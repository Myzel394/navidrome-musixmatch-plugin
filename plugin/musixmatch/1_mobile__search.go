package musixmatch

import (
	"cmp"
	"encoding/json"
	"math"
	"slices"

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
	TrackLength  int        `json:"track_length"`
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
func searchMobileTrackIDs(input lyrics.GetLyricsRequest, token string) []int {
	response, err := mobileGet("track.search", buildMobileSearchQuery(input, token))
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

// Reduces search results to the track IDs worth fetching lyrics for. Skips
// duplicates, entries without lyrics, instrumentals, and entries whose identity
// does not match the requested track. The rest are ordered by how close their
// length is to the requested track, because the same title can be a live or
// extended version whose synced lyrics would drift. Keeps at most
// `mobileCandidateSweepLimit`, since each returned ID costs one request.
func filterMobileTrackIDs(input lyrics.GetLyricsRequest, body mobileSearchBody) []int {
	tracks := make([]mobileSearchTrack, 0, len(body.TrackList))
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
		tracks = append(tracks, track)
	}

	sortMobileTracksByLengthDifference(tracks, float64(input.Track.Duration))

	trackIDs := make([]int, 0, mobileCandidateSweepLimit)
	for _, track := range tracks {
		if len(trackIDs) >= mobileCandidateSweepLimit {
			break
		}
		trackIDs = append(trackIDs, track.TrackID)
	}
	return trackIDs
}

// Keeps the search order for tracks that are equally close, and for every track
// when the requested duration is unknown. Tracks of unknown length go last.
func sortMobileTracksByLengthDifference(tracks []mobileSearchTrack, duration float64) {
	if duration <= 0 {
		return
	}
	lengthDifference := func(track mobileSearchTrack) float64 {
		if track.TrackLength <= 0 {
			return math.Inf(1)
		}
		return math.Abs(float64(track.TrackLength) - duration)
	}
	slices.SortStableFunc(tracks, func(a, b mobileSearchTrack) int {
		return cmp.Compare(lengthDifference(a), lengthDifference(b))
	})
}
