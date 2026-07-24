package musixmatch

import (
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
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
		utils.LogErrorf("FetchLyrics: desktop API lookup failed reason=%s status=%d error=%v", desktopFailure.ReasonValue(), desktopFailure.StatusCodeValue(), err)
	}

	// If no user token provided, we cannot fallback to the website,
	// so we return the error now.
	if utils.ConfigUserToken() == "" {
		utils.LogInfof("FetchLyrics: skipping website fallback because musixmatch_user_token is not configured")

		return lyrics.GetLyricsResponse{}, nil, desktopFailure, nil
	}

	// Fallback, scrape website
	tracks, err, failure := searchForTracks(input)
	if err != nil {
		status := 0
		reason := ""
		if failure != nil {
			status = failure.StatusCode
			reason = failure.ReasonValue()
		}
		utils.LogErrorf("FetchLyrics: website search failed reason=%s status=%d error=%v", reason, status, err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	if len(tracks) > 0 {
		utils.LogInfof("FetchLyrics: website search match_found")
		var lastFailure *utils.LookupFailure
		for _, track := range tracks {
			pdk.Log(pdk.LogDebug, fmt.Sprintf("FetchLyrics: website candidate found artist=%s title=%s album=%s mbz=%s", track.Artist, track.Title, track.CommontrackVanityID, input.Track.MBZRecordingID))
			resp, err, failure, success := scrapeWebsiteLyricsForTrack(track, input)
			if isIdentityRejection(err) {
				utils.LogInfof("FetchLyrics: website candidate rejected reason=%s", err.Error())
				continue
			}
			if err != nil {
				return resp, err, failure, success
			}
			if len(resp.Lyrics) > 0 {
				utils.LogInfof("FetchLyrics: website candidate accepted source=website category=%s", success.CategoryValue())
				return resp, nil, nil, success
			}
			lastFailure = failure
		}

		utils.LogInfof("FetchLyrics: Could not find any lyrics for any of the candidate tracks, returning last failure")
		return lyrics.GetLyricsResponse{}, nil, lastFailure, nil
	}

	utils.LogInfof("FetchLyrics: website search no_match_found")
	failure = utils.NewLookupFailure("search_no_match", "website", nil).WithPhase("website_search")
	return lyrics.GetLyricsResponse{}, nil, failure, nil
}
