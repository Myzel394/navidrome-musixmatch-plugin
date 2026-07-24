package utils

import (
	"net/url"
	"strings"

	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

type HTTPResponse struct {
	Body       []byte
	StatusCode int
	Headers    map[string]string
}

func DoMusixmatchWebsiteGetRequest(endpoint string) (*HTTPResponse, error) {
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

	resp, err := host.HTTPSend(host.HTTPRequest{
		Method:            pdk.MethodGet.String(),
		URL:               endpoint,
		NoFollowRedirects: true,
		Headers: map[string]string{
			"Accept":          httpAcceptHeader,
			"Accept-Language": "en",
			"User-Agent":      userAgent,
			"Cookie":          strings.Join(cookies, "; "),
		},
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return &HTTPResponse{
		Body:       resp.Body,
		StatusCode: int(resp.StatusCode),
		Headers:    resp.Headers,
	}, nil
}
