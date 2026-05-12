package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"findus/backend/internal/domain"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM users`).Scan(&n)
	return n, err
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	theme := domain.NormalizeUITheme(u.UITheme)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, role, is_active, created_at, updated_at, avatar_path, ui_theme)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.Email, u.PasswordHash, string(u.Role), boolToInt(u.IsActive), formatTime(u.CreatedAt), formatTime(u.UpdatedAt), nullString(u.AvatarPath), theme)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role, is_active, created_at, updated_at, avatar_path, ui_theme FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role, is_active, created_at, updated_at, avatar_path, ui_theme FROM users WHERE username = ?`, username)
	return scanUser(row)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, role, is_active, created_at, updated_at, avatar_path, ui_theme FROM users WHERE email = ?`, email)
	return scanUser(row)
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, email, password_hash, role, is_active, created_at, updated_at, avatar_path, ui_theme FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *u)
	}
	return out, rows.Err()
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET username=?, email=?, password_hash=?, role=?, is_active=?, updated_at=?, avatar_path=?, ui_theme=? WHERE id=?`,
		u.Username, u.Email, u.PasswordHash, string(u.Role), boolToInt(u.IsActive), formatTime(u.UpdatedAt), nullString(u.AvatarPath), domain.NormalizeUITheme(u.UITheme), u.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func nullString(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

func scanUser(sc interface {
	Scan(dest ...any) error
}) (*domain.User, error) {
	var u domain.User
	var isActive int
	var created, updated string
	var avatar sql.NullString
	var uiTheme string
	err := sc.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &isActive, &created, &updated, &avatar, &uiTheme)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	u.IsActive = isActive != 0
	u.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	u.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return nil, fmt.Errorf("updated_at: %w", err)
	}
	if avatar.Valid && avatar.String != "" {
		s := avatar.String
		u.AvatarPath = &s
	}
	u.UITheme = domain.NormalizeUITheme(uiTheme)
	return &u, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
