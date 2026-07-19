package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
)

func detectWebsiteGate(phase string, resp *utils.HTTPResponse) (error, *utils.LookupFailure) {
	if resp == nil {
		return nil, nil
	}

	status := resp.StatusCode
	if location := findHeader(resp.Headers, "Location"); location != "" && status >= 300 && status < 400 {
		if isAuthRedirect(location) {
			return websiteGateFailure(phase, "auth_invalid", status, "auth_redirect")
		}
		if isCaptchaRedirect(location) {
			return websiteGateFailure(phase, "captcha_required", status, "captcha_redirect")
		}
	}

	if isHTMLProbablyAuthPage(resp.Body) {
		return websiteGateFailure(phase, "auth_invalid", status, "auth_html")
	}
	if isHTMLProbablyCaptchaPage(resp.Body) {
		return websiteGateFailure(phase, "captcha_required", status, "captcha_html_signature")
	}

	return nil, nil
}

func websiteGateFailure(phase, reason string, status int, signal string) (error, *utils.LookupFailure) {
	utils.LogErrorf("website gate: detected phase=%s reason=%s status=%d signal=%s", phase, reason, status, signal)
	message := "musixmatch website gate detected: " + reason
	switch reason {
	case "auth_invalid":
		message = "musixmatch authentication is invalid; refresh musixmatch_user_token in plugin settings and retry"
	case "captcha_required":
		message = "musixmatch CAPTCHA required; solve the CAPTCHA in a browser, copy captcha_id into plugin settings, and retry"
	}
	err := fmt.Errorf("%s", message)
	websitePhase := websiteGateLookupPhase(phase)
	return err, utils.NewLookupFailure(reason, "website", err).WithPhase(websitePhase).WithStatusCode(status)
}

func websiteGateLookupPhase(phase string) string {
	switch phase {
	case "search":
		return "website_search"
	case "lyrics_page":
		return "website_lyrics"
	default:
		return "website_lyrics"
	}
}

func findHeader(headers map[string]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func isAuthRedirect(location string) bool {
	u, err := url.Parse(location)
	return err == nil && strings.EqualFold(u.Hostname(), "auth.musixmatch.com")
}

func isCaptchaRedirect(location string) bool {
	u, err := url.Parse(location)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	path := strings.ToLower(u.Path)
	return strings.HasPrefix(path, "/captcha") &&
		(host == "" || host == "www.musixmatch.com" || host == "apic.musixmatch.com" || host == "apic-appmobile.musixmatch.com")
}

// Check if the HTML body is probably an authentication page
// Not a 100% guarantee
func isHTMLProbablyAuthPage(body []byte) bool {
	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		return false
	}
	var data map[string]any
	if err := json.Unmarshal(matches[1], &data); err != nil {
		return false
	}
	props, ok := data["props"].(map[string]any)
	if !ok {
		return false
	}
	pageProps, ok := props["pageProps"].(map[string]any)
	if !ok {
		return false
	}
	auth, ok := pageProps["auth"].(map[string]any)
	if !ok {
		return false
	}
	query, ok := data["query"].(map[string]any)
	if !ok {
		return false
	}
	hasAuthCookie, ok := auth["hasAuthCookie"].(bool)
	if !ok {
		return false
	}
	isLoggedIn, ok := pageProps["isLoggedIn"].(bool)
	if !ok {
		return false
	}
	return data["page"] == "/" &&
		auth["appId"] == "mxm-account-v1.0" &&
		hasAuthCookie == false &&
		isLoggedIn == false &&
		query["app_id"] == "mxm-com-v1.0" &&
		query["force_app_redirect"] == "true"
}

// Check if the HTML body is probably a CAPTCHA page
// Not a 100% guarantee
func isHTMLProbablyCaptchaPage(body []byte) bool {
	lower := strings.ToLower(string(body))
	provider := strings.Contains(lower, "www.google.com/recaptcha") ||
		strings.Contains(lower, "g-recaptcha") ||
		strings.Contains(lower, "hcaptcha.com") ||
		strings.Contains(lower, "h-captcha") ||
		strings.Contains(lower, "challenges.cloudflare.com/turnstile") ||
		strings.Contains(lower, "cf-turnstile")
	challenge := strings.Contains(lower, "data-sitekey")
	form := strings.Contains(lower, "mxm-verify-form") || strings.Contains(lower, "challenge-form")
	return provider && challenge && form
}
