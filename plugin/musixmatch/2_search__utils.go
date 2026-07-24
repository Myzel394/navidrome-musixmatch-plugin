package musixmatch

import (
	"strings"
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

func songFromSearchTrack(track *searchTrack) *Song {
	return &Song{Artist: track.ArtistName, Title: track.TrackName, CommontrackVanityID: track.CommontrackVanityID}
}

func collapseSearchWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
