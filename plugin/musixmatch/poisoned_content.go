package musixmatch

import (
	"regexp"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

var lrcTimestampRe = regexp.MustCompile(`(?m)\[[0-9]{1,2}:[0-9]{2}(?:[.:][0-9]{1,3})?\]`)

func lyricsResponseAllowed(source string, resp lyrics.GetLyricsResponse) bool {
	for _, l := range resp.Lyrics {
		if ok, reason := lyricTextAllowed(l.Text); !ok {
			utils.LogInfof("%s: rejected lyrics reason=%s", source, reason)
			return false
		}
	}
	return len(resp.Lyrics) > 0
}

func lyricTextAllowed(text string) (bool, string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false, "empty"
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "<html") || strings.Contains(lower, "<!doctype html") || strings.Contains(lower, "</html>") {
		return false, "html"
	}
	if strings.HasPrefix(lower, "{") || strings.HasPrefix(lower, "[") && strings.Contains(lower, "\"message\"") {
		return false, "json"
	}
	for _, marker := range []string{"captcha", "unauthorized", "access denied", "forbidden", "error occurred", "login"} {
		if strings.Contains(lower, marker) {
			return false, "auth_or_error"
		}
	}
	norm := whitespaceRe.ReplaceAllString(strings.ToLower(lrcTimestampRe.ReplaceAllString(trimmed, "")), " ")
	if strings.HasPrefix(strings.TrimSpace(norm), "wob gopini den tefe woxica fero gogoh vudob wiya") {
		return false, "generated_pseudo_lyrics"
	}
	return true, ""
}
