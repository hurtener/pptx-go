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
	// Rounded-corner clip + drop shadow from theme tokens (D-114). Both self-gate:
	// RadiusNone leaves the picture rectangular and ElevationFlat emits no shadow,
	// so a Phase-10 image (zero tokens) is byte-identical.
	img.SetCornerRadius(v.CornerRadius)
	img.SetElevation(v.Elevation)
	r.stats.Shapes++
	r.stats.Assets++

	// Annotations (R14.17): numbered pins + highlight boxes overlaid on the image
	// interior at fractional coordinates. Drawn after the pic so they sit on top.
	if v.Annotations != nil {
		r.renderImageAnnotations(ps, interior, v.Annotations)
	}
}

// Annotation geometry (R14.17, D-130). Pinned metrics; colors are tokens.
const (
	annPinR = pptx.EMU(146304)  // In(0.16); a numbered pin's radius
	annCapW = pptx.EMU(1280160) // In(1.40); a leader caption box width
	annCapH = pptx.EMU(192024)  // In(0.21); a leader caption line height
	annGap  = pptx.EMU(45720)   // In(0.05)
)

// renderImageAnnotations draws highlight rects (behind pins) then numbered pins
// with optional leader-line captions, at fractional coordinates of box.
func (r *renderer) renderImageAnnotations(ps *pptx.Slide, box pptx.Box, a *ImageAnnotations) {
	for _, h := range a.Highlights {
		hb := pptx.Box{
			X: box.X + pptx.EMU(clampUnit01(h.X)*float64(box.W)),
			Y: box.Y + pptx.EMU(clampUnit01(h.Y)*float64(box.H)),
			W: pptx.EMU(clampUnit01(h.W) * float64(box.W)),
			H: pptx.EMU(clampUnit01(h.H) * float64(box.H)),
		}
		ps.AddShape(pptx.ShapeRect, hb, pptx.WithFill(pptx.NoFill()),
			pptx.WithLine(pptx.Line{Width: pptx.Pt(2), Color: pptx.TokenColor(timelineAccent(h.AccentIndex))}))
		r.stats.Shapes++
	}
	for _, p := range a.Pins {
		px := box.X + pptx.EMU(clampUnit01(p.X)*float64(box.W))
		py := box.Y + pptx.EMU(clampUnit01(p.Y)*float64(box.H))
		accent := timelineAccent(p.AccentIndex)
		// Optional leader + caption to the right of the pin (clamped into the box).
		if p.Caption != "" {
			cx := px + annPinR + annGap
			if cx+annCapW > box.Right() {
				cx = px - annPinR - annGap - annCapW
			}
			if cx < box.X {
				cx = box.X
			}
			r.hvLine(ps, px, py, (cx+annCapW/2)-px, 0, pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)})
			cf := ps.AddTextFrame(pptx.Box{X: cx, Y: py - annCapH/2, W: annCapW, H: annCapH}).Anchor(pptx.AnchorMiddle)
			cp := cf.AddParagraph(pptx.ParagraphOpts{})
			cp.AddRun(p.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextPrimary)})
			r.stats.Shapes++
		}
		// Pin disc + number.
		ps.AddShape(pptx.ShapeEllipse, pptx.Box{X: px - annPinR, Y: py - annPinR, W: 2 * annPinR, H: 2 * annPinR},
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(accent))))
		r.stats.Shapes++
		if p.Label != "" {
			lf := ps.AddTextFrame(pptx.Box{X: px - annPinR, Y: py - annPinR, W: 2 * annPinR, H: 2 * annPinR}).Anchor(pptx.AnchorMiddle)
			lp := lf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			lp.AddRun(p.Label, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: r.cellTextOn(accent)})
			r.stats.Shapes++
		}
	}
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
		case Banner:
			walkImages(v.Trailing, fn)
		}
	}
}
