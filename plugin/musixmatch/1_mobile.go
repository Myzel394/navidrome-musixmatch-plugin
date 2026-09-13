package musixmatch

import (
	"encoding/json"
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func fetchLyricsFromMobileAPI(input lyrics.GetLyricsRequest, assignment mobileGUIDAssignment) (lyrics.GetLyricsResponse, utils.LyricsFormat, error, *utils.LookupFailure, *utils.LookupSuccess) {
	token, err, failure := mobileUserToken(assignment)
	if err != nil {
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
	}

	// Fetch lyrics from mobile API
	response, err := mobileGet("macro.subtitles.get", buildMobileLyricsQuery(input, token), assignment)
	if err != nil {
		failure := utils.NewLookupFailure("mobile_macro_request_failed", "mobile_api", err).WithPhase("mobile_lyrics")
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
	}

	// Check status
	utils.LogInfof("mobile API: lyrics response received status=%d body_bytes=%d", response.Message.Header.StatusCode, len(response.Message.Body))
	if response.Message.Header.StatusCode == utils.HTTPStatusBlocked {
		err := fmt.Errorf("mobile API returned 401 for lyrics request")
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, utils.NewLookupFailure("mobile_blocked", "mobile_api", err).WithPhase("mobile_lyrics").WithStatusCode(response.Message.Header.StatusCode), nil
	}
	if response.Message.Header.StatusCode != utils.HTTPStatusOK {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(response.Message.Body)))
		err := fmt.Errorf("mobile API returned status %d for lyrics request", response.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("mobile_macro_status", "mobile_api", err).WithPhase("mobile_lyrics").WithStatusCode(response.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
	}

	// Parse
	var body macroBody
	if err := json.Unmarshal(response.Message.Body, &body); err != nil {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("mobile API: lyrics response body=%s", string(response.Message.Body)))
		failure := utils.NewLookupFailure("mobile_macro_parse", "mobile_api", err).WithPhase("mobile_lyrics")
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
	}
	matcherCall := body.MacroCalls["matcher.track.get"]
	meta, err := parseResponseToTrackMetadata(matcherCall)
	if err != nil {
		utils.LogInfof("mobile API: matched track metadata could not be parsed body_bytes=%d", len(matcherCall.Message.Body))
	}

	// Check if response matches any lyrics type
	if err := validateMatchedIdentity(input, meta, "mobile API"); err != nil {
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, nil, nil, nil
	}
	if resp, ok := lyricsFromRichsync(body.MacroCalls["track.richsync.get"]); ok && lyricsResponseAllowed(resp) {
		return resp, utils.LyricsFormatSyncedLRC, nil, nil, utils.NewLookupSuccess("mobile_synced")
	} else if ok {
		utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	if resp, ok := lyricsFromSubtitle(body.MacroCalls["track.subtitles.get"]); ok && lyricsResponseAllowed(resp) {
		return resp, utils.LyricsFormatSyncedLRC, nil, nil, utils.NewLookupSuccess("mobile_synced")
	} else if ok {
		utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	if resp, ok := lyricsFromPlain(body.MacroCalls["track.lyrics.get"]); ok && lyricsResponseAllowed(resp) {
		return resp, utils.LyricsFormatPlain, nil, nil, utils.NewLookupSuccess("mobile_plain")
	} else if ok {
		utils.LogInfof("mobile_api: rejected lyrics reason=generated_pseudo_lyrics")
	}

	// Damn it, that did not work.
	err = fmt.Errorf("mobile API did not return lyrics")
	failure = utils.NewLookupFailure("mobile_no_lyrics", "mobile_api", err).WithPhase("mobile_lyrics")
	return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
}
