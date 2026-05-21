package service

import "testing"

func TestComputeLabelLayoutFullPage(t *testing.T) {
	layout := computeLabelLayout(3, 4, true)
	if layout.MarginX != 0 || layout.MarginY != 0 {
		t.Fatalf("margins = (%v, %v), want 0", layout.MarginX, layout.MarginY)
	}
	wantW := labelPageW / 3
	wantH := labelPageH / 4
	if layout.LabelW != wantW || layout.LabelH != wantH {
		t.Fatalf("label size = (%v, %v), want (%v, %v)", layout.LabelW, layout.LabelH, wantW, wantH)
	}
	if layout.LabelsPerPage != 12 {
		t.Fatalf("labels per page = %d, want 12", layout.LabelsPerPage)
	}
}

func TestComputeLabelLayoutWithMargins(t *testing.T) {
	layout := computeLabelLayout(2, 3, false)
	if layout.MarginX != labelMarginX || layout.MarginY != labelMarginY {
		t.Fatalf("margins = (%v, %v)", layout.MarginX, layout.MarginY)
	}
	wantW := labelUsableW / 2
	wantH := labelUsableH / 3
	if layout.LabelW != wantW || layout.LabelH != wantH {
		t.Fatalf("label size = (%v, %v), want (%v, %v)", layout.LabelW, layout.LabelH, wantW, wantH)
	}
}
