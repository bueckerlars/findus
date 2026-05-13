package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
)

func TestInventoryCreateItemUsesSequentialPolicy(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	settings := sqlite.NewSettingsRepo(db)
	items := sqlite.NewItemRepo(db)
	itemQRTokens := sqlite.NewItemQRTokenReservationRepo(db)
	locs := sqlite.NewLocationRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)

	pol := domain.ItemIDPolicy{Kind: domain.ItemIDKindSequential, Prefix: "t-", Width: 3, NextSeq: 1}
	raw, err := domain.ItemIDPolicyJSON(pol)
	require.NoError(t, err)
	require.NoError(t, settings.Set(ctx, domain.SettingItemIDPolicy, string(raw)))

	inv := &service.Inventory{
		Locations:    locs,
		Items:        items,
		ItemQRTokens: itemQRTokens,
		Templates:    templates,
		Settings:     settings,
	}
	it, err := inv.CreateItem(ctx, "one", "", locID, domain.TemplateType("standard"), []byte(`{}`), []byte(`{}`), nil)
	require.NoError(t, err)
	require.Equal(t, "t-_001", it.ID)

	it2, err := inv.CreateItem(ctx, "two", "", locID, domain.TemplateType("standard"), []byte(`{}`), []byte(`{}`), nil)
	require.NoError(t, err)
	require.Equal(t, "t-_002", it2.ID)
}

func TestInventorySetItemIDPolicyMigrates(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	settings := sqlite.NewSettingsRepo(db)
	items := sqlite.NewItemRepo(db)
	itemQRTokens := sqlite.NewItemQRTokenReservationRepo(db)
	locs := sqlite.NewLocationRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)

	inv := &service.Inventory{
		Locations:    locs,
		Items:        items,
		ItemQRTokens: itemQRTokens,
		Templates:    templates,
		Settings:     settings,
	}
	_, err = inv.CreateItem(ctx, "a", "", locID, domain.TemplateType("standard"), []byte(`{}`), []byte(`{}`), nil)
	require.NoError(t, err)

	want := domain.ItemIDPolicy{Kind: domain.ItemIDKindSequential, Prefix: "inv-", Width: 2, NextSeq: 1}
	err = inv.SetItemIDPolicy(ctx, dir, want)
	if err != nil {
		t.Fatalf("SetItemIDPolicy: %v", err)
	}

	got, err := items.GetByID(ctx, "inv-_01")
	require.NoError(t, err)
	require.Equal(t, "a", got.Name)

	pol, err := inv.GetItemIDPolicy(ctx)
	require.NoError(t, err)
	require.Equal(t, domain.ItemIDKindSequential, pol.Kind)
	require.Equal(t, int64(2), pol.NextSeq)
}

func TestInventorySetItemIDPolicyRemapsWhenIDsDoNotMatchPolicy(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	settings := sqlite.NewSettingsRepo(db)
	items := sqlite.NewItemRepo(db)
	itemQRTokens := sqlite.NewItemQRTokenReservationRepo(db)
	locs := sqlite.NewLocationRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)

	now := "2020-01-01T00:00:00Z"
	_, err = db.ExecContext(ctx, `
		INSERT INTO items (id, name, description, location_id, template_type, template_data, additional_data, photo_path, qr_token, created_at, updated_at)
		VALUES ('01ARZ3NDEKTSV4RRFFQ69G5FAV', 'legacy', '', ?, 'standard', '{}', '{}', NULL, 'qt_legacy', ?, ?)`,
		locID, now, now)
	require.NoError(t, err)

	raw, err := domain.ItemIDPolicyJSON(domain.ItemIDPolicy{
		Kind: domain.ItemIDKindSequential, Prefix: "item", Width: 4, NextSeq: 99,
	})
	require.NoError(t, err)
	require.NoError(t, settings.Set(ctx, domain.SettingItemIDPolicy, string(raw)))

	inv := &service.Inventory{
		Locations:    locs,
		Items:        items,
		ItemQRTokens: itemQRTokens,
		Templates:    templates,
		Settings:     settings,
	}
	want := domain.ItemIDPolicy{Kind: domain.ItemIDKindSequential, Prefix: "item", Width: 4, NextSeq: 1}
	require.NoError(t, inv.SetItemIDPolicy(ctx, dir, want))

	got, err := items.GetByID(ctx, "item_0001")
	require.NoError(t, err)
	require.Equal(t, "legacy", got.Name)
}

func TestInventoryCreateItemUsesReservedQRToken(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	defer db.Close()

	settings := sqlite.NewSettingsRepo(db)
	items := sqlite.NewItemRepo(db)
	itemQRTokens := sqlite.NewItemQRTokenReservationRepo(db)
	locs := sqlite.NewLocationRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)

	pol := domain.ItemIDPolicy{Kind: domain.ItemIDKindSequential, Prefix: "item", Width: 4, NextSeq: 1}
	raw, err := domain.ItemIDPolicyJSON(pol)
	require.NoError(t, err)
	require.NoError(t, settings.Set(ctx, domain.SettingItemIDPolicy, string(raw)))
	require.NoError(t, itemQRTokens.Reserve(ctx, "item_0001", "tok_reserved_item_1"))

	inv := &service.Inventory{
		Locations:    locs,
		Items:        items,
		ItemQRTokens: itemQRTokens,
		Templates:    templates,
		Settings:     settings,
	}
	it, err := inv.CreateItem(ctx, "one", "", locID, domain.TemplateType("standard"), []byte(`{}`), []byte(`{}`), nil)
	require.NoError(t, err)
	require.Equal(t, "item_0001", it.ID)
	require.Equal(t, "tok_reserved_item_1", it.QRToken)
}
