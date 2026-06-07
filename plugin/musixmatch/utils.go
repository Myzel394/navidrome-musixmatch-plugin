package musixmatch

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	bracketsRe   = regexp.MustCompile(`[\(\[\{][^\)\]\}]*[\)\]\}]`)
	dashSuffixRe = regexp.MustCompile(`(?i)\s*-\s*(remaster(ed)?|single version|live|deluxe|edit|mix|version|radio edit|extended).*$`)
	whitespaceRe = regexp.MustCompile(`\s+`)
	nonWordRe    = regexp.MustCompile(`[^\p{L}\p{N}\s-]`)
	multiDashRe  = regexp.MustCompile(`-+`)
)

func normalize(s string) string {
	s = strings.ToLower(s)
	s = stripDiacritics(s)
	s = bracketsRe.ReplaceAllString(s, " ")
	s = dashSuffixRe.ReplaceAllString(s, "")
	s = whitespaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = stripDiacritics(s)
	s = nonWordRe.ReplaceAllString(s, "")
	s = whitespaceRe.ReplaceAllString(s, "-")
	s = multiDashRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func stripDiacritics(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return result
}
