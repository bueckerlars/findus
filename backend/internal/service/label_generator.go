package service

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"

	"findus/backend/internal/domain"
)

const DefaultLabelBatchLimit = 400

type LabelPDFGenerator struct {
	QR       *QR
	MaxBatch int64
}

type LabelPDFInput struct {
	From   int64
	To     int64
	Policy domain.ItemIDPolicy
}

type LabelPDFOutput struct {
	Filename string
	PDF      []byte
}

func (g *LabelPDFGenerator) Generate(ctx context.Context, in LabelPDFInput) (LabelPDFOutput, error) {
	_ = ctx
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

	type labelRow struct {
		ItemID string
		QRPath string
	}
	rows := make([]labelRow, 0, count)
	for seq := in.From; seq <= in.To; seq++ {
		itemID := domain.FormatSequentialID(pol.Prefix, pol.Width, seq)
		token := deterministicLabelToken(itemID)
		png, err := g.QR.PNG(token)
		if err != nil {
			return LabelPDFOutput{}, err
		}
		qrPath := filepath.Join(tmpDir, fmt.Sprintf("qr-%d.png", seq))
		if err := os.WriteFile(qrPath, png, 0o644); err != nil {
			return LabelPDFOutput{}, err
		}
		rows = append(rows, labelRow{ItemID: itemID, QRPath: qrPath})
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(false, 10)
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(80, 80, 80)
	pdf.SetTextColor(20, 20, 20)

	const (
		labelW      = 95.0
		labelH      = 38.0
		leftRatio   = 0.46
		topMargin   = 10.0
		leftMargin  = 10.0
		qrSize      = 24.0
		qrTopPad    = 4.0
		lineRight   = 6.0
		lineTopPad  = 8.0
		lineSpacing = 8.0
	)
	usableHeight := 297.0 - 20.0
	rowsPerPage := int(usableHeight / labelH)
	labelsPerPage := rowsPerPage * 2

	for i, row := range rows {
		if i%labelsPerPage == 0 {
			pdf.AddPage()
		}
		pageIdx := i % labelsPerPage
		col := pageIdx % 2
		rw := pageIdx / 2
		x := leftMargin + float64(col)*labelW
		y := topMargin + float64(rw)*labelH
		leftW := labelW * leftRatio

		pdf.Rect(x, y, labelW, labelH, "")
		pdf.Line(x+leftW, y, x+leftW, y+labelH)

		qrX := x + (leftW-qrSize)/2
		qrY := y + qrTopPad
		pdf.ImageOptions(row.QRPath, qrX, qrY, qrSize, qrSize, false, gofpdf.ImageOptions{ImageType: "png"}, 0, "")

		pdf.SetFont("Helvetica", "", 10)
		pdf.SetXY(x+2, y+qrTopPad+qrSize+2)
		pdf.CellFormat(leftW-4, 5, row.ItemID, "", 0, "C", false, 0, "")

		lineStartX := x + leftW + 5
		lineEndX := x + labelW - lineRight
		for li := 0; li < 3; li++ {
			ly := y + lineTopPad + float64(li)*lineSpacing
			pdf.Line(lineStartX, ly, lineEndX, ly)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return LabelPDFOutput{}, err
	}
	return LabelPDFOutput{
		Filename: fmt.Sprintf("labels-%d-%d.pdf", in.From, in.To),
		PDF:      buf.Bytes(),
	}, nil
}

func deterministicLabelToken(itemID string) string {
	sum := sha1.Sum([]byte("findus-label:" + itemID))
	return "lbl_" + hex.EncodeToString(sum[:12])
}
