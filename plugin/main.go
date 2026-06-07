package main

import (
	"github.com/Myzel394/navidrome-shazam-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

type plugin struct{}

//go:wasmexport nd_on_init
func ndOnInit() int32 {
	utils.LogInfof("initialized")
	return 0
}

func init() {
	lyrics.Register(&plugin{})
}

func main() {}
