package musixmatch

import (
	"encoding/json"
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func fetchLyricsFromMobileAPI(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	return fetchLyricsFromMobileAPIWithRetry(input, false)
}

func fetchLyricsFromMobileAPIWithRetry(input lyrics.GetLyricsRequest, retried bool) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	token, err, failure := mobileUserToken()
	if err != nil {
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	var resp macroResponse
	if err := mobileGet("macro.subtitles.get", buildMobileLyricsQuery(input, token), &resp); err != nil {
		failure := utils.NewLookupFailure("mobile_macro_request_failed", "mobile_api", err).WithPhase("mobile_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("mobile API: lyrics response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == utils.HTTPStatusBlocked {
		if !retried {
			mobileInvalidateUserToken()
			utils.LogInfof("mobile API: token invalidated after 401, retrying once")
			return fetchLyricsFromMobileAPIWithRetry(input, true)
		}
		return lyrics.GetLyricsResponse{}, fmt.Errorf("mobile API returned 401 for lyrics request"), utils.NewLookupFailure("mobile_blocked", "mobile_api", fmt.Errorf("mobile API returned 401 for lyrics request")).WithPhase("mobile_lyrics").WithStatusCode(resp.Message.Header.StatusCode), nil
	}
	if resp.Message.Header.StatusCode != utils.HTTPStatusOK {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("mobile API returned status %d for lyrics request", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("mobile_macro_status", "mobile_api", err).WithPhase("mobile_lyrics").WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	var body macroBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(resp.Message.Body)))
		failure := utils.NewLookupFailure("mobile_macro_parse", "mobile_api", err).WithPhase("mobile_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	matcherCall := body.MacroCalls["matcher.track.get"]
	meta, err := parseResponseToTrackMetadata(matcherCall)
	if err != nil {
		utils.LogInfof("mobile API: matched track metadata could not be parsed body_bytes=%d", len(matcherCall.Message.Body))
	}
	if err := validateMatchedIdentity(input, meta, "mobile API"); err != nil {
		return lyrics.GetLyricsResponse{}, nil, nil, nil
	}
	if resp, ok := lyricsFromRichsync(body.MacroCalls["track.richsync.get"]); ok && lyricsResponseAllowed(resp) {
		return resp, nil, nil, utils.NewLookupSuccess("mobile_synced")
	} else if ok {
		utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	if resp, ok := lyricsFromSubtitle(body.MacroCalls["track.subtitles.get"]); ok && lyricsResponseAllowed(resp) {
		return resp, nil, nil, utils.NewLookupSuccess("mobile_synced")
	} else if ok {
		utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	if resp, ok := lyricsFromPlain(body.MacroCalls["track.lyrics.get"]); ok && lyricsResponseAllowed(resp) {
		return resp, nil, nil, utils.NewLookupSuccess("mobile_plain")
	} else if ok {
		utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	err = fmt.Errorf("mobile API did not return lyrics")
	failure = utils.NewLookupFailure("mobile_no_lyrics", "mobile_api", err).WithPhase("mobile_lyrics")
	return lyrics.GetLyricsResponse{}, err, failure, nil
}
