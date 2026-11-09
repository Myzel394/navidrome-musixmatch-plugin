package musixmatch

import (
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func getSyncedLyricsFromMacro(macro macroBody) (lyrics.GetLyricsResponse, bool) {
	if resp, ok := lyricsFromRichsync(macro.MacroCalls["track.richsync.get"]); ok && !isPoisonedLyrics(resp) {
		return resp, true
	}
	if resp, ok := lyricsFromSubtitle(macro.MacroCalls["track.subtitles.get"]); ok && !isPoisonedLyrics(resp) {
		return resp, true
	}
	return lyrics.GetLyricsResponse{}, false
}

func findBestPlainLyrics(macros []macroBody) (lyrics.GetLyricsResponse, utils.LyricsFormat, error, *utils.LookupFailure, *utils.LookupSuccess) {
	for _, macro := range macros {
		if resp, ok := lyricsFromPlain(macro.MacroCalls["track.lyrics.get"]); ok && !isPoisonedLyrics(resp) {
			return resp, utils.LyricsFormatPlain, nil, nil, utils.NewLookupSuccess("mobile_plain")
		}
	}

	// Damn it, that did not work.
	err := fmt.Errorf("mobile API did not return lyrics")
	failure := utils.NewLookupFailure("mobile_no_lyrics", "mobile_api", err).WithPhase("mobile_lyrics")
	return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
}

func isPoisonedLyrics(resp lyrics.GetLyricsResponse) bool {
	if lyricsResponseAllowed(resp) {
		return false
	}
	utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	return true
}
