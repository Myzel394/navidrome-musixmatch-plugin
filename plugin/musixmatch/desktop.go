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

const (
	desktopAppID       = "web-desktop-app-v1.0"
	desktopTokenCache  = "musixmatch_desktop_user_token"
	desktopTokenTTL    = 10 * 60
	desktopUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36"
	desktopAPISuccess  = 200
	desktopAPIBlocked  = 401
	desktopFallbackErr = "desktop API did not return lyrics"
)

type desktopResponse struct {
	Message struct {
		Header desktopHeader   `json:"header"`
		Body   json.RawMessage `json:"body"`
	} `json:"message"`
}

type desktopHeader struct {
	StatusCode int `json:"status_code"`
}

type desktopTokenBody struct {
	UserToken string `json:"user_token"`
}

type desktopMacroBody struct {
	MacroCalls map[string]desktopResponse `json:"macro_calls"`
}

type desktopSubtitleBody struct {
	SubtitleList []struct {
		Subtitle struct {
			Body     string `json:"subtitle_body"`
			Language string `json:"subtitle_language"`
		} `json:"subtitle"`
	} `json:"subtitle_list"`
}

type desktopRichsyncBody struct {
	Richsync struct {
		Body     string `json:"richsync_body"`
		Language string `json:"richssync_language"`
	} `json:"richsync"`
}

type desktopLyricsBody struct {
	Lyrics struct {
		Body     string `json:"lyrics_body"`
		Language string `json:"lyrics_language"`
	} `json:"lyrics"`
}

type desktopRichsyncLine struct {
	Timestamp float64 `json:"ts"`
	Text      string  `json:"x"`
}

func fetchLyricsFromDesktopAPI(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error) {
	utils.LogInfof("desktop API: fetching lyrics for '%s' by '%s'", input.Track.Title, input.Track.Artist)

	token, err := desktopUserToken()
	if err != nil {
		return lyrics.GetLyricsResponse{}, err
	}

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

	var resp desktopResponse
	if err := desktopGet("macro.subtitles.get", query, &resp); err != nil {
		return lyrics.GetLyricsResponse{}, err
	}
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		_ = host.CacheSetString(desktopTokenCache, "", 1)
		return lyrics.GetLyricsResponse{}, fmt.Errorf("desktop API returned 401 for lyrics request")
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		return lyrics.GetLyricsResponse{}, fmt.Errorf("desktop API returned status %d for lyrics request", resp.Message.Header.StatusCode)
	}

	var body desktopMacroBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		return lyrics.GetLyricsResponse{}, fmt.Errorf("failed to parse desktop API macro body: %w", err)
	}

	if resp, ok := desktopLyricsFromRichsync(body.MacroCalls["track.richsync.get"]); ok {
		return resp, nil
	}
	if resp, ok := desktopLyricsFromSubtitle(body.MacroCalls["track.subtitles.get"]); ok {
		return resp, nil
	}
	if resp, ok := desktopLyricsFromPlain(body.MacroCalls["track.lyrics.get"]); ok {
		return resp, nil
	}

	return lyrics.GetLyricsResponse{}, fmt.Errorf(desktopFallbackErr)
}

func desktopUserToken() (string, error) {
	if token, ok, err := host.CacheGetString(desktopTokenCache); err == nil && ok && token != "" {
		return token, nil
	} else if err != nil {
		utils.LogErrorf("desktop API token cache read failed: %v", err)
	}

	query := url.Values{}
	query.Set("user_language", "en")

	var resp desktopResponse
	if err := desktopGet("token.get", query, &resp); err != nil {
		return "", err
	}
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		return "", fmt.Errorf("desktop API returned 401 while fetching token")
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		return "", fmt.Errorf("desktop API returned status %d while fetching token", resp.Message.Header.StatusCode)
	}

	var body desktopTokenBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		return "", fmt.Errorf("failed to parse desktop API token body: %w", err)
	}
	if body.UserToken == "" {
		return "", fmt.Errorf("desktop API returned empty token")
	}

	if err := host.CacheSetString(desktopTokenCache, body.UserToken, desktopTokenTTL); err != nil {
		utils.LogErrorf("desktop API token cache write failed: %v", err)
	}

	return body.UserToken, nil
}

func desktopGet(action string, query url.Values, out any) error {
	query.Set("app_id", desktopAppID)
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	endpoint := fmt.Sprintf(utils.MusixmatchDesktopAPIURL, action) + "?" + query.Encode()
	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Accept", "application/json")
	req.SetHeader("Accept-Language", "en")
	req.SetHeader("Cookie", "AWSELBCORS=0; AWSELB=0")
	req.SetHeader("User-Agent", desktopUserAgent)

	resp := req.Send()
	if resp.Status() != utils.HTTPStatusOK {
		return fmt.Errorf("HTTP %d from desktop API endpoint %s", resp.Status(), action)
	}
	if err := json.Unmarshal(resp.Body(), out); err != nil {
		return fmt.Errorf("failed to parse desktop API response for %s: %w", action, err)
	}
	return nil
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
