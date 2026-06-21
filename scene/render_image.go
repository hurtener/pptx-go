package scene

import (
	"errors"
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/frames"
)

// Image composition (RFC §14.3). renderImage resolves the asset bytes, wraps
// them in a device frame when the node selects one (the recipe draws the bezel
// and returns the interior region), and inserts the picture into that interior.
//
// Phase 10 scope: framing + placement + alt text. Crop/fit, aspect-aware
// fitting, and the media-manager refactor are Phase 11 — this file is the seam
// Phase 11 extends. No product behavior (D-026): an unresolved asset or an
// unregistered frame degrades to a LayoutWarning, never a panic.
func (r *renderer) renderImage(ps *pptx.Slide, box pptx.Box, v Image, slideID string) {
	interior := box
	if name := resolveFrameName(v); name != "" {
		if recipe, ok := r.cfg.frames.Lookup(name); ok {
			in, n := recipe(ps, box)
			interior = in
			r.stats.Shapes += n
		} else {
			// Stage-1 validation rejects this before render; degrade defensively.
			r.warn(slideID, fmt.Sprintf("image frame %q not registered; rendering unframed", name))
		}
	}

	data, ct, err := r.resolve(v.AssetID)
	if err != nil {
		r.warn(slideID, fmt.Sprintf("image asset %q unresolved: %v", v.AssetID, err))
		return
	}
	img, aerr := ps.AddImage(pptx.ImageBytes(data, ct), interior)
	if aerr != nil {
		r.warn(slideID, fmt.Sprintf("image %q: %v", v.AssetID, aerr))
		return
	}
	if v.Alt != "" {
		img.SetAltText(v.Alt)
	}
	// Crop and fit drive the builder's existing srcRect/stretch (D-039). SetFit
	// with the default FitFill is idempotent (AddImage already stretches), so the
	// uncropped/default case stays byte-identical to a Phase-10 image.
	if v.Crop != (Crop{}) {
		img.SetCrop(v.Crop)
	}
	img.SetFit(v.Fit)
	r.stats.Shapes++
	r.stats.Assets++
}

// resolveFrameName resolves an Image's frame to a registry name (D-038):
// FrameName wins when non-empty, else the FrameKind enum maps to a curated
// name (FrameNone → "", i.e. no frame).
func resolveFrameName(v Image) string {
	if v.FrameName != "" {
		return v.FrameName
	}
	return frameKindName(v.Frame)
}

// frameKindName maps a curated FrameKind to its reserved registry name.
func frameKindName(k FrameKind) string {
	switch k {
	case FrameBrowser:
		return frames.NameBrowser
	case FramePhone:
		return frames.NamePhone
	case FrameDesktop:
		return frames.NameDesktop
	case FrameLaptop:
		return frames.NameLaptop
	default:
		return "" // FrameNone
	}
}

// validateFrameRefs is the registry-aware half of Stage-1 validation (RFC
// §14.4, D-038): every Image whose resolved frame name is non-empty must
// resolve in the render's registry (curated ∪ extensions). It runs in Render
// (not the option-free ValidateScene) because the registry derives from render
// options. It walks container children so a nested framed Image is covered.
func validateFrameRefs(s Scene, reg *frames.Registry) error {
	var errs []error
	for i := range s.Slides {
		sl := &s.Slides[i]
		where := sl.ID
		if where == "" {
			where = fmt.Sprintf("#%d", i)
		}
		walkImages(sl.Nodes, func(img Image) {
			name := resolveFrameName(img)
			if name == "" {
				return
			}
			if _, ok := reg.Lookup(name); !ok {
				errs = append(errs, fmt.Errorf(
					"slide %s: image frame %q is not a curated or registered frame (have %v)",
					where, name, reg.Names()))
			}
		})
	}
	return errors.Join(errs...)
}

// walkImages visits every Image in a node tree, recursing into container
// children (two_column, grid, card, card_section).
func walkImages(nodes []SlideNode, fn func(Image)) {
	for _, n := range nodes {
		switch v := n.(type) {
		case Image:
			fn(v)
		case TwoColumn:
			walkImages(v.Left, fn)
			walkImages(v.Right, fn)
		case Grid:
			walkImages(v.Cells, fn)
		case Card:
			walkImages(v.Body, fn)
		case CardSection:
			walkImages(v.Body, fn)
		case Bento:
			walkImages(v.cellNodes(), fn)
		}
	}
}
