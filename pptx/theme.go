package pptx

// The Theme is the single source of visual truth at write time (P2, D-003).
// It maps semantic tokens — color, text color, typography, spacing, radius,
// elevation — to concrete OOXML values. Builder calls (Phase 03) take tokens;
// the resolver (tokenresolve.go) materializes the value against the active
// theme. This file defines the token taxonomy (RFC §7.1) and the Theme model.

// RGB is a 6-hex-digit color string without a leading '#', e.g. "2563EB".
type RGB string

// ColorRole is a semantic page-level surface color (RFC §7.1).
type ColorRole int

const (
	ColorCanvas ColorRole = iota
	ColorSurface
	ColorSurfaceAlt
	ColorAccent
	ColorAccentAlt
	ColorAccentWarm
	ColorSuccess
	ColorWarning
	ColorError
	ColorInfo
	// ColorPaper is a faintly tinted off-white "paper" canvas, distinct from
	// pure white, for a designed background tone (D-104). It is appended last so
	// every prior ColorRole value is unchanged. It defaults to ColorCanvas's
	// value (white) so a deck is byte-identical until a theme overrides the tint.
	// ColorPaper has no theme1.xml slot — like TextMuted it keeps its default on
	// read-back (RFC §7.3); its resolved background RGB still round-trips.
	ColorPaper
)

// TextColorRole is a semantic text color for inline runs (RFC §7.1).
type TextColorRole int

const (
	TextPrimary TextColorRole = iota
	TextSecondary
	TextTertiary
	TextInverse
	TextMuted
	TextAccent
	TextAccentAlt
	TextSuccess
	TextWarning
	TextError
)

// TypeRole is a step on the typography scale (RFC §7.1).
type TypeRole int

const (
	TypeDisplay TypeRole = iota
	TypeH1
	TypeH2
	TypeH3
	TypeH4
	TypeH5
	TypeBody
	TypeBodySmall
	TypeCaption
	TypeMono
	TypeCode
)

// SpaceRole is a step on the spacing scale; resolves to EMU (RFC §7.1).
type SpaceRole int

const (
	SpaceXS SpaceRole = iota
	SpaceSM
	SpaceMD
	SpaceLG
	SpaceXL
	Space2XL
)

// RadiusRole is a corner-radius step; resolves to EMU (RFC §7.1).
type RadiusRole int

const (
	RadiusNone RadiusRole = iota
	RadiusSM
	RadiusMD
	RadiusLG
	RadiusFull
)

// ElevationRole is a shadow/elevation step (RFC §7.1).
type ElevationRole int

const (
	ElevationFlat ElevationRole = iota
	ElevationRaised
	ElevationElevated
)

// FontSpec is a resolved typography value: a font family, size in points,
// weight (100–900, 400 = regular, 700 = bold), italic flag, and letter-spacing.
type FontSpec struct {
	Family string
	Size   float64
	Weight int
	Italic bool
	// Tracking is letter-spacing in points (signed): positive opens glyphs apart
	// (wide-tracked eyebrows/labels), negative tightens them (display headlines).
	// 0 (the zero value) emits nothing — byte-identical to an untracked run.
	// Emitted as the OOXML a:rPr/@spc attribute (1/100 pt). (D-060.)
	Tracking float64
	// LineHeight is the role's line spacing as a percent of single (100 = single,
	// 120 = 1.2×); tight display sets ~100–105, body ~120–135. 0 (the zero value)
	// and 100 emit nothing — byte-identical. The scene renderer applies it to a
	// node's paragraphs; emitted as OOXML a:pPr/a:lnSpc/a:spcPct. (D-061.)
	LineHeight float64
	// Case is the role's case transform (e.g. CaseUpper for tracked-caps
	// eyebrows). It is rendered via the OOXML a:rPr/@cap attribute, so the run
	// text stays original-case (round-trips) and PowerPoint/the rasterizer caps
	// it at display. CaseNone (the zero value) emits nothing — byte-identical.
	// (D-062.)
	Case TextCase
	// AvgCharWidth is the role face's average glyph advance as a fraction of the
	// font size, used only by the deterministic wrap/overflow estimator (it never
	// renders). A soul sets a measured factor for its bundled face (serif/display
	// faces advance differently from the default sans). 0 (the zero value) uses
	// the built-in ~0.5 sans fallback — byte-identical estimate. (D-064.)
	AvgCharWidth float64
	// Fallback is an ordered list of substitute families for this role. When a
	// FontSource is registered and it cannot resolve Family, the write-time
	// fallback pass emits the first Fallback entry the source can resolve (instead
	// of letting the host pick an arbitrary default), so output degrades to a
	// controlled near-match. Empty (the zero value) means no fallback —
	// byte-identical. The chosen face is recorded as the run's a:latin typeface
	// (OOXML run fonts are single-valued). (D-066.)
	Fallback []string
}

