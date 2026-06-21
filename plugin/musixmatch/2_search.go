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

func searchForTrack(input lyrics.GetLyricsRequest) (*Song, error) {
	normArtist := normalize(input.Track.Artist)
	normTitle := normalize(input.Track.Title)

	query := fmt.Sprintf("%s %s", normArtist, normTitle)
	endpoint := fmt.Sprintf(utils.MusixmatchSearchPageURL, url.QueryEscape(query))

	utils.LogInfof("searching for '%s' -> %s", query, endpoint)

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		utils.LogErrorf("search request failed for '%s': %v", query, err)
		return nil, fmt.Errorf("failed to do musixmatch search request for query %s; Error: %v", query, err)
	}

	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to find musixmatch search Next.js data for query %s", query)
	}

	var resp searchResponse
	if err := json.Unmarshal(matches[1], &resp); err != nil {
		utils.LogErrorf("search parse failed for '%s': %v", query, err)
		return nil, fmt.Errorf("failed to parse musixmatch search response for query %s: %v", query, err)
	}

	searchResponse := resp.PageProps.Data.OpenSearch.Data.OpenSearchTrackSearch.Body

	if searchResponse.BestMatch != nil && searchResponse.BestMatch.Type == "track" {
		utils.LogInfof("search found bestMatch: '%s' by '%s' (vanity: %s)", searchResponse.BestMatch.TrackName, searchResponse.BestMatch.ArtistName, searchResponse.BestMatch.CommontrackVanityID)
		return &Song{Artist: searchResponse.BestMatch.ArtistName, Title: searchResponse.BestMatch.TrackName, CommontrackVanityID: searchResponse.BestMatch.CommontrackVanityID}, nil
	}

	if len(searchResponse.Tracks) > 0 {
		utils.LogInfof("search fallback to tracks[0]: '%s' by '%s' (vanity: %s)", searchResponse.Tracks[0].TrackName, searchResponse.Tracks[0].ArtistName, searchResponse.Tracks[0].CommontrackVanityID)
		return &Song{Artist: searchResponse.Tracks[0].ArtistName, Title: searchResponse.Tracks[0].TrackName, CommontrackVanityID: searchResponse.Tracks[0].CommontrackVanityID}, nil
	}

	utils.LogInfof("search returned no results for '%s'", query)
	return nil, nil
}
