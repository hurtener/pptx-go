package ornaments_test

import (
	"testing"

	assetornaments "github.com/hurtener/pptx-go/assets/ornaments"
	"github.com/hurtener/pptx-go/pptx"
)

type recipe = func(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64) int

func curated() map[string]recipe {
	return map[string]recipe{
		"glow_ring":      assetornaments.GlowRing,
		"radial_glow":    assetornaments.RadialGlow,
		"grid_dots":      assetornaments.GridDots,
		"corner_bracket": assetornaments.CornerBracket,
		"chevron_arrow":  assetornaments.ChevronArrow,
		"noise_overlay":  assetornaments.NoiseOverlay,
	}
}

func newSlide(t *testing.T) *pptx.Slide {
	t.Helper()
	return pptx.New().AddSlide("")
}

// TestOrnamentRecipes_EmitShapes checks every curated ornament emits at least
// one shape and reports a stable (deterministic) count.
func TestOrnamentRecipes_EmitShapes(t *testing.T) {
	box := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(3), H: pptx.In(3)}
	for name, rec := range curated() {
		t.Run(name, func(t *testing.T) {
			n1 := rec(newSlide(t), box, pptx.AlphaOpaque, 0)
			if n1 < 1 {
				t.Fatalf("%s emitted %d shapes, want >= 1", name, n1)
			}
			if n2 := rec(newSlide(t), box, pptx.AlphaOpaque, 0); n2 != n1 {
				t.Errorf("%s shape count not deterministic: %d vs %d", name, n1, n2)
			}
		})
	}
}

// TestChevronArrow_Rotation confirms the single-shape chevron accepts a rotation
// without panicking (the rotation-honoring ornament).
func TestChevronArrow_Rotation(t *testing.T) {
	box := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(2), H: pptx.In(2)}
	if n := assetornaments.ChevronArrow(newSlide(t), box, pptx.AlphaOpaque, 90); n != 1 {
		t.Errorf("ChevronArrow emitted %d shapes, want 1", n)
	}
}
