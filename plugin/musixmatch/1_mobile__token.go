package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
)

func mobileUserToken() (string, error, *utils.LookupFailure) {
	if token, ok, err := host.CacheGetString(mobileTokenCache); err == nil && ok && token != "" {
		utils.LogInfof("mobile API: token cache hit")
		return token, nil, nil
	}
	query := url.Values{}
	query.Set("user_language", "en")
	var resp desktopResponse
	if err := mobileGet("token.get", query, &resp); err != nil {
		failure := utils.NewLookupFailure("mobile_token_request_failed", "mobile_api", err).WithPhase("mobile_token")
		return "", err, failure
	}
	if resp.Message.Header.StatusCode != utils.HTTPStatusOK {
		err := fmt.Errorf("mobile API returned status %d while fetching token", resp.Message.Header.StatusCode)
		return "", err, utils.NewLookupFailure("mobile_token_status", "mobile_api", err).WithPhase("mobile_token").WithStatusCode(resp.Message.Header.StatusCode)
	}
	var body desktopTokenBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		return "", err, utils.NewLookupFailure("mobile_token_parse", "mobile_api", err).WithPhase("mobile_token")
	}
	if body.UserToken == "" {
		err := fmt.Errorf("mobile API returned empty token")
		return "", err, utils.NewLookupFailure("mobile_token_empty", "mobile_api", err).WithPhase("mobile_token")
	}
	_ = host.CacheSetString(mobileTokenCache, body.UserToken, int64(mobileTokenTTL/time.Second))
	return body.UserToken, nil, nil
}

func mobileInvalidateUserToken() { _ = host.CacheRemove(mobileTokenCache) }
