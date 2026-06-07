package utils

import (
	"fmt"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func DoGetRequest(endpoint string) ([]byte, error) {
	userAgent := ConfigUserAgent()
	httpAcceptHeader := ConfigSearchHTTPAcceptHeader()
	musixmatchCookie := ConfigUserToken()

	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Accept", httpAcceptHeader)
	req.SetHeader("Accept-Language", "en")
	req.SetHeader("User-Agent", userAgent)
	req.SetHeader("Cookie", "musixmatchUserToken="+musixmatchCookie)

	resp := req.Send()
	if resp.Status() != HTTPStatusOK {
		LogErrorf("HTTP %d from %s", resp.Status(), endpoint)
		return resp.Body(), fmt.Errorf("error code %d returned from Musixmatch for endpoint %s", resp.Status(), endpoint)
	}
	return resp.Body(), nil
}
