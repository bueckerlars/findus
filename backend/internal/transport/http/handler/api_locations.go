package handler

import (
	"net/http"
	"sort"
	"strings"

	"findus/backend/internal/domain"
)

// apiLocationTreeNode is a minimal location row for the list UI (no duplicate long fields).
type apiLocationTreeNode struct {
	ID       string                `json:"ID"`
	Name     string                `json:"Name"`
	Children []apiLocationTreeNode `json:"children"`
}

func buildLocationTreeNodes(byParent map[string][]domain.Location, id string) []apiLocationTreeNode {
	kids := byParent[id]
	out := make([]apiLocationTreeNode, 0, len(kids))
	for _, l := range kids {
		out = append(out, apiLocationTreeNode{
			ID:       l.ID,
			Name:     l.Name,
			Children: buildLocationTreeNodes(byParent, l.ID),
		})
	}
	return out
}

// APILocationsList returns the full location hierarchy for the list page.
func (s *Server) APILocationsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	all, err := s.Locs.ListAll(ctx, 500)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	byParent := make(map[string][]domain.Location)
	for _, l := range all {
		key := ""
		if l.ParentID != nil {
			key = *l.ParentID
		}
		byParent[key] = append(byParent[key], l)
	}
	for k, sl := range byParent {
		sort.Slice(sl, func(i, j int) bool {
			return strings.ToLower(sl[i].Name) < strings.ToLower(sl[j].Name)
		})
		byParent[k] = sl
	}
	tree := buildLocationTreeNodes(byParent, "")
	s.writeJSON(w, http.StatusOK, map[string]any{"tree": tree})
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
	// Nil slices encode as JSON null; clients expect arrays for list fields.
	if crumb == nil {
		crumb = []domain.LocationPathElement{}
	}
	if ch == nil {
		ch = []domain.Location{}
	}
	if items == nil {
		items = []domain.Item{}
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
