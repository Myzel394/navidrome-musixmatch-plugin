package musixmatch

import (
	"encoding/json"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

// Mobile `track.search` response envelope. Same message/header/body shape as
// every other mobile API call, with a track list payload.
type mobileSearchBody struct {
	TrackList []struct {
		Track mobileSearchTrack `json:"track"`
	} `json:"track_list"`
}

type mobileSearchTrack struct {
	TrackID      int        `json:"track_id"`
	TrackName    string     `json:"track_name"`
	ArtistName   string     `json:"artist_name"`
	AlbumName    string     `json:"album_name"`
	HasLyrics    mobileBool `json:"has_lyrics"`
	Instrumental mobileBool `json:"instrumental"`
}

// The API encodes flags as the numbers 0 and 1 rather than JSON booleans.
type mobileBool bool

func (b *mobileBool) UnmarshalJSON(data []byte) error {
	*b = string(data) == "1"
	return nil
}

// `track.search` returns metadata only — a `track_list` of up to 10 entries,
// each {track_id, track_name, artist_name, album_name, has_lyrics,
// instrumental} — never lyric text. Musixmatch keeps duplicate catalog entries
// per song (plain vs synced, solo vs feat. credit), so searching reveals every
// duplicate and lets lyrics be fetched per explicit `track_id` instead of
// trusting the server-side matcher to pick the right entry.
func searchMobileTrackIDs(input lyrics.GetLyricsRequest, token string, assignment mobileGUIDAssignment) []int {
	response, err := mobileGet("track.search", buildMobileSearchQuery(input, token), assignment)
	if err != nil {
		utils.LogInfof("mobile API: candidate search request failed error=%v", err)
		return nil
	}
	if response.Message.Header.StatusCode != utils.HTTPStatusOK {
		utils.LogInfof("mobile API: candidate search returned status=%d", response.Message.Header.StatusCode)
		return nil
	}

	var body mobileSearchBody
	if err := json.Unmarshal(response.Message.Body, &body); err != nil {
		utils.LogInfof("mobile API: candidate search response could not be parsed body_bytes=%d", len(response.Message.Body))
		return nil
	}

	return filterMobileTrackIDs(input, body)
}

// Reduces search results to the track IDs worth fetching lyrics for, keeping
// the search order. Skips duplicates, entries without lyrics, instrumentals,
// and entries whose identity does not match the requested track. Stops at
// `mobileCandidateSweepLimit`, since each returned ID costs one request.
func filterMobileTrackIDs(input lyrics.GetLyricsRequest, body mobileSearchBody) []int {
	trackIDs := make([]int, 0, mobileCandidateSweepLimit)
	seen := map[int]bool{}
	for _, item := range body.TrackList {
		track := item.Track
		if track.TrackID == 0 || seen[track.TrackID] {
			continue
		}
		if !track.HasLyrics || track.Instrumental {
			continue
		}
		seen[track.TrackID] = true
		meta := trackMetadata{Artist: track.ArtistName, Title: track.TrackName, Album: track.AlbumName}
		if err := validateMatchedIdentity(input, meta, "mobile API candidates"); err != nil {
			continue
		}
		trackIDs = append(trackIDs, track.TrackID)
		if len(trackIDs) >= mobileCandidateSweepLimit {
			break
		}
	}
	return trackIDs
}
