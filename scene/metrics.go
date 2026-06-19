package scene

import "github.com/hurtener/pptx-go/pptx"

// Text-width estimation for alignment (Phase 13). naturalWidth is a pure,
// deterministic estimator: no DOM, no measurement, pinned constants. It is
// used only for horizontal centering/right-aligning the body stack; it is
// NOT used for height/line-count estimation (which is a separate later unit).

// avgCharWidthFactor is the ratio of average character width to font point
// size. Calibrated for the default sans-serif (Calibri/Inter); pinned as a
// compile-time constant for determinism. 0.5 is a conservative estimate:
// wide enough to avoid false "narrower than box" conclusions, not so wide
// that it fails to detect short text.
const avgCharWidthFactor = 0.5

// emuPerPointMetrics is the EMU-per-point conversion used in metrics
// (mirroring the private emuPerPoint in pptx/units.go; 914400/72 = 12700).
const emuPerPointMetrics = 12700

// naturalWidth returns a deterministic estimate of the horizontal span of rt
// when rendered on theme. For each run the width contribution is:
//
//	len(text) × floor(fontSize_pt × avgCharWidthFactor × emuPerPoint)
//
// The font size is resolved from the run's Style.TypeRole via
// theme.ResolveType; an absent TypeRole entry falls back to the theme's
// fallback (14 pt Calibri). The TypeRole zero value (TypeDisplay = 0) is used
// as-is (it maps to 40 pt in the default theme). Use naturalWidthAt when the
// caller knows the node's base rendering role and the runs may carry an unset
// TypeRole.
//
// naturalWidth is a pure function: identical (rt, theme) inputs always return
// the same value. It allocates no heap memory.
func naturalWidth(rt RichText, theme *pptx.Theme) pptx.EMU {
	var total pptx.EMU
	for _, run := range rt {
		if len(run.Text) == 0 {
			continue
		}
		spec := theme.ResolveType(run.Style.TypeRole)
		// avgW: average char width in EMU, truncated to integer (deterministic).
		avgW := pptx.EMU(spec.Size * avgCharWidthFactor * emuPerPointMetrics)
		total += avgW * pptx.EMU(len(run.Text))
	}
	return total
}

// naturalWidthAt is naturalWidth with a base TypeRole. For each run whose
// Style.TypeRole is 0 (TypeDisplay — the Go zero value, treated as "unset"
// here), base is substituted before delegating to naturalWidth. Callers that
// know the node's rendering base role (e.g. TypeH2 for a level-2 Heading)
// use this for a more accurate estimate.
//
// If the caller wants TypeDisplay explicitly, set run.Style.TypeRole =
// pptx.TypeDisplay (which is also 0); naturalWidthAt cannot distinguish
// "unset" from "explicitly TypeDisplay" — both will use base. For TypeDisplay
// content, call naturalWidth directly instead.
func naturalWidthAt(rt RichText, base pptx.TypeRole, theme *pptx.Theme) pptx.EMU {
	if len(rt) == 0 {
		return 0
	}
	enriched := make(RichText, len(rt))
	for i, run := range rt {
		enriched[i] = run
		if enriched[i].Style.TypeRole == 0 {
			enriched[i].Style.TypeRole = base
		}
	}
	return naturalWidth(enriched, theme)
}

// nodeNaturalWidth estimates the dominant text width of a leaf node on theme,
// using the node's actual rendering TypeRole as the base. Container and visual
// nodes (Grid, TwoColumn, Table, Card, …) return 0 — they always span the
// full box width and alignment within them is their own concern.
//
// The estimate is a single-width value per node (the dominant run or the
// widest visible part). It is NOT a max-line-width across wrapped paragraphs
// (line-wrapping is a separate later unit).
func nodeNaturalWidth(n SlideNode, theme *pptx.Theme) pptx.EMU {
	switch v := n.(type) {
	case Hero:
		// Title is the dominant visual element; it renders at TypeDisplay
		// (zero TypeRole = TypeDisplay is correct here — call naturalWidth
		// directly so the zero TypeRole resolves to TypeDisplay, not a base).
		return naturalWidth(RichText{{Text: v.Title}}, theme)
	case Heading:
		return naturalWidthAt(v.Text, headingRole(v.Level), theme)
	case Prose:
		if len(v.Paragraphs) == 0 {
			return 0
		}
		// Use the first paragraph as the width representative.
		return naturalWidthAt(v.Paragraphs[0], pptx.TypeBody, theme)
	case Quote:
		return naturalWidthAt(v.Text, pptx.TypeH3, theme)
	case Chip:
		return naturalWidth(RichText{{Text: v.Label, Style: RunStyle{TypeRole: pptx.TypeBodySmall}}}, theme)
	case SectionDivider:
		// Label renders at TypeDisplay (zero TypeRole correct here too).
		return naturalWidth(RichText{{Text: v.Label}}, theme)
	}
	// Containers and visuals are always full-width; callers should not
	// center/right-align them — nodeEffectiveHAlign returns HAlignLeft for them.
	return 0
}
