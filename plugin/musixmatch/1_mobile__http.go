package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
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

func mobileGet(action string, query url.Values, assignment mobileGUIDAssignment) (*macroResponse, error) {
	query.Set("app_id", utils.ConfigMobileAppID())
	query.Set("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	headerOverrides := map[string]string{}
	if assignment.GUID != "" {
		headerOverrides["X-Cookie"] = "x-mxm-token-guid=" + assignment.GUID
	}
	body, err := utils.DoMobileGetRequest(fmt.Sprintf(utils.MusixmatchMobileAPIURL, action)+"?"+query.Encode(), headerOverrides)
	if err != nil {
		return nil, err
	}

	response := macroResponse{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse mobile API response for %s: %w", action, err)
	}

	return &response, nil
}
