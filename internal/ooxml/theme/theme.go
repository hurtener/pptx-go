package theme

import (
	"encoding/xml"
	"fmt"
	"sync"

	"github.com/hurtener/pptx-go/internal/ooxml"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// ThemePart - theme part
// ============================================================================
//
// Corresponds to /ppt/theme/themeN.xml
// Defines the presentation's color scheme, font scheme, and format scheme.
//
// ============================================================================

// ThemePart holds the theme part for a presentation.
type ThemePart struct {
	uri *opc.PackURI

	// theme data
	theme *XTheme

	// cached color lookup table
	colorCache map[ColorRole]*XColorVariant

	mu sync.RWMutex
}

// NewThemePart creates a new theme part with the given numeric ID.
func NewThemePart(id int) *ThemePart {
	return &ThemePart{
		uri:        opc.NewPackURI(fmt.Sprintf("/ppt/theme/theme%d.xml", id)),
		colorCache: make(map[ColorRole]*XColorVariant),
	}
}

// NewThemePartWithURI creates a theme part using the specified URI.
func NewThemePartWithURI(uri *opc.PackURI) *ThemePart {
	return &ThemePart{
		uri:        uri,
		colorCache: make(map[ColorRole]*XColorVariant),
	}
}

// PartURI returns the part URI.
func (t *ThemePart) PartURI() *opc.PackURI {
	return t.uri
}

// Theme returns the theme data.
func (t *ThemePart) Theme() *XTheme {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.theme
}

// SetThemeData sets the theme data (used when assigning a cloned theme).
func (t *ThemePart) SetThemeData(theme *XTheme) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.theme = theme
	// clear the cache
	t.colorCache = make(map[ColorRole]*XColorVariant)
}

// ColorScheme returns the color scheme.
func (t *ThemePart) ColorScheme() *XColorScheme {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.theme == nil || t.theme.ThemeElements == nil {
		return nil
	}
	return t.theme.ThemeElements.ColorScheme
}

// FontScheme returns the font scheme.
func (t *ThemePart) FontScheme() *XFontScheme {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.theme == nil || t.theme.ThemeElements == nil {
		return nil
	}
	return t.theme.ThemeElements.FontScheme
}

// ============================================================================
// Color access methods
// ============================================================================

// GetThemeColor returns the color for the given role.
func (t *ThemePart) GetThemeColor(role ColorRole) *XColorVariant {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// check the cache
	if c, ok := t.colorCache[role]; ok {
		return c
	}

	// look up in the color scheme
	scheme := t.ColorScheme()
	if scheme == nil {
		return nil
	}

	var color *XColorVariant
	switch role {
	case ColorRoleDark1:
		color = scheme.Dark1
	case ColorRoleLight1:
		color = scheme.Light1
	case ColorRoleDark2:
		color = scheme.Dark2
	case ColorRoleLight2:
		color = scheme.Light2
	case ColorRoleAccent1:
		color = scheme.Accent1
	case ColorRoleAccent2:
		color = scheme.Accent2
	case ColorRoleAccent3:
		color = scheme.Accent3
	case ColorRoleAccent4:
		color = scheme.Accent4
	case ColorRoleAccent5:
		color = scheme.Accent5
	case ColorRoleAccent6:
		color = scheme.Accent6
	case ColorRoleHyperlink:
		color = scheme.Hyperlink
	case ColorRoleFollowedHyperlink:
		color = scheme.FollowedHyperlink
	}

	// cache the result
	if color != nil {
		t.colorCache[role] = color
	}

	return color
}

// GetThemeColorRGB returns the RGB value for the given role as a 6-digit hex
// string (e.g. "FF0000"), or an empty string if unavailable.
func (t *ThemePart) GetThemeColorRGB(role ColorRole) string {
	color := t.GetThemeColor(role)
	if color == nil {
		return ""
	}

	if color.SRGBColor != nil {
		return color.SRGBColor.Val
	}

	if color.SysColor != nil && color.SysColor.LastClr != "" {
		return color.SysColor.LastClr
	}

	return ""
}

// GetThemeColorType returns the color type for the given role.
func (t *ThemePart) GetThemeColorType(role ColorRole) ColorType {
	color := t.GetThemeColor(role)
	if color == nil {
		return ColorTypeUnknown
	}

	if color.SRGBColor != nil {
		return ColorTypeRGB
	}

	if color.SysColor != nil {
		return ColorTypeSystem
	}

	return ColorTypeUnknown
}

