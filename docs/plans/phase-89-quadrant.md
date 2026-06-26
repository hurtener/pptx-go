# Phase 89 — 2x2 quadrant / positioning map

**Subsystem:** `scene` (new IR node)
**RFC sections:** §11, §12, §10.1, §7.1
**Deps:** brief 72.
**Status:** Done

## 1. Goal
Add a `Quadrant` scene node — a 2x2 positioning map with labeled axes, optional
per-quadrant tints/titles, and items plotted at (x,y) — native and deterministic.

## 2. Why now
Wave 14 coverage classes; the positioning/landscape class (effort/impact, BCG) is
uncovered. Engine req R14.9 (MED · engine; D-059).

## 3–6. RFC / brief / decisions
RFC §11/§12 (new native node, no media), §10.1 (absent until used), §7.1 (token
colors). Brief 72. Decisions: D-059, D-026 (engine plots; caller supplies x/y),
D-124 (new).

## 7. Architecture
`Quadrant{AxisX, AxisY QuadrantAxis; Quadrants [4]QuadrantCell; Items
[]QuadrantItem}`. `renderQuadrant`: a left gutter (Y labels) + bottom strip (X
labels); the field gets 4 low-alpha quadrant tints + titles, a center cross, and
item dots at (x, inverted-y) with edge-flipped/clamped labels (dot color via the
pinned `timelineAccent` cycle). Pure integer-EMU.

## 8. Files
nodes.go (KindQuadrant + Quadrant/QuadrantAxis/QuadrantCell/QuadrantItem),
policy.go, validate.go (coords in [0,1]), render.go (dispatch + preferredHeight +
nodeUsesAssets), render_quadrant.go (NEW composer), scene_test.go (catalog 31),
render_quadrant_test.go (NEW), render_adversarial_test.go, test/integration (kind
loop ..KindQuadrant + Quadrant on the button slide), scripts/smoke/phase-89.sh,
docs/research/72 + INDEX, this plan, README, glossary, docs/site/catalog/
visual-leaves.md, skills/compose-a-scene.

## 9. Public API
`type Quadrant struct {...}` + `QuadrantAxis`/`QuadrantCell`/`QuadrantItem`.
Additive new node; no break.

## 10–11. Risks / acceptance
Off-canvas labels (mitigated by edge-flip + clamp + the adversarial corner-item
slide); determinism (integer-EMU; 1-vs-8-worker test). Accept: a 2x2 with 6 items
renders axes + 4 tints + dots inside the safe area, conformant; an out-of-range
coord fails validation; worker-count deterministic.

## 12–14. Coverage / smoke / tests
scene 80%. `scripts/smoke/phase-89.sh`. Black-box: a 6-item quadrant (axes +
dividers + ≥6 dots + labels, conformant), invalid-coord validation, determinism;
adversarial corner-item quadrant; integration all-kinds fixture.
