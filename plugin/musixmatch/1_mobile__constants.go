package musixmatch

import "time"

const (
	mobileAppID      = "mac-ios-v2.0"
	mobileAppVersion = "10.1.1"
	mobileUserAgent  = "Musixmatch/2025120901 CFNetwork/3860.300.31 Darwin/25.2.0"

	mobileTokenCache = "musixmatch_mobile_user_token"
	mobileTokenTTL   = 60 * time.Second
)
