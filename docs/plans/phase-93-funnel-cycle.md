# Phase 93 — funnel + cycle non-linear process diagrams

**Subsystem:** `scene` (two new IR nodes)
**RFC sections:** §11, §12, §10.1, §7.1
**Deps:** D-127 (Tree, for branch); brief 76.
**Status:** Done

## 1. Goal
Add `Funnel` (tapering stages) and `Cycle` (ring of stages with arrows) nodes;
branch (1→M) is covered by the `Tree` node.

## 2. Why now
Wave 14 coverage; non-linear process archetypes (funnel/cycle/branch) are common
and uncovered by linear `Flow`. R14.11 (MED · engine; D-059).

## 3-6. RFC/brief/decisions
RFC §11/§12 (new native nodes), §10.1, §7.1. Brief 76. Decisions D-059, D-026,
D-128 (new).

## 7. Architecture
`Funnel{Stages []FunnelStage{Label, Value, AccentIndex}}` → stacked centered
rounded-rect bands of linearly decreasing width (floored at funnelMinFrac), each
with a centered contrast label + value. `Cycle{Stages []CycleStage{Label, Icon,
AccentIndex}}` → stage cards placed clockwise from the top on a ring (trig →
round to int EMU), consecutive (incl. last→first) connected by a flip-aware
straight line + a chevron arrowhead rotated to the chord. Branch = Tree (D-127).
Native; cycle icons recurse in walkIconRefs.

## 8. Files
nodes.go (KindFunnel/KindCycle + Funnel/Cycle + stage structs), policy.go,
validate.go, render.go (dispatch + preferredHeight + nodeUsesAssets), render_card.go
(walkIconRefs case Cycle), render_funnel_cycle.go (NEW), scene_test.go (catalog 35),
render_funnel_cycle_test.go (NEW), render_adversarial_test.go, test/integration
(kind-loop ..KindCycle + Funnel/Cycle on the button slide), scripts/smoke/phase-93.sh,
docs/research/76 + INDEX, this plan, README, THEME, glossary, decisions D-128,
docs/site/catalog/visual-leaves.md, skills/compose-a-scene.

## 9. Public API
`type Funnel struct {...}` + `FunnelStage`; `type Cycle struct {...}` + `CycleStage`.
Additive; no break.

## 10-11. Risks/acceptance
Off-canvas (funnel bands centered; cycle squared box; adversarial slide);
determinism (trig → round, 1-vs-8 test). Accept: a 4-stage funnel tapers + labels
values (conformant); a 5-stage cycle places cards + chevron arrows (conformant);
empty fails; worker-count deterministic.

## 12-14. Coverage/smoke/tests
scene 80%. scripts/smoke/phase-93.sh. Black-box: funnel, cycle, empty-fails,
determinism; adversarial funnel+cycle; integration all-kinds.
