package shazam

import (
	"encoding/json"
	"fmt"
	"strings"
)

/*
jquery: .recordingOf.lyrics.text
*/
type songInfoJSON struct {
	RecordingOf struct {
		Lyrics struct {
			Text string `json:"text"`
		} `json:"lyrics"`
	} `json:"recordingOf"`
}

// extractLyricsFromHTML scrapes lyrics from a Shazam song page.
// Lyrics are embedded in Next.js __next_f.push() script tags as
// escaped JSON with the structure: \"lyricLines\":[{\"content\":\"...\"}]
func extractLyricsFromHTML(html string) (string, error) {
	const startOfScript = `<script type="application/ld+json">`
	const endOfScript = `</script>`

	// Find the script tag containing the lyrics JSON
	startIdx := strings.Index(html, startOfScript)
	if startIdx == -1 {
		return "", fmt.Errorf("could not find start of lyrics script tag")
	}

	endIdx := strings.Index(html[startIdx:], endOfScript)
	if endIdx == -1 {
		return "", fmt.Errorf("could not find end of lyrics script tag")
	}

	scriptContent := html[startIdx+len(startOfScript) : startIdx+endIdx]

	var songInfo songInfoJSON
	if err := json.Unmarshal([]byte(scriptContent), &songInfo); err != nil {
		return "", fmt.Errorf("failed to parse lyrics JSON: %v; %v", err, scriptContent)
	}

	lyricsText := songInfo.RecordingOf.Lyrics.Text

	if lyricsText != "" {
		return lyricsText, nil
	}

	return "", fmt.Errorf("lyrics text not found in JSON")
}
