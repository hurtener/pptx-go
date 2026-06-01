package render

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

func mustTranslate(t *testing.T, svg string) *slide.XCustomGeometry {
	t.Helper()
	g, err := Translate([]byte(svg))
	if err != nil {
		t.Fatalf("Translate(%q): %v", svg, err)
	}
	return g
}

// TestTranslate_Commands covers the supported command subset and the viewBox→
// path coordinate scaling (a 24-unit viewBox → a 2400-unit path grid).
func TestTranslate_Commands(t *testing.T) {
	svg := `<svg viewBox="0 0 24 24"><path d="M2 2 L22 2 H22 V22 C22 10 18 6 12 6 Q6 6 2 12 Z" fill="black"/></svg>`
	g := mustTranslate(t, svg)
	if len(g.PathList.Paths) != 1 {
		t.Fatalf("paths = %d, want 1", len(g.PathList.Paths))
	}
	p := g.PathList.Paths[0]
	if p.W != 2400 || p.H != 2400 {
		t.Errorf("path grid = %d×%d, want 2400×2400", p.W, p.H)
	}
	gotKinds := make([]string, len(p.Commands))
	for i, c := range p.Commands {
		gotKinds[i] = c.Cmd
	}
	wantKinds := []string{
		slide.PathMoveTo, slide.PathLnTo, slide.PathLnTo, slide.PathLnTo,
		slide.PathCubicTo, slide.PathQuadTo, slide.PathClose,
	}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Errorf("command kinds = %v, want %v", gotKinds, wantKinds)
	}
	// M2 2 → (2*100, 2*100) = (200, 200).
	if p.Commands[0].Pts[0] != (slide.XPoint{X: 200, Y: 200}) {
		t.Errorf("moveTo point = %+v, want {200 200}", p.Commands[0].Pts[0])
	}
	// H22 keeps current y (=2 after the L), so → (2200, 200).
	if p.Commands[2].Pts[0] != (slide.XPoint{X: 2200, Y: 200}) {
		t.Errorf("H point = %+v, want {2200 200}", p.Commands[2].Pts[0])
	}
}

// TestTranslate_Relative checks lowercase commands accumulate from the current
// point.
func TestTranslate_Relative(t *testing.T) {
	g := mustTranslate(t, `<svg viewBox="0 0 10 10"><path d="M1 1 l2 0 l0 2 z"/></svg>`)
	p := g.PathList.Paths[0]
	// m1 1 → (100,100); l2 0 → (300,100); l0 2 → (300,300).
	if p.Commands[1].Pts[0] != (slide.XPoint{X: 300, Y: 100}) {
		t.Errorf("relative L1 = %+v, want {300 100}", p.Commands[1].Pts[0])
	}
	if p.Commands[2].Pts[0] != (slide.XPoint{X: 300, Y: 300}) {
		t.Errorf("relative L2 = %+v, want {300 300}", p.Commands[2].Pts[0])
	}
}

// TestTranslate_SmoothReflection checks S reflects the previous cubic's control
// point about the current point.
func TestTranslate_SmoothReflection(t *testing.T) {
	// After "C ... 6 6  10 10" the current point is (10,10) and the last control
	// is (6,6); a following "S 14 10  18 10" must use first control = 2*10-6 = 14.
	g := mustTranslate(t, `<svg viewBox="0 0 20 20"><path d="M0 10 C2 6 6 6 10 10 S14 10 18 10"/></svg>`)
	p := g.PathList.Paths[0]
	s := p.Commands[len(p.Commands)-1]
	if s.Cmd != slide.PathCubicTo {
		t.Fatalf("smooth command kind = %s, want cubicBezTo", s.Cmd)
	}
	if s.Pts[0] != (slide.XPoint{X: 1400, Y: 1400}) { // (14,14)*100 — reflected y too (2*10-6=14)
		t.Errorf("reflected control = %+v, want {1400 1400}", s.Pts[0])
	}
}

// TestTranslate_Deterministic checks the translator is a pure function.
func TestTranslate_Deterministic(t *testing.T) {
	svg := `<svg viewBox="0 0 24 24"><path d="M3 3 L21 3 L21 21 Z"/></svg>`
	a := mustTranslate(t, svg)
	b := mustTranslate(t, svg)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("translation not deterministic")
	}
}

// TestTranslate_Rejections covers each documented constraint violation — all
// fail at translation (i.e. at registration), never silently.
func TestTranslate_Rejections(t *testing.T) {
	cases := []struct {
		name, svg, wantSub string
	}{
		{"multi-path", `<svg viewBox="0 0 24 24"><path d="M0 0 L1 1"/><path d="M2 2 L3 3"/></svg>`, "exactly one"},
		{"circle element", `<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="6"/></svg>`, "circle"},
		{"line element", `<svg viewBox="0 0 24 24"><path d="M0 0 L1 1"/><line x1="0" y1="0" x2="1" y2="1"/></svg>`, "line"},
		{"fill none", `<svg viewBox="0 0 24 24"><path d="M0 0 L1 1" fill="none"/></svg>`, "none"},
		{"gradient fill", `<svg viewBox="0 0 24 24"><path d="M0 0 L1 1" fill="url(#g)"/></svg>`, "gradient"},
		{"arc command", `<svg viewBox="0 0 24 24"><path d="M0 0 A5 5 0 0 1 10 10"/></svg>`, "arc"},
		{"no path", `<svg viewBox="0 0 24 24"></svg>`, "no <path>"},
		{"empty d", `<svg viewBox="0 0 24 24"><path d=""/></svg>`, "no d data"},
		{"no viewBox", `<svg><path d="M0 0 L1 1"/></svg>`, "viewBox"},
		{"bad number", `<svg viewBox="0 0 24 24"><path d="M0 0 L1 ..2"/></svg>`, "malformed"},
		{"unknown command", `<svg viewBox="0 0 24 24"><path d="M0 0 K1 1"/></svg>`, "unexpected character"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Translate([]byte(tc.svg))
			if err == nil {
				t.Fatalf("expected an error for %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("error %q does not contain %q", err, tc.wantSub)
			}
		})
	}
}
