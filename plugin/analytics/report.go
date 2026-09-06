package analytics

import (
	"strings"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

const (
	analyticsSchemaVersion = "2"
	maxAnalyticsTextLen    = 2000
)

type lookupAnalytics struct {
	Input             lyrics.GetLyricsRequest
	Response          lyrics.GetLyricsResponse
	Err               error
	Failure           *utils.LookupFailure
	Success           *utils.LookupSuccess
	MobileGUIDVariant string
	StartedAt         time.Time
	Duration          time.Duration
	Logs              []utils.CapturedLog
}

func ReportLyricsLookup(input lyrics.GetLyricsRequest, resp lyrics.GetLyricsResponse, err error, failure *utils.LookupFailure, success *utils.LookupSuccess, mobileGUIDVariant string, startedAt time.Time, duration time.Duration, logs []utils.CapturedLog) {
	if strings.TrimSpace(utils.OpenObserveAuthToken) == "" {
		return
	}

	lookup := lookupAnalytics{
		Input:             input,
		Response:          resp,
		Err:               err,
		Failure:           failure,
		Success:           success,
		MobileGUIDVariant: mobileGUIDVariant,
		StartedAt:         startedAt,
		Duration:          duration,
		Logs:              logs,
	}

	if utils.ConfigShareMetrics() {
		reportLookupMetrics(lookup, input)
	}
	if utils.ConfigShareErrors() && !lookup.success() && err != nil {
		reportLookupFailureTrace(lookup)
	}
}

func (l lookupAnalytics) success() bool {
	return l.Err == nil && len(l.Response.Lyrics) > 0
}

func (l lookupAnalytics) lookupFailure() *utils.LookupFailure {
	if l.Failure != nil {
		return l.Failure
	}
	return utils.LookupFailureFromError(l.Err)
}

func truncateAnalyticsText(s string) string {
	if len(s) <= maxAnalyticsTextLen {
		return s
	}
	return s[:maxAnalyticsTextLen] + "..."
}
