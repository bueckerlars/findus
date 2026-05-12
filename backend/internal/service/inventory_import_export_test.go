package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
)

func newTestInventory(db *sql.DB) *service.Inventory {
	return &service.Inventory{
		Locations: sqlite.NewLocationRepo(db),
		Items:     sqlite.NewItemRepo(db),
		Labels:    sqlite.NewLabelRepo(db),
		Templates: sqlite.NewItemTemplateRepo(db),
		Settings:  sqlite.NewSettingsRepo(db),
	}
}

func runImportWithTx(t *testing.T, ctx context.Context, db *sql.DB, mainInv *service.Inventory, bundle *service.InventoryExportBundle) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	invTx := &service.Inventory{
		Locations: sqlite.NewLocationRepoConn(tx),
		Items:     sqlite.NewItemRepoConn(tx),
		Labels:    sqlite.NewLabelRepoConn(tx),
		Templates: sqlite.NewItemTemplateRepoConn(tx),
		Settings:  sqlite.NewSettingsRepoConn(tx),
	}
	_, err = invTx.ImportInventoryBundle(ctx, bundle)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.NoError(t, mainInv.PostInventoryImportSearchRefresh(ctx, bundle))
	require.NoError(t, mainInv.ReconcileSequentialNextSeqAfterImport(ctx))
}

func TestInventoryExportImportRoundtripJSON(t *testing.T) {
	ctx := context.Background()
	dir1 := t.TempDir()
	db1, err := sqlite.OpenDB(ctx, dir1)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db1.Close() })

	inv1 := newTestInventory(db1)
	parent := "loc_basement"
	_, err = inv1.CreateLocation(ctx, "ZetaImportChild", "import test child", &parent)
	require.NoError(t, err)

	bundle, err := inv1.BuildInventoryExportBundle(ctx)
	require.NoError(t, err)
	require.NoError(t, service.ValidateInventoryImportBundle(bundle))

	dir2 := t.TempDir()
	db2, err := sqlite.OpenDB(ctx, dir2)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db2.Close() })

	inv2 := newTestInventory(db2)
	runImportWithTx(t, ctx, db2, inv2, bundle)

	locs, err := inv2.Locations.ListAllExport(ctx)
	require.NoError(t, err)
	var found bool
	for _, l := range locs {
		if l.Name == "ZetaImportChild" {
			found = true
			break
		}
	}
	require.True(t, found, "imported location missing on target db")
}

func TestInventoryExportCSVZipRoundtrip(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	inv := newTestInventory(db)
	bundle, err := inv.BuildInventoryExportBundle(ctx)
	require.NoError(t, err)
	zb, err := service.EncodeInventoryCSVZIP(bundle)
	require.NoError(t, err)
	b2, err := service.DecodeInventoryCSVZIP(zb)
	require.NoError(t, err)
	require.Equal(t, bundle.ExportVersion, b2.ExportVersion)
	require.Equal(t, len(bundle.Items), len(b2.Items))
	require.Equal(t, len(bundle.Locations), len(b2.Locations))
}

func TestInventoryImportItemWithLabels(t *testing.T) {
	ctx := context.Background()
	dir1 := t.TempDir()
	db1, err := sqlite.OpenDB(ctx, dir1)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db1.Close() })
	inv1 := newTestInventory(db1)

	lbs, err := inv1.ListLabels(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, lbs)
	labelID := lbs[0].ID

	parent := "loc_basement"
	loc, err := inv1.CreateLocation(ctx, "LabelImportLoc", "x", &parent)
	require.NoError(t, err)

	it, err := inv1.CreateItem(ctx, "LabelledThing", "d", loc.ID, domain.TemplateType("standard"), json.RawMessage(`{}`), json.RawMessage(`{}`), []string{labelID})
	require.NoError(t, err)

	bundle, err := inv1.BuildInventoryExportBundle(ctx)
	require.NoError(t, err)

	dir2 := t.TempDir()
	db2, err := sqlite.OpenDB(ctx, dir2)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db2.Close() })
	inv2 := newTestInventory(db2)
	runImportWithTx(t, ctx, db2, inv2, bundle)

	got, err := inv2.Items.GetByID(ctx, it.ID)
	require.NoError(t, err)
	require.Equal(t, "LabelledThing", got.Name)
	lb2, err := inv2.Items.ListLabelsForItem(ctx, it.ID)
	require.NoError(t, err)
	require.NotEmpty(t, lb2)
}
