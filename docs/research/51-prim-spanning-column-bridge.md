# Brief 51 — prim-spanning-column-bridge

**Subsystem:** scene — Layer 2 renderer (TwoColumn field extension)
**Authored:** 2026-06-23
**Motivating phase:** Phase 68 — prim-spanning-column-bridge (R12.8, MED · engine)

## 1. Question

An option / path slide ("One agent, purpose-built — two ways to get it") wants a labeled
connector *spanning the tops* of two columns, signaling they share a root. `TwoColumn.Join`
(D-055) only places an element ON the vertical seam (center); the recreation collapsed the
bridge into a tiny circle with "One age nt" wrapped mid-word. What extends the join to a
horizontal spanning bridge?

## 2. Prior art surveyed

- **`scene/render_container.go renderColumnJoin`** — the D-055 seam element (badge/arrow
  centered on the gap). The bridge is the same `Join`/`JoinLabel` data drawn at the top /
  bottom edge instead, with a content-fit pill (reusing the badge's fit-to-label idea).
- **`scene/render_card.go` ribbon top-bar (Phase 65)** — the band-reserve pattern: a band
  at the edge shifts the content; `ribbonReserveOf` threaded through the geometry. The
  bridge reserves a band so the bracket sits above (or below) the columns.
- **`scene/metrics.go fitScale`** — the pill label single-line guarantee (no mid-word wrap).
- **D-059:** R12.8 is `engine`.

## 3. Findings

- **A `TwoColumn.JoinPosition` field, not a new node.** `JoinSeam` (zero) is today's
  centered element — byte-identical (the existing column-join tests pass). `JoinTopBridge`
  / `JoinBottomBridge` draw the bracket.
- **The bridge reserves a band.** Like the ribbon top-bar, a `bridgeBandH` band at the top
  (or bottom) edge insets the column layout box, so the bracket spans above (below) the
  columns without overlap; `preferredHeight` adds the band so the slot grows.
- **The bracket = a spanning line + two end stubs + a content-fit label pill.** The line
  runs from the left column's left edge to the right column's right edge; short stubs at
  each end reach toward the columns; the `JoinLabel` is a `RadiusFull` accent pill centered
  on the line, sized to the label (no mid-word wrap — `fitScale` shrinks only if the label
  would exceed the span). Reuses the flow stub/line primitives (plain accent rects).
- **Colors are tokens; metrics pinned.** The accent fill is a token; the band height, stub
  length, stroke, and pill padding are pinned EMU.

## 4. Recommendation

Extend `TwoColumn` with `JoinPosition JoinPosition` (`JoinSeam`/`JoinTopBridge`/
`JoinBottomBridge`) and a `renderColumnBridge` that reserves a band, lays the columns in the
remaining region, and draws the bracket (line + 2 stubs + a content-fit centered label
pill). Validate the position range; add the band to `preferredHeight`. `JoinSeam` (zero) is
byte-identical. Extend the R11.12 adversarial fixture with a top-bridge TwoColumn under a
long hostile label so the on-canvas + no-mid-word-wrap invariants are covered.
