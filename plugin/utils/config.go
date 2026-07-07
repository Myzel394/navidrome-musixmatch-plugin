package utils

import (
	"strconv"

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

func ConfigCaptchaID() string {
	v, _ := pdk.GetConfig(ConfigKeyCaptchaID)
	return v
}

func ConfigShareErrors() bool {
	return getConfigBool(ConfigKeyShareErrors, false)
}

func ConfigShareMetrics() bool {
	return getConfigBool(ConfigKeyShareMetrics, false)
}
