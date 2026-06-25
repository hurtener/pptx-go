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

// Shape is a handle to a shape on a slide — either one the builder added or one
// recovered by Open from a reopened deck (Slide.Shapes, RFC §16). It does not
// expose the underlying OOXML wire type (P3); read accessors (Geometry,
// Rotation, Fill, Line, Shadow) map the recovered wire fields back to the public
// builder types so a reopened shape compares field-equal to the authored one.
//
// Exactly one of the underlying children is set: an auto-shape (sp), a picture
// (pic), or a graphic frame / table (gf). The builder constructs only sp shapes;
// Shapes() wraps whichever child it finds so the read accessors land on a common
// handle (text/table/image read arrive in later read chunks).
type Shape struct {
	s   *Slide // owning slide (read side: resolves hyperlink relationships)
	sp  *slide.XSp
	pic *slide.XPicture
	gf  *slide.XGraphicFrame
}

// props returns the shape's <spPr>, regardless of the underlying child kind, or
// nil when the child carries none (a graphic frame has its transform inline, not
// under spPr).
func (sh *Shape) props() *slide.XShapeProperties {
	switch {
	case sh.sp != nil:
		return sh.sp.ShapeProperties
	case sh.pic != nil:
		return sh.pic.ShapeProperties
	default:
		return nil
	}
}

// xfrm returns the shape's 2D transform: spPr/xfrm for an auto-shape or picture,
// or the graphic frame's own xfrm.
func (sh *Shape) xfrm() *slide.XTransform2D {
	if sh.gf != nil {
		return sh.gf.Transform2D
	}
	if p := sh.props(); p != nil {
		return p.Transform2D
	}
	return nil
}

// shapeConfig accumulates AddShape options.
type shapeConfig struct {
	fill       Fill
	line       Line
	radius     *RadiusRole
	rotation   *float64       // degrees clockwise; nil = unset
	shadow     *Elevation     // literal drop shadow; nil = unset
	shadowRole *ElevationRole // token drop shadow; resolved at AddShape; wins over shadow
	imageFill  ImageSource    // cover-fit image surface fill; nil = unset (wins over fill)
}

// ShapeOption configures a shape at creation time.
type ShapeOption func(*shapeConfig)

// WithFill sets the shape's interior fill (SolidFill, NoFill, …).
func WithFill(f Fill) ShapeOption { return func(c *shapeConfig) { c.fill = f } }

// WithLine sets the shape's outline.
func WithLine(l Line) ShapeOption { return func(c *shapeConfig) { c.line = l } }

// WithImageFill fills the shape's interior with an image (a `<a:blipFill>`)
// instead of a solid/gradient fill — a photo-as-surface for cards and panels
// (R14.1). The image is cover-fit: scaled to fill the shape and center-cropped on
// the overflowing axis (computed from the image's format-header dimensions —
// §7/D-046, not pixel data — so there is no distortion at any aspect; an
// unreadable header falls back to a plain stretch). It wins over WithFill. The
// shape's geometry/corner-radius still clips the fill, so an image-filled
// roundRect keeps its rounded corners. A nil source or an unreadable image leaves
// the prior fill unchanged.
func WithImageFill(src ImageSource) ShapeOption {
	return func(c *shapeConfig) { c.imageFill = src }
}

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
	// Image fill (R14.1): replaces any solid/gradient fill with a cover-fit
	// <a:blipFill>. The part is registered here (the option layer has no slide
	// handle); an unreadable source leaves the prior fill in place.
	if cfg.imageFill != nil {
		if img, err := cfg.imageFill.resolveImage(); err == nil {
			rID := s.addImagePart(img.bytes, img.ext)
			sp.ShapeProperties.SolidFill = nil
			sp.ShapeProperties.GradientFill = nil
			sp.ShapeProperties.NoFill = nil
			sp.ShapeProperties.BlipFill = &slide.XBlipFillProperties{
				Blip:    &slide.XBlip{Embed: rID},
				SrcRect: coverSrcRect(img.bytes, box),
				Stretch: &slide.XStretchProperties{FillRect: &slide.XFillRectProperties{}},
			}
		}
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

	return &Shape{s: s, sp: sp}
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
	if sh == nil {
		return Box{}
	}
	xf := sh.xfrm()
	if xf == nil {
		return Box{}
	}
	var b Box
	if xf.Offset != nil {
		b.X, b.Y = EMU(xf.Offset.X), EMU(xf.Offset.Y)
	}
	if xf.Extent != nil {
		b.W, b.H = EMU(xf.Extent.Cx), EMU(xf.Extent.Cy)
	}
	return b
}

// Geometry returns the shape's preset geometry — the OOXML prst name (e.g.
// ShapeRoundRect). It is the empty string for a custom-geometry shape (an icon
// glyph, custGeom) or one with no geometry (a picture or graphic frame). It is
// the read inverse of AddShape's geom argument.
func (sh *Shape) Geometry() ShapeGeometry {
	if sh == nil {
		return ""
	}
	if p := sh.props(); p != nil && p.PresetGeom != nil {
		return ShapeGeometry(p.PresetGeom.Prst)
	}
	return ""
}

