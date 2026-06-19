package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestNaturalWidth_Deterministic verifies that naturalWidth returns the same
// value on repeated calls with identical inputs (pure-function contract).
func TestNaturalWidth_Deterministic(t *testing.T) {
	theme := pptx.DefaultTheme()
	rt := RichText{{Text: "Hello world", Style: RunStyle{TypeRole: pptx.TypeH2}}}
	a := naturalWidth(rt, theme)
	b := naturalWidth(rt, theme)
	if a != b {
		t.Fatalf("naturalWidth not deterministic: got %d then %d", a, b)
	}
	if a == 0 {
		t.Fatal("naturalWidth returned 0 for non-empty text")
	}
}

// TestNaturalWidth_Monotonic verifies that a longer text produces a wider
// estimate (monotonicity in text length for same TypeRole).
func TestNaturalWidth_Monotonic(t *testing.T) {
	theme := pptx.DefaultTheme()
	role := RunStyle{TypeRole: pptx.TypeBody}
	short := RichText{{Text: "Hi", Style: role}}
	long := RichText{{Text: "Hello world this is a longer text string", Style: role}}
	ws := naturalWidth(short, theme)
	wl := naturalWidth(long, theme)
	if ws >= wl {
		t.Errorf("naturalWidth not monotonic: short=%d >= long=%d", ws, wl)
	}
}

// TestNaturalWidth_Empty verifies that nil or empty text yields 0 width.
func TestNaturalWidth_Empty(t *testing.T) {
	theme := pptx.DefaultTheme()
	if w := naturalWidth(nil, theme); w != 0 {
		t.Errorf("nil RichText: want 0, got %d", w)
	}
	if w := naturalWidth(RichText{{Text: ""}}, theme); w != 0 {
		t.Errorf("empty text run: want 0, got %d", w)
	}
}

// TestNaturalWidth_ZeroTypeRoleIsTypeDisplay verifies that a zero TypeRole
// (the Go zero value = TypeDisplay = 0) resolves to the TypeDisplay font spec,
// not to a separate "unset" path. This is the correct semantic for naturalWidth;
// callers that know the base role should use naturalWidthAt instead.
func TestNaturalWidth_ZeroTypeRoleIsTypeDisplay(t *testing.T) {
	theme := pptx.DefaultTheme()
	zero := RichText{{Text: "abc"}}                                                 // TypeRole = 0 = TypeDisplay
	display := RichText{{Text: "abc", Style: RunStyle{TypeRole: pptx.TypeDisplay}}} // explicit TypeDisplay = 0
	if naturalWidth(zero, theme) != naturalWidth(display, theme) {
		t.Error("zero TypeRole should resolve identically to TypeDisplay")
	}
}

// TestNaturalWidth_LargerFontWider verifies that a larger TypeRole yields a
// wider estimate for the same text (monotonic in font size).
func TestNaturalWidth_LargerFontWider(t *testing.T) {
	theme := pptx.DefaultTheme()
	text := "Same text"
	small := RichText{{Text: text, Style: RunStyle{TypeRole: pptx.TypeCaption}}} // 10pt
	large := RichText{{Text: text, Style: RunStyle{TypeRole: pptx.TypeDisplay}}} // 40pt
	if naturalWidth(small, theme) >= naturalWidth(large, theme) {
		t.Error("larger font should produce wider naturalWidth estimate")
	}
}

// TestNaturalWidthAt_BaseRole verifies that naturalWidthAt uses the given
// base role for runs with zero TypeRole, producing a different (expected)
// width than the raw naturalWidth call.
func TestNaturalWidthAt_BaseRole(t *testing.T) {
	theme := pptx.DefaultTheme()
	// A run with zero TypeRole: naturalWidth uses TypeDisplay (40pt);
	// naturalWidthAt with TypeBody (14pt) should give a smaller result.
	rt := RichText{{Text: "Section"}}
	wDisplay := naturalWidth(rt, theme)               // TypeDisplay = 40pt
	wBody := naturalWidthAt(rt, pptx.TypeBody, theme) // TypeBody = 14pt
	if wBody >= wDisplay {
		t.Errorf("naturalWidthAt(TypeBody) (%d) should be < naturalWidth TypeDisplay (%d)", wBody, wDisplay)
	}
}

// TestNaturalWidthAt_EmptyRichText verifies graceful handling of nil input.
func TestNaturalWidthAt_EmptyRichText(t *testing.T) {
	theme := pptx.DefaultTheme()
	if w := naturalWidthAt(nil, pptx.TypeBody, theme); w != 0 {
		t.Errorf("naturalWidthAt(nil): want 0, got %d", w)
	}
}

// TestNodeNaturalWidth_HeadingUsesHeadingRole verifies that nodeNaturalWidth
// for a Heading uses the heading role (28pt for level-2), not the default
// TypeDisplay that a plain RichText zero-role run would resolve to.
func TestNodeNaturalWidth_HeadingUsesHeadingRole(t *testing.T) {
	theme := pptx.DefaultTheme()
	text := "Section"
	h := Heading{Text: RichText{{Text: text}}, Level: 2}
	nw := nodeNaturalWidth(h, theme)
	// The Heading level-2 font is 28pt; TypeDisplay is 40pt.
	// nodeNaturalWidth should use 28pt, so it must be < TypeDisplay result.
	displayW := naturalWidth(RichText{{Text: text}}, theme) // TypeDisplay = 40pt
	if nw >= displayW {
		t.Errorf("heading natural width %d should be < TypeDisplay width %d (heading uses 28pt, not 40pt)", nw, displayW)
	}
	if nw == 0 {
		t.Error("heading natural width should not be 0")
	}
}

// TestNodeNaturalWidth_ContainersReturnZero verifies that container and
// visual nodes return 0 (they always keep full-width boxes).
func TestNodeNaturalWidth_ContainersReturnZero(t *testing.T) {
	theme := pptx.DefaultTheme()
	containers := []SlideNode{
		Grid{Columns: 2, Cells: []SlideNode{Divider{}, Divider{}}},
		TwoColumn{Left: []SlideNode{Divider{}}, Right: []SlideNode{Divider{}}},
		Card{Header: "x", Body: []SlideNode{Divider{}}},
		Callout{Kind: CalloutNote, Body: RichText{{Text: "x"}}},
		Divider{},
		Arrow{Direction: ArrowRight},
	}
	for _, n := range containers {
		if w := nodeNaturalWidth(n, theme); w != 0 {
			t.Errorf("%T: expected nodeNaturalWidth=0, got %d", n, w)
		}
	}
}
