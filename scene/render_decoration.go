package scene

import (
	"errors"
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/ornaments"
)

// Decoration composition (RFC §14.2, §11.1, §12). A preset decoration renders as
// native shapes via its ornament recipe; an asset decoration renders as a pic.
// The placement box is derived from anchor + offset + size + bleed; opacity maps
// to an OOXML alpha the recipe applies to the accent token. Layer z-order is
// imposed by the renderer's layout() (background before body, foreground after).
// No product behavior (D-026): an unresolved asset or an unregistered ornament
// degrades to a LayoutWarning, never a panic.

// defaultDecorationSize is the ornament box used when Decoration.Size is zero.
var defaultDecorationSize = pptx.Size{W: pptx.In(2), H: pptx.In(2)}

func (r *renderer) renderDecoration(ps *pptx.Slide, region pptx.Box, v Decoration, slideID string) {
	box := decorationBox(region, v)
	if !v.Bleed && offCanvas(region, box) {
		r.warn(slideID, "decoration extends past the slide edge; set Bleed to allow it")
	}
	alpha := opacityAlpha(v.Opacity)

	switch v.Kind {
	case DecorationPreset:
		recipe, ok := r.cfg.ornaments.Lookup(v.Preset)
		if !ok { // Stage-1 validation rejects this before render; degrade defensively.
			r.warn(slideID, fmt.Sprintf("decoration ornament %q not registered; skipped", v.Preset))
			return
		}
		r.stats.Shapes += recipe(ps, box, alpha, v.Rotation)
	case DecorationAsset:
		data, ct, err := r.resolve(v.AssetID)
		if err != nil {
			r.warn(slideID, fmt.Sprintf("decoration asset %q unresolved: %v", v.AssetID, err))
			return
		}
		img, aerr := ps.AddImage(pptx.ImageBytes(data, ct), box)
		if aerr != nil {
			r.warn(slideID, fmt.Sprintf("decoration image %q: %v", v.AssetID, aerr))
			return
		}
		if v.Rotation != 0 {
			img.SetRotation(v.Rotation)
		}
		if alpha != pptx.AlphaOpaque {
			img.SetOpacity(alpha)
		}
		r.stats.Shapes++
		r.stats.Assets++
	}
}

// decorationBox positions the ornament box so its anchor-corresponding point
// (the box's top-left for a top-left anchor, its center for a center anchor,
// etc.) lands on the region's anchor point, shifted by Offset. Size defaults
// when zero.
func decorationBox(region pptx.Box, v Decoration) pptx.Box {
	size := v.Size
	if size.W <= 0 || size.H <= 0 {
		size = defaultDecorationSize
	}
	target := v.Anchor.Point(region)
	// within is the anchor's offset inside a box of this size at the origin, so
	// subtracting it aligns the box's matching corner/edge/center to the target.
	within := v.Anchor.Point(pptx.Box{W: size.W, H: size.H})
	return pptx.Box{
		X: target.X + v.Offset.X - within.X,
		Y: target.Y + v.Offset.Y - within.Y,
		W: size.W,
		H: size.H,
	}
}

// offCanvas reports whether box extends past the region's bounds.
func offCanvas(region, box pptx.Box) bool {
	return box.X < region.X || box.Y < region.Y ||
		box.Right() > region.Right() || box.Bottom() > region.Bottom()
}

// opacityAlpha maps a 0..1 opacity to an OOXML alpha; 0 (the zero value) is
// treated as fully opaque (the default).
func opacityAlpha(opacity float64) int {
	if opacity <= 0 {
		return pptx.AlphaOpaque
	}
	if opacity >= 1 {
		return pptx.AlphaOpaque
	}
	return int(opacity * float64(pptx.AlphaOpaque))
}

// validateOrnamentRefs is the registry-aware half of Stage-1 validation (RFC
// §14.2/§14.4): every preset Decoration's name must resolve in the render's
// ornament registry. It runs in Render (the registry derives from options) and
// recurses container children.
func validateOrnamentRefs(s Scene, reg *ornaments.Registry) error {
	var errs []error
	for i := range s.Slides {
		sl := &s.Slides[i]
		where := sl.ID
		if where == "" {
			where = fmt.Sprintf("#%d", i)
		}
		walkDecorations(sl.Nodes, func(d Decoration) {
			if d.Kind != DecorationPreset {
				return
			}
			if _, ok := reg.Lookup(d.Preset); !ok {
				errs = append(errs, fmt.Errorf(
					"slide %s: decoration ornament %q is not a curated or registered ornament (have %v)",
					where, d.Preset, reg.Names()))
			}
		})
	}
	return errors.Join(errs...)
}

// walkDecorations visits every Decoration in a node tree, recursing into
// container children.
func walkDecorations(nodes []SlideNode, fn func(Decoration)) {
	for _, n := range nodes {
		switch v := n.(type) {
		case Decoration:
			fn(v)
		case TwoColumn:
			walkDecorations(v.Left, fn)
			walkDecorations(v.Right, fn)
		case Grid:
			walkDecorations(v.Cells, fn)
		case Card:
			walkDecorations(v.Body, fn)
		case CardSection:
			walkDecorations(v.Body, fn)
		}
	}
}
