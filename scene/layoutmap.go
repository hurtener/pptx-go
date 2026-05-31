package scene

// LayoutMap maps a scene LayoutKind (a structural intent) to a named layout in
// the active template's master (RFC §13.2). When a scene is rendered with
// WithLayoutMap, each slide's LayoutKind is resolved through the map to a layout
// name and the slide is related to that template layout; an entry whose name the
// template doesn't define falls back to the blank layout and records a
// LayoutWarning (the engine never errors on a layout miss — D-026).
type LayoutMap map[LayoutKind]string

// nameFor returns the mapped layout name for a kind, or "" when the map is nil
// or has no entry for the kind (the slide then uses the default blank layout).
func (m LayoutMap) nameFor(kind LayoutKind) string {
	if m == nil {
		return ""
	}
	return m[kind]
}

// DefaultLayoutMap maps each LayoutKind to the conventional PowerPoint standard
// layout name. It is a convenience for callers ingesting a stock template whose
// layouts use PowerPoint's default English names; a brand kit with custom layout
// names needs a caller-supplied map. Unmapped kinds resolve to the blank layout.
func DefaultLayoutMap() LayoutMap {
	return LayoutMap{
		LayoutCover:        "Title Slide",
		LayoutTitleContent: "Title and Content",
		LayoutTwoColumn:    "Two Content",
		LayoutCardGrid:     "Title and Content",
		LayoutFullBleed:    "Blank",
		LayoutBlank:        "Blank",
	}
}
