# Brief 27 — fill-cap-no-overgrow-sparse

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 44 — fill cap (no over-grow) (R10.6, HIGH · engine)

## 1. Question

`VAlignFill` grows flexible nodes proportionally to their preferred height with no
ceiling, so a near-empty node balloons (the recreation's "Canvas" card grew to an
enormous height holding one sentence). How can the engine, **opt-in and
deterministically**, cap each flexible node's growth so excess slack becomes
balanced whitespace instead of one oversized node, while leaving uncapped
`VAlignFill` byte-identical?

## 2. Prior art surveyed

- `scene/render.go` `distributeFill(nodes, heights, slack)` — grows each flexible
  node by `slack × heights[idx] / flexH`, the last absorbing the rounding
  remainder, so the added heights sum to exactly `slack`. Integer EMU,
  worker-count independent.
- `scene/render.go` `alignedStackIn` — for `VAlignFill`, calls `distributeFill`
  with the positive slack then places top-pinned at the standard gap. `VAlignFit`
  (R10.2, D-071) added a parallel opt-in mode + a `fitCompress` helper — the
  precedent shape for a new fill variant.
- `scene/align.go` — the `VAlign` enum; `VAlignJustify` already distributes slack
  into inter-node gaps (the "even spacing" mechanism this brief reuses).
- DECKARD R10.6 spec: extend `distributeFill` with an opt-in cap — each flexible
  node grows by at most `growthMax × preferredH` (or a per-node `MaxGrow`); slack
  beyond the caps is left as balanced spacing (start offset / inter-node gaps);
  default (no cap) byte-identical; deterministic integer EMU.

## 3. Findings

- The cap is a **bounded** variant of `distributeFill`: grow each flexible node by
  its proportional share *capped* at `growthMax × preferredH`, sum the actual
  growth (`used ≤ slack`), and hand the **residual** (`slack − used`) back to the
  caller as balanced spacing. Because shares are capped (not remainder-pinned to
  the last node), `used` can be < `slack` — that gap is the whole point.
- **Opt-in as a new `VAlign` value (`VAlignFillCapped`)**, mirroring `VAlignFit`.
  This keeps plain `VAlignFill` byte-identical (it still calls the unchanged
  `distributeFill`), needs no `Alignment` struct change, and composes through the
  existing `align.Vertical` switch. A per-node `MaxGrow` field (the spec's
  alternative) would touch every flexible node type for marginal benefit — a
  pinned `growthMax` + a mode is simpler and sufficient for the gap.
- **Pinned `growthMax`.** `fillGrowthMaxBP = 10000` (basis points) → a node grows
  by at most 1.0× its preferred height (at most doubles). Modest enough that a
  one-line card stays proportionate, generous enough that genuine fill still
  works.
- **Balanced residual = even spacing.** Distribute the residual across the `n+1`
  spaces of the stack: `space = residual/(n+1)`; `startY = box.Y + space` (top
  margin) and `effectiveGap = gap + space` (each internal gap), leaving the
  remainder as bottom margin. Pure integer division — deterministic. For a single
  flexible node this centers it (top = bottom = residual/2). This reuses the
  `VAlignJustify`/`VAlignFit` gap-and-offset mechanism, so no new placement path.
- **Byte-identical default.** Uncapped `VAlignFill` is untouched; `VAlignFillCapped`
  is the new branch. When content fits at scale 1 (no slack) both behave like
  top-pinned.
- **Determinism.** `distributeFillCapped` is integer/basis-point over a fixed node
  order; the residual spacing is integer division — worker-count independent.

## 4. Recommendations

1. Add `VAlignFillCapped` to the `VAlign` enum (after `VAlignFit`) + `String()`
   (`"fill-capped"`).
2. Add `distributeFillCapped(nodes, heights, slack) → used`: proportional share
   capped at `fillGrowthMaxBP × preferredH` per node; mutate heights; return the
   total growth used.
3. In `alignedStackIn`, add a `VAlignFillCapped` branch: grow via
   `distributeFillCapped`, then distribute `residual = slack − used` as even
   spacing (`startY += residual/(n+1)`, `effectiveGap += residual/(n+1)`).
4. Tests: a sparse + dense fill stack — the sparse node grows ≤ its cap and the
   residual appears as even spacing; uncapped `VAlignFill` byte-identical;
   determinism guard; smoke `phase-44.sh`.

## 5. Open questions

- **Per-node `MaxGrow`** — deferred; a pinned `growthMax` covers the gap. The
  `distributeFillCapped` seam can take a per-node cap later if a req wants it.
- **Interaction with `VAlignFit`** — orthogonal: Fit compresses an over-full
  stack, FillCapped bounds an under-full one; a stack is one or the other.
