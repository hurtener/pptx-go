# Brief 75 — Hierarchy / org-chart / tree (R14.10)

> Informs Phase 92 (Wave 14). Engine req R14.10
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase
Team/org slides, taxonomy trees, and layered architecture decompositions are a
standard class with no support — `Flow` is sequential, `Bento`/`Grid` are flat.

## 2. Findings
- **A new native node, tidy-tree layout.** `Tree{Root TreeNode; Orientation}`;
  `TreeNode{Label, Detail, Icon, Children, AccentIndex}`. A deterministic two-axis
  layout: leaves spread evenly across the cross-axis (`leafSlot`), depth on the
  main axis (`levelSlot`), internal nodes centered over their first/last leaf
  descendant — so parent→child edges don't cross. Top-down (`FlowVertical`,
  default) vs left-right (`FlowHorizontal`) just swap the axes.
- **Elbow connectors avoid diagonal-line flips.** Each parent→child edge is an
  out·mid-bus·in elbow (all horizontal/vertical `ShapeLine` segments), so no
  `flipV` is needed and the org-chart reads cleanly.
- **Bounded.** Node width derives from `leafSlot`/`levelSlot`, clamped to a max
  and floored (a too-wide tree warns + clamps rather than overflowing). Nodes are
  rounded-rect cards (accent border by `AccentIndex`, label + optional detail +
  optional curated icon). Native (no media); `nodeUsesAssets:false`.

## 3. Recommendations
- Node + composer + full wiring (`walkTreeIcons` recurses node icons for Stage-1).
  Tests: a 3-level tree (cards + edges, conformant), left-right, determinism; an
  adversarial deep/long-label tree. Glossary, THEME, visual-leaves, skill. D-127.

## 4. Open questions
- A full Reingold-Tilford tidy layout (subtree contour packing) → V1.x; the
  leaf-centroid layout satisfies the balanced/non-crossing acceptance.
