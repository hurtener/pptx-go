package frames_test

import (
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/frames"
)

func noopRecipe(_ *pptx.Slide, region pptx.Box) (pptx.Box, int) { return region, 0 }

// TestCurated_HasFourFrames checks the curated factory exposes exactly the four
// reserved names and resolves each.
func TestCurated_HasFourFrames(t *testing.T) {
	reg := frames.Curated()
	want := []string{"browser", "desktop", "laptop", "phone"} // sorted
	if got := reg.Names(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for _, n := range []string{frames.NameBrowser, frames.NamePhone, frames.NameDesktop, frames.NameLaptop} {
		if _, ok := reg.Lookup(n); !ok {
			t.Errorf("curated registry missing %q", n)
		}
	}
}

// TestWith_OverlayIsImmutable checks With returns a copy: the new registry has
// the extension, the original does not (per-render, not global — §14.4).
func TestWith_OverlayIsImmutable(t *testing.T) {
	base := frames.Curated()
	ext := base.With("retro", noopRecipe)

	if _, ok := ext.Lookup("retro"); !ok {
		t.Fatal("extended registry missing the registered frame")
	}
	if _, ok := base.Lookup("retro"); ok {
		t.Fatal("With mutated the base registry (must return a copy)")
	}
}

// TestWith_OverridesCuratedName checks registering a curated name replaces that
// frame for the derived registry only.
func TestWith_OverridesCuratedName(t *testing.T) {
	base := frames.Curated()
	override := base.With(frames.NameBrowser, noopRecipe)

	rec, ok := override.Lookup(frames.NameBrowser)
	if !ok {
		t.Fatal("override registry missing browser")
	}
	// The override recipe returns 0 shapes; the curated browser returns several.
	if _, n := rec(pptx.New().AddSlide(""), pptx.Box{W: pptx.In(4), H: pptx.In(3)}); n != 0 {
		t.Errorf("browser was not overridden: shapes = %d, want 0", n)
	}
	if _, ok := base.Lookup(frames.NameBrowser); !ok {
		t.Error("base registry lost its curated browser")
	}
}

// TestLookup_Unknown and nil-safety.
func TestLookup_UnknownAndNil(t *testing.T) {
	if _, ok := frames.Curated().Lookup("nope"); ok {
		t.Error("Lookup(nope) reported found")
	}
	var nilReg *frames.Registry
	if _, ok := nilReg.Lookup("browser"); ok {
		t.Error("nil registry Lookup reported found")
	}
	if names := nilReg.Names(); names != nil {
		t.Errorf("nil registry Names() = %v, want nil", names)
	}
}
