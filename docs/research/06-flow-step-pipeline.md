# Brief 06 — flow step pipeline (pills + connectors)

**Subsystem:** scene (flow)
**Authored:** 2026-06-03
**Motivating phase:** Phase 15 — Flow (RFC §11.1, §12, D-026)

## 1. Question

How should pptx-go render the `flow` scene node — a sequential step pipeline of
labeled pills joined by connectors (`arrow`, `arrow_dashed`, `cycle`, `plus`),
in horizontal and vertical directions — as **native PPTX shapes** composing the
public builder (P1), and does it need any new builder capability (as Phase 14
needed the shadow primitive), or is it pure composition?

Sub-questions the plan must settle:

1. **Connectors.** The RFC §8.1 sketches `Slide.AddConnector(kind, from, to)`
   but it was **never built**. Do flow connectors need it, or do they compose
   existing preset shapes?
2. **Connector model.** Is the connector kind a single flow-level property, or
   per-pair (a connector between each adjacent pair)?
3. **`cycle` and `plus`.** How does `cycle` "append a return-arrow after the
   last step" (RFC), and how is `plus` ("A + B + C") drawn?
4. **`arrow_dashed`.** The builder's preset arrows are *filled block* shapes
   (`rightArrow`) and `Line` has a `Dash` but **no arrowhead** (no
   `headEnd`/`tailEnd`). How is a dashed *arrow* drawn within that surface?
5. **Step pill.** Reuse the Phase-14 card chrome, or a lighter dedicated pill?
6. **IR growth.** What does `Flow` need beyond `{Orientation, Steps}` today?

## 2. Prior art surveyed

- **The `Arrow` leaf** (`scene/render_leaves.go`): renders a directional arrow
  via `arrowGeom(dir)` → the OOXML preset name (`rightArrow` / `leftArrow` /
  `upArrow` / `downArrow`) through `ps.AddShape`, filled with the accent token,
  optional centered label. Exactly the per-connector primitive a flow needs
  between pills — no new builder API.
- **The card / chip chrome** (`render_card.go`, `renderChip`): rounded-rect +
  centered/anchored text is the step-pill shape. The card chrome
  (`renderCardChrome`) adds a stripe + header row + padding — heavier than a
  flow step needs.
- **The builder shape surface** (`pptx/shape.go`): `ShapeGeometry` is a string
  type, so any OOXML preset is reachable via `pptx.ShapeGeometry("…")` even
  without a named constant (named ones: rect, roundRect, ellipse, triangle,
  diamond, parallelogram, hexagon, chevron, rightArrow, line). `Line{Width,
  Color, Dash}` supports preset dashes (`"dash"`, `"sysDash"`, …) but carries
  **no line-end arrowheads**.
- **The unbuilt `AddConnector`** (RFC §8.1, §8.2 `Anchor{ShapeID, Side}`): a
  true anchored connector (`cxnSp`) that routes between two shapes' anchor
  sides. Not built; a separate capability from a decorative inter-pill arrow.
- **OOXML preset geometries** (ECMA-376 §20.1.10.55, `ST_ShapeType`): beyond
  the block arrows, `mathPlus` (a "+"), `circularArrow` / `bentArrow` /
  `curvedRightArrow` (return/loop arrows), `chevron` (a process chevron). All
  reachable via `AddShape(ShapeGeometry("…"))`.
- **pengui-slides v4 flow**: a row (or column) of step pills with a uniform
  connector glyph between each pair; `cycle` loops the last back to the first;
  `plus` reads the steps as an additive set. The connector style is a
  flow-level choice, not per-pair.
- **Determinism (D-035)** and **parallelism (D-015)**: a flow is native shapes
  only (no media), integer-EMU geometry — media-free and parallel-safe; classify
  in `nodeUsesAssets` as not-asset-bearing.

## 3. Findings

- **F1 — Flow needs no new builder capability (unlike Phase 14).** Pills are
  `roundRect` + text; connectors are preset arrow/plus shapes via `AddShape`.
  Everything composes the existing builder (P1). The `AddConnector` RFC sketch
  is a *different*, unbuilt capability (anchored `cxnSp` routing) that flow does
  not require — flow connectors are decorative glyphs placed in the gaps between
  pills, positioned by the layout, not anchored to shape sides. Building
  `AddConnector` is out of scope for Phase 15 (defer to if/when a node needs
  true routed connectors).
- **F2 — Connectors are a flow-level kind, not per-pair.** RFC §16 maps
  `flow → Flow (incl. connector kinds)` and §11.1 calls it a "sequential step
  pipeline"; pengui v4 uses one connector style per flow. A single
  `Flow.Connector ConnectorKind` applied between every adjacent pair matches the
  acceptance criteria ("a 4-step horizontal flow with arrow connectors") and is
  simpler. Per-pair connectors are a future additive `[]ConnectorKind` if a
  real deck needs mixed connectors.
