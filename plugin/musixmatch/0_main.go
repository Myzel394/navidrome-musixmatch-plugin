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
func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, utils.LyricsFormat, error, *utils.LookupFailure, *utils.LookupSuccess, string) {
	mobileGUIDAssignment := chooseMobileGUIDAssignment()
	var lastErr error
	var lastFailure *utils.LookupFailure
	// // // Mobile
	utils.LogInfof("FetchLyrics: lookup started source=mobile_api artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID)
	if resp, format, err, failure, success := fetchLyricsFromMobileAPI(input, mobileGUIDAssignment); err == nil && len(resp.Lyrics) > 0 {
		utils.LogInfof("FetchLyrics: lookup succeeded source=mobile_api category=%s", success.CategoryValue())
		return resp, format, nil, nil, success, mobileGUIDAssignment.Variant
	} else if err != nil {
		lastErr = err
		lastFailure = failure
		utils.LogErrorf("FetchLyrics: lookup failed source=mobile_api reason=%s status=%d error=%v", failure.ReasonValue(), failure.StatusCodeValue(), err)
	}

	// // // Website
	if utils.ConfigUserToken() != "" {
		utils.LogInfof("FetchLyrics: lookup started source=website artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID)
		resp, format, err, failure, success := fetchLyricsViaWebsiteScraping(input)

		if err == nil && len(resp.Lyrics) > 0 {
			utils.LogInfof("FetchLyrics: lookup succeeded source=website category=%s", success.CategoryValue())
			return resp, format, nil, nil, success, mobileGUIDAssignment.Variant
		} else if err != nil {
			lastErr = err
			lastFailure = failure
			utils.LogErrorf("FetchLyrics: lookup failed source=website reason=%s status=%d error=%v", failure.ReasonValue(), failure.StatusCodeValue(), err)
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
	if lastFailure != nil {
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, lastErr, lastFailure, nil, mobileGUIDAssignment.Variant
	}
	err := fmt.Errorf("No lyrics found for artist=%s title=%s album=%s mbz=%s", input.Track.Artist, input.Track.Title, input.Track.Album, input.Track.MBZRecordingID)
	failure := utils.NewLookupFailure("lyrics_not_found", "all_sources", err).WithPhase("all_sources")
	return lyrics.GetLyricsResponse{},
		utils.LyricsFormatUnknown,
		err,
		failure,
		nil,
		mobileGUIDAssignment.Variant
}
