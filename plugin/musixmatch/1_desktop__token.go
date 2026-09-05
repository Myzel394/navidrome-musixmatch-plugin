package musixmatch

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

const (
	desktopTokenCache = "musixmatch_desktop_user_token"
	desktopGUIDCache  = "musixmatch_desktop_install_guid"
	desktopTokenTTL   = 10 * time.Minute
)

type desktopTokenBody struct {
	UserToken string `json:"user_token"`
}

func desktopUserToken() (string, string, error, *utils.LookupFailure) {
	token, ok, err := host.CacheGetString(desktopTokenCache)
	if err != nil {
		utils.LogErrorf("desktop API token cache read failed: %v", err)
	}
	guid, guidOK, guidErr := host.CacheGetString(desktopGUIDCache)
	if guidErr != nil {
		utils.LogErrorf("desktop API guid cache read failed: %v", guidErr)
	}
	if ok && token != "" && guidOK && guid != "" {
		utils.LogInfof("desktop API: token cache hit")
		return token, guid, nil, nil
	}
	utils.LogInfof("desktop API: token cache miss")

	if guid == "" {
		guid = newDesktopInstallGUID()
		if guid != "" {
			if err := host.CacheSetString(desktopGUIDCache, guid, int64(desktopTokenTTL/time.Second)); err != nil {
				utils.LogErrorf("desktop API guid cache write failed: %v", err)
			}
		}
	}

	query := url.Values{}
	query.Set("user_language", "en")
	query.Set("guid", guid)

	var resp desktopResponse
	err = desktopGet("token.get", query, guid, &resp)
	if err != nil {
		utils.LogErrorf("desktop API: token request failed error=%v", err)
		failure := utils.NewLookupFailure("desktop_token_request_failed", "desktop_api", err).WithPhase("desktop_token")
		return "", guid, err, failure
	}
	utils.LogInfof("desktop API: token response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == desktopAPIBlocked {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: token response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("desktop API returned 401 while fetching token")
		failure := utils.NewLookupFailure("desktop_token_blocked", "desktop_api", err).WithPhase("desktop_token").WithStatusCode(resp.Message.Header.StatusCode)
		return "", guid, err, failure
	}
	if resp.Message.Header.StatusCode != desktopAPISuccess {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: token response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("desktop API returned status %d while fetching token", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("desktop_token_status", "desktop_api", err).WithPhase("desktop_token").WithStatusCode(resp.Message.Header.StatusCode)
		return "", guid, err, failure
	}

	var body desktopTokenBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		utils.LogErrorf("desktop API: failed to parse desktop API token body body_bytes=%d error=%v", len(resp.Message.Body), err)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: token response body=%s", string(resp.Message.Body)))
		failure := utils.NewLookupFailure("desktop_token_parse", "desktop_api", err).WithPhase("desktop_token")
		return "", guid, err, failure
	}
	if body.UserToken == "" {
		utils.LogInfof("desktop API: parsed token response token_present=false")
		err := fmt.Errorf("desktop API returned empty token")
		failure := utils.NewLookupFailure("desktop_token_empty", "desktop_api", err).WithPhase("desktop_token")
		return "", guid, err, failure
	}
	utils.LogInfof("desktop API: parsed token response token_present=true")

	if err := host.CacheSetString(desktopTokenCache, body.UserToken, int64(desktopTokenTTL/time.Second)); err != nil {
		utils.LogErrorf("desktop API token cache write failed: %v", err)
	}

	return body.UserToken, guid, nil, nil
}

func newDesktopInstallGUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func desktopInvalidateUserToken() {
	if err := host.CacheRemove(desktopTokenCache); err != nil {
		utils.LogErrorf("desktop API token cache remove failed: %v", err)
	}
	if err := host.CacheRemove(desktopGUIDCache); err != nil {
		utils.LogErrorf("desktop API guid cache remove failed: %v", err)
	}
}
