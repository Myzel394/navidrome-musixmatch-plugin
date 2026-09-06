package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

type desktopTokenBody struct {
	UserToken string `json:"user_token"`
}

func desktopUserToken() (string, error, *utils.LookupFailure) {
	if token, ok, err := host.CacheGetString(desktopTokenCache); err != nil {
		utils.LogErrorf("desktop API token cache read failed: %v", err)
	} else if ok && token != "" {
		utils.LogInfof("desktop API: token cache hit")
		return token, nil, nil
	}
	utils.LogInfof("desktop API: token cache miss")

	query := url.Values{}
	query.Set("user_language", "en")

	var resp desktopResponse
	err := desktopGet("token.get", query, &resp)
	if err != nil {
		utils.LogErrorf("desktop API: token request failed error=%v", err)
		failure := utils.NewLookupFailure("desktop_token_request_failed", "desktop_api", err).WithPhase("desktop_token")
		return "", err, failure
	}
	utils.LogInfof("desktop API: token response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == utils.HTTPStatusBlocked {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: token response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("desktop API returned 401 while fetching token")
		failure := utils.NewLookupFailure("desktop_token_blocked", "desktop_api", err).WithPhase("desktop_token").WithStatusCode(resp.Message.Header.StatusCode)
		return "", err, failure
	}
	if resp.Message.Header.StatusCode != utils.HTTPStatusOK {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: token response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("desktop API returned status %d while fetching token", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("desktop_token_status", "desktop_api", err).WithPhase("desktop_token").WithStatusCode(resp.Message.Header.StatusCode)
		return "", err, failure
	}

	var body desktopTokenBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		utils.LogErrorf("desktop API: failed to parse desktop API token body body_bytes=%d error=%v", len(resp.Message.Body), err)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: token response body=%s", string(resp.Message.Body)))
		failure := utils.NewLookupFailure("desktop_token_parse", "desktop_api", err).WithPhase("desktop_token")
		return "", err, failure
	}
	if body.UserToken == "" {
		utils.LogInfof("desktop API: parsed token response token_present=false")
		err := fmt.Errorf("desktop API returned empty token")
		failure := utils.NewLookupFailure("desktop_token_empty", "desktop_api", err).WithPhase("desktop_token")
		return "", err, failure
	}
	utils.LogInfof("desktop API: parsed token response token_present=true")

	if err := host.CacheSetString(desktopTokenCache, body.UserToken, int64(desktopTokenTTL/time.Second)); err != nil {
		utils.LogErrorf("desktop API token cache write failed: %v", err)
	}

	return body.UserToken, nil, nil
}

func desktopInvalidateUserToken() {
	if err := host.CacheRemove(desktopTokenCache); err != nil {
		utils.LogErrorf("desktop API token cache remove failed: %v", err)
	}
}
