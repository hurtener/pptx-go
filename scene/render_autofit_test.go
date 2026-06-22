package scene

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Display text shrink-to-fit (Phase 43, R10.5). White-box tests for the pure
// scale function (fitScale) and the per-node display-run helper (displayRunScale).

// TestFitScale_FitsReturnsZero: text that already fits its box gets no scaling.
func TestFitScale_FitsReturnsZero(t *testing.T) {
	if got := fitScale(pptx.In(2), pptx.In(3)); got != 0 {
		t.Errorf("fitting text: fitScale = %v, want 0 (no scaling)", got)
	}
	if got := fitScale(pptx.In(3), pptx.In(3)); got != 0 {
		t.Errorf("exact-fit text: fitScale = %v, want 0", got)
	}
	if got := fitScale(0, pptx.In(3)); got != 0 {
		t.Errorf("unknown width: fitScale = %v, want 0", got)
	}
}

// TestFitScale_OverflowFitsAtOrAboveFloor: an over-wide run scales to a quantized
// factor in (0,1) that makes its width fit the box, never below the pinned floor.
func TestFitScale_OverflowFitsAtOrAboveFloor(t *testing.T) {
	natW, boxW := pptx.In(4), pptx.In(3)
	s := fitScale(natW, boxW)
	if s <= 0 || s >= 1 {
		t.Fatalf("overflow: fitScale = %v, want a fraction in (0,1)", s)
	}
	if s < 0.60 {
		t.Errorf("scale %v below the 0.60 floor", s)
	}
	// The scaled width must fit the box (the acceptance: estimated width <= boxW).
	if scaled := pptx.EMU(float64(natW) * s); scaled > boxW {
		t.Errorf("scaled width %d still exceeds box %d (scale %v)", scaled, boxW, s)
	}
	// Quantized to the 0.025 step.
	if bp := int(s*10000 + 0.5); bp%autofitStepBP != 0 {
		t.Errorf("scale %v (%d bp) not quantized to the %d-bp step", s, bp, autofitStepBP)
	}
}

// TestFitScale_FloorCaps: an extreme overflow is capped at the 0.60 ratio floor.
func TestFitScale_FloorCaps(t *testing.T) {
	if s := fitScale(pptx.In(10), pptx.In(3)); s != 0.60 {
		t.Errorf("extreme overflow: fitScale = %v, want the 0.60 floor", s)
	}
}

// TestFitScale_Deterministic: identical inputs always return the same scale.
func TestFitScale_Deterministic(t *testing.T) {
	a := fitScale(pptx.In(7), pptx.In(4))
	b := fitScale(pptx.In(7), pptx.In(4))
	if a != b {
		t.Errorf("non-deterministic: %v vs %v", a, b)
	}
}

// TestDisplayRunScale_NodeWiring: the per-node helper shrinks an over-wide display
// value, leaves a fitting value unscaled, and is off when AutoFit is false.
func TestDisplayRunScale_NodeWiring(t *testing.T) {
	r := newTestRenderer(t)
	long := strings.Repeat("8", 60) // wide at TypeDisplay
	boxW := pptx.In(3)

	on := r.displayRunScale(true, long, pptx.TypeDisplay, boxW)
	if on <= 0 || on >= 1 {
		t.Errorf("AutoFit on, over-wide value: scale = %v, want (0,1)", on)
	}
	if off := r.displayRunScale(false, long, pptx.TypeDisplay, boxW); off != 0 {
		t.Errorf("AutoFit off: scale = %v, want 0", off)
	}
	if fit := r.displayRunScale(true, "9", pptx.TypeDisplay, pptx.In(6)); fit != 0 {
		t.Errorf("AutoFit on, fitting value: scale = %v, want 0", fit)
	}
}
