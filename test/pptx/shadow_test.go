package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestWithElevation_EmitsAndRoundTrips is Phase 14 PR#1 acceptance criterion 1:
// WithElevation emits an <a:outerShdw> from the theme token, the deck conforms,
// and the shadow survives Open → re-save (G6).
func TestWithElevation_EmitsAndRoundTrips(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.AddShape(pptx.ShapeRoundRect, fxBox,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurface))),
		pptx.WithElevation(pptx.ElevationElevated))

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("elevation deck failed conformance:\n%s", rep)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	// Default ElevationElevated: Blur=Pt(12), OffsetY=Pt(4) straight down (dir
	// 5400000), Color 000000, Alpha 35000.
	for _, want := range []string{
		`<a:effectLst>`, `<a:outerShdw`, `blurRad="152400"`, `dist="50800"`,
		`dir="5400000"`, `rotWithShape="0"`, `<a:srgbClr val="000000"`, `<a:alpha val="35000"`,
	} {
		if !strings.Contains(slide, want) {
			t.Errorf("elevation slide missing %q in:\n%s", want, slide)
		}
	}
	// Round-trip fidelity (G6): the shadow survives Open → re-save, not just a
	// clean reopen.
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("reopen elevation deck: %v", err)
	}
	resaved, err := reopened.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save reopened elevation deck: %v", err)
	}
	if rs := readZipPart(t, resaved, "ppt/slides/slide1.xml"); !strings.Contains(rs, "<a:outerShdw") || !strings.Contains(rs, `blurRad="152400"`) {
		t.Errorf("shadow did not survive round-trip through Open:\n%s", rs)
	}
}

// TestWithShadow_LiteralEscapeHatch is PR#1 acceptance: a literal Elevation
// emits the same wire with the caller's blur/offset/color/alpha.
func TestWithShadow_LiteralEscapeHatch(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	// A literal shadow offset down-right (dx=dy) → dir 45° = 2700000; dist = the
	// hypotenuse; explicit color + alpha.
	s.AddShape(pptx.ShapeRoundRect, fxBox, pptx.WithShadow(pptx.Elevation{
		Blur: pptx.Pt(6), OffsetX: pptx.Pt(3), OffsetY: pptx.Pt(3), Color: "112233", Alpha: 40000,
	}))
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{`<a:outerShdw`, `dir="2700000"`, `<a:srgbClr val="112233"`, `<a:alpha val="40000"`} {
		if !strings.Contains(slide, want) {
			t.Errorf("literal shadow missing %q in:\n%s", want, slide)
		}
	}
}

// TestShadowOmittedWhenFlat is PR#1 acceptance criterion 2: a shape with no
// shadow option, and a shape with a flat elevation, emit no <a:effectLst> — the
// byte-identical guard that the primitive is opt-in and does not perturb
// existing output.
func TestShadowOmittedWhenFlat(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	// No shadow option at all.
	s.AddShape(pptx.ShapeRect, fxBox, pptx.WithFill(pptx.SolidFill(pptx.RGB("FFFFFF"))))
	// Explicit flat elevation — also a no-op.
	s.AddShape(pptx.ShapeRoundRect, fxBox, pptx.WithElevation(pptx.ElevationFlat))

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	if strings.Contains(slide, "<a:effectLst>") || strings.Contains(slide, "<a:outerShdw") {
		t.Errorf("flat / no-shadow shapes must not emit an effect list:\n%s", slide)
	}
}
