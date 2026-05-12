package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"findus/backend/internal/domain"
)

// BuildInventoryExportBundle loads all inventory rows for JSON/ZIP export.
func (s *Inventory) BuildInventoryExportBundle(ctx context.Context) (*InventoryExportBundle, error) {
	if s.Locations == nil || s.Items == nil || s.Templates == nil || s.Labels == nil {
		return nil, fmt.Errorf("%w: inventory not configured", domain.ErrValidation)
	}
	locs, err := s.Locations.ListAllExport(ctx)
	if err != nil {
		return nil, err
	}
	labels, err := s.Labels.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	tpls, err := s.Templates.List(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.Items.ListAllExport(ctx)
	if err != nil {
		return nil, err
	}
	pairs, err := s.Items.ListItemLabelPairs(ctx)
	if err != nil {
		return nil, err
	}
	labelsByItem := make(map[string][]string)
	for _, p := range pairs {
		labelsByItem[p.ItemID] = append(labelsByItem[p.ItemID], p.LabelID)
	}
	outItems := make([]InventoryExportItem, 0, len(items))
	for i := range items {
		it := items[i]
		outItems = append(outItems, InventoryExportItem{
			Item:     it,
			LabelIDs: dedupeStrings(labelsByItem[it.ID]),
		})
	}
	et := make([]InventoryExportTemplate, 0, len(tpls))
	for i := range tpls {
		et = append(et, exportTemplateFromDomain(tpls[i]))
	}
	return &InventoryExportBundle{
		ExportVersion: inventoryExportVersion,
		ExportedAt:    time.Now().UTC(),
		Locations:     locs,
		Labels:        labels,
		ItemTemplates: et,
		Items:         outItems,
	}, nil
}

// ImportInventoryBundle applies a merge import (upsert by id) using the given inventory repos (typically tx-backed).
func (s *Inventory) ImportInventoryBundle(ctx context.Context, bundle *InventoryExportBundle) (*InventoryImportResult, error) {
	if err := ValidateInventoryImportBundle(bundle); err != nil {
		return nil, err
	}
	if s.Locations == nil || s.Items == nil || s.Templates == nil || s.Labels == nil || s.Settings == nil {
		return nil, fmt.Errorf("%w: inventory not configured", domain.ErrValidation)
	}
	res := &InventoryImportResult{}
	locs, err := sortLocationsTopo(bundle.Locations)
	if err != nil {
		return nil, err
	}

	for _, t := range bundle.ItemTemplates {
		if err := importOneTemplate(ctx, s, t, res); err != nil {
			return nil, err
		}
	}
	for _, lb := range bundle.Labels {
		if err := importOneLabel(ctx, s, lb, res); err != nil {
			return nil, err
		}
	}
	for _, loc := range locs {
		if err := importOneLocation(ctx, s, loc, res); err != nil {
			return nil, err
		}
	}
	for _, row := range bundle.Items {
		if err := importOneItem(ctx, s, row, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func importOneTemplate(ctx context.Context, s *Inventory, t InventoryExportTemplate, res *InventoryImportResult) error {
	id := strings.TrimSpace(t.ID)
	if id == "" {
		return fmt.Errorf("%w: template id", domain.ErrValidation)
	}
	if !importTemplateIDRE.MatchString(id) {
		return fmt.Errorf("%w: invalid template id %q", domain.ErrValidation, id)
	}
	fields, err := domain.ParseTemplateFieldsJSON(t.FieldsJSON)
	if err != nil {
		return fmt.Errorf("%w: template %q fields: %w", domain.ErrValidation, id, err)
	}
	canon, err := domain.MarshalTemplateFields(fields)
	if err != nil {
		return err
	}
	_, err = s.Templates.GetByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		now := time.Now().UTC()
		dt := domain.ItemTemplate{
			ID:          id,
			DisplayName: strings.TrimSpace(t.DisplayName),
			FieldsJSON:  canon,
			Fields:      fields,
			SortOrder:   t.SortOrder,
			CreatedAt:   pickTime(t.CreatedAt, now),
			UpdatedAt:   pickTime(t.UpdatedAt, now),
		}
		if err := s.Templates.Create(ctx, &dt); err != nil {
			return err
		}
		res.TemplatesCreated++
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.UpdateItemTemplate(ctx, id, t.DisplayName, canon, t.SortOrder); err != nil {
		return err
	}
	res.TemplatesUpdated++
	return nil
}

func importOneLabel(ctx context.Context, s *Inventory, lb domain.Label, res *InventoryImportResult) error {
	id := strings.TrimSpace(lb.ID)
	if id == "" {
		return fmt.Errorf("%w: label id", domain.ErrValidation)
	}
	existing, err := s.Labels.GetByID(ctx, id)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if errors.Is(err, domain.ErrNotFound) {
		if err := labelNameFree(ctx, s, lb.Name, id); err != nil {
			return err
		}
		now := time.Now().UTC()
		l := domain.Label{
			ID:                  id,
			Name:                strings.TrimSpace(lb.Name),
			Color:               strings.TrimSpace(lb.Color),
			DefaultTemplateType: lb.DefaultTemplateType,
			CreatedAt:           pickTime(lb.CreatedAt, now),
			UpdatedAt:           pickTime(lb.UpdatedAt, now),
		}
		if err := s.Labels.Create(ctx, &l); err != nil {
			if isSQLiteUniqueViolation(err) {
				return fmt.Errorf("%w: label name %q", ErrLabelNameConflict, lb.Name)
			}
			return err
		}
		res.LabelsCreated++
		return nil
	}
	if strings.TrimSpace(existing.Name) != strings.TrimSpace(lb.Name) {
		if err := labelNameFree(ctx, s, lb.Name, id); err != nil {
			return err
		}
	}
	_, err = s.UpdateLabel(ctx, id, lb.Name, lb.Color, lb.DefaultTemplateType)
	if err != nil {
		return err
	}
	res.LabelsUpdated++
	return nil
}

func labelNameFree(ctx context.Context, s *Inventory, name, excludeID string) error {
	list, err := s.Labels.ListAll(ctx)
	if err != nil {
		return err
	}
	n := strings.TrimSpace(name)
	for _, o := range list {
		if o.ID == excludeID {
			continue
		}
		if strings.TrimSpace(o.Name) == n {
			return fmt.Errorf("%w: label name %q already used by %q", ErrLabelNameConflict, n, o.ID)
		}
	}
	return nil
}

func importOneLocation(ctx context.Context, s *Inventory, l domain.Location, res *InventoryImportResult) error {
	id := strings.TrimSpace(l.ID)
	if id == "" {
		return fmt.Errorf("%w: location id", domain.ErrValidation)
	}
	pid := l.ParentID
	if pid != nil && strings.TrimSpace(*pid) == "" {
		pid = nil
	}
	cur, err := s.Locations.GetByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		if pid != nil && *pid != "" {
			if err := s.assertNoCycle(ctx, id, *pid); err != nil {
				return err
			}
		}
		now := time.Now().UTC()
		for attempt := 0; attempt < 24; attempt++ {
			cand := l
			cand.ID = id
			cand.ParentID = pid
			cand.CreatedAt = pickTime(l.CreatedAt, now)
			cand.UpdatedAt = pickTime(l.UpdatedAt, now)
			if attempt > 0 {
				cand.QRToken = newID()
			}
			if err := s.Locations.Create(ctx, &cand); err != nil {
				if isSQLiteUniqueViolation(err) {
					continue
				}
				return err
			}
			res.LocationsCreated++
			return nil
		}
		return fmt.Errorf("location %q: qr_token collision", id)
	}
	if err != nil {
		return err
	}
	if pid != nil && *pid != "" {
		if err := s.assertNoCycle(ctx, id, *pid); err != nil {
			return err
		}
	}
	cur.Name = strings.TrimSpace(l.Name)
	cur.Description = l.Description
	cur.ParentID = pid
	cur.UpdatedAt = time.Now().UTC()
	if err := s.Locations.Update(ctx, cur); err != nil {
		return err
	}
	res.LocationsUpdated++
	return nil
}

func importOneItem(ctx context.Context, s *Inventory, row InventoryExportItem, res *InventoryImportResult) error {
	id := strings.TrimSpace(row.ID)
	if id == "" {
		return fmt.Errorf("%w: item id", domain.ErrValidation)
	}
	if err := s.ensureTemplateType(ctx, row.TemplateType); err != nil {
		return err
	}
	td := row.TemplateData
	if len(td) == 0 {
		td = json.RawMessage(`{}`)
	}
	ad := row.AdditionalData
	if len(ad) == 0 {
		ad = json.RawMessage(`{}`)
	}
	if err := ValidateAdditionalJSON(ad); err != nil {
		return err
	}
	if _, err := s.Locations.GetByID(ctx, row.LocationID); err != nil {
		return fmt.Errorf("item %q: location %q: %w", id, row.LocationID, err)
	}
	ids := dedupeStrings(row.LabelIDs)
	if err := s.validateLabelRefs(ctx, ids); err != nil {
		return fmt.Errorf("item %q: %w", id, err)
	}

	cur, err := s.Items.GetByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		now := time.Now().UTC()
		for attempt := 0; attempt < 24; attempt++ {
			cand := domain.Item{
				ID:             id,
				Name:           trim(row.Name, 1, 200),
				Description:    trim(row.Description, 0, 5000),
				LocationID:     row.LocationID,
				TemplateType:   row.TemplateType,
				TemplateData:   td,
				AdditionalData: ad,
				PhotoPath:      row.PhotoPath,
				QRToken:        row.QRToken,
				CreatedAt:      pickTime(row.CreatedAt, now),
				UpdatedAt:      pickTime(row.UpdatedAt, now),
			}
			if cand.Name == "" {
				return fmt.Errorf("%w: item name", domain.ErrValidation)
			}
			if attempt > 0 {
				cand.QRToken = newID()
			}
			if err := s.Items.Create(ctx, &cand); err != nil {
				if isSQLiteUniqueViolation(err) {
					continue
				}
				return err
			}
			if err := s.Items.ReplaceItemLabels(ctx, id, ids); err != nil {
				return err
			}
			res.ItemsCreated++
			res.ItemLabelsReplaced++
			return nil
		}
		return fmt.Errorf("item %q: qr_token collision", id)
	}
	if err != nil {
		return err
	}
	cur.Name = trim(row.Name, 1, 200)
	cur.Description = trim(row.Description, 0, 5000)
	cur.LocationID = row.LocationID
	cur.TemplateType = row.TemplateType
	cur.TemplateData = td
	cur.AdditionalData = ad
	cur.PhotoPath = row.PhotoPath
	cur.UpdatedAt = time.Now().UTC()
	if cur.Name == "" {
		return fmt.Errorf("%w: item name", domain.ErrValidation)
	}
	if err := s.Items.Update(ctx, cur); err != nil {
		return err
	}
	if err := s.Items.ReplaceItemLabels(ctx, id, ids); err != nil {
		return err
	}
	res.ItemsUpdated++
	res.ItemLabelsReplaced++
	return nil
}

func pickTime(t time.Time, fallback time.Time) time.Time {
	if t.IsZero() {
		return fallback
	}
	return t.UTC()
}

func sortLocationsTopo(all []domain.Location) ([]domain.Location, error) {
	idSet := make(map[string]struct{}, len(all))
	for _, l := range all {
		id := strings.TrimSpace(l.ID)
		if id == "" {
			return nil, fmt.Errorf("%w: empty location id", domain.ErrValidation)
		}
		idSet[id] = struct{}{}
	}
	outSet := make(map[string]struct{})
	out := make([]domain.Location, 0, len(all))
	for len(out) < len(all) {
		progressed := false
		for _, l := range all {
			id := strings.TrimSpace(l.ID)
			if _, done := outSet[id]; done {
				continue
			}
			var p string
			if l.ParentID != nil {
				p = strings.TrimSpace(*l.ParentID)
			}
			if p != "" {
				if _, ok := idSet[p]; !ok {
					return nil, fmt.Errorf("%w: location %q references unknown parent %q", domain.ErrValidation, id, p)
				}
				if _, ok := outSet[p]; !ok {
					continue
				}
			}
			out = append(out, l)
			outSet[id] = struct{}{}
			progressed = true
		}
		if !progressed {
			return nil, fmt.Errorf("%w: location parent cycle or unresolved graph", domain.ErrValidation)
		}
	}
	return out, nil
}

// ReconcileSequentialNextSeqAfterImport raises next_seq when imported or existing item IDs exceed the counter.
func (s *Inventory) ReconcileSequentialNextSeqAfterImport(ctx context.Context) error {
	if s.Settings == nil || s.Items == nil {
		return nil
	}
	pol, err := s.GetItemIDPolicy(ctx)
	if err != nil {
		return err
	}
	pol = pol.Normalize()
	if pol.Kind != domain.ItemIDKindSequential {
		return nil
	}
	items, err := s.Items.ListAllExport(ctx)
	if err != nil {
		return err
	}
	var maxSeq int64
	for _, it := range items {
		if seq, ok := parseSequentialItemSeq(it.ID, pol); ok && seq > maxSeq {
			maxSeq = seq
		}
	}
	want := maxSeq + 1
	if want < pol.NextSeq {
		want = pol.NextSeq
	}
	if want != pol.NextSeq {
		pol.NextSeq = want
		return s.saveItemIDPolicy(ctx, pol)
	}
	return nil
}

func parseSequentialItemSeq(id string, pol domain.ItemIDPolicy) (int64, bool) {
	pol = pol.Normalize()
	if pol.Kind != domain.ItemIDKindSequential {
		return 0, false
	}
	prefix := regexp.QuoteMeta(strings.TrimSpace(pol.Prefix))
	w := pol.Width
	if prefix == "" || w < 1 {
		return 0, false
	}
	re1, err := regexp.Compile(fmt.Sprintf(`^%s_(\d{%d})$`, prefix, w))
	if err != nil {
		return 0, false
	}
	if m := re1.FindStringSubmatch(strings.TrimSpace(id)); len(m) == 2 {
		v, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}
	re2, err := regexp.Compile(fmt.Sprintf(`^%s(\d{%d})$`, prefix, w))
	if err != nil {
		return 0, false
	}
	if m := re2.FindStringSubmatch(strings.TrimSpace(id)); len(m) == 2 {
		v, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	}
	return 0, false
}

// PostInventoryImportSearchRefresh updates denormalized search columns after a committed import.
func (s *Inventory) PostInventoryImportSearchRefresh(ctx context.Context, bundle *InventoryExportBundle) error {
	if bundle == nil {
		return nil
	}
	for _, it := range bundle.Items {
		id := strings.TrimSpace(it.ID)
		if id == "" {
			continue
		}
		if err := s.RefreshItemSearchDenorm(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// --- CSV / ZIP (same logical bundle) ---

const (
	csvFileManifest      = "manifest.json"
	csvFileLocations     = "locations.csv"
	csvFileLabels        = "labels.csv"
	csvFileItemTemplates = "item_templates.csv"
	csvFileItems         = "items.csv"
	csvFileItemLabels    = "item_labels.csv"
)

type csvManifest struct {
	ExportVersion int       `json:"export_version"`
	ExportedAt    time.Time `json:"exported_at"`
}

// EncodeInventoryCSVZIP builds a zip with manifest + CSV tables from a bundle.
func EncodeInventoryCSVZIP(bundle *InventoryExportBundle) ([]byte, error) {
	// implemented in inventory_export_csv.go
	return encodeInventoryCSVZIP(bundle)
}

// DecodeInventoryCSVZIP parses an export zip back into a bundle.
func DecodeInventoryCSVZIP(data []byte) (*InventoryExportBundle, error) {
	return decodeInventoryCSVZIP(data)
}
