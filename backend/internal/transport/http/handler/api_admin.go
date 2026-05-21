package handler

import (
	"encoding/json"
	"errors"
	"math"
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
	ugMap, _ := s.Admin.MapUserGroupIDs(ctx)
	for i := range apiUsers {
		if gids := ugMap[apiUsers[i].ID]; len(gids) > 0 {
			apiUsers[i].GroupIDs = gids
		}
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
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/settings"})
}

func (s *Server) APIAdminSettingsRegistrationGet(w http.ResponseWriter, r *http.Request) {
	_ = r
	ctx := r.Context()
	mode, err := s.Admin.GetRegistrationMode(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"registration_mode": string(mode)})
}

func (s *Server) APIAdminSettingsItemIDsGet(w http.ResponseWriter, r *http.Request) {
	_ = r
	ctx := r.Context()
	pol, err := s.Inventory.GetItemIDPolicy(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	n, err := s.Items.Count(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"policy": pol, "item_count": n})
}

type apiAdminItemIDPolicyReq struct {
	Prefix  string  `json:"prefix"`
	Width   float64 `json:"width"`
	NextSeq int64   `json:"next_seq,omitempty"` // accepted for forward-compatible clients; ignored here
}

func (s *Server) APIAdminSettingsItemIDsPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminItemIDPolicyReq
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	want := domain.ItemIDPolicy{Kind: domain.ItemIDKindSequential}
	want.Prefix = strings.TrimSpace(req.Prefix)
	if req.Width < 0 || req.Width > 64 {
		s.writeJSONError(w, http.StatusBadRequest, "bad width")
		return
	}
	want.Width = int(req.Width)
	cur, err := s.Inventory.GetItemIDPolicy(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	want = want.Normalize()
	cur = cur.Normalize()
	if cur.EffectiveKey() == want.EffectiveKey() {
		want.NextSeq = cur.NextSeq
	}
	if err := want.Validate(); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.Inventory.SetItemIDPolicy(ctx, s.Config.DataDir, want); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/settings"})
}

type apiAdminLabelGenerateReq struct {
	From     float64 `json:"from"`
	To       float64 `json:"to"`
	Cols     int     `json:"cols"`
	Rows     int     `json:"rows"`
	FullPage bool    `json:"full_page"`
}

func (s *Server) APIAdminLabelsGenerate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminLabelGenerateReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.From < 1 || req.To < 1 || math.Trunc(req.From) != req.From || math.Trunc(req.To) != req.To {
		s.writeJSONError(w, http.StatusBadRequest, "invalid range")
		return
	}
	pol, err := s.Inventory.GetItemIDPolicy(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	gen := service.LabelPDFGenerator{
		QR:           s.QR,
		MaxBatch:     service.DefaultLabelBatchLimit,
		Reservations: s.ItemQRTokens,
	}
	out, err := gen.Generate(ctx, service.LabelPDFInput{
		From:     int64(req.From),
		To:       int64(req.To),
		Policy:   pol,
		Cols:     req.Cols,
		Rows:     req.Rows,
		FullPage: req.FullPage,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writePDFResponse(w, out.Filename, out.PDF)
}

type apiAdminLocationLabelGenerateReq struct {
	LocationIDs []string `json:"location_ids"`
	Cols        int      `json:"cols"`
	Rows        int      `json:"rows"`
	FullPage    bool     `json:"full_page"`
}

func (s *Server) APIAdminLocationLabelsGenerate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminLocationLabelGenerateReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.LocationIDs) == 0 {
		s.writeJSONError(w, http.StatusBadRequest, "no locations selected")
		return
	}
	gen := service.LocationLabelPDFGenerator{
		QR:   s.QR,
		Locs: s.Locs,
	}
	out, err := gen.Generate(ctx, service.LocationLabelPDFInput{
		LocationIDs: req.LocationIDs,
		Cols:        req.Cols,
		Rows:        req.Rows,
		FullPage:    req.FullPage,
	})
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writePDFResponse(w, out.Filename, out.PDF)
}

func (s *Server) writePDFResponse(w http.ResponseWriter, filename string, data []byte) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="`+filename+`"`)
	w.Header().Set("X-Filename", filename)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
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
		n, err := s.Items.CountByTemplateType(ctx, t.ID)
		if err != nil {
			s.Log.Error("admin templates: count by template", "template_id", t.ID, "err", err)
			s.writeJSONError(w, http.StatusInternalServerError, "server error")
			return
		}
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

type apiAdminGroupRow struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	MemberCount int64    `json:"member_count"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func (s *Server) APIAdminGroupsList(w http.ResponseWriter, r *http.Request) {
	_ = r
	ctx := r.Context()
	gs, counts, err := s.Admin.ListPermissionGroups(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	rows := make([]apiAdminGroupRow, 0, len(gs))
	for i := range gs {
		g := &gs[i]
		if s.Groups == nil {
			s.writeJSONError(w, http.StatusInternalServerError, "server error")
			return
		}
		perms, err := s.Groups.GetGroupPermissions(ctx, g.ID)
		if err != nil {
			s.writeJSONError(w, http.StatusInternalServerError, "server error")
			return
		}
		ps := make([]string, 0, len(perms))
		for _, p := range perms {
			ps = append(ps, string(p))
		}
		rows = append(rows, apiAdminGroupRow{
			ID:          g.ID,
			Name:        g.Name,
			Permissions: ps,
			MemberCount: counts[g.ID],
			CreatedAt:   g.CreatedAt.UTC().Format(time.RFC3339Nano),
			UpdatedAt:   g.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"groups": rows})
}

type apiAdminGroupSaveReq struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func (s *Server) APIAdminGroupsCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiAdminGroupSaveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var perms []domain.Permission
	for _, permStr := range req.Permissions {
		p, ok := domain.ParsePermission(strings.TrimSpace(permStr))
		if !ok {
			s.writeJSONError(w, http.StatusBadRequest, "unknown permission: "+permStr)
			return
		}
		perms = append(perms, p)
	}
	g, err := s.Admin.CreatePermissionGroup(ctx, req.Name, perms)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"id": g.ID, "next": "/admin/groups"})
}

func (s *Server) APIAdminGroupGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	g, perms, err := s.Admin.GetPermissionGroup(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	ps := make([]string, 0, len(perms))
	for _, p := range perms {
		ps = append(ps, string(p))
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"group": map[string]any{
			"id": g.ID, "name": g.Name,
			"permissions": ps,
			"created_at":  g.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updated_at":  g.UpdatedAt.UTC().Format(time.RFC3339Nano),
		},
	})
}

func (s *Server) APIAdminGroupUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiAdminGroupSaveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var perms []domain.Permission
	for _, permStr := range req.Permissions {
		p, ok := domain.ParsePermission(strings.TrimSpace(permStr))
		if !ok {
			s.writeJSONError(w, http.StatusBadRequest, "unknown permission: "+permStr)
			return
		}
		perms = append(perms, p)
	}
	if err := s.Admin.UpdatePermissionGroup(ctx, id, req.Name, perms); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/groups/" + id})
}

func (s *Server) APIAdminGroupDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Admin.DeletePermissionGroup(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/groups"})
}

type apiAdminUserGroupsReq struct {
	GroupIDs []string `json:"group_ids"`
}

func (s *Server) APIAdminUserGroupsSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiAdminUserGroupsReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := s.Admin.SetUserPermissionGroups(ctx, id, req.GroupIDs); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, domain.ErrValidation) {
			s.writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/admin/users"})
}
