package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/render"
)

// Icons (RFC §14.1). A curated or caller icon is a single-path SVG translated to
// native custom path geometry (<a:custGeom>) and filled with a theme token. The
// translator (internal/render) enforces the documented SVG subset; the OOXML
// wire types stay isolated (P3) — AddIcon takes SVG bytes and returns an opaque
// *Shape.

// ValidateIcon reports whether svg satisfies the icon translator constraints
// (single path, solid fill, no gradients, no elliptical arcs), without drawing
// it. It is the registration-time check curated assets and scene.WithIconExtension
// use to fail fast (D-005).
func ValidateIcon(svg []byte) error {
	if _, err := render.Translate(svg); err != nil {
		return fmt.Errorf("pptx: invalid icon SVG: %w", err)
	}
	return nil
}

// AddIcon adds a single-path SVG glyph as a native custom-geometry shape,
// positioned by box. By default it fills with the accent token (P2); pass
// WithFill to override the color, or WithLine to add an outline. It errors if
// the SVG violates the translator constraints (the same check as ValidateIcon).
func (s *Slide) AddIcon(svg []byte, box Box, opts ...ShapeOption) (*Shape, error) {
	geom, err := render.Translate(svg)
	if err != nil {
		return nil, fmt.Errorf("pptx: AddIcon: %w", err)
	}

	var cfg shapeConfig
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}

	sp := s.builder.AddCustomShape(int(box.X), int(box.Y), int(box.W), int(box.H), geom)

	theme := s.activeTheme()
	fill := cfg.fill
	if fill == nil {
		fill = SolidFill(TokenColor(ColorAccent)) // the icon's documented default (RFC §14.1)
	}
	fill.applyFill(sp.ShapeProperties, theme)
	cfg.line.apply(sp.ShapeProperties, theme)

	return &Shape{s: s, sp: sp}, nil
}
