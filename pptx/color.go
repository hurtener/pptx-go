// Package pptx provides a high-level interface for working with PPTX files.
package pptx

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================================
// Color system - color mapping and validation
// ============================================================================
//
// OOXML color specification:
//  1. Output format must be 6-digit hex without a # prefix (e.g. "FF0000").
//  2. Opacity uses a separate tag: <a:alpha val="50000"/>; range 0-100000.
//  3. Unified input format: #RRGGBBAA (8 digits; last two digits are alpha).
//     - #FF0000FF = 100% opaque red
//     - #FF000000 = 0% fully transparent red
//
// PowerPoint supports two color representations:
//  1. RGB color (srgbClr) - specified directly, e.g. "FF0000" for red.
//  2. Theme color (schemeClr) - references a theme color, e.g. "accent1".
//
// ============================================================================

// Alpha constants (OOXML range: 0-100000).
const (
	// AlphaOpaque is fully opaque.
	AlphaOpaque = 100000
	// AlphaTransparent is fully transparent.
	AlphaTransparent = 0
	// AlphaDefault is the default opacity (100%).
	AlphaDefault = 100000
)

// ColorType identifies the kind of color value.
type ColorType int

const (
	// ColorTypeRGB is a direct RGB color.
	ColorTypeRGB ColorType = iota
	// ColorTypeScheme is a theme (scheme) color.
	ColorTypeScheme
	// ColorTypeInvalid represents an invalid or unparseable color.
	ColorTypeInvalid
)

// Color holds a parsed color value.
type Color struct {
	// Type is the color kind.
	Type ColorType
	// RGB is the 6-digit hex value (e.g. "FF0000", no # prefix).
	RGB string
	// Scheme is the theme color name (e.g. "accent1").
	Scheme string
	// Alpha is the opacity in OOXML units (0-100000).
	// 100000 = fully opaque, 0 = fully transparent.
	Alpha int
	// IsValid reports whether the color was successfully parsed.
	IsValid bool
}

// ============================================================================
// Preset color constants
// ============================================================================

// Preset colors (6-digit hex, no # prefix).
var (
	// Basic colors.
	ColorBlack   = RGBColor("000000")
	ColorWhite   = RGBColor("FFFFFF")
	ColorRed     = RGBColor("FF0000")
	ColorGreen   = RGBColor("00FF00")
	ColorBlue    = RGBColor("0000FF")
	ColorYellow  = RGBColor("FFFF00")
	ColorCyan    = RGBColor("00FFFF")
	ColorMagenta = RGBColor("FF00FF")

	// Transparent color.
	ColorTransparent = Color{Type: ColorTypeRGB, RGB: "000000", Alpha: AlphaTransparent, IsValid: true}

	// Common UI colors.
	ColorGray      = RGBColor("808080")
	ColorLightGray = RGBColor("C0C0C0")
	ColorDarkGray  = RGBColor("404040")
	ColorOrange    = RGBColor("FFA500")
	ColorPurple    = RGBColor("800080")
	ColorPink      = RGBColor("FFC0CB")
	ColorBrown     = RGBColor("A52A2A")
	ColorNavy      = RGBColor("000080")
	ColorTeal      = RGBColor("008080")
	ColorOlive     = RGBColor("808000")
	ColorMaroon    = RGBColor("800000")
	ColorLime      = RGBColor("00FF00")
	ColorAqua      = RGBColor("00FFFF")
	ColorSilver    = RGBColor("C0C0C0")
	ColorGold      = RGBColor("FFD700")
)

// ============================================================================
// Theme color constants
// ============================================================================

// Theme color name constants.
const (
	// Background colors.
	SchemeBg1 = "bg1" // background color 1 (typically white or light)
	SchemeBg2 = "bg2" // background color 2
	SchemeFg1 = "fg1" // foreground/text color 1 (typically black or dark)
	SchemeFg2 = "fg2" // foreground/text color 2

	// Accent colors.
	SchemeAccent1 = "accent1" // accent color 1
	SchemeAccent2 = "accent2" // accent color 2
	SchemeAccent3 = "accent3" // accent color 3
	SchemeAccent4 = "accent4" // accent color 4
	SchemeAccent5 = "accent5" // accent color 5
	SchemeAccent6 = "accent6" // accent color 6

	// Hyperlink colors.
	SchemeHlink    = "hlink"    // hyperlink color
	SchemeFolHlink = "folHlink" // followed hyperlink color

	// Special colors.
	SchemePhClr = "phClr" // slide title color (placeholder color)
	SchemeTx1   = "tx1"   // text color 1
	SchemeTx2   = "tx2"   // text color 2
)

