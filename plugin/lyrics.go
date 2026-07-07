package main

import (
	"fmt"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/musixmatch"
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func (p *plugin) GetLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error) {
	utils.StartLogCapture()

	startedAt := time.Now().UTC()
	resp, err, failure := musixmatch.FetchLyrics(input)
	duration := time.Since(startedAt)

	if err != nil {
		utils.LogErrorf("GetLyrics failed: %v", err)
	} else if len(resp.Lyrics) == 0 {
		utils.LogInfof("GetLyrics completed without lyrics")
	}

	logs := utils.StopLogCapture()
	utils.ReportLyricsLookup(input, resp, err, failure, startedAt, duration, logs)

	if err != nil {
		return resp, fmt.Errorf("%s%w", utils.LogPrefix, err)
	}
	return resp, nil
}
