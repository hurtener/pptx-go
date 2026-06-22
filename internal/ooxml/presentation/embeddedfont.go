package presentation

// Embedded-font support for presentation.xml (<p:embeddedFontLst>). PowerPoint
// renders an embedded font only if it is declared here with a relationship to
// the font-data part; the bytes alone are not enough. RFC §7.6 / D-019.

// EmbeddedFontEntry records one embedded face style for a typeface, keyed to
// the relationship that points at its font-data part.
type EmbeddedFontEntry struct {
	Typeface string
	Style    string // "regular" | "bold" | "italic" | "boldItalic"
	RID      string
}

// XEmbeddedFontList models <p:embeddedFontLst>.
type XEmbeddedFontList struct {
	XMLName struct{}        `xml:"embeddedFontLst"`
	Fonts   []XEmbeddedFont `xml:"embeddedFont"`
}

// XEmbeddedFont models one <p:embeddedFont> (a typeface + its face refs).
type XEmbeddedFont struct {
	Font       XEmbeddedFontName `xml:"font"`
	Regular    *XFontStyleRef    `xml:"regular,omitempty"`
	Bold       *XFontStyleRef    `xml:"bold,omitempty"`
	Italic     *XFontStyleRef    `xml:"italic,omitempty"`
	BoldItalic *XFontStyleRef    `xml:"boldItalic,omitempty"`
}

// XEmbeddedFontName is the <p:font typeface="…"/> element.
type XEmbeddedFontName struct {
	Typeface string `xml:"typeface,attr"`
}

// XFontStyleRef is a face ref (<p:regular r:id="…"/> etc.).
type XFontStyleRef struct {
	RID string `xml:"rid,attr"`
}

// AddEmbeddedFont records an embedded font face. style is one of
// "regular", "bold", "italic", "boldItalic"; rID is the relationship to the
// font-data part. Entries serialize into <p:embeddedFontLst> on ToXML.
func (p *PresentationPart) AddEmbeddedFont(typeface, style, rID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.embeddedFonts = append(p.embeddedFonts, EmbeddedFontEntry{Typeface: typeface, Style: style, RID: rID})
}

// EmbeddedFontCount returns the number of recorded embedded font faces.
func (p *PresentationPart) EmbeddedFontCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.embeddedFonts)
}

// HasEmbeddedFace reports whether a face for the given typeface and style slot
// ("regular" | "bold" | "italic" | "boldItalic") is already recorded. The
// automatic embedding pass uses it to skip a face a caller embedded by hand,
// keeping the pass idempotent against manual EmbedFont calls.
func (p *PresentationPart) HasEmbeddedFace(typeface, style string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, e := range p.embeddedFonts {
		if e.Typeface == typeface && e.Style == style {
			return true
		}
	}
	return false
}

// buildEmbeddedFontList groups the recorded entries by typeface (preserving
// first-seen order) into the XML model. Returns nil when there are none.
func buildEmbeddedFontList(entries []EmbeddedFontEntry) *XEmbeddedFontList {
	if len(entries) == 0 {
		return nil
	}
	var order []string
	byFace := map[string]*XEmbeddedFont{}
	for _, e := range entries {
		ef, ok := byFace[e.Typeface]
		if !ok {
			ef = &XEmbeddedFont{Font: XEmbeddedFontName{Typeface: e.Typeface}}
			byFace[e.Typeface] = ef
			order = append(order, e.Typeface)
		}
		ref := &XFontStyleRef{RID: e.RID}
		switch e.Style {
		case "bold":
			ef.Bold = ref
		case "italic":
			ef.Italic = ref
		case "boldItalic":
			ef.BoldItalic = ref
		default:
			ef.Regular = ref
		}
	}
	lst := &XEmbeddedFontList{}
	for _, face := range order {
		lst.Fonts = append(lst.Fonts, *byFace[face])
	}
	return lst
}
