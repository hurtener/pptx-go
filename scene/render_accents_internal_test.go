package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestAccentResolvers_BrandPalette verifies accentColorAt/accentRGBAt cycle a
// theme's brand-accent palette by index (R8.4).
func TestAccentResolvers_BrandPalette(t *testing.T) {
	r := newTestRenderer(t)
	r.theme = pptx.NewTheme(pptx.WithAccents("AA0000", "00BB00", "0000CC"))

	// accentColorAt returns the literal RGB hue at the cycled index.
	cases := map[int]pptx.RGB{0: "AA0000", 1: "00BB00", 2: "0000CC", 3: "AA0000", 5: "0000CC"}
	for idx, want := range cases {
		got, ok := r.accentColorAt(idx).(pptx.RGB)
		if !ok || got != want {
			t.Errorf("accentColorAt(%d) = %v, want RGB %q", idx, r.accentColorAt(idx), want)
		}
		if rgb := r.accentRGBAt(idx); rgb != want {
			t.Errorf("accentRGBAt(%d) = %q, want %q", idx, rgb, want)
		}
	}
	// Negative index clamps to 0.
	if got := r.accentRGBAt(-1); got != "AA0000" {
		t.Errorf("accentRGBAt(-1) = %q, want AA0000 (clamped)", got)
	}
}

// TestAccentResolvers_PinnedFallback verifies that with no brand palette the
// resolvers reproduce the pinned five-role cycle (byte-identical, R8.4).
func TestAccentResolvers_PinnedFallback(t *testing.T) {
	r := newTestRenderer(t) // default theme, no Accents
	if len(r.theme.Accents) != 0 {
		t.Fatal("default theme should carry no brand-accent palette")
	}
	// accentRGBAt(idx) must equal ResolveColor(timelineAccent(idx)) for the cycle.
	want := []pptx.ColorRole{pptx.ColorAccent, pptx.ColorAccentAlt, pptx.ColorInfo, pptx.ColorSuccess, pptx.ColorWarning}
	for idx, role := range want {
		if got := r.accentRGBAt(idx); got != r.theme.ResolveColor(role) {
			t.Errorf("accentRGBAt(%d) = %q, want pinned role %v = %q", idx, got, role, r.theme.ResolveColor(role))
		}
	}
	// And it wraps with the cycle length (5).
	if r.accentRGBAt(5) != r.accentRGBAt(0) {
		t.Error("pinned cycle did not wrap at length 5")
	}
}
