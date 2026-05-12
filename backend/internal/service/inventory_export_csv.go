package service

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"findus/backend/internal/domain"
)

var importTemplateIDRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func encodeInventoryCSVZIP(bundle *InventoryExportBundle) ([]byte, error) {
	if bundle == nil {
		return nil, fmt.Errorf("%w: nil bundle", domain.ErrValidation)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mf := csvManifest{ExportVersion: bundle.ExportVersion, ExportedAt: bundle.ExportedAt}
	mb, err := json.Marshal(mf)
	if err != nil {
		_ = zw.Close()
		return nil, err
	}
	w, err := zw.Create(csvFileManifest)
	if err != nil {
		_ = zw.Close()
		return nil, err
	}
	if _, err := w.Write(mb); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeLocationsCSV(zw, bundle.Locations); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeLabelsCSV(zw, bundle.Labels); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeTemplatesCSV(zw, bundle.ItemTemplates); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeItemsCSV(zw, bundle.Items); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := writeItemLabelsCSV(zw, bundle.Items); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeLocationsCSV(zw *zip.Writer, locs []domain.Location) error {
	w, err := zw.Create(csvFileLocations)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "name", "parent_id", "description", "qr_token", "created_at", "updated_at"}); err != nil {
		return err
	}
	for _, l := range locs {
		pid := ""
		if l.ParentID != nil {
			pid = *l.ParentID
		}
		row := []string{
			l.ID, l.Name, pid, l.Description, l.QRToken,
			l.CreatedAt.UTC().Format(time.RFC3339Nano),
			l.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeLabelsCSV(zw *zip.Writer, labels []domain.Label) error {
	w, err := zw.Create(csvFileLabels)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "name", "color", "default_template_type", "created_at", "updated_at"}); err != nil {
		return err
	}
	for _, lb := range labels {
		dtp := ""
		if lb.DefaultTemplateType != nil {
			dtp = *lb.DefaultTemplateType
		}
		row := []string{
			lb.ID, lb.Name, lb.Color, dtp,
			lb.CreatedAt.UTC().Format(time.RFC3339Nano),
			lb.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeTemplatesCSV(zw *zip.Writer, tpls []InventoryExportTemplate) error {
	w, err := zw.Create(csvFileItemTemplates)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "display_name", "fields_json", "sort_order", "created_at", "updated_at"}); err != nil {
		return err
	}
	for _, t := range tpls {
		row := []string{
			t.ID, t.DisplayName, string(t.FieldsJSON),
			strconv.Itoa(t.SortOrder),
			t.CreatedAt.UTC().Format(time.RFC3339Nano),
			t.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeItemsCSV(zw *zip.Writer, items []InventoryExportItem) error {
	w, err := zw.Create(csvFileItems)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"id", "name", "description", "location_id", "template_type", "template_data", "additional_data", "photo_path", "qr_token", "created_at", "updated_at"}); err != nil {
		return err
	}
	for _, it := range items {
		pp := ""
		if it.PhotoPath != nil {
			pp = *it.PhotoPath
		}
		row := []string{
			it.ID, it.Name, it.Description, it.LocationID, string(it.TemplateType),
			string(it.TemplateData), string(it.AdditionalData), pp, it.QRToken,
			it.CreatedAt.UTC().Format(time.RFC3339Nano),
			it.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func writeItemLabelsCSV(zw *zip.Writer, items []InventoryExportItem) error {
	w, err := zw.Create(csvFileItemLabels)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"item_id", "label_id"}); err != nil {
		return err
	}
	for _, it := range items {
		for _, lid := range it.LabelIDs {
			lid = strings.TrimSpace(lid)
			if lid == "" {
				continue
			}
			if err := cw.Write([]string{it.ID, lid}); err != nil {
				return err
			}
		}
	}
	cw.Flush()
	return cw.Error()
}

func decodeInventoryCSVZIP(data []byte) (*InventoryExportBundle, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("zip: %w", err)
	}
	files := make(map[string][]byte)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		files[f.Name] = b
	}
	var man csvManifest
	if mb, ok := files[csvFileManifest]; ok {
		if err := json.Unmarshal(mb, &man); err != nil {
			return nil, fmt.Errorf("manifest: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing %s in zip", csvFileManifest)
	}
	if man.ExportVersion != inventoryExportVersion {
		return nil, fmt.Errorf("%w: export_version %d", domain.ErrValidation, man.ExportVersion)
	}
	if man.ExportedAt.IsZero() {
		man.ExportedAt = time.Now().UTC()
	}
	locs, err := readLocationsCSV(files[csvFileLocations])
	if err != nil {
		return nil, err
	}
	labels, err := readLabelsCSV(files[csvFileLabels])
	if err != nil {
		return nil, err
	}
	tpls, err := readTemplatesCSV(files[csvFileItemTemplates])
	if err != nil {
		return nil, err
	}
	items, err := readItemsCSV(files[csvFileItems])
	if err != nil {
		return nil, err
	}
	if err := mergeItemLabelsCSV(files[csvFileItemLabels], items); err != nil {
		return nil, err
	}
	return &InventoryExportBundle{
		ExportVersion: man.ExportVersion,
		ExportedAt:    man.ExportedAt.UTC(),
		Locations:     locs,
		Labels:        labels,
		ItemTemplates: tpls,
		Items:         items,
	}, nil
}

func readLocationsCSV(raw []byte) ([]domain.Location, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("missing %s", csvFileLocations)
	}
	r := csv.NewReader(bytes.NewReader(raw))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, fmt.Errorf("empty locations csv")
	}
	h := rows[0]
	idx, err := csvHeaderIndex(h, "id", "name", "parent_id", "description", "qr_token", "created_at", "updated_at")
	if err != nil {
		return nil, err
	}
	var out []domain.Location
	for _, row := range rows[1:] {
		if err := safeCSVRowLength(row, 512000); err != nil {
			return nil, err
		}
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		var pid *string
		if p := strings.TrimSpace(getCol(row, idx, "parent_id")); p != "" {
			pid = &p
		}
		ct, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "created_at"))
		if err != nil {
			return nil, fmt.Errorf("location created_at: %w", err)
		}
		ut, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "updated_at"))
		if err != nil {
			return nil, fmt.Errorf("location updated_at: %w", err)
		}
		out = append(out, domain.Location{
			ID:          strings.TrimSpace(getCol(row, idx, "id")),
			Name:        getCol(row, idx, "name"),
			ParentID:    pid,
			Description: getCol(row, idx, "description"),
			QRToken:     strings.TrimSpace(getCol(row, idx, "qr_token")),
			CreatedAt:   ct.UTC(),
			UpdatedAt:   ut.UTC(),
		})
	}
	return out, nil
}

