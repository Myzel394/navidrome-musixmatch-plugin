package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func buildMobileLyricsQuery(input lyrics.GetLyricsRequest, token string) url.Values {
	q := url.Values{}
	q.Set("format", "json")
	q.Set("namespace", "lyrics_richsynched")
	q.Set("optional_calls", "track.richsync")
	q.Set("subtitle_format", "lrc")
	q.Set("q_artist", input.Track.Artist)
	q.Set("q_track", input.Track.Title)
	q.Set("usertoken", token)
	if input.Track.Album != "" {
		q.Set("q_album", input.Track.Album)
	}
	if input.Track.Duration > 0 {
		q.Set("q_duration", strconv.Itoa(int(input.Track.Duration+0.5)))
	}
	return q
}

func buildMobileSearchQuery(input lyrics.GetLyricsRequest, token string) url.Values {
	q := url.Values{}
	q.Set("format", "json")
	q.Set("q_track", input.Track.Title)
	q.Set("q_artist", input.Track.Artist)
	q.Set("page", "1")
	q.Set("page_size", "10")
	q.Set("f_has_lyrics", "1")
	q.Set("usertoken", token)
	return q
}

func buildMobileTrackLyricsQuery(trackID int, token string) url.Values {
	q := url.Values{}
	q.Set("format", "json")
	q.Set("namespace", "lyrics_richsynched")
	q.Set("optional_calls", "track.richsync")
	q.Set("subtitle_format", "lrc")
	q.Set("track_id", strconv.Itoa(trackID))
	q.Set("usertoken", token)
	return q
}

func mobileGet(action string, query url.Values) (*macroResponse, error) {
	query.Set("app_id", utils.ConfigMobileAppID())
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	body, err := utils.DoMobileGetRequest(fmt.Sprintf(utils.MusixmatchMobileAPIURL, action) + "?" + query.Encode())
	if err != nil {
		return nil, err
	}

	response := macroResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse mobile API response for %s: %w", action, err)
	}

	return &response, nil
}

func mobileGetMacro(query url.Values) (macroBody, error, *utils.LookupFailure) {
	response, err := mobileGet("macro.subtitles.get", query)
	if err != nil {
		failure := utils.NewLookupFailure("mobile_macro_request_failed", "mobile_api", err).WithPhase("mobile_lyrics")
		return macroBody{}, err, failure
	}

	// Check status
	utils.LogInfof("mobile API: lyrics response received status=%d body_bytes=%d", response.Message.Header.StatusCode, len(response.Message.Body))
	if response.Message.Header.StatusCode == utils.HTTPStatusBlocked {
		err := fmt.Errorf("mobile API returned 401 for lyrics request")
		failure := utils.NewLookupFailure("mobile_blocked", "mobile_api", err).WithPhase("mobile_lyrics").WithStatusCode(response.Message.Header.StatusCode)
		return macroBody{}, err, failure
	}
	if response.Message.Header.StatusCode != utils.HTTPStatusOK {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(response.Message.Body)))
		err := fmt.Errorf("mobile API returned status %d for lyrics request", response.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("mobile_macro_status", "mobile_api", err).WithPhase("mobile_lyrics").WithStatusCode(response.Message.Header.StatusCode)
		return macroBody{}, err, failure
	}

	// Parse
	var body macroBody
	if err := json.Unmarshal(response.Message.Body, &body); err != nil {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(response.Message.Body)))
		failure := utils.NewLookupFailure("mobile_macro_parse", "mobile_api", err).WithPhase("mobile_lyrics")
		return macroBody{}, err, failure
	}
	return body, nil, nil
}
