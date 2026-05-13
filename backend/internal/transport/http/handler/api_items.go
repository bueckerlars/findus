package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"findus/backend/internal/domain"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/middleware"
)

type itemListRow struct {
	domain.Item
	LocationName string `json:"location_name"`
}

func (s *Server) itemsAsListRows(ctx context.Context, items []domain.Item) []itemListRow {
	if len(items) == 0 {
		return []itemListRow{}
	}
	ids := make([]string, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].LocationID)
	}
	names := s.locationDisplayNamesForIDs(ctx, ids)
	out := make([]itemListRow, len(items))
	for i := range items {
		out[i] = itemListRow{Item: items[i], LocationName: names[items[i].LocationID]}
	}
	return out
}

func (s *Server) APIItemsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, err := s.Items.ListAll(ctx, 500)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"items": s.itemsAsListRows(ctx, items)})
}

func (s *Server) APIItemNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	opts, err := s.flatLocations(ctx)
	if err != nil || len(opts) == 0 {
		s.writeJSON(w, http.StatusOK, map[string]any{"no_locations": true})
		return
	}
	sel := strings.TrimSpace(r.URL.Query().Get("location_id"))
	if sel == "" {
		sel = opts[0].ID
	}
	labels, _ := s.Inventory.ListLabels(ctx)
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	defTpl := ""
	if len(templates) > 0 {
		defTpl = templates[0].ID
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"locations": opts, "selected_location": sel, "all_labels": labels,
		"templates": templates, "default_template_id": defTpl,
	})
}

func (s *Server) APIItemNewFields(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tpl := strings.TrimSpace(r.URL.Query().Get("template_type"))
	if tpl == "" {
		list, _ := s.Inventory.ListItemTemplates(ctx)
		if len(list) > 0 {
			tpl = list[0].ID
		}
	}
	t, err := s.Inventory.GetItemTemplate(ctx, tpl)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Unknown template.")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"fields": t.Fields, "template_id": t.ID})
}

func (s *Server) APIItemCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "bad form")
		return
	}
	tt, err := s.Inventory.GetItemTemplate(ctx, r.FormValue("template_type"))
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid template type.")
		return
	}
	add, err := service.ParseAdditionalFromForm(r)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	form := make(map[string]string)
	for _, f := range tt.Fields {
		form[f.Key] = r.FormValue(f.Key)
	}
	td, err := domain.NormalizeTemplateDataFromFields(tt.Fields, form)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	merged, err := service.MergeTemplateIntoAdditional(add, td)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	labelIDs := parseLabelIDs(r)
	emptyTD := json.RawMessage(`{}`)
	it, err := s.Inventory.CreateItem(ctx, r.FormValue("name"), r.FormValue("description"), r.FormValue("location_id"), domain.TemplateType(tt.ID), emptyTD, merged, labelIDs)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if f, hdr, err := r.FormFile("photo"); err == nil {
		defer f.Close()
		_ = hdr
		if err := s.saveItemPhoto(ctx, it.ID, f); err != nil {
			s.Log.Error("photo", "err", err)
		} else {
			rel := filepath.ToSlash(filepath.Join("images", it.ID+".webp"))
			it.PhotoPath = &rel
			it.UpdatedAt = time.Now().UTC()
			_ = s.Items.Update(ctx, it)
		}
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"id": it.ID, "next": "/items/" + it.ID})
}

func (s *Server) APIItemGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	canEdit := middleware.Can(ctx, domain.PermItemsWrite)
	editMode := r.URL.Query().Get("edit") == "1"
	if editMode && !canEdit {
		editMode = false
	}
	out := s.itemDetailJSON(ctx, it, editMode)
	s.writeJSON(w, http.StatusOK, out)
}

func (s *Server) itemDetailJSON(ctx context.Context, it *domain.Item, editMode bool) map[string]any {
	labels, _ := s.Items.ListLabelsForItem(ctx, it.ID)
	labelIDs := make([]string, len(labels))
	for i := range labels {
		labelIDs[i] = labels[i].ID
	}
	selectedLabels := make(map[string]bool, len(labelIDs))
	for _, id := range labelIDs {
		selectedLabels[id] = true
	}
	opts, _ := s.flatLocations(ctx)
	allLabels, _ := s.Inventory.ListLabels(ctx)
	tpl, tplErr := s.Inventory.GetItemTemplate(ctx, string(it.TemplateType))
	if tplErr != nil {
		tpl = nil
	}
	merged, err := service.MergeTemplateIntoAdditional(it.AdditionalData, it.TemplateData)
	if err != nil {
		s.Log.Error("item merge", "err", err)
		merged = it.AdditionalData
	}
	attrRows := service.ItemPropertyDetailRows(tpl, merged)
	tplCaption := strings.TrimSpace(string(it.TemplateType))
	if tpl != nil && strings.TrimSpace(tpl.DisplayName) != "" {
		tplCaption = tpl.DisplayName
	}
	const tsLayout = "2 Jan 2006, 15:04 MST"
	systemRows := []map[string]string{
		{"label": "Item ID", "value": it.ID},
		{"label": "Created", "value": it.CreatedAt.UTC().Format(tsLayout)},
		{"label": "Last updated", "value": it.UpdatedAt.UTC().Format(tsLayout)},
		{"label": "Template", "value": tplCaption},
	}
	locPath, _ := s.Locs.PathFromRoot(ctx, it.LocationID)
	if labels == nil {
		labels = []domain.Label{}
	}
	if locPath == nil {
		locPath = []domain.LocationPathElement{}
	}
	if attrRows == nil {
		attrRows = []service.ItemPropertyDetailRow{}
	}
	if opts == nil {
		opts = []locOpt{}
	}
	if allLabels == nil {
		allLabels = []domain.Label{}
	}
	photo := ""
	if it.PhotoPath != nil && *it.PhotoPath != "" {
		photo = "/items/" + it.ID + "/photo"
	}
	atts, err := s.itemAttachmentsPayload(ctx, it.ID)
	if err != nil {
		s.Log.Error("attachments payload", "err", err)
		atts = []map[string]any{}
	}
	return map[string]any{
		"item": it, "labels": labels, "photo_url": photo, "location_path": locPath,
		"edit_mode": editMode, "attr_rows": attrRows, "system_rows": systemRows,
		"locations": opts, "all_labels": allLabels, "selected_labels": selectedLabels,
		"attachments": atts,
	}
}

