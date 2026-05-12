package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"findus/backend/internal/domain"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/middleware"
)

func (s *Server) APIAdminHome(w http.ResponseWriter, _ *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) APIAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	us, err := s.Admin.ListUsers(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	list, err := s.Admin.ListInvites(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	mode, err := s.Admin.GetRegistrationMode(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	apiUsers := make([]apiUser, 0, len(us))
	for i := range us {
		apiUsers = append(apiUsers, apiUserFrom(&us[i]))
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"users": apiUsers, "invites": list, "registration_mode": string(mode),
	})
}

type apiAdminCreateUserReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (s *Server) APIAdminUsersCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminCreateUserReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var role domain.Role
	switch req.Role {
	case "admin":
		role = domain.RoleAdmin
	case "user":
		role = domain.RoleUser
	default:
		s.writeJSONError(w, http.StatusBadRequest, "bad role")
		return
	}
	if _, err := s.Admin.CreateUser(ctx, req.Username, req.Email, req.Password, role); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/users"})
}

type apiAdminRoleReq struct {
	Role string `json:"role"`
}

func (s *Server) APIAdminUserRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiAdminRoleReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	role := domain.Role(req.Role)
	if role != domain.RoleAdmin && role != domain.RoleUser {
		s.writeJSONError(w, http.StatusBadRequest, "bad role")
		return
	}
	_ = s.Admin.SetUserRole(ctx, id, role)
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/users"})
}

type apiAdminActiveReq struct {
	Active bool `json:"active"`
}

func (s *Server) APIAdminUserActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiAdminActiveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_ = s.Admin.SetUserActive(ctx, id, req.Active)
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/users"})
}

type apiAdminInviteReq struct {
	Role     string `json:"role"`
	TTLHours int    `json:"ttl_hours"`
}

func (s *Server) APIAdminInvitesCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminInviteReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	role := domain.Role(req.Role)
	if role != domain.RoleAdmin && role != domain.RoleUser {
		s.writeJSONError(w, http.StatusBadRequest, "bad role")
		return
	}
	ttl := time.Duration(req.TTLHours) * time.Hour
	u, ok := middleware.User(ctx)
	if !ok {
		s.writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if _, err := s.Admin.CreateInvite(ctx, u.ID, role, ttl); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/users"})
}

type apiAdminSettingsReq struct {
	Mode string `json:"mode"`
}

func (s *Server) APIAdminSettingsRegistration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminSettingsReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	m, ok := domain.ParseRegistrationMode(req.Mode)
	if !ok {
		s.writeJSONError(w, http.StatusBadRequest, "bad mode")
		return
	}
	if err := s.Admin.SetRegistrationMode(ctx, m); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/users"})
}

type apiTemplateListRow struct {
	Template        domain.ItemTemplate `json:"template"`
	Count           int64               `json:"count"`
	HasAlternate    bool                `json:"has_alternate"`
	FallbackDisplay string              `json:"fallback_display"`
}

func (s *Server) APIAdminTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.Inventory.ListItemTemplates(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	rows := make([]apiTemplateListRow, 0, len(list))
	for _, t := range list {
		n, _ := s.Items.CountByTemplateType(ctx, t.ID)
		_, fdn, ok := service.ReassignTargetForDelete(list, t.ID)
		rows = append(rows, apiTemplateListRow{Template: t, Count: n, HasAlternate: ok, FallbackDisplay: fdn})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"rows": rows})
}

type apiAdminTemplateSaveReq struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	FieldsJSON  string `json:"fields_json"`
	SortOrder   int    `json:"sort_order"`
}

func (s *Server) APIAdminTemplateCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminTemplateSaveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.Inventory.CreateItemTemplate(ctx, req.ID, req.DisplayName, []byte(req.FieldsJSON), req.SortOrder); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/templates"})
}

func (s *Server) APIAdminTemplateGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	t, err := s.Inventory.GetItemTemplate(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	init := string(t.FieldsJSON)
	if strings.TrimSpace(init) == "" {
		init = "[]"
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"template": t, "fields_json": init})
}

func (s *Server) APIAdminTemplateUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiAdminTemplateSaveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.Inventory.UpdateItemTemplate(ctx, id, req.DisplayName, []byte(req.FieldsJSON), req.SortOrder); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/templates"})
}

func (s *Server) APIAdminTemplateDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Inventory.DeleteItemTemplate(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/templates"})
}

// APIAdminTemplateNewEmpty returns defaults for new template form.
func (s *Server) APIAdminTemplateNewEmpty(w http.ResponseWriter, r *http.Request) {
	_ = r
	s.writeJSON(w, http.StatusOK, map[string]any{"sort_order": 10, "fields_json": "[]"})
}
