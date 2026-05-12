package middleware

import "net/http"

// Chain wraps h with middlewares in order: the first middleware in ms is the outermost.
func Chain(h http.Handler, ms ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range ms {
		h = m(h)
	}
	return h
}