func readLabelsCSV(raw []byte) ([]domain.Label, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("missing %s", csvFileLabels)
	}
	r := csv.NewReader(bytes.NewReader(raw))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, fmt.Errorf("empty labels csv")
	}
	h := rows[0]
	idx, err := csvHeaderIndex(h, "id", "name", "color", "default_template_type", "created_at", "updated_at")
	if err != nil {
		return nil, err
	}
	var out []domain.Label
	for _, row := range rows[1:] {
		if err := safeCSVRowLength(row, 512000); err != nil {
			return nil, err
		}
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		dtp := strings.TrimSpace(getCol(row, idx, "default_template_type"))
		var dtpPtr *string
		if dtp != "" {
			dtpPtr = &dtp
		}
		ct, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "created_at"))
		if err != nil {
			return nil, fmt.Errorf("label created_at: %w", err)
		}
		ut, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "updated_at"))
		if err != nil {
			return nil, fmt.Errorf("label updated_at: %w", err)
		}
		out = append(out, domain.Label{
			ID:                  strings.TrimSpace(getCol(row, idx, "id")),
			Name:                getCol(row, idx, "name"),
			Color:               getCol(row, idx, "color"),
			DefaultTemplateType: dtpPtr,
			CreatedAt:           ct.UTC(),
			UpdatedAt:           ut.UTC(),
		})
	}
	return out, nil
}

func readTemplatesCSV(raw []byte) ([]InventoryExportTemplate, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("missing %s", csvFileItemTemplates)
	}
	r := csv.NewReader(bytes.NewReader(raw))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, fmt.Errorf("empty templates csv")
	}
	h := rows[0]
	idx, err := csvHeaderIndex(h, "id", "display_name", "fields_json", "sort_order", "created_at", "updated_at")
	if err != nil {
		return nil, err
	}
	var out []InventoryExportTemplate
	for _, row := range rows[1:] {
		if err := safeCSVRowLength(row, 512000); err != nil {
			return nil, err
		}
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		so, err := strconv.Atoi(getCol(row, idx, "sort_order"))
		if err != nil {
			return nil, fmt.Errorf("template sort_order: %w", err)
		}
		ct, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "created_at"))
		if err != nil {
			return nil, fmt.Errorf("template created_at: %w", err)
		}
		ut, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "updated_at"))
		if err != nil {
			return nil, fmt.Errorf("template updated_at: %w", err)
		}
		fj := json.RawMessage(strings.TrimSpace(getCol(row, idx, "fields_json")))
		if len(fj) == 0 {
			fj = json.RawMessage(`[]`)
		}
		out = append(out, InventoryExportTemplate{
			ID:          strings.TrimSpace(getCol(row, idx, "id")),
			DisplayName: getCol(row, idx, "display_name"),
			FieldsJSON:  fj,
			SortOrder:   so,
			CreatedAt:   ct.UTC(),
			UpdatedAt:   ut.UTC(),
		})
	}
	return out, nil
}

