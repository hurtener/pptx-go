# Brief 49 — prim-inter-column-connectors

**Subsystem:** scene — Layer 2 renderer (Grid field extension)
**Authored:** 2026-06-23
**Motivating phase:** Phase 66 — prim-inter-column-connectors (R12.4, HIGH · engine)

## 1. Question

A 3-column architecture slide (people → operating layer → knowledge) reads as data flow
only when connector arrows join the columns in the gutters. The recreation's plain `Grid`
left the cards floating disconnected; `TwoColumn.Join` (D-055) only places a single
centered seam element, never an N-column gutter glyph. What field lets a `Grid` draw
connectors between adjacent columns?

## 2. Prior art surveyed

- **`scene/render_flow.go renderConnector(ps, gap, kind, vertical)`** — already draws an
  inter-step glyph (`ConnectorArrow`/`ArrowDashed`/`Cycle`/`Plus`) centered in a gap box.
  The grid connector reuses it verbatim, passing the derived gutter box.
- **`scene/render_container.go renderGrid` + `layout.Grid`** — the grid already computes a
  per-cell box slice (row-major); the gutter between columns `c` and `c+1` is derivable
  from the first-row cells' edges.
- **`TwoColumn.Join` (D-055)** — the single-seam precedent; R12.8 will extend *that* for a
  spanning bridge, while R12.4 is the N-column gutter case.
- **D-059:** R12.4 is `engine`. The field + render is the requirement.

## 3. Findings

- **`Grid.Connectors []GridConnector` — a field extension, not a new node.** No catalog
  change; empty ⇒ byte-identical (the existing grid tests pass). `GridConnector{Between
  [2]int; Kind ConnectorKind; Label string}`.
- **The gutter box is derived from the cell boxes.** For `Between {c, c+1}`, the gutter is
  `{X: cells[c].Right(), W: cells[c+1].X − cells[c].Right(), Y: box.Y, H: box.H}` — the
  vertical gutter spanning the grid height. `renderConnector` centers the glyph there.
  Pure integer geometry from the deterministic `layout.Grid` output.
- **`ConnectorBiArrow` is the one new glyph.** Reuse the Flow connector set and add a
  bidirectional arrow: `leftRightArrow` (horizontal) / `upDownArrow` (vertical) preset
  geometry — a `prst` attribute on the already-registered `prstGeom` element, so no
  `restorenamespaces` change.
- **Adjacency is validated at Stage-1.** `Between[1] == Between[0]+1` and both in
  `0..Columns−1`; the kind in range. (Grid already requires `Columns` in 2..4 and a full
  first row, so the referenced cells always exist.)
- **The optional `Label` sits below the glyph** in the lower third of the gutter (a muted
  `TypeCaption`); the glyph centers in the upper part. Gutters are narrow, so labels are
  best kept short — documented.

## 4. Recommendation

Add `Grid.Connectors []GridConnector` + `ConnectorBiArrow`, drawn by a
`renderGridConnectors` helper that derives each gutter box from the cell layout and calls
the existing `renderConnector`. Validate adjacency/range/kind at Stage-1. Field extension
— no catalog/kind change, no `walkIconRefs`. Empty ⇒ byte-identical. Extend the R11.12
adversarial fixture with connectors on the 3-column cards grid so the gutter glyphs are
covered by the on-canvas invariant.
