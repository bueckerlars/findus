package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"findus/internal/domain"
)

type InviteRepo struct{ db *sql.DB }

func NewInviteRepo(db *sql.DB) *InviteRepo { return &InviteRepo{db: db} }

func (r *InviteRepo) Create(ctx context.Context, inv *domain.Invite) error {
	var used any
	if inv.UsedAt != nil {
		used = formatTime(*inv.UsedAt)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO invites (id, token, created_by, role, expires_at, used_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		inv.ID, inv.Token, inv.CreatedBy, string(inv.Role), formatTime(inv.ExpiresAt), used, formatTime(inv.CreatedAt))
	return err
}

func (r *InviteRepo) GetByToken(ctx context.Context, token string) (*domain.Invite, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, token, created_by, role, expires_at, used_at, created_at FROM invites WHERE token = ?`, token)
	return scanInvite(row)
}

func (r *InviteRepo) MarkUsed(ctx context.Context, id string, at time.Time) error {
	res, err := r.db.ExecContext(ctx, `UPDATE invites SET used_at=? WHERE id=?`, formatTime(at), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *InviteRepo) ListRecent(ctx context.Context, limit int) ([]domain.Invite, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, token, created_by, role, expires_at, used_at, created_at FROM invites ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Invite
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *inv)
	}
	return out, rows.Err()
}

func scanInvite(sc interface {
	Scan(dest ...any) error
}) (*domain.Invite, error) {
	var inv domain.Invite
	var used sql.NullString
	var exp, created string
	err := sc.Scan(&inv.ID, &inv.Token, &inv.CreatedBy, &inv.Role, &exp, &used, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	inv.ExpiresAt, err = parseTime(exp)
	if err != nil {
		return nil, fmt.Errorf("expires_at: %w", err)
	}
	if used.Valid {
		t, err := parseTime(used.String)
		if err != nil {
			return nil, fmt.Errorf("used_at: %w", err)
		}
		inv.UsedAt = &t
	}
	inv.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	return &inv, nil
}

type SettingsRepo struct{ db *sql.DB }

func NewSettingsRepo(db *sql.DB) *SettingsRepo { return &SettingsRepo{db: db} }

func (r *SettingsRepo) Get(ctx context.Context, key string) (string, bool, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (r *SettingsRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}
