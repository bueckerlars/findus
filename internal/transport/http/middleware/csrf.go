package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	csrfCookieName = "findus_csrf"
	csrfHeaderName = "X-CSRF-Token"
	csrfFormName   = "csrf_token"
)

type csrfCtxKeyType struct{}

var csrfCtxKey = csrfCtxKeyType{}

func CSRFToken(ctx context.Context) string {
	v, _ := ctx.Value(csrfCtxKey).(string)
	return v
}

func CSRF(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkipCSRF(r) {
				next.ServeHTTP(w, r)
				return
			}
			c, err := r.Cookie(csrfCookieName)
			token := ""
			if err != nil || c.Value == "" {
				b := make([]byte, 16)
				_, _ = rand.Read(b)
				token = hex.EncodeToString(b)
				http.SetCookie(w, &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     "/",
					MaxAge:   86400 * 30,
					HttpOnly: false,
					Secure:   secure,
					SameSite: http.SameSiteLaxMode,
				})
			} else {
				token = c.Value
			}
			ctx := context.WithValue(r.Context(), csrfCtxKey, token)
			r = r.WithContext(ctx)
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)
				return
			}
			if err := parseRequestForCSRF(r); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			got := strings.TrimSpace(r.Header.Get(csrfHeaderName))
			if got == "" {
				got = strings.TrimSpace(r.FormValue(csrfFormName))
			}
			if got == "" || token == "" || !strings.EqualFold(got, token) {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func shouldSkipCSRF(r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		return true
	}
	if r.URL.Path == "/healthz" {
		return true
	}
	return false
}

func parseRequestForCSRF(r *http.Request) error {
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		return r.ParseMultipartForm(32 << 20)
	}
	return r.ParseForm()
}
