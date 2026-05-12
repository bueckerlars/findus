package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic in HTTP handler",
						slog.String("event", "panic"),
						slog.String("http.method", r.Method),
						slog.String("http.path", r.URL.Path),
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
