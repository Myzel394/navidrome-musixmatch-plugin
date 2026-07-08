package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

const (
	analyticsSchemaVersion = "1"
	maxAnalyticsTextLen    = 2000

	metricLookupSuccessTotal = "navidrome_musixmatch_lookup_success_total"
	metricLookupFailureTotal = "navidrome_musixmatch_lookup_failure_total"
	metricLookupDurationMS   = "navidrome_musixmatch_lookup_duration_ms"
)

type lookupAnalytics struct {
	Input     lyrics.GetLyricsRequest
	Response  lyrics.GetLyricsResponse
	Err       error
	Failure   *LookupFailure
	Success   *LookupSuccess
	StartedAt time.Time
	Duration  time.Duration
	Logs      []CapturedLog
}

type metricRecord map[string]any

type otlpExportTraceServiceRequest struct {
	ResourceSpans []otlpResourceSpans `json:"resourceSpans"`
}

type otlpResourceSpans struct {
	Resource   otlpResource     `json:"resource"`
	ScopeSpans []otlpScopeSpans `json:"scopeSpans"`
}

type otlpResource struct {
	Attributes []otlpAttribute `json:"attributes,omitempty"`
}

type otlpScopeSpans struct {
	Scope otlpScope  `json:"scope"`
	Spans []otlpSpan `json:"spans"`
}

type otlpScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type otlpSpan struct {
	TraceID           string          `json:"traceId"`
	SpanID            string          `json:"spanId"`
	Name              string          `json:"name"`
	Kind              int             `json:"kind,omitempty"`
	StartTimeUnixNano string          `json:"startTimeUnixNano"`
	EndTimeUnixNano   string          `json:"endTimeUnixNano"`
	Attributes        []otlpAttribute `json:"attributes,omitempty"`
	Events            []otlpEvent     `json:"events,omitempty"`
	Status            otlpStatus      `json:"status"`
}

type otlpEvent struct {
	TimeUnixNano string          `json:"timeUnixNano"`
	Name         string          `json:"name"`
	Attributes   []otlpAttribute `json:"attributes,omitempty"`
}

type otlpStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

type otlpAttribute struct {
	Key   string    `json:"key"`
	Value otlpValue `json:"value"`
}

