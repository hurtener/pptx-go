package pptx

import (
	"bytes"
	"image"
	_ "image/gif"  // register GIF for DecodeConfig
	_ "image/jpeg" // register JPEG for DecodeConfig
	_ "image/png"  // register PNG for DecodeConfig

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// coverSrcRect computes the source-rectangle crop that makes an image cover-fit
// box: the image is scaled to fill the box and the overflowing axis is
// center-cropped, so it fills the surface with no distortion at any aspect. The
// crop is derived from the image's format-header dimensions only (image.
// DecodeConfig — not pixel data, §7/D-046; the chart composer reads the same
// header for aspect-fit), in integer thousandths-of-a-percent, so it is
// deterministic regardless of worker count.
//
// It returns nil when the dimensions are unreadable (a plain stretch — best
// effort) or when the image and box already share an aspect (no crop needed, so
// stretch == cover and the output stays minimal).
func coverSrcRect(data []byte, box Box) *slide.XSrcRect {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || box.W <= 0 || box.H <= 0 {
		return nil
	}
	// Compare aspects by cross-multiplication (no floats): imgW/imgH vs boxW/boxH.
	imgW, imgH := int64(cfg.Width), int64(cfg.Height)
	boxW, boxH := int64(box.W), int64(box.H)
	lhs := imgW * boxH
	rhs := boxW * imgH
	if lhs == rhs {
		return nil // aspects match: a full stretch already covers without distortion
	}
	if lhs > rhs {
		// Image is wider than the box → crop left/right. Visible width fraction =
		// boxAR/imgAR = rhs/lhs; crop the remainder, split evenly between edges.
		cropTotal := 100000 - int(rhs*100000/lhs)
		l := cropTotal / 2
		return &slide.XSrcRect{L: l, R: cropTotal - l}
	}
	// Image is taller than the box → crop top/bottom.
	cropTotal := 100000 - int(lhs*100000/rhs)
	t := cropTotal / 2
	return &slide.XSrcRect{T: t, B: cropTotal - t}
}
