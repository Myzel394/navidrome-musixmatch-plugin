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

func ConfigSearchCountry() string {
	return getConfigString(ConfigKeySearchCountry, DefaultSearchCountry)
}

func ConfigSearchLanguage() string {
	return getConfigString(ConfigKeySearchLanguage, DefaultSearchLanguage)
}

func ConfigShazamCookie() string {
	v, _ := pdk.GetConfig(ConfigKeyCookie)
	return v
}

func ConfigSearchHTTPAcceptHeader() string {
	return getConfigString(ConfigKeyHTTPAccept, DefaultHTTPAccept)
}

func ConfigUserAgent() string {
	return getConfigString(ConfigKeyUserAgent, DefaultUserAgent)
}

func ConfigSearchLimit() int {
	v, ok := pdk.GetConfig(ConfigKeySearchLimit)
	if !ok || v == "" {
		return DefaultSearchLimit
	}

	limit, err := strconv.Atoi(v)
	if err != nil {
		LogErrorf("failed to parse %s config value, defaulting to %d", ConfigKeySearchLimit, DefaultSearchLimit)
		return DefaultSearchLimit
	}

	return limit
}

func ConfigSearchLevenshteinThreshold() float64 {
	v, ok := pdk.GetConfig(ConfigKeyLevenshteinThreshold)
	if !ok || v == "" {
		return DefaultLevenshteinThreshold
	}

	threshold, err := strconv.ParseFloat(v, 64)
	if err != nil {
		LogErrorf("failed to parse %s config value, defaulting to %v", ConfigKeyLevenshteinThreshold, DefaultLevenshteinThreshold)
		return DefaultLevenshteinThreshold
	}

	return threshold
}
