package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
)

func TestItemRepoMigrateItemPrimaryKeys_Sequential(t *testing.T) {
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
		VALUES ('it_ulid_a', 'A', '', ?, 'standard', '{}', '{}', 'images/it_ulid_a.webp', 'qt_a', ?, ?),
		       ('it_ulid_b', 'B', '', ?, 'standard', '{}', '{}', NULL, 'qt_b', ?, ?)`,
		locID, now, now, locID, now, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO item_labels (item_id, label_id) VALUES ('it_ulid_a', 'lbl_standard')`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO items_fts(items_fts) VALUES('rebuild')`)
	require.NoError(t, err)

	p1 := "images/doc-01.webp"
	p2 := (*string)(nil)
	repo := sqlite.NewItemRepo(db)
	err = repo.MigrateItemPrimaryKeys(ctx, []domain.ItemIDMigration{
		{OldID: "it_ulid_a", NewID: "doc-01", NewPhotoPath: &p1},
		{OldID: "it_ulid_b", NewID: "doc-02", NewPhotoPath: p2},
	})
	require.NoError(t, err)

	err = repo.UpdateItemSearchDenorm(ctx, "doc-01", "lab", "loc", "body")
	require.NoError(t, err, "UpdateItemSearchDenorm after PK migrate")

	var cnt int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM items WHERE id IN ('doc-01','doc-02')`).Scan(&cnt)
	require.NoError(t, err)
	require.Equal(t, 2, cnt)

	var lbl string
	err = db.QueryRowContext(ctx, `SELECT label_id FROM item_labels WHERE item_id = 'doc-01'`).Scan(&lbl)
	require.NoError(t, err)
	require.Equal(t, "lbl_standard", lbl)

	var photo sql.NullString
	err = db.QueryRowContext(ctx, `SELECT photo_path FROM items WHERE id = 'doc-01'`).Scan(&photo)
	require.NoError(t, err)
	require.True(t, photo.Valid)
	require.Equal(t, "images/doc-01.webp", photo.String)

	it, err := repo.GetByID(ctx, "doc-02")
	require.NoError(t, err)
	require.Nil(t, it.PhotoPath)
}

func TestItemRepoMigrateItemPrimaryKeys_TwoPhaseSwap(t *testing.T) {
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
		VALUES ('x1', 'one', '', ?, 'standard', '{}', '{}', NULL, 'qx1', ?, ?),
		       ('x2', 'two', '', ?, 'standard', '{}', '{}', NULL, 'qx2', ?, ?)`,
		locID, now, now, locID, now, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO item_labels (item_id, label_id) VALUES ('x1', 'lbl_standard'), ('x2', 'lbl_electronics')`)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	err = repo.MigrateItemPrimaryKeys(ctx, []domain.ItemIDMigration{
		{OldID: "x1", NewID: "x2"},
		{OldID: "x2", NewID: "x1"},
	})
	require.NoError(t, err)

	var n1, n2 string
	err = db.QueryRowContext(ctx, `SELECT name FROM items WHERE id = 'x1'`).Scan(&n1)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `SELECT name FROM items WHERE id = 'x2'`).Scan(&n2)
	require.NoError(t, err)
	require.Equal(t, "two", n1)
	require.Equal(t, "one", n2)

	var c1, c2 int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM item_labels WHERE item_id = 'x1' AND label_id = 'lbl_electronics'`).Scan(&c1)
	require.NoError(t, err)
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM item_labels WHERE item_id = 'x2' AND label_id = 'lbl_standard'`).Scan(&c2)
	require.NoError(t, err)
	require.Equal(t, 1, c1)
	require.Equal(t, 1, c2)
}
