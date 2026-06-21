package musixmatch

import "encoding/json"

const (
	desktopAppID       = "web-desktop-app-v1.0"
	desktopTokenCache  = "musixmatch_desktop_user_token"
	desktopTokenTTL    = 10 * 60 // 10 minutes
	desktopUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36"
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

type desktopTokenBody struct {
	UserToken string `json:"user_token"`
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

type desktopRichsyncLine struct {
	Timestamp float64 `json:"ts"`
	Text      string  `json:"x"`
}
