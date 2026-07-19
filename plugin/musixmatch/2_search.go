package musixmatch

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type searchTrack struct {
	TrackName           string `json:"track_name"`
	ArtistName          string `json:"artist_name"`
	Type                string `json:"type"`
	CommontrackVanityID string `json:"commontrack_vanity_id"`
}

type searchResponse struct {
	PageProps struct {
		Data struct {
			OpenSearch struct {
				Data struct {
					OpenSearchTrackSearch struct {
						Body struct {
							Tracks    []searchTrack `json:"tracks"`
							BestMatch *searchTrack  `json:"bestMatch"`
						} `json:"body"`
					} `json:"opensearchTrackSearch"`
				} `json:"data"`
			} `json:"openSearch"`
		} `json:"data"`
	} `json:"pageProps"`
}

func searchForTrack(input lyrics.GetLyricsRequest) (*Song, error, *utils.LookupFailure) {
	normArtist := normalize(input.Track.Artist)
	normTitle := normalize(input.Track.Title)

	query := fmt.Sprintf("%s %s", normArtist, normTitle)
	endpoint := fmt.Sprintf(utils.MusixmatchSearchPageURL, url.QueryEscape(query))

	utils.LogInfof("website search: started artist_present=%t title_present=%t", normArtist != "", normTitle != "")

	httpResp, err := utils.DoMusixmatchWebsiteGetRequest(endpoint)
	if err != nil || httpResp == nil {
		utils.LogErrorf("website search: search request failed body_present=%t error=%v", httpResp != nil && httpResp.Body != nil, err)
		failure := utils.NewLookupFailure("search_request_failed", "website", err).WithPhase("website_search")
		return nil, err, failure
	}
	if err, failure := detectWebsiteGate("search", httpResp); failure != nil {
		return nil, err, failure
	}
	body := httpResp.Body
	if httpResp.StatusCode != utils.HTTPStatusOK {
		utils.LogErrorf("HTTP %d from Musixmatch", httpResp.StatusCode)
		err := &utils.HTTPError{StatusCode: httpResp.StatusCode}
		failure := utils.NewLookupFailure("search_request_failed", "website", err).WithPhase("website_search")
		return nil, err, failure
	}
	utils.LogInfof("website search: response received body_bytes=%d", len(body))

	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		utils.LogInfof("website search: failed to find musixmatch search Next.js data body_bytes=%d", len(body))
		err := fmt.Errorf("failed to find musixmatch search Next.js data")
		failure := utils.NewLookupFailure("search_next_data_missing", "website", err).WithPhase("website_search")
		return nil, err, failure
	}
	utils.LogInfof("website search: Next.js data found bytes=%d", len(matches[1]))

	var resp searchResponse
	if err := json.Unmarshal(matches[1], &resp); err != nil {
		utils.LogErrorf("website search: could not parse Next.js data bytes=%d error=%v", len(matches[1]), err)
		failure := utils.NewLookupFailure("search_json_parse_failed", "website", err).WithPhase("website_search")
		return nil, err, failure
	}

	searchResponse := resp.PageProps.Data.OpenSearch.Data.OpenSearchTrackSearch.Body
	bestMatchIsTrack := searchResponse.BestMatch != nil && searchResponse.BestMatch.Type == "track"
	utils.LogInfof("website search: parsed results result_count=%d best_match_present=%t best_match_is_track=%t", len(searchResponse.Tracks), searchResponse.BestMatch != nil, bestMatchIsTrack)

	if bestMatchIsTrack {
		utils.LogInfof("website search: match_found using best match")
		return &Song{Artist: searchResponse.BestMatch.ArtistName, Title: searchResponse.BestMatch.TrackName, CommontrackVanityID: searchResponse.BestMatch.CommontrackVanityID}, nil, nil
	}

	if len(searchResponse.Tracks) > 0 {
		utils.LogInfof("website search: match_found using first result")
		return &Song{Artist: searchResponse.Tracks[0].ArtistName, Title: searchResponse.Tracks[0].TrackName, CommontrackVanityID: searchResponse.Tracks[0].CommontrackVanityID}, nil, nil
	}

	utils.LogInfof("website search: no_match_found because results were empty best_match_present=%t best_match_is_track=%t", searchResponse.BestMatch != nil, bestMatchIsTrack)
	err = fmt.Errorf("no Musixmatch search results")
	failure := utils.NewLookupFailure("search_no_match", "website", err).WithPhase("website_search")
	return nil, err, failure
}
