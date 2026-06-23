package scene

import "github.com/hurtener/pptx-go/pptx"

// Text-width estimation (Phase 13) and wrapped-line-count estimation
// (Phase 22). naturalWidth is a pure, deterministic estimator: no DOM, no
// measurement, pinned constants. It horizontally centers/right-aligns the body
// stack, and — via wrappedLines — feeds the content-aware preferredHeight so a
// node's slot grows with the text that wraps into it.

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
		// Per-face factor (D-064): the role's measured AvgCharWidth when set, else
		// the built-in sans fallback (0.5) — byte-identical for an unset face.
		factor := spec.AvgCharWidth
		if factor <= 0 {
			factor = avgCharWidthFactor
		}
		// avgW: average char width in EMU, truncated to integer (deterministic).
		avgW := pptx.EMU(spec.Size * factor * emuPerPointMetrics)
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
// widest visible part). It is NOT a max-line-width across wrapped paragraphs;
// wrappedLines handles vertical line-count estimation separately.
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

// AutoFit (shrink-to-fit) pinned constants (R10.5, D-074). The scale is
// quantized to a fixed step and floored at a minimum ratio so the result is
// deterministic and never shrinks past a legible bound.
const (
	autofitRatioMinBP = 6000 // 0.60 — the minimum scale (pinned ratio floor)
	autofitStepBP     = 250  // 0.025 — quantization step (keeps the float deterministic)
)

// fitScale returns the font-scale multiplier that shrinks a run whose estimated
// natural width is natW toward boxW, or 0 when it already fits (or the inputs are
// unknown) — 0 means "no scaling", so an AutoFit-off or already-fitting run is
// byte-identical. When natW > boxW it returns floor(boxW/natW) quantized down to
// autofitStepBP and floored at autofitRatioMinBP, expressed as a fraction in
// [0.60, 1). It never returns >= 1 (never upscales). The 0.60 floor is a legibility
// bound, so it does NOT guarantee the run fits: text much wider than its box
// (natW >> boxW, e.g. a value in a sub-~0.85" cell) can still overflow at the floor
// — that residual overflow is accepted (see D-088). Pure integer / basis-point math:
// identical (natW, boxW) inputs always return the same scale.
func fitScale(natW, boxW pptx.EMU) float64 {
	if natW <= 0 || boxW <= 0 || natW <= boxW {
		return 0
	}
	raw := boxW * 10000 / natW // floored basis points; < 10000 since natW > boxW
	q := (raw / autofitStepBP) * autofitStepBP
	if q < autofitRatioMinBP {
		q = autofitRatioMinBP
	}
	if q >= 10000 {
		return 0
	}
	return float64(q) / 10000
}

// wrappedLines estimates how many lines rt occupies when laid out in a column
// of width avail, using the same pinned char-width model as naturalWidth. It is
// the vertical complement of naturalWidth: where naturalWidth answers "how wide
// is this text on one line", wrappedLines answers "how many lines does it take
// in this width". The estimate is the char-budget model
//
//	lines = ceil(naturalWidthAt(rt, base, …) / avail)
//
// floored at 1. base is the node's rendering TypeRole, substituted for runs
// that carry the zero TypeRole (see naturalWidthAt).
//
// wrappedLines is a pure, deterministic function: pure integer ceil division
// over the (already deterministic) naturalWidth result, no measurement, no
// allocation. It deliberately returns 1 — i.e. defers to the caller's existing
// fixed per-line height — when avail <= 0 or theme is nil, so a content-aware
// preferredHeight that lacks a real width/theme reproduces the pre-Phase-22
// fixed-height output byte-for-byte. It never claims long text is one line,
// which is the only property the overlap/overflow guarantees require.
func wrappedLines(rt RichText, base pptx.TypeRole, avail pptx.EMU, theme *pptx.Theme) int {
	if avail <= 0 || theme == nil {
		return 1
	}
	w := naturalWidthAt(rt, base, theme)
	if w <= 0 {
		return 1
	}
	// ceil(w / avail) with positive integers.
	lines := int((w + avail - 1) / avail)
	if lines < 1 {
		lines = 1
	}
	return lines
}
