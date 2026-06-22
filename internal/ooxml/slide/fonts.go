package slide

// Font-face collection for the automatic font-embedding pass (RFC §7.6, R9.1).
// The walk mirrors XSpTree.DroppedDescendants: it visits every text body in the
// shape tree — shape text bodies and table-cell text bodies — and reports the
// distinct typefaces explicitly set on runs. Group shapes carry no text in the
// V1 model, so they are not traversed.

// FontFace is one distinct text face used on a slide: a Latin typeface plus its
// bold/italic flags. The emitted run property (a:rPr) carries only b/i, not a
// numeric weight, so a face buckets into the four OOXML embeddedFont slots
// (regular / bold / italic / boldItalic) — this is the unit the embedding pass
// embeds.
type FontFace struct {
	Typeface string
	Bold     bool
	Italic   bool
}

// UsedFontFaces returns the distinct faces explicitly set on the slide's runs,
// in document order (deduped per slide). A run that does not set a Latin
// typeface inherits the theme major/minor fonts (carried by theme1.xml, not a
// per-run face) and is not reported — the embedding pass embeds the explicitly
// set per-run faces, which is where a brand display/heading face lands. The
// caller merges and sorts across slides for determinism.
func (s *SlidePart) UsedFontFaces() []FontFace {
	if s == nil {
		return nil
	}
	var out []FontFace
	seen := map[FontFace]bool{}
	add := func(tb *XTextBody) {
		if tb == nil {
			return
		}
		for pi := range tb.Paragraphs {
			for _, r := range tb.Paragraphs[pi].Runs() {
				pr := r.TextProperties
				if pr == nil || pr.Latin == nil || pr.Latin.Typeface == "" {
					continue
				}
				f := FontFace{
					Typeface: pr.Latin.Typeface,
					Bold:     pr.Bold == "1",
					Italic:   pr.Italic == "1",
				}
				if seen[f] {
					continue
				}
				seen[f] = true
				out = append(out, f)
			}
		}
	}
	for _, child := range s.SpTree().Children {
		switch c := child.(type) {
		case *XSp:
			add(c.TextBody)
		case *XGraphicFrame:
			if c.Graphic == nil || c.Graphic.GraphicData == nil || c.Graphic.GraphicData.Table == nil {
				continue
			}
			for ri := range c.Graphic.GraphicData.Table.Rows {
				for ci := range c.Graphic.GraphicData.Table.Rows[ri].Cells {
					add(c.Graphic.GraphicData.Table.Rows[ri].Cells[ci].TextBody)
				}
			}
		}
	}
	return out
}

// RewriteFontFaces rewrites every run's Latin typeface that matches a key in
// mapping to the mapped face, in place, over the same text bodies UsedFontFaces
// walks (shape + table cells). It realizes the declared font fallback chain
// (R9.6, D-066): the chosen face is recorded as the run's single-valued a:latin
// typeface. A nil/empty mapping is a no-op. It reports the number of runs
// rewritten.
func (s *SlidePart) RewriteFontFaces(mapping map[string]string) int {
	if s == nil || len(mapping) == 0 {
		return 0
	}
	n := 0
	rewrite := func(tb *XTextBody) {
		if tb == nil {
			return
		}
		for pi := range tb.Paragraphs {
			for _, r := range tb.Paragraphs[pi].Runs() {
				pr := r.TextProperties
				if pr == nil || pr.Latin == nil {
					continue
				}
				if to, ok := mapping[pr.Latin.Typeface]; ok && to != pr.Latin.Typeface {
					pr.Latin.Typeface = to
					n++
				}
			}
		}
	}
	for _, child := range s.SpTree().Children {
		switch c := child.(type) {
		case *XSp:
			rewrite(c.TextBody)
		case *XGraphicFrame:
			if c.Graphic == nil || c.Graphic.GraphicData == nil || c.Graphic.GraphicData.Table == nil {
				continue
			}
			for ri := range c.Graphic.GraphicData.Table.Rows {
				for ci := range c.Graphic.GraphicData.Table.Rows[ri].Cells {
					rewrite(c.Graphic.GraphicData.Table.Rows[ri].Cells[ci].TextBody)
				}
			}
		}
	}
	return n
}
