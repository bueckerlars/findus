package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"findus/backend/internal/domain"
	"findus/backend/internal/repository"
)

const DefaultLabelBatchLimit = 400

type LabelPDFGenerator struct {
	QR           *QR
	MaxBatch     int64
	Reservations repository.ItemQRTokenReservationRepository
}

type LabelPDFInput struct {
	From   int64
	To     int64
	Policy domain.ItemIDPolicy
	Cols   int // 0 = default (2)
	Rows   int // 0 = auto-fit rows for A4
}

type LabelPDFOutput struct {
	Filename string
	PDF      []byte
}

func (g *LabelPDFGenerator) Generate(ctx context.Context, in LabelPDFInput) (LabelPDFOutput, error) {
	pol := in.Policy.Normalize()
	if pol.Kind != domain.ItemIDKindSequential {
		return LabelPDFOutput{}, fmt.Errorf("%w: kind", domain.ErrValidation)
	}
	if in.From < 1 || in.To < 1 || in.From > in.To {
		return LabelPDFOutput{}, fmt.Errorf("%w: invalid range", domain.ErrValidation)
	}
	limit := g.MaxBatch
	if limit < 1 {
		limit = DefaultLabelBatchLimit
	}
	count := in.To - in.From + 1
	if count > limit {
		return LabelPDFOutput{}, fmt.Errorf("%w: range too large (max %d)", domain.ErrValidation, limit)
	}

	tmpDir, err := os.MkdirTemp("", "findus-label-pdf-*")
	if err != nil {
		return LabelPDFOutput{}, err
	}
	defer os.RemoveAll(tmpDir)

	contents := make([]labelContent, 0, count)
	for seq := in.From; seq <= in.To; seq++ {
		itemID := domain.FormatSequentialID(pol.Prefix, pol.Width, seq)
		token, err := g.reserveOrLoadToken(ctx, itemID)
		if err != nil {
			return LabelPDFOutput{}, err
		}
		png, err := g.QR.PNG(token)
		if err != nil {
			return LabelPDFOutput{}, err
		}
		qrPath := filepath.Join(tmpDir, fmt.Sprintf("qr-%d.png", seq))
		if err := os.WriteFile(qrPath, png, 0o644); err != nil {
			return LabelPDFOutput{}, err
		}
		contents = append(contents, labelContent{QRPath: qrPath, PrimaryText: itemID})
	}

	layout := computeLabelLayout(in.Cols, in.Rows)
	pdfBytes, err := renderLabelsPDF(contents, layout)
	if err != nil {
		return LabelPDFOutput{}, err
	}
	return LabelPDFOutput{
		Filename: fmt.Sprintf("labels-%d-%d.pdf", in.From, in.To),
		PDF:      pdfBytes,
	}, nil
}

func (g *LabelPDFGenerator) reserveOrLoadToken(ctx context.Context, itemID string) (string, error) {
	if g.Reservations == nil {
		return "", fmt.Errorf("%w: qr reservations", domain.ErrValidation)
	}
	token, ok, err := g.Reservations.GetTokenByItemID(ctx, itemID)
	if err != nil {
		return "", err
	}
	if ok {
		return token, nil
	}
	for attempt := 0; attempt < 12; attempt++ {
		candidate := newID()
		if err := g.Reservations.Reserve(ctx, itemID, candidate); err == nil {
			return candidate, nil
		} else if !isSQLiteUniqueViolation(err) {
			return "", err
		}
		token, ok, err = g.Reservations.GetTokenByItemID(ctx, itemID)
		if err != nil {
			return "", err
		}
		if ok {
			return token, nil
		}
	}
	return "", fmt.Errorf("%w: qr token reserve", domain.ErrValidation)
}
