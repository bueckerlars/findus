package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"findus/internal/domain"
)

type ItemRepo struct{ db *sql.DB }

func NewItemRepo(db *sql.DB) *ItemRepo { return &ItemRepo{db: db} }

func (r *ItemRepo) Create(ctx context.Context, it *domain.Item) error {
	ad := it.AdditionalData
	if len(ad) == 0 {
		ad = json.RawMessage(`{}`)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		it.ID, it.Name, it.Description, it.LocationID, string(it.TemplateType), string(it.TemplateData), string(ad), strPtr(it.PhotoPath), it.QRToken, formatTime(it.CreatedAt), formatTime(it.UpdatedAt))
	return err
}

func (r *ItemRepo) GetByID(ctx context.Context, id string) (*domain.Item, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at FROM items WHERE id = ?`, id)
	return scanItem(row)
}

func (r *ItemRepo) GetByQRToken(ctx context.Context, token string) (*domain.Item, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at FROM items WHERE qr_token = ?`, token)
	return scanItem(row)
}

func (r *ItemRepo) Update(ctx context.Context, it *domain.Item) error {
	ad := it.AdditionalData
	if len(ad) == 0 {
		ad = json.RawMessage(`{}`)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE items SET name=?, description=?, location_id=?, template_type=?, template_data=?, additional_data=?, photo_path=?, updated_at=? WHERE id=?`,
		it.Name, it.Description, it.LocationID, string(it.TemplateType), string(it.TemplateData), string(ad), strPtr(it.PhotoPath), formatTime(it.UpdatedAt), it.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ItemRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ItemRepo) ListByLocation(ctx context.Context, locationID string) ([]domain.Item, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at FROM items WHERE location_id = ? ORDER BY name`, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	return out, rows.Err()
}

func (r *ItemRepo) Search(ctx context.Context, q string, limit int) ([]domain.Item, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	match := ftsMatchQuery(q)
	if match == "" {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT i.id, i.name, i.description, i.location_id, i.template_type, i.template_data, i.additional_data, i.photo_path, i.qr_token, i.created_at, i.updated_at
		FROM items i
		INNER JOIN items_fts ON i.rowid = items_fts.rowid
		WHERE items_fts MATCH ?
		ORDER BY i.name
		LIMIT ?`, match, limit)
	if err != nil {
		return r.searchLike(ctx, q, limit)
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	if err := rows.Err(); err != nil {
		return r.searchLike(ctx, q, limit)
	}
	return out, nil
}

func (r *ItemRepo) searchLike(ctx context.Context, q string, limit int) ([]domain.Item, error) {
	pat := "%" + strings.ReplaceAll(q, "%", "\\%") + "%"
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT i.id, i.name, i.description, i.location_id, i.template_type, i.template_data, i.additional_data, i.photo_path, i.qr_token, i.created_at, i.updated_at
		FROM items i
		WHERE i.name LIKE ? ESCAPE '\\' OR i.description LIKE ? ESCAPE '\\'
		OR i.additional_data LIKE ? ESCAPE '\\'
		OR EXISTS (
			SELECT 1 FROM item_labels il INNER JOIN labels lb ON lb.id = il.label_id
			WHERE il.item_id = i.id AND lb.name LIKE ? ESCAPE '\\'
		)
		ORDER BY i.name LIMIT ?`, pat, pat, pat, pat, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	return out, rows.Err()
}

func (r *ItemRepo) ListAll(ctx context.Context, limit int) ([]domain.Item, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at
		FROM items ORDER BY name LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	return out, rows.Err()
}

func (r *ItemRepo) ListRecentByUpdated(ctx context.Context, limit int) ([]domain.Item, error) {
	if limit <= 0 || limit > 100 {
		limit = 12
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at
		FROM items ORDER BY updated_at DESC, id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Item
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *it)
	}
	return out, rows.Err()
}

func (r *ItemRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM items`).Scan(&n)
	return n, err
}

func (r *ItemRepo) ReplaceItemLabels(ctx context.Context, itemID string, labelIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM item_labels WHERE item_id = ?`, itemID); err != nil {
		return err
	}
	for _, lid := range labelIDs {
		if lid == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO item_labels (item_id, label_id) VALUES (?, ?)`, itemID, lid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *ItemRepo) ListLabelsForItem(ctx context.Context, itemID string) ([]domain.Label, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.name, l.color, l.default_template_type, l.created_at, l.updated_at
		FROM labels l
		INNER JOIN item_labels il ON il.label_id = l.id
		WHERE il.item_id = ?
		ORDER BY l.name`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Label
	for rows.Next() {
		lb, err := scanLabel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *lb)
	}
	return out, rows.Err()
}

func ftsMatchQuery(q string) string {
	parts := strings.Fields(q)
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for i, p := range parts {
		p = strings.ReplaceAll(p, `"`, `""`)
		if i > 0 {
			b.WriteString(" OR ")
		}
		b.WriteByte('"')
		b.WriteString(p)
		b.WriteByte('*')
		b.WriteByte('"')
	}
	return b.String()
}

func scanItem(sc interface {
	Scan(dest ...any) error
}) (*domain.Item, error) {
	var it domain.Item
	var photo sql.NullString
	var td, ad string
	var created, updated string
	err := sc.Scan(&it.ID, &it.Name, &it.Description, &it.LocationID, &it.TemplateType, &td, &ad, &photo, &it.QRToken, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	it.TemplateData = json.RawMessage(td)
	it.AdditionalData = json.RawMessage(ad)
	if photo.Valid {
		s := photo.String
		it.PhotoPath = &s
	}
	it.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	it.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return nil, fmt.Errorf("updated_at: %w", err)
	}
	return &it, nil
}

func (r *ItemRepo) CountByTemplateType(ctx context.Context, templateType string) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM items WHERE template_type = ?`, templateType).Scan(&n)
	return n, err
}

// ReassignTemplateType moves all items from fromID to toID and resets template_data to {}.
func (r *ItemRepo) ReassignTemplateType(ctx context.Context, fromID, toID string) error {
	now := formatTime(time.Now().UTC())
	_, err := r.db.ExecContext(ctx, `
		UPDATE items SET template_type = ?, template_data = '{}', updated_at = ? WHERE template_type = ?`,
		toID, now, fromID)
	return err
}
