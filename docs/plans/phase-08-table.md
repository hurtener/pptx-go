# Phase 08 — Table

**Subsystem:** pptx (table builder) + scene (table node)
**RFC sections:** §8.5, §11.1 (table), §12 (table → native `tbl`)
**Deps:** Phase 04 (rich-text cells), Phase 07 (containers — tables-in-grids),
Phase 03 (the `graphicFrame`/`tbl` codec, left best-effort).
**Status:** In progress

---

## 1. Goal

A first-class table: a `pptx.Table` builder (header row, banding, merged cells,
cell fills/borders, rich-text cells) over a hardened `tbl` codec, plus the
`scene` composer for the `table` IR node with a caption above it.

## 2. Why now

Tables are a core deck primitive and the last leaf the scene renderer is missing
before Wave 3. Phase 03 shipped a best-effort `tbl`/`graphicFrame` codec (with
`a:xfrm` where PowerPoint wants `p:xfrm`) and deferred "full table namespace
fidelity … when tables are formally shipped." This is that phase. It also closes
the Phase 07 dependency (a table can live in a grid cell).

## 3. RFC sections implemented

- `RFC §8.5` — the builder table API exactly as written: `slide.AddTable(box,
  rows, cols)`, `SetHeaderRow`, `SetBanding(rowBand, colBand)`, `Cell(r,c)`,
  `cell.MergeRight(n)` / `MergeDown(n)`, `cell.SetFill(Fill)`, cell borders via
  `pptx.Line`, `cell.TextFrame()`.
- `RFC §11.1` / `RFC §12` — the `table` IR node renders **native** (`tbl`), one
  `RichText` per cell; a header row when `Headers` is non-empty; a caption is a
  separate `TextFrame` above the table (the scene renderer composes it).
- Hardens the `internal/ooxml/slide` `tbl`/`graphicFrame` codec (the Phase 03
  deviation): `p:xfrm` on the graphic frame, `tblPr`, `tcPr` (fill, borders,
  margins, anchor), and `hMerge`/`vMerge` continuation cells — all round-tripped.

## 4. Brief findings incorporated

No informing brief — the table API is specified directly in RFC §8.5;
`docs/research/INDEX.md` lists a "table merged-cell semantics" survey only as a
candidate. The OOXML `tbl` shape is vendored knowledge in the codec.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-026` — engine, not product: banding/header are concrete emitted fills the
  caller asked for, no legibility opinions.
- `D-032` — one emission path; the graphic-frame `p:xfrm` fix extends the
  `RestoreNamespaces` rewrite mechanism (a write-side element rename, mirroring
  the `rid→r:id` attribute rewrite) rather than a second writer.
- `D-033` — cell fills/borders compose the sealed `Color`/`Fill`/`Line` model and
  resolve against the active theme.

## 7. Architecture

```text
internal/ooxml/slide/slide_types.go  tblPr, tcPr, hMerge/vMerge, row height;
                                     XTableCellProps with fill/borders/anchor
internal/ooxml/slide/slide_marshal.go XGraphicFrame.MarshalXML → p:xfrm (read
                                     stays struct-tag "xfrm"; write emits "pxfrm")
internal/ooxml/restorenamespaces.go  elementLocal rewrite: pxfrm → p:xfrm;
                                     tblPr/tcPr/lnL.. → a:
pptx/table.go                        Table, Cell builder (RFC §8.5)
pptx/slide.go                        AddTable(box, rows, cols) *Table (replaces
                                     the inherited px AddTable/SetTableCellText)
