package ornaments_test

import (
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/ornaments"
)

func noopRecipe(_ *pptx.Slide, _ pptx.Box, _ int, _ float64, _ pptx.ColorRole, _ pptx.EMU) int {
	return 0
}

// TestCurated_HasSixOrnaments checks the curated factory exposes the six
// reserved names and resolves each.
func TestCurated_HasSixOrnaments(t *testing.T) {
	reg := ornaments.Curated()
	want := []string{"chevron_arrow", "corner_bracket", "glow_ring", "grid_dots", "noise_overlay", "radial_glow", "starfield"} // sorted
	if got := reg.Names(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for _, n := range want {
		if _, ok := reg.Lookup(n); !ok {
			t.Errorf("curated registry missing %q", n)
		}
	}
}

// TestWith_OverlayIsImmutable checks With returns a copy (per-render, not global).
func TestWith_OverlayIsImmutable(t *testing.T) {
	base := ornaments.Curated()
	ext := base.With("spark", noopRecipe)
	if _, ok := ext.Lookup("spark"); !ok {
		t.Fatal("extended registry missing the registered ornament")
	}
	if _, ok := base.Lookup("spark"); ok {
		t.Fatal("With mutated the base registry (must return a copy)")
	}
}

// TestLookup_UnknownAndNil covers misses and nil-safety.
func TestLookup_UnknownAndNil(t *testing.T) {
	if _, ok := ornaments.Curated().Lookup("nope"); ok {
		t.Error("Lookup(nope) reported found")
	}
	var nilReg *ornaments.Registry
	if _, ok := nilReg.Lookup("glow_ring"); ok {
		t.Error("nil registry Lookup reported found")
	}
	if names := nilReg.Names(); names != nil {
		t.Errorf("nil registry Names() = %v, want nil", names)
	}
}
