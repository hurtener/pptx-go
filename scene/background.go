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

	// BackgroundRadial fills the slide canvas with a center-out radial gradient
	// (a spotlight/vignette) from Background.Stops, or the legacy two-role
	// Background.Gradient pair when Stops is empty. The focal point is centered
	// (a 50%-inset circle); a focal offset is not yet exposed (D-106). Appended
	// last so existing BackgroundKind values are unchanged (byte-identical).
	BackgroundRadial

	// BackgroundMesh draws a soft "mesh" wash: a base canvas fill plus the N
	// low-alpha radial glows in Background.Mesh, pooled at caller-chosen anchors
	// over the canvas (the cover/section mesh look — D-112). An empty Mesh draws
	// nothing (absent config). Appended last so existing values are unchanged.
	BackgroundMesh
)

// MeshGlow is one pooled radial glow in a BackgroundMesh (D-112): a soft circle
// of light at Anchor, of the surface role Color, radius Radius (EMU), fading from
// the center alpha Alpha (OOXML 0..100000) to transparent at the edge.
type MeshGlow struct {
	// Anchor is where the glow pools on the slide (its center).
	Anchor Anchor
	// Color is the glow's surface color role (resolved against the active theme).
	Color pptx.ColorRole
	// Radius is the glow circle's radius in EMU; a non-positive radius is skipped.
	Radius pptx.EMU
	// Alpha is the glow center's OOXML opacity (0..100000); keep it low for a
	// subtle pool. The edge fades to fully transparent.
	Alpha int
}

// Scrim is an optional darkening (or tinting) overlay drawn over a slide's
// background fill so text reads legibly over a photographic or busy background
// (R14.1). It is a general mechanism: the engine draws it; the caller (soul)
// chooses the color and opacity that meet its contrast target (D-026). A nil
// Background.Scrim draws nothing (byte-identical).
//
// Color is the overlay's surface color role (resolved against the active theme).
// The zero value (ColorCanvas) is a real color (white), so set it deliberately —
// a darkening scrim typically uses a dark surface or a literal-backed role.
//
// Opacity is the overlay's OOXML opacity (0..100000); 0 draws an invisible
// overlay. For a solid scrim the whole overlay carries Opacity; for a gradient
// scrim the dense edge carries Opacity and the opposite edge is transparent.
//
// Gradient, when true, draws a linear gradient scrim (transparent → Color at
// Opacity) instead of a flat wash — the classic bottom-heavy caption scrim.
// GradientAngle (degrees clockwise from the positive x-axis; 0° = left-to-right,
// 90° = top-to-bottom) orients it; the zero value defaults to 90° (top
// transparent, bottom dense). Gradient is ignored for a solid scrim.
type Scrim struct {
	// Color is the overlay's surface color role.
	Color pptx.ColorRole
	// Opacity is the dense edge's OOXML opacity (0..100000).
	Opacity int
	// Gradient selects a transparent→Color linear gradient overlay when true,
	// else a flat solid wash at Opacity.
	Gradient bool
	// GradientAngle orients a gradient scrim in degrees; zero defaults to 90°.
	GradientAngle int
}

// Duotone is an optional two-tone recolor applied to a photographic background
// (R14.1): the photo's shadows map to Shadow and its highlights to Highlight,
// producing an on-brand tint. Both are surface color roles resolved against the
// active theme (P2), so a theme swap re-tints the photo. A nil Background.Duotone
// leaves the photo at its natural colors (byte-identical). Applies only when
// Kind == BackgroundAsset.
type Duotone struct {
	// Shadow is the role the photo's dark tones map to.
	Shadow pptx.ColorRole
	// Highlight is the role the photo's light tones map to.
	Highlight pptx.ColorRole
}

// GradientStop is one color stop in a multi-stop background gradient (D-105).
// Pos is the stop position in [0,1] (0 = start, 1 = end); Color is a surface
// role resolved against the active theme so a theme swap re-paints the stop (P2).
type GradientStop struct {
	// Pos is the stop position along the gradient axis, in [0,1].
	Pos float64
	// Color is the surface color role at this stop.
	Color pptx.ColorRole
}

// String returns the background kind's name.
func (k BackgroundKind) String() string {
	switch k {
	case BackgroundColor:
		return "color"
	case BackgroundGradient:
		return "gradient"
	case BackgroundAsset:
		return "asset"
	case BackgroundRadial:
		return "radial"
	case BackgroundMesh:
		return "mesh"
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
//
// Stops is an optional multi-stop gradient (D-105). When non-empty it
// supersedes Gradient for a BackgroundGradient; when empty the two-role
// Gradient + Angle path runs unchanged (byte-identical to pre-D-105 output).
type Background struct {
	// Kind selects the fill type; zero (BackgroundNone) draws nothing.
	Kind BackgroundKind

	// Color is the surface color role for a solid-color background
	// (Kind == BackgroundColor). Resolves against the active theme.
	Color pptx.ColorRole

	// Gradient holds the two surface color roles for a linear gradient
	// (Kind == BackgroundGradient) when Stops is empty. Index 0 is the start
	// stop (Pos 0.0), index 1 is the end stop (Pos 1.0). Both resolve against
	// the active theme.
	Gradient [2]pptx.ColorRole

	// Stops is an optional multi-stop gradient (2..8 ascending stops in [0,1])
	// for Kind == BackgroundGradient. When non-empty it supersedes Gradient (the
	// legacy two-role pair); when empty, Gradient + Angle drive a two-stop linear
	// gradient (byte-identical to pre-D-105 output). Invalid stops (<2, >8, out
	// of [0,1], or not strictly ascending) record a LayoutWarning and skip the
	// fill (RFC §10.2 — degrade to a warning, no panic). The slice makes
	// Background non-comparable; compare with reflect.DeepEqual.
	Stops []GradientStop

	// Angle is the linear gradient angle in degrees clockwise from the positive
	// x-axis (used when Kind == BackgroundGradient). 0° = left-to-right,
	// 90° = top-to-bottom. Valid range [0, 360); values outside are normalized.
	Angle int

	// AssetID is the asset reference for a full-bleed picture background
	// (Kind == BackgroundAsset). Resolved via the render's AssetResolver.
	AssetID AssetID

	// Mesh holds the pooled radial glows for a BackgroundMesh (D-112), drawn over
	// the base canvas fill in slice order. Empty draws nothing (absent config).
	// Adds no comparability constraint beyond the Stops slice above.
	Mesh []MeshGlow

	// Scrim is an optional darkening/tinting overlay drawn over the background
	// fill for text legibility over a photo or busy background (R14.1). nil draws
	// nothing (byte-identical). It applies over any drawn background kind.
	Scrim *Scrim

	// Duotone is an optional two-tone recolor of a photographic background
	// (R14.1), applied only when Kind == BackgroundAsset. nil leaves the photo at
	// its natural colors (byte-identical).
	Duotone *Duotone
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
