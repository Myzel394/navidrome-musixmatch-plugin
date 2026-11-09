package musixmatch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
	"github.com/stretchr/testify/mock"
)

// Musixmatch keeps duplicate track entries per song: one twin carries synced
// lyrics, the other carries nothing. The search returns both twins, with the
// empty one first; the lookup must still return the synced sibling.
func TestFetchLyricsStoryTimeDuplicateSweep(t *testing.T) {
	mockStoryTimeDuplicateAPI(t)

	resp, format, err, failure, success, _ := FetchLyrics(lyrics.GetLyricsRequest{
		Track: lyrics.TrackInfo{
			Artist:   "Fivio Foreign",
			Title:    "Story Time",
			Duration: 162,
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
	if format != utils.LyricsFormatSyncedLRC {
		t.Fatalf("expected synced LRC format, got %v", format)
	}
	if len(resp.Lyrics) == 0 {
		t.Fatal("expected at least one lyrics result")
	}
	if !strings.Contains(resp.Lyrics[0].Text, "[00:") {
		t.Fatalf("expected synced LRC lyrics, got %q", resp.Lyrics[0].Text)
	}
	if !strings.Contains(strings.ToLower(resp.Lyrics[0].Text), "sad story") {
		t.Fatalf("expected sweep to resolve the synced twin, got %q", resp.Lyrics[0].Text)
	}

	host.CacheMock.AssertExpectations(t)
	host.HTTPMock.AssertExpectations(t)
	pdk.PDKMock.AssertExpectations(t)
}

func TestMobileGetMacroReportsFailureReason(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantReason string
	}{
		{"blocked", `{"message":{"header":{"status_code":401},"body":""}}`, "mobile_blocked"},
		{"bad status", `{"message":{"header":{"status_code":404},"body":""}}`, "mobile_macro_status"},
		{"malformed body", `{"message":{"header":{"status_code":200},"body":"not a macro"}}`, "mobile_macro_parse"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
			pdk.PDKMock.On("GetConfig", mock.Anything).Return("", false).Maybe()
			host.HTTPMock.On("Send", mock.Anything).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: []byte(test.body)}, nil).Once()
			t.Cleanup(func() {
				host.HTTPMock.ExpectedCalls = nil
				host.HTTPMock.Calls = nil
				pdk.PDKMock.ExpectedCalls = nil
				pdk.PDKMock.Calls = nil
			})

			_, err, failure := mobileGetMacro(buildMobileTrackLyricsQuery(1, "fixture-token"), mobileGUIDAssignment{})
			if err == nil {
				t.Fatal("expected an error")
			}
			if failure.ReasonValue() != test.wantReason {
				t.Fatalf("expected reason %q, got %q", test.wantReason, failure.ReasonValue())
			}
		})
	}
}

