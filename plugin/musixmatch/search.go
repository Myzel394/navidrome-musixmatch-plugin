package musixmatch

import (
	"fmt"
	"net/url"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func searchForTrack(input lyrics.GetLyricsRequest) (*Song, error) {
	normArtist := normalize(input.Track.Artist)
	normTitle := normalize(input.Track.Title)

	query := fmt.Sprintf("%s %s", normArtist, normTitle)
	endpoint := fmt.Sprintf(utils.MusixmatchSearchPageURL, url.QueryEscape(query))

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		return nil, fmt.Errorf("failed to do musixmatch search request for query %s; Error: %v", query, err)
	}
}
