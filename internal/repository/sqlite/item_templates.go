package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"findus/internal/domain"
)

type ItemTemplateRepo struct{ db *sql.DB }

func NewItemTemplateRepo(db *sql.DB) *ItemTemplateRepo { return &ItemTemplateRepo{db: db} }

func (r *ItemTemplateRepo) List(ctx context.Context) ([]domain.ItemTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, display_name, fields_json, sort_order, created_at, updated_at
		FROM item_templates ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ItemTemplate
	for rows.Next() {
		t, err := scanItemTemplate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *ItemTemplateRepo) GetByID(ctx context.Context, id string) (*domain.ItemTemplate, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, display_name, fields_json, sort_order, created_at, updated_at
		FROM item_templates WHERE id = ?`, id)
	return scanItemTemplate(row)
}

func (r *ItemTemplateRepo) Create(ctx context.Context, t *domain.ItemTemplate) error {
	fj := t.FieldsJSON
	if len(fj) == 0 {
		fj = []byte("[]")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO item_templates (id, display_name, fields_json, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.DisplayName, string(fj), t.SortOrder, formatTime(t.CreatedAt), formatTime(t.UpdatedAt))
	return err
}

func (r *ItemTemplateRepo) Update(ctx context.Context, t *domain.ItemTemplate) error {
	fj := t.FieldsJSON
	if len(fj) == 0 {
		fj = []byte("[]")
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE item_templates SET display_name=?, fields_json=?, sort_order=?, updated_at=? WHERE id=?`,
		t.DisplayName, string(fj), t.SortOrder, formatTime(t.UpdatedAt), t.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ItemTemplateRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM item_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanItemTemplate(sc interface {
	Scan(dest ...any) error
}) (*domain.ItemTemplate, error) {
	var t domain.ItemTemplate
	var fj string
	var created, updated string
	err := sc.Scan(&t.ID, &t.DisplayName, &fj, &t.SortOrder, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.FieldsJSON = json.RawMessage(fj)
	fields, err := domain.ParseTemplateFieldsJSON(t.FieldsJSON)
	if err != nil {
		return nil, fmt.Errorf("item_template %q: %w", t.ID, err)
	}
	t.Fields = fields
	t.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	t.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return nil, fmt.Errorf("updated_at: %w", err)
	}
	return &t, nil
}
