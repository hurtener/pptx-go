package scene

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Estimate/actual parity (Phase 48, R10.10). White-box tests: the Card chrome
// estimate grows with a wrapped multi-line header (single-line byte-identical),
// and the Bento estimate measures each cell at its actual span width (span-1
// byte-identical, wide-span no longer over-counts).

// TestPreferredHeight_WrappedCardGrows: a card with a header that wraps to 2+
// lines in a narrow column has a larger preferredHeight than the same card with a
// short (single-line) header — the slot accounts for the wrapped header.
func TestPreferredHeight_WrappedCardGrows(t *testing.T) {
	theme := pptx.DefaultTheme()
	avail := pptx.In(2.5) // narrow card → a long header wraps
	long := Card{Header: strings.Repeat("Platform White Label ", 4)}
	short := Card{Header: "ARR"}

	gotLong := preferredHeight(long, avail, theme)
	gotShort := preferredHeight(short, avail, theme)
	if gotLong <= gotShort {
		t.Errorf("wrapped-header card (%d) should be taller than single-line (%d)", gotLong, gotShort)
	}
	// The growth equals the wrapped increment (extra title lines × cardTitleRowH).
	c := cardChrome{header: long.Header, size: long.Size, layout: long.Layout, paddingScale: long.PaddingScale}
	wantExtra := cardHeaderExtraHeight(theme, avail, c)
	if wantExtra <= 0 {
		t.Fatal("test setup: header did not wrap at this width")
	}
	if gotLong-gotShort != wantExtra {
		t.Errorf("card grew by %d, want the wrapped increment %d", gotLong-gotShort, wantExtra)
	}
}

// TestPreferredHeight_SingleLineCardUnchanged: a single-line-header card's
// estimate is exactly the pre-R10.10 formula (cardChromeEst + body + estGap),
// i.e. the wrapped increment is 0 (byte-identical).
func TestPreferredHeight_SingleLineCardUnchanged(t *testing.T) {
	theme := pptx.DefaultTheme()
	avail := pptx.In(4)
	card := Card{Header: "Revenue", Body: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "one line"}}}}}}

	got := preferredHeight(card, avail, theme)
	want := cardChromeEst + nodesHeight(card.Body, avail-2*cardBodyInsetEst, theme) + estGap
	if got != want {
		t.Errorf("single-line card estimate = %d, want the unchanged baseline %d", got, want)
	}
}

// TestPreferredHeight_BentoSpanWidth: a wide-span bento cell is measured at its
// span width, so a bento whose dense cell spans 2 columns is no taller than the
// same content measured at the (wider) span width — and a span-1 bento is
// byte-identical to a hand-computed unit-width estimate.
func TestPreferredHeight_BentoSpanWidth(t *testing.T) {
	theme := pptx.DefaultTheme()
	avail := pptx.In(9)
	long := Prose{Paragraphs: []RichText{{{Text: strings.Repeat("dense capability text ", 6)}}}}

	// A 3-column bento with the long prose in a span-2 cell vs the same in a span-1
	// cell: the span-2 cell renders wider, wraps less, so its estimate is <= span-1.
	span2 := Bento{Columns: 3, Rows: []BentoRow{{Cells: []BentoCell{{Span: 2, Node: long}, {Span: 1, Node: Prose{}}}}}}
	span1 := Bento{Columns: 3, Rows: []BentoRow{{Cells: []BentoCell{{Span: 1, Node: long}, {Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}}}}

	hSpan2 := preferredHeight(span2, avail, theme)
	hSpan1 := preferredHeight(span1, avail, theme)
	if hSpan2 > hSpan1 {
		t.Errorf("wide-span bento (%d) should be no taller than unit-width (%d) — span width must widen the cell", hSpan2, hSpan1)
	}
	if hSpan2 == hSpan1 {
		t.Errorf("test setup: span-2 and span-1 estimates equal (%d) — the long prose did not wrap differently", hSpan2)
	}
}

// TestPreferredHeight_BentoSpanOneByteIdentical: a bento with only span-1 cells is
// byte-identical to the unit-width estimate (the span fix only affects span>1).
func TestPreferredHeight_BentoSpanOneByteIdentical(t *testing.T) {
	theme := pptx.DefaultTheme()
	avail := pptx.In(9)
	cell := Prose{Paragraphs: []RichText{{{Text: strings.Repeat("text ", 8)}}}}
	b := Bento{Columns: 3, Rows: []BentoRow{
		{Cells: []BentoCell{{Span: 1, Node: cell}, {Span: 1, Node: cell}, {Span: 1, Node: cell}}},
		{Cells: []BentoCell{{Span: 1, Node: cell}}},
	}}
	// Recompute the unit-width estimate by hand (span-1 → spanW == unitW).
	cols := pptx.EMU(3)
	unitW := (avail - estGap*(cols-1)) / cols
	var maxCell pptx.EMU
	for _, row := range b.Rows {
		for _, c := range row.Cells {
			if h := preferredHeight(c.Node, unitW, theme); h > maxCell {
				maxCell = h
			}
		}
	}
	nRows := pptx.EMU(len(b.Rows))
	want := nRows*maxCell + estGap*(nRows-1)
	if got := preferredHeight(b, avail, theme); got != want {
		t.Errorf("span-1 bento estimate = %d, want the unit-width %d", got, want)
	}
}
