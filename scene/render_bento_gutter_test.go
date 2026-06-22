package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for R11.9 bento row-label gutter fit (D-089).

// TestBentoGutterWidth_FitsLabels: the gutter tracks the widest row label — 0 when
// unlabeled, the circular minimum for a short label, the label's natural width plus
// padding for a medium one, and the cap for a very long one.
func TestBentoGutterWidth_FitsLabels(t *testing.T) {
	theme := pptx.DefaultTheme()

	none := Bento{Columns: 2, Rows: []BentoRow{{Cells: []BentoCell{{Span: 1, Node: Prose{}}}}}}
	if g := bentoGutterWidthOf(theme, none); g != 0 {
		t.Errorf("unlabeled gutter = %d, want 0", g)
	}

	short := Bento{Columns: 2, Rows: []BentoRow{{Label: "A", Cells: []BentoCell{{Span: 1, Node: Prose{}}}}}}
	if g := bentoGutterWidthOf(theme, short); g != bentoGutterMinW {
		t.Errorf("short-label gutter = %d, want the minimum %d", g, bentoGutterMinW)
	}

	// A medium label sizes to naturalWidth + padding (between min and max).
	medium := Bento{Columns: 2, Rows: []BentoRow{{Label: "Control plane", Cells: []BentoCell{{Span: 1, Node: Prose{}}}}}}
	gMed := bentoGutterWidthOf(theme, medium)
	wantMed := naturalWidth(RichText{{Text: "Control plane", Style: RunStyle{TypeRole: pptx.TypeCaption}}}, theme) + 2*bentoGutterPadX
	if gMed != wantMed {
		t.Errorf("medium-label gutter = %d, want naturalWidth+2pad = %d", gMed, wantMed)
	}
	if gMed <= bentoGutterMinW || gMed >= bentoGutterMaxW {
		t.Errorf("medium-label gutter %d should be strictly between min %d and max %d", gMed, bentoGutterMinW, bentoGutterMaxW)
	}

	// A very long label caps at the maximum.
	long := Bento{Columns: 2, Rows: []BentoRow{{Label: "An extremely long row label that exceeds the cap", Cells: []BentoCell{{Span: 1, Node: Prose{}}}}}}
	if g := bentoGutterWidthOf(theme, long); g != bentoGutterMaxW {
		t.Errorf("over-long gutter = %d, want the cap %d", g, bentoGutterMaxW)
	}
}

// TestBentoGutter_GeometryUsesFit: the gutter the geometry reserves equals
// bentoGutterWidthOf, and a label's natural width fits inside it (up to the cap).
func TestBentoGutter_GeometryUsesFit(t *testing.T) {
	theme := pptx.DefaultTheme()
	gap := theme.ResolveSpace(pptx.SpaceMD)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(12), H: pptx.In(6)}
	v := Bento{Columns: 2, Rows: []BentoRow{
		{Label: "The core", Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
		{Label: "Control plane", Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
	}}
	gutterW, _, _, _ := bentoGeometry(box, v, gap, nil, theme)
	if want := bentoGutterWidthOf(theme, v); gutterW != want {
		t.Errorf("geometry gutter = %d, want bentoGutterWidthOf = %d", gutterW, want)
	}
	// The widest label fits inside the gutter (label natural width <= gutter, since
	// it is below the cap).
	widest := naturalWidth(RichText{{Text: "Control plane", Style: RunStyle{TypeRole: pptx.TypeCaption}}}, theme)
	if widest > gutterW {
		t.Errorf("widest label width %d exceeds the gutter %d (would clip)", widest, gutterW)
	}
}
