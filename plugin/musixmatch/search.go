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

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		return nil, fmt.Errorf("failed to do musixmatch search request for query %s; Error: %v", query, err)
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse musixmatch search response for query %s: %v", query, err)
	}

	searchResponse := resp.PageProps.Data.OpenSearch.Data.OpenSearchTrackSearch.Body

	if searchResponse.BestMatch != nil && searchResponse.BestMatch.Type == "track" {
		return &Song{Artist: searchResponse.BestMatch.ArtistName, Title: searchResponse.BestMatch.TrackName, CommontrackVanityID: searchResponse.BestMatch.CommontrackVanityID}, nil
	}

	if len(searchResponse.Tracks) > 0 {
		return &Song{Artist: searchResponse.Tracks[0].ArtistName, Title: searchResponse.Tracks[0].TrackName, CommontrackVanityID: searchResponse.Tracks[0].CommontrackVanityID}, nil
	}

	return nil, nil
}
