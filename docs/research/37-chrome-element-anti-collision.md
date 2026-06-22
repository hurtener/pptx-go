# Brief 37 — chrome-element-anti-collision

**Subsystem:** scene — Layer 2 renderer (card chrome)
**Authored:** 2026-06-22
**Motivating phase:** Phase 54 — chrome-element-anti-collision (R11.6, HIGH · engine)

## 1. Question

A card's header pill (right-aligned, top of the header row) and its status dot
(top-right corner) are positioned independently and both anchor to the same
top-right point, so when both are set they overlap — the dot sits on the pill's
right edge (recreation slide 9, "POPULAR" pill + dot). How can they be placed so
their boxes never intersect, byte-identically when only one is set?

## 2. Prior art surveyed

- **`scene/render_card.go renderCardChrome`** — the pill is drawn at
  `pillBox.X = innerX + innerW − pillW` (its right edge = `box.X + box.W − pad`); the
  status dot at `box.X + box.W − pad − cardStatusDotSz` (right edge also =
  `box.X + box.W − pad`). Same right edge, both at the top → overlap.
- **`cardPillWidthOf`** (R11.5/D-085) — the now-fitted pill width, available to
  compute the pill's left edge.
- DECKARD R11.6 spec: when both `c.pill != ""` and `c.statusDot != nil`, shift the
  dot left by `pillW + gap` (or above the pill), so the dot and pill boxes do not
  intersect; inert and byte-identical when only one is set.

## 3. Findings

- **Shift the dot left of the pill.** The pill already occupies the top of the
  header row; placing the dot *above* it would push it off the card top inset, so
  *left* is the correct axis. When both are set, set
  `dotX = pillX − gapSM − cardStatusDotSz` where `pillX = innerX + innerW −
  cardPillWidthOf(theme, pill, innerW)`. The dot's right edge is then `pillX − gapSM
  < pillX` (the pill's left edge), so the boxes are disjoint with a `gapSM`
  separation. A floor at `innerX` keeps the dot on-card for a pathologically wide
  pill.
- **Byte-identical when only one is set.** The shift is guarded by `c.pill != ""`;
  a dot-only card keeps the corner placement `box.X + box.W − pad −
  cardStatusDotSz`, so the existing rich-visuals goldens (which set the dot without a
  pill) are unchanged.
- **Disjointness is by construction, not chance.** Because the dot derives from the
  *same* `cardPillWidthOf` the pill is drawn with, `dot.Right() = pillX − gapSM`
  always sits a gap to the left of the pill's left edge — no overlap for any pill
  label length.
- **Scope: pill × dot only.** The watermark (bottom-anchored, low opacity, behind
  the body) is a different z-order/region concern handled by R11.11; the pill × dot
  pair is the reported top-right collision and the R11.6 acceptance.

## 4. Recommendation

In the status-dot block of `renderCardChrome`, compute `dotX` left of the pill when
both are set (floored at `innerX`), else keep the corner placement. Test: a
render-level check that the dot's x-offset with a pill is smaller than the dot-only
corner x (the shift fired), and that a dot-only card is unchanged. D-086 records the
anti-collision placement.

## 5. Open questions

- A long eyebrow could still reach under the shifted dot's left region; out of scope
  for R11.6 (its acceptance is pill ∩ dot = ∅). The header text column already
  reserves the pill width; the dot lives in that reserved gap.
