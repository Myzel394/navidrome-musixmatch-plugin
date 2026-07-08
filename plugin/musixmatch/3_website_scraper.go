package musixmatch

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

var nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type subtitleLine struct {
	Text string `json:"text"`
	Time struct {
		Total      float64 `json:"total"`
		Minutes    int     `json:"minutes"`
		Seconds    int     `json:"seconds"`
		Hundredths int     `json:"hundredths"`
	} `json:"time"`
	Type string `json:"type"`
}

type sectionLine struct {
	Title string         `json:"title"`
	Lines []subtitleLine `json:"lines"`
}

type nextDataResponse struct {
	Props struct {
		PageProps struct {
			Data struct {
				TrackInfo struct {
					Data struct {
						Lyrics struct {
							Body     string `json:"body"`
							Language string `json:"language"`
						} `json:"lyrics"`
						TrackStructureList []sectionLine  `json:"trackStructureList"`
						Subtitle           []subtitleLine `json:"subtitle"`
					} `json:"data"`
				} `json:"trackInfo"`
			} `json:"data"`
		} `json:"pageProps"`
	} `json:"props"`
}

func scrapeWebsiteLyricsForTrack(track *Song) (lyrics.GetLyricsResponse, error, *utils.LookupFailure, *utils.LookupSuccess) {
	endpoint := fmt.Sprintf(utils.MusixmatchFetchPageURL, track.CommontrackVanityID)

	utils.LogInfof("fetching lyrics page for %s", track.CommontrackVanityID)

	body, err := utils.DoGetRequest(endpoint)
	if err != nil || body == nil {
		if err == nil {
			err = fmt.Errorf("empty musixmatch lyrics page response body for %s", track.CommontrackVanityID)
		}
		utils.LogErrorf("fetch page failed for %s: %v", track.CommontrackVanityID, err)
		err = fmt.Errorf("failed to fetch musixmatch page for %s: %w", track.CommontrackVanityID, err)
		failure := utils.NewLookupFailure("lyrics_page_request_failed", "website", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		utils.LogErrorf("no __NEXT_DATA__ on page for %s", track.CommontrackVanityID)
		err := fmt.Errorf("could not find __NEXT_DATA__ on page for %s", track.CommontrackVanityID)
		failure := utils.NewLookupFailure("lyrics_page_next_data_missing", "website", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	var data nextDataResponse
	if err := json.Unmarshal(matches[1], &data); err != nil {
		utils.LogErrorf("__NEXT_DATA__ parse failed for %s: %v", track.CommontrackVanityID, err)
		err = fmt.Errorf("failed to parse __NEXT_DATA__ for %s: %w", track.CommontrackVanityID, err)
		failure := utils.NewLookupFailure("lyrics_page_json_parse_failed", "website", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	trackData := data.Props.PageProps.Data.TrackInfo.Data
	lang := trackData.Lyrics.Language

	// Check for synced lyrics
	if len(trackData.TrackStructureList) > 0 {
		lrc := buildLRCForTrackStructure(trackData.TrackStructureList)
		if lrc != "" {
			utils.LogInfof("got synced lyrics (via trackStructureList) for %s (%d lines LRC)", track.CommontrackVanityID, strings.Count(lrc, "\n"))
			success := utils.NewLookupSuccess("website_sync")
			return lyrics.GetLyricsResponse{
				Lyrics: []lyrics.LyricsText{{Lang: lang, Text: lrc}},
			}, nil, nil, success
		}
	} else if len(trackData.Subtitle) > 0 {
		lrc := buildLRCForSubtitle(trackData.Subtitle)
		if lrc != "" {
			utils.LogInfof("got synced lyrics (via subtitle) for %s (%d lines LRC)", track.CommontrackVanityID, strings.Count(lrc, "\n"))
			success := utils.NewLookupSuccess("website_sync")
			return lyrics.GetLyricsResponse{
				Lyrics: []lyrics.LyricsText{{Lang: lang, Text: lrc}},
			}, nil, nil, success
		}
	}

	// Fallback to plain lyrics
	lyricsBody := trackData.Lyrics.Body
	if lyricsBody == "" {
		utils.LogErrorf("empty lyrics body for %s", track.CommontrackVanityID)
		err := fmt.Errorf("no lyrics found for track %s", track.CommontrackVanityID)
		failure := utils.NewLookupFailure("lyrics_empty", "website", err)
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	utils.LogInfof("got plain lyrics for %s (%d chars)", track.CommontrackVanityID, len(lyricsBody))
	success := utils.NewLookupSuccess("website plain")
	return lyrics.GetLyricsResponse{
		Lyrics: []lyrics.LyricsText{{Lang: lang, Text: lyricsBody}},
	}, nil, nil, success
}
