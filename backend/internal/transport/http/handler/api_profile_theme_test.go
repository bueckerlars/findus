package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/config"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/handler"
	"findus/backend/internal/transport/http/middleware"
)

func TestAPIProfileThemePatch(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	users := sqlite.NewUserRepo(db)
	settings := sqlite.NewSettingsRepo(db)
	invites := sqlite.NewInviteRepo(db)
	auth := &service.Auth{Users: users, Settings: settings, Invites: invites}

	jwtSecret := []byte("test-secret-test-secret-test-secret")

	srv := &handler.Server{
		Log:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		Config:    config.Config{DataDir: dir, CookieSecure: false},
		Users:     users,
		Settings:  settings,
		Invites:   invites,
		Auth:      auth,
		JWTSecret: jwtSecret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", srv.APIAuthRegister)
	mux.Handle("GET /api/me", middleware.RequireAuth(http.HandlerFunc(srv.APIMe)))
	mux.Handle("PATCH /api/profile/theme", middleware.RequireAuth(http.HandlerFunc(srv.APIProfileThemePatch)))
	h := middleware.Chain(mux,
		middleware.AuthOptional(users, jwtSecret, false),
		middleware.CSRF(false),
	)

	rr0 := httptest.NewRecorder()
	h.ServeHTTP(rr0, httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil))
	csrf := cookieValue(rr0.Result().Cookies(), "findus_csrf")
	require.NotEmpty(t, csrf)

	body := []byte(`{"username":"bob","email":"bob@example.com","password":"password1234","invite":""}`)
	reqReg := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	reqReg.Header.Set("Content-Type", "application/json")
	reqReg.Header.Set("X-CSRF-Token", csrf)
	reqReg.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, reqReg)
	require.Equal(t, http.StatusOK, rr1.Code, rr1.Body.String())

	session := cookieValue(rr1.Result().Cookies(), "findus_session")
	require.NotEmpty(t, session)

	patch := func(theme string) *httptest.ResponseRecorder {
		b, _ := json.Marshal(map[string]string{"theme": theme})
		req := httptest.NewRequest(http.MethodPatch, "/api/profile/theme", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF-Token", csrf)
		req.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
		req.AddCookie(&http.Cookie{Name: "findus_session", Value: session})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		return rr
	}

	rrBad := patch("not-a-theme")
	require.Equal(t, http.StatusBadRequest, rrBad.Code)

	rrOk := patch("forest")
	require.Equal(t, http.StatusOK, rrOk.Code)
	var patchOut struct {
		User struct {
			ID    string `json:"id"`
			Theme string `json:"theme"`
		} `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rrOk.Body).Decode(&patchOut))
	require.Equal(t, "forest", patchOut.User.Theme)

	rrMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	reqMe.Header.Set("X-CSRF-Token", csrf)
	reqMe.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	reqMe.AddCookie(&http.Cookie{Name: "findus_session", Value: session})
	h.ServeHTTP(rrMe, reqMe)
	require.Equal(t, http.StatusOK, rrMe.Code)
	var meOut struct {
		User struct {
			ID    string `json:"id"`
			Theme string `json:"theme"`
		} `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rrMe.Body).Decode(&meOut))
	require.Equal(t, patchOut.User.ID, meOut.User.ID)
	require.Equal(t, "forest", meOut.User.Theme)

	rrNight := patch("night")
	require.Equal(t, http.StatusOK, rrNight.Code)
	var nightOut struct {
		User struct {
			Theme string `json:"theme"`
		} `json:"user"`
	}
	require.NoError(t, json.NewDecoder(rrNight.Body).Decode(&nightOut))
	require.Equal(t, "night", nightOut.User.Theme)
}

func cookieValue(cookies []*http.Cookie, name string) string {
	for _, c := range cookies {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}