func TestFetchLyricsPicksLyricsAcrossTracks(t *testing.T) {
	const (
		firstID  = 224027345
		secondID = 223175197
	)
	syncedMacro := func(t *testing.T) []byte { return macroResponseJSON(t, "[00:01.00] synced line\n", "") }
	poisonedMacro := func(t *testing.T) []byte { return macroResponseJSON(t, poisonedLyricsFixture, "") }
	plainMacro := func(text string) func(t *testing.T) []byte {
		return func(t *testing.T) []byte { return macroResponseJSON(t, "", text) }
	}
	tests := []struct {
		name string
		// Tracks without an entry have no mock, so fetching one panics. That
		// is how the early return on synced lyrics is checked.
		macros       map[int]func(t *testing.T) []byte
		wantFormat   utils.LyricsFormat
		wantCategory string
		wantText     string
	}{
		{
			name:         "stops at first synced",
			macros:       map[int]func(t *testing.T) []byte{firstID: syncedMacro},
			wantFormat:   utils.LyricsFormatSyncedLRC,
			wantCategory: "mobile_synced",
			wantText:     "synced line",
		},
		{
			name:         "later synced beats earlier plain",
			macros:       map[int]func(t *testing.T) []byte{firstID: plainMacro("plain line"), secondID: syncedMacro},
			wantFormat:   utils.LyricsFormatSyncedLRC,
			wantCategory: "mobile_synced",
			wantText:     "synced line",
		},
		{
			name:         "poisoned synced is skipped",
			macros:       map[int]func(t *testing.T) []byte{firstID: poisonedMacro, secondID: syncedMacro},
			wantFormat:   utils.LyricsFormatSyncedLRC,
			wantCategory: "mobile_synced",
			wantText:     "synced line",
		},
		{
			name:         "first plain wins without synced",
			macros:       map[int]func(t *testing.T) []byte{firstID: plainMacro("first plain"), secondID: plainMacro("second plain")},
			wantFormat:   utils.LyricsFormatPlain,
			wantCategory: "mobile_plain",
			wantText:     "first plain",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStoryTimeToken(t)
			mockStoryTimeSearch(t, firstID, secondID)
			for trackID, macro := range test.macros {
				mockStoryTimeMacro(trackID, macro(t))
			}

			resp, format, err, failure, success, _ := FetchLyrics(lyrics.GetLyricsRequest{
				Track: lyrics.TrackInfo{Artist: "Fivio Foreign", Title: "Story Time", Duration: 162},
			})
			if err != nil || failure != nil || success == nil {
				t.Fatalf("expected success, got err=%v failure=%v success=%v", err, failure, success)
			}
			if format != test.wantFormat {
				t.Fatalf("expected format %v, got %v", test.wantFormat, format)
			}
			if success.CategoryValue() != test.wantCategory {
				t.Fatalf("expected category %q, got %q", test.wantCategory, success.CategoryValue())
			}
			if !strings.Contains(resp.Lyrics[0].Text, test.wantText) {
				t.Fatalf("expected lyrics containing %q, got %q", test.wantText, resp.Lyrics[0].Text)
			}
			host.HTTPMock.AssertExpectations(t)
		})
	}
}

func TestFilterMobileTrackIDsOrdersByLengthDifference(t *testing.T) {
	tests := []struct {
		name     string
		duration float32
		lengths  []int
		want     []int
	}{
		{"closest length first", 391, []int{423, 420, 391, 380}, []int{3, 4, 2, 1}},
		{"ties keep search order", 100, []int{90, 110, 100}, []int{3, 1, 2}},
		{"no duration keeps search order", 0, []int{423, 391}, []int{1, 2}},
		{"unknown length goes last", 391, []int{0, 500}, []int{2, 1}},
		{"sorted before limit", 200, []int{100, 110, 120, 130, 140, 200}, []int{6, 5, 4, 3, 2}},
	}
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return().Maybe()
	t.Cleanup(func() {
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trackList := make([]any, 0, len(test.lengths))
			for i, length := range test.lengths {
				trackList = append(trackList, map[string]any{"track": map[string]any{"track_id": i + 1, "track_name": "Story Time", "artist_name": "Fivio Foreign", "track_length": length, "has_lyrics": 1, "instrumental": 0}})
			}
			data, err := json.Marshal(map[string]any{"track_list": trackList})
			if err != nil {
				t.Fatalf("could not build search fixture: %v", err)
			}
			var body mobileSearchBody
			if err := json.Unmarshal(data, &body); err != nil {
				t.Fatalf("could not parse search fixture: %v", err)
			}

			got := filterMobileTrackIDs(lyrics.GetLyricsRequest{
				Track: lyrics.TrackInfo{Artist: "Fivio Foreign", Title: "Story Time", Duration: test.duration},
			}, body)
			if !slices.Equal(got, test.want) {
				t.Fatalf("expected track IDs %v, got %v", test.want, got)
			}
		})
	}
}

var blockedMacroFixture = []byte(`{"message":{"header":{"status_code":401},"body":""}}`)