// TextCase is a type role's case transform, rendered as the OOXML a:rPr/@cap
// attribute (the run text is preserved; the display is cased). The zero value
// CaseNone leaves text exactly as authored.
type TextCase int

const (
	CaseNone      TextCase = iota // as authored (no cap attribute)
	CaseUpper                     // all caps — a:rPr cap="all"
	CaseSmallCaps                 // small caps — a:rPr cap="small"
)

// capAttr returns the OOXML cap attribute value, or "" for CaseNone.
func (c TextCase) capAttr() string {
	switch c {
	case CaseUpper:
		return "all"
	case CaseSmallCaps:
		return "small"
	default:
		return ""
	}
}

// textCaseFromCap is capAttr's read inverse.
func textCaseFromCap(v string) TextCase {
	switch v {
	case "all":
		return CaseUpper
	case "small":
		return CaseSmallCaps
	default:
		return CaseNone
	}
}

// Bold reports whether the weight is bold (≥600).
func (f FontSpec) Bold() bool { return f.Weight >= 600 }

// Elevation is a resolved shadow specification. A zero Elevation (Blur and
// offsets all zero) is "flat" — no shadow.
type Elevation struct {
	Blur    EMU // blur radius
	OffsetX EMU
	OffsetY EMU
	Color   RGB
	Alpha   int // 0–100000 (OOXML alpha), 0 = transparent
}

// IsFlat reports whether the elevation renders no shadow.
func (e Elevation) IsFlat() bool { return e.Blur == 0 && e.OffsetX == 0 && e.OffsetY == 0 }

// ColorPalette maps the surface and text color roles to concrete RGB values.
type ColorPalette struct {
	Surfaces map[ColorRole]RGB
	Text     map[TextColorRole]RGB
}

// DarkPalette is a theme's optional VariantDark color override set (R8.3). The
// scene renderer's dark-variant derivation writes its pinned neutral-gray
// default first, then overlays these surface/text roles when DarkColors is
// non-nil — so a brand renders its own deep dark side (e.g. navy) instead of
// the pinned gray. A nil DarkColors (the zero value) keeps the pinned gray
// default, byte-identical. It mirrors ColorPalette so any role can be
// overridden; the same maps carry the dark accent/extension overrides (R8.7).
//
// DarkPalette has no theme1.xml slot — it is consumed only by the scene
// renderer to derive the VariantDark theme and is never serialized. The
// resolved dark RGB a slide renders with round-trips (and is reported via the
// render Stats.Colors hook, D-058); the field itself does not (like ColorPaper,
// D-104).
type DarkPalette struct {
	Surfaces map[ColorRole]RGB     // VariantDark surface overrides
	Text     map[TextColorRole]RGB // VariantDark text overrides
}

// Typography maps each type role to a resolved FontSpec.
type Typography map[TypeRole]FontSpec

// Spacing maps each spacing role to an EMU value.
type Spacing map[SpaceRole]EMU

// Radii maps each radius role to an EMU value.
type Radii map[RadiusRole]EMU

// Elevations maps each elevation role to a resolved Elevation.
type Elevations map[ElevationRole]Elevation

