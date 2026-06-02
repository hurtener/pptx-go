// Package ornaments holds the curated preset ornament shape recipes (RFC §14.2,
// D-005). Each recipe composes native PPTX shapes into a box in the active
// accent token, at a caller opacity (alpha) and rotation. Recipes use the public
// pptx builder only (P1) — gradients (glows), rotation, and token-alpha all land
// in Phase-13 PR #1 (D-041). Geometry is pure integer-EMU arithmetic, so a
// re-render is byte-identical (D-035); no recipe uses randomness or wall-clock.
//
// The scene/ornaments registry wires these functions to the curated names and
// adds the §14.4 caller-extension overlay; its Recipe type matches the functions
// here exactly.
package ornaments

import "github.com/hurtener/pptx-go/pptx"

// accent is a solid accent-token fill at the given OOXML alpha (the ornament
// color path — P2). alpha == 0 is fully transparent; AlphaOpaque is solid.
func accent(alpha int) pptx.ShapeOption {
	return pptx.WithFill(pptx.SolidFill(pptx.TokenColorAlpha(pptx.ColorAccent, alpha)))
}

// minEMU returns the smaller of a and b.
func minEMU(a, b pptx.EMU) pptx.EMU {
	if a < b {
		return a
	}
	return b
}
