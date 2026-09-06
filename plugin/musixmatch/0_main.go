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

// Return error whenever a lyrics could not be fetched.
// Return a failure only if the lookup failed due to a fixable error.
// Some lyrics just don't exist, and we do not need a failure for that.
func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess, string) {
	mobileGUIDAssignment := chooseMobileGUIDAssignment()
	// // // Mobile
	utils.LogInfof("FetchLyrics: lookup started for artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID)
	if resp, err, failure, success := fetchLyricsFromMobileAPI(input, mobileGUIDAssignment); err == nil && len(resp.Lyrics) > 0 {
		utils.LogInfof("FetchLyrics: lookup succeeded source=mobile_api category=%s", success.CategoryValue())
		return resp, nil, nil, success, mobileGUIDAssignment.Variant
	} else if err != nil {
		utils.LogErrorf("FetchLyrics: mobile API lookup failed reason=%s status=%d error=%v", failure.ReasonValue(), failure.StatusCodeValue(), err)
	}

	// // // Website
	if utils.ConfigUserToken() != "" {
		resp, err, failure, success := fetchLyricsViaWebsiteScraping(input)

		if err == nil && len(resp.Lyrics) > 0 {
			utils.LogInfof("FetchLyrics: lookup succeeded source=mobile_api category=%s", success.CategoryValue())
			return resp, nil, nil, success, mobileGUIDAssignment.Variant
		} else if err != nil {
			utils.LogErrorf("FetchLyrics: mobile API lookup failed reason=%s status=%d error=%v", failure.ReasonValue(), failure.StatusCodeValue(), err)
		}
	}

	// // // Desktop
	// Currently disabled because Musixmatch is actively poisoning the responses.
	// resp, err, failure, success := fetchLyricsFromDesktopAPI(input)

	// if err == nil && len(resp.Lyrics) > 0 {
	// 	utils.LogInfof("FetchLyrics: lookup succeeded source=desktop_api category=%s", success.CategoryValue())
	// 	return resp, nil, success
	// }
	// if err != nil {
	// 	utils.LogErrorf("FetchLyrics: desktop API lookup failed reason=%s status=%d error=%v", failure.ReasonValue(), failure.StatusCodeValue(), err)
	// }
	failure := utils.NewLookupFailure(
		"lyrics_not_found",
		"all_sources",
		fmt.Errorf("No lyrics found for artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID),
	).WithPhase("all_sources")
	return lyrics.GetLyricsResponse{},
		fmt.Errorf("No lyrics found for artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID),
		failure,
		nil,
		mobileGUIDAssignment.Variant
}
