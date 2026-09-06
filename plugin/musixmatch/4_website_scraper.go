package musixmatch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func fetchLyricsViaWebsiteScraping(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	// Fallback, scrape website
	tracks, err, failure := searchForTracks(input)
	if err == nil {
		if len(tracks) > 0 {
			var lastFailure *utils.LookupFailure
			utils.LogInfof("FetchLyrics: website search match_found")
			for _, track := range tracks {
				pdk.Log(pdk.LogDebug, fmt.Sprintf("FetchLyrics: website candidate found artist=%s title=%s album=%s mbz=%s", track.Artist, track.Title, track.CommontrackVanityID, input.Track.MBZRecordingID))
				resp, err, failure, success := scrapeWebsiteLyricsForTrack(track, input)
				if isIdentityRejection(err) {
					utils.LogInfof("FetchLyrics: website candidate rejected reason=%s", err.Error())
					continue
				}
				if err != nil {
					return resp, err, failure, success
				}
				if len(resp.Lyrics) > 0 {
					utils.LogInfof("FetchLyrics: website candidate accepted source=website category=%s", success.CategoryValue())
					return resp, nil, nil, success
				}
			}

			if lastFailure != nil {
				return lyrics.GetLyricsResponse{}, fmt.Errorf("website search failed"), lastFailure, nil
			} else {
				return lyrics.GetLyricsResponse{}, fmt.Errorf("website search found no lyrics"), nil, nil
			}
		} else {
			return lyrics.GetLyricsResponse{}, fmt.Errorf("website search found no tracks"), nil, nil
		}
	} else {
		status := 0
		reason := ""
		if failure != nil {
			status = failure.StatusCode
			reason = failure.ReasonValue()
		}

		utils.LogErrorf("FetchLyrics: website search failed reason=%s status=%d error=%v", reason, status, err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
}

var doMusixmatchWebsiteLyricsGetRequest = utils.DoMusixmatchWebsiteGetRequest

func scrapeWebsiteLyricsForTrack(track *Song, input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	endpoint := fmt.Sprintf(utils.MusixmatchFetchPageURL, track.CommontrackVanityID)

	utils.LogInfof("website lyrics page: started")

	resp, err := doMusixmatchWebsiteLyricsGetRequest(endpoint)
	if err != nil || resp == nil {
		utils.LogErrorf("website lyrics page: request failed body_present=%t error=%v", resp != nil && resp.Body != nil, err)
		failure := utils.NewLookupFailure("lyrics_page_request_failed", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	if err, failure := detectWebsiteGate("lyrics_page", resp); failure != nil {
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website lyrics page: blocked response body=%s", string(resp.Body)))
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	body := resp.Body
	if resp.StatusCode != utils.HTTPStatusOK {
		utils.LogErrorf("HTTP %d from Musixmatch", resp.StatusCode)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website lyrics page: response body=%s", string(body)))
		err := &utils.HTTPError{StatusCode: resp.StatusCode}
		failure := utils.NewLookupFailure("lyrics_page_request_failed", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("website lyrics page: response received body_bytes=%d", len(body))

	matches := WEBSITE_NEXT_DATA_REGEX.FindSubmatch(body)
	if len(matches) < 2 {
		utils.LogErrorf("website lyrics page: Next.js data not found body_bytes=%d", len(body))
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website lyrics page: response body=%s", string(body)))
		err := fmt.Errorf("could not find __NEXT_DATA__ on page")
		failure := utils.NewLookupFailure("lyrics_page_next_data_missing", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("website lyrics page: Next.js data found bytes=%d", len(matches[1]))

	var data nextDataResponse
	if err := json.Unmarshal(matches[1], &data); err != nil {
		utils.LogErrorf("website lyrics page: could not parse Next.js data bytes=%d error=%v", len(matches[1]), err)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("website lyrics page: Next.js data=%s", string(matches[1])))
		failure := utils.NewLookupFailure("lyrics_page_json_parse_failed", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	trackData := data.Props.PageProps.Data.TrackInfo.Data
	// Checking the album is not a good idea since the album could be a compilation or a different edition, so we will not check it for now.
	// if err := validateMatchedIdentity(input, trackMetadata{Artist: trackData.Track.ArtistName, Album: trackData.Track.AlbumName}, "website lyrics page"); err != nil {
	// 	return lyrics.GetLyricsResponse{}, err, nil, nil
	// }
	lang := trackData.Lyrics.Language
	structureLines := 0
	for _, section := range trackData.TrackStructureList {
		structureLines += len(section.Lines)
	}
	utils.LogInfof("website lyrics page: parsed lyrics data track_structure_sections=%d track_structure_lines=%d subtitle_lines=%d plain_present=%t plain_bytes=%d", len(trackData.TrackStructureList), structureLines, len(trackData.Subtitle), trackData.Lyrics.Body != "", len(trackData.Lyrics.Body))

	// Check for synced lyrics
	if len(trackData.TrackStructureList) > 0 {
		utils.LogInfof("website lyrics page: trying source=track_structure")
		lrc := buildLRCForTrackStructure(trackData.TrackStructureList)
		if lrc != "" {
			utils.LogInfof("website lyrics page: got synced lyrics source=track_structure lrc_lines=%d", strings.Count(lrc, "\n"))
			success := utils.NewLookupSuccess("website_sync")
			return lyrics.GetLyricsResponse{
				Lyrics: []lyrics.LyricsText{{Lang: lang, Text: lrc}},
			}, nil, nil, success
		}
	}
	if len(trackData.Subtitle) > 0 {
		utils.LogInfof("website lyrics page: trying source=subtitle")
		lrc := buildLRCForSubtitle(trackData.Subtitle)
		if lrc != "" {
			utils.LogInfof("website lyrics page: got synced lyrics source=subtitle lrc_lines=%d", strings.Count(lrc, "\n"))
			success := utils.NewLookupSuccess("website_sync")
			return lyrics.GetLyricsResponse{
				Lyrics: []lyrics.LyricsText{{Lang: lang, Text: lrc}},
			}, nil, nil, success
		}
	}

	utils.LogInfof("website lyrics page: no synced lyrics found, falling back to plain lyrics")

	// Fallback to plain lyrics
	lyricsBody := trackData.Lyrics.Body
	if lyricsBody == "" {
		utils.LogErrorf("website lyrics page: plain lyrics unavailable because body was empty")
		failure := utils.NewLookupFailure("lyrics_empty", "website", err).WithPhase("website_lyrics")
		err = fmt.Errorf("body is empty")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	utils.LogInfof("website lyrics page: got plain lyrics bytes=%d", len(lyricsBody))
	success := utils.NewLookupSuccess("website_plain")
	return lyrics.GetLyricsResponse{
		Lyrics: []lyrics.LyricsText{{Lang: lang, Text: lyricsBody}},
	}, nil, nil, success
}