- **F3 — `cycle` is a flow mode, not just a connector glyph.** "Appends a
  return-arrow after the last step" (RFC) means: render `arrow` connectors
  between pairs, then add one extra return arrow from the last pill back toward
  the first (a `circularArrow`/`bentArrow` preset, or a routed arrow along the
  margin). So `cycle` ≈ `arrow` + a trailing loop-back. Modeling it as a
  `ConnectorKind` value (cycle) that the renderer special-cases is cleanest.
- **F4 — `plus` is the `mathPlus` preset** placed in each gap (the steps read
  as an additive set, "A + B + C"). No arrowhead, no direction — symmetric.
- **F5 — `arrow_dashed` is the one geometry wrinkle.** The block-arrow presets
  can't be "dashed" meaningfully (they're filled), and `Line` has no arrowhead.
  Two viable renderings within the current surface: (a) a thin **dashed line**
  (`ShapeLine` + `Line.Dash`) plus a small solid `chevron`/`triangle` head — two
  shapes per connector; or (b) a `rightArrow` block with a **dashed outline**
  and no fill (an outline-only dashed arrow). (a) reads more like a dashed
  connector; (b) is one shape. The plan picks one; neither needs new builder
  API. (A third option — add line-end arrowheads to `pptx.Line` — is a real
  builder addition; defer unless the plan wants it.)
- **F6 — A dedicated lighter step-pill, not the card chrome.** A flow step is a
  compact `roundRect` (themed fill/border) with a centered label + optional
  smaller detail line + optional icon — no accent stripe, no header row. Reusing
  `renderCardChrome` would impose stripe/header padding inappropriate for a
  pill. A small `renderFlowStep` helper (sibling to `renderChip`) is the fit.
- **F7 — IR growth is additive (D-043 pattern).** Add `Connector ConnectorKind`
  to `Flow` (zero = `arrow`, the most common, so existing `Flow{Orientation,
  Steps}` keeps a sensible default). Optionally `Icon string` on `FlowStep`
  (steps often carry an icon, like cards — resolves through the same icon
  registry wired in Phase 14). Re-export `ConnectorKind` into the scene package.
- **F8 — Layout.** Horizontal: pills are equal-width columns with connector
  glyphs centered in the inter-pill gaps; vertical: pills stack with connectors
  rotated 90° (down arrows). The `scene/layout` engine already splits a box into
  weighted columns/rows; the flow reserves gap slots for connectors. All integer
  EMU → deterministic.

## 4. Recommendations

(Inputs to the plan; the plan binds.)

- **R1 — Pure composition, no new builder API.** Render the flow with
  `AddShape` (pills + preset connector glyphs) and `AddTextFrame` (labels).
  Explicitly *do not* build `AddConnector` in Phase 15.
- **R2 — Flow-level `Connector ConnectorKind`** with values `{ConnectorArrow
  (zero), ConnectorArrowDashed, ConnectorCycle, ConnectorPlus}`. Applied between
  every adjacent pair; `cycle` adds the trailing return arrow.
- **R3 — `arrow_dashed` = dashed line + small chevron head** (option F5-a) — it
  reads as a dashed connector and stays within the builder. Record the choice in
  the D-NNN; note the deferred alternative (line-end arrowheads on `pptx.Line`).
- **R4 — Dedicated `renderFlowStep` pill** (roundRect + centered label + detail
  + optional icon), not the card chrome.
- **R5 — Additive IR**: `Flow.Connector` (zero = arrow) and `FlowStep.Icon`
  (optional, via the Phase-14 icon registry + `validateIconRefs` extended to
  walk flow steps). Re-export `ConnectorKind`.
- **R6 — Classify flow as media-free** in `nodeUsesAssets` (native shapes +
  custGeom icons only), so flows render in parallel.

## 5. Open questions

- **Q1 — `arrow_dashed` rendering** (dashed line + chevron head vs dashed-outline
  block arrow vs building line-end arrowheads). Plan + a fork to the maintainer.
- **Q2 — `cycle` return-arrow geometry** (a `circularArrow` preset glyph after
  the last pill vs a routed bent arrow along the margin back to the first). The
  former is simpler and deterministic; the latter is prettier but fiddly. Plan
  decides; default to the simpler preset.
- **Q3 — Per-step icons in V1?** Flow steps commonly carry icons. Including
  `FlowStep.Icon` now reuses the Phase-14 registry cheaply; deferring keeps
  Phase 15 smaller. A fork.
- **Q4 — Connector model** (flow-level uniform vs per-pair). Recommended uniform
  for V1; per-pair is a future additive field. Confirm with the maintainer.
