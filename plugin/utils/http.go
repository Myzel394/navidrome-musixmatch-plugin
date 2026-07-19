package utils

import (
	"strings"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func DoGetRequest(endpoint string) ([]byte, error) {
	userAgent := ConfigUserAgent()
	httpAcceptHeader := ConfigSearchHTTPAcceptHeader()

	cookies := make([]string, 0, 2)

	cookie := strings.Join(cookies, "; ")

	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Accept", httpAcceptHeader)
	req.SetHeader("Accept-Language", "en")
	req.SetHeader("User-Agent", userAgent)
	req.SetHeader("Cookie", cookie)

	resp := req.Send()
	if resp.Status() != HTTPStatusOK {
		LogErrorf("HTTP %d received", resp.Status())
		return resp.Body(), &HTTPError{StatusCode: int(resp.Status())}
	}
	return resp.Body(), nil
}
