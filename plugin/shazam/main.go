package shazam

import (
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type Song struct {
	ID     string
	Artist string
	Title  string
}

func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error) {
	track, err := searchForTrack(input)
	if err != nil || track == nil {
		return lyrics.GetLyricsResponse{}, err
	}

	return fetchLyricsForTrack(track)
}
