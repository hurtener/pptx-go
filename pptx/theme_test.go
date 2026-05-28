package pptx_test

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestDefaultThemeResolvesEveryRole(t *testing.T) {
	th := pptx.DefaultTheme()

	surfaces := []pptx.ColorRole{
		pptx.ColorCanvas, pptx.ColorSurface, pptx.ColorSurfaceAlt, pptx.ColorAccent,
		pptx.ColorAccentAlt, pptx.ColorAccentWarm, pptx.ColorSuccess, pptx.ColorWarning,
		pptx.ColorError, pptx.ColorInfo,
	}
	for _, r := range surfaces {
		if v := th.ResolveColor(r); len(v) != 6 {
			t.Errorf("ColorRole %d resolved to %q (want 6-hex)", r, v)
		}
	}
	texts := []pptx.TextColorRole{
		pptx.TextPrimary, pptx.TextSecondary, pptx.TextTertiary, pptx.TextInverse,
		pptx.TextMuted, pptx.TextAccent, pptx.TextAccentAlt, pptx.TextSuccess,
		pptx.TextWarning, pptx.TextError,
	}
	for _, r := range texts {
		if v := th.ResolveTextColor(r); len(v) != 6 {
			t.Errorf("TextColorRole %d resolved to %q (want 6-hex)", r, v)
		}
	}
	types := []pptx.TypeRole{
		pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3, pptx.TypeH4, pptx.TypeH5,
		pptx.TypeBody, pptx.TypeBodySmall, pptx.TypeCaption, pptx.TypeMono, pptx.TypeCode,
	}
	for _, r := range types {
		if spec := th.ResolveType(r); spec.Family == "" || spec.Size == 0 {
			t.Errorf("TypeRole %d resolved to %+v", r, spec)
		}
	}
	for _, r := range []pptx.SpaceRole{pptx.SpaceXS, pptx.SpaceSM, pptx.SpaceMD, pptx.SpaceLG, pptx.SpaceXL, pptx.Space2XL} {
		if th.ResolveSpace(r) <= 0 {
			t.Errorf("SpaceRole %d resolved to 0", r)
		}
	}
	// Radii: RadiusNone is legitimately 0.
	if th.ResolveRadius(pptx.RadiusMD) <= 0 {
		t.Error("RadiusMD resolved to 0")
	}
	if !th.ResolveElevation(pptx.ElevationFlat).IsFlat() {
		t.Error("ElevationFlat should be flat")
	}
	if th.ResolveElevation(pptx.ElevationRaised).IsFlat() {
		t.Error("ElevationRaised should not be flat")
	}
}

func TestResolutionDeterministic(t *testing.T) {
	th := pptx.DefaultTheme()
	for i := 0; i < 3; i++ {
		if th.ResolveColor(pptx.ColorAccent) != th.ResolveColor(pptx.ColorAccent) {
			t.Fatal("ResolveColor not deterministic")
		}
		if th.ResolveType(pptx.TypeH1) != th.ResolveType(pptx.TypeH1) {
			t.Fatal("ResolveType not deterministic")
		}
	}
}

func TestThemeSwapReResolves(t *testing.T) {
	a := pptx.DefaultTheme()
	b := pptx.NewTheme(pptx.WithAccent("FF0000"))
	if a.ResolveColor(pptx.ColorAccent) == b.ResolveColor(pptx.ColorAccent) {
		t.Fatal("theme swap: ColorAccent should differ between themes")
	}
	if b.ResolveColor(pptx.ColorAccent) != "FF0000" {
		t.Errorf("WithAccent: got %q want FF0000", b.ResolveColor(pptx.ColorAccent))
	}
	// The default theme must be unaffected by NewTheme's mutation (Clone).
	if a.ResolveColor(pptx.ColorAccent) == "FF0000" {
		t.Fatal("NewTheme leaked a mutation into the default theme")
	}
}

func TestResolveUnsetFallback(t *testing.T) {
	empty := &pptx.Theme{Colors: pptx.ColorPalette{Surfaces: map[pptx.ColorRole]pptx.RGB{}, Text: map[pptx.TextColorRole]pptx.RGB{}}}
	if empty.ResolveColor(pptx.ColorAccent) != "FFFFFF" {
		t.Errorf("unset surface fallback: got %q", empty.ResolveColor(pptx.ColorAccent))
	}
	if empty.ResolveTextColor(pptx.TextPrimary) != "000000" {
		t.Errorf("unset text fallback: got %q", empty.ResolveTextColor(pptx.TextPrimary))
	}
	if empty.ResolveType(pptx.TypeBody).Family == "" {
		t.Error("unset type fallback should have a family")
	}
}

func TestCloneIndependence(t *testing.T) {
	a := pptx.DefaultTheme()
	c := a.Clone()
	c.Colors.Surfaces[pptx.ColorAccent] = "00FF00"
	if a.ResolveColor(pptx.ColorAccent) == "00FF00" {
		t.Fatal("Clone shares the surfaces map with the original")
	}
}

func TestWithFonts(t *testing.T) {
	th := pptx.NewTheme(pptx.WithFonts("Inter", "Inter"))
	if th.HeadingFont != "Inter" || th.BodyFont != "Inter" {
		t.Errorf("WithFonts: heading=%q body=%q", th.HeadingFont, th.BodyFont)
	}
	if th.ResolveType(pptx.TypeH1).Family != "Inter" {
		t.Errorf("H1 family: got %q", th.ResolveType(pptx.TypeH1).Family)
	}
	if th.ResolveType(pptx.TypeBody).Family != "Inter" {
		t.Errorf("Body family: got %q", th.ResolveType(pptx.TypeBody).Family)
	}
	// Mono stays mono.
	if th.ResolveType(pptx.TypeMono).Family != "Consolas" {
		t.Errorf("Mono family should stay Consolas, got %q", th.ResolveType(pptx.TypeMono).Family)
	}
}
