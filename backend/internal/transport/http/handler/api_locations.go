package handler

import (
	"net/http"
	"strings"

	"findus/backend/internal/domain"
)

// APILocationsList returns root locations.
func (s *Server) APILocationsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roots, err := s.Locs.ListChildren(ctx, nil)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"roots": roots})
}

// APILocationNew returns parent options for create form.
func (s *Server) APILocationNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	opts, err := s.parentLocationOptions(ctx, "")
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	pid := strings.TrimSpace(r.URL.Query().Get("parent_id"))
	s.writeJSON(w, http.StatusOK, map[string]any{"parent_options": opts, "selected_parent": pid})
}

type apiLocationCreateReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
}

func (s *Server) APILocationCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiLocationCreateReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	parent := strings.TrimSpace(req.ParentID)
	var p *string
	if parent != "" {
		p = &parent
	}
	loc, err := s.Inventory.CreateLocation(ctx, req.Name, req.Description, p)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"id": loc.ID, "next": "/locations/" + loc.ID})
}

// APILocationGet returns one location with children, items, breadcrumb.
func (s *Server) APILocationGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	loc, err := s.Locs.GetByID(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	path, err := s.Locs.PathFromRoot(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	var crumb []domain.LocationPathElement
	if len(path) > 0 {
		crumb = path[:len(path)-1]
	}
	backHref := "/locations"
	backLabel := "All locations"
	if len(crumb) > 0 {
		p := crumb[len(crumb)-1]
		backHref = "/locations/" + p.ID
		backLabel = "← " + p.Name
	}
	ch, err := s.Locs.ListChildren(ctx, &id)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	items, err := s.Items.ListByLocation(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"location": loc, "breadcrumb": crumb, "children": ch, "items": items,
		"back_href": backHref, "back_label": backLabel,
	})
}

// APILocationEdit returns location + parent options.
func (s *Server) APILocationEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	loc, err := s.Locs.GetByID(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	parent := ""
	if loc.ParentID != nil {
		parent = *loc.ParentID
	}
	opts, err := s.parentLocationOptions(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"location": loc, "parent_options": opts, "selected_parent": parent,
	})
}

func (s *Server) APILocationUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiLocationCreateReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	parent := strings.TrimSpace(req.ParentID)
	var p *string
	if parent != "" {
		p = &parent
	}
	if _, err := s.Inventory.UpdateLocation(ctx, id, req.Name, req.Description, p); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/locations/" + id})
}

func (s *Server) APILocationDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Inventory.DeleteLocation(ctx, id); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/locations"})
}
