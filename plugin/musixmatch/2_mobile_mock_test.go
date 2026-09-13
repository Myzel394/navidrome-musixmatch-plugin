package musixmatch

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
	"github.com/stretchr/testify/mock"
)

func TestFetchLyricsBirdsOfAFeatherMobileMock(t *testing.T) {
	mockMobileAPI(t)

	resp, format, err, failure, success, variant := FetchLyrics(lyrics.GetLyricsRequest{
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
	if variant != mobileGUIDVariantRandom {
		t.Fatalf("expected random_guid variant, got %q", variant)
	}
	if len(resp.Lyrics) == 0 {
		t.Fatal("expected at least one lyrics result")
	}
	if format != utils.LyricsFormatSyncedLRC {
		t.Fatalf("expected synced LRC format, got %v", format)
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

	host.CacheMock.AssertExpectations(t)
	host.HTTPMock.AssertExpectations(t)
	pdk.PDKMock.AssertExpectations(t)
}

func mockMobileAPI(t *testing.T) {
	t.Helper()
	originalReader := mobileExperimentRandomReader
	mobileExperimentRandomReader = bytes.NewReader([]byte{1, 'f', 'i', 'x', 't', 'u', 'r', 'e', '-', 'g', 'u', 'i', 'd', '-', '0', '0', '1'})

	tokenBody := []byte(`{"message":{"header":{"status_code":200},"body":{"user_token":"fixture-token"}}}`)
	macroBody := mobileFirstMacroFixture(t)

	host.CacheMock.On("GetString", mobileTokenCache).Return("", false, nil).Once()
	host.CacheMock.On("SetString", mobileTokenCache, mock.AnythingOfType("string"), int64(mobileTokenTTL/time.Second)).Return(nil).Once()
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyUserToken).Return("", true).Maybe()
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyMobileUserAgent).Return("", false).Maybe()
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyMobileAppVersion).Return("", false).Maybe()
	pdk.PDKMock.On("GetConfig", utils.ConfigKeyMobileAppID).Return("", false).Maybe()
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return request.Method == pdk.MethodGet.String() &&
			request.NoFollowRedirects &&
			strings.Contains(request.URL, "apic-appmobile.musixmatch.com") &&
			strings.Contains(request.URL, "token.get") &&
			strings.Contains(request.URL, "app_id="+utils.DefaultMobileAppID) &&
			strings.Contains(request.URL, "user_language=en")
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: tokenBody}, nil).Once()
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return request.Method == pdk.MethodGet.String() &&
			request.NoFollowRedirects &&
			strings.Contains(request.URL, "apic-appmobile.musixmatch.com") &&
			strings.Contains(request.URL, "macro.subtitles.get") &&
			strings.Contains(request.URL, "format=json") &&
			strings.Contains(request.URL, "app_id="+utils.DefaultMobileAppID) &&
			strings.Contains(request.URL, "q_artist=Billie+Eilish") &&
			strings.Contains(request.URL, "q_track=Birds+of+a+Feather") &&
			strings.Contains(request.URL, "q_duration=213") &&
			strings.Contains(request.URL, "usertoken=fixture-token")
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: macroBody}, nil).Once()

	t.Cleanup(func() {
		mobileExperimentRandomReader = originalReader
		host.CacheMock.ExpectedCalls = nil
		host.CacheMock.Calls = nil
		host.HTTPMock.ExpectedCalls = nil
		host.HTTPMock.Calls = nil
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