// Theme is the semantic visual contract. HeadingFont/BodyFont are the
// theme1.xml major/minor font-scheme faces; the Typography map may override
// the family per type role.
type Theme struct {
	Name        string
	HeadingFont string
	// DisplayFont, when non-empty, is the family for the TypeDisplay role (the big
	// editorial face), independent of HeadingFont. Empty (the zero value) makes
	// TypeDisplay inherit HeadingFont — byte-identical to a 2-font theme. (D-063.)
	DisplayFont string
	BodyFont    string
	Colors      ColorPalette
	// Accents, when non-empty, is an ordered brand-accent palette (R8.4): the
	// scene renderer's per-element accent cycle (timeline / funnel / cycle /
	// quadrant / tree / image-pin markers) rotates through these literal hues by
	// index instead of the engine's pinned five-role cycle. Empty (the zero
	// value) keeps the pinned cycle, byte-identical. Like DarkColors it has no
	// theme1.xml slot — the resolved accent RGB round-trips, the field does not.
	Accents []RGB
	// DarkColors, when non-nil, supplies soul-driven VariantDark overrides that
	// the scene renderer overlays over its pinned neutral-gray dark default
	// (R8.3). nil (the zero value) keeps the pinned gray, byte-identical. The
	// field has no theme1.xml slot — see DarkPalette.
	DarkColors *DarkPalette
	// Gradients holds named brand gradients (R8.5), each requested by a scene
	// Background's GradientName. nil/empty means no named gradients (the legacy
	// role-based gradient path is byte-identical). Like DarkColors it has no
	// theme1.xml slot — the resolved gradient fill round-trips, the map does not.
	Gradients  map[string]GradientSpec
	Typography Typography
	Spacing    Spacing
	Radii      Radii
	Elevations Elevations
}

// ThemeOption customizes a Theme built with NewTheme.
type ThemeOption func(*Theme)

// WithName sets the theme name.
func WithName(name string) ThemeOption { return func(t *Theme) { t.Name = name } }

// WithAccent overrides the accent surface color.
func WithAccent(c RGB) ThemeOption {
	return func(t *Theme) { t.Colors.Surfaces[ColorAccent] = c }
}

// WithPaper overrides the ColorPaper surface tint — the faintly tinted off-white
// "paper" canvas (D-104). Pass a low-chroma off-white (e.g. RGB("FAFAF8")) to
// give content slides a designed paper tone; the default is white (= ColorCanvas).
func WithPaper(c RGB) ThemeOption {
	return func(t *Theme) { t.Colors.Surfaces[ColorPaper] = c }
}

// WithAccents sets the theme's ordered brand-accent palette (R8.4): the scene
// renderer's per-element accent cycle rotates through these literal hues by
// index instead of the pinned five-role cycle, so a brand can render 4+
// coordinated accent hues across timeline phases, quadrant points, tree nodes,
// etc. Passing none (or an empty slice) keeps the pinned cycle (byte-identical).
func WithAccents(palette ...RGB) ThemeOption {
	return func(t *Theme) {
		if len(palette) == 0 {
			return
		}
		t.Accents = append([]RGB(nil), palette...)
	}
}

// WithGradient registers a named brand gradient on the theme (R8.5). A scene
// Background requests it by name (Background.GradientName) and the renderer feeds
// it to pptx.LinearGradient / RadialGradient per the spec's Radial flag. Calling
// it with the same name twice replaces the earlier spec. Themes that register no
// gradients are byte-identical to the legacy role-based gradient path.
func WithGradient(name string, spec GradientSpec) ThemeOption {
	return func(t *Theme) {
		if t.Gradients == nil {
			t.Gradients = map[string]GradientSpec{}
		}
		t.Gradients[name] = spec
	}
}

// Gradient returns the named brand gradient (R8.5) and whether it is registered.
func (t *Theme) Gradient(name string) (GradientSpec, bool) {
	spec, ok := t.Gradients[name]
	return spec, ok
}

// WithDarkSurface sets a soul-driven VariantDark override for a surface role
// (R8.3). It lazily allocates Theme.DarkColors; the scene renderer overlays the
// override over its pinned neutral-gray dark default, so a brand renders its own
// dark canvas/surface (e.g. deep navy). Setting none leaves the pinned gray
// default (byte-identical). Composable with WithDarkText and order-independent.
func WithDarkSurface(role ColorRole, c RGB) ThemeOption {
	return func(t *Theme) {
		ensureDarkColors(t).Surfaces[role] = c
	}
}

