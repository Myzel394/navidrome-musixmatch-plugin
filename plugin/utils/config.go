package utils

import (
	"strconv"
	"strings"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func getConfigString(key, def string) string {
	v, ok := pdk.GetConfig(key)
	if !ok || v == "" {
		return def
	}
	return v
}

func getConfigBool(key string, def bool) bool {
	v, ok := pdk.GetConfig(key)
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getConfigInt(key string, def int) int {
	v, ok := pdk.GetConfig(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func ConfigUserToken() string {
	v, _ := pdk.GetConfig(ConfigKeyUserToken)
	return v
}

func ConfigUserAgent() string {
	return getConfigString(ConfigKeyUserAgent, DefaultUserAgent)
}

func ConfigSearchHTTPAcceptHeader() string {
	return getConfigString(ConfigKeyHTTPAccept, DefaultHTTPAccept)
}

func ConfigMobileUserAgent() string {
	return getConfigString(ConfigKeyMobileUserAgent, DefaultMobileUserAgent)
}

func ConfigMobileAppVersion() string {
	return getConfigString(ConfigKeyMobileAppVersion, DefaultMobileAppVersion)
}

func ConfigMobileAppID() string {
	return getConfigString(ConfigKeyMobileAppID, DefaultMobileAppID)
}

func ConfigCaptchaID() string {
	v, _ := pdk.GetConfig(ConfigKeyCaptchaID)
	return v
}

func ConfigMaxWebsiteCandidateAttempts() int {
	attempts := getConfigInt(ConfigKeyMaxWebsiteCandidateAttempts, DefaultMaxWebsiteCandidateAttempts)
	if attempts < MinWebsiteCandidateAttempts || attempts > MaxWebsiteCandidateAttempts {
		return DefaultMaxWebsiteCandidateAttempts
	}
	return attempts
}

func ConfigShareErrors() bool {
	return getConfigBool(ConfigKeyShareErrors, true)
}

func ConfigShareMetrics() bool {
	return getConfigBool(ConfigKeyShareMetrics, true)
}
