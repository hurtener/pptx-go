// Package embeddings holds the wire constants and helpers for embedding font
// data in a PPTX: the *.fntdata parts and the relationship that ties them to
// presentation.xml's <p:embeddedFontLst> (RFC §7.6, D-019). The presentation
// part owns the list XML; this package owns the part typing and naming.
package embeddings

import "fmt"

const (
	// ContentTypeFontData is the OOXML content type for an embedded font part.
	ContentTypeFontData = "application/x-fontdata"

	// RelTypeFont is the relationship type from presentation.xml to a font part.
	RelTypeFont = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/font"
)

// FontPartURI returns the package URI for the nth embedded font part (1-based).
func FontPartURI(n int) string {
	return fmt.Sprintf("/ppt/fonts/font%d.fntdata", n)
}

// FontRelTarget returns the relationship target (relative to presentation.xml)
// for the nth embedded font part.
func FontRelTarget(n int) string {
	return fmt.Sprintf("fonts/font%d.fntdata", n)
}

// StyleFor maps a weight + italic flag to the <p:embeddedFont> face slot:
// "regular", "bold", "italic", or "boldItalic". Weight ≥ 600 is bold.
func StyleFor(weight int, italic bool) string {
	bold := weight >= 600
	switch {
	case bold && italic:
		return "boldItalic"
	case bold:
		return "bold"
	case italic:
		return "italic"
	default:
		return "regular"
	}
}
