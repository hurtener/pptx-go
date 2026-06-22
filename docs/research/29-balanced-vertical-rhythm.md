# Brief 29 — balanced-vertical-rhythm-sparse

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 46 — balanced vertical rhythm (R10.8, MED · engine)

## 1. Question

On sparse slides (a cover, a closing) the body stack uses fixed inter-node gaps,
so the elements cluster with a large void: the recreation cover clusters
eyebrow/title/subtitle in the middle then drops the description far below; the
closing leaves the whole lower frame empty. `VAlignJustify` spreads slack into the
gaps but pins the stack edge-to-edge (no margins); `VAlignCenter` adds margins but
keeps fixed gaps. How can a sparse stack distribute whitespace as an **even
rhythm** — proportional spacing that fills the region without one large residual
void — optically centered, opt-in and byte-identical by default?

## 2. Prior art surveyed

- `scene/render.go` `alignedStackIn` — the vertical modes: `VAlignTop` (box.Y,
  fixed gap), `VAlignCenter` (centered, fixed gap), `VAlignBottom`, `VAlignJustify`
  (box.Y, gaps widened to `slack/(n-1)`), `VAlignFill`/`VAlignFillCapped` (grow /
  capped-grow). `VAlignFillCapped` (D-075) already distributes a residual across
  `n+1` even units (top margin + widened gaps) — the exact even-rhythm primitive
  this brief reuses.
- `scene/align.go` — the `VAlign` enum.
- DECKARD R10.8 spec: make `VAlignJustify` (and/or a new distribute/balanced mode)
  gap distribution proportional and group-aware — distribute slack into inter-node
  gaps weighted by a pinned ratio, with an optical-center offset (slightly above
  geometric center); keep `VAlignTop` default byte-identical; deterministic
  integer EMU; mechanism only.

## 3. Findings

- The even-rhythm primitive **already exists** in `VAlignFillCapped`'s residual
  handling: distribute the slack beyond the base gaps across `n+1` units —
  `unit = slack/(n+1)` — into a top margin and the `n-1` internal gaps
  (`effectiveGap = gap + unit`), the bottom absorbing the remainder. This gives
  proportional spacing with *no single large void*, which is exactly the
  acceptance.
- **Distinct from the existing modes.** `VAlignJustify` puts all slack into
  internal gaps (no margins → content spans edge-to-edge). `VAlignCenter` puts all
  slack into equal top/bottom margins (fixed internal gaps → content clusters).
  `VAlignBalanced` puts slack into *both* margins and gaps evenly → an even rhythm
  that reads balanced on a sparse slide.
- **Optical center.** "Slightly above geometric center" = bias the stack upward by
  shrinking the top margin below an even unit: `startY = box.Y + unit ×
  balancedOpticalBP/10000` with `balancedOpticalBP ≈ 8500` (top is 85 % of a
  unit). The freed space falls to the bottom margin (the remainder), so the stack
  sits a touch above center — the canonical optical-centering rule. Pure integer /
  basis-point — deterministic.
- **Group-aware weighting is the caller's, not the engine's.** The spec's "larger
  gap before a description block" requires knowing which node is a "description
  block" — that is content taste (D-026). The engine mechanism is the even rhythm
  + optical center; a caller that wants a larger pre-description gap orders its
  nodes or inserts a spacer. The optical bias is the pinned-ratio the engine owns.
- **A new mode, not a change to Justify.** Adding `VAlignBalanced` keeps
  `VAlignJustify`/`VAlignTop`/`VAlignCenter` byte-identical (the spec's hard
  requirement) and composes through the existing `align.Vertical` switch, mirroring
  `VAlignFit`/`VAlignFillCapped`.

## 4. Recommendations

1. Add `VAlignBalanced` to the `VAlign` enum (after `VAlignFillCapped`) +
   `String()` (`"balanced"`).
2. In `alignedStackIn`, add a `VAlignBalanced` branch: `slack = box.H − totalH`;
   when `slack > 0`, `unit = slack/(n+1)`; `startY = box.Y + unit ×
   balancedOpticalBP/10000`; `effectiveGap = gap + unit`.
3. Tests: a 3-node sparse stack under balanced mode has a non-zero top margin and
   widened gaps (the slack distributed, no single void), the stack stays in the
   box, and the optical bias keeps the top margin below the bottom;
   `VAlignTop`/`VAlignCenter` byte-identical; determinism guard; smoke `phase-46.sh`.

## 5. Open questions

- **Per-node gap weighting** (a literally larger gap before a chosen node) —
  deferred; it is content taste, and a caller can order nodes / insert a spacer.
- **Optical factor tuning** — `8500` is a pinned default; a future req could expose
  it, but the engine ships one deterministic value.