func (s *Server) APIItemEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	opts, err := s.flatLocations(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	itemLabels, err := s.Items.ListLabelsForItem(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	selected := make(map[string]bool)
	for _, lb := range itemLabels {
		selected[lb.ID] = true
	}
	allLabels, _ := s.Inventory.ListLabels(ctx)
	if opts == nil {
		opts = []locOpt{}
	}
	if allLabels == nil {
		allLabels = []domain.Label{}
	}
	s.writeJSON(w, http.StatusOK, map[string]any{
		"item": it, "locations": opts, "all_labels": allLabels, "selected_labels": selected,
		"additional_rows": additionalFieldPairs(mergedAdditionalForEdit(it)),
	})
}

func (s *Server) APIItemUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "bad form")
		return
	}
	tplID := strings.TrimSpace(r.FormValue("template_type"))
	if _, err := s.Inventory.GetItemTemplate(ctx, tplID); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Invalid template type.")
		return
	}
	tt := domain.TemplateType(tplID)
	if tt != it.TemplateType {
		s.writeJSONError(w, http.StatusBadRequest, "Template type cannot be changed here.")
		return
	}
	add, err := service.ParseAdditionalFromForm(r)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	labelIDs := parseLabelIDs(r)
	emptyTD := json.RawMessage(`{}`)
	it2, err := s.Inventory.UpdateItem(ctx, id, r.FormValue("name"), r.FormValue("description"), r.FormValue("location_id"), tt, emptyTD, add, labelIDs)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if f, hdr, err := r.FormFile("photo"); err == nil {
		defer f.Close()
		_ = hdr
		if err := s.saveItemPhoto(ctx, id, f); err != nil {
			s.Log.Error("photo", "err", err)
		} else {
			rel := filepath.ToSlash(filepath.Join("images", id+".webp"))
			it2.PhotoPath = &rel
			it2.UpdatedAt = time.Now().UTC()
			_ = s.Items.Update(ctx, it2)
		}
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/items/" + id})
}

func (s *Server) APIItemDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	it, err := s.Items.GetByID(ctx, id)
	if err == nil && it.PhotoPath != nil {
		_ = os.Remove(filepath.Join(s.Config.DataDir, filepath.FromSlash(*it.PhotoPath)))
	}
	if err := s.Inventory.DeleteItem(ctx, id); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = service.RemoveItemAttachmentDir(s.Config.DataDir, id)
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/items"})
}

func (s *Server) APISearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	res, err := s.Items.Search(ctx, q, 50)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	if res == nil {
		res = []domain.Item{}
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"query": q, "results": s.itemsAsListRows(ctx, res)})
}

type apiLabelRow struct {
	Label                domain.Label `json:"label"`
	ChipHref             string       `json:"chip_href"`
	DefaultTemplateTitle string       `json:"default_template_title,omitempty"`
}

func (s *Server) APILabelsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	canEdit := middleware.Can(ctx, domain.PermLabelsWrite)
	list, err := s.Inventory.ListLabels(ctx)
	if err != nil {
		s.writeJSONError(w, http.StatusInternalServerError, "server error")
		return
	}
	tpls, _ := s.Inventory.ListItemTemplates(ctx)
	titles := make(map[string]string, len(tpls))
	for _, t := range tpls {
		titles[t.ID] = t.DisplayName
	}
	rows := s.labelRows(canEdit, list, titles)
	out := make([]apiLabelRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, apiLabelRow{
			Label: row.LB, ChipHref: row.ChipHref,
			DefaultTemplateTitle: row.DefaultTemplateTitle,
		})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"label_rows": out})
}

func (s *Server) APILabelNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	s.writeJSON(w, http.StatusOK, map[string]any{"templates": templates})
}

type apiLabelSaveReq struct {
	Name                string `json:"name"`
	Color               string `json:"color"`
	DefaultTemplateType string `json:"default_template_type"`
}

func (s *Server) APILabelCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req apiLabelSaveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	rawTpl := strings.TrimSpace(req.DefaultTemplateType)
	var def *string
	if rawTpl != "" {
		def = &rawTpl
	}
	if _, err := s.Inventory.CreateLabel(ctx, req.Name, req.Color, def); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/labels"})
}

func (s *Server) APILabelGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	lb, err := s.Labels.GetByID(ctx, id)
	if err != nil {
		s.writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	sel := ""
	if lb.DefaultTemplateType != nil {
		sel = *lb.DefaultTemplateType
	}
	templates, _ := s.Inventory.ListItemTemplates(ctx)
	s.writeJSON(w, http.StatusOK, map[string]any{"label": lb, "selected_template": sel, "templates": templates})
}

func (s *Server) APILabelUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	var req apiLabelSaveReq
	if err := readJSON(r, &req); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	rawTpl := strings.TrimSpace(req.DefaultTemplateType)
	var def *string
	if rawTpl != "" {
		def = &rawTpl
	}
	if _, err := s.Inventory.UpdateLabel(ctx, id, req.Name, req.Color, def); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/labels"})
}

func (s *Server) APILabelDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if err := s.Inventory.DeleteLabel(ctx, id); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"next": "/labels"})
}
