// Package ornaments is the scene-side ornament registry: it wires the curated
// ornament recipes (assets/ornaments) to their names and provides the per-render
// caller-extension overlay (RFC §14.2/§14.4, D-005, D-038).
//
// It mirrors scene/frames and scene/icons (the curated-asset seam): a closed
// curated name set plus caller extension, an immutable per-render overlay,
// read-only during compose. A recipe draws an ornament into a box at a caller
// opacity (OOXML alpha) and rotation, composing the public pptx builder only
// (P1).
package ornaments

import (
	"sort"

	assetornaments "github.com/hurtener/pptx-go/assets/ornaments"
	"github.com/hurtener/pptx-go/pptx"
)

// Recipe draws an ornament into box at the given OOXML alpha (0..100000),
// rotation (degrees), surface color role, and lattice pitch (EMU), returning the
// number of shapes emitted. The role is the decoration color (Decoration.Color,
// default ColorAccent — D-107). The pitch is the pattern spacing
// (Decoration.Pitch, D-111): pattern recipes (grid_dots/noise_overlay/starfield)
// derive their repeat count from it (0 = a legacy fixed count); the others ignore
// it. A recipe may ignore either. The signature matches the curated recipes in
// assets/ornaments exactly.
type Recipe func(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64, role pptx.ColorRole, pitch pptx.EMU) int

// The reserved curated ornament names (RFC §14.2).
const (
	NameGlowRing      = "glow_ring"
	NameRadialGlow    = "radial_glow"
	NameGridDots      = "grid_dots"
	NameCornerBracket = "corner_bracket"
	NameChevronArrow  = "chevron_arrow"
	NameNoiseOverlay  = "noise_overlay"
	NameStarfield     = "starfield"   // organic scatter of size/alpha-varied dots (R13.6, D-110)
	NameScatterDot    = "scatter_dot" // scatter family (R14.20, D-131)
	NameScatterStar   = "scatter_star"
	NameScatterPlus   = "scatter_plus"
	NameScatterRing   = "scatter_ring"
)

// PatternMaxDots re-exports the pattern-recipe dot cap (R13.7, D-111) so scene
// code (render_decoration's pitch-projection warning) shares the single source of
// truth with the recipes and the two can never drift (D-115).
const PatternMaxDots = assetornaments.PatternMaxDots

// Registry is an immutable, name-keyed set of ornament recipes. Lookup and Names
// are safe on a nil *Registry (treated as empty).
type Registry struct {
	m map[string]Recipe
}

// Curated returns a registry seeded with the seven curated ornaments.
func Curated() *Registry {
	return &Registry{m: map[string]Recipe{
		NameGlowRing:      assetornaments.GlowRing,
		NameRadialGlow:    assetornaments.RadialGlow,
		NameGridDots:      assetornaments.GridDots,
		NameCornerBracket: assetornaments.CornerBracket,
		NameChevronArrow:  assetornaments.ChevronArrow,
		NameNoiseOverlay:  assetornaments.NoiseOverlay,
		NameStarfield:     assetornaments.Starfield,
		NameScatterDot:    assetornaments.ScatterDot,
		NameScatterStar:   assetornaments.ScatterStar,
		NameScatterPlus:   assetornaments.ScatterPlus,
		NameScatterRing:   assetornaments.ScatterRing,
	}}
}

// With returns a copy of the registry with name bound to rec (overriding any
// existing entry). The receiver is not mutated — extensions are per-render, not
// global. A blank name or nil recipe is ignored.
func (r *Registry) With(name string, rec Recipe) *Registry {
	size := 0
	if r != nil {
		size = len(r.m)
	}
	cp := &Registry{m: make(map[string]Recipe, size+1)}
	if r != nil {
		for k, v := range r.m {
			cp.m[k] = v
		}
	}
	if name != "" && rec != nil {
		cp.m[name] = rec
	}
	return cp
}

// Lookup returns the recipe registered under name, or (nil, false).
func (r *Registry) Lookup(name string) (Recipe, bool) {
	if r == nil {
		return nil, false
	}
	rec, ok := r.m[name]
	return rec, ok
}

// Names returns the registered ornament names in sorted order (validation
// messages).
func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}
	out := make([]string, 0, len(r.m))
	for k := range r.m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
