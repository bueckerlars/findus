package handler

import (
	"encoding/json"
	"errors"
	htempl "html/template"
	"net/http"
	"strconv"
	"strings"

	"findus/internal/domain"
	"findus/internal/service"
)

type adminTemplateRow struct {
	T               domain.ItemTemplate
	Count           int64
	HasAlternate    bool
	FallbackDisplay string
}

func fieldsInitFromForm(raw string) htempl.JS {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return htempl.JS("[]")
	}
	var a any
	if json.Unmarshal([]byte(raw), &a) != nil {
		return htempl.JS("[]")
	}
	return htempl.JS(raw)
}

func (s *Server) AdminTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.Inventory.ListItemTemplates(ctx)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	rows := make([]adminTemplateRow, 0, len(list))
	for _, t := range list {
		n, _ := s.Items.CountByTemplateType(ctx, t.ID)
		_, fdn, ok := service.ReassignTargetForDelete(list, t.ID)
		rows = append(rows, adminTemplateRow{T: t, Count: n, HasAlternate: ok, FallbackDisplay: fdn})
	}
	sh := s.shell(r, "Item templates")
	s.render(w, http.StatusOK, "page_admin_templates", struct {
		shell
		Rows []adminTemplateRow
	}{shell: sh, Rows: rows})
}

type adminTemplateNewPage struct {
	shell
	Error      string
	FormID     string
	FormName   string
	FormSort   string
	FieldsInit htempl.JS
}

func (s *Server) AdminTemplateNewGet(w http.ResponseWriter, r *http.Request) {
	sh := s.shell(r, "New item template")
	s.render(w, http.StatusOK, "page_admin_template_new", adminTemplateNewPage{
		shell: sh, FormSort: "10", FieldsInit: htempl.JS("[]"),
	})
}

func (s *Server) AdminTemplateNewPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	sortOrder, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("sort_order")))
	if err := s.Inventory.CreateItemTemplate(ctx, r.FormValue("id"), r.FormValue("display_name"), []byte(r.FormValue("fields_json")), sortOrder); err != nil {
		sh := s.shell(r, "New item template")
		s.render(w, http.StatusOK, "page_admin_template_new", adminTemplateNewPage{
			shell: sh, Error: err.Error(),
			FormID: r.FormValue("id"), FormName: r.FormValue("display_name"),
			FormSort:   strings.TrimSpace(r.FormValue("sort_order")),
			FieldsInit: fieldsInitFromForm(r.FormValue("fields_json")),
		})
		return
	}
	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

type adminTemplateEditPage struct {
	shell
	Template   *domain.ItemTemplate
	FieldsInit htempl.JS
	Error      string
}

func (s *Server) AdminTemplateEditGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	t, err := s.Inventory.GetItemTemplate(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	sh := s.shell(r, "Edit item template")
	init := string(t.FieldsJSON)
	if strings.TrimSpace(init) == "" {
		init = "[]"
	}
	s.render(w, http.StatusOK, "page_admin_template_edit", adminTemplateEditPage{
		shell: sh, Template: t, FieldsInit: htempl.JS(init),
	})
}

func (s *Server) AdminTemplateEditPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	sortOrder, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("sort_order")))
	if err := s.Inventory.UpdateItemTemplate(ctx, id, r.FormValue("display_name"), []byte(r.FormValue("fields_json")), sortOrder); err != nil {
		t, gerr := s.Inventory.GetItemTemplate(ctx, id)
		if gerr != nil {
			http.NotFound(w, r)
			return
		}
		sh := s.shell(r, "Edit item template")
		s.render(w, http.StatusOK, "page_admin_template_edit", adminTemplateEditPage{
			shell: sh, Template: t, FieldsInit: fieldsInitFromForm(r.FormValue("fields_json")),
			Error: err.Error(),
		})
		return
	}
	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}

func (s *Server) AdminTemplateDeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Inventory.DeleteItemTemplate(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin/templates", http.StatusSeeOther)
}
