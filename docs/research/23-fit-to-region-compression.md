# Brief 23 — fit-to-region-compression

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 40 — fit-to-region-compression (R10.2, CRITICAL · engine)

## 1. Question

When a slide's body stack is taller than its region, the scene renderer
today places the overflowing nodes off-box and only records a
`content overflows its region` warning — the content is clipped below the
slide edge (the recreation drew its bottom bento row partially off-canvas).
How can the engine, **opt-in and deterministically**, compress an
over-full stack so its last node lands inside the region, while leaving
fitting content and every existing alignment mode byte-identical?

## 2. Prior art surveyed

- `scene/render.go` `alignedStackIn` (the sole body-stack layout, Phase 13):
  computes per-node `preferredHeight`, a `bodyH` (sum of node heights), and
  `totalH = bodyH + gap*(n-1)`; positions by `align.Vertical`
  (Top/Center/Bottom/Justify/Fill). `VAlignJustify` already *expands* gaps
  into slack; `VAlignFill` grows flexible nodes via `distributeFill`. The
  overflow warning fires unconditionally on `totalH > box.H`.
- `scene/render.go` `stackIn` (container sub-layout for TwoColumn/Grid/Card
  cells) — fixed `SpaceMD` gap, same warn.
- `pptx/theme.go` spacing tokens: `SpaceXS=Pt(2)`, `SpaceSM=Pt(4)`,
  `SpaceMD=Pt(8)` (the stack gap), `SpaceLG=Pt(16)`, `SpaceXL=Pt(24)`.
  `ResolveSpace` maps the role → EMU.
- Phase 23 (R2, `VAlignFill` / `distributeFill`, D-052) — the established
  "grow to fit" inverse: integer-EMU proportional distribution with the
  rounding remainder pinned to the last node, proven worker-count
  independent. Phase 39 (R10.1, D-070) — wrapped-header geometry, the first
  Wave-10 content-fit unit.
- DECKARD R10.2 spec: opt-in fit mode honored by `alignedStackIn`; when
  `totalH > box.H`, `s = clamp((box.H - fixedFloors)/flexibleH, sMin, 1)`
  applied in priority — (1) inter-node gaps → `gapMin`, (2) card padding →
  `padMin`, (3) one type-scale step within a ratio floor. Integer-EMU,
  worker-count-independent. Default OFF byte-identical + the existing warning.

## 3. Findings

- The off-slide clip is a **placement** problem, not a content opinion: the
  body stack's preferred geometry exceeds the region and the engine places
  it anyway. Compressing the slot geometry (gaps, then slot heights) so the
  last node bottom ≤ `box.Bottom()` is a pure mechanism (D-026) — it does
  not decide *what* the deck contains, only that its boxes fit the frame.
- The two levers fully inside `alignedStackIn` are **inter-node gaps** and
  **node slot heights**. The spec's deeper sub-steps — card interior padding
  (→ `padMin`) and an explicit font type-scale step — live in `renderCard`
  (R10.7 `density-aware-card-padding`) and in the display-text shrink
  (R10.5 `display-text-shrink-to-fit`) respectively. The R10.2 spec text
  itself cross-references those: R10.7 "an auto-tighten step **inside the
  fit-to-region pass**", R10.5 the type shrink. So the fit pass is layered:
  Phase 40 establishes the gap-then-slot-height primitive with pinned
  floors; R10.7/R10.5 plug their steps into the same pass.
- A `VAlignFit` value on the existing `VAlign` enum is the minimal additive
  surface: it routes through the `align.Vertical` switch exactly like
  `VAlignFill` (its inverse), needs no `Alignment` struct change (keeps the
  struct comparable), preserves all existing iota values, and is opt-in per
  slide via `SceneSlide.Content.Vertical`.
- **Determinism:** gap reduction is integer EMU; slot-height scaling is
  basis-point integer math (`h*sBP/10000`), a pure function of the heights,
  the gap, and `box.H` — no float, no worker-order dependency. Mirrors the
  proven `distributeFill` discipline.
- **Pinned floors:** `gapMin = SpaceXS` (the smallest positive spacing
  token, Pt(2)) and `sMin = 0.60` (basis-point `6000`). A ~25% overflow
  (`totalH ≈ 1.25·box.H`) needs `s ≈ 0.8` after gaps minimize — inside
  `[0.60, 1]`, so the acceptance band fits via the pinned steps alone.
- **Byte-identical guarantees:** with the flag off, no code path changes.
  With the flag on but `totalH ≤ box.H`, the compression branch is skipped
  and `VAlignFit` falls through to top-pinned placement with the standard
  gap — identical to `VAlignTop`. Only an actually-overflowing fit slide
  changes (the fix).
- **Truthful warning:** for non-Fit modes keep the existing
  `totalH > box.H` warn verbatim (Stats-observed, tested). For `VAlignFit`,
  recompute after compression and warn iff content *still* overflows (it hit
  the pinned floors and could not fully fit) — so a successful fit suppresses
  the warning and the signal stays trustworthy (R10.10/R10.11).

## 4. Recommendations

1. Add `VAlignFit` to the `VAlign` enum (after `VAlignFill`) + its `String()`
   case (`"fit"`). Document it as the compression inverse of `VAlignFill`.
2. Add a renderer method `fitCompress(heights, bodyH, gap, box) → effGap`
   that (a) floors the inter-node gap toward `gapMin` to absorb the overflow,
   then (b) if still overflowing at `gapMin`, scales every slot height by
   `sBP = clamp(avail*10000/bodyH, 6000, 10000)` where
   `avail = box.H - effGap*(n-1)`. Mutates `heights` in place, returns the
   compressed gap. Pure integer/basis-point math.
3. In `alignedStackIn`, invoke `fitCompress` only when
   `align.Vertical == VAlignFit && totalH > box.H`; keep `startY = box.Y`
   (top-pinned). Make the overflow warn mode-aware as in Finding 7.
4. Keep `fitCompress` a reusable primitive (theme-aware for `gapMin`) so the
   container stackers can adopt it later (R10.3 bento rows, R10.4 card body).
5. Tests: white-box `fitCompress` (gap-only fit, gap+height fit, floor-capped
   residual, no-op when fitting); placement-level acceptance (last node
   bottom ≤ `box.Bottom()` for a ~25% overflow); byte-identical `VAlignFit`
   vs `VAlignTop` when content fits; a determinism guard
   (`TestRenderDeterministic_VAlignFit`); smoke `phase-40.sh`.

## 5. Open questions

- **Container-internal fit.** `stackIn` (TwoColumn/Grid/Card cells) has no
  alignment param; threading a fit mode through every caller is out of scope
  here. R10.3 (`content-weighted-bento-grid-rows`) and R10.4
  (`card-body-vertical-distribution`) own container-internal distribution and
  will reuse this phase's `fitCompress` primitive + pinned floors.
- **Card padding / type-scale sub-steps.** Deferred to R10.7 (padding
  auto-tighten inside the fit pass) and R10.5 (display-text shrink), per the
  spec's own cross-references. Phase 40 delivers the gap + slot-height steps,
  which satisfy the ≤25%-overflow acceptance on their own.
- **Estimate/actual parity.** The fit pass is only as trustworthy as
  `preferredHeight`; R10.10 (`estimate-actual-parity-fit-budget`) closes the
  remaining estimator gaps (wrapped headers already done in R10.1, wide-span
  bento cells pending).