// SchemeColors is the list of all valid theme color names.
var SchemeColors = []string{
	SchemeBg1, SchemeBg2, SchemeFg1, SchemeFg2,
	SchemeAccent1, SchemeAccent2, SchemeAccent3, SchemeAccent4, SchemeAccent5, SchemeAccent6,
	SchemeHlink, SchemeFolHlink,
	SchemeTx1, SchemeTx2,
}

// ============================================================================
// Color parsing
// ============================================================================

// hexColorRegex matches 6- or 8-digit hex colors (with or without #).
var hexColorRegex = regexp.MustCompile(`^#?([0-9A-Fa-f]{6})([0-9A-Fa-f]{2})?$`)

// hexColor3Regex matches 3-digit hex colors (with or without #).
var hexColor3Regex = regexp.MustCompile(`^#?([0-9A-Fa-f]{3})$`)

// rgbColorRegex matches rgb(r, g, b) notation.
var rgbColorRegex = regexp.MustCompile(`^rgb\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)$`)

// rgbaColorRegex matches rgba(r, g, b, a) notation.
var rgbaColorRegex = regexp.MustCompile(`^rgba\s*\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*,\s*([\d.]+)\s*\)$`)

// ParseColor parses a color string. Supported formats:
//   - "#FF0000" or "FF0000" (6-digit hex)
//   - "#FF0000FF" (8-digit hex; last two digits are alpha)
//   - "rgb(255, 0, 0)" (RGB)
//   - "rgba(255, 0, 0, 0.5)" (RGBA; alpha 0-1)
//   - "accent1", "bg1", etc. (theme color names)
func ParseColor(s string) Color {
	s = strings.TrimSpace(s)
	if s == "" {
		return Color{Type: ColorTypeInvalid, IsValid: false}
	}

	// Try 8-digit hex (#RRGGBBAA).
	if matches := hexColorRegex.FindStringSubmatch(s); matches != nil {
		hex := strings.ToUpper(matches[1])
		alpha := AlphaDefault
		if matches[2] != "" {
			alpha = hexToAlpha(matches[2])
		}
		return Color{
			Type:    ColorTypeRGB,
			RGB:     hex,
			Alpha:   alpha,
			IsValid: true,
		}
	}

	// Try 3-digit hex.
	if matches := hexColor3Regex.FindStringSubmatch(s); matches != nil {
		hex3 := matches[1]
		hex := strings.ToUpper(string(hex3[0]) + string(hex3[0]) + string(hex3[1]) + string(hex3[1]) + string(hex3[2]) + string(hex3[2]))
		return Color{
			Type:    ColorTypeRGB,
			RGB:     hex,
			Alpha:   AlphaDefault,
			IsValid: true,
		}
	}

	// Try rgba(…).
	if matches := rgbaColorRegex.FindStringSubmatch(s); matches != nil {
		r, _ := strconv.Atoi(matches[1])
		g, _ := strconv.Atoi(matches[2])
		b, _ := strconv.Atoi(matches[3])
		a, _ := strconv.ParseFloat(matches[4], 64)
		alpha := int(a * float64(AlphaOpaque))
		return ColorFromRGBA(r, g, b, alpha)
	}

	// Try rgb(…).
	if matches := rgbColorRegex.FindStringSubmatch(s); matches != nil {
		r, _ := strconv.Atoi(matches[1])
		g, _ := strconv.Atoi(matches[2])
		b, _ := strconv.Atoi(matches[3])
		return ColorFromRGB(r, g, b)
	}

	// Try theme color name.
	if IsSchemeColor(s) {
		return SchemeColor(s)
	}

	return Color{Type: ColorTypeInvalid, IsValid: false}
}

// hexToAlpha converts a 2-digit hex string to an OOXML alpha value (0-100000).
func hexToAlpha(hex string) int {
	val, _ := strconv.ParseInt(hex, 16, 64)
	// 0x00 -> 0 (transparent), 0xFF -> 100000 (opaque)
	return int(float64(val) / 255.0 * float64(AlphaOpaque))
}

// alphaToHex converts an OOXML alpha value to a 2-digit uppercase hex string.
func alphaToHex(alpha int) string {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > AlphaOpaque {
		alpha = AlphaOpaque
	}
	val := int(float64(alpha) / float64(AlphaOpaque) * 255.0)
	return fmt.Sprintf("%02X", val)
}

// ============================================================================
// Color constructors
// ============================================================================

