package handler

import (
	"io/fs"
	"net/http"

	"findus/backend/internal/transport/http/middleware"
	"findus/frontend"
)

// MountApp serves Vite build from embedded dist/ (index.html + assets/).
func (s *Server) MountApp(mux *http.ServeMux) error {
	sub, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		return err
	}
	assets, err := fs.Sub(sub, "assets")
	if err != nil {
		return err
	}
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	return nil
}

// ServeSPA returns index.html for Vue Router (client-side routes).
func (s *Server) ServeSPA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	b, err := fs.ReadFile(frontend.Assets, "dist/index.html")
	if err != nil {
		http.Error(w, "spa not built", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

// MountAPI registers JSON API routes (caller wraps with auth/admin as needed).
func (s *Server) MountAPI(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/bootstrap", s.APIBootstrap)
	mux.HandleFunc("GET /api/auth/username-available", s.APIAuthUsernameAvailable)
	mux.HandleFunc("POST /api/auth/login", s.APIAuthLogin)
	mux.HandleFunc("POST /api/auth/register", s.APIAuthRegister)

	mux.Handle("GET /api/me", middleware.RequireAuth(http.HandlerFunc(s.APIMe)))
	mux.Handle("POST /api/auth/logout", middleware.RequireAuth(http.HandlerFunc(s.APIAuthLogout)))

	mux.Handle("GET /api/home", middleware.RequireAuth(http.HandlerFunc(s.APIHome)))
	mux.Handle("GET /api/profile", middleware.RequireAuth(http.HandlerFunc(s.APIProfileGet)))
	mux.Handle("POST /api/profile", middleware.RequireAuth(http.HandlerFunc(s.APIProfilePost)))
	mux.Handle("PATCH /api/profile/theme", middleware.RequireAuth(http.HandlerFunc(s.APIProfileThemePatch)))

	mux.Handle("GET /api/locations", middleware.RequireAuth(http.HandlerFunc(s.APILocationsList)))
	mux.Handle("GET /api/locations/new", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILocationNew))))
	mux.Handle("POST /api/locations", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILocationCreate))))
	mux.Handle("GET /api/locations/{id}", middleware.RequireAuth(http.HandlerFunc(s.APILocationGet)))
	mux.Handle("GET /api/locations/{id}/edit", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILocationEdit))))
	mux.Handle("POST /api/locations/{id}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILocationUpdate))))
	mux.Handle("POST /api/locations/{id}/delete", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILocationDelete))))

	mux.Handle("GET /api/items", middleware.RequireAuth(http.HandlerFunc(s.APIItemsList)))
	mux.Handle("GET /api/items/new", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemNew))))
	mux.Handle("GET /api/items/new/fields", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemNewFields))))
	mux.Handle("POST /api/items", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemCreate))))
	mux.Handle("GET /api/items/{id}", middleware.RequireAuth(http.HandlerFunc(s.APIItemGet)))
	mux.Handle("GET /api/items/{id}/edit", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemEdit))))
	mux.Handle("POST /api/items/{id}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemUpdate))))
	mux.Handle("POST /api/items/{id}/delete", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemDelete))))
	mux.Handle("GET /api/items/{id}/attachments", middleware.RequireAuth(http.HandlerFunc(s.APIItemAttachmentsList)))
	mux.Handle("POST /api/items/{id}/attachments", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemAttachmentCreate))))
	mux.Handle("PATCH /api/items/{id}/attachments/{attachmentId}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemAttachmentPatch))))
	mux.Handle("DELETE /api/items/{id}/attachments/{attachmentId}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIItemAttachmentDelete))))

	mux.Handle("GET /api/search", middleware.RequireAuth(http.HandlerFunc(s.APISearch)))
	mux.Handle("GET /api/command-search", middleware.RequireAuth(http.HandlerFunc(s.CommandSearchGet)))

	mux.Handle("GET /api/labels", middleware.RequireAuth(http.HandlerFunc(s.APILabelsList)))
	mux.Handle("GET /api/labels/new", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILabelNew))))
	mux.Handle("POST /api/labels", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILabelCreate))))
	mux.Handle("GET /api/labels/{id}/edit", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILabelGet))))
	mux.Handle("POST /api/labels/{id}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILabelUpdate))))
	mux.Handle("POST /api/labels/{id}/delete", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APILabelDelete))))

	mux.Handle("GET /api/admin", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminHome))))
	mux.Handle("GET /api/admin/users", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminUsers))))
	mux.Handle("POST /api/admin/users", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminUsersCreate))))
	mux.Handle("POST /api/admin/users/{id}/role", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminUserRole))))
	mux.Handle("POST /api/admin/users/{id}/active", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminUserActive))))
	mux.Handle("POST /api/admin/invites", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminInvitesCreate))))
	mux.Handle("GET /api/admin/settings/registration", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminSettingsRegistrationGet))))
	mux.Handle("POST /api/admin/settings/registration", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminSettingsRegistration))))
	mux.Handle("GET /api/admin/settings/item-ids", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminSettingsItemIDsGet))))
	mux.Handle("POST /api/admin/settings/item-ids", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminSettingsItemIDsPost))))
	mux.Handle("POST /api/admin/labels/generate", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminLabelsGenerate))))
	mux.Handle("GET /api/admin/inventory-export", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminInventoryExport))))
	mux.Handle("POST /api/admin/inventory-import", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminInventoryImport))))

	mux.Handle("GET /api/admin/templates", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminTemplates))))
	mux.Handle("GET /api/admin/templates/new", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminTemplateNewEmpty))))
	mux.Handle("POST /api/admin/templates", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminTemplateCreate))))
	mux.Handle("GET /api/admin/templates/{id}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminTemplateGet))))
	mux.Handle("POST /api/admin/templates/{id}", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminTemplateUpdate))))
	mux.Handle("POST /api/admin/templates/{id}/delete", middleware.RequireAuth(middleware.RequireAdmin(http.HandlerFunc(s.APIAdminTemplateDelete))))
}

func (s *Server) Handler() (http.Handler, error) {
	mux := http.NewServeMux()
	if err := s.MountApp(mux); err != nil {
		return nil, err
	}
	mux.HandleFunc("GET /healthz", s.Health)
	s.MountAPI(mux)

	mux.HandleFunc("GET /q/{token}", s.QRedirect)

	auth := func(h http.HandlerFunc) http.Handler { return middleware.RequireAuth(h) }
	admin := func(h http.HandlerFunc) http.Handler {
		return middleware.RequireAuth(middleware.RequireAdmin(h))
	}

	mux.Handle("GET /profile/photo", auth(s.ProfilePhoto))
	mux.Handle("GET /items/{id}/photo", auth(s.ItemPhoto))
	mux.Handle("GET /items/{id}/attachments/{attachmentId}", auth(s.ItemAttachmentDownload))
	mux.Handle("GET /items/{id}/qr.png", auth(s.ItemQR))
	mux.Handle("GET /locations/{id}/qr.png", auth(s.LocationQR))
	mux.Handle("GET /admin/backup.zip", admin(s.AdminBackup))

	mux.Handle("GET /{path...}", http.HandlerFunc(s.ServeSPA))

	return middleware.Chain(mux,
		middleware.AuthOptional(s.Users, s.JWTSecret, s.Config.CookieSecure),
		middleware.CSRF(s.Config.CookieSecure),
		middleware.RequestLog(s.Log),
		middleware.Recover(s.Log),
	), nil
}
