package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const slowRequestThreshold = 2 * time.Second

func skipAccessLog(path string) bool {
	switch {
	case path == "/healthz":
		return true
	case strings.HasPrefix(path, "/static/"):
		return true
	default:
		return false
	}
}

// RequestLog emits structured access records: skips health and static assets;
// successful fast requests log at Debug; client/server errors and slow requests log at Warn/Error.
func RequestLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipAccessLog(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			start := time.Now()
			wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(wrapped, r)
			dur := time.Since(start)
			attrs := []any{
				slog.String("event", "http_access"),
				slog.String("http.method", r.Method),
				slog.String("http.path", r.URL.Path),
				slog.Int("http.status_code", wrapped.status),
				slog.Int64("duration_ms", dur.Milliseconds()),
			}
			switch {
			case wrapped.status >= 500:
				log.Error("http request failed", attrs...)
			case wrapped.status >= 400:
				log.Warn("http client error", attrs...)
			case dur >= slowRequestThreshold:
				log.Info("http slow request", append(attrs, slog.Bool("slow", true))...)
			default:
				log.Debug("http request", attrs...)
			}
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
