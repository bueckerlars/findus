package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
)

func TestAuthUpdateProfile(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	users := sqlite.NewUserRepo(db)
	settings := sqlite.NewSettingsRepo(db)
	invites := sqlite.NewInviteRepo(db)
	auth := &service.Auth{Users: users, Settings: settings, Invites: invites}

	u, err := auth.Register(ctx, "alice", "alice@example.com", "password1234", "")
	require.NoError(t, err)

	out, err := auth.UpdateProfile(ctx, u.ID, "alice2", "alice2@example.com", "password1234", "newpassword12")
	require.NoError(t, err)
	require.Equal(t, "alice2", out.Username)
	require.Equal(t, "alice2@example.com", out.Email)

	_, err = auth.UpdateProfile(ctx, u.ID, "alice2", "alice2@example.com", "wrongpassword", "")
	require.ErrorIs(t, err, domain.ErrInvalidCurrentPassword)
}
