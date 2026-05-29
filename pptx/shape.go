package pptx

import "github.com/hurtener/pptx-go/internal/ooxml/slide"

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
	fill Fill
	line Line
}

// ShapeOption configures a shape at creation time.
type ShapeOption func(*shapeConfig)

// WithFill sets the shape's interior fill (SolidFill, NoFill, …).
func WithFill(f Fill) ShapeOption { return func(c *shapeConfig) { c.fill = f } }

// WithLine sets the shape's outline.
func WithLine(l Line) ShapeOption { return func(c *shapeConfig) { c.line = l } }

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

	return &Shape{sp: sp}
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
