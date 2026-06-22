# Brief 24 — content-weighted-bento-rows

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 41 — content-weighted bento rows (R10.3, HIGH · engine)

## 1. Question

`bentoGeometry` forces every row to the same height
`rowH = (box.H − gaps)/nRows`, regardless of content. A sparse one-line row and a
dense four-line row get identical height, so the sparse row wastes a huge band
while the dense row starves and overflows off-slide (the recreation's slide-6
"Canvas" card). How can the engine, **opt-in and deterministically**, size each
bento row to its content's preferred height — clamped so the total never exceeds
the region — while keeping the default equal-row layout byte-identical?

## 2. Prior art surveyed

- `scene/render_bento.go` `bentoGeometry(box, v, gap) → (gutterW, rowH, cells)`:
  a pure function computing the left-gutter width, a single equal `rowH`, and one
  `Box` per cell (row-major). `renderBento` independently recomputes
  `rowY = box.Y + ri*(rowH+gap)` for both the gutter label and the cells.
- `scene/render.go` `preferredHeight` — the deterministic per-node slot
  estimator; the bento case already takes the max cell preferred height across
  all rows (`nRows × maxCell`).
- Phase 40 (R10.2, D-071) `fitCompress` — the established integer/basis-point
  "scale toward fit" primitive (`h × sBP / 10000`), proven worker-count
  independent. R10.3 reuses the same scaling shape for the overflow clamp.
- Phase 27 (R5c, D-056) — the `Bento` node + the span-weighted column grid.
- DECKARD R10.3 spec: opt-in content-weighted row mode: per-row
  `h_i = max preferred height of that row's cells at their span widths`; if
  `Σh_i + gaps > box.H` scale all rows by the fit factor, else distribute the
  leftover slack (top-align or proportional). Default equal rows byte-identical.
  Gutter labels anchor-middle within their actual row height. Integer math.

## 3. Findings

- The fix is purely **vertical distribution** of the bento's slot among its rows;
  the horizontal (gutter + span column) geometry is unchanged. So the cleanest
  refactor factors the shared horizontal math (`bentoColumns` → gutterW,
  contentX, unitW; `cellWidth(span)`) out of `bentoGeometry`, then lets the row
  heights be either equal (default) or a caller-supplied per-row slice.
- Per-row preferred height = `max over the row's cells of preferredHeight(cell,
  cellWidth, theme)` — exactly the metric the bento estimator already uses
  globally, applied per row instead. It needs the theme, so the weighted heights
  are computed by a renderer method, while `bentoGeometry` stays a pure function
  that accepts the resulting `[]EMU`.
- **Overflow clamp must guarantee fit** (the acceptance demands
  `Σ + gaps ≤ box.H`). Unlike `fitCompress`'s 0.60 ratio floor (which may leave
  residual), the bento row clamp scales fully to fit:
  `sBP = avail·10000/Σpref`, `h_i = pref_i·sBP/10000` — flooring guarantees
  `Σ ≤ avail`. No ratio floor, because a bento with no slack must still place all
  rows inside the frame (the whole point of the req).
- **Under-full (fits):** leave rows at preferred height (top-aligned, leftover as
  bottom whitespace). This already satisfies "a sparse row does not steal space
  from a dense one." A proportional row-fill option is a future refinement; the
  spec lists top-align as an acceptable default.
- **Byte-identical default:** with `WeightedRows=false`, `bentoGeometry` uses the
  equal `rowH` for every row and accumulates `rowY` identically to today — the
  per-row-array refactor reproduces the exact same boxes and label Ys.
- **Gutter labels** already anchor-middle (`AnchorMiddle`); routing them through
  the per-row Y + per-row H makes them anchor-middle within the *actual* row
  height automatically.
- The slot **estimate** (`preferredHeight` for the bento as a whole) is left
  unchanged this phase — the internal row distribution operates on whatever
  `box.H` the bento is given. Estimator/actual parity for bento (wide-span cells,
  weighted rows) is R10.10's explicit job; touching it here would widen scope.
- The **Grid analog** named in the spec is a smaller, separable change: `Grid`
  cells are single children laid out by the pure `layout.Grid` (no theme); making
  Grid rows content-weighted means `renderGrid` building boxes itself. Out of
  scope for the bento-focused acceptance criterion; deferred (R10.10 / follow-up).

## 4. Recommendations

1. Add an additive `WeightedRows bool` to `Bento` (zero = equal rows, today's
   byte-identical layout).
2. Refactor `bentoGeometry` to `(box, v, gap, rowHs []EMU) → (gutterW, rowYs,
   rowHs, cells)`: `rowHs == nil` ⇒ equal mode; otherwise use the supplied
   per-row heights. Extract `bentoColumns` + `cellWidth` for the shared
   horizontal math.
3. Add `r.bentoWeightedRowHeights(box, v, gap) []EMU`: per-row max cell preferred
   height, clamped to fit via a single basis-point scale when `Σpref + gaps >
   box.H`.
4. `renderBento`: when `v.WeightedRows`, compute the weighted heights and pass
   them in; use the returned `rowYs[ri]`/`rowHs[ri]` for both labels and cells.
5. Tests: weighted dense-row-taller-than-sparse-and-fits; equal-mode
   byte-identical to today; determinism guard; smoke `phase-41.sh`.

## 5. Open questions

- **Proportional row-fill of leftover slack** (vs top-align) — deferred; the spec
  offers it as an option, top-align is the chosen default.
- **Grid content-weighted rows** — deferred (separable; `layout.Grid` is pure).
- **Estimator parity** for weighted bento slot height — R10.10.
