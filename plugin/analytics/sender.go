package analytics

import (
	"encoding/json"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func sendOpenObserveJSON(endpoint string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		utils.LogErrorf("analytics marshal failed: %v", err)
		return
	}

	_, err = host.HTTPSend(host.HTTPRequest{
		Method:            pdk.MethodGet.String(),
		Body: body,
		URL:               endpoint,
		NoFollowRedirects: true,
		Headers: map[string]string{
			"Authorization": openObserveAuthorizationHeader(),
			"Content-Type":  "application/json",
			"User-Agent":    utils.PluginName,
		},
	})

	if err != nil {
		utils.LogErrorf("analytics POST failed: %v", err)
	}
}

func openObserveAuthorizationHeader() string {
	token := strings.TrimSpace(utils.OpenObserveAuthToken)
	return "Basic " + token
}
