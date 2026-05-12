package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/repository/sqlite"
)

func TestItemRepoUpdateItemSearchDenormAfterInsert(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := "denorm_" + t.Name()
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at)
		VALUES (?, 'Hello', '', ?, 'standard', '{}', '{}', NULL, ?, ?, ?)`,
		id, locID, id+"tok", now, now)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	err = repo.UpdateItemSearchDenorm(ctx, id, "L1", "locpath", "body")
	require.NoError(t, err)
}
