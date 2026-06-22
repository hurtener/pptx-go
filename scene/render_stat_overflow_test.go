package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for R11.8 stat-value overflow guard (D-088).

// TestStatValueFit_RoleLadder: the value steps down the pinned role ladder
// (TypeDisplay → H1 → H2) as the box narrows, and falls back to a FontScale shrink
// at the H2 floor when even that wraps. AutoFit-off is always (TypeDisplay, 0).
func TestStatValueFit_RoleLadder(t *testing.T) {
	r := newTestRenderer(t)
	const value = "$4,000+"

	// AutoFit off: full display, no scale (byte-identical), regardless of width.
	if role, scale := r.statValueFit(false, value, pptx.In(1)); role != pptx.TypeDisplay || scale != 0 {
		t.Errorf("AutoFit-off = (%v, %v), want (TypeDisplay, 0)", role, scale)
	}

	// Wide box: fits at TypeDisplay.
	if role, scale := r.statValueFit(true, value, pptx.In(2.5)); role != pptx.TypeDisplay || scale != 0 {
		t.Errorf("wide box = (%v, %v), want (TypeDisplay, 0)", role, scale)
	}

	// Medium box: steps to H1.
	if role, scale := r.statValueFit(true, value, pptx.In(1.7)); role != pptx.TypeH1 || scale != 0 {
		t.Errorf("medium box = (%v, %v), want (TypeH1, 0)", role, scale)
	}

	// Narrow box: steps to the H2 floor.
	if role, scale := r.statValueFit(true, value, pptx.In(1.4)); role != pptx.TypeH2 || scale != 0 {
		t.Errorf("narrow box = (%v, %v), want (TypeH2, 0)", role, scale)
	}

	// Tiny box: H2 still wraps → floor + a shrink scale.
	role, scale := r.statValueFit(true, value, pptx.In(0.8))
	if role != pptx.TypeH2 {
		t.Errorf("tiny box role = %v, want TypeH2 (floor)", role)
	}
	if scale <= 0 || scale >= 1 {
		t.Errorf("tiny box scale = %v, want a shrink in (0,1)", scale)
	}
}

// TestStatValueFit_OneLine: at the chosen role (and scale), the value occupies a
// single line — the R11.8 guarantee — across a sweep of box widths.
func TestStatValueFit_OneLine(t *testing.T) {
	r := newTestRenderer(t)
	const value = "$4,000+"
	for _, w := range []pptx.EMU{pptx.In(2.5), pptx.In(1.7), pptx.In(1.4), pptx.In(1.0)} {
		role, scale := r.statValueFit(true, value, w)
		// Effective width after scale: a 0 scale means full role size.
		natW := naturalWidthAt(RichText{{Text: value}}, role, r.theme)
		eff := natW
		if scale > 0 {
			eff = pptx.EMU(float64(natW) * scale)
		}
		if eff > w {
			t.Errorf("width %d: value effective width %d (role %v scale %v) exceeds box — would wrap", w, eff, role, scale)
		}
	}
}
