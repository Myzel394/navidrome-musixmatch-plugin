package utils

import (
	"fmt"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

func LogInfof(format string, args ...any) {
	pdk.Log(pdk.LogInfo, fmt.Sprintf(LogPrefix+format, args...))
}

func LogErrorf(format string, args ...any) {
	pdk.Log(pdk.LogError, fmt.Sprintf(LogPrefix+format, args...))
}
