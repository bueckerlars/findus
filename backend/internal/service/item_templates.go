package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"findus/backend/internal/domain"
)

var templateSlugRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func (s *Inventory) resolveDefaultTemplate(ctx context.Context, t string) (*string, error) {
	if s.Templates == nil {
		return nil, fmt.Errorf("%w: templates not configured", domain.ErrValidation)
	}
	if _, err := s.Templates.GetByID(ctx, t); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, fmt.Errorf("%w: default_template_type", domain.ErrValidation)
		}
		return nil, err
	}
	return &t, nil
}

func (s *Inventory) ensureTemplateType(ctx context.Context, tt domain.TemplateType) error {
	if s.Templates == nil {
		return fmt.Errorf("%w: templates not configured", domain.ErrValidation)
	}
	if _, err := s.Templates.GetByID(ctx, string(tt)); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fmt.Errorf("%w: template_type", domain.ErrValidation)
		}
		return err
	}
	return nil
}

// ListItemTemplates returns all templates ordered by sort_order then id.
func (s *Inventory) ListItemTemplates(ctx context.Context) ([]domain.ItemTemplate, error) {
	return s.Templates.List(ctx)
}

// GetItemTemplate loads one template by id (slug).
func (s *Inventory) GetItemTemplate(ctx context.Context, id string) (*domain.ItemTemplate, error) {
	return s.Templates.GetByID(ctx, id)
}

// CreateItemTemplate inserts a new template; id must match templateSlugRE.
func (s *Inventory) CreateItemTemplate(ctx context.Context, id, displayName string, fieldsJSON []byte, sortOrder int) error {
	id = strings.TrimSpace(id)
	if !templateSlugRE.MatchString(id) {
		return fmt.Errorf("%w: id", domain.ErrValidation)
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || utf8.RuneCountInString(displayName) > 128 {
		return fmt.Errorf("%w: display_name", domain.ErrValidation)
	}
	fields, err := domain.ParseTemplateFieldsJSON(fieldsJSON)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrValidation, err)
	}
	canon, err := domain.MarshalTemplateFields(fields)
	if err != nil {
		return err
	}
	if _, err := s.Templates.GetByID(ctx, id); err == nil {
		return fmt.Errorf("%w: id already exists", domain.ErrConflict)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	now := time.Now().UTC()
	return s.Templates.Create(ctx, &domain.ItemTemplate{
		ID:          id,
		DisplayName: displayName,
		FieldsJSON:  canon,
		SortOrder:   sortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

// UpdateItemTemplate updates display name, field schema JSON, and sort order for an existing id.
func (s *Inventory) UpdateItemTemplate(ctx context.Context, id, displayName string, fieldsJSON []byte, sortOrder int) error {
	t, err := s.Templates.GetByID(ctx, id)
	if err != nil {
		return err
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" || utf8.RuneCountInString(displayName) > 128 {
		return fmt.Errorf("%w: display_name", domain.ErrValidation)
	}
	fields, err := domain.ParseTemplateFieldsJSON(fieldsJSON)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrValidation, err)
	}
	canon, err := domain.MarshalTemplateFields(fields)
	if err != nil {
		return err
	}
	t.DisplayName = displayName
	t.FieldsJSON = canon
	t.Fields = fields
	t.SortOrder = sortOrder
	t.UpdatedAt = time.Now().UTC()
	return s.Templates.Update(ctx, t)
}

// DeleteItemTemplate removes a template. Items using it are reassigned to another template
// (lowest sort_order, then id) with template_data reset to {}. Labels referencing it as default lose that default.
func (s *Inventory) DeleteItemTemplate(ctx context.Context, id string) error {
	if s.Templates == nil || s.Items == nil {
		return fmt.Errorf("%w: not configured", domain.ErrValidation)
	}
	list, err := s.Templates.List(ctx)
	if err != nil {
		return err
	}
	found := false
	for _, t := range list {
		if t.ID == id {
			found = true
			break
		}
	}
	if !found {
		return domain.ErrNotFound
	}
	cnt, err := s.Items.CountByTemplateType(ctx, id)
	if err != nil {
		return err
	}
	if cnt > 0 {
		toID, ok := pickReassignTargetID(list, id)
		if !ok {
			return fmt.Errorf("%w: cannot delete the only template while items still use it", domain.ErrValidation)
		}
		if err := s.Items.ReassignTemplateType(ctx, id, toID); err != nil {
			return err
		}
	}
	if s.Labels != nil {
		_ = s.Labels.ClearDefaultTemplateType(ctx, id)
	}
	return s.Templates.Delete(ctx, id)
}
