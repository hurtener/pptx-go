package scene

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"  // register GIF for DecodeConfig
	_ "image/jpeg" // register JPEG for DecodeConfig
	_ "image/png"  // register PNG for DecodeConfig
	"math"

	"github.com/hurtener/pptx-go/pptx"
)

// Chart composer (RFC §15.1 / §12, D-004/D-046). A V1 chart renders as a
// caller-rasterized pic that contains-to-fit its slot (aspect preserved), with
// an optional caption below. The chart image's dimensions are read from its
// header (image.DecodeConfig — not pixel data, D-046) to fit it and to warn when
// its aspect ratio diverges from the slot. An unresolved asset draws a labeled
// ChartPlaceholder instead of a blank gap (D-036).

const (
	chartCaptionH = pptx.EMU(365760) // 0.4"
	chartARThresh = 0.15             // aspect divergence that triggers a warning
)

func (r *renderer) renderChart(ps *pptx.Slide, box pptx.Box, v Chart, slideID string) {
	slot := box
	if v.Caption != "" {
		slot.H = box.H - chartCaptionH
	}

	data, ct, err := r.resolve(v.AssetID)
	if err != nil {
		r.warn(slideID, fmt.Sprintf("chart asset %q unresolved: %v", v.AssetID, err))
		ps.ChartPlaceholder(slot)
		r.stats.Shapes += 2 // placeholder rect + label
	} else {
		placed := slot
		if w, h, ok := imageDims(data); ok {
			placed = containFit(slot, w, h)
			if pct := aspectDivergencePct(slot, w, h); pct > int(chartARThresh*100) {
				r.warn(slideID, fmt.Sprintf("chart %q aspect ratio diverges from its slot by ~%d%%; fit within the slot", v.AssetID, pct))
			}
		}
		if _, aerr := ps.AddImage(pptx.ImageBytes(data, ct), placed); aerr != nil {
			r.warn(slideID, fmt.Sprintf("chart image %q: %v", v.AssetID, aerr))
		} else {
			r.stats.Shapes++
			r.stats.Assets++
		}
	}

	if v.Caption != "" {
		capBox := pptx.Box{X: box.X, Y: slot.Y + slot.H, W: box.W, H: chartCaptionH}
		tf := ps.AddTextFrame(capBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
	}
}

// imageDims reads an image's pixel dimensions from its format header only
// (image.DecodeConfig does not decode pixel data — §7/D-046). Returns ok=false
// on any unreadable/zero header, so the caller degrades silently.
func imageDims(b []byte) (w, h int, ok bool) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, false
	}
	return cfg.Width, cfg.Height, true
}

// containFit returns the largest box at the image's w:h aspect that fits inside
// slot, centered. Integer-EMU, deterministic (D-035).
func containFit(slot pptx.Box, w, h int) pptx.Box {
	slotAR := float64(slot.W) / float64(slot.H)
	imgAR := float64(w) / float64(h)
	out := slot
	if imgAR > slotAR { // wider than the slot → width-bound
		out.W = slot.W
		out.H = pptx.EMU(math.Round(float64(slot.W) / imgAR))
	} else { // taller than the slot → height-bound
		out.H = slot.H
		out.W = pptx.EMU(math.Round(float64(slot.H) * imgAR))
	}
	out.X = slot.X + (slot.W-out.W)/2
	out.Y = slot.Y + (slot.H-out.H)/2
	return out
}

// aspectDivergencePct is the rounded percent by which the slot's aspect ratio
// diverges from the image's (relative to the image), for a deterministic warning.
func aspectDivergencePct(slot pptx.Box, w, h int) int {
	slotAR := float64(slot.W) / float64(slot.H)
	imgAR := float64(w) / float64(h)
	return int(math.Round(math.Abs(slotAR-imgAR) / imgAR * 100))
}
