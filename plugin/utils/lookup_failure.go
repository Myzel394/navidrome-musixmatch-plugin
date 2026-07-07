package utils

import (
	"errors"
	"fmt"
)

const (
	LookupSourceDesktopAPI = "desktop_api"
	LookupSourceWebsite    = "website"
	LookupSourceUnknown    = "unknown"

	LookupFailureStageDesktopToken            = "desktop_token"
	LookupFailureStageDesktopRequest          = "desktop_request"
	LookupFailureStageDesktopBlocked          = "desktop_blocked"
	LookupFailureStageDesktopNoLyrics         = "desktop_no_lyrics"
	LookupFailureStageWebsiteFallbackDisabled = "website_fallback_disabled"
	LookupFailureStageSearchRequest           = "search_request"
	LookupFailureStageSearchParse             = "search_parse"
	LookupFailureStageSearchNoMatch           = "search_no_match"
	LookupFailureStageLyricsPageRequest       = "lyrics_page_request"
	LookupFailureStageLyricsPageParse         = "lyrics_page_parse"
	LookupFailureStageLyricsEmpty             = "lyrics_empty"
	LookupFailureStageUnknown                 = "unknown"
	LookupFailureReasonUnknown                = "unknown"
)

type LookupFailure struct {
	Stage      string
	Reason     string
	Source     string
	Endpoint   string
	StatusCode int
	Err        error
}

type HTTPError struct {
	StatusCode int
	Endpoint   string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d from %s", e.StatusCode, SanitizeAnalyticsText(e.Endpoint))
}

func NewLookupFailure(stage, reason, source string, err error) *LookupFailure {
	failure := &LookupFailure{
		Stage:  stage,
		Reason: reason,
		Source: source,
		Err:    err,
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		failure.Endpoint = SanitizeAnalyticsText(httpErr.Endpoint)
		failure.StatusCode = httpErr.StatusCode
	}

	return failure
}

func LookupFailureFromError(err error) *LookupFailure {
	var failure *LookupFailure
	if errors.As(err, &failure) {
		return failure
	}
	return nil
}

func (f *LookupFailure) Error() string {
	if f == nil {
		return ""
	}
	if f.Err != nil {
		return f.Err.Error()
	}
	if f.Reason != "" {
		return f.Reason
	}
	return f.StageValue()
}

func (f *LookupFailure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Err
}

func (f *LookupFailure) WithEndpoint(endpoint string) *LookupFailure {
	if f != nil {
		f.Endpoint = SanitizeAnalyticsText(endpoint)
	}
	return f
}

func (f *LookupFailure) WithStatusCode(statusCode int) *LookupFailure {
	if f != nil {
		f.StatusCode = statusCode
	}
	return f
}

func (f *LookupFailure) StageValue() string {
	if f == nil || f.Stage == "" {
		return LookupFailureStageUnknown
	}
	return f.Stage
}

func (f *LookupFailure) ReasonValue() string {
	if f == nil || f.Reason == "" {
		return LookupFailureReasonUnknown
	}
	return f.Reason
}

func (f *LookupFailure) SourceValue() string {
	if f == nil || f.Source == "" {
		return LookupSourceUnknown
	}
	return f.Source
}
