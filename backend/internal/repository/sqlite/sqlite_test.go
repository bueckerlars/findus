package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
)

func TestOpenDBRunsMigrations(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var n int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&n)
	require.NoError(t, err)
	require.Equal(t, 1, n)

	var tc int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM item_templates`).Scan(&tc)
	require.NoError(t, err)
	require.Equal(t, 6, tc)

	var lc int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM labels`).Scan(&lc)
	require.NoError(t, err)
	require.Equal(t, 6, lc)

	var dt string
	err = db.QueryRowContext(ctx, `SELECT default_template_type FROM labels WHERE id = 'lbl_servers'`).Scan(&dt)
	require.NoError(t, err)
	require.Equal(t, "servers", dt)

	var locCnt int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM locations`).Scan(&locCnt)
	require.NoError(t, err)
	require.GreaterOrEqual(t, locCnt, 1)

	for _, id := range []string{"documents", "servers", "iot"} {
		var fj string
		err = db.QueryRowContext(ctx, `SELECT fields_json FROM item_templates WHERE id = ?`, id).Scan(&fj)
		require.NoError(t, err)
		_, err = domain.ParseTemplateFieldsJSON([]byte(fj))
		require.NoError(t, err, id)
	}
}

func TestItemRepoSearchFTSMatchQuery(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at)
		VALUES ('it1', 'Clothing rack', '', ?, 'documents', '{}', '{}', NULL, 'tok1', ?, ?)`,
		locID, now, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO items_fts(items_fts) VALUES('rebuild')`)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	res, err := repo.Search(ctx, "Clothing", 50)
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "it1", res[0].ID)
}
