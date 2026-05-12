package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"findus/backend/internal/domain"
)

type LabelRepo struct{ db *sql.DB }

func NewLabelRepo(db *sql.DB) *LabelRepo { return &LabelRepo{db: db} }

func (r *LabelRepo) ListAll(ctx context.Context) ([]domain.Label, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, color, default_template_type, created_at, updated_at FROM labels ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Label
	for rows.Next() {
		l, err := scanLabel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *LabelRepo) GetByID(ctx context.Context, id string) (*domain.Label, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, color, default_template_type, created_at, updated_at FROM labels WHERE id = ?`, id)
	return scanLabel(row)
}

func (r *LabelRepo) Create(ctx context.Context, l *domain.Label) error {
	var dtp any
	if l.DefaultTemplateType != nil {
		dtp = *l.DefaultTemplateType
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO labels (id, name, color, default_template_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		l.ID, l.Name, l.Color, dtp, formatTime(l.CreatedAt), formatTime(l.UpdatedAt))
	return err
}

func (r *LabelRepo) Update(ctx context.Context, l *domain.Label) error {
	var dtp any
	if l.DefaultTemplateType != nil {
		dtp = *l.DefaultTemplateType
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE labels SET name=?, color=?, default_template_type=?, updated_at=? WHERE id=?`,
		l.Name, l.Color, dtp, formatTime(l.UpdatedAt), l.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LabelRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM labels WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LabelRepo) ClearDefaultTemplateType(ctx context.Context, templateID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE labels SET default_template_type = NULL, updated_at = ? WHERE default_template_type = ?`,
		formatTime(time.Now().UTC()), templateID)
	return err
}

func (r *LabelRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM labels`).Scan(&n)
	return n, err
}

func scanLabel(sc interface {
	Scan(dest ...any) error
}) (*domain.Label, error) {
	var l domain.Label
	var dtp sql.NullString
	var created, updated string
	err := sc.Scan(&l.ID, &l.Name, &l.Color, &dtp, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if dtp.Valid {
		s := dtp.String
		l.DefaultTemplateType = &s
	}
	l.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	l.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return nil, fmt.Errorf("updated_at: %w", err)
	}
	return &l, nil
}
