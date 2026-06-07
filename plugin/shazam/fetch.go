package shazam

import (
	"fmt"
	"net/url"

	"github.com/Myzel394/navidrome-shazam-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func fetchLyricsForTrack(track *Song) (lyrics.GetLyricsResponse, error) {
	// Normalize already stripped brackets, so "Never Gonna Give You Up (2022 Remaster)"
	// becomes "never gonna give you up" → slug: "never-gonna-give-you-up"
	slug := slugify(track.Title)

	endpoint := fmt.Sprintf(utils.ShazamFetchPageURL, track.ID, url.PathEscape(slug))

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		return lyrics.GetLyricsResponse{}, fmt.Errorf("failed to do shazam fetchLyrics request for track ID %s; Error: %v", track.ID, err)
	}

	text, err := extractLyricsFromHTML(string(body))
	if err != nil {
		return lyrics.GetLyricsResponse{}, err
	}

	utils.LogInfof("found lyrics for track ID %s", track.ID)

	return lyrics.GetLyricsResponse{
		Lyrics: []lyrics.LyricsText{
			{
				Text: text,
			},
		},
	}, nil
}
