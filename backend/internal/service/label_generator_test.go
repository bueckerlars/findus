package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/domain"
	"findus/backend/internal/service"
)

func TestLabelPDFGeneratorGenerate(t *testing.T) {
	gen := service.LabelPDFGenerator{
		QR:       &service.QR{BaseURL: "http://localhost:8080"},
		MaxBatch: 50,
	}
	out, err := gen.Generate(context.Background(), service.LabelPDFInput{
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
	gen := service.LabelPDFGenerator{
		QR:       &service.QR{BaseURL: "http://localhost:8080"},
		MaxBatch: 2,
	}
	_, err := gen.Generate(context.Background(), service.LabelPDFInput{
		From: 10,
		To:   9,
		Policy: domain.ItemIDPolicy{
			Kind:   domain.ItemIDKindSequential,
			Prefix: "item",
			Width:  4,
		},
	})
	require.Error(t, err)

	_, err = gen.Generate(context.Background(), service.LabelPDFInput{
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
