# Phase 27 — bento grid

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §11.2 (container nodes), §10.4 (Stage-1 validation)
**Deps:** Phase 07 (containers + `layout`), Phase 22 (content-aware
`preferredHeight`), Phase 23 (`isFlexible`). External: none.
**Status:** In progress

---

## 1. Goal

Add a `Bento` container node — rows with an optional left label and cells of
variable column span against a shared column grid — so a deck can lay out a
row-labeled bento grid, leaving `Grid`/`TwoColumn` untouched.

## 2. Why now

Completes **Wave 8** unit R5 (`DECKARD-PRODUCT-REQUIREMENTS.md`): sub-units (a)+(b)
(the TwoColumn column join) shipped in Phase 26, and this phase lands sub-unit
(c), the row-labeled bento grid — which R5 explicitly allows as a separate
sub-unit. `Grid.Ratio` is per-column, so a per-row-labeled, per-row-variable-span
layout is a structurally different layout and warrants a new node (brief 14).

## 3. RFC sections implemented

- `RFC §11.2` — a new native container node (`Bento`) that subdivides its slot
  into labelled rows of span-weighted cells and renders each cell through the
  normal dispatch, composing from existing primitives (no new OOXML — P1).
- `RFC §10.4` — Stage-1 structural validation for the new node.

## 4. Brief findings incorporated

- `docs/research/14-bento-grid.md` — *a new `Bento` node is the right shape*
  (not an overload of `Grid`, whose `Ratio` is per-column) → `Bento{Columns,
  Rows}`, `BentoRow{Label, Cells}`, `BentoCell{Span, Node}`.
- `docs/research/14-bento-grid.md` — *absolute spans keep columns aligned* → a
  shared `unitW` from `Columns`; a span-S cell is `S·unitW + (S−1)·gap`.
- `docs/research/14-bento-grid.md` — *reserve the gutter only when used* → the
  left label gutter is reserved iff any row has a non-empty `Label`.
- `docs/research/14-bento-grid.md` — *the new-node wiring checklist + a
  `cellNodes()` helper* → the helper flattens cells so every recursion
  (`validate`, `nodeUsesAssets`, `walkIconRefs`/`walkImages`/`walkDecorations`)
  is a one-liner.
- `docs/research/14-bento-grid.md` — *Stage-1 validation is the safety net* →
  `Columns ≥ 1`, non-empty rows/cells, `Span ≥ 1`, non-nil nodes, row spans ≤
  `Columns`.

## 5. Findings I'm departing from

None. The brief's open-questions (content-height rows, rowspan/per-cell valign,
alternate gutter placements) are explicitly deferred there.

## 6. Decisions referenced

- `D-026` — *Engine, not product.* The caller supplies labels, spans, and cell
  content; the engine lays them out and judges nothing.
- `D-011`/`D-018` — per-node policy: `Bento` is a native container (children
  render per their own policy); its `policyTable` entry is `{}`.
- This plan files **D-056 — Bento node** in `docs/decisions.md`.

## 7. Architecture

```text
scene/nodes.go        Bento / BentoRow / BentoCell types + KindBento (+ String)
                      Bento.cellNodes() []SlideNode  — flatten cells, row-major

scene/policy.go       policyTable[KindBento] = {}

scene/validate.go     case Bento: Columns≥1, rows/cells non-empty, Span≥1,
                      non-nil nodes, row spans ≤ Columns; validateChildren(cellNodes)

scene/render.go       renderNode:       case Bento → renderBento
                      preferredHeight:  case Bento (nRows × max cell height)
                      isFlexible:       + Bento (grows under VAlignFill)
                      nodeUsesAssets:   case Bento → nodesUseAssets(cellNodes)

scene/render_card.go     walkIconRefs:    + Bento → walkIconRefs(cellNodes)
scene/render_image.go    walkImages:      + Bento → walkImages(cellNodes)
scene/render_decoration.go walkDecorations: + Bento → walkDecorations(cellNodes)

scene/render_bento.go  renderBento: gutter (iff any label) + equal-height rows +
                       per-row absolute-span cell layout via a shared unitW   NEW
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — Bento/BentoRow/BentoCell + KindBento + String + cellNodes
scene/policy.go                      # CHANGED — policyTable[KindBento]
scene/validate.go                    # CHANGED — case Bento
scene/render.go                      # CHANGED — renderNode, preferredHeight, isFlexible, nodeUsesAssets
scene/render_card.go                 # CHANGED — walkIconRefs recursion
scene/render_image.go                # CHANGED — walkImages recursion
scene/render_decoration.go           # CHANGED — walkDecorations recursion
scene/render_bento.go                # NEW — renderBento + gutter/unit constants
scene/render_bento_test.go           # NEW — labels, spans, validation, flex, determinism
scene/scene_test.go                  # CHANGED — allNodes + catalog count 20 → 21
test/integration/roundtrip_test.go   # CHANGED — everyNodeScene + collectKinds + kind-range loop
scripts/smoke/phase-27.sh            # NEW — phase smoke
docs/research/14-bento-grid.md       # NEW — informing brief
docs/research/INDEX.md               # CHANGED — registers brief 14
docs/plans/phase-27-bento-grid.md    # NEW — this plan
docs/plans/README.md                 # CHANGED — adds Phase 27 to Wave 8
docs/decisions.md                    # CHANGED — adds D-056
docs/glossary.md                     # CHANGED — adds "Bento", "Column span"
docs/site/catalog/containers.md      # CHANGED — Bento node docs (§19)
skills/compose-a-scene/SKILL.md      # CHANGED — Bento node entry (§19)
```

