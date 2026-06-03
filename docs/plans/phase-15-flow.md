# Phase 15 — Flow

**Subsystem:** scene (flow)
**RFC sections:** §11.1 (flow), §12 (per-node policy), §16 (v4 mapping)
**Deps:** Phase 14 (card-like step pill + icon registry wiring)
**Status:** Draft

---

## 1. Goal

Render the `flow` scene node as a native sequential step pipeline — labeled
pills joined by connector glyphs (`arrow`, `arrow_dashed`, `cycle`, `plus`) in
horizontal or vertical direction — composing the public builder only (P1), with
no new builder capability.

## 2. Why now

Phase 15 is next in the master plan (`docs/plans/README.md`, Wave 4) and its one
dependency shipped: Phase 14 built the card-like rounded pill and wired the icon
registry into compose, both of which the flow step reuses. Flow completes the
§11.1 process-visual set alongside cards.

## 3. RFC sections implemented

- `RFC §11.1` — the `flow` node ("Sequential step pipeline (horizontal/vertical)").
- `RFC §12` — the per-node policy row "`flow` → Native (step pills + connectors)"
  (§12.1); flow carries no `asset_id`, so it renders natively.
- `RFC §16` — the v4 → IR mapping `flow → Flow (incl. connector kinds)`.

## 4. Brief findings incorporated

- `docs/research/06-flow-step-pipeline.md`:
  - **F1 (no new builder capability)** → flow composes `AddShape` (pills +
    preset connector glyphs) + `AddTextFrame`. The RFC's unbuilt `AddConnector`
    (anchored `cxnSp`) is **not** built — flow connectors are decorative glyphs
    placed in inter-pill gaps by the layout, not routed between anchors.
  - **F2 / Q4 (flow-level connector)** → a single `Flow.Connector ConnectorKind`
    applied between every adjacent pair (per-pair is a future additive field).
  - **F3 / Q2 (cycle)** → `cycle` renders `arrow` connectors plus a trailing
    return arrow (a `circularArrow` preset glyph after the last pill) — the
    simpler, deterministic option.
  - **F4 (plus)** → `mathPlus` preset glyph in each gap.
  - **F5 / Q1 (arrow_dashed)** → a thin **dashed line** (`ShapeLine` +
    `Line.Dash`) plus a small solid **chevron** head — two shapes per connector,
    within the current builder.
  - **F6 (lighter pill)** → a dedicated `renderFlowStep` (roundRect + centered
    label + optional detail + optional icon), not the heavier card chrome.
  - **F7 / Q3 (additive IR + step icons)** → `Flow.Connector` (zero = arrow) and
    `FlowStep.Icon` (optional, via the Phase-14 registry); re-export
    `ConnectorKind`.
  - **F8 (layout)** → horizontal: equal-width pill columns with connector glyphs
    centered in the gaps; vertical: stacked pills with connectors rotated to
    point down. Integer EMU, deterministic.

## 5. Findings I'm departing from

- `docs/research/06-…` **R3 / F5 also floated building line-end arrowheads on
  `pptx.Line`** as the "cleanest result." **Departing:** Phase 15 ships the
  dashed-line + chevron-head composition (no new builder API), per the
  maintainer's choice. Real OOXML `lnEnd` arrowheads remain a future builder
  addition (V1.x) if a node needs true arrow-terminated lines; recorded in D-044.

## 6. Decisions referenced

- `D-040` — icon engine + `AddIcon` — flow step icons resolve through it.
- `D-043` — Phase-14 icon-registry wiring + `validateIconRefs` closed-name
  Stage-1 check — extended here to walk flow steps; additive-IR pattern reused.
- `D-035` — byte-identical determinism — flow geometry is integer EMU, no map
  iteration; connectors and pills place deterministically.
- `D-015` — parallel render — flow is media-free (native shapes + custGeom
  icons), classified not-asset-bearing in `nodeUsesAssets`.
- `D-026` — engine, not product — flow renders the IR's typed connector/steps
  with no heuristics.
- **`D-044` (new, this PR)** — Flow renders by composition (no new builder API);
  flow-level connector kind; `arrow_dashed` = dashed line + chevron head (defer
  `lnEnd` arrowheads); additive `Flow.Connector` + `FlowStep.Icon`.

## 7. Architecture

Single PR (pure composition — no builder change, so no split). All in `scene`.

```text
scene/nodes.go        # CHANGED — Flow.Connector + FlowStep.Icon; ConnectorKind enum
scene/render_flow.go  # NEW — renderFlow, renderFlowStep, connector glyphs
scene/render.go       # CHANGED — dispatch Flow; nodeUsesAssets; preferredHeight
scene/render_card.go  # CHANGED — walkCards→ also walk flow steps for validateIconRefs
                      #           (or a sibling walk; icon refs include flow steps)
```

Layout (horizontal, N steps): the slot splits into N pill columns and N−1
connector gaps via the `scene/layout` engine; `cycle` adds a return glyph after
the last pill. Vertical: rows instead of columns, connectors rotated 90°.

```text
┌─────┐  →  ┌─────┐  →  ┌─────┐        (arrow)
│ [i] │     │ [i] │     │ [i] │
│ Lbl │     │ Lbl │     │ Lbl │
└─────┘     └─────┘     └─────┘
   └──────────── ⟳ ──────────┘         (cycle adds the return arrow)
```

## 8. Files added or changed

