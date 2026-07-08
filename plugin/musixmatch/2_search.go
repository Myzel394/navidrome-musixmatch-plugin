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

	utils.LogInfof("searching for '%s'", query)

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		if err == nil {
			err = fmt.Errorf("empty musixmatch search response body for query %s", query)
		}
		utils.LogErrorf("search request failed for '%s': %v", query, err)
		err = fmt.Errorf("failed to do musixmatch search request for query %s: %w", query, err)
		failure := utils.NewLookupFailure("search_request_failed", "website", err)
		return nil, err, failure
	}

	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		err := fmt.Errorf("failed to find musixmatch search Next.js data for query %s", query)
		failure := utils.NewLookupFailure("search_next_data_missing", "website", err)
		return nil, err, failure
	}

	var resp searchResponse
	if err := json.Unmarshal(matches[1], &resp); err != nil {
		utils.LogErrorf("search parse failed for '%s': %v", query, err)
		err = fmt.Errorf("failed to parse musixmatch search response for query %s: %w", query, err)
		failure := utils.NewLookupFailure("search_json_parse_failed", "website", err)
		return nil, err, failure
	}

	searchResponse := resp.PageProps.Data.OpenSearch.Data.OpenSearchTrackSearch.Body

	if searchResponse.BestMatch != nil && searchResponse.BestMatch.Type == "track" {
		utils.LogInfof("search found bestMatch: '%s' by '%s' (vanity: %s)", searchResponse.BestMatch.TrackName, searchResponse.BestMatch.ArtistName, searchResponse.BestMatch.CommontrackVanityID)
		return &Song{Artist: searchResponse.BestMatch.ArtistName, Title: searchResponse.BestMatch.TrackName, CommontrackVanityID: searchResponse.BestMatch.CommontrackVanityID}, nil, nil
	}

	if len(searchResponse.Tracks) > 0 {
		utils.LogInfof("search fallback to tracks[0]: '%s' by '%s' (vanity: %s)", searchResponse.Tracks[0].TrackName, searchResponse.Tracks[0].ArtistName, searchResponse.Tracks[0].CommontrackVanityID)
		return &Song{Artist: searchResponse.Tracks[0].ArtistName, Title: searchResponse.Tracks[0].TrackName, CommontrackVanityID: searchResponse.Tracks[0].CommontrackVanityID}, nil, nil
	}

	utils.LogInfof("search returned no results for '%s'", query)
	err = fmt.Errorf("no Musixmatch search results for query %s", query)
	failure := utils.NewLookupFailure("search_no_match", "website", err)
	return nil, err, failure
}
