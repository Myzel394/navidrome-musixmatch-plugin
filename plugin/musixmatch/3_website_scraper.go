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

	utils.LogInfof("website lyrics page: started")

	resp, err := utils.DoMusixmatchWebsiteGetRequest(endpoint)
	if err != nil || resp == nil {
		utils.LogErrorf("website lyrics page: request failed body_present=%t error=%v", resp != nil && resp.Body != nil, err)
		failure := utils.NewLookupFailure("lyrics_page_request_failed", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	if err, failure := detectWebsiteGate("lyrics_page", resp); failure != nil {
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	body := resp.Body
	if resp.StatusCode != utils.HTTPStatusOK {
		utils.LogErrorf("HTTP %d from Musixmatch", resp.StatusCode)
		err := &utils.HTTPError{StatusCode: resp.StatusCode}
		failure := utils.NewLookupFailure("lyrics_page_request_failed", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("website lyrics page: response received body_bytes=%d", len(body))

	matches := nextDataRe.FindSubmatch(body)
	if len(matches) < 2 {
		utils.LogErrorf("website lyrics page: Next.js data not found body_bytes=%d", len(body))
		err := fmt.Errorf("could not find __NEXT_DATA__ on page")
		failure := utils.NewLookupFailure("lyrics_page_next_data_missing", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}
	utils.LogInfof("website lyrics page: Next.js data found bytes=%d", len(matches[1]))

	var data nextDataResponse
	if err := json.Unmarshal(matches[1], &data); err != nil {
		utils.LogErrorf("website lyrics page: could not parse Next.js data bytes=%d error=%v", len(matches[1]), err)
		failure := utils.NewLookupFailure("lyrics_page_json_parse_failed", "website", err).WithPhase("website_lyrics")
		return lyrics.GetLyricsResponse{}, err, failure, nil
	}

	trackData := data.Props.PageProps.Data.TrackInfo.Data
	lang := trackData.Lyrics.Language
	structureLines := 0
	for _, section := range trackData.TrackStructureList {
		structureLines += len(section.Lines)
	}
	utils.LogInfof("website lyrics page: parsed lyrics data track_structure_sections=%d track_structure_lines=%d subtitle_lines=%d plain_present=%t plain_bytes=%d", len(trackData.TrackStructureList), structureLines, len(trackData.Subtitle), trackData.Lyrics.Body != "", len(trackData.Lyrics.Body))

	// Check for synced lyrics
	if len(trackData.TrackStructureList) > 0 {
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
		utils.LogInfof("website lyrics page: trying subtitle fallback")
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
