package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/repository/sqlite"
)

func TestItemSearchRankingNameOverDescription(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	prefix := "ranktest_" + t.Name()
	q := prefix + "needle"

	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at, search_labels, search_location, search_body)
		VALUES (?, ?, '', ?, 'standard', '{}', '{}', NULL, ?, ?, ?, '', '', '')`,
		prefix+"desc", "zzz", locID, prefix+"tokd", now, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at, search_labels, search_location, search_body)
		VALUES (?, ?, ?, ?, 'standard', '{}', '{}', NULL, ?, ?, ?, '', '', '')`,
		prefix+"name", q, "", locID, prefix+"tokn", now, now)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	_, err = db.ExecContext(ctx, `UPDATE items SET description = ? WHERE id = ?`, q, prefix+"desc")
	require.NoError(t, err)

	res, err := repo.Search(ctx, prefix+"needle", 20)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(res), 2)
	require.Equal(t, prefix+"name", res[0].ID, "name hit should rank before description-only hit")
}

func TestItemSearchInfixTrigram(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := "infix_" + t.Name()
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at, search_labels, search_location, search_body)
		VALUES (?, 'xx-raretoken-zz', '', ?, 'standard', '{}', '{}', NULL, ?, ?, ?, '', '', '')`,
		id, locID, id+"tok", now, now)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	res, err := repo.Search(ctx, "raretoken", 20)
	require.NoError(t, err)
	var found bool
	for _, it := range res {
		if it.ID == id {
			found = true
			break
		}
	}
	require.True(t, found, "infix substring should match via trigram FTS")
}

func TestItemSearchTemplateDataValue(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	id := "tmpl_" + t.Name()
	td := `{"serial":"UNIQSERIAL999X"}`
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at, search_labels, search_location, search_body)
		VALUES (?, 'box', '', ?, 'servers', ?, '{}', NULL, ?, ?, ?, '', '', '')`,
		id, locID, td, id+"tok", now, now)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	res, err := repo.Search(ctx, "UNIQSERIAL999X", 20)
	require.NoError(t, err)
	var found bool
	for _, it := range res {
		if it.ID == id {
			found = true
			break
		}
	}
	require.True(t, found, "template_data should be indexed in FTS")
}

func TestItemSearchFuzzyRerankAmongHits(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	p := "fuzz_" + t.Name()
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at, search_labels, search_location, search_body)
		VALUES (?, 'gadgetaaa', '', ?, 'standard', '{}', '{}', NULL, ?, ?, ?, '', '', '')`,
		p+"a", locID, p+"toka", now, now)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at, search_labels, search_location, search_body)
		VALUES (?, 'gadgetz', '', ?, 'standard', '{}', '{}', NULL, ?, ?, ?, '', '', '')`,
		p+"z", locID, p+"tokz", now, now)
	require.NoError(t, err)

	repo := sqlite.NewItemRepo(db)
	res, err := repo.Search(ctx, "gadget", 20)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(res), 2)
	require.Equal(t, p+"z", res[0].ID, "closer name to query should rank first (fuzzy + bm25)")
}
