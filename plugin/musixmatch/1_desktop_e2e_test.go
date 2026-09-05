package musixmatch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
	"github.com/stretchr/testify/mock"
)

func TestFetchLyricsBirdsOfAFeatherDesktopEndToEnd(t *testing.T) {
	mockDesktopAPI(t)

	resp, err, failure, success := FetchLyrics(lyrics.GetLyricsRequest{
		Track: lyrics.TrackInfo{
			Artist:   "Billie Eilish",
			Title:    "Birds of a Feather",
			Duration: 213,
		},
	})
	if err != nil {
		t.Fatalf("FetchLyrics returned error: %v", err)
	}
	if failure != nil {
		t.Fatalf("FetchLyrics returned failure: %v", failure)
	}
	if success == nil {
		t.Fatal("expected success metadata")
	}
	if len(resp.Lyrics) == 0 {
		t.Fatal("expected at least one lyrics result")
	}

	got := resp.Lyrics[0]
	if got.Lang != "en" {
		t.Fatalf("expected language en, got %q", got.Lang)
	}
	if !strings.Contains(strings.ToLower(got.Text), "birds of a feather") {
		t.Fatalf("expected fetched lyrics to contain chorus, got %q", got.Text)
	}
	if !strings.Contains(got.Text, "[00:") {
		t.Fatalf("expected synced LRC lyrics, got %q", got.Text)
	}
}

func mockDesktopAPI(t *testing.T) {
	t.Helper()

	guid := newDesktopInstallGUID()
	if guid == "" {
		t.Fatal("could not generate install guid")
	}
	tokenBody := mustGetDesktopTokenBody(t, guid)
	macroBody := mustGetDesktopMacroBody(t, tokenBody, guid)

	host.CacheMock.On("GetString", desktopTokenCache).Return("", false, nil).Once()
	host.CacheMock.On("GetString", desktopGUIDCache).Return("", false, nil).Once()
	host.CacheMock.On("SetString", desktopGUIDCache, mock.AnythingOfType("string"), int64(desktopTokenTTL/time.Second)).Return(nil).Once()
	host.CacheMock.On("SetString", desktopTokenCache, mock.AnythingOfType("string"), int64(desktopTokenTTL/time.Second)).Return(nil).Once()
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "token.get")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, tokenBody)).Once()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "macro.subtitles.get") && strings.Contains(endpoint, "q_artist=Billie+Eilish")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, macroBody)).Once()

	t.Cleanup(func() {
		host.CacheMock.ExpectedCalls = nil
		host.CacheMock.Calls = nil
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})
}

func mustGetDesktopTokenBody(t *testing.T, guid string) []byte {
	t.Helper()

	query := "app_id=" + desktopAppID + "&user_language=en&guid=" + guid + "&t=" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	body, err := realDesktopGetRequest(fmt.Sprintf(utils.MusixmatchDesktopAPIURL, "token.get")+"?"+query, guid)
	if err != nil {
		t.Fatalf("failed to fetch desktop token fixture: %v", err)
	}
	return body
}

func mustGetDesktopMacroBody(t *testing.T, tokenBody []byte, guid string) []byte {
	t.Helper()

	var tokenResp desktopResponse
	if err := json.Unmarshal(tokenBody, &tokenResp); err != nil {
		t.Fatalf("failed to parse desktop token fixture: %v", err)
	}
	if tokenResp.Message.Header.StatusCode != desktopAPISuccess {
		t.Fatalf("desktop token fixture returned status %d: %s", tokenResp.Message.Header.StatusCode, tokenBody)
	}
	if len(tokenResp.Message.Body) == 0 {
		t.Fatalf("desktop token fixture returned empty body: %s", tokenBody)
	}
	var token desktopTokenBody
	if err := json.Unmarshal(tokenResp.Message.Body, &token); err != nil {
		t.Fatalf("failed to parse desktop token body fixture: %v", err)
	}
	if token.UserToken == "" {
		t.Fatalf("desktop token fixture returned empty user token: %s", tokenBody)
	}

	query := desktopLyricsQuery(lyrics.GetLyricsRequest{Track: lyrics.TrackInfo{
		Artist:   "Billie Eilish",
		Title:    "Birds of a Feather",
		Duration: 213,
	}}, token.UserToken)
	query.Set("app_id", desktopAppID)
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	body, err := realDesktopGetRequest(fmt.Sprintf(utils.MusixmatchDesktopAPIURL, "macro.subtitles.get")+"?"+query.Encode(), guid)
	if err != nil {
		t.Fatalf("failed to fetch desktop macro fixture: %v", err)
	}
	return body
}

func realDesktopGetRequest(endpoint string, guid string) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range desktopAPIHeaders(guid) {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != utils.HTTPStatusOK {
		return body, fmt.Errorf("error code %d returned from Musixmatch for endpoint %s", resp.StatusCode, endpoint)
	}
	return body, nil
}
