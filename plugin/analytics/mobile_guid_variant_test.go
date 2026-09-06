package analytics

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
	"github.com/stretchr/testify/mock"
)

func TestLookupMetricsIncludeMobileGUIDVariantWithoutGUID(t *testing.T) {
	resetAnalyticsMocks(t)
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyUserToken).Return("", false).Maybe()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodPost, utils.OpenObserveMetricsURL).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.MatchedBy(bodyContains("\"mobile_guid_variant\":\"random_guid\"", "00010203-0405-4607-8809-0a0b0c0d0e0f"))).Return(pdk.NewStubHTTPResponse(200, nil, nil)).Once()

	reportLookupMetrics(lookupAnalytics{
		Response:          lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Text: "lyrics"}}},
		Success:           utils.NewLookupSuccess("mobile_synced"),
		MobileGUIDVariant: "random_guid",
	}, lyrics.GetLyricsRequest{})

	pdk.PDKMock.ExpectedCalls = nil
	pdk.PDKMock.Calls = nil
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return().Maybe()
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyUserToken).Return("", false).Maybe()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodPost, utils.OpenObserveMetricsURL).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.MatchedBy(bodyContains("\"mobile_guid_variant\":\"no_guid\"", "00010203-0405-4607-8809-0a0b0c0d0e0f"))).Return(pdk.NewStubHTTPResponse(200, nil, nil)).Once()

	reportLookupMetrics(lookupAnalytics{
		Err:               errors.New("boom"),
		Failure:           utils.NewLookupFailure("reason", "mobile_api", errors.New("boom")),
		MobileGUIDVariant: "no_guid",
	}, lyrics.GetLyricsRequest{})
}

func TestLookupFailureTraceIncludesMobileGUIDVariantWithoutGUID(t *testing.T) {
	resetAnalyticsMocks(t)
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyUserToken).Return("", false).Maybe()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodPost, utils.OpenObserveTraceURL).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.MatchedBy(bodyContains("\"stringValue\":\"no_guid\"", "00010203-0405-4607-8809-0a0b0c0d0e0f"))).Return(pdk.NewStubHTTPResponse(200, nil, nil)).Once()

	reportLookupFailureTrace(lookupAnalytics{
		Err:               errors.New("boom"),
		Failure:           utils.NewLookupFailure("reason", "mobile_api", errors.New("boom")),
		StartedAt:         time.Now().UTC(),
		Duration:          time.Millisecond,
		MobileGUIDVariant: "no_guid",
	})
}

func bodyContains(want, forbidden string) func(*pdk.HTTPRequest) bool {
	return func(r *pdk.HTTPRequest) bool {
		body := reflect.ValueOf(r).Elem().FieldByName("body").Bytes()
		text := string(body)
		return strings.Contains(text, want) && !strings.Contains(text, forbidden)
	}
}

func resetAnalyticsMocks(t *testing.T) {
	t.Helper()
	pdk.PDKMock.ExpectedCalls = nil
	pdk.PDKMock.Calls = nil
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return().Maybe()
	t.Cleanup(func() {
		pdk.PDKMock.AssertExpectations(t)
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})
}
