package pptx

import "strings"

// Color is a write-time-resolved color: either a literal RGB value or a theme
// token. A token resolves against the active Theme when the color is applied,
// so the same builder input re-renders in a new palette after a theme swap
// (P2; D-012, D-030). Construct colors with RGB / RGBA (literals) or
// TokenColor / TokenTextColor (tokens).
//
// The interface is sealed — resolve is unexported — so the only colors are the
// ones this package defines; callers cannot inject a type the codec can't emit.
type Color interface {
	// resolve materializes the color against a theme (which may be nil, in
	// which case DefaultTheme is used).
	resolve(t *Theme) resolvedColor
}

// resolvedColor is a concrete sRGB value plus an OOXML alpha (0..100000).
type resolvedColor struct {
	rgb   RGB
	alpha int
}

// Alpha bounds (OOXML range 0..100000; 100000 = fully opaque).
const (
	AlphaOpaque      = 100000
	AlphaTransparent = 0
)

// RGB — the 6-hex color type from the theme model — is itself a literal Color,
// so pptx.RGB("2563EB") is both a value and a fill color.
func (c RGB) resolve(*Theme) resolvedColor {
	return resolvedColor{rgb: normalizeHex(c), alpha: AlphaOpaque}
}

// literalColor is a literal RGB with an explicit alpha.
type literalColor struct {
	rgb   RGB
	alpha int
}

func (c literalColor) resolve(*Theme) resolvedColor {
	return resolvedColor{rgb: normalizeHex(c.rgb), alpha: clampAlpha(c.alpha)}
}

// RGBA returns a literal color with the given OOXML alpha (0..100000).
func RGBA(hex RGB, alpha int) Color {
	return literalColor{rgb: hex, alpha: alpha}
}

// surfaceToken is a surface color role resolved against the active theme, with
// an OOXML alpha (AlphaOpaque unless set via TokenColorAlpha).
type surfaceToken struct {
	role  ColorRole
	alpha int
}

func (s surfaceToken) resolve(t *Theme) resolvedColor {
	return resolvedColor{rgb: themeOr(t).ResolveColor(s.role), alpha: clampAlpha(s.alpha)}
}

// TokenColor returns a color bound to a semantic surface role (e.g.
// ColorAccent). It resolves to the active theme's value at apply time — swap
// the theme and the same token yields the new palette's color.
func TokenColor(role ColorRole) Color { return surfaceToken{role: role, alpha: AlphaOpaque} }

// TokenColorAlpha returns a token color (TokenColor) at the given OOXML alpha
// (0..100000) — the token analogue of RGBA. It keeps the value token-bound (P2)
// while letting a caller dim it, e.g. a Decoration's opacity or a gradient
// glow's transparent edge.
func TokenColorAlpha(role ColorRole, alpha int) Color {
	return surfaceToken{role: role, alpha: clampAlpha(alpha)}
}

// textToken is a text color role resolved against the active theme.
type textToken struct{ role TextColorRole }

func (t2 textToken) resolve(t *Theme) resolvedColor {
	return resolvedColor{rgb: themeOr(t).ResolveTextColor(t2.role), alpha: AlphaOpaque}
}

// TokenTextColor returns a color bound to a semantic text role (e.g.
// TextPrimary), resolved against the active theme at apply time.
func TokenTextColor(role TextColorRole) Color { return textToken{role: role} }

// themeOr returns t, or DefaultTheme when t is nil.
func themeOr(t *Theme) *Theme {
	if t == nil {
		return DefaultTheme()
	}
	return t
}

// normalizeHex strips a leading '#' and upper-cases a hex color.
func normalizeHex(c RGB) RGB {
	return RGB(strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(string(c)), "#")))
}

// clampAlpha bounds an alpha value to the OOXML range.
func clampAlpha(a int) int {
	switch {
	case a < AlphaTransparent:
		return AlphaTransparent
	case a > AlphaOpaque:
		return AlphaOpaque
	default:
		return a
	}
}