type otlpValue struct {
	StringValue string   `json:"stringValue,omitempty"`
	IntValue    string   `json:"intValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
}

func ReportLyricsLookup(input lyrics.GetLyricsRequest, resp lyrics.GetLyricsResponse, err error, failure *LookupFailure, success *LookupSuccess, startedAt time.Time, duration time.Duration, logs []CapturedLog) {
	if strings.TrimSpace(OpenObserveAuthToken) == "" {
		return
	}

	lookup := lookupAnalytics{
		Input:     input,
		Response:  resp,
		Err:       err,
		Failure:   failure,
		Success:   success,
		StartedAt: startedAt,
		Duration:  duration,
		Logs:      logs,
	}

	if ConfigShareMetrics() {
		reportLookupMetrics(lookup)
	}
	if ConfigShareErrors() && !lookup.success() {
		reportLookupFailureTrace(lookup)
	}
}

func (l lookupAnalytics) success() bool {
	return l.Err == nil && len(l.Response.Lyrics) > 0
}

func (l lookupAnalytics) lookupFailure() *LookupFailure {
	if l.Failure != nil {
		return l.Failure
	}
	return LookupFailureFromError(l.Err)
}

func reportLookupMetrics(lookup lookupAnalytics) {
	result := "failure"
	if lookup.success() {
		result = "success"
	}

	now := time.Now().UTC().Unix()
	base := metricRecord{
		"plugin":         PluginName,
		"schema_version": analyticsSchemaVersion,
		"result":         result,
		"_timestamp":     now,
	}

	if lookup.success() {
		base["source"] = classifyLookupSuccess(lookup.Success)
	} else {
		failure := lookup.lookupFailure()
		base["failure_reason"] = failure.ReasonValue()
		base["failure_source"] = failure.SourceValue()
	}

	metrics := []metricRecord{
		metricWith(base, metricLookupDurationMS, "gauge", float64(lookup.Duration.Milliseconds())),
	}
	if lookup.success() {
		metrics = append(metrics, metricWith(base, metricLookupSuccessTotal, "counter", 1))
	} else {
		metrics = append(metrics, metricWith(base, metricLookupFailureTotal, "counter", 1))
	}

	sendOpenObserveJSON(OpenObserveMetricsURL, metrics)
}

func metricWith(base metricRecord, name, metricType string, value any) metricRecord {
	record := make(metricRecord, len(base)+3)
	for k, v := range base {
		record[k] = v
	}
	record["__name__"] = name
	record["__type__"] = metricType
	record["value"] = value
	return record
}

func reportLookupFailureTrace(lookup lookupAnalytics) {
	end := lookup.StartedAt.Add(lookup.Duration)
	traceID, spanID := traceIDs(lookup)
	failure := lookup.lookupFailure()
	failureMessage := "lyrics lookup returned no lyrics"
	if lookup.Err != nil {
		failureMessage = lookup.Err.Error()
	} else if failure != nil {
		failureMessage = failure.Error()
	}
	failureMessage = truncateAnalyticsText(failureMessage)

	span := otlpSpan{
		TraceID:           traceID,
		SpanID:            spanID,
		Name:              "lyrics lookup failed",
		Kind:              1,
		StartTimeUnixNano: unixNanoString(lookup.StartedAt),
		EndTimeUnixNano:   unixNanoString(end),
		Attributes: []otlpAttribute{
			otlpStringAttr("lookup.result", "failure"),
			otlpStringAttr("lookup.failure_reason", failure.ReasonValue()),
			otlpStringAttr("lookup.source", failure.SourceValue()),
			otlpDoubleAttr("lookup.duration_ms", float64(lookup.Duration.Milliseconds())),
			otlpStringAttr("track.artist", lookup.Input.Track.Artist),
			otlpStringAttr("track.title", lookup.Input.Track.Title),
			otlpBoolAttr("config.has_musixmatch_user_token", ConfigUserToken() != ""),

			otlpStringAttr("plugin.source", OpenObserveAttributePluginSource),
			otlpStringAttr("service.version", OpenObserveAttributeVersion),
		},
		Events: failureLogEvents(lookup.Logs),
		Status: otlpStatus{Code: 2, Message: failureMessage},
	}
	if lookup.Input.Track.Album != "" {
		span.Attributes = append(span.Attributes, otlpStringAttr("track.album", lookup.Input.Track.Album))
	}
	if lookup.Input.Track.MBZRecordingID != "" {
		span.Attributes = append(span.Attributes, otlpStringAttr("track.mbz_recording_id", lookup.Input.Track.MBZRecordingID))
	}
	if failure != nil && failure.StatusCode > 0 {
		span.Attributes = append(span.Attributes, otlpIntAttr("http.status_code", failure.StatusCode))
	}

	payload := otlpExportTraceServiceRequest{
		ResourceSpans: []otlpResourceSpans{{
			Resource: otlpResource{Attributes: []otlpAttribute{
				otlpStringAttr("service.name", PluginName),
				otlpStringAttr("telemetry.sdk.name", "navidrome-plugin"),
				otlpStringAttr("analytics.schema_version", analyticsSchemaVersion),
			}},
			ScopeSpans: []otlpScopeSpans{{
				Scope: otlpScope{Name: PluginName},
				Spans: []otlpSpan{span},
			}},
		}},
	}

	sendOpenObserveJSON(OpenObserveTraceURL, payload)
}

func failureLogEvents(logs []CapturedLog) []otlpEvent {
	events := make([]otlpEvent, 0, len(logs))
	for _, log := range logs {
		logTime, err := time.Parse(time.RFC3339Nano, log.Timestamp)
		if err != nil {
			logTime = time.Now().UTC()
		}
		events = append(events, otlpEvent{
			TimeUnixNano: unixNanoString(logTime),
			Name:         "log",
			Attributes: []otlpAttribute{
				otlpStringAttr("log.level", log.Level),
				otlpStringAttr("log.message", truncateAnalyticsText(log.Message)),
			},
		})
	}
	return events
}

func sendOpenObserveJSON(endpoint string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		LogErrorf("analytics marshal failed: %v", err)
		return
	}

	req := pdk.NewHTTPRequest(pdk.MethodPost, endpoint)
	req.SetHeader("Authorization", openObserveAuthorizationHeader())
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("User-Agent", PluginName)
	req.SetBody(body)

	resp := req.Send()
	if resp.Status() < 200 || resp.Status() >= 300 {
		LogErrorf("analytics POST failed: HTTP %d from %s", resp.Status(), endpoint)
	}
}

func openObserveAuthorizationHeader() string {
	token := strings.TrimSpace(OpenObserveAuthToken)
	return "Basic " + token
}

func classifyLookupSuccess(success *LookupSuccess) string {
	return success.CategoryValue()
}

func traceIDs(lookup lookupAnalytics) (string, string) {
	seed := fmt.Sprintf("%d:%s:%s:%s", lookup.StartedAt.UnixNano(), lookup.Input.Track.Artist, lookup.Input.Track.Title, lookup.Err)
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:16]), hex.EncodeToString(sum[16:24])
}

func unixNanoString(t time.Time) string {
	return strconv.FormatInt(t.UTC().UnixNano(), 10)
}

func otlpStringAttr(key, value string) otlpAttribute {
	return otlpAttribute{Key: key, Value: otlpValue{StringValue: truncateAnalyticsText(value)}}
}

func otlpDoubleAttr(key string, value float64) otlpAttribute {
	return otlpAttribute{Key: key, Value: otlpValue{DoubleValue: &value}}
}

func otlpIntAttr(key string, value int) otlpAttribute {
	return otlpAttribute{Key: key, Value: otlpValue{IntValue: strconv.Itoa(value)}}
}

func otlpBoolAttr(key string, value bool) otlpAttribute {
	return otlpAttribute{Key: key, Value: otlpValue{BoolValue: &value}}
}

func truncateAnalyticsText(s string) string {
	if len(s) <= maxAnalyticsTextLen {
		return s
	}
	return s[:maxAnalyticsTextLen] + "..."
}
