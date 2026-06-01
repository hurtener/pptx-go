package frames_test

import (
	"testing"

	assetframes "github.com/hurtener/pptx-go/assets/frames"
	"github.com/hurtener/pptx-go/pptx"
)

// recipe is the shape every curated frame matches.
type recipe = func(sl *pptx.Slide, region pptx.Box) (pptx.Box, int)

func curated() map[string]recipe {
	return map[string]recipe{
		"browser": assetframes.Browser,
		"phone":   assetframes.Phone,
		"desktop": assetframes.Desktop,
		"laptop":  assetframes.Laptop,
	}
}

func newSlide(t *testing.T) *pptx.Slide {
	t.Helper()
	return pptx.New().AddSlide("")
}

// TestRecipe_InteriorWithinRegion is acceptance criterion 1 at the recipe level:
// every curated frame returns an interior box strictly inside the region (the
// bezel occupies the margin) with positive extent, and emits at least one bezel
// shape.
func TestRecipe_InteriorWithinRegion(t *testing.T) {
	region := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(6), H: pptx.In(4)}
	for name, rec := range curated() {
		t.Run(name, func(t *testing.T) {
			interior, shapes := rec(newSlide(t), region)
			if shapes < 1 {
				t.Fatalf("%s: shapes = %d, want >= 1", name, shapes)
			}
			if interior.W <= 0 || interior.H <= 0 {
				t.Fatalf("%s: interior has non-positive extent: %+v", name, interior)
			}
			if interior.X < region.X || interior.Y < region.Y ||
				interior.Right() > region.Right() || interior.Bottom() > region.Bottom() {
				t.Fatalf("%s: interior %+v escapes region %+v", name, interior, region)
			}
			if interior == region {
				t.Fatalf("%s: interior equals region (no bezel margin)", name)
			}
			if interior.W >= region.W && interior.H >= region.H {
				t.Fatalf("%s: interior %+v not strictly inside region %+v", name, interior, region)
			}
		})
	}
}

// TestRecipe_Deterministic is the D-035 invariant at the recipe level: the
// geometry is a pure function of the region — two calls return identical
// interior boxes and shape counts.
func TestRecipe_Deterministic(t *testing.T) {
	region := pptx.Box{X: pptx.In(0.5), Y: pptx.In(0.5), W: pptx.In(5), H: pptx.In(3)}
	for name, rec := range curated() {
		t.Run(name, func(t *testing.T) {
			a, an := rec(newSlide(t), region)
			b, bn := rec(newSlide(t), region)
			if a != b || an != bn {
				t.Fatalf("%s: non-deterministic geometry: (%+v,%d) vs (%+v,%d)", name, a, an, b, bn)
			}
		})
	}
}
