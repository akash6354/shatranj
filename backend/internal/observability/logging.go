package observability

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger builds the process logger from SHATRANJ_LOG_LEVEL and
// SHATRANJ_LOG_FORMAT (text or json).
func NewLogger() *slog.Logger {
	level := new(slog.LevelVar)
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SHATRANJ_LOG_LEVEL"))) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}

	options := &slog.HandlerOptions{Level: level}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("SHATRANJ_LOG_FORMAT")), "json") {
		return slog.New(slog.NewJSONHandler(os.Stdout, options))
	}
	return slog.New(slog.NewTextHandler(os.Stdout, options))
}
