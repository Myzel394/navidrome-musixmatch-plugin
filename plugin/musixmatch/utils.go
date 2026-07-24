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

func runeLevenshteinSimilarity(a, b string) float64 {
	ar, br := []rune(a), []rune(b)
	maxLen := max(len(ar), len(br))
	if maxLen == 0 {
		return 1
	}
	distance := runeLevenshteinDistance(ar, br)
	return 1 - float64(distance)/float64(maxLen)
}

func runeLevenshteinDistance(a, b []rune) int {
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur := make([]int, len(b)+1)
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			cur[j] = min(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(b)]
}

func validStrictSimilarity(a, b string) bool {
	a = normalize(a)
	b = normalize(b)
	if a == "" || b == "" {
		return false
	}
	return runeLevenshteinSimilarity(a, b) > 0.9
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
