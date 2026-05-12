package middleware

import (
	"context"

	"findus/backend/internal/domain"
)

type ctxKey string

const userKey ctxKey = "findus_user"

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
