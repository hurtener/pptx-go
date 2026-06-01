package icons_test

import (
	"testing"

	"github.com/hurtener/pptx-go/scene/icons"
)

// TestCurated_HasStarterSet checks the curated factory exposes the embedded
// starter set and resolves a known icon.
func TestCurated_HasStarterSet(t *testing.T) {
	reg := icons.Curated()
	if len(reg.Names()) < 16 {
		t.Fatalf("curated registry has %d icons, want the starter set (>= 16)", len(reg.Names()))
	}
	for _, n := range []string{"arrow-right", "check", "circle", "star"} {
		if _, ok := reg.Lookup(n); !ok {
			t.Errorf("curated registry missing %q", n)
		}
	}
}

// TestWith_OverlayIsImmutable checks With returns a copy (per-render, not global).
func TestWith_OverlayIsImmutable(t *testing.T) {
	base := icons.Curated()
	ext := base.With("custom", []byte("<svg/>"))
	if _, ok := ext.Lookup("custom"); !ok {
		t.Fatal("extended registry missing the registered icon")
	}
	if _, ok := base.Lookup("custom"); ok {
		t.Fatal("With mutated the base registry (must return a copy)")
	}
}

// TestLookup_UnknownAndNil covers misses and nil-safety.
func TestLookup_UnknownAndNil(t *testing.T) {
	if _, ok := icons.Curated().Lookup("nope"); ok {
		t.Error("Lookup(nope) reported found")
	}
	var nilReg *icons.Registry
	if _, ok := nilReg.Lookup("check"); ok {
		t.Error("nil registry Lookup reported found")
	}
	if names := nilReg.Names(); names != nil {
		t.Errorf("nil registry Names() = %v, want nil", names)
	}
}