// Rotation returns the shape's clockwise rotation in degrees within [0, 360°),
// or 0 if unset — the read inverse of WithRotation.
func (sh *Shape) Rotation() float64 {
	if sh == nil {
		return 0
	}
	if xf := sh.xfrm(); xf != nil {
		return float64(xf.Rotation) / 60000.0
	}
	return 0
}

// Fill returns the shape's interior fill, or nil when the shape has no explicit
// fill (it inherits its style fill). A reopened fill surfaces resolved literal
// colors (D-030); inspect it via Fill.Kind / SolidColor / Gradient. It is the
// read inverse of WithFill.
func (sh *Shape) Fill() Fill {
	if sh == nil {
		return nil
	}
	return fillFromX(sh.props())
}

// Line returns the shape's outline, or a zero Line when the shape has no
// explicit outline (it inherits its style line). It is the read inverse of
// WithLine.
func (sh *Shape) Line() Line {
	if sh == nil {
		return Line{}
	}
	p := sh.props()
	if p == nil || p.Line == nil {
		return Line{}
	}
	x := p.Line
	ln := Line{Width: EMU(x.Width)}
	if x.SolidFill != nil {
		ln.Color = colorFromSrgb(x.SolidFill.SrgbClr)
	}
	if x.PrstDash != nil {
		ln.Dash = x.PrstDash.Val
	}
	return ln
}

// Shadow returns the shape's drop shadow as an Elevation and true, or a zero
// Elevation and false when the shape casts none. The OOXML outerShdw stores the
// offset in polar form (dist/dir); Shadow reconstructs the cartesian
// OffsetX/OffsetY, so an axis-aligned shadow round-trips exactly and an oblique
// one to within sub-EMU rounding (D-035). It is the read inverse of
// WithElevation / WithShadow.
func (sh *Shape) Shadow() (Elevation, bool) {
	if sh == nil {
		return Elevation{}, false
	}
	p := sh.props()
	if p == nil || p.EffectList == nil || p.EffectList.OuterShdw == nil {
		return Elevation{}, false
	}
	o := p.EffectList.OuterShdw
	rad := float64(o.Dir) / 60000.0 * math.Pi / 180
	e := Elevation{
		Blur:    EMU(o.BlurRad),
		OffsetX: EMU(math.Round(float64(o.Dist) * math.Cos(rad))),
		OffsetY: EMU(math.Round(float64(o.Dist) * math.Sin(rad))),
	}
	if o.SrgbClr != nil {
		e.Color = RGB(o.SrgbClr.Val)
		if o.SrgbClr.Alpha != nil {
			e.Alpha = o.SrgbClr.Alpha.Val
		}
	}
	return e, true
}

// TextFrame returns the shape's rich-text container and true, or nil and false
// when the shape carries no text body. On a reopened deck the returned frame
// enumerates paragraphs → runs with their authored style / color / hyperlink /
// bullet (Slide.Shapes, RFC §16). Table-cell text is reached via the table read
// accessors, not here.
func (sh *Shape) TextFrame() (*TextFrame, bool) {
	if sh == nil || sh.sp == nil || sh.sp.TextBody == nil {
		return nil, false
	}
	return &TextFrame{s: sh.s, body: sh.sp.TextBody}, true
}

// Table returns the table this shape bears and true, or nil and false when the
// shape is not a table (a graphic frame carrying an <a:tbl>). On a reopened deck
// the returned handle exposes the authored row/column counts, column widths,
// header/banding intent, and per-cell text / fill / merge via the Table and Cell
// read accessors (RFC §16).
func (sh *Shape) Table() (*Table, bool) {
	if sh == nil || sh.gf == nil || sh.gf.Graphic == nil ||
		sh.gf.Graphic.GraphicData == nil || sh.gf.Graphic.GraphicData.Table == nil {
		return nil, false
	}
	tbl := sh.gf.Graphic.GraphicData.Table
	cols := 0
	if tbl.Grid != nil {
		cols = len(tbl.Grid.GridCols)
	}
	t := &Table{slide: sh.s, gf: sh.gf, rows: len(tbl.Rows), cols: cols}
	if tbl.Pr != nil {
		t.headerOn = tbl.Pr.FirstRow == "1"
		t.bandRowOn = tbl.Pr.BandRow == "1"
	}
	return t, true
}

// Image returns the image this shape bears and true, or nil and false when the
// shape is not a picture. On a reopened deck the returned handle exposes the
// authored alt text / crop / fit / rotation / opacity and resolves the embedded
// bytes via Image.Bytes (RFC §16).
func (sh *Shape) Image() (*Image, bool) {
	if sh == nil || sh.pic == nil {
		return nil, false
	}
	return &Image{s: sh.s, pic: sh.pic}, true
}

// activeTheme returns the presentation's theme, or DefaultTheme if unavailable.
func (s *Slide) activeTheme() *Theme {
	if s.presentation != nil {
		return s.presentation.Theme()
	}
	return DefaultTheme()
}
