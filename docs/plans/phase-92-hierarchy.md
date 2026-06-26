# Phase 92 — hierarchy / org-chart / tree

**Subsystem:** `scene` (new IR node)
**RFC sections:** §11, §12, §10.1, §7.1
**Deps:** brief 75.
**Status:** Done

## 1. Goal
Add a `Tree` scene node — a node-link hierarchy (org chart / taxonomy / layered
decomposition) laid out as a balanced tidy tree with elbow edges.

## 2. Why now
Wave 14 coverage; the hierarchy class is uncovered (Flow is sequential). R14.10
(MED · engine; D-059).

## 3-6. RFC/brief/decisions
RFC §11/§12 (new native node), §10.1 (absent until used), §7.1 (token colors).
Brief 75. Decisions D-059, D-026, D-127 (new).

## 7. Architecture
`Tree{Root TreeNode; Orientation}`, `TreeNode{Label, Detail, Icon, Children,
AccentIndex}`. `renderTree`: a deterministic tidy layout (leaves spread on the
cross-axis, depth on the main-axis, internal nodes centered over leaf descendants);
parent→child elbow connectors (out·mid-bus·in, all H/V); node cards = rounded rect
+ accent border (timelineAccent) + label/detail/icon. Bounded (clamp + warn).
Top-down (FlowVertical) / left-right (FlowHorizontal) swap the axes.

## 8. Files
nodes.go (KindTree + Tree/TreeNode), policy.go, validate.go, render.go (dispatch +
preferredHeight + nodeUsesAssets), render_tree.go (NEW), render_card.go
(walkIconRefs case Tree + walkTreeIcons), scene_test.go (catalog 33),
render_tree_test.go (NEW), render_adversarial_test.go, test/integration (kind-loop
..KindTree + Tree on the button slide), scripts/smoke/phase-92.sh, docs/research/75
+ INDEX, this plan, README, THEME, glossary, decisions D-127,
docs/site/catalog/visual-leaves.md, skills/compose-a-scene.

## 9. Public API
`type Tree struct {...}` + `TreeNode`. Additive new node; no break.

## 10-11. Risks/acceptance
Crossing edges (mitigated: leaf-centroid layout + elbows; adversarial tree);
overflow (clamp + warn); determinism (integer-EMU; 1-vs-8 test). Accept: a 3-level
tree renders balanced cards + non-crossing elbow edges (conformant); left-right
works; worker-count deterministic.

## 12-14. Coverage/smoke/tests
scene 80%. scripts/smoke/phase-92.sh. Black-box: 3-level tree, left-right,
determinism; adversarial deep/long-label tree; integration all-kinds.
