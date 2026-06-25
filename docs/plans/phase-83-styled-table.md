# Phase 83 — styled table / comparison matrix

**Subsystem:** `scene` (Table node + renderer)
**RFC sections:** §11.1 (table), §12 (per-node policy), §10.1 (backward-compat), §7.1 (token colors)
**Deps:** D-082 (auto-contrast), D-095/D-100 (glyph nodes); brief 66.
**Status:** Done

---

## 1. Goal

Add comparison-matrix styling to the scene `Table` — a header band, zebra
striping, a highlighted column, an emphasized row-label column, and grouped
header spans — all from theme tokens, additive and byte-identical when unused.

## 2. Why now

Wave 14 coverage classes (`docs/plans/README.md`); a feature×plan comparison
matrix is the most common enterprise/SaaS slide and the plain `Table` can't style
it. Engine req R14.3 (HIGH · both; engine half per D-059).

## 3. RFC sections implemented

- `RFC §11.1` — the `Table` gains designed styling over the native `a:tbl`.
- `RFC §12` — the styled table is native shapes (a `tbl`), not media.
- `RFC §10.1` — a nil `Style` is byte-identical to the plain banded table.
- `RFC §7.1` — every fill/border resolves from theme tokens (P2).

## 4. Brief findings incorporated

- `docs/research/66-styled-table-matrix.md` — *"the native table builder already
  exposes every styling hook"* → compose `SetFill`/`SetBorders`/`MergeRight`; no
  new builder capability (P1).
- `66` — *"`applyStyling` overwrites cell fills, so the styled path must avoid
  it"* → the styled path sets fills explicitly and skips `SetHeaderRow`/
  `SetBanding`; the nil-Style path keeps them verbatim (byte-identical).
- `66` — *"auto-contrast is reusable for the header band"* → `cellTextOn` wraps
  `onCardSurface` (D-082).
- `66` — *"CellKind glyphs are not a native-table feature"* → documented as
  composed with `Bento`+`Checklist`/`IconRows`, not a `Table` field.
- `66` — *"grouped header adds one row; the estimate must account for it"* →
  `tableHeight += In(0.4)` for the group row.

## 5. Findings I'm departing from

- **CellKind (check / cross / dot / mini-bar cells)** is *not* added to `Table`. A
  native OOXML table cell (`a:tc`) holds only a text body — no shape children — so
  the glyphs can't be embedded as native shapes, and font glyphs would reintroduce
  the empty-box risk D-095 fixed. The matrix-with-glyphs use case is already
  served by a `Bento` of `Checklist` / `IconRows` cells (the requirement's gap
  text itself notes ref-07 renders its matrix as a Bento). Documented (D-118).

## 6. Decisions referenced

- `D-059` — engine extension; engine half of a `both` req.
- `D-082` — the `onCardSurface` auto-contrast mechanism (header-band text).
- `D-095` / `D-100` — the `Checklist` / `IconRows` glyph nodes (the CellKind path).
- `D-026` — the engine exposes the styling mechanism; the soul drives it.
- `D-118` (new) — files `TableStyle` + the CellKind decision.

## 7. Architecture

`Table.Style *TableStyle{HeaderFill, Zebra bool; HighlightCol int; RowLabelCol
bool; HeaderGroups []HeaderGroup{Label; Span}}`. `renderTable` branches: a nil
`Style` runs the existing plain path (`SetHeaderRow().SetBanding()`, byte-
identical); a non-nil `Style` runs `renderStyledTable`, which sets every cell fill
explicitly (no builder banding) — a grouped header row (`MergeRight` per span,
accent band), a header band (accent + `cellTextOn` contrast text), body rows with
`styleBodyCell` (highlight column → accent tint + heavier accent border; row-label
column → SurfaceAlt + bold; zebra → SurfaceAlt on odd rows). `tableHeight` adds a
row for the grouped header so the slot estimate stays truthful.

```text
Table{Headers, Rows, Style:{HeaderFill, Zebra, HighlightCol:3, RowLabelCol,
                            HeaderGroups:[{Plan,1},{Paid,3}]}}
  → group row (a:gridSpan="3") + accent header band + zebra + accent-tinted col 3
Table{Headers, Rows}  (nil Style) → plain SetHeaderRow + SetBanding (byte-identical)
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — Table.Style + TableStyle + HeaderGroup
scene/render_table.go                # CHANGED — renderStyledTable / styleBodyCell / fillCellColored / cellTextOn
scene/render.go                      # CHANGED — tableHeight accounts for the grouped header row
scene/render_table_styled_test.go    # NEW — styled emit, grouped header, nil byte-identical, determinism
scene/render_adversarial_test.go     # CHANGED — a styled matrix slide in the torture fixture
scripts/smoke/phase-83.sh            # NEW — phase smoke
docs/research/66-styled-table-matrix.md  # NEW — brief
docs/research/INDEX.md               # CHANGED — registers brief 66
docs/plans/phase-83-styled-table.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — Phase 83 detail
docs/design/THEME.md                 # CHANGED — table styling mechanism note
docs/glossary.md                     # CHANGED — comparison matrix / table style term
docs/decisions.md                    # CHANGED — adds D-118
docs/site/reference/scene.md         # CHANGED — Table.Style / TableStyle
skills/compose-a-scene/SKILL.md      # CHANGED — Table.Style
```

## 9. Public API surface

```go
// scene
type TableStyle struct { HeaderFill, Zebra bool; HighlightCol int; RowLabelCol bool; HeaderGroups []HeaderGroup }
type HeaderGroup struct { Label string; Span int }
// Table gains: Style *TableStyle
```

Additive; no break.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** nil `Style` runs the unchanged plain
  path; a byte-identity test pins it.
- **R2 — `applyStyling` clobbering explicit fills.** **Mitigation:** the styled
  path never calls `SetHeaderRow`/`SetBanding`; an emit test asserts the accent +
  surfaceAlt fills are present.
- **R3 — determinism.** **Mitigation:** a 1-vs-8-worker test asserts byte-identity.
- **R4 — grouped-header overflow.** **Mitigation:** `tableHeight` adds the group
  row; the R11.3 clamp caps an over-tall table.

## 11. Acceptance criteria

1. A styled matrix renders a header band, zebra, a highlighted column, and row
   labels in theme tokens (accent + surfaceAlt fills present); conformant; no
   warnings.
2. A grouped header row merges its spans (`gridSpan`) and labels them.
3. A `Table` with a nil `Style` is byte-identical to the pre-Phase-83 build.
4. The styled table is worker-count deterministic.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | styled table composer |

## 13. Smoke check

`scripts/smoke/phase-83.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `Table.Style` / `TableStyle` / `HeaderGroup` / `renderStyledTable` /
   `styleBodyCell` present.
3. `OK:` styled emit / grouped header / nil byte-identical / determinism tests.

## 14. Tests

- **Black-box (`scene_test`):** a styled matrix emits the accent + surfaceAlt
  fills and is conformant; a grouped header merges spans; a nil `Style` is
  byte-identical; the styled table is worker-count deterministic.
- **Adversarial:** a fully-styled matrix with long wrapping cells in the fixture.
- **Integration / Fuzz:** no (no new node; `Table` is already integration-covered).
