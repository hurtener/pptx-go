package theme

// ============================================================================
// Theme XML struct type definitions — corresponds to /ppt/theme/themeN.xml
// ============================================================================
//
// The theme file defines the color scheme, font scheme, and format scheme
// for a presentation.
// Namespace: http://schemas.openxmlformats.org/drawingml/2006/main
//
// Example file locations:
//   - /ppt/theme/theme1.xml
//   - /ppt/theme/theme2.xml
//
// ============================================================================

// XTheme is the root element of a theme XML file.
type XTheme struct {
	XMLName       struct{}        `xml:"theme"`
	XmlnsA        string          `xml:"xmlns:a,attr"`
	Name          string          `xml:"name,attr,omitempty"`
	ThemeElements *XThemeElements `xml:"themeElements"`
}

// XThemeElements groups all theme element collections.
type XThemeElements struct {
	ColorScheme *XColorScheme `xml:"clrScheme"`
	FontScheme  *XFontScheme  `xml:"fontScheme,omitempty"`
	FmtScheme   *XFmtScheme   `xml:"fmtScheme,omitempty"`
}

// ============================================================================
// Color Scheme
// ============================================================================

// XColorScheme defines the 12 standard colors used in a presentation.
type XColorScheme struct {
	XMLName           struct{}       `xml:"clrScheme"`
	XmlnsA            string         `xml:"xmlns:a,attr,omitempty"`
	Name              string         `xml:"name,attr,omitempty"`
	Dark1             *XColorVariant `xml:"dk1"`      // Dark 1
	Light1            *XColorVariant `xml:"lt1"`      // Light 1
	Dark2             *XColorVariant `xml:"dk2"`      // Dark 2
	Light2            *XColorVariant `xml:"lt2"`      // Light 2
	Accent1           *XColorVariant `xml:"accent1"`  // Accent 1
	Accent2           *XColorVariant `xml:"accent2"`  // Accent 2
	Accent3           *XColorVariant `xml:"accent3"`  // Accent 3
	Accent4           *XColorVariant `xml:"accent4"`  // Accent 4
	Accent5           *XColorVariant `xml:"accent5"`  // Accent 5
	Accent6           *XColorVariant `xml:"accent6"`  // Accent 6
	Hyperlink         *XColorVariant `xml:"hlink"`    // Hyperlink
	FollowedHyperlink *XColorVariant `xml:"folHlink"` // Followed hyperlink
}

// XColorVariant holds a color that may be either an RGB or a system color.
type XColorVariant struct {
	SRGBColor *XSRGBColor `xml:"srgbClr,omitempty"`
	SysColor  *XSysColor  `xml:"sysClr,omitempty"`
}

// XSRGBColor holds an RGB color value.
type XSRGBColor struct {
	Val string `xml:"val,attr"` // 6-digit hex RGB value (e.g. "FF0000")
}

// XSysColor holds a system color value.
type XSysColor struct {
	Val     string `xml:"val,attr"`               // system color name
	LastClr string `xml:"lastClr,attr,omitempty"` // last-used RGB value (fallback color)
}

// ColorType enumerates color types.
type ColorType int

const (
	ColorTypeUnknown ColorType = iota
	ColorTypeRGB               // RGB color
	ColorTypeSystem            // system color
)

// ColorRole enumerates the roles of colors within a color scheme.
type ColorRole int

const (
	ColorRoleDark1             ColorRole = iota // Dark 1 (typically black)
	ColorRoleLight1                             // Light 1 (typically white)
	ColorRoleDark2                              // Dark 2
	ColorRoleLight2                             // Light 2
	ColorRoleAccent1                            // Accent 1
	ColorRoleAccent2                            // Accent 2
	ColorRoleAccent3                            // Accent 3
	ColorRoleAccent4                            // Accent 4
	ColorRoleAccent5                            // Accent 5
	ColorRoleAccent6                            // Accent 6
	ColorRoleHyperlink                          // Hyperlink
	ColorRoleFollowedHyperlink                  // Followed hyperlink
)

// String returns the name of the color role.
func (r ColorRole) String() string {
	names := []string{
		"Dark1",
		"Light1",
		"Dark2",
		"Light2",
		"Accent1",
		"Accent2",
		"Accent3",
		"Accent4",
		"Accent5",
		"Accent6",
		"Hyperlink",
		"FollowedHyperlink",
	}
	if int(r) < len(names) {
		return names[r]
	}
	return "Unknown"
}

// ============================================================================
// Font Scheme
// ============================================================================

// XFontScheme defines the font scheme for a presentation.
type XFontScheme struct {
	XMLName   struct{}         `xml:"fontScheme"`
	XmlnsA    string           `xml:"xmlns:a,attr,omitempty"`
	Name      string           `xml:"name,attr,omitempty"`
	MajorFont *XFontCollection `xml:"majorFont,omitempty"` // heading font
	MinorFont *XFontCollection `xml:"minorFont,omitempty"` // body font
}

// XFontCollection holds a font collection. The Latin/EastAsia/Complex faces
// are nested elements (<latin typeface="…"/>), not attributes of the
// collection — modelling them as a string attribute (the prior shape) emitted
// invalid OOXML that PowerPoint ignored.
type XFontCollection struct {
	Latin    *XFontFace    `xml:"latin,omitempty"` // Latin font face
	EastAsia *XFontFace    `xml:"ea,omitempty"`    // East Asian font face
	Complex  *XFontFace    `xml:"cs,omitempty"`    // complex-script font face
	Fonts    []XScriptFont `xml:"font"`            // script-specific fonts
}

// XFontFace is a single font face: a typeface name plus an optional panose
// classification, preserved across a round-trip.
type XFontFace struct {
	Typeface string `xml:"typeface,attr"`
	Panose   string `xml:"panose,attr,omitempty"`
}

// XScriptFont holds a script-specific font definition.
type XScriptFont struct {
	Script   string `xml:"script,attr"`   // script code (e.g. "Jpan", "Hans")
	Typeface string `xml:"typeface,attr"` // font name
}

// ============================================================================
// Format Scheme
// ============================================================================

// XFmtScheme defines the format scheme.
// Contains style lists for fills, lines, effects, and background fills.
type XFmtScheme struct {
	XMLName        struct{}          `xml:"fmtScheme"`
	XmlnsA         string            `xml:"xmlns:a,attr,omitempty"`
	Name           string            `xml:"name,attr,omitempty"`
	FillStyleLst   *XFillStyleList   `xml:"fillStyleLst,omitempty"`
	LnStyleLst     *XLineStyleList   `xml:"lnStyleLst,omitempty"`
	EffectStyleLst *XEffectStyleList `xml:"effectStyleLst,omitempty"`
	BgFillStyleLst *XFillStyleList   `xml:"bgFillStyleLst,omitempty"`
}

// XFillStyleList holds a fill style list.
type XFillStyleList struct {
	InnerXML string `xml:",innerxml"` // preserve raw XML content to avoid data loss
}

// XLineStyleList holds a line style list.
type XLineStyleList struct {
	InnerXML string `xml:",innerxml"` // preserve raw XML content
}

// XEffectStyleList holds an effect style list.
type XEffectStyleList struct {
	InnerXML string `xml:",innerxml"` // preserve raw XML content
}
