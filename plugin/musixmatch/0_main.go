package musixmatch

import (
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type Song struct {
	Artist              string
	Title               string
	CommontrackVanityID string
}

func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	utils.LogInfof("FetchLyrics: artist='%s' title='%s'", input.Track.Artist, input.Track.Title)
	if resp, err, _, success := fetchLyricsFromDesktopAPI(input); err == nil && len(resp.Lyrics) > 0 {
		return resp, nil, nil, success
	} else if err != nil {
		utils.LogErrorf("FetchLyrics desktop API fallback: %v", err)
	}

	// Fallback, scrape website when a user token is configured.
	if utils.ConfigUserToken() == "" {
		utils.LogInfof("FetchLyrics: skipping website fallback because musixmatch_user_token is not configured")
		err := fmt.Errorf("website fallback disabled because musixmatch_user_token is not configured")
		failure := utils.NewLookupFailure("website_fallback_disabled", "website", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	track, err, failure := searchForTrack(input)
	if err != nil {
		utils.LogErrorf("FetchLyrics search error: %v", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	if track == nil {
		utils.LogInfof("FetchLyrics: no match found for '%s' - '%s'", input.Track.Artist, input.Track.Title)
		err := fmt.Errorf("no Musixmatch match found for %q by %q", input.Track.Title, input.Track.Artist)
		failure := utils.NewLookupFailure("search_no_match", "website", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	utils.LogInfof("FetchLyrics: matched '%s' by '%s'", track.Title, track.Artist)
	return scrapeWebsiteLyricsForTrack(track)
}
