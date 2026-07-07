package utils

import (
	"net/url"
	"strings"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func DoGetRequest(endpoint string) ([]byte, error) {
	userAgent := ConfigUserAgent()
	httpAcceptHeader := ConfigSearchHTTPAcceptHeader()
	musixmatchCookie := ConfigUserToken()
	musixmatchConfigID := ConfigCaptchaID()

	cookies := make([]string, 0, 2)

	if musixmatchCookie != "" {
		cookies = append(cookies, "musixmatchUserToken="+url.QueryEscape(musixmatchCookie))

		if musixmatchConfigID != "" {
			cookies = append(cookies, "captcha_id="+url.QueryEscape(musixmatchConfigID))
		}
	}

	cookie := strings.Join(cookies, "; ")

	req := pdk.NewHTTPRequest(pdk.MethodGet, endpoint)
	req.SetHeader("Accept", httpAcceptHeader)
	req.SetHeader("Accept-Language", "en")
	req.SetHeader("User-Agent", userAgent)
	req.SetHeader("Cookie", cookie)

	resp := req.Send()
	if resp.Status() != HTTPStatusOK {
		LogErrorf("HTTP %d from %s", resp.Status(), endpoint)
		return resp.Body(), &HTTPError{StatusCode: int(resp.Status()), Endpoint: endpoint}
	}
	return resp.Body(), nil
}
