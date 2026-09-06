package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

var doMusixmatchWebsiteSearchGetRequest = utils.DoMusixmatchWebsiteGetRequest

func searchForTracks(input lyrics.GetLyricsRequest) ([]*Song, error, *utils.LookupFailure) {
	normArtist := normalize(input.Track.Artist)
	normTitle := normalize(input.Track.Title)
	candidateLimit := utils.ConfigMaxWebsiteCandidateAttempts()
	candidates := make([]*Song, 0, candidateLimit)
	seen := map[string]bool{}

	utils.LogInfof("website search: started artist_present=%t title_present=%t", normArtist != "", normTitle != "")
	utils.LogInfof("website search: searching using normalized artist and title artist=%s title=%s", normArtist, normTitle)
	// Normalized search
	if normArtist != "" && normTitle != "" {
		tracks, err, failure := _searchForQuery(normArtist+" "+normTitle, "artist_title", input)
		if err != nil || failure != nil {
			return nil, err, failure
		}
		candidates = appendWebsiteCandidates(candidates, seen, tracks)
	}
	utils.LogInfof("website search: got %d candidates now", len(candidates))
	if len(candidates) >= candidateLimit {
		return candidates, nil, nil
	}

	utils.LogInfof("website search: searching using raw artist and title artist=%s title=%s", input.Track.Artist, input.Track.Title)
	rawArtist := collapseSearchWhitespace(input.Track.Artist)
	rawTitle := collapseSearchWhitespace(input.Track.Title)
	if rawArtist != "" && rawTitle != "" {
		tracks, err, failure := _searchForQuery(rawArtist+" "+rawTitle, "artist_title", input)
		if err != nil || failure != nil {
			return nil, err, failure
		}
		candidates = appendWebsiteCandidates(candidates, seen, tracks)
	}
	utils.LogInfof("website search: got %d candidates now", len(candidates))
	if len(candidates) >= candidateLimit {
		return candidates, nil, nil
	}

	utils.LogInfof("website search: searching using raw artist only artist=%s", input.Track.Artist)
	if rawTitle != "" {
		tracks, err, failure := _searchForQuery(rawTitle, "title_only", input)
		if err != nil || failure != nil {
			return nil, err, failure
		}
		candidates = appendWebsiteCandidates(candidates, seen, tracks)
	}
	utils.LogInfof("website search: got %d candidates now", len(candidates))
	if len(candidates) > 0 {
		return candidates, nil, nil
	}

	utils.LogInfof("website search: no_match_found after all attempts")
	return nil, nil, nil
}

func _searchForQuery(query, mode string, input lyrics.GetLyricsRequest) ([]*Song, error, *utils.LookupFailure) {
	endpoint := fmt.Sprintf(utils.MusixmatchSearchPageURL, url.QueryEscape(query))
	utils.LogInfof("website search: query attempt query_chars=%d", len(query))

	httpResp, err := doMusixmatchWebsiteSearchGetRequest(endpoint)
	if err != nil || httpResp == nil {
		utils.LogErrorf("website search: search request failed body_present=%t error=%v", httpResp != nil && httpResp.Body != nil, err)
		failure := utils.NewLookupFailure("search_request_failed", "website", err).WithPhase("website_search")
		return nil, err, failure
	}
	if err, failure := detectWebsiteGate("search", httpResp); failure != nil {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website search: blocked response body=%s", string(httpResp.Body)))
		return nil, err, failure
	}
	body := httpResp.Body
	if httpResp.StatusCode != utils.HTTPStatusOK {
		utils.LogErrorf("HTTP %d from Musixmatch", httpResp.StatusCode)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website search: response body=%s", string(body)))
		err := &utils.HTTPError{StatusCode: httpResp.StatusCode}
		failure := utils.NewLookupFailure("search_request_failed", "website", err).WithPhase("website_search")
		return nil, err, failure
	}
	utils.LogInfof("website search: response received body_bytes=%d", len(body))

	matches := WEBSITE_NEXT_DATA_REGEX.FindSubmatch(body)
	if len(matches) < 2 {
		utils.LogInfof("website search: failed to find musixmatch search Next.js data body_bytes=%d", len(body))
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website search: response body=%s", string(body)))
		err := fmt.Errorf("failed to find musixmatch search Next.js data")
		failure := utils.NewLookupFailure("search_next_data_missing", "website", err).WithPhase("website_search")
		return nil, err, failure
	}
	utils.LogInfof("website search: Next.js data found bytes=%d", len(matches[1]))

	var resp searchResponse
	if err := json.Unmarshal(matches[1], &resp); err != nil {
		utils.LogErrorf("website search: could not parse Next.js data bytes=%d error=%v", len(matches[1]), err)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website search: Next.js data=%s", string(matches[1])))
		failure := utils.NewLookupFailure("search_json_parse_failed", "website", err).WithPhase("website_search")
		return nil, err, failure
	}

	searchResponse := resp.PageProps.Data.OpenSearch.Data.OpenSearchTrackSearch.Body
	bestMatchIsTrack := searchResponse.BestMatch != nil && searchResponse.BestMatch.Type == "track"
	utils.LogInfof("website search: parsed results result_count=%d best_match_present=%t best_match_is_track=%t", len(searchResponse.Tracks), searchResponse.BestMatch != nil, bestMatchIsTrack)

	if bestMatchIsTrack {
		utils.LogInfof("website search: match_found using best match")
	}
	tracks := orderWebsiteCandidates(searchResponse.BestMatch, searchResponse.Tracks, mode, input)
	if len(tracks) == 0 {
		utils.LogInfof("website search: no_match_found because results were empty best_match_present=%t best_match_is_track=%t", searchResponse.BestMatch != nil, bestMatchIsTrack)
	}
	return tracks, nil, nil
}