## 9. Public API surface

```go
// scene (nodes.go)
type BentoCell struct {
    Span int       // column units this cell occupies (>= 1)
    Node SlideNode // cell content
}
type BentoRow struct {
    Label string      // left-gutter label; "" = no label for this row
    Cells []BentoCell
}
type Bento struct {
    node
    Columns int // shared column-unit count a row's spans are measured against (>= 1)
    Rows    []BentoRow
}
func (Bento) NodeKind() NodeKind // KindBento
```

New scene IR node ⇒ a smoke check and Stage-1 validation land in this PR
(§4.2/§13). No new builder API, no new theme token (P2 — reuses `TextMuted`).

## 10. Risks

- **R1 — incomplete node wiring.** A new node must touch many switches; a missed
  one is a silent bug. **Mitigation:** the round-trip `everyNodeScene` +
  `collectKinds` + the contiguous kind-range loop (`KindHero..KindBento`) fail
  loudly if `Bento` isn't exercised; the catalog count assertion (21) and
  `TestPolicy_MatchesStructs` guard the enum/policy.
- **R2 — column misalignment / overflow.** **Mitigation:** absolute `unitW`
  spans; Stage-1 rejects a row whose spans exceed `Columns`; a test asserts a
  span-2 cell is ~twice a span-1 cell's width.
- **R3 — determinism / parallel safety.** **Mitigation:** integer-EMU geometry;
  `nodeUsesAssets` recurses cells so an asset-bearing cell still forces serial; a
  determinism test renders a bento deck across 1 vs N workers.

## 11. Acceptance criteria

1. A `Bento` with labelled rows renders each row's left label and its cells; a
   span-2 cell is about twice the width of a span-1 cell (columns align).
2. A `Bento` with no labelled row reserves no gutter (cells use the full width).
3. Stage-1 rejects: `Columns < 1`, an empty row, a `Span < 1`, a nil cell node,
   and a row whose spans exceed `Columns`.
4. `Bento` is flexible (grows under `VAlignFill`) and recurses for assets/icons/
   images/decorations (a framed image in a cell resolves; a bento with a
   media cell composes sequentially).
5. The round-trip "every node" guard covers `Bento`; the catalog has 21 kinds.
6. A bento deck renders byte-identical across 1 vs N workers.
7. `make coverage` shows `scene` ≥ its band; `make preflight` passes.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; the new branches are covered by
`render_bento_test.go` + the catalog/round-trip updates.

## 13. Smoke check

`scripts/smoke/phase-27.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` bento renders labels + span-aligned cells (criterion 1).
3. `OK:` no gutter when unlabelled (criterion 2).
4. `OK:` Stage-1 rejects malformed bento (criterion 3).
5. `OK:` catalog has 21 kinds / round-trip covers Bento (criterion 5).
6. `OK:` bento render is deterministic across workers (criterion 6).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — `renderBento` geometry (label present, span widths, gutter
  reservation), `validateNode` rejections, `isFlexible`, determinism (white/black
  box).
- **Round-trip golden:** the existing `everyNodeScene` round-trip extends to
  `Bento` (write → reopen → byte-identical), which is the round-trip guard.
- **Integration** (`test/integration/`): the round-trip "every node" test gains
  `Bento`.
- **Fuzz / Benchmark:** no.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Bento` — a scene IR container node: rows with an optional left label and cells
  of variable column span against a shared column grid.
- `Column span` — a `BentoCell`'s width in `Bento.Columns` column units
  (`Span`); a span-S cell occupies S units.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-27.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entry D-056 added.
- [ ] Docs site updated for the Bento node (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
