// Package frames holds the curated device-frame shape recipes (RFC §14.3).
//
// A recipe draws a device bezel — browser, phone, desktop, laptop — as native
// PPTX shapes into a region and returns the interior box the renderer inserts
// the framed image into, plus the number of bezel shapes it emitted. Recipes
// compose the public pptx builder only (P1): they never reach under it and
// never construct raw OOXML (P3). All visible color flows through Theme tokens
// (P2) — bezels reuse ColorSurface/ColorSurfaceAlt and the browser's traffic
// lights reuse ColorError/ColorWarning/ColorSuccess; no new token is
// introduced. Geometry is pure integer-EMU arithmetic on the region, so a
// re-render is byte-identical (D-035).
//
// The scene/frames registry wires these functions to the curated names and
// adds the §14.4 caller-extension overlay; the recipe signature there
// (scene/frames.Recipe) matches the functions in this package exactly.
package frames

import "github.com/hurtener/pptx-go/pptx"

// fill is a solid token fill — the bezel color path (P2).
func fill(role pptx.ColorRole) pptx.ShapeOption {
	return pptx.WithFill(pptx.SolidFill(pptx.TokenColor(role)))
}

// maxEMU returns the larger of a and b.
func maxEMU(a, b pptx.EMU) pptx.EMU {
	if a > b {
		return a
	}
	return b
}
