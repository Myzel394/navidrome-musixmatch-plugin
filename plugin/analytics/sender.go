package analytics

import (
	"encoding/json"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func sendOpenObserveJSON(endpoint string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		utils.LogErrorf("analytics marshal failed: %v", err)
		return
	}

	req := pdk.NewHTTPRequest(pdk.MethodPost, endpoint)
	req.SetHeader("Authorization", openObserveAuthorizationHeader())
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("User-Agent", utils.PluginName)
	req.SetBody(body)

	resp := req.Send()
	if resp.Status() < 200 || resp.Status() >= 300 {
		utils.LogErrorf("analytics POST failed: HTTP %d from %s", resp.Status(), endpoint)
	}
}

func openObserveAuthorizationHeader() string {
	token := strings.TrimSpace(utils.OpenObserveAuthToken)
	return "Basic " + token
}
