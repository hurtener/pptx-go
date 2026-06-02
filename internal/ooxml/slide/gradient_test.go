package slide

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestGradientFill_Linear checks a linear gradient marshals to a gsLst + lin.
func TestGradientFill_Linear(t *testing.T) {
	g := &XGradientFill{
		GsLst: XGradientStopList{Gs: []XGradientStop{
			{Pos: 0, SrgbClr: &XSrgbClr{Val: "2563EB"}},
			{Pos: 100000, SrgbClr: &XSrgbClr{Val: "FFFFFF"}},
		}},
		Lin: &XLinearGradient{Ang: 5400000},
	}
	out, err := xml.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		`<gradFill>`, `<gsLst>`,
		`<gs pos="0"><srgbClr val="2563EB"></srgbClr></gs>`,
		`<gs pos="100000"><srgbClr val="FFFFFF"></srgbClr></gs>`,
		`<lin ang="5400000">`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("linear gradient missing %q in:\n%s", want, got)
		}
	}
}

// TestGradientFill_Radial checks a radial gradient marshals to path="circle" +
// fillToRect, and that an alpha stop (a glow edge) survives.
func TestGradientFill_Radial(t *testing.T) {
	g := &XGradientFill{
		GsLst: XGradientStopList{Gs: []XGradientStop{
			{Pos: 0, SrgbClr: &XSrgbClr{Val: "2563EB"}},
			{Pos: 100000, SrgbClr: &XSrgbClr{Val: "2563EB", Alpha: &XAlpha{Val: 0}}},
		}},
		Path: &XPathGradient{Path: "circle", FillToRect: &XFillToRect{L: 50000, T: 50000, R: 50000, B: 50000}},
	}
	out, err := xml.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		`<path path="circle">`,
		`<fillToRect l="50000" t="50000" r="50000" b="50000">`,
		`<alpha val="0">`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("radial gradient missing %q in:\n%s", want, got)
		}
	}
}

// TestGradientFill_RoundTrip checks a gradient survives marshal → unmarshal →
// re-marshal byte-identically (D-032).
func TestGradientFill_RoundTrip(t *testing.T) {
	g := &XGradientFill{
		GsLst: XGradientStopList{Gs: []XGradientStop{
			{Pos: 0, SrgbClr: &XSrgbClr{Val: "112233"}},
			{Pos: 100000, SrgbClr: &XSrgbClr{Val: "112233", Alpha: &XAlpha{Val: 0}}},
		}},
		Path: &XPathGradient{Path: "circle", FillToRect: &XFillToRect{L: 50000, T: 50000, R: 50000, B: 50000}},
	}
	first, err := xml.Marshal(g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back XGradientFill
	if err := xml.Unmarshal(first, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if n := len(back.GsLst.Gs); n != 2 {
		t.Fatalf("stops = %d, want 2", n)
	}
	if back.Path == nil || back.Path.Path != "circle" {
		t.Fatalf("radial path not reconstructed: %+v", back.Path)
	}
	second, err := xml.Marshal(&back)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("gradient round-trip not byte-identical:\n first=%s\nsecond=%s", first, second)
	}
}
