package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"findus/internal/domain"
)

type LocationRepo struct{ db *sql.DB }

func NewLocationRepo(db *sql.DB) *LocationRepo { return &LocationRepo{db: db} }

func (r *LocationRepo) Create(ctx context.Context, l *domain.Location) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO locations (id, name, parent_id, description, qr_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		l.ID, l.Name, strPtr(l.ParentID), l.Description, l.QRToken, formatTime(l.CreatedAt), formatTime(l.UpdatedAt))
	return err
}

func (r *LocationRepo) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, parent_id, description, qr_token, created_at, updated_at FROM locations WHERE id = ?`, id)
	return scanLocation(row)
}

func (r *LocationRepo) GetByQRToken(ctx context.Context, token string) (*domain.Location, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, parent_id, description, qr_token, created_at, updated_at FROM locations WHERE qr_token = ?`, token)
	return scanLocation(row)
}

func (r *LocationRepo) Update(ctx context.Context, l *domain.Location) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE locations SET name=?, parent_id=?, description=?, updated_at=? WHERE id=?`,
		l.Name, strPtr(l.ParentID), l.Description, formatTime(l.UpdatedAt), l.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LocationRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM locations WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *LocationRepo) ListChildren(ctx context.Context, parentID *string) ([]domain.Location, error) {
	var rows *sql.Rows
	var err error
	if parentID == nil {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, name, parent_id, description, qr_token, created_at, updated_at FROM locations WHERE parent_id IS NULL ORDER BY name`)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, name, parent_id, description, qr_token, created_at, updated_at FROM locations WHERE parent_id = ? ORDER BY name`, *parentID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Location
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *LocationRepo) ListRecentByUpdated(ctx context.Context, limit int) ([]domain.Location, error) {
	if limit <= 0 || limit > 100 {
		limit = 12
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, parent_id, description, qr_token, created_at, updated_at
		FROM locations ORDER BY updated_at DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Location
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *LocationRepo) ListAll(ctx context.Context, limit int) ([]domain.Location, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, parent_id, description, qr_token, created_at, updated_at
		FROM locations ORDER BY name LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Location
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func (r *LocationRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM locations`).Scan(&n)
	return n, err
}

func (r *LocationRepo) ChildCountsByParentID(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT parent_id, COUNT(*) FROM locations WHERE parent_id IS NOT NULL GROUP BY parent_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int64)
	for rows.Next() {
		var parentID string
		var n int64
		if err := rows.Scan(&parentID, &n); err != nil {
			return nil, err
		}
		out[parentID] = n
	}
	return out, rows.Err()
}

func (r *LocationRepo) PathFromRoot(ctx context.Context, id string) ([]domain.LocationPathElement, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH RECURSIVE anc(id, name, parent_id, depth) AS (
			SELECT id, name, parent_id, 0 FROM locations WHERE id = ?
			UNION ALL
			SELECT l.id, l.name, l.parent_id, anc.depth + 1 FROM locations l INNER JOIN anc ON l.id = anc.parent_id
		)
		SELECT id, name FROM anc ORDER BY depth DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.LocationPathElement
	for rows.Next() {
		var e domain.LocationPathElement
		if err := rows.Scan(&e.ID, &e.Name); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, domain.ErrNotFound
	}
	return out, nil
}

func scanLocation(sc interface {
	Scan(dest ...any) error
}) (*domain.Location, error) {
	var l domain.Location
	var parent sql.NullString
	var created, updated string
	err := sc.Scan(&l.ID, &l.Name, &parent, &l.Description, &l.QRToken, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	l.ParentID = sqlNullString(parent)
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
