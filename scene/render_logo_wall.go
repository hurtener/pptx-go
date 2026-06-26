package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// LogoWall composer (R14.7, D-125). An N-up grid of logo assets, each contained
// (not cropped) and centered in its cell at a common optical size, optionally
// recolored to a uniform tone (mono/brand) so a mixed set reads as one cohesive
// wall. Asset-bearing; a missing logo warns and is skipped (RFC §10.2).

const (
	logoWallDefaultCols = 4
	logoCellPad         = pptx.EMU(91440)  // In(0.10); padding inside each logo cell
	logoRowH            = pptx.EMU(914400) // In(1.0); a pinned row height for the estimate
	logoCaptionH        = pptx.EMU(365760) // In(0.40); the optional caption strip
)

func logoWallPreferredHeight(v LogoWall) pptx.EMU {
	cols := v.Columns
	if cols < 1 {
		cols = logoWallDefaultCols
	}
	rows := (len(v.Logos) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	h := logoRowH * pptx.EMU(rows)
	if v.Caption != "" {
		h += logoCaptionH
	}
	return h
}

func (r *renderer) renderLogoWall(ps *pptx.Slide, box pptx.Box, v LogoWall, slideID string) {
	cols := v.Columns
	if cols < 1 {
		cols = logoWallDefaultCols
	}
	grid := box
	if v.Caption != "" {
		cf := ps.AddTextFrame(pptx.Box{X: box.X, Y: box.Y, W: box.W, H: logoCaptionH}).Anchor(pptx.AnchorMiddle)
		cp := cf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		cp.AddRun(v.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
		grid = pptx.Box{X: box.X, Y: box.Y + logoCaptionH, W: box.W, H: box.H - logoCaptionH}
	}
	n := len(v.Logos)
	rows := (n + cols - 1) / cols
	if rows < 1 || grid.W <= 0 || grid.H <= 0 {
		return
	}
	cellW := grid.W / pptx.EMU(cols)
	cellH := grid.H / pptx.EMU(rows)

	for i, logo := range v.Logos {
		col := i % cols
		row := i / cols
		cell := pptx.Box{X: grid.X + cellW*pptx.EMU(col), Y: grid.Y + cellH*pptx.EMU(row), W: cellW, H: cellH}
		inner := pptx.Box{X: cell.X + logoCellPad, Y: cell.Y + logoCellPad, W: cell.W - 2*logoCellPad, H: cell.H - 2*logoCellPad}
		data, ct, err := r.resolve(logo.AssetID)
		if err != nil {
			r.warn(slideID, fmt.Sprintf("logo %q unresolved: %v", logo.AssetID, err))
			continue
		}
		dst := containBox(inner, data)
		img, aerr := ps.AddImage(pptx.ImageBytes(data, ct), dst)
		if aerr != nil {
			r.warn(slideID, fmt.Sprintf("logo %q: %v", logo.AssetID, aerr))
			continue
		}
		if logo.Alt != "" {
			img.SetAltText(logo.Alt)
		}
		switch v.Tone {
		case LogoToneMono:
			img.SetDuotone(pptx.TokenTextColor(pptx.TextPrimary), pptx.TokenColor(pptx.ColorCanvas))
		case LogoToneBrand:
			img.SetDuotone(pptx.TokenColor(pptx.ColorAccent), pptx.TokenColor(pptx.ColorCanvas))
		}
		r.stats.Shapes++
		r.stats.Assets++
	}
}

// containBox returns the largest box of the image's aspect that fits within cell,
// centered (so a logo is shown whole, not cropped). When the dimensions are
// unreadable it returns the cell unchanged (best effort).
func containBox(cell pptx.Box, data []byte) pptx.Box {
	w, h, ok := imageDims(data)
	if !ok || w <= 0 || h <= 0 || cell.W <= 0 || cell.H <= 0 {
		return cell
	}
	// Fit the image aspect inside the cell: compare imgW/imgH vs cellW/cellH.
	imgW, imgH := int64(w), int64(h)
	cellW, cellH := int64(cell.W), int64(cell.H)
	var dw, dh pptx.EMU
	if imgW*cellH >= cellW*imgH {
		// Image is wider than the cell → width-bound.
		dw = cell.W
		dh = pptx.EMU(int64(cell.W) * imgH / imgW)
	} else {
		dh = cell.H
		dw = pptx.EMU(int64(cell.H) * imgW / imgH)
	}
	return pptx.Box{X: cell.X + (cell.W-dw)/2, Y: cell.Y + (cell.H-dh)/2, W: dw, H: dh}
}
