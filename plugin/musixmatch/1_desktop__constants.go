package musixmatch

import (
	"encoding/json"
)

const (
	desktopAppID       = "mac-ios-v2.0"
	desktopAppVersion  = "10.1.1"
	desktopClientID    = "Musixmatch/2025120901 CFNetwork/3860.300.31 Darwin/25.2.0"
	desktopAPISuccess  = 200
	desktopAPIBlocked  = 401
	desktopFallbackErr = "desktop API did not return lyrics"
)

type desktopResponse struct {
	Message struct {
		Header desktopHeader   `json:"header"`
		Body   json.RawMessage `json:"body"`
	} `json:"message"`
}

type desktopHeader struct {
	StatusCode int `json:"status_code"`
}

type desktopMacroBody struct {
	MacroCalls map[string]desktopResponse `json:"macro_calls"`
}

type desktopSubtitleBody struct {
	SubtitleList []struct {
		Subtitle struct {
			Body     string `json:"subtitle_body"`
			Language string `json:"subtitle_language"`
		} `json:"subtitle"`
	} `json:"subtitle_list"`
}

type desktopRichsyncBody struct {
	Richsync struct {
		Body     string `json:"richsync_body"`
		Language string `json:"richssync_language"`
	} `json:"richsync"`
}

type desktopLyricsBody struct {
	Lyrics struct {
		Body     string `json:"lyrics_body"`
		Language string `json:"lyrics_language"`
	} `json:"lyrics"`
}

type desktopTrackBody struct {
	Track struct {
		TrackName  string `json:"track_name"`
		ArtistName string `json:"artist_name"`
		AlbumName  string `json:"album_name"`
	} `json:"track"`
}

type desktopRichsyncLine struct {
	Timestamp float64 `json:"ts"`
	Text      string  `json:"x"`
}