// RGBColor creates an opaque RGB color from a 6-digit hex string.
func RGBColor(hex string) Color {
	hex = strings.ToUpper(strings.TrimSpace(hex))
	hex = strings.TrimPrefix(hex, "#")

	// Validate hex format.
	if len(hex) != 6 {
		return Color{Type: ColorTypeInvalid, IsValid: false}
	}
	if _, err := strconv.ParseInt(hex, 16, 64); err != nil {
		return Color{Type: ColorTypeInvalid, IsValid: false}
	}

	return Color{
		Type:    ColorTypeRGB,
		RGB:     hex,
		Alpha:   AlphaDefault,
		IsValid: true,
	}
}

// RGBAColor creates an RGB color with the given opacity.
// hex is a 6-digit hex string; alpha is in OOXML units (0-100000).
func RGBAColor(hex string, alpha int) Color {
	c := RGBColor(hex)
	if !c.IsValid {
		return c
	}
	c.Alpha = alpha
	if c.Alpha < 0 {
		c.Alpha = 0
	}
	if c.Alpha > AlphaOpaque {
		c.Alpha = AlphaOpaque
	}
	return c
}

// ColorFromRGB creates an opaque color from integer R, G, B components (0-255).
func ColorFromRGB(r, g, b int) Color {
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return Color{Type: ColorTypeInvalid, IsValid: false}
	}
	hex := fmt.Sprintf("%02X%02X%02X", r, g, b)
	return RGBColor(hex)
}

// ColorFromRGBA creates a color from integer R, G, B components (0-255) and an
// OOXML alpha value (0-100000).
func ColorFromRGBA(r, g, b, alpha int) Color {
	c := ColorFromRGB(r, g, b)
	if !c.IsValid {
		return c
	}
	c.Alpha = alpha
	if c.Alpha < 0 {
		c.Alpha = 0
	}
	if c.Alpha > AlphaOpaque {
		c.Alpha = AlphaOpaque
	}
	return c
}

// SchemeColor creates a theme color from a scheme color name.
func SchemeColor(name string) Color {
	name = strings.ToLower(strings.TrimSpace(name))
	if !IsSchemeColor(name) {
		return Color{Type: ColorTypeInvalid, IsValid: false}
	}
	return Color{
		Type:    ColorTypeScheme,
		Scheme:  name,
		Alpha:   AlphaDefault,
		IsValid: true,
	}
}

// IsSchemeColor reports whether name is a valid theme color name.
func IsSchemeColor(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, scheme := range SchemeColors {
		if scheme == name {
			return true
		}
	}
	return false
}

// ============================================================================
// Color output - OOXML-compliant
// ============================================================================

// ToRGB returns the 6-digit uppercase hex string (no # prefix).
// Suitable for use in <a:srgbClr val="FF0000"/>.
func (c Color) ToRGB() string {
	if c.Type == ColorTypeRGB {
		return c.RGB
	}
	return ""
}

// ToHex returns the color as a #-prefixed 6-digit hex string (for display).
func (c Color) ToHex() string {
	if c.Type == ColorTypeRGB {
		return "#" + c.RGB
	}
	return ""
}

// ToHexA returns the color as a #-prefixed 8-digit hex string including alpha.
func (c Color) ToHexA() string {
	if c.Type == ColorTypeRGB {
		return "#" + c.RGB + alphaToHex(c.Alpha)
	}
	return ""
}

// ToScheme returns the theme color name.
func (c Color) ToScheme() string {
	if c.Type == ColorTypeScheme {
		return c.Scheme
	}
	return ""
}

// AlphaValue returns the opacity in OOXML units (0-100000).
// Suitable for use in <a:alpha val="50000"/>.
func (c Color) AlphaValue() int {
	return c.Alpha
}

// AlphaPercent returns the opacity as a percentage (0-100).
func (c Color) AlphaPercent() float64 {
	return float64(c.Alpha) / float64(AlphaOpaque) * 100
}

// String returns a human-readable representation of the color.
func (c Color) String() string {
	switch c.Type {
	case ColorTypeRGB:
		if c.Alpha != AlphaDefault {
			return "#" + c.RGB + alphaToHex(c.Alpha)
		}
		return "#" + c.RGB
	case ColorTypeScheme:
		return c.Scheme
	default:
		return "invalid"
	}
}

// RGBComponents returns the individual R, G, B components (0-255) and whether
// they are available.
func (c Color) RGBComponents() (r, g, b int, ok bool) {
	if c.Type != ColorTypeRGB || len(c.RGB) != 6 {
		return 0, 0, 0, false
	}
	ri, _ := strconv.ParseInt(c.RGB[0:2], 16, 64)
	gi, _ := strconv.ParseInt(c.RGB[2:4], 16, 64)
	bi, _ := strconv.ParseInt(c.RGB[4:6], 16, 64)
	return int(ri), int(gi), int(bi), true
}

