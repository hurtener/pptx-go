package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestDarkThemeFrom_NilFallback verifies that a base theme with no dark palette
// yields the pinned Tailwind-gray dark theme verbatim (byte-identical fallback,
// R8.3).
func TestDarkThemeFrom_NilFallback(t *testing.T) {
	base := pptx.DefaultTheme()
	if base.DarkColors != nil {
		t.Fatal("DefaultTheme should carry no dark palette")
	}
	dark := darkThemeFrom(base)
	checks := map[pptx.ColorRole]pptx.RGB{
		pptx.ColorCanvas:     darkCanvas,
		pptx.ColorSurface:    darkSurface,
		pptx.ColorSurfaceAlt: darkSurfaceAlt,
	}
	for role, want := range checks {
		if got := dark.Colors.Surfaces[role]; got != want {
			t.Errorf("nil fallback surface %v = %q, want pinned %q", role, got, want)
		}
	}
	if got := dark.Colors.Text[pptx.TextPrimary]; got != darkTextPrimary {
		t.Errorf("nil fallback TextPrimary = %q, want pinned %q", got, darkTextPrimary)
	}
}

// TestDarkThemeFrom_Overlay verifies the soul-driven dark palette overlays the
// pinned default role-by-role: overridden roles take the soul value, unset roles
// keep the pinned default (R8.3).
func TestDarkThemeFrom_Overlay(t *testing.T) {
	base := pptx.NewTheme(
		pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"),
		pptx.WithDarkText(pptx.TextPrimary, "F4F6FF"),
		pptx.WithDarkSurface(pptx.ColorAccent, "5EEAD4"), // R8.7-style accent override
	)
	dark := darkThemeFrom(base)

	// Overridden roles take the soul value.
	if got := dark.Colors.Surfaces[pptx.ColorCanvas]; got != "0A0E1A" {
		t.Errorf("overlay canvas = %q, want 0A0E1A", got)
	}
	if got := dark.Colors.Text[pptx.TextPrimary]; got != "F4F6FF" {
		t.Errorf("overlay TextPrimary = %q, want F4F6FF", got)
	}
	if got := dark.Colors.Surfaces[pptx.ColorAccent]; got != "5EEAD4" {
		t.Errorf("overlay accent = %q, want 5EEAD4", got)
	}
	// Unset roles keep the pinned default (surface/surfaceAlt were not overridden).
	if got := dark.Colors.Surfaces[pptx.ColorSurface]; got != darkSurface {
		t.Errorf("unset surface = %q, want pinned %q", got, darkSurface)
	}
	if got := dark.Colors.Text[pptx.TextSecondary]; got != darkTextSecondary {
		t.Errorf("unset TextSecondary = %q, want pinned %q", got, darkTextSecondary)
	}
	// The overlay must not have mutated the base theme's dark palette.
	if base.DarkColors.Surfaces[pptx.ColorCanvas] != "0A0E1A" {
		t.Error("darkThemeFrom mutated the base theme's dark palette")
	}
}