scene/nodes.go                       Table gains a Caption field (additive)
scene/render_table.go                composes the tbl + optional caption frame
scene/render.go                      Table added to dispatch + preferredHeight
```

**`p:xfrm` fix.** Go's unmarshal is context-aware (it matches `xfrm` inside
`graphicFrame` to that field), but `RestoreNamespaces` is context-free and maps
`xfrm → a:` globally. So the fix is write-only: `XGraphicFrame.MarshalXML` emits
the transform under a sentinel local name `pxfrm`; `RestoreNamespaces` rewrites
`pxfrm → p:xfrm` (an element rename, like the `rid → r:id` attribute rename). The
struct keeps the `xfrm` tag, so read (default unmarshal) is unchanged and
round-trips.

**Builder.** `AddTable` builds the grid (equal column widths from the box) and
empty cells. `Cell(r,c)` returns a handle; `TextFrame()` returns a `pptx.TextFrame`
over the cell's `txBody` (reusing the Phase 04 model). `MergeRight(n)`/`MergeDown(n)`
set `gridSpan`/`rowSpan` on the anchor cell and `hMerge`/`vMerge` on the covered
cells. `SetFill`/borders write `tcPr`, resolving colors against the slide's theme
(the Chunk B pattern). `SetHeaderRow`/`SetBanding` apply concrete fills (header
fill + alternating row fills) so they render without a table-style part, and set
the `tblPr` flags for fidelity.

**Scene.** `render_table` builds the table from `Headers` + `Rows` of `RichText`
(header row present ⇒ `SetHeaderRow(true)` + banding), maps each cell's `RichText`
through the existing `addRichText`, and — when `Caption` is set — composes a
caption `TextFrame` above the table.

## 8. Files added or changed

Per §7. Plus `scripts/smoke/phase-08.sh`, `docs/glossary.md` (Table/Cell builder
terms if useful), `CHANGELOG.md`, and test files. The inherited `Slide.AddTable(x,
y,cx,cy,rows,cols)` / `SetTableCellText` are replaced (pre-V1) by the Box-based
`AddTable` + `Cell`; the one affected inherited test is updated.

## 9. Public API surface

```go
func (s *Slide) AddTable(box Box, rows, cols int) *Table   // replaces the px form
type Table struct{ ... }
func (t *Table) SetHeaderRow(on bool) *Table
func (t *Table) SetBanding(rowBand, colBand bool) *Table
func (t *Table) Cell(row, col int) *Cell
func (t *Table) SetColumnWidths(emu ...EMU) *Table
type Cell struct{ ... }
func (c *Cell) TextFrame() *TextFrame
func (c *Cell) SetText(text string) *Cell
func (c *Cell) SetFill(f Fill) *Cell
func (c *Cell) SetBorders(line Line) *Cell
func (c *Cell) MergeRight(n int) *Cell
func (c *Cell) MergeDown(n int) *Cell
// scene
type Table struct { ...; Caption string }   // Caption is new (additive)
```

`Slide.AddTable` and `SetTableCellText` change shape — documented in
`CHANGELOG.md` (v0.x).

## 10. Risks

- **R1 — `p:xfrm` rewrite correctness.** *Mitigation:* a write-only element
  rename keyed on a sentinel name; a codec golden asserts `<p:xfrm>` on a table
  and a round-trip proves read is unaffected.
- **R2 — merged-cell round-trip.** `hMerge`/`vMerge` continuation cells must
  survive. *Mitigation:* a merged-table round-trip golden (write → Open → assert
  spans + continuation flags).
- **R3 — banding without a table style.** Built-in style banding needs the app's
  style library. *Mitigation:* emit **explicit** alternating cell fills
  (deterministic, visible, asserted in XML); the `tblPr` flags carry intent.
- **R4 — inherited table-test break.** *Mitigation:* update the one
  `presentation_test.go` table test to the new API in this PR.

## 11. Acceptance criteria

1. A table with merged cells (a `MergeRight`/`MergeDown`) round-trips losslessly
   through `pptx.Open` (spans + `hMerge`/`vMerge` preserved).
2. A banded table alternates row fills correctly (distinct fills on odd/even
   body rows in the emitted XML).
3. A scene `table` node with a caption renders the `tbl` plus a caption text
   shape positioned above it; the deck is conformant.
4. The graphic frame emits `<p:xfrm>` (not `<a:xfrm>`), and a table deck opens
   without the repaired prompt.
5. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` green; prior
   smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` (table) | 85% | new builder file (measured via `test/pptx`) |
| `internal/ooxml/slide` (table codec) | 85% | codec band (round-trip goldens) |
| `scene` | 80% | existing band |

## 13. Smoke check

`scripts/smoke/phase-08.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` merged-cell round-trip test passes.
3. `OK:` banded-table alternating-fill test passes.
4. `OK:` `p:xfrm` codec golden passes.
5. `OK:` scene table + caption render + conformance test passes.

## 14. Tests

- **Unit:** `pptx` table builder (merge/banding/header/fill/borders/text);
  codec goldens (`p:xfrm`, merged-cell round-trip) in `internal/ooxml/slide`.
- **Round-trip golden:** merged table write → Open → model equality.
- **Integration:** scene table render → conformance (extends the scene→pptx seam).

## 15. Vocabulary added

- `Table` / `Cell` (builder) — table-builder primitives. Glossary entries if not
  already covered by the scene `Table` node.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-08.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] A builder API for a visual property added ⇒ THEME.md entry (cell fills
      compose existing tokens — no new token).
- [ ] Round-trip golden lands in this PR.
- [ ] Glossary / CHANGELOG updated.