```text
scene/nodes.go             # CHANGED — Flow.Connector, FlowStep.Icon, ConnectorKind (re-exported scene enum)
scene/render_flow.go       # NEW — renderFlow + renderFlowStep + connector composers
scene/render.go            # CHANGED — dispatch Flow; nodeUsesAssets(Flow)=media-free; preferredHeight(Flow)
scene/render_card.go       # CHANGED — validateIconRefs also walks Flow steps (icon names)
scene/render_flow_test.go  # NEW — render + connectors + icon + parallel idempotency
scripts/smoke/phase-15.sh  # NEW — phase smoke
docs/decisions.md          # CHANGED — adds D-044
docs/glossary.md           # CHANGED — flow step, connector kind
docs/research/06-flow-step-pipeline.md  # NEW (committed with plan)
docs/plans/phase-15-flow.md             # NEW (this file)
```

No new visual token (connectors/pills reuse existing color/radius/space tokens),
so no `docs/design/THEME.md` change is required.

## 9. Public API surface

```go
// scene — additive
type ConnectorKind int
const (
    ConnectorArrow       ConnectorKind = iota // solid arrow (zero / default)
    ConnectorArrowDashed                      // dashed line + chevron head
    ConnectorCycle                            // arrows + trailing return arrow
    ConnectorPlus                             // mathPlus glyph between steps
)

type FlowStep struct {
    Label  RichText
    Detail RichText
    Icon   string // NEW — curated/extension icon name (closed-name, Stage-1 validated)
}

type Flow struct {
    node
    Orientation FlowOrientation
    Steps       []FlowStep
    Connector   ConnectorKind // NEW — zero = ConnectorArrow (preserves prior default)
}
```

No prior surface breaks: `Connector`/`Icon` zero values preserve a plain arrow
flow with no step icons.

## 10. Risks

- **R1 — `arrow_dashed` two-shape connector misaligns** (line + chevron head).
  **Mitigation:** compute both from the same gap box (line spans the gap, head
  centered at the gap's leading edge); a render test asserts both shapes emit and
  the deck conforms.
- **R2 — Vertical connector rotation.** **Mitigation:** reuse `WithRotation`
  (D-041) for the chevron/arrow in vertical flows; a test asserts a vertical flow
  emits rotated glyphs and stays conformant.
- **R3 — Step icon makes a flow sequential (loses parallelism).**
  **Mitigation:** icons are `custGeom` (native, not media); `nodeUsesAssets(Flow)`
  returns false. A workers=1 vs N byte-identical test guards it.
- **R4 — `cycle` return glyph overlaps content.** **Mitigation:** place the
  `circularArrow` in a reserved trailing slot (horizontal) / below the last pill
  (vertical); overflow records a `LayoutWarning`, not an error (D-026).

## 11. Acceptance criteria

1. A 4-step horizontal `Flow` with `ConnectorArrow` renders 4 pills + 3 arrow
   glyphs as native shapes; the deck conforms; no `pic`.
2. A `ConnectorCycle` flow renders the inter-pair arrows **plus** one trailing
   return arrow after the last step.
3. A vertical `Flow` renders connectors rotated to point down.
4. A `ConnectorArrowDashed` flow renders a dashed line + a chevron head per gap.
5. A `ConnectorPlus` flow renders a `mathPlus` glyph per gap.
6. A `FlowStep.Icon: "<curated>"` places a `custGeom` icon in the pill; an
   unknown icon name fails Stage-1 with a render error **before** compose.
7. `scene.Render` is byte-identical for a flow scene at `workers=1` and
   `workers=N` (D-035/D-015).
8. A plain `Flow{Orientation, Steps}` (zero `Connector`, no step `Icon`) renders
   as an arrow flow (additive-IR default preserved).
9. `make coverage` shows `scene` ≥ its band.
10. `scripts/smoke/phase-15.sh` reports `OK ≥ count(criteria)`, `FAIL = 0`;
    prior smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; flow composer adds to the existing package |

No new package → no `coverage.json` entry.

## 13. Smoke check

`scripts/smoke/phase-15.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` a 4-step horizontal arrow flow renders (pills + connectors).
3. `OK:` a cycle flow appends a return arrow.
4. `OK:` a vertical flow rotates connectors.
5. `OK:` arrow_dashed and plus connectors render.
6. `OK:` a flow step icon resolves; an unknown name fails Stage-1.
7. `OK:` flow render is byte-identical workers=1 vs N.

## 14. Tests

- **Unit:** `scene` (flow layout, each connector kind, step icon resolution +
  unknown-name failure, `nodeUsesAssets` classification).
- **Round-trip golden:** not a new builder primitive (composition only); flow
  output is covered by conformance + render assertions. (No `pptx` API added, so
  no new G6 golden — the shapes it composes are already round-trip-tested.)
- **Integration** (`test/integration/`): the flow step icon consumes the icon
  registry seam (Phase 12/14). A test renders a flow with a step icon through
  real `internal/opc` + `encoding/xml` and asserts the `custGeom` reaches the
  slide; an unknown name errors.
- **Fuzz:** none (no new parse surface).
- **Benchmark:** optional.

## 15. Vocabulary added

File in `docs/glossary.md` (alphabetical) in this PR:

- `connector kind` — a flow's inter-step glyph: arrow / arrow_dashed / cycle /
  plus (`ConnectorKind`).
- `flow step` — one pill in a `Flow`: label + optional detail + optional icon
  (`FlowStep`).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-15.sh` reports `OK ≥ count(criteria)` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (D-044).
- [ ] (Phase 20+) Docs site updated for user-facing surface changes. (inert)
- [ ] (Phase 20+) Affected agent skill(s) updated. (inert)
