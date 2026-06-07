package musixmatch

import (
	"fmt"
	"strings"
)

func totalToLRC(total float64) string {
	mins := int(total) / 60
	secs := int(total) % 60
	frac := int((total - float64(int(total))) * 100)
	return fmt.Sprintf("[%02d:%02d.%02d]", mins, secs, frac)
}

func buildLRCForTrackStructure(sections []sectionLine) string {
	var b strings.Builder
	for _, section := range sections {
		for _, line := range section.Lines {
			if line.Type == "instrumental" {
				continue
			}
			if line.Text == "" {
				continue
			}
			b.WriteString(totalToLRC(line.Time.Total))
			b.WriteString(line.Text)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func buildLRCForSubtitle(subtitles []subtitleLine) string {
	var b strings.Builder
	for _, line := range subtitles {
		if line.Type == "instrumental" {
			continue
		}
		if line.Text == "" {
			continue
		}
		b.WriteString(totalToLRC(line.Time.Total))
		b.WriteString(line.Text)
		b.WriteByte('\n')
	}
	return b.String()
}
