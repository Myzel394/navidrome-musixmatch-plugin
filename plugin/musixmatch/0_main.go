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

// Return error whenever a lyrics could not be fetched.
// Return a failure only if the lookup failed due to a fixable error.
// Some lyrics just don't exist, and we do not need a failure for that.
func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	utils.LogInfof("FetchLyrics: lookup started for artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID)
	var desktopFailure *utils.LookupFailure
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

	// If no user token provided, we cannot fallback to the website,
	// so we return the error now.
	if utils.ConfigUserToken() == "" {
		utils.LogInfof("FetchLyrics: skipping website fallback because musixmatch_user_token is not configured")

		return lyrics.GetLyricsResponse{}, nil, desktopFailure, nil
	}

	// Fallback, scrape website
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
	failure = utils.NewLookupFailure("search_no_match", "website", nil).WithPhase("website_search")
	return lyrics.GetLyricsResponse{}, nil, failure, nil
}
