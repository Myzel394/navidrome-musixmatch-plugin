package musixmatch

import (
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

// Richsync times each word (JSON, only returned with
// `optional_calls=track.richsync`), subtitles time each line (LRC). Richsync is
// preferred, with subtitles as the fallback. Both end up as line-by-line LRC,
// because `lyricsFromRichsync` only keeps each line's start time and text.
//
// Subtitle:
//
//	[00:01.00]First line of the song
//	[00:04.50]Second line of the song
//
// Richsync, one object per line (`ts`/`te` are the line's start and end, `o` is
// each word's offset from the line's start):
//
//	{
//	  "ts": 1.0,
//	  "te": 4.5,
//	  "x": "First line of the song",
//	  "l": [{ "c": "First", "o": 0.0 }, { "c": "line", "o": 0.4 }]
//	}
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
