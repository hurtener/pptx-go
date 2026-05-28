package theme

// Domain accessors for the font scheme, so consumers (the pptx theme codec)
// can read font faces without traversing the raw XML structs (P3).

// MajorLatinFont returns the heading (major) Latin typeface, or "" if unset.
func (t *ThemePart) MajorLatinFont() string {
	if fs := t.FontScheme(); fs != nil && fs.MajorFont != nil {
		if fs.MajorFont.Latin != nil {
			return fs.MajorFont.Latin.Typeface
		}
	}
	return ""
}

// MinorLatinFont returns the body (minor) Latin typeface, or "" if unset.
func (t *ThemePart) MinorLatinFont() string {
	if fs := t.FontScheme(); fs != nil && fs.MinorFont != nil {
		if fs.MinorFont.Latin != nil {
			return fs.MinorFont.Latin.Typeface
		}
	}
	return ""
}
