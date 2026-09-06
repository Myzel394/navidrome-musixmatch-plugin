package musixmatch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func lyricsFromRichsync(call macroResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != utils.HTTPStatusOK {
		utils.LogInfof("desktop API: richsync lyrics unavailable because status was not successful status=%d body_present=%t", call.Message.Header.StatusCode, len(call.Message.Body) > 0)
		return lyrics.GetLyricsResponse{}, false
	}
	if len(call.Message.Body) == 0 {
		utils.LogInfof("desktop API: richsync lyrics unavailable because response body was empty status=%d", call.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, false
	}

	var body macroRichsyncBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		utils.LogInfof("desktop API: richsync lyrics unavailable because response could not be parsed body_bytes=%d", len(call.Message.Body))
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: richsync response body=%s", string(call.Message.Body)))
		return lyrics.GetLyricsResponse{}, false
	}
	if body.Richsync.Body == "" {
		utils.LogInfof("desktop API: richsync lyrics unavailable because lyrics body was empty")
		return lyrics.GetLyricsResponse{}, false
	}

	var lines []macroRichsyncLine
	if err := json.Unmarshal([]byte(body.Richsync.Body), &lines); err != nil {
		utils.LogErrorf("desktop API richsync parse failed: %v", err)
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: richsync payload body=%s", body.Richsync.Body))
		return lyrics.GetLyricsResponse{}, false
	}

	var builder strings.Builder
	for _, line := range lines {
		if line.Text == "" {
			builder.WriteByte('\n')
			continue
		}
		builder.WriteString(totalToLRC(line.Timestamp))
		builder.WriteString(line.Text)
		builder.WriteByte('\n')
	}
	if builder.Len() == 0 {
		utils.LogInfof("desktop API: richsync lyrics unavailable because no LRC lines were rendered parsed_lines=%d", len(lines))
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got richsync lyrics (%d lines LRC)", strings.Count(builder.String(), "\n"))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: body.Richsync.Language, Text: builder.String()}}}, true
}

func lyricsFromSubtitle(call macroResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != utils.HTTPStatusOK {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because status was not successful status=%d body_present=%t", call.Message.Header.StatusCode, len(call.Message.Body) > 0)
		return lyrics.GetLyricsResponse{}, false
	}
	if len(call.Message.Body) == 0 {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because response body was empty status=%d", call.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, false
	}

	var body macroSubtitleBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because response could not be parsed body_bytes=%d", len(call.Message.Body))
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: subtitle response body=%s", string(call.Message.Body)))
		return lyrics.GetLyricsResponse{}, false
	}
	if len(body.SubtitleList) == 0 {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because subtitle list was empty")
		return lyrics.GetLyricsResponse{}, false
	}

	subtitle := body.SubtitleList[0].Subtitle
	if subtitle.Body == "" {
		utils.LogInfof("desktop API: subtitle lyrics unavailable because lyrics body was empty list_count=%d", len(body.SubtitleList))
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got subtitle lyrics (%d lines LRC)", strings.Count(subtitle.Body, "\n"))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: subtitle.Language, Text: subtitle.Body}}}, true
}

func lyricsFromPlain(call macroResponse) (lyrics.GetLyricsResponse, bool) {
	if call.Message.Header.StatusCode != utils.HTTPStatusOK {
		utils.LogInfof("desktop API: plain lyrics unavailable because status was not successful status=%d body_present=%t", call.Message.Header.StatusCode, len(call.Message.Body) > 0)
		return lyrics.GetLyricsResponse{}, false
	}
	if len(call.Message.Body) == 0 {
		utils.LogInfof("desktop API: plain lyrics unavailable because response body was empty status=%d", call.Message.Header.StatusCode)
		return lyrics.GetLyricsResponse{}, false
	}

	var body macroLyricsBody
	if err := json.Unmarshal(call.Message.Body, &body); err != nil {
		utils.LogInfof("desktop API: plain lyrics unavailable because response could not be parsed body_bytes=%d", len(call.Message.Body))
		pdk.Log(pdk.LogDebug, fmt.Sprintf("desktop API: plain response body=%s", string(call.Message.Body)))
		return lyrics.GetLyricsResponse{}, false
	}
	if body.Lyrics.Body == "" {
		utils.LogInfof("desktop API: plain lyrics unavailable because lyrics body was empty")
		return lyrics.GetLyricsResponse{}, false
	}

	utils.LogInfof("desktop API: got plain lyrics (%d chars)", len(body.Lyrics.Body))
	return lyrics.GetLyricsResponse{Lyrics: []lyrics.LyricsText{{Lang: body.Lyrics.Language, Text: body.Lyrics.Body}}}, true
}
