package slide

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

// FuzzParseSlide exercises the slide shape-tree parser — the read surface an
// external (non-pptx-go) deck degrades against (RFC §16; D-048). The seeds are
// external-style structures pptx-go does not model: group shapes,
// mc:AlternateContent, foreign namespaces, and truncated parts. Invariant: the
// parser never panics; it returns either a populated part or an error, and the
// dropped-child collection is always safe to read.
func FuzzParseSlide(f *testing.F) {
	// Well-formed minimal slide.
	f.Add([]byte(`<p:sld xmlns:p="x" xmlns:a="y"><p:cSld><p:spTree><p:nvGrpSpPr/><p:grpSpPr/></p:spTree></p:cSld></p:sld>`))
	// A modeled shape alongside an unmodeled group shape (dropped).
	f.Add([]byte(`<p:sld><p:cSld><p:spTree><p:sp><p:nvSpPr><p:cNvPr id="2" name="t"/></p:nvSpPr></p:sp><p:grpSp/></p:spTree></p:cSld></p:sld>`))
	// mc:AlternateContent — the canonical PowerPoint compatibility wrapper.
	f.Add([]byte(`<p:sld><p:cSld><p:spTree><mc:AlternateContent><mc:Choice Requires="a14"/><mc:Fallback/></mc:AlternateContent></p:spTree></p:cSld></p:sld>`))
	// Foreign namespace element with attributes and text.
	f.Add([]byte(`<p:sld><p:cSld><p:spTree><x:custom xmlns:x="urn:foreign"><x:inner k="v">payload</x:inner></x:custom></p:spTree></p:cSld></p:sld>`))
	// Structural edge cases: no spTree, empty, garbage, truncated.
	f.Add([]byte(`<p:sld><p:cSld/></p:sld>`))
	f.Add([]byte(``))
	f.Add([]byte(`not xml`))
	f.Add([]byte(`<p:sld><p:cSld><p:spTree><p:sp`))

	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	f.Fuzz(func(t *testing.T, data []byte) {
		sp := NewSlidePartWithURI(uri)
		// The only invariant is no panic; an error is an acceptable outcome for a
		// malformed external part (best-effort read).
		_ = sp.FromXML(data)
		// Reading the dropped-child collection must always be safe, even after a
		// failed parse.
		_ = sp.SpTree().DroppedChildren()
	})
}
