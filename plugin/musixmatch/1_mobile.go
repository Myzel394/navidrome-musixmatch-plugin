package musixmatch

import (
	"encoding/json"
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func fetchLyricsFromMobileAPI(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	token, err, failure := mobileUserToken()
	if err != nil {
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	var resp desktopResponse
	if err := mobileGet("macro.subtitles.get", mobileLyricsQuery(input, token), &resp); err != nil {
		failure := utils.NewLookupFailure("mobile_macro_request_failed", "mobile_api", err).WithPhase("mobile_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("mobile API: lyrics response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == statusCodeAPIBlocked {
		utils.LogInfof("mobile API: token invalidated after 401, retrying once")
		return lyrics.GetLyricsResponse{}, fmt.Errorf("mobile API returned 401 for lyrics request"), utils.NewLookupFailure("mobile_blocked", "mobile_api", fmt.Errorf("mobile API returned 401 for lyrics request")).WithPhase("mobile_lyrics").WithStatusCode(resp.Message.Header.StatusCode), nil
	}
	if resp.Message.Header.StatusCode != statusCodePISuccess {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("mobile API returned status %d for lyrics request", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("mobile_macro_status", "mobile_api", err).WithPhase("mobile_lyrics").WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	var body desktopMacroBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(resp.Message.Body)))
		failure := utils.NewLookupFailure("mobile_macro_parse", "mobile_api", err).WithPhase("mobile_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	if err := validateMatchedIdentity(input, desktopMatchedMetadata(body.MacroCalls["matcher.track.get"]), "mobile API"); err != nil {
		return lyrics.GetLyricsResponse{}, nil, nil, nil
	}
	if resp, ok := lyricsFromDesktopRichsync(body.MacroCalls["track.richsync.get"]); ok && lyricsResponseAllowed("mobile_api", resp) {
		return resp, nil, nil, utils.NewLookupSuccess("mobile_synced")
	}
	if resp, ok := lyricsFromDesktopSubtitle(body.MacroCalls["track.subtitles.get"]); ok && lyricsResponseAllowed("mobile_api", resp) {
		return resp, nil, nil, utils.NewLookupSuccess("mobile_synced")
	}
	if resp, ok := lyricsFromDesktopPlain(body.MacroCalls["track.lyrics.get"]); ok && lyricsResponseAllowed("mobile_api", resp) {
		return resp, nil, nil, utils.NewLookupSuccess("mobile_plain")
	}
	err = fmt.Errorf("mobile API did not return lyrics")
	failure = utils.NewLookupFailure("mobile_no_lyrics", "mobile_api", err).WithPhase("mobile_lyrics")
	return lyrics.GetLyricsResponse{}, err, failure, nil
}
