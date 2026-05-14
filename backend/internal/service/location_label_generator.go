package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"findus/backend/internal/domain"
)

const DefaultLocationLabelBatchLimit = 400

// LocationLabelRepository is the subset of the location repo used by the generator.
type LocationLabelRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Location, error)
}

type LocationLabelPDFGenerator struct {
	QR   *QR
	Locs LocationLabelRepository
}

type LocationLabelPDFInput struct {
	LocationIDs []string
	Cols        int // 0 = default (2)
	Rows        int // 0 = auto-fit for A4
}

type LocationLabelPDFOutput struct {
	Filename string
	PDF      []byte
}

func (g *LocationLabelPDFGenerator) Generate(ctx context.Context, in LocationLabelPDFInput) (LocationLabelPDFOutput, error) {
	if len(in.LocationIDs) == 0 {
		return LocationLabelPDFOutput{}, fmt.Errorf("%w: no locations selected", domain.ErrValidation)
	}
	if len(in.LocationIDs) > DefaultLocationLabelBatchLimit {
		return LocationLabelPDFOutput{}, fmt.Errorf("%w: too many locations (max %d)", domain.ErrValidation, DefaultLocationLabelBatchLimit)
	}

	tmpDir, err := os.MkdirTemp("", "findus-location-label-pdf-*")
	if err != nil {
		return LocationLabelPDFOutput{}, err
	}
	defer os.RemoveAll(tmpDir)

	contents := make([]labelContent, 0, len(in.LocationIDs))
	for i, id := range in.LocationIDs {
		loc, err := g.Locs.GetByID(ctx, id)
		if err != nil {
			return LocationLabelPDFOutput{}, fmt.Errorf("location %q: %w", id, err)
		}
		png, err := g.QR.PNG(loc.QRToken)
		if err != nil {
			return LocationLabelPDFOutput{}, err
		}
		qrPath := filepath.Join(tmpDir, fmt.Sprintf("qr-%d.png", i))
		if err := os.WriteFile(qrPath, png, 0o644); err != nil {
			return LocationLabelPDFOutput{}, err
		}
		contents = append(contents, labelContent{QRPath: qrPath, PrimaryText: loc.Name})
	}

	layout := computeLabelLayout(in.Cols, in.Rows)
	pdfBytes, err := renderLabelsPDF(contents, layout)
	if err != nil {
		return LocationLabelPDFOutput{}, err
	}
	return LocationLabelPDFOutput{
		Filename: "location-labels.pdf",
		PDF:      pdfBytes,
	}, nil
}
