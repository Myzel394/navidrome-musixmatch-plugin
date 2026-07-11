package analytics

import (
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
)

const (
	metricLookupSuccessTotal = "navidrome_musixmatch_lookup_success_total"
	metricLookupFailureTotal = "navidrome_musixmatch_lookup_failure_total"
	metricLookupDurationMS   = "navidrome_musixmatch_lookup_duration_ms"
)

type metricRecord map[string]any

func reportLookupMetrics(lookup lookupAnalytics) {
	result := "failure"
	if lookup.success() {
		result = "success"
	}

	now := time.Now().UTC().Unix()
	base := metricRecord{
		"plugin":                    utils.PluginName,
		"plugin_version":            utils.OpenObserveAttributeVersion,
		"has_musixmatch_user_token": utils.ConfigUserToken() != "",
		"schema_version":            analyticsSchemaVersion,
		"result":                    result,
		"_timestamp":                now,
	}

	if lookup.success() {
		category := classifyLookupSuccess(lookup.Success)
		base["category"] = category
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

	sendOpenObserveJSON(utils.OpenObserveMetricsURL, metrics)
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

func classifyLookupSuccess(success *utils.LookupSuccess) string {
	return success.CategoryValue()
}
