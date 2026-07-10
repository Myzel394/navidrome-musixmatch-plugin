package analytics

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
)

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
			otlpBoolAttr("config.has_musixmatch_user_token", utils.ConfigUserToken() != ""),

			otlpStringAttr("plugin.source", utils.OpenObserveAttributePluginSource),
			otlpStringAttr("service.version", utils.OpenObserveAttributeVersion),
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
				otlpStringAttr("service.name", utils.PluginName),
				otlpStringAttr("telemetry.sdk.name", "navidrome-plugin"),
				otlpStringAttr("analytics.schema_version", analyticsSchemaVersion),
			}},
			ScopeSpans: []otlpScopeSpans{{
				Scope: otlpScope{Name: utils.PluginName},
				Spans: []otlpSpan{span},
			}},
		}},
	}

	sendOpenObserveJSON(utils.OpenObserveTraceURL, payload)
}

func failureLogEvents(logs []utils.CapturedLog) []otlpEvent {
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

func traceIDs(lookup lookupAnalytics) (string, string) {
	seed := fmt.Sprintf("%d:%s:%s:%s", lookup.StartedAt.UnixNano(), lookup.Input.Track.Artist, lookup.Input.Track.Title, lookup.Err)
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:16]), hex.EncodeToString(sum[16:24])
}
