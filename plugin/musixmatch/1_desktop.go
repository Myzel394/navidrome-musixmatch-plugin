package musixmatch

import (
	"encoding/json"
	"fmt"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func fetchLyricsFromDesktopAPI(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	utils.LogInfof("desktop API: lookup started duration_filter=%t", input.Track.Duration > 0)

	token, err, failure := desktopUserToken()
	if err != nil {
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	query := desktopLyricsQuery(input, token)

	var resp macroResponse
	err = desktopGet("macro.subtitles.get", query, &resp)
	if err != nil {
		utils.LogErrorf("desktop API: lyrics request failed error=%v", err)
		failure := utils.NewLookupFailure("desktop_macro_request_failed", "desktop_api", err).WithPhase("desktop_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("desktop API: lyrics response received status=%d body_bytes=%d", resp.Message.Header.StatusCode, len(resp.Message.Body))
	if resp.Message.Header.StatusCode == utils.HTTPStatusBlocked {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: lyrics response body=%s", string(resp.Message.Body)))
		desktopInvalidateUserToken()
		err := fmt.Errorf("desktop API returned 401 for lyrics request")
		failure := utils.NewLookupFailure("desktop_macro_blocked", "desktop_api", err).WithPhase("desktop_lyrics").WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	if resp.Message.Header.StatusCode != utils.HTTPStatusOK {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: lyrics response body=%s", string(resp.Message.Body)))
		err := fmt.Errorf("desktop API returned status %d for lyrics request", resp.Message.Header.StatusCode)
		failure := utils.NewLookupFailure("desktop_macro_status", "desktop_api", err).WithPhase("desktop_lyrics").WithStatusCode(resp.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	var body macroBody
	if err := json.Unmarshal(resp.Message.Body, &body); err != nil {
		utils.LogErrorf("desktop API: failed to parse desktop API macro body body_bytes=%d error=%v", len(resp.Message.Body), err)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: lyrics response body=%s", string(resp.Message.Body)))
		failure := utils.NewLookupFailure("desktop_macro_parse", "desktop_api", err).WithPhase("desktop_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("desktop API: parsed lyrics response calls=%d richsync_available=%t subtitle_available=%t plain_available=%t", len(body.MacroCalls), body.MacroCalls["track.richsync.get"].Message.Header.StatusCode != 0, body.MacroCalls["track.subtitles.get"].Message.Header.StatusCode != 0, body.MacroCalls["track.lyrics.get"].Message.Header.StatusCode != 0)

	meta := trackMetadata{}
	matcherCall := body.MacroCalls["matcher.track.get"]
	if matcherCall.Message.Header.StatusCode == utils.HTTPStatusOK && len(matcherCall.Message.Body) > 0 {
		var trackBody macroTrackBody
		if err := json.Unmarshal(matcherCall.Message.Body, &trackBody); err != nil {
			utils.LogInfof("desktop API: matched track metadata could not be parsed body_bytes=%d", len(matcherCall.Message.Body))
		} else {
			meta = trackMetadata{Artist: trackBody.Track.ArtistName, Title: trackBody.Track.TrackName, Album: trackBody.Track.AlbumName}
		}
	}
	if err := validateMatchedIdentity(input, meta, "desktop API"); err != nil {
		utils.LogInfof("desktop API: matched metadata rejected")
		return lyrics.GetLyricsResponse{}, nil, nil, nil
	}
	if resp, ok := lyricsFromRichsync(body.MacroCalls["track.richsync.get"]); ok && lyricsResponseAllowed(resp) {
		utils.LogInfof("desktop API: selected lyrics representation=richsync")
		success := utils.NewLookupSuccess("desktop_synced")
		return resp, nil, nil, success
	} else if ok {
		utils.LogInfof("desktop_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	if resp, ok := lyricsFromSubtitle(body.MacroCalls["track.subtitles.get"]); ok && lyricsResponseAllowed(resp) {
		utils.LogInfof("desktop API: selected lyrics representation=subtitle")
		success := utils.NewLookupSuccess("desktop_synced")
		return resp, nil, nil, success
	} else if ok {
		utils.LogInfof("desktop_api: rejected lyrics reason=generated_pseudo_lyrics")
	}
	if resp, ok := lyricsFromPlain(body.MacroCalls["track.lyrics.get"]); ok && lyricsResponseAllowed(resp) {
		utils.LogInfof("desktop API: selected lyrics representation=plain")
		success := utils.NewLookupSuccess("desktop_plain")
		return resp, nil, nil, success
	} else if ok {
		utils.LogInfof("desktop_api: rejected lyrics reason=generated_pseudo_lyrics")
	}

	utils.LogInfof("desktop API: lookup finished without lyrics")
	err = fmt.Errorf(desktopFallbackErr)
	failure = utils.NewLookupFailure("desktop_no_lyrics", "desktop_api", err).WithPhase("desktop_lyrics")
	return lyrics.GetLyricsResponse{}, err, failure, nil
}
