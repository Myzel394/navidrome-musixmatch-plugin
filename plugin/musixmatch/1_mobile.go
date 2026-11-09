package musixmatch

import (
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func fetchLyricsFromMobileAPI(input lyrics.GetLyricsRequest, assignment mobileGUIDAssignment) (lyrics.GetLyricsResponse, utils.LyricsFormat, error, *utils.LookupFailure, *utils.LookupSuccess) {
	token, err, failure := mobileUserToken(assignment)
	if err != nil {
		return lyrics.GetLyricsResponse{}, utils.LyricsFormatUnknown, err, failure, nil
	}

	// Resolve explicit track IDs first, so lyrics are fetched for known catalog
	// entries instead of whichever duplicate the server-side matcher picks.
	trackIDs := searchMobileTrackIDs(input, token, assignment)
	utils.LogInfof("mobile API: search resolved track_ids=%d", len(trackIDs))

	// No track ID found: fall back to a query-based macro call and let the
	// server-side matcher pick the track.
	if len(trackIDs) == 0 {
		return fetchMobileLyricsByQuery(input, token, assignment)
	}

	// Stop at the first synced lyrics. Plain lyrics are only used once no track
	// has synced ones, so those macros are kept until every track was tried.
	macros := make([]macroBody, 0, len(trackIDs))
	for _, trackID := range trackIDs {
		macro, err, failure := mobileGetMacro(buildMobileTrackLyricsQuery(trackID, token), assignment)
		if err != nil {
			utils.LogInfof("mobile API: track macro failed track_id=%d reason=%s error=%v", trackID, failure.ReasonValue(), err)
			continue
		}
		if resp, ok := getSyncedLyricsFromMacro(macro); ok {
			return resp, utils.LyricsFormatSyncedLRC, nil, nil, utils.NewLookupSuccess("mobile_synced")
		}
		macros = append(macros, macro)
	}
	return findBestPlainLyrics(macros)
}

func fetchMobileLyricsByQuery(input lyrics.GetLyricsRequest, token string, assignment mobileGUIDAssignment) (lyrics.GetLyricsResponse, utils.LyricsFormat, error, *utils.LookupFailure, *utils.LookupSuccess) {
	body, err, failure := mobileGetMacro(buildMobileLyricsQuery(input, token), assignment)
	if err != nil {
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
	if resp, ok := getSyncedLyricsFromMacro(body); ok {
		return resp, utils.LyricsFormatSyncedLRC, nil, nil, utils.NewLookupSuccess("mobile_synced")
	}
	return findBestPlainLyrics([]macroBody{body})
}
