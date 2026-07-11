package musixmatch

import (
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type Song struct {
	Artist              string
	Title               string
	CommontrackVanityID string
}

func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	utils.LogInfof("FetchLyrics: lookup started for artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID)
	if resp, err, desktopFailure, success := fetchLyricsFromDesktopAPI(input); err == nil && len(resp.Lyrics) > 0 {
		utils.LogInfof("FetchLyrics: lookup succeeded source=desktop_api category=%s", success.CategoryValue())
		return resp, nil, nil, success
	} else if err != nil {
		status := 0
		if desktopFailure != nil {
			status = desktopFailure.StatusCode
		}
		utils.LogErrorf("FetchLyrics: desktop API lookup failed reason=%s status=%d error=%v", desktopFailure.ReasonValue(), status, err)
	}

	// Fallback, scrape website when a user token is configured.
	if utils.ConfigUserToken() == "" {
		utils.LogInfof("FetchLyrics: skipping website fallback because musixmatch_user_token is not configured")
		failure := utils.NewLookupFailure("website_fallback_disabled", "website", nil)
		return lyrics.GetLyricsResponse{}, nil, failure, nil
	}

	track, err, failure := searchForTrack(input)
	if err != nil {
		status := 0
		if failure != nil {
			status = failure.StatusCode
		}
		utils.LogErrorf("FetchLyrics: website search failed reason=%s status=%d error=%v", failure.ReasonValue(), status, err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	if track != nil {
		utils.LogInfof("FetchLyrics: website search match_found")
		return scrapeWebsiteLyricsForTrack(track)
	}

	utils.LogInfof("FetchLyrics: website search no_match_found")
	failure = utils.NewLookupFailure("search_no_match", "website", nil)
	return lyrics.GetLyricsResponse{}, nil, failure, nil
}
