# Phase 95 — scatter ornament family

**Subsystem:** `assets/ornaments` + `scene/ornaments`
**RFC sections:** §14.1 (ornaments), §10.1
**Deps:** D-110 (starfield); brief 78.
**Status:** Done

## 1. Goal
Restate the sample-specific `starfield` ornament as a general scatter/particle
family sharing one deterministic placement engine.

## 2. Why now
Wave 14 META req R14.20 (engine half): generalize the flagged starfield to its
class. (The pricing-card-recipe half is product.)

## 3-6. RFC/brief/decisions
RFC §14.1 (curated ornaments), §10.1 (starfield byte-identical). Brief 78.
Decisions D-059, D-110 (starfield), D-026, D-131 (new).

## 7. Architecture
Extract the starfield placement (lattice + hash-of-index + size/alpha variance)
into a shared `scatter(sl, box, alpha, role, pitch, shape scatterShape)` in
assets/ornaments. `Starfield` = `scatter(scatterDot)` (byte-identical). New
`ScatterDot/Star/Plus/Ring` pass a different shape (ellipse / star5 / mathPlus /
ring outline). Registered in Curated() (11) + isPatternPreset (Pitch cap). Recipe
signature unchanged.

## 8. Files
assets/ornaments/patterns.go (scatter engine + shapes + 4 recipes; Starfield
refactor), scene/ornaments/registry.go (names + Curated), scene/render_decoration.go
(isPatternPreset), scene/ornaments/registry_test.go (count 11),
render_scatter_test.go (NEW), scripts/smoke/phase-95.sh + phase-76.sh (test rename),
docs/research/78 + INDEX, this plan, README, glossary, decisions D-131,
skills/compose-a-scene.

## 9. Public API
New curated ornament names: scatter_dot/_star/_plus/_ring. Additive; starfield
unchanged.

## 10-11. Risks/acceptance
Starfield byte-identity (scatter+scatterDot reproduces it; pinned by a test);
determinism. Accept: the family renders a starfield AND a dust/dot/plus/ring field
from one engine with different shapes, each byte-identical across renders;
starfield == scatter_dot.

## 12-14. Coverage/smoke/tests
scene/ornaments 80%. scripts/smoke/phase-95.sh. Tests: family shapes, starfield
byte-identity, determinism, curated count 11.