// ============================================================================
// Color setter methods
// ============================================================================

// SetThemeColorRGB sets the RGB color for the given role.
// rgb must be a 6-digit hex string (e.g. "FF0000").
func (t *ThemePart) SetThemeColorRGB(role ColorRole, rgb string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// ensure the theme structure exists
	t.ensureThemeStructure()

	scheme := t.theme.ThemeElements.ColorScheme

	// create the color variant
	color := &XColorVariant{
		SRGBColor: &XSRGBColor{Val: rgb},
	}

	// assign the color to the appropriate role
	switch role {
	case ColorRoleDark1:
		scheme.Dark1 = color
	case ColorRoleLight1:
		scheme.Light1 = color
	case ColorRoleDark2:
		scheme.Dark2 = color
	case ColorRoleLight2:
		scheme.Light2 = color
	case ColorRoleAccent1:
		scheme.Accent1 = color
	case ColorRoleAccent2:
		scheme.Accent2 = color
	case ColorRoleAccent3:
		scheme.Accent3 = color
	case ColorRoleAccent4:
		scheme.Accent4 = color
	case ColorRoleAccent5:
		scheme.Accent5 = color
	case ColorRoleAccent6:
		scheme.Accent6 = color
	case ColorRoleHyperlink:
		scheme.Hyperlink = color
	case ColorRoleFollowedHyperlink:
		scheme.FollowedHyperlink = color
	}

	// update the cache
	t.colorCache[role] = color
}

// SetThemeColorSystem sets a system color for the given role.
// sysColorName is the system color name (e.g. "windowText", "window").
// lastClr is the fallback RGB value (6-digit hex).
func (t *ThemePart) SetThemeColorSystem(role ColorRole, sysColorName, lastClr string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// ensure the theme structure exists
	t.ensureThemeStructure()

	scheme := t.theme.ThemeElements.ColorScheme

	// create the color variant
	color := &XColorVariant{
		SysColor: &XSysColor{
			Val:     sysColorName,
			LastClr: lastClr,
		},
	}

	// assign the color to the appropriate role
	switch role {
	case ColorRoleDark1:
		scheme.Dark1 = color
	case ColorRoleLight1:
		scheme.Light1 = color
	case ColorRoleDark2:
		scheme.Dark2 = color
	case ColorRoleLight2:
		scheme.Light2 = color
	case ColorRoleAccent1:
		scheme.Accent1 = color
	case ColorRoleAccent2:
		scheme.Accent2 = color
	case ColorRoleAccent3:
		scheme.Accent3 = color
	case ColorRoleAccent4:
		scheme.Accent4 = color
	case ColorRoleAccent5:
		scheme.Accent5 = color
	case ColorRoleAccent6:
		scheme.Accent6 = color
	case ColorRoleHyperlink:
		scheme.Hyperlink = color
	case ColorRoleFollowedHyperlink:
		scheme.FollowedHyperlink = color
	}

	// update the cache
	t.colorCache[role] = color
}

// ensureThemeStructure ensures the theme structure is initialized (must be called with the lock held).
func (t *ThemePart) ensureThemeStructure() {
	if t.theme == nil {
		t.theme = &XTheme{
			XmlnsA: "http://schemas.openxmlformats.org/drawingml/2006/main",
		}
	}
	if t.theme.ThemeElements == nil {
		t.theme.ThemeElements = &XThemeElements{}
	}
	if t.theme.ThemeElements.ColorScheme == nil {
		t.theme.ThemeElements.ColorScheme = &XColorScheme{}
	}
	if t.theme.ThemeElements.FontScheme == nil {
		t.theme.ThemeElements.FontScheme = &XFontScheme{}
	}
}

// ============================================================================
// Font setter methods
// ============================================================================

// SetThemeMajorFont sets the major (heading) font.
func (t *ThemePart) SetThemeMajorFont(latin, eastAsia, complex string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ensureThemeStructure()

	fontScheme := t.theme.ThemeElements.FontScheme
	if fontScheme.MajorFont == nil {
		fontScheme.MajorFont = &XFontCollection{}
	}

	fontScheme.MajorFont.Latin = &XFontFace{Typeface: latin}
	fontScheme.MajorFont.EastAsia = &XFontFace{Typeface: eastAsia}
	fontScheme.MajorFont.Complex = &XFontFace{Typeface: complex}
}

