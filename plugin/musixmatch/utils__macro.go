package musixmatch

import (
	"encoding/json"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
)

type macroResponse struct {
	Message struct {
		Header macroHeader     `json:"header"`
		Body   json.RawMessage `json:"body"`
	} `json:"message"`
}

type macroHeader struct {
	StatusCode int `json:"status_code"`
}

type macroBody struct {
	MacroCalls map[string]macroResponse `json:"macro_calls"`
}

type macroSubtitleBody struct {
	SubtitleList []struct {
		Subtitle struct {
			Body     string `json:"subtitle_body"`
			Language string `json:"subtitle_language"`
		} `json:"subtitle"`
	} `json:"subtitle_list"`
}

type macroRichsyncBody struct {
	Richsync struct {
		Body     string `json:"richsync_body"`
		Language string `json:"richssync_language"`
	} `json:"richsync"`
}

type macroLyricsBody struct {
	Lyrics struct {
		Body     string `json:"lyrics_body"`
		Language string `json:"lyrics_language"`
	} `json:"lyrics"`
}

type macroTrackBody struct {
	Track struct {
		TrackName  string `json:"track_name"`
		ArtistName string `json:"artist_name"`
		AlbumName  string `json:"album_name"`
	} `json:"track"`
}

type macroRichsyncLine struct {
	Timestamp float64 `json:"ts"`
	Text      string  `json:"x"`
}

// Parse response to common `trackMetadata` struct
func parseResponseToTrackMetadata(call macroResponse) (trackMetadata, error) {
	if call.Message.Header.StatusCode != utils.HTTPStatusOK || len(call.Message.Body) == 0 {
		return trackMetadata{}, nil
	}
	var body macroTrackBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		return trackMetadata{}, err
	}
	return trackMetadata{Artist: body.Track.ArtistName, Title: body.Track.TrackName, Album: body.Track.AlbumName}, nil
}
