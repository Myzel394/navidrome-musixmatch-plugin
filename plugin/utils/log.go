package utils

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/navidrome/navidrome/plugins/pdk/go/pdk"
)

type CapturedLog struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

var (
	logCaptureMu      sync.Mutex
	logCaptureActive  bool
	logCaptureEntries []CapturedLog
	sensitiveLogRe    = regexp.MustCompile(`(?i)(usertoken|musixmatchUserToken|captcha_id|authorization)=([^&\s;]+)`)
)

func LogInfof(format string, args ...any) {
	msg := fmt.Sprintf(LogPrefix+format, args...)
	pdk.Log(pdk.LogInfo, msg)
	captureLog("info", msg)
}

func LogErrorf(format string, args ...any) {
	msg := fmt.Sprintf(LogPrefix+format, args...)
	pdk.Log(pdk.LogError, msg)
	captureLog("error", msg)
}

func StartLogCapture() {
	logCaptureMu.Lock()
	defer logCaptureMu.Unlock()

	logCaptureActive = true
	logCaptureEntries = nil
}

func StopLogCapture() []CapturedLog {
	logCaptureMu.Lock()
	defer logCaptureMu.Unlock()

	logs := make([]CapturedLog, len(logCaptureEntries))
	copy(logs, logCaptureEntries)
	logCaptureActive = false
	logCaptureEntries = nil

	return logs
}

func captureLog(level, msg string) {
	logCaptureMu.Lock()
	defer logCaptureMu.Unlock()

	if !logCaptureActive {
		return
	}

	logCaptureEntries = append(logCaptureEntries, CapturedLog{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   SanitizeAnalyticsText(msg),
	})
}

func SanitizeAnalyticsText(s string) string {
	return sensitiveLogRe.ReplaceAllStringFunc(s, func(match string) string {
		parts := strings.SplitN(match, "=", 2)
		return parts[0] + "=<redacted>"
	})
}
