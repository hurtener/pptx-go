package scene

import "github.com/hurtener/pptx-go/pptx"

// BackgroundKind selects a slide's full-bleed background fill. The zero value
// (BackgroundNone) draws nothing — the slide inherits the presentation's default
// background — preserving byte-identical output for all slides that do not set a
// background (RFC §10.1 backward-compatibility guarantee).
type BackgroundKind int

const (
	// BackgroundNone draws no explicit background; the slide inherits the
	// presentation's default. This is the zero value and pre-Phase-13 behavior.
	// For VariantDark slides a dark canvas rect is drawn automatically when this
	// is set (see renderBackground).
	BackgroundNone BackgroundKind = iota

	// BackgroundColor fills the entire slide canvas with a single solid color
	// resolved from the active theme via Background.Color.
	BackgroundColor

	// BackgroundGradient fills the slide canvas with a two-stop linear gradient
	// between the two roles in Background.Gradient at Background.Angle degrees.
	BackgroundGradient

	// BackgroundAsset fills the slide canvas with a full-bleed picture resolved
	// via Background.AssetID from the render's AssetResolver.
	BackgroundAsset
)

// String returns the background kind's name.
func (k BackgroundKind) String() string {
	switch k {
	case BackgroundColor:
		return "color"
	case BackgroundGradient:
		return "gradient"
	case BackgroundAsset:
		return "asset"
	default:
		return "none"
	}
}

// Background is a slide's full-bleed background specification. It is drawn
// before all body content and decorations — behind the bg-decoration layer
// and behind the body stack — so it forms the lowest layer in the slide's
// z-order. The zero value (Kind == BackgroundNone) draws nothing; all existing
// slides are byte-identical after this field is added to SceneSlide.
//
// Color is the surface color role used when Kind == BackgroundColor; it resolves
// against the active theme so a theme swap re-paints the background (P2).
//
// Gradient is a pair of ColorRole values; index 0 is at position 0 (the start
// of the gradient, Pos 0.0) and index 1 is at position 1 (the end, Pos 1.0).
// Angle is measured in degrees clockwise from the positive x-axis (0° =
// left-to-right, 90° = top-to-bottom).
//
// AssetID is the asset reference passed to the render's AssetResolver when
// Kind == BackgroundAsset. A missing resolver or an unresolvable ID records a
// LayoutWarning and skips the fill; the slide renders without a background
// rather than failing (RFC §10.2 — no panics, degrade to warning).
type Background struct {
	// Kind selects the fill type; zero (BackgroundNone) draws nothing.
	Kind BackgroundKind

	// Color is the surface color role for a solid-color background
	// (Kind == BackgroundColor). Resolves against the active theme.
	Color pptx.ColorRole

	// Gradient holds the two surface color roles for a linear gradient
	// (Kind == BackgroundGradient). Index 0 is the start stop (Pos 0.0),
	// index 1 is the end stop (Pos 1.0). Both resolve against the active theme.
	Gradient [2]pptx.ColorRole

	// Angle is the linear gradient angle in degrees clockwise from the positive
	// x-axis (used when Kind == BackgroundGradient). 0° = left-to-right,
	// 90° = top-to-bottom. Valid range [0, 360); values outside are normalized.
	Angle int

	// AssetID is the asset reference for a full-bleed picture background
	// (Kind == BackgroundAsset). Resolved via the render's AssetResolver.
	AssetID AssetID
}

// ─── VariantDark pinned palette ──────────────────────────────────────────────
//
// These six hex values are hard-coded (not derived from the base theme) so that
// a VariantDark slide produces deterministic, byte-identical output across
// renders regardless of the presentation's concrete palette (RFC §10.1).
//
// Palette: Tailwind CSS gray scale — WCAG 2.1 AA at typical slide font sizes.
const (
	// darkCanvas is the slide canvas for a dark-variant slide (#111827, gray-900).
	darkCanvas = pptx.RGB("111827")
	// darkSurface is the card/panel surface (#1F2937, gray-800).
	darkSurface = pptx.RGB("1F2937")
	// darkSurfaceAlt is the subtle alternate surface (#374151, gray-700).
	darkSurfaceAlt = pptx.RGB("374151")
	// darkTextPrimary is primary body/heading text (#F9FAFB, gray-50).
	darkTextPrimary = pptx.RGB("F9FAFB")
	// darkTextSecondary is secondary/caption text (#E5E7EB, gray-200).
	darkTextSecondary = pptx.RGB("E5E7EB")
	// darkTextTertiary is muted/label text (#9CA3AF, gray-400).
	darkTextTertiary = pptx.RGB("9CA3AF")
)

// darkThemeFrom derives a VariantDark theme from base by cloning it and
// overriding the surface and body-text palette with the pinned dark values.
// Accent and semantic color roles (error, warning, success, info, accent*) are
// preserved from the base theme so brand identity survives the swap. TextInverse
// is also preserved — it remains white for text placed on accent-colored shapes
// (e.g. SectionDivider, Arrow), where it is still correct.
func darkThemeFrom(base *pptx.Theme) *pptx.Theme {
	dark := base.Clone()
	dark.Colors.Surfaces[pptx.ColorCanvas] = darkCanvas
	dark.Colors.Surfaces[pptx.ColorSurface] = darkSurface
	dark.Colors.Surfaces[pptx.ColorSurfaceAlt] = darkSurfaceAlt
	dark.Colors.Text[pptx.TextPrimary] = darkTextPrimary
	dark.Colors.Text[pptx.TextSecondary] = darkTextSecondary
	dark.Colors.Text[pptx.TextTertiary] = darkTextTertiary
	return dark
}
