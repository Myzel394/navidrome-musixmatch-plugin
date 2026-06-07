package musixmatch

import (
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type Song struct {
	Artist              string
	Title               string
	CommontrackVanityID string
}

func FetchLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error) {
	utils.LogInfof("FetchLyrics: artist='%s' title='%s'", input.Track.Artist, input.Track.Title)

	track, err := searchForTrack(input)
	if err != nil {
		utils.LogErrorf("FetchLyrics search error: %v", err)
		return lyrics.GetLyricsResponse{}, err
	}
	if track == nil {
		utils.LogInfof("FetchLyrics: no match found for '%s' - '%s'", input.Track.Artist, input.Track.Title)
		return lyrics.GetLyricsResponse{}, nil
	}

	utils.LogInfof("FetchLyrics: matched '%s' by '%s'", track.Title, track.Artist)
	return fetchLyricsForTrack(track)
}
