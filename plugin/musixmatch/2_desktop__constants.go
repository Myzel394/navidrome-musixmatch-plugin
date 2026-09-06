package musixmatch

import "time"

const (
	desktopAppID       = "web-desktop-app-v1.0"
	desktopUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36"
	desktopFallbackErr = "desktop API did not return lyrics"

	desktopTokenCache = "musixmatch_desktop_user_token"
	desktopTokenTTL   = 10 * time.Minute
)
