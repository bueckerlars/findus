package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository/sqlite"
	"findus/backend/internal/service"
)

func TestLabelPDFGeneratorGenerate(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gen := service.LabelPDFGenerator{
		QR:           &service.QR{BaseURL: "http://localhost:8080"},
		MaxBatch:     50,
		Reservations: sqlite.NewItemQRTokenReservationRepo(db),
	}
	out, err := gen.Generate(ctx, service.LabelPDFInput{
		From: 1,
		To:   6,
		Policy: domain.ItemIDPolicy{
			Kind:    domain.ItemIDKindSequential,
			Prefix:  "item",
			Width:   4,
			NextSeq: 1,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "labels-1-6.pdf", out.Filename)
	require.NotEmpty(t, out.PDF)
	require.Contains(t, string(out.PDF[:12]), "%PDF")
}

func TestLabelPDFGeneratorValidateRange(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	gen := service.LabelPDFGenerator{
		QR:           &service.QR{BaseURL: "http://localhost:8080"},
		MaxBatch:     2,
		Reservations: sqlite.NewItemQRTokenReservationRepo(db),
	}
	_, err = gen.Generate(ctx, service.LabelPDFInput{
		From: 10,
		To:   9,
		Policy: domain.ItemIDPolicy{
			Kind:   domain.ItemIDKindSequential,
			Prefix: "item",
			Width:  4,
		},
	})
	require.Error(t, err)

	_, err = gen.Generate(ctx, service.LabelPDFInput{
		From: 1,
		To:   4,
		Policy: domain.ItemIDPolicy{
			Kind:   domain.ItemIDKindSequential,
			Prefix: "item",
			Width:  4,
		},
	})
	require.Error(t, err)
}

func TestLabelPDFGeneratorReusesReservedToken(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	db, err := sqlite.OpenDB(ctx, dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := sqlite.NewItemQRTokenReservationRepo(db)
	require.NoError(t, repo.Reserve(ctx, "item_0001", "tok_fixed_1"))

	gen := service.LabelPDFGenerator{
		QR:           &service.QR{BaseURL: "http://localhost:8080"},
		MaxBatch:     50,
		Reservations: repo,
	}
	_, err = gen.Generate(ctx, service.LabelPDFInput{
		From: 1,
		To:   1,
		Policy: domain.ItemIDPolicy{
			Kind:    domain.ItemIDKindSequential,
			Prefix:  "item",
			Width:   4,
			NextSeq: 1,
		},
	})
	require.NoError(t, err)

	token, ok, err := repo.GetTokenByItemID(ctx, "item_0001")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "tok_fixed_1", token)
}
