package pptx

import (
	"fmt"
	"math"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// ShapeGeometry is a preset shape outline, expressed as its OOXML preset
// geometry name (ST_ShapeType). Use the Shape* constants.
type ShapeGeometry string

// Preset geometries (a curated subset of ST_ShapeType; the value is the OOXML
// prst attribute).
const (
	ShapeRect          ShapeGeometry = "rect"
	ShapeRoundRect     ShapeGeometry = "roundRect"
	ShapeEllipse       ShapeGeometry = "ellipse"
	ShapeTriangle      ShapeGeometry = "triangle"
	ShapeDiamond       ShapeGeometry = "diamond"
	ShapeParallelogram ShapeGeometry = "parallelogram"
	ShapeHexagon       ShapeGeometry = "hexagon"
	ShapeChevron       ShapeGeometry = "chevron"
	ShapeRightArrow    ShapeGeometry = "rightArrow"
	ShapeLine          ShapeGeometry = "line"
)

// Shape is an opaque handle to a shape that was added to a slide. It does not
// expose the underlying OOXML wire type (P3); it exists so callers can hold a
// reference for future, type-safe mutators.
type Shape struct {
	sp *slide.XSp
}

// shapeConfig accumulates AddShape options.
type shapeConfig struct {
	fill       Fill
	line       Line
	radius     *RadiusRole
	rotation   *float64       // degrees clockwise; nil = unset
	shadow     *Elevation     // literal drop shadow; nil = unset
	shadowRole *ElevationRole // token drop shadow; resolved at AddShape; wins over shadow
}

// ShapeOption configures a shape at creation time.
type ShapeOption func(*shapeConfig)

// WithFill sets the shape's interior fill (SolidFill, NoFill, …).
func WithFill(f Fill) ShapeOption { return func(c *shapeConfig) { c.fill = f } }

// WithLine sets the shape's outline.
func WithLine(l Line) ShapeOption { return func(c *shapeConfig) { c.line = l } }

// WithRadius sets a rounded-corner radius from a theme radius token (P2). It
// applies to ShapeRoundRect only — the corner radius is OOXML's roundRect adjust
// handle — and is ignored for other geometries. The token resolves against the
// active theme at AddShape time, so a theme swap re-rounds the same input;
// RadiusFull yields a full capsule (pill).
func WithRadius(role RadiusRole) ShapeOption {
	return func(c *shapeConfig) { c.radius = &role }
}

// WithRotation rotates the shape clockwise by deg degrees about its centre
// (OOXML <a:xfrm rot>, D-041). The angle is normalized to [0, 360°).
func WithRotation(deg float64) ShapeOption {
	return func(c *shapeConfig) { c.rotation = &deg }
}

// WithElevation casts a drop shadow from the active theme's Elevation token for
// role (the documented token path — P2, D-043). The token resolves at AddShape
// time, so a theme swap re-renders the shadow in the brand's elevation.
// ElevationFlat (and any flat token) emits no effect — byte-identical to a shape
// with no shadow.
func WithElevation(role ElevationRole) ShapeOption {
	return func(c *shapeConfig) {
		r := role
		c.shadowRole = &r
		c.shadow = nil
	}
}

// WithShadow casts a drop shadow from a literal Elevation (the escape hatch;
// the documented path is WithElevation). A flat Elevation (IsFlat) emits no
// effect.
func WithShadow(e Elevation) ShapeOption {
	return func(c *shapeConfig) {
		c.shadow = &e
		c.shadowRole = nil
	}
}

// AddShape adds a preset-geometry shape positioned by box (EMU) and returns a
// handle to it. Fills and lines are resolved against the presentation's active
// theme at this point, so a theme token reflects the theme in force now — the
// mechanism behind theme swaps (P2). This is the token-aware shape API (RFC
// §8.2/§8.3); the older Add* helpers remain for convenience.
func (s *Slide) AddShape(geom ShapeGeometry, box Box, opts ...ShapeOption) *Shape {
	var cfg shapeConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}

	sp := s.builder.AddAutoShape(int(box.X), int(box.Y), int(box.W), int(box.H), string(geom))
	if sp.ShapeProperties == nil {
		sp.ShapeProperties = &slide.XShapeProperties{}
	}

	theme := s.activeTheme()
	if cfg.fill != nil {
		cfg.fill.applyFill(sp.ShapeProperties, theme)
	}
	cfg.line.apply(sp.ShapeProperties, theme)

	if cfg.radius != nil && geom == ShapeRoundRect {
		applyCornerRadius(sp.ShapeProperties, theme.ResolveRadius(*cfg.radius), box)
	}

	if cfg.rotation != nil && sp.ShapeProperties.Transform2D != nil {
		sp.ShapeProperties.Transform2D.Rotation = normalizeAngle60k(*cfg.rotation)
	}

	// Drop shadow (D-043): token role resolves against the active theme; a literal
	// Elevation is the escape hatch. A flat elevation emits no effect, keeping a
	// no-shadow shape byte-identical.
	switch {
	case cfg.shadowRole != nil:
		applyShadow(sp.ShapeProperties, theme.ResolveElevation(*cfg.shadowRole))
	case cfg.shadow != nil:
		applyShadow(sp.ShapeProperties, *cfg.shadow)
	}

	return &Shape{sp: sp}
}

