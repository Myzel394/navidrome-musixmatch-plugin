package musixmatch

import (
	"errors"
	"regexp"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

var (
	errIdentityRejected = errors.New("identity_rejected")
	artistJoinRe        = regexp.MustCompile(`(?i)\s+(feat\.|featuring|ft\.|&|and)\s+`)
)

type trackMetadata struct {
	Artist string
	Title  string
	Album  string
}

func isIdentityRejection(err error) bool { return errors.Is(err, errIdentityRejected) }

func rejectIdentity(reason string) error { return errors.Join(errIdentityRejected, errors.New(reason)) }

func validateMatchedIdentity(input lyrics.GetLyricsRequest, meta trackMetadata, source string) error {
	if meta.Artist == "" || meta.Title == "" {
		utils.LogInfof("%s: metadata_unavailable, rejecting lyrics", source)
		return rejectIdentity("metadata_unavailable")
	}
	if input.Track.Artist != "" && meta.Artist != "" && !artistsSimilar(input.Track.Artist, meta.Artist) {
		utils.LogInfof("%s: artist_mismatch, rejecting lyrics; requested_artist=%s actual_artist=%s", source, input.Track.Artist, meta.Artist)
		return rejectIdentity("artist_mismatch")
	}
	if input.Track.Title != "" && meta.Title != "" && !validStrictSimilarity(input.Track.Title, meta.Title) {
		utils.LogInfof("%s: title_mismatch, rejecting lyrics; requested_title=%s actual_title=%s", source, input.Track.Title, meta.Title)
		return rejectIdentity("title_mismatch")
	}
	if input.Track.Album != "" && meta.Album != "" && !validStrictSimilarity(input.Track.Album, meta.Album) {
		utils.LogInfof("%s: album_mismatch informational; requested_album=%s actual_album=%s", source, input.Track.Album, meta.Album)
	}
	return nil
}

func artistsSimilar(a, b string) bool {
	if validStrictSimilarity(a, b) {
		return true
	}
	return validStrictSimilarity(primaryArtist(a), primaryArtist(b))
}

func primaryArtist(s string) string {
	parts := artistJoinRe.Split(s, 2)
	return strings.TrimSpace(parts[0])
}

func appendWebsiteCandidates(dst []*Song, seen map[string]bool, src []*Song) []*Song {
	limit := utils.ConfigMaxWebsiteCandidateAttempts()
	for _, s := range src {
		if len(dst) >= limit {
			break
		}
		if s == nil || s.CommontrackVanityID == "" || seen[s.CommontrackVanityID] {
			continue
		}
		seen[s.CommontrackVanityID] = true
		dst = append(dst, s)
	}
	return dst
}

func orderWebsiteCandidates(best *searchTrack, tracks []searchTrack, mode string, input lyrics.GetLyricsRequest) []*Song {
	limit := utils.ConfigMaxWebsiteCandidateAttempts()
	ordered := make([]*Song, 0, limit)
	seen := map[string]bool{}
	add := func(track *searchTrack) {
		if len(ordered) >= limit || track == nil || track.Type != "track" || track.CommontrackVanityID == "" || seen[track.CommontrackVanityID] {
			return
		}
		if mode == "title_only" && titleOnlyArtistMismatch(input, track.ArtistName) {
			return
		}
		seen[track.CommontrackVanityID] = true
		ordered = append(ordered, songFromSearchTrack(track))
	}
	add(best)
	for i := range tracks {
		add(&tracks[i])
	}
	return ordered
}

func titleOnlyArtistMismatch(input lyrics.GetLyricsRequest, candidateArtist string) bool {
	return input.Track.Artist != "" && candidateArtist != "" && !artistsSimilar(input.Track.Artist, candidateArtist)
}
