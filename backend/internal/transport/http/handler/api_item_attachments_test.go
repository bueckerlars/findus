package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/config"
	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/handler"
	"findus/backend/internal/transport/http/middleware"
)

// 1×1 transparent PNG (valid image/jpeg sniff target: image/png).
var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func TestItemAttachmentsAuthMatrix(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	users := sqlite.NewUserRepo(db)
	locs := sqlite.NewLocationRepo(db)
	items := sqlite.NewItemRepo(db)
	labels := sqlite.NewLabelRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)
	invites := sqlite.NewInviteRepo(db)
	settings := sqlite.NewSettingsRepo(db)
	attachments := sqlite.NewItemAttachmentRepo(db)
	inv := &service.Inventory{
		Locations: locs, Items: items, Labels: labels, Templates: templates, Settings: settings,
		ItemAttachments: attachments,
	}
	auth := &service.Auth{Users: users, Settings: settings, Invites: invites}
	jwtSecret := []byte("test-secret-test-secret-test-secret")

	srv := &handler.Server{
		Log:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		Config:    config.Config{DataDir: dir, CookieSecure: false},
		Users:     users,
		Locs:      locs,
		Items:     items,
		Labels:    labels,
		Templates: templates,
		Settings:  settings,
		Auth:      auth,
		Inventory: inv,
		JWTSecret: jwtSecret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", srv.APIAuthRegister)
	mux.Handle("GET /api/items/{id}/attachments", middleware.RequireAuth(http.HandlerFunc(srv.APIItemAttachmentsList)))
	mux.Handle("POST /api/items/{id}/attachments", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(srv.APIItemAttachmentCreate))))
	mux.Handle("PATCH /api/items/{id}/attachments/{attachmentId}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(srv.APIItemAttachmentPatch))))
	mux.Handle("DELETE /api/items/{id}/attachments/{attachmentId}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(srv.APIItemAttachmentDelete))))
	mux.Handle("GET /items/{id}/attachments/{attachmentId}", middleware.RequireAuth(http.HandlerFunc(srv.ItemAttachmentDownload)))
	h := middleware.Chain(mux,
		middleware.AuthOptional(users, jwtSecret, false),
		middleware.CSRF(false),
	)

	csrf := bootstrapCSRF(t, h)

	// First user = admin
	regAdmin := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(
		`{"username":"admin","email":"admin@example.com","password":"password1234","invite":""}`,
	)))
	regAdmin.Header.Set("Content-Type", "application/json")
	regAdmin.Header.Set("X-CSRF-Token", csrf)
	regAdmin.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	rrA := httptest.NewRecorder()
	h.ServeHTTP(rrA, regAdmin)
	require.Equal(t, http.StatusOK, rrA.Code, rrA.Body.String())
	adminSession := cookieValue(rrA.Result().Cookies(), "findus_session")
	require.NotEmpty(t, adminSession)

	var locID string
	require.NoError(t, db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID))
	now := time.Now().UTC().Format(time.RFC3339Nano)
	itemID := "it_attach_test"
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at)
		VALUES (?, 'Doc item', '', ?, 'documents', '{}', '{}', NULL, 'tok_attach_1', ?, ?)`,
		itemID, locID, now, now)
	require.NoError(t, err)

	// Upload as admin
	attID := postAttachment(t, h, csrf, adminSession, itemID, "Warranty", tinyPNG)
	require.NotEmpty(t, attID)

	// Patch title
	rrP := httptest.NewRecorder()
	patchBody := []byte(`{"title":"Updated title"}`)
	reqP := httptest.NewRequest(http.MethodPatch, "/api/items/"+itemID+"/attachments/"+attID, bytes.NewReader(patchBody))
	reqP.Header.Set("Content-Type", "application/json")
	reqP.Header.Set("X-CSRF-Token", csrf)
	reqP.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	reqP.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	h.ServeHTTP(rrP, reqP)
	require.Equal(t, http.StatusOK, rrP.Code, rrP.Body.String())

	// List
	rrL := httptest.NewRecorder()
	reqL := httptest.NewRequest(http.MethodGet, "/api/items/"+itemID+"/attachments", nil)
	reqL.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	h.ServeHTTP(rrL, reqL)
	require.Equal(t, http.StatusOK, rrL.Code)
	var listOut struct {
		Attachments []struct {
			ID               string `json:"ID"`
			DownloadURL      string `json:"DownloadURL"`
			OriginalFilename string `json:"OriginalFilename"`
			Title            string `json:"Title"`
		} `json:"attachments"`
	}
	require.NoError(t, json.NewDecoder(rrL.Body).Decode(&listOut))
	require.Len(t, listOut.Attachments, 1)
	require.Equal(t, attID, listOut.Attachments[0].ID)
	require.Equal(t, "Updated title", listOut.Attachments[0].Title)

	// Download as admin
	rrD := httptest.NewRecorder()
	reqD := httptest.NewRequest(http.MethodGet, "/items/"+itemID+"/attachments/"+attID, nil)
	reqD.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	h.ServeHTTP(rrD, reqD)
	require.Equal(t, http.StatusOK, rrD.Code)
	require.Equal(t, "image/png", rrD.Header().Get("Content-Type"))
	require.Equal(t, len(tinyPNG), rrD.Body.Len())

	// Open registration + second user (non-admin)
	require.NoError(t, settings.Set(ctx, domain.SettingRegistrationMode, string(domain.RegistrationOpen)))
	regUser := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte(
		`{"username":"member","email":"member@example.com","password":"password1234","invite":""}`,
	)))
	regUser.Header.Set("Content-Type", "application/json")
	regUser.Header.Set("X-CSRF-Token", csrf)
	regUser.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	rrU := httptest.NewRecorder()
	h.ServeHTTP(rrU, regUser)
	require.Equal(t, http.StatusOK, rrU.Code, rrU.Body.String())
	userSession := cookieValue(rrU.Result().Cookies(), "findus_session")
	require.NotEmpty(t, userSession)

	// Member can list and download
	rrL2 := httptest.NewRecorder()
	reqL2 := httptest.NewRequest(http.MethodGet, "/api/items/"+itemID+"/attachments", nil)
	reqL2.AddCookie(&http.Cookie{Name: "findus_session", Value: userSession})
	h.ServeHTTP(rrL2, reqL2)
	require.Equal(t, http.StatusOK, rrL2.Code)

	rrD2 := httptest.NewRecorder()
	reqD2 := httptest.NewRequest(http.MethodGet, "/items/"+itemID+"/attachments/"+attID, nil)
	reqD2.AddCookie(&http.Cookie{Name: "findus_session", Value: userSession})
	h.ServeHTTP(rrD2, reqD2)
	require.Equal(t, http.StatusOK, rrD2.Code)

	// Member cannot upload
	rrForbidden := httptest.NewRecorder()
	reqF := newAttachmentRequest(t, csrf, userSession, itemID, "", tinyPNG)
	h.ServeHTTP(rrForbidden, reqF)
	require.Equal(t, http.StatusForbidden, rrForbidden.Code)

	// Member cannot delete
	rrDelDeny := httptest.NewRecorder()
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/items/"+itemID+"/attachments/"+attID, nil)
	reqDel.Header.Set("X-CSRF-Token", csrf)
	reqDel.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	reqDel.AddCookie(&http.Cookie{Name: "findus_session", Value: userSession})
	h.ServeHTTP(rrDelDeny, reqDel)
	require.Equal(t, http.StatusForbidden, rrDelDeny.Code)

	// Wrong item id in download path
	rr404 := httptest.NewRecorder()
	req404 := httptest.NewRequest(http.MethodGet, "/items/wrong-id/attachments/"+attID, nil)
	req404.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	h.ServeHTTP(rr404, req404)
	require.Equal(t, http.StatusNotFound, rr404.Code)

	// Admin delete
	rrDel := httptest.NewRecorder()
	reqDelOk := httptest.NewRequest(http.MethodDelete, "/api/items/"+itemID+"/attachments/"+attID, nil)
	reqDelOk.Header.Set("X-CSRF-Token", csrf)
	reqDelOk.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	reqDelOk.AddCookie(&http.Cookie{Name: "findus_session", Value: adminSession})
	h.ServeHTTP(rrDel, reqDelOk)
	require.Equal(t, http.StatusOK, rrDel.Code)
}

func bootstrapCSRF(t *testing.T, h http.Handler) string {
	t.Helper()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/bootstrap", nil))
	csrf := cookieValue(rr.Result().Cookies(), "findus_csrf")
	require.NotEmpty(t, csrf)
	return csrf
}

func postAttachment(t *testing.T, h http.Handler, csrf, session, itemID, title string, file []byte) string {
	t.Helper()
	req := newAttachmentRequest(t, csrf, session, itemID, title, file)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())
	var out struct {
		Attachment struct {
			ID string `json:"ID"`
		} `json:"attachment"`
	}
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&out))
	return out.Attachment.ID
}

func newAttachmentRequest(t *testing.T, csrf, session, itemID, title string, file []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "tiny.png")
	require.NoError(t, err)
	_, err = fw.Write(file)
	require.NoError(t, err)
	if title != "" {
		require.NoError(t, w.WriteField("title", title))
	}
	require.NoError(t, w.Close())
	req := httptest.NewRequest(http.MethodPost, "/api/items/"+itemID+"/attachments", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "findus_csrf", Value: csrf})
	req.AddCookie(&http.Cookie{Name: "findus_session", Value: session})
	return req
}
