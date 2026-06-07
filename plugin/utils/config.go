package utils

import (
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func getConfigString(key, def string) string {
	v, ok := pdk.GetConfig(key)
	if !ok || v == "" {
		return def
	}
	return v
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