// WithDarkText sets a soul-driven VariantDark override for a text role (R8.3).
// It lazily allocates Theme.DarkColors; the scene renderer overlays the override
// over its pinned dark-text default. Setting none leaves the pinned default
// (byte-identical). Composable with WithDarkSurface and order-independent.
func WithDarkText(role TextColorRole, c RGB) ThemeOption {
	return func(t *Theme) {
		ensureDarkColors(t).Text[role] = c
	}
}

// ensureDarkColors lazily allocates t.DarkColors (and its maps) and returns it.
func ensureDarkColors(t *Theme) *DarkPalette {
	if t.DarkColors == nil {
		t.DarkColors = &DarkPalette{
			Surfaces: map[ColorRole]RGB{},
			Text:     map[TextColorRole]RGB{},
		}
	}
	return t.DarkColors
}

// WithFonts overrides the heading and body font families (and updates the
// Typography families to match).
func WithFonts(heading, body string) ThemeOption {
	return func(t *Theme) {
		t.HeadingFont, t.BodyFont = heading, body
		for role, spec := range t.Typography {
			switch {
			case role == TypeDisplay && t.DisplayFont != "":
				spec.Family = t.DisplayFont // a distinct display face wins for TypeDisplay
			case role <= TypeH5:
				spec.Family = heading
			case role == TypeMono || role == TypeCode:
				// mono untouched
			default:
				spec.Family = body
			}
			t.Typography[role] = spec
		}
	}
}

// WithDisplayFont sets a distinct display face for the TypeDisplay role (the big
// editorial face), independent of the heading face (D-063). Order-independent
// with WithFonts. Omitting it leaves TypeDisplay on HeadingFont (byte-identical).
func WithDisplayFont(family string) ThemeOption {
	return func(t *Theme) {
		t.DisplayFont = family
		if spec, ok := t.Typography[TypeDisplay]; ok {
			spec.Family = family
			t.Typography[TypeDisplay] = spec
		}
	}
}

// NewTheme returns a copy of the default theme with the options applied.
func NewTheme(opts ...ThemeOption) *Theme {
	t := DefaultTheme()
	for _, o := range opts {
		o(t)
	}
	return t
}

// DefaultTheme returns the V1 default theme: a light surface, a neutral
// palette, and a system font stack (Calibri / Calibri Light / Consolas) that
// renders every node legibly with no embedding (RFC §7.5). The returned
// theme is a fresh copy — callers may mutate it freely.
func DefaultTheme() *Theme {
	const (
		heading = "Calibri Light"
		body    = "Calibri"
		mono    = "Consolas"
	)
	return &Theme{
		Name:        "pptx-go default",
		HeadingFont: heading,
		BodyFont:    body,
		Colors: ColorPalette{
			Surfaces: map[ColorRole]RGB{
				ColorCanvas:     "FFFFFF",
				ColorSurface:    "FFFFFF",
				ColorSurfaceAlt: "F1F3F5",
				ColorAccent:     "2563EB",
				ColorAccentAlt:  "7C3AED",
				ColorAccentWarm: "EA580C",
				ColorSuccess:    "16A34A",
				ColorWarning:    "D97706",
				ColorError:      "DC2626",
				ColorInfo:       "0EA5E9",
				ColorPaper:      "FFFFFF", // = ColorCanvas; settable to an off-white paper tint (D-104)
			},
			Text: map[TextColorRole]RGB{
				TextPrimary:   "111827",
				TextSecondary: "374151",
				TextTertiary:  "6B7280",
				TextInverse:   "FFFFFF",
				TextMuted:     "9CA3AF",
				TextAccent:    "2563EB",
				TextAccentAlt: "7C3AED",
				TextSuccess:   "16A34A",
				TextWarning:   "D97706",
				TextError:     "DC2626",
			},
		},
		Typography: Typography{
			TypeDisplay:   {Family: heading, Size: 40, Weight: 700},
			TypeH1:        {Family: heading, Size: 32, Weight: 700},
			TypeH2:        {Family: heading, Size: 28, Weight: 600},
			TypeH3:        {Family: heading, Size: 24, Weight: 600},
			TypeH4:        {Family: heading, Size: 20, Weight: 600},
			TypeH5:        {Family: heading, Size: 16, Weight: 600},
			TypeBody:      {Family: body, Size: 14, Weight: 400},
			TypeBodySmall: {Family: body, Size: 12, Weight: 400},
			TypeCaption:   {Family: body, Size: 10, Weight: 400},
			TypeMono:      {Family: mono, Size: 13, Weight: 400},
			TypeCode:      {Family: mono, Size: 12, Weight: 400},
		},
		Spacing: Spacing{
			SpaceXS:  Pt(2),
			SpaceSM:  Pt(4),
			SpaceMD:  Pt(8),
			SpaceLG:  Pt(16),
			SpaceXL:  Pt(24),
			Space2XL: Pt(40),
		},
		Radii: Radii{
			RadiusNone: 0,
			RadiusSM:   Pt(2),
			RadiusMD:   Pt(6),
			RadiusLG:   Pt(12),
			RadiusFull: Pt(7200), // effectively pill-shaped at slide scale
		},
		Elevations: Elevations{
			ElevationFlat:     {},
			ElevationRaised:   {Blur: Pt(4), OffsetY: Pt(1), Color: "000000", Alpha: 25000},
			ElevationElevated: {Blur: Pt(12), OffsetY: Pt(4), Color: "000000", Alpha: 35000},
		},
	}
}

