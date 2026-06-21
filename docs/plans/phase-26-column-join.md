# Phase 26 — column join

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §11.2 (TwoColumn container)
**Deps:** Phase 07 (containers). External: none.
**Status:** In progress

---

## 1. Goal

Add an optional centered element on a `TwoColumn`'s seam — a "VS"-style text
badge or a connector arrow — so a deck can compare two columns with a badge
between them or link them with an arrow, opt-in and byte-identical when unset.

## 2. Why now

Fifth unit of **Wave 8 — post-V1 engine extensions**
(`DECKARD-PRODUCT-REQUIREMENTS.md` R5, LOW), sub-units **(a)** center badge and
**(b)** inter-column connector. R5's third sub-unit **(c)** — the row-labeled
bento grid — is a distinct new IR node and lands as **Phase 27**, which R5
explicitly permits ("these can land as separate sub-units"). Picked up after
R1–R4 per the one-requirement-per-PR cadence.

## 3. RFC sections implemented

- `RFC §11.2` — extends the native `TwoColumn` container with an optional
  inter-column element, composed from existing builder primitives (`ShapeEllipse`
  + a text run for the badge, `ShapeRightArrow` for the connector) — no new OOXML
  capability (P1).

## 4. Brief findings incorporated

- `docs/research/13-column-join.md` — *one enum with a `None` zero covers (a) and
  (b)* → `ColumnJoin` (`JoinNone`/`JoinBadge`/`JoinArrow`) + `JoinLabel`; the zero
  value is naturally "absent" (byte-identical), no pointer/bool.
- `docs/research/13-column-join.md` — *the element sits on the seam, overlapping
  both columns* → centered on `(left.X+left.W + right.X)/2`, drawn after the
  column content so it sits on top.
- `docs/research/13-column-join.md` — *byte-identity / determinism are automatic*
  → join shapes emit only for non-`JoinNone`; pinned EMU geometry, native shapes,
  no media.
- `docs/research/13-column-join.md` — *two columns only; N-column connectors
  deferred* → the connector is the 2-column A → B case; the architecture-diagram
  multi-column connector is out of scope (not in R5's acceptance).

## 5. Findings I'm departing from

None. The brief's open-questions (N-column connectors, badge size/shape knobs,
vertical placement) are explicitly deferred there.

## 6. Decisions referenced

- `D-044` — *Flow connectors compose preset shapes.* The column connector follows
  the same native-shape precedent (no anchored `AddConnector`).
- `D-026` — *Engine, not product.* The caller supplies the badge label and picks
  badge-or-arrow; the engine draws it on the seam and invents nothing.
- This plan files **D-055 — TwoColumn column join** in `docs/decisions.md`.

## 7. Architecture

```text
scene/nodes.go         ColumnJoin enum (JoinNone/JoinBadge/JoinArrow)   NEW type
                       TwoColumn: + Join ColumnJoin, + JoinLabel string

scene/render_container.go
  renderTwoColumn:     after the two column stacks, when Join != JoinNone,
                       draw the seam element                              NEW call
  renderColumnJoin:    badge (accent ellipse + centered inverse label) or
                       connector (accent right-arrow), centered on the seam  NEW
  consts:              joinBadgeSz, joinArrowW, joinArrowH                NEW
```

The seam midpoint is `(left.X + left.W + right.X) / 2`; the element centers there
vertically at `box.Y + box.H/2`, overlapping both columns (the reference look).

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — ColumnJoin enum + TwoColumn.Join/JoinLabel
scene/render_container.go            # CHANGED — renderColumnJoin + call + consts
scene/render_container_test.go       # CHANGED — badge/arrow render, omit-when-none, byte-identical, determinism
scripts/smoke/phase-26.sh            # NEW — phase smoke
docs/research/13-column-join.md      # NEW — informing brief
docs/research/INDEX.md               # CHANGED — registers brief 13
docs/plans/phase-26-column-join.md   # NEW — this plan
docs/plans/README.md                 # CHANGED — adds Phase 26 to Wave 8
docs/decisions.md                    # CHANGED — adds D-055
docs/glossary.md                     # CHANGED — adds "Column join", "VS badge"
docs/site/catalog/containers.md      # CHANGED — TwoColumn field docs (§19)
skills/compose-a-scene/SKILL.md      # CHANGED — TwoColumn field list (§19)
```

## 9. Public API surface

```go
// scene (nodes.go)
type ColumnJoin int
const (
    JoinNone  ColumnJoin = iota // default: nothing between the columns (byte-identical)
    JoinBadge                   // a circular text badge (JoinLabel), e.g. "VS"
    JoinArrow                   // a right-arrow connector between the columns
)

type TwoColumn struct {
    // ... existing ...
    Join      ColumnJoin // optional inter-column element; JoinNone = none
    JoinLabel string     // badge text when Join == JoinBadge
}
```

New public scene surface (an enum + two `TwoColumn` fields) ⇒ a smoke check
lands in this PR (§4.2/§13). No new builder API, no new scene IR node, no new
theme token (P2 — reuses `ColorAccent` / `TextInverse`).

## 10. Risks

- **R1 — byte-identity regression for existing TwoColumn.** **Mitigation:** the
  join draws only for non-`JoinNone`; a test renders a `JoinNone` TwoColumn and
  asserts it is byte-identical to one with the fields absent.
- **R2 — determinism.** **Mitigation:** fixed integer-EMU geometry; a determinism
  test renders a join deck across 1 vs N workers.
- **R3 — element mispositioned.** **Mitigation:** a test asserts a badge/arrow
  shape is emitted and (for the badge) the label text is present; the seam math
  reuses the column boxes `layout.Columns` already returns.

## 11. Acceptance criteria

1. A `TwoColumn` with `Join: JoinBadge, JoinLabel: "VS"` renders a badge (an
   ellipse + the "VS" text) centered on the column seam.
2. A `TwoColumn` with `Join: JoinArrow` renders a right-arrow connector on the
   seam.
3. A `TwoColumn` with `Join: JoinNone` (zero value) renders **byte-identical** to
   today (no join shapes).
4. A deck of join TwoColumns renders byte-identical across 1 vs N workers.
5. `make coverage` shows `scene` ≥ its band; `make preflight` passes.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; new branches covered by the container
tests.

## 13. Smoke check

`scripts/smoke/phase-26.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` JoinBadge renders an ellipse + label (criterion 1).
3. `OK:` JoinArrow renders a connector arrow (criterion 2).
4. `OK:` JoinNone is byte-identical (criterion 3).
5. `OK:` join render is deterministic across workers (criterion 4).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — black-box render assertions on the emitted slide XML
  (ellipse + `<a:t>VS</a:t>` for the badge, `prst="rightArrow"` for the
  connector, and their absence for `JoinNone`); byte-identity + determinism.
- **Round-trip golden:** N/A — no builder primitive / scene node added.
- **Integration** (`test/integration/`): no — internal to `scene` TwoColumn.
- **Fuzz / Benchmark:** no.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Column join` — the optional element a `TwoColumn` draws on its seam
  (`ColumnJoin`): nothing, a text badge, or a connector arrow.
- `VS badge` — a `Column join` text badge (`JoinBadge` + `JoinLabel`), a circular
  label straddling the seam between two compared columns.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-26.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entry D-055 added.
- [ ] Docs site updated for the TwoColumn fields (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
