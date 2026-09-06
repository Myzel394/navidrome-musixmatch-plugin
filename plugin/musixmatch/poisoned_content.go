package musixmatch

import (
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"

	"regexp"
	"strings"
)

func lyricsResponseAllowed(resp lyrics.GetLyricsResponse) bool {
	for _, l := range resp.Lyrics {
		if detectPoisonedLyrics(l.Text) {
			return false
		}
	}
	return len(resp.Lyrics) > 0
}

var poisonedLyricsLRCTimestampRe = regexp.MustCompile(`(?m)\[[0-9]{1,2}:[0-9]{2}(?:[.:][0-9]{1,3})?\]`)

func detectPoisonedLyrics(text string) bool {
	normalized := poisonedLyricsLRCTimestampRe.ReplaceAllString(text, "")
	normalized = strings.ToLower(normalized)
	normalized = strings.Join(strings.Fields(normalized), " ")
	return strings.HasPrefix(normalized, "wob gopini den tefe woxica fero gogoh vudob wiya")
}