// Clone returns a deep copy of the theme so callers can mutate without
// affecting the original (themes are reusable artifacts — CLAUDE.md §5).
func (t *Theme) Clone() *Theme {
	c := *t
	c.Colors.Surfaces = make(map[ColorRole]RGB, len(t.Colors.Surfaces))
	for k, v := range t.Colors.Surfaces {
		c.Colors.Surfaces[k] = v
	}
	c.Colors.Text = make(map[TextColorRole]RGB, len(t.Colors.Text))
	for k, v := range t.Colors.Text {
		c.Colors.Text[k] = v
	}
	// Accents is an ordered slice (R8.4) — copy it so a clone's palette can be
	// mutated without aliasing the original (themes are reusable, §5). A nil/empty
	// slice stays nil (byte-identical fallback preserved).
	if len(t.Accents) > 0 {
		c.Accents = append([]RGB(nil), t.Accents...)
	}
	// DarkColors is an optional pointer (R8.3) — deep-copy both maps when present
	// so a clone's dark palette can be mutated without aliasing the original
	// (themes are reusable artifacts — CLAUDE.md §5). nil stays nil (the shallow
	// c := *t already copied the nil pointer).
	if t.DarkColors != nil {
		dc := &DarkPalette{
			Surfaces: make(map[ColorRole]RGB, len(t.DarkColors.Surfaces)),
			Text:     make(map[TextColorRole]RGB, len(t.DarkColors.Text)),
		}
		for k, v := range t.DarkColors.Surfaces {
			dc.Surfaces[k] = v
		}
		for k, v := range t.DarkColors.Text {
			dc.Text[k] = v
		}
		c.DarkColors = dc
	}
	// Gradients is an optional map of named brand gradients (R8.5) — deep-copy the
	// map and each spec's stop slice so a clone's gradients can be mutated without
	// aliasing the original. A nil/empty map stays nil (byte-identical fallback).
	if len(t.Gradients) > 0 {
		g := make(map[string]GradientSpec, len(t.Gradients))
		for name, spec := range t.Gradients {
			spec.Stops = append([]GradientStop(nil), spec.Stops...)
			g[name] = spec
		}
		c.Gradients = g
	}
	c.Typography = make(Typography, len(t.Typography))
	for k, v := range t.Typography {
		c.Typography[k] = v
	}
	c.Spacing = make(Spacing, len(t.Spacing))
	for k, v := range t.Spacing {
		c.Spacing[k] = v
	}
	c.Radii = make(Radii, len(t.Radii))
	for k, v := range t.Radii {
		c.Radii[k] = v
	}
	c.Elevations = make(Elevations, len(t.Elevations))
	for k, v := range t.Elevations {
		c.Elevations[k] = v
	}
	return &c
}
