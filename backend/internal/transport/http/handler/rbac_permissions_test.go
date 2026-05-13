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
	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/handler"
	"findus/backend/internal/transport/http/middleware"
)

func TestRBACItemsWriteWithoutLocationsWrite(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	users := sqlite.NewUserRepo(db)
	groups := sqlite.NewGroupRepo(db)
	locs := sqlite.NewLocationRepo(db)
	items := sqlite.NewItemRepo(db)
	itemQRTokens := sqlite.NewItemQRTokenReservationRepo(db)
	labels := sqlite.NewLabelRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)
	attachments := sqlite.NewItemAttachmentRepo(db)
	settings := sqlite.NewSettingsRepo(db)
	invites := sqlite.NewInviteRepo(db)
	inv := &service.Inventory{
		Locations: locs, Items: items, ItemQRTokens: itemQRTokens, Labels: labels, Templates: templates, Settings: settings,
		ItemAttachments: attachments,
	}
	auth := &service.Auth{Users: users, Settings: settings, Invites: invites}
	adminSvc := &service.Admin{Users: users, Settings: settings, Invites: invites, Groups: groups}
	jwtSecret := []byte("test-secret-test-secret-test-secret")

	srv := &handler.Server{
		Log:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		Config:       config.Config{DataDir: dir, CookieSecure: false},
		Users:        users,
		Groups:       groups,
		Locs:         locs,
		Items:        items,
		ItemQRTokens: itemQRTokens,
		Labels:       labels,
		Templates:    templates,
		Settings:     settings,
		Invites:      invites,
		Auth:         auth,
		Admin:        adminSvc,
		Inventory:    inv,
		JWTSecret:    jwtSecret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", srv.APIAuthRegister)
	mux.Handle("GET /api/items/new", middleware.RequireAuth(middleware.RequirePermission(domain.PermItemsWrite)(http.HandlerFunc(srv.APIItemNew))))
	mux.Handle("GET /api/locations/new", middleware.RequireAuth(middleware.RequirePermission(domain.PermLocationsWrite)(http.HandlerFunc(srv.APILocationNew))))
	mux.Handle("POST /api/admin/groups", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(srv.APIAdminGroupsCreate))))
	mux.Handle("POST /api/admin/users/{id}/groups", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(srv.APIAdminUserGroupsSet))))

	h := middleware.Chain(mux,
		middleware.AuthOptional(users, groups, jwtSecret, false),
		middleware.CSRF(false),
	)

	csrf := bootstrapCSRF(t, h)

	regAdmin := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(
		`{"username":"owner","email":"owner@example.com","password":"password1234","invite":""}`,
	)))
	regAdmin.Header.Set("Content-Type", "application/json")
	regAdmin.Header.Set("X-CSRF-Token", csrf)
	regAdmin.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	rrA := httptest.NewRecorder()
	h.ServeHTTP(rrA, regAdmin)
	require.Equal(t, http.StatusOK, rrA.Code, rrA.Body.String())
	adminSession := cookieValue(rrA.Result().Cookies(), "findus_session")
	require.NotEmpty(t, adminSession)

	require.NoError(t, settings.Set(ctx, string(domain.SettingRegistrationMode), string(domain.RegistrationOpen)))

	regMember := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(
		`{"username":"editor","email":"editor@example.com","password":"password1234","invite":""}`,
	)))
	regMember.Header.Set("Content-Type", "application/json")
	regMember.Header.Set("X-CSRF-Token", csrf)
	regMember.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	rrM := httptest.NewRecorder()
	h.ServeHTTP(rrM, regMember)
	require.Equal(t, http.StatusOK, rrM.Code, rrM.Body.String())
	memberSession := cookieValue(rrM.Result().Cookies(), "findus_session")
	require.NotEmpty(t, memberSession)

	var memberID string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id FROM users WHERE username = 'editor'`).Scan(&memberID))

	createGroupBody := []byte(`{"name":"Items only","permissions":["items.write"]}`)
	reqG := httptest.NewRequest(http.MethodPost, "/api/admin/groups", bytes.NewReader(createGroupBody))
	reqG.Header.Set("Content-Type", "application/json")
	reqG.Header.Set("X-CSRF-Token", csrf)
	reqG.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	reqG.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	rrG := httptest.NewRecorder()
	h.ServeHTTP(rrG, reqG)
	require.Equal(t, http.StatusOK, rrG.Code, rrG.Body.String())
	var groupOut struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(rrG.Body).Decode(&groupOut))
	require.NotEmpty(t, groupOut.ID)

	assignBody, _ := json.Marshal(map[string]any{"group_ids": []string{groupOut.ID}})
	reqAs := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+memberID+"/groups", bytes.NewReader(assignBody))
	reqAs.Header.Set("Content-Type", "application/json")
	reqAs.Header.Set("X-CSRF-Token", csrf)
	reqAs.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	reqAs.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	rrAs := httptest.NewRecorder()
	h.ServeHTTP(rrAs, reqAs)
	require.Equal(t, http.StatusOK, rrAs.Code, rrAs.Body.String())

	reqItems := httptest.NewRequest(http.MethodGet, "/api/items/new", nil)
	reqItems.AddCookie(&http.Cookie{Name: "findus_session", Value: memberSession})
	rrItems := httptest.NewRecorder()
	h.ServeHTTP(rrItems, reqItems)
	require.Equal(t, http.StatusOK, rrItems.Code, rrItems.Body.String())

	reqLoc := httptest.NewRequest(http.MethodGet, "/api/locations/new", nil)
	reqLoc.AddCookie(&http.Cookie{Name: "findus_session", Value: memberSession})
	rrLoc := httptest.NewRecorder()
	h.ServeHTTP(rrLoc, reqLoc)
	require.Equal(t, http.StatusForbidden, rrLoc.Code)
}
