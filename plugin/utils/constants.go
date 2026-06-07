package utils

const LogPrefix = "navidrome-shazam-plugin: "

const HTTPStatusOK = 200

const (
	DefaultSearchCountry        = "US"
	DefaultSearchLanguage       = "en"
	DefaultSearchLimit          = 10
	DefaultLevenshteinThreshold = 0.85
	DefaultUserAgent            = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1"
	DefaultHTTPAccept           = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
)

const (
	ConfigKeySearchCountry        = "shazam_search_country"
	ConfigKeySearchLanguage       = "shazam_language"
	ConfigKeySearchLimit          = "shazam_search_limit"
	ConfigKeyLevenshteinThreshold = "shazam_search_levenshtein_threshold"
	ConfigKeyUserAgent            = "shazam_user_agent"
	ConfigKeyHTTPAccept           = "shazam_http_accept"
	ConfigKeyCookie               = "shazam_cookie"
)

const (
	ShazamFetchPageURL = "https://www.shazam.com/song/%s/%s"
	ShazamSearchAPIURL = "https://www.shazam.com/services/amapi/v1/catalog/%s/search?term=%s&types=songs&limit=%d"
)
