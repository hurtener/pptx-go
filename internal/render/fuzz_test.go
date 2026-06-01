package render

import "testing"

// FuzzTranslate exercises the SVG parse/translate surface (RFC §11 fuzz
// requirement). Invariant: Translate never panics and never returns (nil, nil)
// — it yields either a geometry or an error for any input. The seed corpus runs
// as an ordinary CI test.
func FuzzTranslate(f *testing.F) {
	seeds := []string{
		`<svg viewBox="0 0 24 24"><path d="M2 2 L22 2 L22 22 Z"/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="M0 0 C1 1 2 2 3 3 S4 4 5 5"/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="M0 0 Q1 1 2 2 T4 4"/></svg>`,
		`<svg viewBox="0 0 10 10"><path d="m1 1 l2 0 h3 v3 z"/></svg>`,
		`<svg><path d=""/></svg>`,
		`<svg viewBox="0 0 24 24"><path d="M0 0 A5 5 0 0 1 10 10"/></svg>`,
		`<svg viewBox="0 0 24 24"><circle cx="1" cy="1" r="1"/></svg>`,
		`<svg viewBox="bad"><path d="M.. .."/></svg>`,
		`not xml at all`,
		``,
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, svg []byte) {
		g, err := Translate(svg)
		if err == nil && g == nil {
			t.Fatal("Translate returned (nil, nil)")
		}
		if err != nil && g != nil {
			t.Fatal("Translate returned both a geometry and an error")
		}
	})
}
