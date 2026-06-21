# Brief 14 — bento-grid

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-20
**Motivating phase:** Phase 27 — row-labeled bento grid

## 1. Question

Reference decks use a **row-labeled bento grid**: rows that each carry a left
label and a sequence of cells with **variable column spans** (a wide cell next to
two narrow ones, the next row split differently — all aligned to a shared column
grid). The existing `Grid` is uniform: N equal columns, one child per cell, no
row labels, no spans. How can the engine add a bento layout — per-row labels +
variable spans against a shared column grid — additively, without disturbing
`Grid`/`TwoColumn`?

This is sub-unit (c) of `DECKARD-PRODUCT-REQUIREMENTS.md` R5; sub-units (a)+(b)
(the TwoColumn column join) shipped in the prior phase.

## 2. Prior art surveyed

- **`scene/Grid` + `layout.Grid`.** `Grid.Ratio` is **per-column** weights, not
  per-row, and `Grid` places exactly one child per uniform cell. The requirement's
  hint ("Ratio per row already partly exists — extend") doesn't hold: a bento with
  per-row labels and per-row variable spans is a structurally different layout, so
  it warrants a **new node** rather than overloading `Grid`.
- **`layout.Columns`.** Splits a box into weighted columns — directly usable to
  lay a row's cells by span, but it always *fills* the row (proportional), which
  loses cross-row column alignment. A 12-column-style **absolute** span model
  (a fixed unit width, a span-S cell = S units) keeps columns aligned across
  rows, matching the reference bento. So the renderer computes a unit width from
  the column count and lays cells left-to-right by absolute span.
- **Card / chrome label idioms.** A left-gutter label is a vertically-centered
  caption run (`AnchorMiddle`, `TextMuted`), the same idiom the card eyebrow and
  chrome footer use. No new primitive.
- **The container wiring checklist.** A new IR node must touch a fixed set of
  switches: `NodeKind` + `String`, `policyTable`, `validateNode`, `renderNode`,
  `preferredHeight`, `isFlexible`, `nodeUsesAssets`, and the four `walk*`
  recursions (icons, images, decorations, plus the round-trip `collectKinds`).
  A `cellNodes()` helper that flattens a bento's cells makes every recursion a
  one-liner and keeps them from drifting.
- **D-026 (engine, not product).** The caller supplies the labels, the spans, and
  the cell content; the engine lays them out on a shared column grid. It makes no
  decision about *what* belongs in a bento.

## 3. Findings

- **A new `Bento` node is the right shape.** `Bento{Columns, Rows}` where each
  `BentoRow{Label, Cells}` and each `BentoCell{Span, Node}`. `Columns` is the
  shared column-unit count; a cell of span S occupies S units; a row's spans sum
  to ≤ `Columns` (ragged rows allowed). This expresses the reference directly and
  leaves `Grid`/`TwoColumn` untouched.
- **Absolute spans keep columns aligned.** Compute `unitW` once from `Columns`
  and the content width; a span-S cell is `S·unitW + (S−1)·gap`. A span-1 cell is
  always `unitW`, so columns line up across rows — the bento look. (Proportional
  per-row would not align.)
- **The gutter is reserved only when used.** Reserve the left label gutter only
  when at least one row has a non-empty `Label`; otherwise the bento is a pure
  span-grid with the full width. This keeps a label-less bento from wasting space
  and a labelled one from overlapping content.
- **Equal-height rows, deterministic.** Rows split the box height equally (like
  `layout.Grid`); all geometry is integer EMU, so the bento is deterministic and
  parallel-safe (it registers no media itself — only an asset-bearing cell does).
- **Stage-1 validation is the safety net.** `Columns ≥ 1`, non-empty rows, each
  row non-empty, each `Span ≥ 1`, each cell node non-nil, and a row's spans ≤
  `Columns` — surfaced as joined errors, recursing into cell nodes.

## 4. Recommendations

1. Add `Bento` / `BentoRow` / `BentoCell` + `KindBento` (+ `String`), a
   `cellNodes()` helper, and a `policyTable` entry (`{}` — native container).
2. Wire the new node through every switch: `validateNode`, `renderNode`,
   `preferredHeight`, `isFlexible` (a bento is flexible — grows under
   `VAlignFill`), `nodeUsesAssets`, and the `walk*` recursions, using
   `cellNodes()`.
3. `renderBento`: reserve the gutter iff any row is labelled; equal-height rows;
   per-row absolute-span cell layout via a shared `unitW`; left labels
   vertically centered.
4. Extend the catalog count (20 → 21 kinds) and the round-trip `everyNodeScene`
   / `collectKinds` / kind-range loop so the "every node round-trips" guard
   covers `Bento`.

## 5. Open questions

- **Content-height rows.** Rows are equal-height; a content-driven row height
  (tall row for a big cell, short for a label-only row) is a refinement, deferred
  — equal rows match the reference and compose with grow-to-fit (R2).
- **Cell alignment within a span / row-spanning cells.** Cells don't span
  multiple *rows* (no rowspan), and a cell fills its span box. Rowspan and
  per-cell vertical alignment are deferred; not in R5's acceptance.
- **Gutter on the right / labels per column.** Only a left row-label gutter is
  modeled (the reference). Other label placements are deferred.
