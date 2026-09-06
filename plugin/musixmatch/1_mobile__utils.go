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

func mobileLyricsQuery(input lyrics.GetLyricsRequest, token string) url.Values {
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

func mobileGet(action string, query url.Values, out any) error {
	query.Set("app_id", mobileAppID)
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	body, err := _doMobileGetRequest(fmt.Sprintf(utils.MusixmatchMobileAPIURL, action) + "?" + query.Encode())
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func _doMobileGetRequest(endpoint string) ([]byte, error) {
	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Host", "apic-appmobile.musixmatch.com")
	req.SetHeader("authority", "apic-appmobile.musixmatch.com")
	req.SetHeader("X-Cookie", "x-mxm-token-guid=")
	req.SetHeader("x-mxm-app-version", mobileAppVersion)
	req.SetHeader("X-User-Agent", mobileUserAgent)
	req.SetHeader("Accept-Language", "en-US,en;q=0.9")
	req.SetHeader("Connection", "keep-alive")
	req.SetHeader("Accept", "application/json")
	resp := req.Send()
	if resp.Status() != utils.HTTPStatusOK {
		return resp.Body(), &utils.HTTPError{StatusCode: int(resp.Status())}
	}
	return resp.Body(), nil
}
