package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"findus/backend/internal/domain"
	"findus/backend/internal/transport/http/middleware"
)

type apiUser struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	IsActive   bool      `json:"is_active"`
	Theme      string    `json:"theme"`
	AvatarPath *string   `json:"avatar_path,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func apiUserFrom(u *domain.User) apiUser {
	if u == nil {
		return apiUser{}
	}
	return apiUser{
		ID: u.ID, Username: u.Username, Email: u.Email,
		Role: string(u.Role), IsActive: u.IsActive,
		Theme:      domain.NormalizeUITheme(u.UITheme),
		AvatarPath: u.AvatarPath,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

// APIMe returns the current session user or 401.
func (s *Server) APIMe(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.User(r.Context())
	if !ok {
		s.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if fresh, err := s.Users.GetByID(r.Context(), u.ID); err == nil {
		u = fresh
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"user": apiUserFrom(u)})
}

type apiBootstrap struct {
	UserCount        int64  `json:"user_count"`
	RegistrationMode string `json:"registration_mode"`
	HasMode          bool   `json:"has_registration_mode"`
}

// APIBootstrap exposes registration-related flags for the register page (no auth required).
func (s *Server) APIBootstrap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	n, err := s.Users.Count(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	mode, hasMode, _ := s.AuthSettingsMode(ctx)
	out := apiBootstrap{UserCount: n, HasMode: hasMode}
	if hasMode {
		out.RegistrationMode = string(mode)
	}
	s.writeJSON(w, http.StatusOK, out)
}

type apiLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Next     string `json:"next"`
}

// APIAuthLogin sets session cookie; JSON body.
func (s *Server) APIAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req apiLoginReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	u, err := s.Auth.Login(r.Context(), strings.TrimSpace(req.Username), req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			s.writeJSONError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		s.Log.Error("login", "err", err)
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	if err := middleware.IssueSessionCookie(w, s.JWTSecret, u, s.Config.CookieSecure); err != nil {
		s.Log.Error("issue cookie", "err", err)
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	next := strings.TrimSpace(req.Next)
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		next = "/"
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": next})
}

// APIAuthLogout clears session (auth required by mux).
func (s *Server) APIAuthLogout(w http.ResponseWriter, r *http.Request) {
	middleware.ClearSessionCookie(w, s.Config.CookieSecure)
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/login"})
}

type apiRegisterReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Invite   string `json:"invite"`
}

// APIAuthRegister creates account and session when allowed.
func (s *Server) APIAuthRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiRegisterReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	n, err := s.Users.Count(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	invite := strings.TrimSpace(req.Invite)
	mode, hasMode, _ := s.AuthSettingsMode(ctx)
	if n > 0 {
		if !hasMode {
			mode = domain.RegistrationAdminOnly
		}
		switch mode {
		case domain.RegistrationAdminOnly:
			s.writeJSONError(w, http.StatusForbidden, "Registration is closed.")
			return
		case domain.RegistrationInvite:
			if invite == "" {
				s.writeJSONError(w, http.StatusForbidden, "invite token required")
				return
			}
			if _, err := s.Invites.GetByToken(ctx, invite); err != nil {
				s.writeJSONError(w, http.StatusForbidden, "invalid invite")
				return
			}
		}
	}
	u, err := s.Auth.Register(ctx, req.Username, req.Email, req.Password, invite)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRegistrationClosed):
			s.writeJSONError(w, http.StatusForbidden, "Registration is closed.")
		case errors.Is(err, domain.ErrInvalidInvite):
			s.writeJSONError(w, http.StatusForbidden, "Invalid or expired invite.")
		default:
			if errors.Is(err, domain.ErrValidation) {
				s.writeJSONError(w, http.StatusBadRequest, err.Error())
				return
			}
			s.Log.Error("register", "err", err)
			s.writeJSONError(w, http.StatusInternalServerError, "Could not register.")
		}
		return
	}
	if err := middleware.IssueSessionCookie(w, s.JWTSecret, u, s.Config.CookieSecure); err != nil {
		s.Log.Error("issue cookie", "err", err)
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/"})
}