// applyShadow attaches an <a:effectLst><a:outerShdw> realizing e. A flat
// elevation is a no-op (no effect list), so it does not perturb existing output.
// The Theme's cartesian OffsetX/OffsetY become outerShdw's polar dist/dir,
// rounded to integers so the serialized bytes are deterministic (D-035).
func applyShadow(spPr *slide.XShapeProperties, e Elevation) {
	if spPr == nil || e.IsFlat() {
		return
	}
	dist := int(math.Round(math.Hypot(float64(e.OffsetX), float64(e.OffsetY))))
	dir := 0
	if e.OffsetX != 0 || e.OffsetY != 0 {
		deg := math.Atan2(float64(e.OffsetY), float64(e.OffsetX)) * 180 / math.Pi
		dir = normalizeAngle60k(deg)
	}
	spPr.EffectList = &slide.XEffectList{
		OuterShdw: &slide.XOuterShadow{
			BlurRad:      int(e.Blur),
			Dist:         dist,
			Dir:          dir,
			RotWithShape: 0,
			SrgbClr:      &slide.XSrgbClr{Val: string(e.Color), Alpha: &slide.XAlpha{Val: e.Alpha}},
		},
	}
}

// applyCornerRadius sets a roundRect's corner radius via its adjust guide. The
// theme radius token is an absolute EMU length, but OOXML's roundRect adjust is
// a fraction of the shorter side (×100000, so 50000 = 50% = a full capsule), so
// the absolute radius is converted against the shape box and clamped.
func applyCornerRadius(spPr *slide.XShapeProperties, radius EMU, box Box) {
	if spPr == nil || spPr.PresetGeom == nil {
		return
	}
	minDim := box.W
	if box.H < minDim {
		minDim = box.H
	}
	if minDim <= 0 {
		return
	}
	adj := int(math.Round(float64(radius) / float64(minDim) * 100000))
	if adj < 0 {
		adj = 0
	}
	if adj > 50000 {
		adj = 50000 // 50% of the shorter side = fully rounded (capsule)
	}
	spPr.PresetGeom.AvLst = &slide.XAvLst{
		Gd: []slide.XShapeGuide{{Name: "adj", Fmla: fmt.Sprintf("val %d", adj)}},
	}
}

// Box returns the shape's position and size in EMU.
func (sh *Shape) Box() Box {
	if sh == nil || sh.sp == nil || sh.sp.ShapeProperties == nil || sh.sp.ShapeProperties.Transform2D == nil {
		return Box{}
	}
	xf := sh.sp.ShapeProperties.Transform2D
	var b Box
	if xf.Offset != nil {
		b.X, b.Y = EMU(xf.Offset.X), EMU(xf.Offset.Y)
	}
	if xf.Extent != nil {
		b.W, b.H = EMU(xf.Extent.Cx), EMU(xf.Extent.Cy)
	}
	return b
}

// activeTheme returns the presentation's theme, or DefaultTheme if unavailable.
func (s *Slide) activeTheme() *Theme {
	if s.presentation != nil {
		return s.presentation.Theme()
	}
	return DefaultTheme()
}
