# Brief 72 — 2x2 quadrant / positioning map (R14.9)

> Informs Phase 89 (Wave 14). Engine req R14.9
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase

Investor/strategy decks routinely use a 2x2 positioning map (competitive
landscape, effort/impact, BCG-style). Bento/Grid give equal cells but nothing
draws labeled axes and positions items by (x,y). Phase 89 adds a `Quadrant` node.

## 2. Findings

- **A new native node.** Axes, center dividers, per-quadrant tints, item dots, and
  labels all compose from preset rects/lines/ellipses + text — no media
  (`nodeUsesAssets:false`, `HasAsset:false`). Catalog 30 → 31.
- **Coordinates are fractions** in `[0,1]` with the origin bottom-left; the
  renderer inverts Y to screen space. Item labels flip to the dot's left near the
  right edge and clamp into the field, so they stay on-canvas (the timeline
  stagger/clamp pattern). Item dot colors reuse `timelineAccent` (a pinned token
  cycle, P2).
- **Layout.** A left gutter holds the Y end labels, a bottom strip the X end
  labels; the field is the remainder. Per-quadrant tints are low-alpha; the cross
  dividers are SurfaceAlt. Pure integer-EMU → byte-identical / deterministic.

## 3. Recommendations

- Node: `Quadrant{AxisX, AxisY QuadrantAxis{LowLabel, HighLabel}; Quadrants
  [4]QuadrantCell{Title; Fill *ColorRole}; Items []QuadrantItem{X, Y float64;
  Label; AccentIndex}}`. Full new-node wiring; `validate` rejects out-of-[0,1]
  coordinates. Tests: a 6-item 2x2 (axes + dividers + dots + labels, conformant),
  invalid-coord validation, determinism; an adversarial corner-item quadrant.
  Glossary, visual-leaves, skill. D-124.

## 4. Open questions

- An NxM grid variant and item anti-collision beyond edge-flip → V1.x (the 2x2
  with edge-flipped labels satisfies the acceptance).
