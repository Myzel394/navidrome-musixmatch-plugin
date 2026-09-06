package musixmatch

import (
	"strings"
	"testing"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
	"github.com/stretchr/testify/mock"
)

func TestFetchLyricsFromMobileAPIRetriesOnceAfter401WithFreshToken(t *testing.T) {
	oldToken := "stale-token"
	tokenBody := []byte(`{"message":{"header":{"status_code":200},"body":{"user_token":"fresh-token"}}}`)
	blockedMacroBody := []byte(`{"message":{"header":{"status_code":401},"body":{}}}`)
	successMacroBody := mobileFirstMacroFixture(t)

	host.CacheMock.On("GetString", mobileTokenCache).Return(oldToken, true, nil).Once()
	host.CacheMock.On("Remove", mobileTokenCache).Return(nil).Once()
	host.CacheMock.On("GetString", mobileTokenCache).Return("", false, nil).Once()
	host.CacheMock.On("SetString", mobileTokenCache, "fresh-token", int64(mobileTokenTTL/time.Second)).Return(nil).Once()
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "macro.subtitles.get") && strings.Contains(endpoint, "usertoken=stale-token")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, blockedMacroBody)).Once()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "token.get")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, tokenBody)).Once()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "macro.subtitles.get") && strings.Contains(endpoint, "usertoken=fresh-token")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, successMacroBody)).Once()

	t.Cleanup(func() {
		host.CacheMock.ExpectedCalls = nil
		host.CacheMock.Calls = nil
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})

	resp, err, failure, success := fetchLyricsFromMobileAPI(lyrics.GetLyricsRequest{Track: lyrics.TrackInfo{Artist: "Billie Eilish", Title: "Birds of a Feather"}})
	if err != nil {
		t.Fatalf("fetchLyricsFromMobileAPI returned error: %v", err)
	}
	if failure != nil {
		t.Fatalf("fetchLyricsFromMobileAPI returned failure: %v", failure)
	}
	if success == nil || success.CategoryValue() != "mobile_synced" {
		t.Fatalf("expected mobile synced success, got %#v", success)
	}
	if len(resp.Lyrics) == 0 || !strings.Contains(resp.Lyrics[0].Text, "Birds of a feather") {
		t.Fatalf("expected lyrics from retry success, got %#v", resp.Lyrics)
	}
}
