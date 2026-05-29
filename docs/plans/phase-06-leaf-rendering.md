# Phase 06 — Leaf-node rendering

**Subsystem:** scene (text leaves + a deterministic body layout)
**RFC sections:** §10.2 (body layout), §10.6 (asset resolution at render),
§11.1, §12 (rows: hero, prose, heading, list, divider, quote, callout, chip,
arrow, code_block, section_divider)
**Deps:** Phase 05 (IR catalog + validation + asset seam), Phase 04 (the
builder rich-text model the composers call), Phase 03 (shapes/media/notes).
**Status:** Done

---

## 1. Goal

Make `scene.Render` emit: a deterministic top-level vertical body layout plus a
per-node composer for each text-heavy leaf, turning a `Scene` of leaf nodes
into a real, conformant PPTX.

## 2. Why now

Phase 05 shipped the IR and a no-op `Render`; this is the first phase that
actually composes the builder from the IR (P1), and the foundation containers
(Phase 07) place leaves into sub-regions. The leaf composers + the body layout
are the substrate every later rendering phase reuses.

## 3. RFC sections implemented

- `RFC §10.2` — a deterministic, priority-ordered body layout (the minimal V1
  form: top-level nodes stack vertically in IR order within a margin-inset body
  region; `section_divider` overrides to full-bleed). The full container
  sub-layout is Phase 07; the constraint engine is not built (RFC: "not a full
  constraint solver").
- `RFC §10.6` — asset resolution at render time for `code_block` (the one
  asset-bearing leaf in scope); a missing/!resolvable asset surfaces a
  `LayoutWarning` and the node is skipped (graceful — see §16).
- `RFC §11.1` / `RFC §12` — per-node composers for the text-heavy leaves:
  `hero`, `prose`, `heading`, `list`, `divider`, `quote`, `callout`, `chip`,
  `arrow`, `code_block`, `section_divider`. Each follows its §12 policy (native
  shapes, except `code_block` → `pic`).

`image`, `chart`, `decoration`, `table`, `flow` and the containers
(`two_column`, `grid`, `card`, `card_section`) are later phases; encountering
one in Phase 06 emits a "not yet rendered" `LayoutWarning` and skips it.

## 4. Brief findings incorporated

No informing brief — the layout policy and per-node dispositions are specified
in RFC §10.2/§11/§12 and D-026 (engine, not product). `docs/research/INDEX.md`
lists a layout-engine survey only as a candidate.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-026` — engine, not product: **no render modes, no legibility heuristics,
  no text-size boosting, no render-time policy options.** A 9pt run renders at
  9pt. Composers map the IR to builder calls and nothing more.
- `D-011`/`D-018` — per-node policy is intrinsic; composers follow the §12 table
  (`code_block` → image, the rest native).
- `D-014` — `code_block` is a caller-rasterized `pic` + optional caption.
- `D-024` — assets resolve through the `AssetResolver`.
- `D-012`/`D-033` — run colors/typography flow through theme tokens; the scene
  theme is applied to the presentation before composing.

## 7. Architecture

`Render` becomes: apply the scene theme to the presentation, validate (Stage 1),
then for each `SceneSlide` add a `pptx.Slide`, lay out its top-level nodes, and
dispatch each to a per-node composer; accumulate `Stats`.

```text
scene/render.go        renderer, Render, body layout, node dispatch,
                       RichText → pptx.Paragraph mapping, Stats accounting
scene/render_leaves.go per-leaf composers (hero/prose/heading/list/divider/
                       quote/callout/chip/arrow/code_block/section_divider)
```

**Layout (V1, deterministic).** A body region = slide box inset by a uniform
margin. Top-level nodes stack top-to-bottom in IR order; each gets the full body
width and a per-node preferred height, separated by a `SpaceMD` gap. Total
height over the body emits an overflow `LayoutWarning` (not an error — RFC
§10.2). A `section_divider` node takes the full slide box. Composers in the
`scene` package call the public `pptx` builder only (P1); the
`scene/layout` package stays a placeholder until the container/grid engine
(Phase 07+).

**Text mapping.** A scene `RichText` maps onto a `pptx.Paragraph`: one builder
run per `TextRun`, carrying the run's bold/italic/underline/strike/code/link and
its `TextColor` (token → `pptx.TokenTextColor`, literal → `pptx.RGB`). The
node's base `TypeRole` sets the typography scale (prose→Body, heading→H1..H5,
hero title→Display, …); run-level `TypeRole` overrides are deferred (V1 leaf
typography is node-level). A link run becomes `Paragraph.AddHyperlink`.

**Idempotency.** The composition is deterministic (IR order, fixed heights, no
map iteration over emitted order, no randomness), so the same scene+theme yields
the same shapes (RFC §10.1). (ZIP entry timestamps are an OPC-layer concern,
out of scope here.)

## 8. Files added or changed

```text
scene/scene.go         # CHANGED — Render delegates to the renderer
scene/render.go        # NEW — renderer, layout, dispatch, richtext mapping, Stats
scene/render_leaves.go # NEW — per-leaf composers
scene/*_test.go        # NEW — render tests + a scene→pptx→reopen round-trip
test/integration/      # NEW — scene render conformance (the first scene→pptx seam)
scripts/smoke/phase-06.sh  # NEW
docs/glossary.md       # CHANGED if new vocab (body region / placement)
```

## 9. Public API surface

No new exported types (the surface was fixed in Phase 05). `Render` gains a real
body; `Stats` is populated (`Slides`, `Shapes`, `Assets`, `Warnings`).

## 10. Risks

- **R1 — layout determinism.** Non-deterministic placement breaks RFC §10.1.
  *Mitigation:* fixed per-node heights + IR-order iteration; no maps in the
  emit path; a round-trip + a "render twice → same shape model" check.
- **R2 — required-asset failure policy.** RFC §10.6 says a required asset that
  won't resolve fails the render; failing a whole deck for one missing image is
  harsh for previews. *Mitigation:* emit a `LayoutWarning` and skip the node
  (graceful degradation); documented as a §16 deviation — the caller inspects
  `Stats.Warnings`. The happy path (registered asset) is covered.
- **R3 — text-size fidelity.** *Mitigation:* the composer passes the run/node
  `TypeRole` straight to the builder; a test renders a theme whose body role is
  9pt and asserts `sz="900"` — no boosting (D-026).

## 11. Acceptance criteria

1. A scene with one of each in-scope text leaf renders to a conformant PPTX;
   native leaves emit native shapes and `code_block` emits a `pic`.
2. A `code_block` with a registered `AssetID` renders the image and its caption.
3. A scene whose body type role is 9pt renders the text at 9pt (no boosting).
4. A scene rendered with default options opens without the repaired prompt (the
   Phase 03 hygiene pass runs on save).
5. Round-trip: scene → PPTX → re-read the shape model (via `pptx.Open`).
6. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` green; prior
   smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default scene band (already configured) |

## 13. Smoke check

`scripts/smoke/phase-06.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` one-of-each-leaf render + conformance test passes.
3. `OK:` code_block image + caption test passes.
4. `OK:` 9pt-stays-9pt test passes.
5. `OK:` scene→PPTX→reopen round-trip test passes.

`SKIP` until landed; `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` — per-leaf composer output (shapes/XML spot-checks),
  layout stacking, asset-warning path.
- **Round-trip golden:** scene → PPTX → `pptx.Open` → assert shapes (criterion 5).
- **Integration** (`test/integration/`): conformance over a multi-leaf scene
  (the first scene→pptx seam; Deps name Phases 03/04).
- **Fuzz/Bench:** none.

## 15. Vocabulary added

- none required (the IR/render vocabulary landed in Phase 05). A "body region"
  / "placement" note may be added to the glossary if useful.

## 16. Plan deviations encountered during implementation

- **Required-asset failure is graceful (R2 resolved toward warnings).** RFC §10.6
  says a required asset that won't resolve fails the render; the implementation
  instead emits a `LayoutWarning` and skips the node, so a preview render with no
  resolver still produces a deck. Callers treat `Stats.Warnings` as fatal if they
  need to (no strict mode — RFC §10.2). The happy path (registered asset) is
  fully covered.
- **Run-level `TypeRole` is not used in V1 leaf rendering.** Each composer sets a
  node-level base `TypeRole` (prose→Body, heading→H{level}, hero title→Display,
  …) and maps each scene `TextRun`'s bold/italic/underline/strike/code/link +
  color; the run's own `TypeRole` field is ignored (node-level typography is the
  norm). Run-level type overrides can be honored in a later phase without an API
  change.
- **Composers live in the `scene` package** (`render.go` + `render_leaves.go`),
  not `scene/layout/text/` — the master plan allowed either; keeping them in
  `scene` avoids an import cycle (they call the public `pptx` builder, P1) and
  the `scene/layout` package stays a placeholder for the geometry engine.
- **`section_divider` is laid out full-bleed** (full slide box) rather than in
  the body stack, per RFC §10.2's "override body to full-bleed".

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages (scene 87.6%).
- [x] `scripts/smoke/phase-06.sh` reports `OK ≥ 5` and `FAIL = 0` (5 OK).
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (no new vocab needed).
- [x] Decision entries added (none — D-011/014/018/024/026 suffice).