// SetThemeMinorFont sets the minor (body) font.
func (t *ThemePart) SetThemeMinorFont(latin, eastAsia, complex string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ensureThemeStructure()

	fontScheme := t.theme.ThemeElements.FontScheme
	if fontScheme.MinorFont == nil {
		fontScheme.MinorFont = &XFontCollection{}
	}

	fontScheme.MinorFont.Latin = &XFontFace{Typeface: latin}
	fontScheme.MinorFont.EastAsia = &XFontFace{Typeface: eastAsia}
	fontScheme.MinorFont.Complex = &XFontFace{Typeface: complex}
}

// SetThemeScriptFont sets a script-specific font.
// isMajor: true for the major (heading) font collection, false for minor (body).
func (t *ThemePart) SetThemeScriptFont(isMajor bool, script, typeface string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.ensureThemeStructure()

	fontScheme := t.theme.ThemeElements.FontScheme
	var fontColl *XFontCollection
	if isMajor {
		if fontScheme.MajorFont == nil {
			fontScheme.MajorFont = &XFontCollection{}
		}
		fontColl = fontScheme.MajorFont
	} else {
		if fontScheme.MinorFont == nil {
			fontScheme.MinorFont = &XFontCollection{}
		}
		fontColl = fontScheme.MinorFont
	}

	// update existing script font if found
	for i, f := range fontColl.Fonts {
		if f.Script == script {
			fontColl.Fonts[i].Typeface = typeface
			return
		}
	}

	// otherwise append a new entry
	fontColl.Fonts = append(fontColl.Fonts, XScriptFont{
		Script:   script,
		Typeface: typeface,
	})
}

// ============================================================================
// XML serialization / deserialization
// ============================================================================

// ToXML serializes the theme to XML.
func (t *ThemePart) ToXML() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.theme == nil {
		return nil, fmt.Errorf("theme data is nil")
	}

	output, err := xml.MarshalIndent(t.theme, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(ooxml.XMLDeclaration), output...), nil
}

// FromXML deserializes the theme from XML.
func (t *ThemePart) FromXML(data []byte) error {
	// strip namespace prefixes for compatibility with Go's xml.Unmarshal
	cleanData, err := ooxml.StripNamespacePrefixes(data)
	if err != nil {
		return fmt.Errorf("failed to clean XML: %w", err)
	}

	var theme XTheme
	if err := xml.Unmarshal(cleanData, &theme); err != nil {
		return fmt.Errorf("failed to unmarshal theme XML: %w", err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.theme = &theme

	// clear the cache
	t.colorCache = make(map[ColorRole]*XColorVariant)

	return nil
}

// ParseTheme parses a theme from XML bytes.
func ParseTheme(data []byte) (*XTheme, error) {
	cleanData, err := ooxml.StripNamespacePrefixes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to clean XML: %w", err)
	}

	var theme XTheme
	if err := xml.Unmarshal(cleanData, &theme); err != nil {
		return nil, fmt.Errorf("failed to parse theme: %w", err)
	}

	return &theme, nil
}

// ============================================================================
// XColorVariant helper methods
// ============================================================================

// Type returns the color type.
func (c *XColorVariant) Type() ColorType {
	if c == nil {
		return ColorTypeUnknown
	}
	if c.SRGBColor != nil {
		return ColorTypeRGB
	}
	if c.SysColor != nil {
		return ColorTypeSystem
	}
	return ColorTypeUnknown
}

// RGB returns the RGB color value.
// For RGB colors the value is returned directly; for system colors the
// LastClr fallback is returned.
func (c *XColorVariant) RGB() string {
	if c == nil {
		return ""
	}

	if c.SRGBColor != nil {
		return c.SRGBColor.Val
	}

	if c.SysColor != nil {
		return c.SysColor.LastClr
	}

	return ""
}

// IsRGB reports whether this is an RGB color.
func (c *XColorVariant) IsRGB() bool {
	return c != nil && c.SRGBColor != nil
}

// IsSystem reports whether this is a system color.
func (c *XColorVariant) IsSystem() bool {
	return c != nil && c.SysColor != nil
}

// SystemColorName returns the system color name.
func (c *XColorVariant) SystemColorName() string {
	if c == nil || c.SysColor == nil {
		return ""
	}
	return c.SysColor.Val
}
