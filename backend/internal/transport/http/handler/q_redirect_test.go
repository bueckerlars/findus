package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/config"
	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
	"findus/backend/internal/transport/http/handler"
)

func TestQRedirectResolvesReservedTokenAfterItemCreation(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	locs := sqlite.NewLocationRepo(db)
	items := sqlite.NewItemRepo(db)
	settings := sqlite.NewSettingsRepo(db)
	templates := sqlite.NewItemTemplateRepo(db)
	reservations := sqlite.NewItemQRTokenReservationRepo(db)

	inv := &service.Inventory{
		Locations:    locs,
		Items:        items,
		ItemQRTokens: reservations,
		Templates:    templates,
		Settings:     settings,
	}
	qr := &service.QR{BaseURL: "http://localhost:8080"}
	gen := service.LabelPDFGenerator{
		QR:           qr,
		MaxBatch:     10,
		Reservations: reservations,
	}
	pol, err := inv.GetItemIDPolicy(ctx)
	require.NoError(t, err)
	_, err = gen.Generate(ctx, service.LabelPDFInput{From: 1, To: 1, Policy: pol})
	require.NoError(t, err)

	token, ok, err := reservations.GetTokenByItemID(ctx, "item_0001")
	require.NoError(t, err)
	require.True(t, ok)

	srv := &handler.Server{
		Log:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		Config:       config.Config{DataDir: dir},
		Locs:         locs,
		Items:        items,
		ItemQRTokens: reservations,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /q/{token}", srv.QRedirect)

	rrBefore := httptest.NewRecorder()
	reqBefore := httptest.NewRequest(http.MethodGet, "/q/"+token, nil)
	mux.ServeHTTP(rrBefore, reqBefore)
	require.Equal(t, http.StatusNotFound, rrBefore.Code)

	var locID string
	err = db.QueryRowContext(ctx, `SELECT id FROM locations LIMIT 1`).Scan(&locID)
	require.NoError(t, err)
	created, err := inv.CreateItem(ctx, "reserved", "", locID, domain.TemplateType("standard"), []byte(`{}`), []byte(`{}`), nil)
	require.NoError(t, err)
	require.Equal(t, "item_0001", created.ID)
	require.Equal(t, token, created.QRToken)

	rrAfter := httptest.NewRecorder()
	reqAfter := httptest.NewRequest(http.MethodGet, "/q/"+token, nil)
	mux.ServeHTTP(rrAfter, reqAfter)
	require.Equal(t, http.StatusFound, rrAfter.Code)
	require.Equal(t, "/items/item_0001", rrAfter.Header().Get("Location"))
}
