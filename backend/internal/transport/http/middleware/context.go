package middleware

import (
	"context"

	"findus/backend/internal/domain"
)

type ctxKey string

const userKey ctxKey = "findus_user"
const permsKey ctxKey = "findus_permissions"

func WithUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func User(ctx context.Context) (*domain.User, bool) {
	v := ctx.Value(userKey)
	if v == nil {
		return nil, false
	}
	u, ok := v.(*domain.User)
	return u, ok
}

// WithPermissions attaches effective permissions for the current user (non-admin only; admins ignore this).
func WithPermissions(ctx context.Context, perms []domain.Permission) context.Context {
	return context.WithValue(ctx, permsKey, perms)
}

// Permissions returns permissions loaded for the session (empty slice = none). Second return is false if unset.
func Permissions(ctx context.Context) ([]domain.Permission, bool) {
	v := ctx.Value(permsKey)
	if v == nil {
		return nil, false
	}
	p, ok := v.([]domain.Permission)
	return p, ok
}

// Can returns true if the user is admin or holds the given permission.
func Can(ctx context.Context, p domain.Permission) bool {
	u, ok := User(ctx)
	if !ok {
		return false
	}
	if u.Role.IsAdmin() {
		return true
	}
	list, ok := Permissions(ctx)
	if !ok {
		return false
	}
	for _, x := range list {
		if x == p {
			return true
		}
	}
	return false
}
