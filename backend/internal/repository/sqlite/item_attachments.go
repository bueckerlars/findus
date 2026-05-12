package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"findus/backend/internal/domain"
)

type ItemAttachmentRepo struct{ db DBConn }

func NewItemAttachmentRepo(db *sql.DB) *ItemAttachmentRepo { return &ItemAttachmentRepo{db: db} }

func NewItemAttachmentRepoConn(c DBConn) *ItemAttachmentRepo { return &ItemAttachmentRepo{db: c} }

func (r *ItemAttachmentRepo) ListByItemID(ctx context.Context, itemID string) ([]domain.ItemAttachment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, item_id, original_filename, storage_path, mime_type, size_bytes, title, metadata_json, created_at
		FROM item_attachments WHERE item_id = ? ORDER BY created_at ASC`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ItemAttachment
	for rows.Next() {
		a, err := scanItemAttachment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func (r *ItemAttachmentRepo) GetByID(ctx context.Context, id string) (*domain.ItemAttachment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, item_id, original_filename, storage_path, mime_type, size_bytes, title, metadata_json, created_at
		FROM item_attachments WHERE id = ?`, id)
	return scanItemAttachment(row)
}

func (r *ItemAttachmentRepo) Create(ctx context.Context, a *domain.ItemAttachment) error {
	meta := a.MetadataJSON
	if len(meta) == 0 {
		meta = json.RawMessage(`{}`)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO item_attachments (id, item_id, original_filename, storage_path, mime_type, size_bytes, title, metadata_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ItemID, a.OriginalFilename, a.StoragePath, a.MimeType, a.SizeBytes, a.Title, string(meta), formatTime(a.CreatedAt))
	return err
}

func (r *ItemAttachmentRepo) UpdateTitle(ctx context.Context, id string, title string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE item_attachments SET title = ? WHERE id = ?`, title, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ItemAttachmentRepo) DeleteByID(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM item_attachments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanItemAttachment(sc interface {
	Scan(dest ...any) error
}) (*domain.ItemAttachment, error) {
	var (
		a         domain.ItemAttachment
		meta      string
		createdAt string
	)
	if err := sc.Scan(&a.ID, &a.ItemID, &a.OriginalFilename, &a.StoragePath, &a.MimeType, &a.SizeBytes, &a.Title, &meta, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan item attachment: %w", err)
	}
	t, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	a.CreatedAt = t
	a.MetadataJSON = json.RawMessage(meta)
	if len(a.MetadataJSON) == 0 {
		a.MetadataJSON = json.RawMessage(`{}`)
	}
	return &a, nil
}
