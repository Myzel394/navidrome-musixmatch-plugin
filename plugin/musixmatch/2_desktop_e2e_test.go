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

	tokenBody := []byte(`{"message":{"header":{"status_code":200},"body":{"user_token":"fixture-token"}}}`)
	macroBody := mobileFirstMacroFixture(t)

	host.CacheMock.On("GetString", mobileTokenCache).Return("", false, nil).Once()
	host.CacheMock.On("SetString", mobileTokenCache, mock.AnythingOfType("string"), int64(mobileTokenTTL/time.Second)).Return(nil).Once()
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyUserToken).Return("", true).Maybe()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "apic-appmobile.musixmatch.com") && strings.Contains(endpoint, "token.get")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, tokenBody)).Once()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "apic-appmobile.musixmatch.com") && strings.Contains(endpoint, "macro.subtitles.get") && strings.Contains(endpoint, "q_artist=Billie+Eilish") && strings.Contains(endpoint, "q_duration=213")
	})).Return(&pdk.HTTPRequest{}).Once()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, macroBody)).Once()
	host.CacheMock.On("GetString", desktopTokenCache).Return("", false, nil).Maybe()
	host.CacheMock.On("SetString", desktopTokenCache, mock.AnythingOfType("string"), int64(desktopTokenTTL/time.Second)).Return(nil).Maybe()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "apic-desktop.musixmatch.com") && strings.Contains(endpoint, "token.get")
	})).Return(&pdk.HTTPRequest{}).Maybe()
	pdk.PDKMock.On("NewHTTPRequest", pdk.MethodGet, mock.MatchedBy(func(endpoint string) bool {
		return strings.Contains(endpoint, "apic-desktop.musixmatch.com") && strings.Contains(endpoint, "macro.subtitles.get")
	})).Return(&pdk.HTTPRequest{}).Maybe()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, tokenBody)).Maybe()
	pdk.PDKMock.On("Send", mock.AnythingOfType("*pdk.HTTPRequest")).Return(pdk.NewStubHTTPResponse(utils.HTTPStatusOK, nil, macroBody)).Maybe()

	t.Cleanup(func() {
		host.CacheMock.ExpectedCalls = nil
		host.CacheMock.Calls = nil
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})
}

func mobileFirstMacroFixture(t *testing.T) []byte {
	t.Helper()
	macro := macroBody{MacroCalls: map[string]macroResponse{}}
	trackBody, _ := json.Marshal(map[string]any{"track": map[string]string{"track_name": "Birds of a Feather", "artist_name": "Billie Eilish", "album_name": ""}})
	subtitleBody, _ := json.Marshal(map[string]any{"subtitle_list": []any{map[string]any{"subtitle": map[string]string{"subtitle_body": "[00:01.00]Birds of a feather\n[00:02.00]We should stick together\n", "subtitle_language": "en"}}}})
	track := macroResponse{}
	track.Message.Header.StatusCode = utils.HTTPStatusOK
	track.Message.Body = trackBody
	subtitle := macroResponse{}
	subtitle.Message.Header.StatusCode = utils.HTTPStatusOK
	subtitle.Message.Body = subtitleBody
	macro.MacroCalls["matcher.track.get"] = track
	macro.MacroCalls["track.subtitles.get"] = subtitle
	body, _ := json.Marshal(macro)
	outer := macroResponse{}
	outer.Message.Header.StatusCode = utils.HTTPStatusOK
	outer.Message.Body = body
	out, _ := json.Marshal(outer)
	return out
}

func withMatchedTrackFixture(t *testing.T, macroFixture []byte, title, artist, album string) []byte {
	t.Helper()
	var outer macroResponse
	if err := json.Unmarshal(macroFixture, &outer); err != nil {
		t.Fatalf("failed to parse macro fixture: %v", err)
	}
	var macro macroBody
	if err := json.Unmarshal(outer.Message.Body, &macro); err != nil {
		t.Fatalf("failed to parse macro body fixture: %v", err)
	}
	trackBody, _ := json.Marshal(map[string]any{"track": map[string]string{"track_name": title, "artist_name": artist, "album_name": album}})
	call := macroResponse{}
	call.Message.Header.StatusCode = utils.HTTPStatusOK
	call.Message.Body = trackBody
	if macro.MacroCalls == nil {
		macro.MacroCalls = map[string]macroResponse{}
	}
	macro.MacroCalls["matcher.track.get"] = call
	body, _ := json.Marshal(macro)
	outer.Message.Body = body
	out, _ := json.Marshal(outer)
	return out
}

func mustGetDesktopTokenBody(t *testing.T) []byte {
	t.Helper()

	query := "app_id=" + desktopAppID + "&user_language=en&t=" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	body, err := realDesktopGetRequest(fmt.Sprintf(utils.MusixmatchDesktopAPIURL, "token.get") + "?" + query)
	if err != nil {
		t.Fatalf("failed to fetch desktop token fixture: %v", err)
	}
	return body
}

func mustGetDesktopMacroBody(t *testing.T, tokenBody []byte) []byte {
	t.Helper()

	var tokenResp macroResponse
	if err := json.Unmarshal(tokenBody, &tokenResp); err != nil {
		t.Fatalf("failed to parse desktop token fixture: %v", err)
	}
	if tokenResp.Message.Header.StatusCode != utils.HTTPStatusOK {
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

	body, err := realDesktopGetRequest(fmt.Sprintf(utils.MusixmatchDesktopAPIURL, "macro.subtitles.get") + "?" + query.Encode())
	if err != nil {
		t.Fatalf("failed to fetch desktop macro fixture: %v", err)
	}
	return body
}

func realDesktopGetRequest(endpoint string) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Cookie", "AWSELBCORS=0; AWSELB=0")
	req.Header.Set("User-Agent", desktopUserAgent)

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
