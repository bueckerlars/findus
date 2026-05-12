package logger

import (
	"log/slog"
	"os"
	"strings"
)

const defaultService = "findus"

// Options configures the application slog.Logger.
type Options struct {
	Level   string // debug, info, warn, error
	Format  string // text (default), json for structured logs
	Service string // logical service name for log records
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New returns a slog.Logger with text (human-readable) or JSON output and optional default attributes.
func New(opts Options) *slog.Logger {
	lvl := parseLevel(opts.Level)
	service := strings.TrimSpace(opts.Service)
	if service == "" {
		service = defaultService
	}
	addSource := lvl == slog.LevelDebug
	handlerOpts := &slog.HandlerOptions{
		Level:     lvl,
		AddSource: addSource,
	}

	var h slog.Handler
	switch strings.ToLower(strings.TrimSpace(opts.Format)) {
	case "json", "structured":
		h = slog.NewJSONHandler(os.Stdout, handlerOpts)
	default:
		h = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	return slog.New(h).With(slog.String("service", service))
}
