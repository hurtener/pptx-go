package presentation

import "testing"

// FuzzPresentationFromXML exercises the presentation.xml parser, which runs on
// every deck open (including untrusted external decks). Invariant: no panic; the
// parser returns parsed data or an error (CLAUDE.md §11).
func FuzzPresentationFromXML(f *testing.F) {
	f.Add([]byte(`<p:presentation xmlns:p="x" xmlns:r="y"><p:sldIdLst><p:sldId id="256" r:id="rId1"/></p:sldIdLst></p:presentation>`))
	f.Add([]byte(`<p:presentation><p:sldSz cx="12192000" cy="6858000"/></p:presentation>`))
	f.Add([]byte(`<p:presentation><p:sldIdLst><p:sldId/></p:sldIdLst></p:presentation>`))
	f.Add([]byte(``))
	f.Add([]byte(`not xml`))
	f.Add([]byte(`<p:presentation`))

	f.Fuzz(func(t *testing.T, data []byte) {
		p := NewPresentationPart()
		// The only invariant is no panic; an error is acceptable for malformed
		// input, and the part must remain safe to interrogate afterward.
		_ = p.FromXML(data)
		_ = p.SlideCount()
	})
}
