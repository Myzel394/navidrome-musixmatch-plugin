package utils

const PluginName = "navidrome-musixmatch-plugin"

const HTTPStatusOK = 200

const (
	DefaultUserAgent                   = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1"
	DefaultHTTPAccept                  = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"
	DefaultMaxWebsiteCandidateAttempts = 3
	MinWebsiteCandidateAttempts        = 1
	MaxWebsiteCandidateAttempts        = 10
)

const (
	ConfigKeyUserToken                   = "musixmatch_user_token"
	ConfigKeyUserAgent                   = "musixmatch_user_agent"
	ConfigKeyHTTPAccept                  = "musixmatch_http_accept"
	ConfigKeyCaptchaID                   = "musixmatch_captcha_id"
	ConfigKeyMaxWebsiteCandidateAttempts = "max_website_candidate_attempts"
	ConfigKeyShareErrors                 = "analytics_share_errors"
	ConfigKeyShareMetrics                = "analytics_share_metrics"
)

const (
	MusixmatchSearchPageURL = "https://www.musixmatch.com/search?query=%s"
	MusixmatchFetchPageURL  = "https://www.musixmatch.com/lyrics/%s"
	MusixmatchDesktopAPIURL = "https://apic-desktop.musixmatch.com/ws/1.1/%s"
	MusixmatchMobileAPIURL  = "https://apic-appmobile.musixmatch.com/ws/1.1/%s"
)

const (
	OpenObserveBaseURL    = "https://bugs.myzel394.app"
	OpenObserveOrg        = "default"
	OpenObserveTraceURL   = OpenObserveBaseURL + "/api/" + OpenObserveOrg + "/v1/traces"
	OpenObserveMetricsURL = OpenObserveBaseURL + "/api/" + OpenObserveOrg + "/ingest/metrics/_json"
)

const (
	statusCodePISuccess  = 200
	statusCodeAPIBlocked = 401
)

var (
	OpenObserveAttributePluginSource string
	OpenObserveAttributeVersion      string
	OpenObserveAuthToken             string
)
