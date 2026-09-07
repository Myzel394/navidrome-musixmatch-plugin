package lyrics_writer

import "github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"

func getExtensionForLyricsFormat(format utils.LyricsFormat) string {
	switch format {
	case utils.LyricsFormatSyncedLRC:
		return ".lrc"
	case utils.LyricsFormatPlain:
		return ".txt"
	default:
		return ".txt"
	}
}