func mockStoryTimeQueryMacro(body []byte) {
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return strings.Contains(request.URL, "macro.subtitles.get") &&
			strings.Contains(request.URL, "q_track=Story+Time") &&
			!strings.Contains(request.URL, "track_id=")
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: body}, nil).Once()
}

func TestFetchLyricsRunsQueryWhenEveryTrackLookupFails(t *testing.T) {
	mockStoryTimeToken(t)
	mockStoryTimeSearch(t, 224027345)
	mockStoryTimeMacro(224027345, blockedMacroFixture)
	mockStoryTimeQueryMacro(macroResponseJSON(t, "[00:01.00] synced line\n", ""))

	resp, format, err, failure, success, _ := FetchLyrics(lyrics.GetLyricsRequest{
		Track: lyrics.TrackInfo{Artist: "Fivio Foreign", Title: "Story Time", Duration: 162},
	})
	if err != nil || failure != nil || success == nil {
		t.Fatalf("expected success, got err=%v failure=%v success=%v", err, failure, success)
	}
	if format != utils.LyricsFormatSyncedLRC {
		t.Fatalf("expected synced LRC format, got %v", format)
	}
	if !strings.Contains(resp.Lyrics[0].Text, "synced line") {
		t.Fatalf("expected query lyrics, got %q", resp.Lyrics[0].Text)
	}
	host.HTTPMock.AssertExpectations(t)
}

func TestFetchLyricsReportsBlockWhenQueryIsBlockedToo(t *testing.T) {
	mockStoryTimeToken(t)
	mockStoryTimeSearch(t, 224027345)
	mockStoryTimeMacro(224027345, blockedMacroFixture)
	mockStoryTimeQueryMacro(blockedMacroFixture)

	_, _, err, failure, success, _ := FetchLyrics(lyrics.GetLyricsRequest{
		Track: lyrics.TrackInfo{Artist: "Fivio Foreign", Title: "Story Time", Duration: 162},
	})
	if err == nil || success != nil {
		t.Fatalf("expected an error, got err=%v success=%v", err, success)
	}
	if failure.ReasonValue() != "mobile_blocked" {
		t.Fatalf("expected reason mobile_blocked, got %q", failure.ReasonValue())
	}
	host.HTTPMock.AssertExpectations(t)
}

// The query call has no mock here, so running it would panic.
func TestFetchLyricsSkipsQueryWhenATrackLookupSucceeds(t *testing.T) {
	mockStoryTimeToken(t)
	mockStoryTimeSearch(t, 224027345, 223175197)
	mockStoryTimeMacro(224027345, blockedMacroFixture)
	mockStoryTimeMacro(223175197, macroResponseJSON(t, "", ""))

	_, _, err, failure, success, _ := FetchLyrics(lyrics.GetLyricsRequest{
		Track: lyrics.TrackInfo{Artist: "Fivio Foreign", Title: "Story Time", Duration: 162},
	})
	if err == nil || success != nil {
		t.Fatalf("expected an error, got err=%v success=%v", err, success)
	}
	if failure.ReasonValue() != "mobile_no_lyrics" {
		t.Fatalf("expected reason mobile_no_lyrics, got %q", failure.ReasonValue())
	}
	host.HTTPMock.AssertExpectations(t)
}

const poisonedLyricsFixture = "[00:01.00] wob gopini den tefe woxica fero gogoh vudob wiya\n"

