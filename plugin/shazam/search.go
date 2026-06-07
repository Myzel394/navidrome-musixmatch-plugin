package shazam

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Myzel394/navidrome-shazam-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type searchHit struct {
	ID         string `json:"id"`
	Attributes struct {
		ArtistName string `json:"artistName"`
		Name       string `json:"name"`
	} `json:"attributes"`
}

type searchResponse struct {
	Results struct {
		Songs struct {
			Data []searchHit `json:"data"`
		} `json:"songs"`
	} `json:"results"`
}

func doSearch(query, country string, searchLimit int) ([]searchHit, error) {
	endpoint := fmt.Sprintf(utils.ShazamSearchAPIURL, country, url.QueryEscape(query), searchLimit)

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		return nil, fmt.Errorf("failed to do shazam search request for query %s; Error: %v", query, err)
	}

	var result searchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse shazam search response for query %s: %v", query, err)
	}

	return result.Results.Songs.Data, nil
}

func searchForTrack(input lyrics.GetLyricsRequest) (*Song, error) {
	normArtist := normalize(input.Track.Artist)
	normTitle := normalize(input.Track.Title)
	country := utils.ConfigSearchCountry()
	searchLimit := utils.ConfigSearchLimit()

	// Primary: artist + track query.
	hits, err := doSearch(normArtist+" "+normTitle, country, searchLimit)
	if err != nil {
		return nil, err
	}
	if best := pickBestMatch(hits, normArtist, normTitle); best != nil {
		return best, nil
	}

	// Fallback: search track-only (sometimes short queries rank better).
	hits, err = doSearch(normTitle, country, searchLimit)
	if err != nil {
		return nil, err
	}
	return pickBestMatch(hits, normArtist, normTitle), nil
}

// pickBestMatch scores every hit against the normalized artist/title and
// returns the Song with the highest combined Levenshtein ratio, provided
// both individual ratios meet the minimum threshold.
func pickBestMatch(hits []searchHit, normArtist, normTitle string) *Song {
	matchThreshold := utils.ConfigSearchLevenshteinThreshold()

	var bestSong *Song
	var bestScore float64

	for _, song := range hits {
		hitArtist := normalize(song.Attributes.ArtistName)
		hitTitle := normalize(song.Attributes.Name)

		artistRatio := levenshteinRatio(normArtist, hitArtist)
		titleRatio := levenshteinRatio(normTitle, hitTitle)

		if artistRatio < matchThreshold || titleRatio < matchThreshold {
			continue
		}

		score := (artistRatio + titleRatio) / 2
		if score > bestScore {
			bestScore = score
			bestSong = &Song{
				ID:     song.ID,
				Artist: hitArtist,
				Title:  hitTitle,
			}
		}
	}

	return bestSong
}