// WithAlpha returns a copy of the color with the given OOXML alpha value (0-100000).
func (c Color) WithAlpha(alpha int) Color {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > AlphaOpaque {
		alpha = AlphaOpaque
	}
	c.Alpha = alpha
	return c
}

// WithAlphaPercent returns a copy of the color with the given opacity percentage (0-100).
func (c Color) WithAlphaPercent(percent float64) Color {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	c.Alpha = int(percent / 100.0 * float64(AlphaOpaque))
	return c
}

// ============================================================================
// Color validation
// ============================================================================

// ColorValidationResult holds the result of a color validation check.
type ColorValidationResult struct {
	// IsValid reports whether the color is valid.
	IsValid bool
	// Color is the parsed color.
	Color Color
	// Original is the original input string.
	Original string
	// Message is a human-readable validation message.
	Message string
}

// ValidateColor validates a color string and returns the result.
func ValidateColor(s string) ColorValidationResult {
	color := ParseColor(s)
	return ColorValidationResult{
		IsValid:  color.IsValid,
		Color:    color,
		Original: s,
		Message:  color.validateMessage(),
	}
}

// validateMessage returns a human-readable description of the color's validity.
func (c Color) validateMessage() string {
	if c.IsValid {
		switch c.Type {
		case ColorTypeRGB:
			if c.Alpha != AlphaDefault {
				return fmt.Sprintf("valid RGB color: #%s (alpha: %d/100000)", c.RGB, c.Alpha)
			}
			return fmt.Sprintf("valid RGB color: #%s", c.RGB)
		case ColorTypeScheme:
			return fmt.Sprintf("valid theme color: %s", c.Scheme)
		}
	}
	return "invalid color format"
}

// ============================================================================
// Color map
// ============================================================================

// ColorMap maps color names to Color values.
type ColorMap struct {
	colors map[string]Color
}

// NewColorMap creates an empty ColorMap.
func NewColorMap() *ColorMap {
	return &ColorMap{
		colors: make(map[string]Color),
	}
}

// DefaultColorMap returns a ColorMap pre-populated with common color names.
func DefaultColorMap() *ColorMap {
	cm := NewColorMap()
	cm.Set("black", ColorBlack)
	cm.Set("white", ColorWhite)
	cm.Set("red", ColorRed)
	cm.Set("green", ColorGreen)
	cm.Set("blue", ColorBlue)
	cm.Set("yellow", ColorYellow)
	cm.Set("cyan", ColorCyan)
	cm.Set("magenta", ColorMagenta)
	cm.Set("gray", ColorGray)
	cm.Set("gray", ColorGray)
	cm.Set("lightgray", ColorLightGray)
	cm.Set("lightgrey", ColorLightGray)
	cm.Set("darkgray", ColorDarkGray)
	cm.Set("darkgrey", ColorDarkGray)
	cm.Set("orange", ColorOrange)
	cm.Set("purple", ColorPurple)
	cm.Set("pink", ColorPink)
	cm.Set("brown", ColorBrown)
	cm.Set("navy", ColorNavy)
	cm.Set("teal", ColorTeal)
	cm.Set("olive", ColorOlive)
	cm.Set("maroon", ColorMaroon)
	cm.Set("lime", ColorLime)
	cm.Set("aqua", ColorAqua)
	cm.Set("silver", ColorSilver)
	cm.Set("gold", ColorGold)
	cm.Set("transparent", ColorTransparent)
	return cm
}

// Set adds or updates a color name mapping.
func (cm *ColorMap) Set(name string, color Color) {
	cm.colors[strings.ToLower(name)] = color
}

// Get looks up a color by name.
func (cm *ColorMap) Get(name string) (Color, bool) {
	color, ok := cm.colors[strings.ToLower(name)]
	return color, ok
}

// Resolve resolves a color string by first checking the map, then parsing it.
func (cm *ColorMap) Resolve(s string) Color {
	// Try the map first.
	if color, ok := cm.Get(s); ok {
		return color
	}
	// Fall back to parsing.
	return ParseColor(s)
}

// All returns a copy of all name-to-color mappings.
func (cm *ColorMap) All() map[string]Color {
	result := make(map[string]Color, len(cm.colors))
	for k, v := range cm.colors {
		result[k] = v
	}
	return result
}
