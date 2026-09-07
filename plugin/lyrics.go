package main

import (
	"fmt"
	"time"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/analytics"
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/lyrics_writer"
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/musixmatch"
	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func (p *plugin) GetLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error) {
	utils.StartLogCapture()

	startedAt := time.Now().UTC()
	resp, format, err, failure, success, mobileGUIDVariant := musixmatch.FetchLyrics(input)
	duration := time.Since(startedAt)

	if err != nil {
		utils.LogErrorf("GetLyrics failed: %v", err)
	} else if len(resp.Lyrics) == 0 {
		utils.LogInfof("GetLyrics completed without lyrics")
	}

	logs := utils.StopLogCapture()
	analytics.ReportLyricsLookup(input, resp, err, failure, success, mobileGUIDVariant, startedAt, duration, logs)
	print(fmt.Sprintf("%v", err))

	if err != nil {
		return resp, fmt.Errorf("%s: %w", utils.PluginName, err)
	}

	lyrics_writer.WriteLyrics(input.Track, resp, format)

	return resp, nil
}