func TestFetchMobileLyricsByQueryRejectsPoisonedLyrics(t *testing.T) {
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	pdk.PDKMock.On("GetConfig", mock.Anything).Return("", false).Maybe()
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return strings.Contains(request.URL, "macro.subtitles.get") &&
			strings.Contains(request.URL, "q_track=Story+Time")
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: macroResponseJSON(t, poisonedLyricsFixture, "plain line")}, nil).Once()
	t.Cleanup(func() {
		host.HTTPMock.ExpectedCalls = nil
		host.HTTPMock.Calls = nil
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})

	resp, format, err, failure, success := fetchMobileLyricsByQuery(lyrics.GetLyricsRequest{
		Track: lyrics.TrackInfo{Artist: "Fivio Foreign", Title: "Story Time", Duration: 162},
	}, "fixture-token", mobileGUIDAssignment{})
	if err != nil || failure != nil || success == nil {
		t.Fatalf("expected success, got err=%v failure=%v success=%v", err, failure, success)
	}
	if format != utils.LyricsFormatPlain {
		t.Fatalf("expected plain format, got %v", format)
	}
	if !strings.Contains(resp.Lyrics[0].Text, "plain line") {
		t.Fatalf("expected plain lyrics, got %q", resp.Lyrics[0].Text)
	}
}

func TestFindBestPlainLyricsReportsNoLyrics(t *testing.T) {
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	t.Cleanup(func() {
		pdk.PDKMock.ExpectedCalls = nil
		pdk.PDKMock.Calls = nil
	})

	_, _, err, failure, success := findBestPlainLyrics([]macroBody{macroBodyFromJSON(t, macroResponseJSON(t, "", ""))})
	if err == nil || success != nil {
		t.Fatalf("expected no lyrics error, got err=%v success=%v", err, success)
	}
	if failure.ReasonValue() != "mobile_no_lyrics" {
		t.Fatalf("expected reason mobile_no_lyrics, got %q", failure.ReasonValue())
	}
}

func macroBodyFromJSON(t *testing.T, data []byte) macroBody {
	t.Helper()
	var outer macroResponse
	if err := json.Unmarshal(data, &outer); err != nil {
		t.Fatalf("could not parse macro envelope: %v", err)
	}
	var body macroBody
	if err := json.Unmarshal(outer.Message.Body, &body); err != nil {
		t.Fatalf("could not parse macro body: %v", err)
	}
	return body
}

func TestMobileSearchTrackParsesFlagsAsBools(t *testing.T) {
	tests := []struct {
		name             string
		payload          string
		wantHasLyrics    mobileBool
		wantInstrumental mobileBool
	}{
		{"has lyrics", `{"has_lyrics":1,"instrumental":0}`, true, false},
		{"instrumental", `{"has_lyrics":0,"instrumental":1}`, false, true},
		{"missing", `{}`, false, false},
		{"null", `{"has_lyrics":null,"instrumental":null}`, false, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var track mobileSearchTrack
			if err := json.Unmarshal([]byte(test.payload), &track); err != nil {
				t.Fatalf("unmarshal returned error: %v", err)
			}
			if track.HasLyrics != test.wantHasLyrics {
				t.Fatalf("expected HasLyrics=%v, got %v", test.wantHasLyrics, track.HasLyrics)
			}
			if track.Instrumental != test.wantInstrumental {
				t.Fatalf("expected Instrumental=%v, got %v", test.wantInstrumental, track.Instrumental)
			}
		})
	}
}

func mockStoryTimeToken(t *testing.T) {
	t.Helper()
	originalReader := mobileExperimentRandomReader
	mobileExperimentRandomReader = bytes.NewReader([]byte{1, 'f', 'i', 'x', 't', 'u', 'r', 'e', '-', 'g', 'u', 'i', 'd', '-', '0', '0', '1'})

	tokenBody := []byte(`{"message":{"header":{"status_code":200},"body":{"user_token":"fixture-token"}}}`)

	host.CacheMock.On("GetString", mobileTokenCache).Return("", false, nil).Once()
	host.CacheMock.On("SetString", mobileTokenCache, mock.AnythingOfType("string"), int64(mobileTokenTTL/time.Second)).Return(nil).Once()
	pdk.PDKMock.On("Log", mock.Anything, mock.Anything).Return()
	pdk.PDKMock.On("GetConfig", mock.Anything).Return("", false).Maybe()
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return request.Method == pdk.MethodGet.String() &&
			request.NoFollowRedirects &&
			strings.Contains(request.URL, "apic-appmobile.musixmatch.com") &&
			strings.Contains(request.URL, "token.get")
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: tokenBody}, nil).Once()

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

