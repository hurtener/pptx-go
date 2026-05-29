# Phase 07 — Container nodes (two_column, grid)

**Subsystem:** scene (containers + sub-region layout)
**RFC sections:** §10.2 (container sub-layout), §11.2 (two_column, grid),
§12 (containers render nothing themselves; children render per their policy)
**Deps:** Phase 06 (leaf composers + the body layout + node dispatch).
**Status:** In progress

---

## 1. Goal

Render the `two_column` and `grid` containers: a pure geometry engine
(`scene/layout`) that subdivides a parent box into ratio/column slots, and
composers that stack each child node into its slot via the existing dispatch.

## 2. Why now

Phase 06 placed top-level leaves in a vertical body stack and warned on
containers. Containers are the next layer of structure every real deck uses, and
`card`/`card_section` (Phase 14) build on the same sub-region engine. This phase
turns `scene/layout` from a placeholder into the geometry library the rest of
Wave 2 reuses.

## 3. RFC sections implemented

- `RFC §11.2` — `two_column` (1:1 / 1:2 / 2:1 split with leaf children) and
  `grid` (2/3/4 columns, weighted ratios, one child per cell).
- `RFC §10.2` — each container has its own internal layout; children render per
  their own policy (recursive dispatch). Overflow surfaces a `LayoutWarning`.
- `RFC §10.4` — Stage 1 gains the grid completeness check (cell count is a
  multiple of `columns`, i.e. `columns × ⌈cells/columns⌉ == cells`).

`card` / `card_section` are Phase 14; other leaves remain as Phase 06 left them.

## 4. Brief findings incorporated

No informing brief — the container geometry is specified in RFC §10.2/§11.2;
`docs/research/INDEX.md` lists a layout-engine survey only as a candidate (the
engine here is deterministic slot division, "not a full constraint solver").

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-026` — engine, not product: slot division is deterministic geometry, no
  legibility/overflow opinions (overflow → a warning the caller may act on).
- `D-011`/`D-018` — containers render nothing themselves; each child renders per
  its intrinsic policy via the existing dispatch.

## 7. Architecture

The geometry is pure and lives in `scene/layout` (it imports only `pptx`, no
node types — no import cycle). The composers live in the `scene` package and
call it, then recurse into `renderNode` per child.

```text
scene/layout/columns.go  Columns(parent, weights, gap) []pptx.Box
scene/layout/grid.go     Grid(parent, cols, colWeights, gap, count) []pptx.Box
scene/render_container.go renderTwoColumn / renderGrid (stack children per slot)
scene/render.go          stackIn extracted (body + column stacking share it);
                         TwoColumn/Grid added to dispatch + preferredHeight
scene/validate.go        grid completeness check
```

**Columns.** Width = `(parent.W − gap·(n−1)) · wᵢ / Σw`, left-to-right, full
parent height. **Grid.** `rows = ⌈count/cols⌉`; column x/width from `Columns`;
equal row heights `(parent.H − gap·(rows−1)) / rows`; cells row-major.

**Child stacking.** `two_column` stacks each side's children vertically in its
column box (reusing `stackIn`, extracted from the Phase 06 body layout). `grid`
places one child per cell box. A child that is itself a container recurses, so
nesting (a grid inside a column) composes for free.

**Height estimation.** When a container is itself a body-stack item, its
`preferredHeight` is estimated from its children (`two_column` = max of the two
sides' summed child heights; `grid` = rows × max child height), so it reserves a
sensible slot.

## 8. Files added or changed

```text
scene/layout/columns.go    # NEW — Columns split
scene/layout/grid.go       # NEW — Grid split
scene/layout/layout_test.go# NEW — geometry unit tests (cell widths/heights)
scene/render_container.go  # NEW — renderTwoColumn / renderGrid
scene/render.go            # CHANGED — stackIn extraction, dispatch, preferredHeight
scene/validate.go          # CHANGED — grid completeness (cells % columns == 0)
scene/*_test.go            # CHANGED — fixtures use complete grids; new render tests
docs/plans/phase-07-containers.md  # NEW
scripts/smoke/phase-07.sh  # NEW
```

## 9. Public API surface

No new `scene` exported types. New `scene/layout` package functions `Columns`
and `Grid` are the geometry primitives (consumed by `scene`; also usable by
later phases). `Grid` validation tightens (a partial last row is now a Stage 1
error).

## 10. Risks

- **R1 — rounding.** Integer EMU division can lose a few EMU across columns.
  *Mitigation:* deterministic floor division; tests assert widths within the
  expected value (and that columns fit inside the parent). Not pixel-perfect by
  design.
- **R2 — grid completeness break.** Tightening Stage 1 invalidates partial-grid
  fixtures from Phase 05/06. *Mitigation:* update those fixtures to complete
  rows in the same PR; add a negative test.
- **R3 — nesting depth.** A container in a container recurses. *Mitigation:* the
  dispatch already recurses; a deep-nest test (grid-in-column) covers it.

## 11. Acceptance criteria

1. `1:1`, `1:2`, `2:1` two_column ratios produce the expected cell widths
   (within rounding) and fit inside the parent.
2. A grid with 2/3/4 columns and a weighted ratio produces the expected cell
   widths and row count.
3. A grid whose cell count is not a multiple of `columns` raises a Stage 1
   validation error.
4. A two_column / grid renders its children as native shapes in their slots
   (the container emits no shape of its own), and the deck stays conformant.
5. Nesting (a grid inside a two_column column) composes.
6. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` green; prior
   smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | existing band |
| `scene/layout` | 80% | new scene package (geometry) |

`scene/layout` is added to `coverage.json`.

## 13. Smoke check

`scripts/smoke/phase-07.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` two_column ratio widths test passes.
3. `OK:` grid column widths + row count test passes.
4. `OK:` grid completeness validation test passes.
5. `OK:` container render + conformance (incl. nesting) test passes.

## 14. Tests

- **Unit:** `scene/layout` — `Columns`/`Grid` widths/heights/fit; `scene` —
  container render output + the grid validation negative.
- **Round-trip / integration:** a container scene through render → conformance
  (extends the Phase 06 scene→pptx seam).

## 15. Vocabulary added

- none required (containers are already in the glossary; "slot"/geometry terms
  are internal).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages (scene + scene/layout).
- [ ] `scripts/smoke/phase-07.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated (if vocab added).
- [ ] Decision entries added (if any).
