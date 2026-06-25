# Brief 66 — Styled table / comparison matrix (R14.3)

> Informs Phase 83 (Wave 14). Engine req R14.3
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · both — engine half; D-059).

## 1. Motivating phase

A feature-by-plan comparison matrix is the single most common enterprise/SaaS
slide, but the scene `Table` is plain RichText cells with no visual styling — no
header band, zebra, highlighted column, row labels, or grouped headers. Phase 83
adds an additive `TableStyle` so any data table reads as a designed matrix.

## 2. Subsystem / files

- `scene/nodes.go` — the `Table` node.
- `scene/render_table.go` — `renderTable` / `fillCell` (the native `a:tbl` path).
- `scene/render.go` — `tableHeight` (the slot estimate).
- `pptx/table.go` — the builder `Table`/`Cell` API (`SetFill`, `SetBorders`,
  `MergeRight`, `SetBanding`, `applyStyling`).
- `scene/contrast.go` — `onCardSurface` (auto-contrast text, D-082).

## 3. Findings

- **The native table builder already exposes every styling hook.** `Cell.SetFill`
  (token-resolved), `Cell.SetBorders`, `Cell.MergeRight` (grouped header spans),
  and `SetColumnWidths` cover header band, zebra, highlighted column, row labels,
  and grouped headers without any new builder capability (P1) — the scene
  renderer composes them.
- **`applyStyling` overwrites cell fills, so the styled path must avoid it.**
  `SetHeaderRow`/`SetBanding` call `applyStyling`, which sets header/odd-row fills
  *after* the call — they would clobber explicit `SetFill`s. So the styled path
  sets every cell fill itself and does **not** call `SetHeaderRow`/`SetBanding`;
  the plain (nil-Style) path keeps the existing `SetHeaderRow().SetBanding()`
  verbatim → byte-identical.
- **Auto-contrast is reusable for the header band.** `onCardSurface(role)` returns
  the inverse text token on a dark fill (else nil); a `cellTextOn` wrapper falls
  back to `TextPrimary`, so header-band / accent-column text contrasts (D-082).
- **CellKind glyphs are *not* a native-table feature.** A native OOXML table cell
  (`a:tc`) holds only a text body — no `pic`/shape children — so check / cross /
  dot / **mini-bar** glyphs cannot be embedded as native shapes, and font glyphs
  would reintroduce the empty-box risk D-095 fixed (it chose `custGeom` over font
  checkboxes). The comparison-matrix-with-glyphs use case is already served by a
  `Bento` of `Checklist` / `IconRows` cells (the glyph nodes shipped in D-095 /
  D-100; the requirement's own gap text notes ref-07 renders its matrix as a
  Bento). So Phase 83 ships `TableStyle` and documents CellKind as composed, not a
  native-table field.
- **Grouped header adds one row; the estimate must account for it.**
  `tableHeight` adds one `In(0.4)` row when `HeaderGroups` is non-empty so the slot
  is allocated enough height and the overflow clamp (R11.3) stays truthful.
- **Graceful degradation, no hard validation.** An out-of-range `HighlightCol`
  simply never matches a column; `HeaderGroups` spans are clamped to the column
  count. No Stage-1 validation needed (RFC §10.2 / D-026).

## 4. Recommendations

- Scene: `Table.Style *TableStyle{HeaderFill, Zebra bool; HighlightCol int;
  RowLabelCol bool; HeaderGroups []HeaderGroup{Label; Span}}`; nil = plain.
- `renderStyledTable` controls every fill: header band (accent + contrast text),
  zebra (SurfaceAlt on odd body rows), highlighted column (accent tint +
  heavier accent border), row labels (SurfaceAlt + bold col 0), grouped header
  (MergeRight per span, accent band). `tableHeight += group row`.
- Tests: styled emit (accent + surfaceAlt fills, conformant), grouped header
  (`gridSpan="3"`), nil byte-identical, determinism; an adversarial matrix slide.
  THEME.md note, glossary, compose-a-scene skill, docs/site scene.md. D-118.

## 5. Open questions

- CellKind glyphs (check/cross/dot/bar) → composed with `Bento`+`Checklist`
  today; a future shape-grid "matrix" node could embed them natively if demand
  warrants (not a native `a:tbl`).
- Per-cell fills / column widths from content → a follow-up (the builder
  `SetColumnWidths` is available; auto-sizing columns to content is V1.x).
