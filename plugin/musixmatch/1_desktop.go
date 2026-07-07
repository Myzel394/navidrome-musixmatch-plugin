package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func fetchLyricsFromDesktopAPI(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure) {
	utils.LogInfof("desktop API: fetching lyrics for '%s' by '%s'", input.Track.Title, input.Track.Artist)

	token, err, failure := desktopUserToken()
	if err != nil {
		return lyrics.GetLyricsResponse{}, err, failure
	}

	query := desktopLyricsQuery(input, token)

	var resp desktopResponse
	endpoint, err := desktopGet("macro.subtitles.get", query, &resp)
	if err != nil {
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopRequest, "macro_request_failed", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint)
		return lyrics.GetLyricsResponse{}, err, failure
	}
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		_ = host.CacheRemove(desktopTokenCache)
		err := fmt.Errorf("desktop API returned 401 for lyrics request")
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopBlocked, "macro_blocked", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint).WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		err := fmt.Errorf("desktop API returned status %d for lyrics request", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopRequest, "macro_status", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint).WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure
	}

	var body desktopMacroBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		err = fmt.Errorf("failed to parse desktop API macro body: %w", err)
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopRequest, "macro_parse", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint)
		return lyrics.GetLyricsResponse{}, err, failure
	}

	if resp, ok := desktopLyricsFromRichsync(body.MacroCalls["track.richsync.get"]); ok {
		return resp, nil, nil
	}
	if resp, ok := desktopLyricsFromSubtitle(body.MacroCalls["track.subtitles.get"]); ok {
		return resp, nil, nil
	}
	if resp, ok := desktopLyricsFromPlain(body.MacroCalls["track.lyrics.get"]); ok {
		return resp, nil, nil
	}

	err = fmt.Errorf(desktopFallbackErr)
	failure = utils.NewLookupFailure(utils.LookupFailureStageDesktopNoLyrics, "no_desktop_lyrics", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint)
	return lyrics.GetLyricsResponse{}, err, failure
}

func desktopLyricsQuery(input lyrics.GetLyricsRequest, token string) url.Values {
	query := url.Values{}
	query.Set("format", "json")
	query.Set("namespace", "lyrics_richsynced")
	query.Set("optional_calls", "track.richsync")
	query.Set("subtitle_format", "lrc")
	query.Set("q_artist", input.Track.Artist)
	query.Set("q_track", input.Track.Title)
	query.Set("usertoken", token)
	if input.Track.Duration > 0 {
		query.Set("f_subtitle_length", strconv.Itoa(int(input.Track.Duration+0.5)))
		query.Set("f_subtitle_length_max_deviation", "3")
	}
	return query
}

func desktopUserToken() (string, error, *utils.LookupFailure) {
	if token, ok, err := host.CacheGetString(desktopTokenCache); err != nil {
		utils.LogErrorf("desktop API token cache read failed: %v", err)
	} else if ok && token != "" {
		return token, nil, nil
	}

	query := url.Values{}
	query.Set("user_language", "en")

	var resp desktopResponse
	endpoint, err := desktopGet("token.get", query, &resp)
	if err != nil {
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopToken, "token_request_failed", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint)
		return "", err, failure
	}
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		err := fmt.Errorf("desktop API returned 401 while fetching token")
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopBlocked, "token_blocked", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint).WithStatusCode(resp.Message.Header.StatusCode)
		return "", err, failure
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		err := fmt.Errorf("desktop API returned status %d while fetching token", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopToken, "token_status", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint).WithStatusCode(resp.Message.Header.StatusCode)
		return "", err, failure
	}

	var body desktopTokenBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		err = fmt.Errorf("failed to parse desktop API token body: %w", err)
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopToken, "token_parse", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint)
		return "", err, failure
	}
	if body.UserToken == "" {
		err := fmt.Errorf("desktop API returned empty token")
		failure := utils.NewLookupFailure(utils.LookupFailureStageDesktopToken, "token_empty", utils.LookupSourceDesktopAPI, err).WithEndpoint(endpoint)
		return "", err, failure
	}

	if err := host.CacheSetString(desktopTokenCache, body.UserToken, int64(desktopTokenTTL/time.Second)); err != nil {
		utils.LogErrorf("desktop API token cache write failed: %v", err)
	}

	return body.UserToken, nil, nil
}

func desktopGet(action string, query url.Values, out any) (string, error) {
	query.Set("app_id", desktopAppID)
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	endpoint := fmt.Sprintf(utils.MusixmatchDesktopAPIURL, action) + "?" + query.Encode()
	body, err := doDesktopGetRequest(endpoint)
	if err != nil {
		return endpoint, err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return endpoint, fmt.Errorf("failed to parse desktop API response for %s: %w", action, err)
	}
	return endpoint, nil
}

func doDesktopGetRequest(endpoint string) ([]byte, error) {
	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Accept-Language", "en")
	req.SetHeader("Cookie", "AWSELBCORS=0; AWSELB=0")
	req.SetHeader("User-Agent", desktopUserAgent)

	resp := req.Send()
	if resp.Status() != utils.HTTPStatusOK {
		return resp.Body(), &utils.HTTPError{StatusCode: int(resp.Status()), Endpoint: endpoint}
	}
	return resp.Body(), nil
}

func desktopLyricsFromRichsync(call desktopResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != desktopAPISuccess || len(call.Message.Body) == 0 {
		return lyrics.GetLyricsResponse{}, false
	}

	var body desktopRichsyncBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil || body.Richsync.Body == "" {
		return lyrics.GetLyricsResponse{}, false
	}

	var lines []desktopRichsyncLine
	if err := json.Unmarshal([]byte(body.Richsync.Body), &lines); err != nil {
		utils.LogErrorf("desktop API richsync parse failed: %v", err)
		return lyrics.GetLyricsResponse{}, false
	}

	var b strings.Builder
	for _, line := range lines {
		if line.Text == "" {
			continue
		}
		b.WriteString(totalToLRC(line.Timestamp))
		b.WriteString(line.Text)
		b.WriteByte('\n')
	}
	if b.Len() == 0 {
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got richsync lyrics (%d lines LRC)", strings.Count(b.String(), "\n"))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: body.Richsync.Language, Text: b.String()}}}, true
}

func desktopLyricsFromSubtitle(call desktopResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != desktopAPISuccess || len(call.Message.Body) == 0 {
		return lyrics.GetLyricsResponse{}, false
	}

	var body desktopSubtitleBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil || len(body.SubtitleList) == 0 {
		return lyrics.GetLyricsResponse{}, false
	}

	subtitle := body.SubtitleList[0].Subtitle
	if subtitle.Body == "" {
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got subtitle lyrics (%d lines LRC)", strings.Count(subtitle.Body, "\n"))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: subtitle.Language, Text: subtitle.Body}}}, true
}

func desktopLyricsFromPlain(call desktopResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != desktopAPISuccess || len(call.Message.Body) == 0 {
		return lyrics.GetLyricsResponse{}, false
	}

	var body desktopLyricsBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil || body.Lyrics.Body == "" {
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got plain lyrics (%d chars)", len(body.Lyrics.Body))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: body.Lyrics.Language, Text: body.Lyrics.Body}}}, true
}
