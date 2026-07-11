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

func desktopGet(action string, query url.Values, out any) error {
	query.Set("app_id", desktopAppID)
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))

	endpoint := fmt.Sprintf(utils.MusixmatchDesktopAPIURL, action) + "?" + query.Encode()
	body, err := _doDesktopGetRequest(endpoint)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("failed to parse desktop API response for %s: %w", action, err)
	}
	return nil
}

func _doDesktopGetRequest(endpoint string) ([]byte, error) {
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
