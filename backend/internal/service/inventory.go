package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository"
)

type Inventory struct {
	Locations    repository.LocationRepository
	Items        repository.ItemRepository
	ItemQRTokens repository.ItemQRTokenReservationRepository
	Labels       repository.LabelRepository
	Templates    repository.ItemTemplateRepository
	Settings     repository.SettingsRepository

	// ItemAttachments is optional for tests; when nil, attachment APIs should not be registered.
	ItemAttachments repository.ItemAttachmentRepository
	// AttachmentPostProcessor optional hook after successful create (e.g. OCR); nil uses a no-op.
	AttachmentPostProcessor AttachmentPostProcessor
}

func (s *Inventory) CreateLocation(ctx context.Context, name, description string, parentID *string) (*domain.Location, error) {
	name = trim(name, 1, 200)
	if name == "" {
		return nil, fmt.Errorf("%w: name", domain.ErrValidation)
	}
	if parentID != nil && *parentID != "" {
		if _, err := s.Locations.GetByID(ctx, *parentID); err != nil {
			return nil, err
		}
	} else {
		parentID = nil
	}
	now := time.Now().UTC()
	loc := &domain.Location{
		ID:          newID(),
		Name:        name,
		ParentID:    parentID,
		Description: trim(description, 0, 2000),
		QRToken:     newID(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.Locations.Create(ctx, loc); err != nil {
		return nil, err
	}
	return loc, nil
}

func (s *Inventory) UpdateLocation(ctx context.Context, id, name, description string, parentID *string) (*domain.Location, error) {
	l, err := s.Locations.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	name = trim(name, 1, 200)
	if name == "" {
		return nil, fmt.Errorf("%w: name", domain.ErrValidation)
	}
	if parentID != nil && *parentID == id {
		return nil, fmt.Errorf("%w: parent", domain.ErrValidation)
	}
	if parentID != nil && *parentID != "" {
		if err := s.assertNoCycle(ctx, id, *parentID); err != nil {
			return nil, err
		}
		if _, err := s.Locations.GetByID(ctx, *parentID); err != nil {
			return nil, err
		}
	} else {
		parentID = nil
	}
	l.Name = name
	l.Description = trim(description, 0, 2000)
	l.ParentID = parentID
	l.UpdatedAt = time.Now().UTC()
	if err := s.Locations.Update(ctx, l); err != nil {
		return nil, err
	}
	if err := s.refreshSearchLocationsForSubtree(ctx, l.ID); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Inventory) assertNoCycle(ctx context.Context, locID, newParentID string) error {
	cur := newParentID
	for i := 0; i < 1000; i++ {
		if cur == locID {
			return fmt.Errorf("%w: cycle", domain.ErrValidation)
		}
		p, err := s.Locations.GetByID(ctx, cur)
		if err != nil {
			return err
		}
		if p.ParentID == nil || *p.ParentID == "" {
			return nil
		}
		cur = *p.ParentID
	}
	return fmt.Errorf("%w: depth", domain.ErrValidation)
}

func (s *Inventory) DeleteLocation(ctx context.Context, id string) error {
	ch, err := s.Locations.ListChildren(ctx, &id)
	if err != nil {
		return err
	}
	if len(ch) > 0 {
		return fmt.Errorf("%w: has children", domain.ErrValidation)
	}
	items, err := s.Items.ListByLocation(ctx, id)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		return fmt.Errorf("%w: has items", domain.ErrValidation)
	}
	return s.Locations.Delete(ctx, id)
}

func (s *Inventory) validateLabelRefs(ctx context.Context, labelIDs []string) error {
	if s.Labels == nil {
		if len(labelIDs) > 0 {
			return fmt.Errorf("%w: labels not configured", domain.ErrValidation)
		}
		return nil
	}
	for _, id := range labelIDs {
		if _, err := s.Labels.GetByID(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Inventory) CreateItem(ctx context.Context, name, description, locationID string, tt domain.TemplateType, td, additional json.RawMessage, labelIDs []string) (*domain.Item, error) {
	name = trim(name, 1, 200)
	if name == "" {
		return nil, fmt.Errorf("%w: name", domain.ErrValidation)
	}
	if _, err := s.Locations.GetByID(ctx, locationID); err != nil {
		return nil, err
	}
	if err := s.ensureTemplateType(ctx, tt); err != nil {
		return nil, err
	}
	if len(additional) == 0 {
		additional = json.RawMessage(`{}`)
	}
	if err := ValidateAdditionalJSON(additional); err != nil {
		return nil, err
	}
	ids := dedupeStrings(labelIDs)
	if err := s.validateLabelRefs(ctx, ids); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var it *domain.Item
	var createErr error
	for attempt := 0; attempt < 24; attempt++ {
		itemID, err := s.allocateNextItemID(ctx)
		if err != nil {
			return nil, err
		}
		cand := &domain.Item{
			ID:             itemID,
			Name:           name,
			Description:    trim(description, 0, 5000),
			LocationID:     locationID,
			TemplateType:   tt,
			TemplateData:   td,
			AdditionalData: additional,
			QRToken:        newID(),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if s.ItemQRTokens != nil {
			reservedToken, ok, err := s.ItemQRTokens.GetTokenByItemID(ctx, itemID)
			if err != nil {
				return nil, err
			}
			if ok {
				cand.QRToken = reservedToken
			}
		}
		createErr = s.Items.Create(ctx, cand)
		if createErr == nil {
			it = cand
			if err := s.sequentialBumpAfterSuccessfulCreate(ctx, itemID); err != nil {
				return nil, err
			}
			break
		}
		if !isSQLiteUniqueViolation(createErr) {
			return nil, createErr
		}
	}
	if it == nil {
		if createErr == nil {
			createErr = fmt.Errorf("%w: item id", domain.ErrValidation)
		}
		return nil, createErr
	}
	if err := s.Items.ReplaceItemLabels(ctx, it.ID, ids); err != nil {
		return nil, err
	}
	if err := s.refreshItemSearchDenorm(ctx, it.ID); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *Inventory) UpdateItem(ctx context.Context, id, name, description, locationID string, tt domain.TemplateType, td, additional json.RawMessage, labelIDs []string) (*domain.Item, error) {
	it, err := s.Items.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.Locations.GetByID(ctx, locationID); err != nil {
		return nil, err
	}
	name = trim(name, 1, 200)
	if name == "" {
		return nil, fmt.Errorf("%w: name", domain.ErrValidation)
	}
	if len(additional) == 0 {
		additional = json.RawMessage(`{}`)
	}
	if err := ValidateAdditionalJSON(additional); err != nil {
		return nil, err
	}
	ids := dedupeStrings(labelIDs)
	if err := s.validateLabelRefs(ctx, ids); err != nil {
		return nil, err
	}
	if err := s.ensureTemplateType(ctx, tt); err != nil {
		return nil, err
	}
	it.Name = name
	it.Description = trim(description, 0, 5000)
	it.LocationID = locationID
	it.TemplateType = tt
	it.TemplateData = td
	it.AdditionalData = additional
	it.UpdatedAt = time.Now().UTC()
	if err := s.Items.Update(ctx, it); err != nil {
		return nil, err
	}
	if err := s.Items.ReplaceItemLabels(ctx, it.ID, ids); err != nil {
		return nil, err
	}
	if err := s.refreshItemSearchDenorm(ctx, it.ID); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *Inventory) DeleteItem(ctx context.Context, id string) error {
	return s.Items.Delete(ctx, id)
}

func (s *Inventory) ListLabels(ctx context.Context) ([]domain.Label, error) {
	if s.Labels == nil {
		return nil, nil
	}
	return s.Labels.ListAll(ctx)
}

func (s *Inventory) CreateLabel(ctx context.Context, name, color string, defaultTemplate *string) (*domain.Label, error) {
	if s.Labels == nil {
		return nil, fmt.Errorf("%w: labels", domain.ErrValidation)
	}
	name = trim(name, 1, 120)
	if name == "" {
		return nil, fmt.Errorf("%w: name", domain.ErrValidation)
	}
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#64748b"
	}
	if utf8.RuneCountInString(color) > 32 {
		return nil, fmt.Errorf("%w: color", domain.ErrValidation)
	}
	var dtp *string
	if defaultTemplate != nil {
		t := strings.TrimSpace(*defaultTemplate)
		if t != "" {
			ref, err := s.resolveDefaultTemplate(ctx, t)
			if err != nil {
				return nil, err
			}
			dtp = ref
		}
	}
	now := time.Now().UTC()
	l := &domain.Label{
		ID:                  newID(),
		Name:                name,
		Color:               color,
		DefaultTemplateType: dtp,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.Labels.Create(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Inventory) UpdateLabel(ctx context.Context, id, name, color string, defaultTemplate *string) (*domain.Label, error) {
	if s.Labels == nil {
		return nil, fmt.Errorf("%w: labels", domain.ErrValidation)
	}
	l, err := s.Labels.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	name = trim(name, 1, 120)
	if name == "" {
		return nil, fmt.Errorf("%w: name", domain.ErrValidation)
	}
	color = strings.TrimSpace(color)
	if color == "" {
		color = "#64748b"
	}
	if utf8.RuneCountInString(color) > 32 {
		return nil, fmt.Errorf("%w: color", domain.ErrValidation)
	}
	var dtp *string
	if defaultTemplate != nil {
		t := strings.TrimSpace(*defaultTemplate)
		if t != "" {
			ref, err := s.resolveDefaultTemplate(ctx, t)
			if err != nil {
				return nil, err
			}
			dtp = ref
		}
	}
	l.Name = name
	l.Color = color
	l.DefaultTemplateType = dtp
	l.UpdatedAt = time.Now().UTC()
	if err := s.Labels.Update(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Inventory) DeleteLabel(ctx context.Context, id string) error {
	if s.Labels == nil {
		return fmt.Errorf("%w: labels", domain.ErrValidation)
	}
	return s.Labels.Delete(ctx, id)
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func trim(s string, minLen, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen > 0 && len(s) > maxLen {
		s = s[:maxLen]
	}
	_ = minLen
	return s
}
