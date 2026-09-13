package utils

import (
	"maps"

	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func DoMobileGetRequest(endpoint string, extraHeaders map[string]string) ([]byte, error) {
	appVersion := ConfigMobileAppVersion()
	userAgent := ConfigMobileUserAgent()

	headers := map[string]string{
		"Host":              "apic-appmobile.musixmatch.com",
		"authority":         "apic-appmobile.musixmatch.com",
		"x-mxm-app-version": appVersion,
		"X-User-Agent":      userAgent,
		"Accept-Language":   "en-US,en;q=0.9",
		"Connection":        "keep-alive",
		"Accept":            "application/json",
	}

	maps.Copy(headers, extraHeaders)

	resp, err := host.HTTPSend(host.HTTPRequest{
		Method:            pdk.MethodGet.String(),
		URL:               endpoint,
		NoFollowRedirects: true,
		Headers:           headers,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return resp.Body, nil
}
