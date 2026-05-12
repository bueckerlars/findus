package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"findus/backend/internal/domain"
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
	_ = r.deleteSubstrFTSForItemID(ctx, id)
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
	return r.searchItemsAdvanced(ctx, q, limit)
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
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM items WHERE template_type = ?`, fromID)
	if err != nil {
		return err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE items SET template_type = ?, template_data = '{}', updated_at = ? WHERE template_type = ?`,
		toID, now, fromID); err != nil {
		return err
	}
	for _, id := range ids {
		if err := r.upsertSubstrFTS(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
