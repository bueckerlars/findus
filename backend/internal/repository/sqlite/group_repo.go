package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"findus/backend/internal/domain"
)

type GroupRepo struct{ db *sql.DB }

func NewGroupRepo(db *sql.DB) *GroupRepo { return &GroupRepo{db: db} }

func (r *GroupRepo) EffectivePermissions(ctx context.Context, userID string) ([]domain.Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT gp.permission
		FROM user_perm_groups ug
		JOIN perm_group_permissions gp ON gp.group_id = ug.group_id
		WHERE ug.user_id = ?
		ORDER BY gp.permission`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Permission
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		if p, ok := domain.ParsePermission(s); ok {
			out = append(out, p)
		}
	}
	return out, rows.Err()
}

func (r *GroupRepo) ListGroupsForUser(ctx context.Context, userID string) ([]domain.PermissionGroup, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, g.created_at, g.updated_at
		FROM perm_groups g
		INNER JOIN user_perm_groups ug ON ug.group_id = g.id
		WHERE ug.user_id = ?
		ORDER BY g.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PermissionGroup
	for rows.Next() {
		g, err := scanPermissionGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, rows.Err()
}

func (r *GroupRepo) ListGroups(ctx context.Context) ([]domain.PermissionGroup, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, created_at, updated_at FROM perm_groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PermissionGroup
	for rows.Next() {
		g, err := scanPermissionGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *g)
	}
	return out, rows.Err()
}

func (r *GroupRepo) ListMemberCounts(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT group_id, COUNT(1) FROM user_perm_groups GROUP BY group_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]int64)
	for rows.Next() {
		var gid string
		var n int64
		if err := rows.Scan(&gid, &n); err != nil {
			return nil, err
		}
		out[gid] = n
	}
	return out, rows.Err()
}

func (r *GroupRepo) GetGroup(ctx context.Context, id string) (*domain.PermissionGroup, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, created_at, updated_at FROM perm_groups WHERE id = ?`, id)
	return scanPermissionGroup(row)
}

func (r *GroupRepo) GetGroupPermissions(ctx context.Context, groupID string) ([]domain.Permission, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT permission FROM perm_group_permissions WHERE group_id = ? ORDER BY permission`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Permission
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		if p, ok := domain.ParsePermission(s); ok {
			out = append(out, p)
		}
	}
	return out, rows.Err()
}

func (r *GroupRepo) ListUserGroupIDs(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT user_id, group_id FROM user_perm_groups`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]string)
	for rows.Next() {
		var uid, gid string
		if err := rows.Scan(&uid, &gid); err != nil {
			return nil, err
		}
		out[uid] = append(out[uid], gid)
	}
	return out, rows.Err()
}

func (r *GroupRepo) CreateGroup(ctx context.Context, g *domain.PermissionGroup) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO perm_groups (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		g.ID, g.Name, formatTime(g.CreatedAt), formatTime(g.UpdatedAt))
	return err
}

func (r *GroupRepo) UpdateGroup(ctx context.Context, g *domain.PermissionGroup) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE perm_groups SET name = ?, updated_at = ? WHERE id = ?`,
		g.Name, formatTime(g.UpdatedAt), g.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GroupRepo) DeleteGroup(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM perm_groups WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GroupRepo) ReplaceGroupPermissions(ctx context.Context, groupID string, perms []domain.Permission) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM perm_group_permissions WHERE group_id = ?`, groupID); err != nil {
		return err
	}
	for _, p := range perms {
		if !domain.ValidPermission(p) {
			return fmt.Errorf("%w: unknown permission %q", domain.ErrValidation, p)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO perm_group_permissions (group_id, permission) VALUES (?, ?)`,
			groupID, string(p)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *GroupRepo) ReplaceUserGroups(ctx context.Context, userID string, groupIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_perm_groups WHERE user_id = ?`, userID); err != nil {
		return err
	}
	for _, gid := range groupIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO user_perm_groups (user_id, group_id) VALUES (?, ?)`, userID, gid); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func scanPermissionGroup(sc interface {
	Scan(dest ...any) error
}) (*domain.PermissionGroup, error) {
	var g domain.PermissionGroup
	var created, updated string
	err := sc.Scan(&g.ID, &g.Name, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	g.CreatedAt, err = parseTime(created)
	if err != nil {
		return nil, fmt.Errorf("created_at: %w", err)
	}
	g.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return nil, fmt.Errorf("updated_at: %w", err)
	}
	return &g, nil
}