func mockStoryTimeDuplicateAPI(t *testing.T) {
	t.Helper()
	mockStoryTimeToken(t)

	// Search returns the empty twin first, synced twin second.
	mockStoryTimeSearch(t, 224027345, 223175197)

	// Explicit track_id macros: empty twin stays empty, synced twin delivers.
	mockStoryTimeMacro(224027345, macroResponseJSON(t, "", ""))
	mockStoryTimeMacro(223175197, macroResponseJSON(t, "[00:12.85] Let me tell a little sad story about this young boy\n", ""))
}

func mockStoryTimeSearch(t *testing.T, trackIDs ...int) {
	t.Helper()
	trackList := make([]any, 0, len(trackIDs))
	for _, trackID := range trackIDs {
		trackList = append(trackList, map[string]any{"track": map[string]any{"track_id": trackID, "track_name": "Story Time", "artist_name": "Fivio Foreign", "album_name": "Story Time", "has_lyrics": 1, "instrumental": 0}})
	}
	searchBody, err := json.Marshal(map[string]any{"message": map[string]any{
		"header": map[string]any{"status_code": 200},
		"body":   map[string]any{"track_list": trackList},
	}})
	if err != nil {
		t.Fatalf("could not build search fixture: %v", err)
	}
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return strings.Contains(request.URL, "track.search") &&
			strings.Contains(request.URL, "q_track=Story+Time")
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: searchBody}, nil).Once()
}

func mockStoryTimeMacro(trackID int, body []byte) {
	host.HTTPMock.On("Send", mock.MatchedBy(func(request host.HTTPRequest) bool {
		return strings.Contains(request.URL, "macro.subtitles.get") &&
			strings.Contains(request.URL, "optional_calls=track.richsync") &&
			strings.Contains(request.URL, fmt.Sprintf("track_id=%d", trackID))
	})).Return(&host.HTTPResponse{StatusCode: int32(utils.HTTPStatusOK), Body: body}, nil).Once()
}

// Macro envelope with a track identity plus optional subtitle/plain bodies.
// Empty strings produce empty-but-present calls, like the real API.
func macroResponseJSON(t *testing.T, subtitleText, plainText string) []byte {
	t.Helper()
	trackBody, _ := json.Marshal(map[string]any{"track": map[string]string{"track_name": "Story Time", "artist_name": "Fivio Foreign", "album_name": "Story Time"}})
	macro := macroBody{MacroCalls: map[string]macroResponse{}}
	track := macroResponse{}
	track.Message.Header.StatusCode = utils.HTTPStatusOK
	track.Message.Body = trackBody
	macro.MacroCalls["matcher.track.get"] = track
	subtitle := macroResponse{}
	subtitle.Message.Header.StatusCode = utils.HTTPStatusOK
	subBody, _ := json.Marshal(map[string]any{"subtitle_list": []any{map[string]any{"subtitle": map[string]string{"subtitle_body": subtitleText, "subtitle_language": "en"}}}})
	subtitle.Message.Body = subBody
	macro.MacroCalls["track.subtitles.get"] = subtitle
	plain := macroResponse{}
	plain.Message.Header.StatusCode = utils.HTTPStatusOK
	plainBody, _ := json.Marshal(map[string]any{"lyrics": map[string]string{"lyrics_body": plainText, "lyrics_language": "en"}})
	plain.Message.Body = plainBody
	macro.MacroCalls["track.lyrics.get"] = plain
	body, _ := json.Marshal(macro)
	outer := macroResponse{}
	outer.Message.Header.StatusCode = utils.HTTPStatusOK
	outer.Message.Body = body
	out, _ := json.Marshal(outer)
	return out
}