func readItemsCSV(raw []byte) ([]InventoryExportItem, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("missing %s", csvFileItems)
	}
	r := csv.NewReader(bytes.NewReader(raw))
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, fmt.Errorf("empty items csv")
	}
	h := rows[0]
	idx, err := csvHeaderIndex(h, "id", "name", "description", "location_id", "template_type", "template_data", "additional_data", "photo_path", "qr_token", "created_at", "updated_at")
	if err != nil {
		return nil, err
	}
	var out []InventoryExportItem
	for _, row := range rows[1:] {
		if err := safeCSVRowLength(row, 512000); err != nil {
			return nil, err
		}
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		pp := strings.TrimSpace(getCol(row, idx, "photo_path"))
		var ppPtr *string
		if pp != "" {
			ppPtr = &pp
		}
		ct, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "created_at"))
		if err != nil {
			return nil, fmt.Errorf("item created_at: %w", err)
		}
		ut, err := time.Parse(time.RFC3339Nano, getCol(row, idx, "updated_at"))
		if err != nil {
			return nil, fmt.Errorf("item updated_at: %w", err)
		}
		td := json.RawMessage(strings.TrimSpace(getCol(row, idx, "template_data")))
		if len(td) == 0 {
			td = json.RawMessage(`{}`)
		}
		ad := json.RawMessage(strings.TrimSpace(getCol(row, idx, "additional_data")))
		if len(ad) == 0 {
			ad = json.RawMessage(`{}`)
		}
		out = append(out, InventoryExportItem{
			Item: domain.Item{
				ID:             strings.TrimSpace(getCol(row, idx, "id")),
				Name:           getCol(row, idx, "name"),
				Description:    getCol(row, idx, "description"),
				LocationID:     strings.TrimSpace(getCol(row, idx, "location_id")),
				TemplateType:   domain.TemplateType(strings.TrimSpace(getCol(row, idx, "template_type"))),
				TemplateData:   td,
				AdditionalData: ad,
				PhotoPath:      ppPtr,
				QRToken:        strings.TrimSpace(getCol(row, idx, "qr_token")),
				CreatedAt:      ct.UTC(),
				UpdatedAt:      ut.UTC(),
			},
			LabelIDs: nil,
		})
	}
	return out, nil
}

func mergeItemLabelsCSV(raw []byte, items []InventoryExportItem) error {
	if len(raw) == 0 {
		return nil
	}
	r := csv.NewReader(bytes.NewReader(raw))
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return nil
	}
	h := rows[0]
	idx, err := csvHeaderIndex(h, "item_id", "label_id")
	if err != nil {
		return err
	}
	byItem := make(map[string][]string)
	for _, row := range rows[1:] {
		if err := safeCSVRowLength(row, 65536); err != nil {
			return err
		}
		if len(row) < 2 {
			continue
		}
		iid := strings.TrimSpace(getCol(row, idx, "item_id"))
		lid := strings.TrimSpace(getCol(row, idx, "label_id"))
		if iid == "" || lid == "" {
			continue
		}
		byItem[iid] = append(byItem[iid], lid)
	}
	for i := range items {
		if extra, ok := byItem[items[i].ID]; ok {
			items[i].LabelIDs = dedupeStrings(append(items[i].LabelIDs, extra...))
		}
	}
	return nil
}

func getCol(row []string, idx map[string]int, col string) string {
	i := idx[col]
	if i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}

func csvHeaderIndex(header []string, names ...string) (map[string]int, error) {
	idx := make(map[string]int)
	for _, n := range names {
		idx[n] = -1
	}
	for i, c := range header {
		c = strings.TrimSpace(strings.ToLower(c))
		for _, n := range names {
			if c == strings.ToLower(n) {
				idx[n] = i
				break
			}
		}
	}
	for _, n := range names {
		if idx[n] < 0 {
			return nil, fmt.Errorf("csv: missing column %q", n)
		}
	}
	return idx, nil
}

func safeCSVRowLength(row []string, max int) error {
	n := 0
	for _, c := range row {
		n += utf8.RuneCountInString(c)
		if n > max {
			return errors.New("csv row too large")
		}
	}
	return nil
}
