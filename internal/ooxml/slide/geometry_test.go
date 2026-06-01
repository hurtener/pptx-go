package slide

import (
	"encoding/xml"
	"strings"
	"testing"
)

// sampleGeom is a small custom geometry exercising every command kind.
func sampleGeom() *XCustomGeometry {
	return &XCustomGeometry{
		AvLst: &XAvLst{},
		GdLst: &XGdLst{},
		PathList: XPathList{Paths: []XPath{{
			W: 2400, H: 2400,
			Commands: []XPathCommand{
				{Cmd: PathMoveTo, Pts: []XPoint{{X: 100, Y: 100}}},
				{Cmd: PathLnTo, Pts: []XPoint{{X: 2300, Y: 100}}},
				{Cmd: PathCubicTo, Pts: []XPoint{{X: 2300, Y: 800}, {X: 1800, Y: 1200}, {X: 1200, Y: 1200}}},
				{Cmd: PathQuadTo, Pts: []XPoint{{X: 400, Y: 1200}, {X: 100, Y: 800}}},
				{Cmd: PathClose},
			},
		}}},
	}
}

// TestCustomGeometry_Marshal checks the wire shape: a custGeom with a single
// path carrying ordered commands and points (bare element names — namespaces
// are restored downstream).
func TestCustomGeometry_Marshal(t *testing.T) {
	out, err := xml.Marshal(sampleGeom())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(out)
	for _, want := range []string{
		`<custGeom>`, `<pathLst>`, `<path w="2400" h="2400">`,
		`<moveTo><pt x="100" y="100"></pt></moveTo>`,
		`<lnTo><pt x="2300" y="100"></pt></lnTo>`,
		`<cubicBezTo><pt x="2300" y="800"></pt><pt x="1800" y="1200"></pt><pt x="1200" y="1200"></pt></cubicBezTo>`,
		`<quadBezTo><pt x="400" y="1200"></pt><pt x="100" y="800"></pt></quadBezTo>`,
		`<close></close>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("marshaled custGeom missing %q in:\n%s", want, got)
		}
	}
}

// TestCustomGeometry_RoundTrip is the D-032 invariant: a custGeom marshals,
// unmarshals, and re-marshals to identical bytes (the path commands + points
// survive).
func TestCustomGeometry_RoundTrip(t *testing.T) {
	first, err := xml.Marshal(sampleGeom())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back XCustomGeometry
	if err := xml.Unmarshal(first, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if n := len(back.PathList.Paths); n != 1 {
		t.Fatalf("paths = %d, want 1", n)
	}
	p := back.PathList.Paths[0]
	if p.W != 2400 || p.H != 2400 {
		t.Errorf("path w/h = %d/%d, want 2400/2400", p.W, p.H)
	}
	if len(p.Commands) != 5 {
		t.Fatalf("commands = %d, want 5", len(p.Commands))
	}
	if p.Commands[2].Cmd != PathCubicTo || len(p.Commands[2].Pts) != 3 {
		t.Errorf("cubic command not reconstructed: %+v", p.Commands[2])
	}
	second, err := xml.Marshal(&back)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("round-trip not byte-identical:\n first=%s\nsecond=%s", first, second)
	}
}
