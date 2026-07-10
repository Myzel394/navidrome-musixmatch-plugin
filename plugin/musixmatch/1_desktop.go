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

func fetchLyricsFromDesktopAPI(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	utils.LogInfof("desktop API: lookup started duration_filter=%t", input.Track.Duration > 0)

	token, err, failure := desktopUserToken()
	if err != nil {
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	query := desktopLyricsQuery(input, token)

	var resp desktopResponse
	err = desktopGet("macro.subtitles.get", query, &resp)
	if err != nil {
		utils.LogErrorf("desktop API: lyrics request failed error=%v", err)
		failure := utils.NewLookupFailure("desktop_macro_request_failed", "desktop_api", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("desktop API: lyrics response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		_ = host.CacheRemove(desktopTokenCache)
		err := fmt.Errorf("desktop API returned 401 for lyrics request")
		failure := utils.NewLookupFailure("desktop_macro_blocked", "desktop_api", err).WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		err := fmt.Errorf("desktop API returned status %d for lyrics request", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("desktop_macro_status", "desktop_api", err).WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	var body desktopMacroBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		utils.LogErrorf("desktop API: failed to parse desktop API macro body body_bytes=%d error=%v", len(resp.Message.Body), err)
		failure := utils.NewLookupFailure("desktop_macro_parse", "desktop_api", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("desktop API: parsed lyrics response calls=%d richsync_available=%t subtitle_available=%t plain_available=%t", len(body.MacroCalls), body.MacroCalls["track.richsync.get"].Message.Header.StatusCode != 0, body.MacroCalls["track.subtitles.get"].Message.Header.StatusCode != 0, body.MacroCalls["track.lyrics.get"].Message.Header.StatusCode != 0)

	if resp, ok := desktopLyricsFromRichsync(body.MacroCalls["track.richsync.get"]); ok {
		success := utils.NewLookupSuccess("desktop_synced")
		return resp, nil, nil, success
	}
	if resp, ok := desktopLyricsFromSubtitle(body.MacroCalls["track.subtitles.get"]); ok {
		success := utils.NewLookupSuccess("desktop_synced")
		return resp, nil, nil, success
	}
	if resp, ok := desktopLyricsFromPlain(body.MacroCalls["track.lyrics.get"]); ok {
		success := utils.NewLookupSuccess("desktop_plain")
		return resp, nil, nil, success
	}

	utils.LogInfof("desktop API: lookup finished without lyrics")
	err = fmt.Errorf(desktopFallbackErr)
	failure = utils.NewLookupFailure("desktop_no_lyrics", "desktop_api", err)
	return lyrics.GetLyricsResponse{}, err, failure, nil
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
		utils.LogInfof("desktop API: token cache hit")
		return token, nil, nil
	}
	utils.LogInfof("desktop API: token cache miss")

	query := url.Values{}
	query.Set("user_language", "en")

	var resp desktopResponse
	err := desktopGet("token.get", query, &resp)
	if err != nil {
		utils.LogErrorf("desktop API: token request failed error=%v", err)
		failure := utils.NewLookupFailure("desktop_token_request_failed", "desktop_api", err)
		return "", err, failure
	}
	utils.LogInfof("desktop API: token response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		err := fmt.Errorf("desktop API returned 401 while fetching token")
		failure := utils.NewLookupFailure("desktop_token_blocked", "desktop_api", err).WithStatusCode(resp.Message.Header.StatusCode)
		return "", err, failure
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		err := fmt.Errorf("desktop API returned status %d while fetching token", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("desktop_token_status", "desktop_api", err).WithStatusCode(resp.Message.Header.StatusCode)
		return "", err, failure
	}

	var body desktopTokenBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		utils.LogErrorf("desktop API: failed to parse desktop API token body body_bytes=%d error=%v", len(resp.Message.Body), err)
		failure := utils.NewLookupFailure("desktop_token_parse", "desktop_api", err)
		return "", err, failure
	}
	if body.UserToken == "" {
		utils.LogInfof("desktop API: parsed token response token_present=false")
		err := fmt.Errorf("desktop API returned empty token")
		failure := utils.NewLookupFailure("desktop_token_empty", "desktop_api", err)
		return "", err, failure
	}
	utils.LogInfof("desktop API: parsed token response token_present=true")

	if err := host.CacheSetString(desktopTokenCache, body.UserToken, int64(desktopTokenTTL/time.Second)); err != nil {
		utils.LogErrorf("desktop API token cache write failed: %v", err)
	}

	return body.UserToken, nil, nil
}

func desktopGet(action string, query url.Values, out any) error {
	query.Set("app_id", desktopAppID)
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	endpoint := fmt.Sprintf(utils.MusixmatchDesktopAPIURL, action) + "?" + query.Encode()
	body, err := doDesktopGetRequest(endpoint)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("failed to parse desktop API response for %s: %w", action, err)
	}
	return nil
}

func doDesktopGetRequest(endpoint string) ([]byte, error) {
	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Accept-Language", "en")
	req.SetHeader("Cookie", "AWSELBCORS=0; AWSELB=0")
	req.SetHeader("User-Agent", desktopUserAgent)

	resp := req.Send()
	if resp.Status() != utils.HTTPStatusOK {
		return resp.Body(), &utils.HTTPError{StatusCode: int(resp.Status())}
	}
	return resp.Body(), nil
}

func desktopLyricsFromRichsync(call desktopResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != desktopAPISuccess {
		utils.LogInfof("desktop API: richsync lyrics unavailable because status was not successful status=%d body_present=%t", call.Message.Header.StatusCode, len(call.Message.Body) > 0)
		return lyrics.GetLyricsResponse{}, false
	}
	if len(call.Message.Body) == 0 {
		utils.LogInfof("desktop API: richsync lyrics unavailable because response body was empty status=%d", call.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, false
	}

	var body desktopRichsyncBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		utils.LogInfof("desktop API: richsync lyrics unavailable because response could not be parsed body_bytes=%d", len(call.Message.Body))
		return lyrics.GetLyricsResponse{}, false
	}
	if body.Richsync.Body == "" {
		utils.LogInfof("desktop API: richsync lyrics unavailable because lyrics body was empty")
		return lyrics.GetLyricsResponse{}, false
	}

	var lines []desktopRichsyncLine
	if err := json.Unmarshal([]byte(body.Richsync.Body), &lines); err != nil {
		utils.LogErrorf("desktop API richsync parse failed: %v", err)
		return lyrics.GetLyricsResponse{}, false
	}

	var builder strings.Builder
	for _, line := range lines {
		if line.Text == "" {
			builder.WriteByte('\n')
			continue
		}
		builder.WriteString(totalToLRC(line.Timestamp))
		builder.WriteString(line.Text)
		builder.WriteByte('\n')
	}
	if builder.Len() == 0 {
		utils.LogInfof("desktop API: richsync lyrics unavailable because no LRC lines were rendered parsed_lines=%d", len(lines))
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got richsync lyrics (%d lines LRC)", strings.Count(builder.String(), "\n"))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: body.Richsync.Language, Text: builder.String()}}}, true
}

func desktopLyricsFromSubtitle(call desktopResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != desktopAPISuccess {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because status was not successful status=%d body_present=%t", call.Message.Header.StatusCode, len(call.Message.Body) > 0)
		return lyrics.GetLyricsResponse{}, false
	}
	if len(call.Message.Body) == 0 {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because response body was empty status=%d", call.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, false
	}

	var body desktopSubtitleBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because response could not be parsed body_bytes=%d", len(call.Message.Body))
		return lyrics.GetLyricsResponse{}, false
	}
	if len(body.SubtitleList) == 0 {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because subtitle list was empty")
		return lyrics.GetLyricsResponse{}, false
	}

	subtitle := body.SubtitleList[0].Subtitle
	if subtitle.Body == "" {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because lyrics body was empty list_count=%d", len(body.SubtitleList))
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got subtitle lyrics (%d lines LRC)", strings.Count(subtitle.Body, "\n"))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: subtitle.Language, Text: subtitle.Body}}}, true
}

func desktopLyricsFromPlain(call desktopResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != desktopAPISuccess {
		utils.LogInfof("desktop API: plain lyrics unavailable because status was not successful status=%d body_present=%t", call.Message.Header.StatusCode, len(call.Message.Body) > 0)
		return lyrics.GetLyricsResponse{}, false
	}
	if len(call.Message.Body) == 0 {
		utils.LogInfof("desktop API: plain lyrics unavailable because response body was empty status=%d", call.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, false
	}

	var body desktopLyricsBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		utils.LogInfof("desktop API: plain lyrics unavailable because response could not be parsed body_bytes=%d", len(call.Message.Body))
		return lyrics.GetLyricsResponse{}, false
	}
	if body.Lyrics.Body == "" {
		utils.LogInfof("desktop API: plain lyrics unavailable because lyrics body was empty")
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got plain lyrics (%d chars)", len(body.Lyrics.Body))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: body.Lyrics.Language, Text: body.Lyrics.Body}}}, true
}
