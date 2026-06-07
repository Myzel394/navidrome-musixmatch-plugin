package musixmatch

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type nextDataResponse struct {
	Props struct {
		PageProps struct {
			Data struct {
				TrackInfo struct {
					Data struct {
						Lyrics struct {
							Body     string `json:"body"`
							Language string `json:"language"`
						} `json:"lyrics"`
					} `json:"data"`
				} `json:"trackInfo"`
			} `json:"data"`
		} `json:"pageProps"`
	} `json:"props"`
}

func fetchLyricsForTrack(track *Song) (lyrics.GetLyricsResponse, error) {
	endpoint := fmt.Sprintf(utils.MusixmatchFetchPageURL, track.CommontrackVanityID)

	utils.LogInfof("fetching lyrics page: %s", endpoint)

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		utils.LogErrorf("fetch page failed for %s: %v", track.CommontrackVanityID, err)
		return lyrics.GetLyricsResponse{}, fmt.Errorf("failed to fetch musixmatch page for %s: %v", track.CommontrackVanityID, err)
	}

	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		utils.LogErrorf("no __NEXT_DATA__ on page for %s", track.CommontrackVanityID)
		return lyrics.GetLyricsResponse{}, fmt.Errorf("could not find __NEXT_DATA__ on page for %s", track.CommontrackVanityID)
	}

	var data nextDataResponse
	if err := json.Unmarshal(matches[1], &data); err != nil {
		utils.LogErrorf("__NEXT_DATA__ parse failed for %s: %v", track.CommontrackVanityID, err)
		return lyrics.GetLyricsResponse{}, fmt.Errorf("failed to parse __NEXT_DATA__ for %s: %v", track.CommontrackVanityID, err)
	}

	lyricsBody := data.Props.PageProps.Data.TrackInfo.Data.Lyrics.Body
	if lyricsBody == "" {
		utils.LogErrorf("empty lyrics body for %s", track.CommontrackVanityID)
		return lyrics.GetLyricsResponse{}, fmt.Errorf("no lyrics found for track %s", track.CommontrackVanityID)
	}

	utils.LogInfof("got lyrics for %s (%d chars)", track.CommontrackVanityID, len(lyricsBody))
	return lyrics.GetLyricsResponse{
		Lyrics: []lyrics.LyricsText{{Text: lyricsBody}},
	}, nil
}
