package main

import (
	"fmt"

	"github.com/Myzel394/navidrome-shazam-plugin/plugin/shazam"
	"github.com/Myzel394/navidrome-shazam-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func (p *plugin) GetLyrics(input lyrics.GetLyricsRequest) (lyrics.GetLyricsResponse, error) {
	resp, err := shazam.FetchLyrics(input)
	if err != nil {
		return resp, fmt.Errorf("%s%w", utils.LogPrefix, err)
	}
	return resp, nil
}
