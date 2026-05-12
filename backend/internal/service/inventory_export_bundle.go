package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"findus/backend/internal/domain"
)

const inventoryExportVersion = 1

// InventoryExportBundle is the canonical JSON shape for admin inventory export/import.
type InventoryExportBundle struct {
	ExportVersion int                       `json:"export_version"`
	ExportedAt    time.Time                 `json:"exported_at"`
	Locations     []domain.Location         `json:"locations"`
	Labels        []domain.Label            `json:"labels"`
	ItemTemplates []InventoryExportTemplate `json:"item_templates"`
	Items         []InventoryExportItem     `json:"items"`
}

// InventoryExportTemplate is stored template metadata + fields_json (no duplicate Fields slice).
type InventoryExportTemplate struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"display_name"`
	FieldsJSON  json.RawMessage `json:"fields_json"`
	SortOrder   int             `json:"sort_order"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func exportTemplateFromDomain(t domain.ItemTemplate) InventoryExportTemplate {
	return InventoryExportTemplate{
		ID:          t.ID,
		DisplayName: t.DisplayName,
		FieldsJSON:  t.FieldsJSON,
		SortOrder:   t.SortOrder,
		CreatedAt:   t.CreatedAt.UTC(),
		UpdatedAt:   t.UpdatedAt.UTC(),
	}
}

// InventoryExportItem is one item plus label membership for round-trip JSON.
type InventoryExportItem struct {
	domain.Item
	LabelIDs []string `json:"label_ids"`
}

func (it InventoryExportItem) MarshalJSON() ([]byte, error) {
	type alias struct {
		ID             string          `json:"id"`
		Name           string          `json:"name"`
		Description    string          `json:"description"`
		LocationID     string          `json:"location_id"`
		TemplateType   string          `json:"template_type"`
		TemplateData   json.RawMessage `json:"template_data"`
		AdditionalData json.RawMessage `json:"additional_data"`
		PhotoPath      *string         `json:"photo_path,omitempty"`
		QRToken        string          `json:"qr_token"`
		CreatedAt      time.Time       `json:"created_at"`
		UpdatedAt      time.Time       `json:"updated_at"`
		LabelIDs       []string        `json:"label_ids"`
	}
	a := alias{
		ID:             it.ID,
		Name:           it.Name,
		Description:    it.Description,
		LocationID:     it.LocationID,
		TemplateType:   string(it.TemplateType),
		TemplateData:   it.TemplateData,
		AdditionalData: it.AdditionalData,
		PhotoPath:      it.PhotoPath,
		QRToken:        it.QRToken,
		CreatedAt:      it.CreatedAt,
		UpdatedAt:      it.UpdatedAt,
		LabelIDs:       it.LabelIDs,
	}
	return json.Marshal(a)
}

func (it *InventoryExportItem) UnmarshalJSON(b []byte) error {
	type alias struct {
		ID             string          `json:"id"`
		Name           string          `json:"name"`
		Description    string          `json:"description"`
		LocationID     string          `json:"location_id"`
		TemplateType   string          `json:"template_type"`
		TemplateData   json.RawMessage `json:"template_data"`
		AdditionalData json.RawMessage `json:"additional_data"`
		PhotoPath      *string         `json:"photo_path"`
		QRToken        string          `json:"qr_token"`
		CreatedAt      time.Time       `json:"created_at"`
		UpdatedAt      time.Time       `json:"updated_at"`
		LabelIDs       []string        `json:"label_ids"`
	}
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	it.ID = strings.TrimSpace(a.ID)
	it.Name = a.Name
	it.Description = a.Description
	it.LocationID = strings.TrimSpace(a.LocationID)
	it.TemplateType = domain.TemplateType(strings.TrimSpace(a.TemplateType))
	it.TemplateData = a.TemplateData
	it.AdditionalData = a.AdditionalData
	it.PhotoPath = a.PhotoPath
	it.QRToken = strings.TrimSpace(a.QRToken)
	it.CreatedAt = a.CreatedAt.UTC()
	it.UpdatedAt = a.UpdatedAt.UTC()
	it.LabelIDs = a.LabelIDs
	return nil
}

// InventoryImportResult summarizes an import run.
type InventoryImportResult struct {
	TemplatesCreated   int `json:"templates_created"`
	TemplatesUpdated   int `json:"templates_updated"`
	LabelsCreated      int `json:"labels_created"`
	LabelsUpdated      int `json:"labels_updated"`
	LocationsCreated   int `json:"locations_created"`
	LocationsUpdated   int `json:"locations_updated"`
	ItemsCreated       int `json:"items_created"`
	ItemsUpdated       int `json:"items_updated"`
	ItemLabelsReplaced int `json:"item_labels_replaced"`
}

// ValidateInventoryImportBundle checks version and required fields before DB writes.
func ValidateInventoryImportBundle(b *InventoryExportBundle) error {
	if b == nil {
		return fmt.Errorf("%w: empty bundle", domain.ErrValidation)
	}
	if b.ExportVersion != inventoryExportVersion {
		return fmt.Errorf("%w: unsupported export_version %d (want %d)", domain.ErrValidation, b.ExportVersion, inventoryExportVersion)
	}
	if b.ExportedAt.IsZero() {
		return fmt.Errorf("%w: exported_at required", domain.ErrValidation)
	}
	return nil
}

// ErrLabelNameConflict is returned when an import would violate labels.name uniqueness.
var ErrLabelNameConflict = errors.New("label name conflicts with a different existing id")
