# pptx-go — Master Phase Plan

> Cross-cutting conventions + the wave-structured phase index. Each phase
> in the table below has (or will have) a `docs/plans/phase-NN-slug.md`
> plan file authored from `docs/plans/_template.md`.
>
> **Authoritative sources** (priority order):
> 1. `RFC-001-pptx-go.md`
> 2. This file
> 3. Per-phase plans (`docs/plans/phase-NN-*.md`)
> 4. `CLAUDE.md` / `AGENTS.md`
>
> Conflicts resolve toward the higher entry. The §16 phase-authoring
> workflow in `CLAUDE.md` is binding for every contributor starting a
> phase.

---

## 1. Cross-cutting conventions

### 1.1 Phase numbering

Two-digit, zero-padded, monotonically increasing. Numbers are reserved when
a phase plan lands; gaps are allowed (a planned-then-canceled phase
leaves its number unfilled rather than re-numbering downstream phases).

### 1.2 Wave grouping

Phases group into **waves**. A wave is a milestone with a coherent
deliverable. Wave boundaries get a read-only **checkpoint audit** —
the wave's punch list lands as one `chore(checkpoint)` PR
(`CLAUDE.md §17`).

### 1.3 Phase plan contract

Every phase plan, from `_template.md`:

- Names its owning subsystem and the RFC sections it implements.
- Lists `Deps:` — prior phases or external prereqs.
- Lists `Risks:` — known unknowns and how the plan handles them.
- Cites the informing research briefs (or notes "no informing brief —
  this is foundational"). Brief absence is a drift signal.
- States `Files added or changed:` in the same PR.
- States `Acceptance criteria:` — binding, smoke-checked.
- States `Coverage targets:` — defaults in `CLAUDE.md §11`; phase may
  raise but not lower.
- References the smoke script: `scripts/smoke/phase-NN.sh`.

### 1.4 Definition of done

A phase is **done** when:

1. Every acceptance criterion in the plan passes.
2. Coverage targets are met (`make coverage` clean).
3. `scripts/smoke/phase-NN.sh` reports `OK ≥ count(criteria)` and
   `FAIL = 0`.
4. Prior phases' smoke scripts still pass.
5. The §14 pre-merge checklist in `CLAUDE.md` is satisfied.

### 1.5 Reasonable deviations

Plans are specifications, not straitjackets. Document the deviation in the
PR body and update the plan in the same PR (`CLAUDE.md §4.3`). Silent
divergence is drift.

### 1.6 Coverage band defaults

Per-package coverage band defaults (override-per-phase in `coverage.json`):

| Class | Default |
|---|---|
| New `pptx` builder package | 85% |
| New `scene` renderer package | 80% |
| `internal/opc`, `internal/ooxml/*` (codecs, conformance-tested) | 85% |
| `internal/render`, `internal/ids` | 80% |
| CLI / tooling (V1: none) | 70% |
| Examples | not coverage-gated |

### 1.7 Smoke scripts

`scripts/smoke/phase-NN.sh` runs the acceptance criteria mechanically.
SKIPs gracefully when the surface isn't built yet. Format: prints
`OK: <criterion>` / `SKIP: <criterion> — <reason>` / `FAIL: <criterion>
— <details>`. Phase done requires `OK ≥ count(criteria)` and `FAIL = 0`.

The smoke script is a thin script — it doesn't re-implement the test
suite; it spot-checks user-visible behavior (the binary builds, the
example runs, the example output passes a round-trip read, etc.).

### 1.8 Round-trip golden tests

From Phase 03 onward, every phase that adds builder API ships a
round-trip golden test: write → read → assert model equality. The
`internal/golden` helper provides the harness. A phase that changes
on-the-wire shape must update goldens in the same PR with a one-line
rationale in the commit message.

### 1.9 Integration tests

A phase ships an integration test (`test/integration/`) when its `Deps:`
name a different subsystem's shipped phase, OR it closes a seam another
phase opened, OR it introduces a public interface other phases will build
on (`CLAUDE.md §17`).

### 1.10 PR title and branch convention

- Branch: `feat/phase-NN-slug`, `chore/phase-NN-slug`, `docs/phase-NN-slug`.
- PR title: `feat(<subsystem>): phase NN — <slug>`.
- PR body: link the phase plan, list deviations, link the RFC sections
  the phase implements.

### 1.11 PPTX validity layers (D-031)

Round-trip tests prove read-back fidelity, not validity. Emitted decks are
checked in four layers (cheapest first): **(1)** `internal/conformance` —
pure-Go OPC integrity, gating every emitted deck in tests; **(2)** `xmllint`
against vendored ISO 29500 transitional XSDs (`docs/specifications/`, SKIPs
until vendored); **(3)** a LibreOffice headless open-proxy CI job
(`.github/workflows/validate.yml`); **(4)** a manual per-wave PowerPoint
check (`docs/validation/POWERPOINT-CHECKS.md`). The harness landed before
Phase 03; Phase 03 (first complete deck + the D-020 hygiene pass) turns on
the full-deck completeness gate.

---

## 2. Wave map

```text
Wave 0 — Foundation                  Phase 00
Wave 1 — Theme + Builder spine       Phase 01–04
Wave 2 — Scene renderer spine        Phase 05–08
Wave 3 — Templates, masters, frames  Phase 09–11
Wave 4 — Curated assets + composites Phase 12–16
Wave 5 — Charts                       Phase 17
Wave 6 — Reading + round-trip        Phase 18–19
Wave 7 — Docs, skills, release       Phase 20–21
Wave 8 — Post-V1 engine extensions   Phase 22–29
Wave 9 — Typography & type system    Phase 30–…   (R9 engine)
Wave 10 — Content fit & density       (R10 engine)
Wave 11 — Rendering robustness        (R11 engine)
Wave 12 — Component primitives        (R12 engine)
Wave 13 — Backgrounds & finish        (R13 engine)
Wave 14 — Coverage classes            (R14 engine)
Wave 15 — Theme/soul engine bits      (R8 engine: dark palette, multi-accent, gradients, dark ext)
```

Each wave ends with a checkpoint audit (`CLAUDE.md §17`). V1.0.0 ships at
the end of Wave 7. **Wave 8** was the first post-V1 wave: caller-driven engine
mechanisms requested by the product built on pptx-go
(`DECKARD-PRODUCT-REQUIREMENTS.md` R1–R7). Each is additive and backward-compatible —
a new optional capability whose zero value reproduces the prior render
byte-for-byte (the one intentional exception is Phase 22, which changes
multi-line text layout by design).

**Waves 9–15** implement the *professional-bar* requirements
(`DECKARD-PRODUCT-REQUIREMENTS.md` R8–R14, "Deckard Wave 2"). The requirements
doc tags each sub-requirement `engine` / `product` / `both`: pptx-go implements
the **engine** mechanisms (and the engine side of `both`); the `product`-tagged
requirements operate on Deckard's own packages (`internal/soul/`, `contracts/`,
`exportstore/`) and are out of scope for this repo (see **D-059**). The same
invariants hold — additive, deterministic, byte-identical when unused,
mechanism-not-taste (D-026). Waves group by requirement family; each
sub-requirement (or a tightly-coupled cluster) is one phase / PR.

---

## 3. Phase index

### Wave 0 — Foundation

#### Phase 00 — Module rename and hygiene scaffolding

**Subsystem:** Repo / build / docs
**RFC sections:** §3.3, §3.4, §5
**Plan:** `docs/plans/phase-00-foundation.md`
**Deps:** none.
**What lands:**
- Module rename to `github.com/hurtener/pptx-go`. Every import path
  updated. `go.mod`, `go.work`, `go.sum`, all `.go` files.
- LICENSE updated to Apache-2.0; upstream MIT preserved at
  `LICENSE.upstream` per MIT attribution.
- `Makefile` with canonical targets (build, test, coverage, vet, lint,
  drift-audit, check-mirror, preflight, install-hooks).
- `scripts/preflight.sh`, `scripts/drift-audit.sh`, `scripts/smoke/_template.sh`,
  `scripts/hooks/pre-commit`, `scripts/install-hooks.sh`.
- `.github/workflows/ci.yml` running `make preflight` + `make test` +
  `make coverage` + `make check-mirror` on every push and PR.
- `.golangci.yml`, `.editorconfig`, `.gitignore`.
- `CHANGELOG.md` (Keep a Changelog format).
- `docs/plans/_template.md`, `docs/plans/phase-00-foundation.md` (this
  phase's own plan — eats its own dogfood).
- The `internal/coveragecheck` package + initial `coverage.json`.
**Acceptance criteria:**
- `make build` succeeds (upstream code compiles under the new module
  name with no behavior change).
- `make test` passes (upstream tests still green).
- `make preflight` passes.
- `make check-mirror` passes (`AGENTS.md == CLAUDE.md`).
- `.github/workflows/ci.yml` green on a sample PR.

---

### Wave 1 — Theme + Builder spine

#### Phase 01 — OPC + OOXML reorg

**Subsystem:** internal/opc, internal/ooxml
**RFC sections:** §6
**Plan:** `docs/plans/phase-01-opc-ooxml-reorg.md`
**Deps:** Phase 00.
**What lands:**
- Move upstream `opc/` → `internal/opc/` (rename only; no behavior change).
- Reorganize upstream `parts/` → `internal/ooxml/{presentation, slide,
  theme, core, drawing, relations, media}/`. Each subpackage's types
  are extracted from the upstream's monolithic `parts/` package.
- `internal/ooxml/namespaces.go` — canonical namespace URIs.
- All upstream tests preserved and re-located alongside their new
  packages.
**Acceptance criteria:**
- `make build` and `make test` pass after the reorg.
- A spot-check round-trip test (open upstream's fixture PPTX, save,
  reopen, assert byte-identical-ish equivalence — modulo allowed
  reordering) passes.
- `make drift-audit` enforces no import of `internal/...` from outside
  the module.

#### Phase 02 — Theme & token model + font embedding

**Subsystem:** pptx (theme + fonts)
**RFC sections:** §7, §7.6
**Deps:** Phase 01.
**What lands:**
- `pptx/theme.go` — `Theme`, `ColorPalette`, `Typography`, `Spacing`,
  `Radii`, `Elevations`.
- `pptx/tokenresolve.go` — `Resolve(token, theme) → OOXMLValue`.
- `pptx/units.go` — `Pt`, `Cm`, `In`, `Px`, EMU constants.
- `pptx/geom.go` — `Box`, `Position`, `Size`, `Anchor`, `Inset`.
- `pptx/fonts.go` — `FontSource` interface, default OS-font source,
  `pptx.WithFontSource(...)`, `pres.EmbedFont(name, style, weight)`.
  **No auto-embedding** — caller controls (D-019).
- Font-embedding writer in `internal/ooxml/embeddings/` —
  `embeddings/*.fntdata` parts + presentation-level font references.
- Default theme + `templates/_default-theme.pptx` (Phase 02 scaffolds;
  Phase 03 wires the default into `pptx.New`).
- Theme-to-OOXML codec in `internal/ooxml/theme/`.
**Acceptance criteria:**
- A `Theme` can be constructed in Go and round-tripped through OOXML.
- A `Theme` can be loaded from a brand `.pptx` template's `theme1.xml`.
- A token resolves to the same OOXML value across two writes with the
  same theme.
- A theme swap re-resolves: same builder state, two themes → two
  different OOXML colors.
- A presentation with a registered `FontSource` and an explicit
  `pres.EmbedFont(...)` call ships the font bytes (round-trip
  preserves the embed).
- A presentation with no `EmbedFont` calls ships no embedded fonts.

#### Phase 03 — Builder spine (Presentation, Slide, Shape, Media, Sections, Notes, Format)

**Subsystem:** pptx (core builder)
**RFC sections:** §5, §8.1–8.3, §8.6, §8.7, §8.8
**Deps:** Phase 02.
**What lands:**
- `pptx/presentation.go` — `Presentation`, `New`, `Open`, `OpenStream`,
  `Save`, `SaveStream`, `Write`, `AddSlide`, `Slides`, `Theme`, `SetTheme`,
  `AddSection`, `Sections`, `Close`. `pptx.WithFormat(...)` accepts
  `Slides16x9` (default) and `Slides4x3`.
- `pptx/slide.go` — `Slide`, `AddShape`, `AddImage`, `AddGroup`,
  `AddConnector`, `Background`, `ID`, `Layout`, `SpeakerNotes`.
- `pptx/section.go` — `Section`, `Include(slide)`, `Slides()`,
  `Name()`, `SetName()`.
- `pptx/notes.go` — speaker notes `TextFrame` + `notesSlide` part
  emission.
- `pptx/shape.go` — `Shape`, `ShapeGeometry`, `Fill`, `Line`, `SolidFill`,
  `GradientFill`, `PatternFill`, `BlipFill`, `NoFill`.
- `pptx/media.go` — `Image`, `ImageSource`, `ImageFile`, `ImageBytes`,
  `ImageReader`, alt-text, crop, fit.
- `pptx/stream.go` — streaming open/save passthrough to `internal/opc`.
- `internal/render/hygiene.go` — always-on XML repair-prompt
  post-processor (D-020). Runs as part of every write; no caller-facing
  option. Trigger list documented in `docs/design/HYGIENE.md`.
- The upstream `pptx/` is migrated incrementally; new files supersede
  old ones; old API kept as deprecated aliases where the new API isn't
  a drop-in.
**Acceptance criteria:**
- Construct a 1-slide presentation with a rect + an image; save; reopen
  via the new builder; assert shape model equals what was written
  (round-trip golden).
- A presentation with `pptx.Slides4x3` produces a deck with 4:3 canvas
  dimensions and an unbroken round-trip.
- A slide's `SpeakerNotes()` text round-trips losslessly.
- A `Section` containing slides round-trips losslessly; PowerPoint
  opens the deck with the section visible in the slide-sorter.
- Open a sample upstream-authored deck and re-emit byte-equivalently
  via streaming save.
- Emitted decks open in PowerPoint without the "this file has been
  repaired" prompt (the hygiene post-processor runs unconditionally
  on every write).

#### Phase 04 — Rich text model

**Subsystem:** pptx (text)
**RFC sections:** §8.4, §9
**Deps:** Phase 03.
**What lands:**
- `pptx/text.go` — `TextFrame`, `Paragraph`, `Run`, `RunStyle`,
  `BulletKind`, `Alignment`, `AutoFitMode`.
- `pptx/text_hyperlink.go` — hyperlink runs.
- `pptx/text_layout.go` — paragraph layout helpers.
- Round-trip tests for: plain run, bold/italic/underline run, colored
  run (token), hyperlinked run, inline-code run, list bullets, numbered
  lists, checklist (rendered as bullet with check glyph).
**Acceptance criteria:**
- A `TextFrame` with multiple paragraphs and styled runs round-trips
  losslessly.
- An inline-code run (`Run.Code = true`) renders with mono font + subtle
  background tint per the default theme.
- A hyperlinked run carries the URL through the relationships layer.

---

### Wave 2 — Scene renderer spine

#### Phase 05 — Scene package scaffold + IR catalog + AssetResolver

**Subsystem:** scene (types only)
**RFC sections:** §10.1, §10.6, §11
**Deps:** Phase 04.
**What lands:**
- `scene/scene.go` — `Scene`, `SceneSlide`, `Render` (no-op stub),
  `RenderOption`, `Stats`, `LayoutWarning`.
- `scene/nodes.go` — every IR node struct from `RFC §11` (leaves +
  containers) with discriminated `type` field.
- `scene/richtext.go` — `TextRun`, `RunStyle`, `TextColor` (token +
  literal).
- `scene/tokens.go` — `ColorRole`, `TextColorRole`, `TypeRole`,
  `SpaceRole`, `RadiusRole`, `ElevationRole` re-exports from `pptx`.
- `scene/validate.go` — Stage 1 validation: well-formed unions, field
  constraints.
- `scene/asset.go` — `AssetID`, `AssetResolver` interface,
  `URIAssetResolver` helper for `asset://`-prefixed URIs. (D-024)
- `scene/policy.go` — documentation/test file asserting per-node
  rendering policy per `RFC §12`. The policy is intrinsic to whether
  the node's IR carries an `asset_id` field (D-018).
- `scene/layout/` — placeholder package; layout engine lands in
  subsequent phases.
**Acceptance criteria:**
- The full IR catalog compiles.
- Stage 1 validation correctly accepts/rejects fixture scenes (a fixture
  per node type, plus negatives).
- `scene.Render` is callable and returns a zero `Stats` on an empty
  `Scene`.
- `URIAssetResolver` resolves `asset://<uuid>` URIs to bytes via the
  caller's callback.
- A compile-time assertion (in `scene/policy_test.go`) verifies that
  every node type whose IR carries an `asset_id` field matches the
  per-node policy table in `RFC §12.1`.

#### Phase 06 — Leaf-node rendering

**Subsystem:** scene (text leaves)
**RFC sections:** §11.1, §12 (rows: hero, prose, heading, list,
divider, quote, callout, chip, arrow, code_block, section_divider)
**Deps:** Phase 05.
**What lands:**
- Per-node composers under `scene/layout/text/` (or `scene/render_*.go`
  — phase plan picks). Each composes `pptx` calls per the per-node
  rendering policy (`RFC §12`).
- Code_block path: the IR's `asset_id` resolves through `AssetResolver`;
  the renderer composes a `pic` shape + optional caption text shape.
- No render modes. No legibility heuristics. No render-time policy
  options. Product behavior lives in callers (D-026).
**Acceptance criteria:**
- A scene with one of each text-heavy leaf renders to a PPTX whose
  shape count matches the per-node policy table.
- A code_block with a registered `AssetID` renders the image and the
  caption.
- A scene with text at 9pt renders the text at 9pt (the library does
  not boost text sizes).
- A scene rendered with default options produces a PPTX that opens in
  PowerPoint without the "repaired" prompt (the always-on hygiene
  pass from Phase 03 runs).
- Round-trip golden: scene → PPTX → re-read shape model.

#### Phase 07 — Container nodes (two_column, grid)

**Subsystem:** scene (containers)
**RFC sections:** §11.2 (two_column, grid)
**Deps:** Phase 06.
**What lands:**
- `scene/layout/twocolumn.go` — ratio + gap; cell layout.
- `scene/layout/grid.go` — columns + weighted ratio + align_items + gap.
- Layout warnings for overflow.
**Acceptance criteria:**
- `1:1`, `1:2`, `2:1` two_column ratios produce correct cell widths.
- Grid with 2/3/4 columns and a weighted ratio produces correct cell
  widths.
- A grid cell count not matching `columns × rows` raises a Stage 1
  validation error.

#### Phase 08 — Table

**Subsystem:** pptx (table) + scene (table node)
**RFC sections:** §8.5, §11.1 (table)
**Deps:** Phase 04 (text), Phase 07 (containers — for tables-in-grids).
**What lands:**
- `pptx/table.go` — `Table`, `Cell`, header rows, banding, merged cells,
  cell-level borders + fills.
- `scene/render_table.go` — composes the builder for a `Table` IR node.
- Header row / banding driven by `Table.headers` presence.
**Acceptance criteria:**
- Table with merged cells round-trips losslessly.
- Banded table alternates fills correctly.
- Table caption renders above the table.

---

### Wave 3 — Templates, masters, frames

#### Phase 09 — Template ingestion (Theme + Masters)

**Subsystem:** pptx (template) + internal/ooxml
**RFC sections:** §13
**Deps:** Phase 02, Phase 03.
**What lands:**
- `pptx.FromTemplate(brand)` — option for `pptx.New` to seed
  presentation from a template (masters + theme + default layouts
  copied).
- `pptx/master.go` — `Master`, `Layout`, `LayoutMap`.
- `scene.WithTheme`, `scene.WithLayoutMap` — render options.
- (`pptx.LoadTheme(path)` already shipped with Phase 02's theme work —
  `pptx/themecodec.go`; Phase 09 consumes it. See
  `docs/plans/phase-09-template-ingestion.md` §5.)
**Acceptance criteria:**
- Loading a PowerPoint-emitted template's theme produces a `Theme`
  whose `Resolve(ColorAccent)` returns the template's accent.
- A scene rendered with `scene.WithTheme(brandTheme)` uses the brand's
  colors.

#### Phase 10 — Frame chrome

**Subsystem:** assets/frames + scene
**RFC sections:** §14.3
**Deps:** Phase 09.
**What lands:**
- `assets/frames/{browser,phone,desktop,laptop}.go` — shape recipes.
- `scene/frames/` — frame registry, extension hook.
- `scene/render_image.go` (extension) — wraps an image with a frame
  when `Image.Frame != none`.
**Acceptance criteria:**
- Each curated frame renders the inner image inside the bezel region.
- A caller-extended frame works through `scene.WithFrameExtension`.

#### Phase 11 — Image node + media manager refactor

**Subsystem:** scene (image) + pptx (media)
**RFC sections:** §8.6, §11.1 (image)
**Deps:** Phase 10.
**What lands:**
- `scene/render_image.go` — full image node composition (asset
  resolution, alt text, crop, fit, frame).
- `pptx/media.go` — refactor of upstream media manager: dedup pool
  moved to `internal/opc` (or a new `internal/media`), alt-text first
  class, MIME detection.
**Acceptance criteria:**
- Inserting the same image twice writes one part (dedup).
- Alt text round-trips.
- A frame + image renders the composite correctly.

---

### Wave 4 — Curated assets + composites

#### Phase 12 — Curated icons

**Subsystem:** assets/icons + scene/icons + internal/render (SVG→OOXML)
**RFC sections:** §14.1
**Deps:** Phase 09.
**What lands:**
- `assets/icons/<name>.svg` — initial set of ~60 lucide-subset icons.
- `internal/render/svgpath.go` — SVG path → OOXML preset/path geom
  translator (single path, solid fill).
- `scene/icons/registry.go` — closed-name registry + extension hook.
**Acceptance criteria:**
- Each curated icon renders as a native PPTX shape path.
- Caller extension via `scene.WithIconExtension(name, svg)` works.
- A icon SVG that violates the translator constraints fails at
  registration (not at render).

#### Phase 13 — Curated ornaments + Decoration node

**Subsystem:** assets/ornaments + scene (decoration)
**RFC sections:** §14.2, §11.1 (decoration)
**Deps:** Phase 12.
**What lands:**
- `assets/ornaments/<name>.go` — six preset recipes.
- `scene/ornaments/registry.go` — closed-name registry + extension hook.
- `scene/render_decoration.go` — `Decoration` IR composition: anchor +
  offset + bleed + opacity + rotation + size.
**Acceptance criteria:**
- Each curated ornament renders at the named anchor.
- A bleed-anchored ornament uses negative offsets correctly.
- Foreground vs background layer ordering is honored.

#### Phase 14 — Card + CardSection

**Subsystem:** scene (composites)
**RFC sections:** §11.2 (card, card_section), §12
**Deps:** Phase 07, Phase 12, Phase 13.
**What lands:**
- `scene/render_card.go` — Card chrome: rounded rect + accent stripe +
  optional icon + eyebrow + header_pill + body + fill/border/elevation.
- `scene/render_card_section.go` — CardSection (top-level container
  with card chrome accepting grids/two_columns/cards inside).
- All v4 card knobs (`fill`, `border_style`, `size`, `elevation`,
  `body_layout`, `layout`, `header_pill`) implemented.
**Acceptance criteria:**
- Each card variant from the pengui-slides Galici/Databricks reference
  decks renders correctly.
- Card-of-cards composition via `card_section` works.

#### Phase 15 — Flow

**Subsystem:** scene (flow)
**RFC sections:** §11.1 (flow), §12
**Deps:** Phase 14 (uses card-like step pill).
**What lands:**
- `scene/render_flow.go` — sequential step pipeline: step pills + per-
  pair connectors (`arrow`, `arrow_dashed`, `cycle`, `plus`).
- Horizontal + vertical directions.
**Acceptance criteria:**
- A 4-step horizontal flow with arrow connectors renders correctly.
- A `cycle` flow appends a return-arrow after the last step.
- A vertical flow rotates connectors.

#### Phase 16 — CodeBlock (raster path)

**Subsystem:** scene (code_block)
**RFC sections:** §11.1 (code_block), §12 (D-014)
**Deps:** Phase 11.
**What lands:**
- `scene/render_code_block.go` — finalize the raster path: caller
  provides `AssetID` of pre-rendered code image; renderer composes
  image + caption + optional language badge.
**Acceptance criteria:**
- A code_block with a registered raster renders correctly.
- Caption renders below the raster.

---

### Wave 5 — Charts

#### Phase 17 — Chart (image-shape V1)

**Subsystem:** scene (chart) + pptx (chart placeholder helper)
**RFC sections:** §15.1, §11.1 (chart)
**Deps:** Phase 11.
**What lands:**
- `scene/render_chart.go` — image-shape disposition: composes a
  `pic` shape from the caller-supplied `AssetID`; caption below.
- `pptx.ChartPlaceholder(box)` builder helper — sizes and positions a
  chart slot without committing bytes.
- Aspect-ratio warning when caller bytes don't match slot.
**Acceptance criteria:**
- A chart node with a PNG raster renders at the assigned slot.
- An aspect-ratio mismatch surfaces a `LayoutWarning`.

### Wave 6 — Reading + round-trip

#### Phase 18 — Round-trip read of self-authored decks

**Subsystem:** pptx (read) + internal/ooxml (parsers)
**RFC sections:** §16
**Deps:** Phase 03 onward (every shipped builder API has a parser
counterpart).
**What lands:**
- `pptx.Open` / `pptx.OpenStream` build a full builder model from a
  pptx-go-authored deck: every shape, text run, fill, line, table,
  image is reconstructed.
- A comprehensive `test/integration/roundtrip_test.go` walks every
  shipped IR node + every shipped builder primitive.
**Acceptance criteria:**
- Every IR node V1.0.0 ships, round-trip is lossless.
- A fixture deck authored by V1.0.0 reopens byte-identically (modulo
  documented permissible reorderings).

#### Phase 19 — External-deck read robustness (best-effort)

**Subsystem:** pptx (read) + internal/ooxml
**RFC sections:** §16
**Deps:** Phase 18.
**What lands:** (scope set by **D-048** — best-effort graceful degradation, not
opaque-carrier preservation; the RFC parks fidelity preservation in V2)
- Parsers gracefully **ignore** unrecognized OOXML shape-tree elements and
  surface them in `Presentation.ReadWarnings()` (warn, don't preserve).
- Parts pptx-go does not model **pass through unchanged** on re-save (the OPC
  pass-through; "`RawPart`" is realized as that, not a new carrier type). Opaque
  `RawShape` *preservation* of unrecognized shapes is deferred to V2.
- Documented degradation modes when external-deck features don't map
  to the builder model.
**Acceptance criteria:**
- A library of synthetic external-style sample decks loads without panic.
- Unsupported elements surface in a `ReadWarnings` slice.

---

### Wave 7 — Docs, skills, release

#### Phase 20 — Agent skills + published docs site

**Subsystem:** docs/site, skills
**RFC sections:** `CLAUDE.md §19`
**Deps:** Wave 6 complete.
**What lands:**
- `skills/` — the eight SKILL.md workflows per `CLAUDE.md §19`: "scaffold a
  presentation", "define a Theme", "load a brand template", "compose a
  scene", "embed a chart raster", "embed a code-block raster", "extend the
  icon set", "register an asset". (Reconciled from the earlier six-skill
  wording to the authoritative `CLAUDE.md §19` list — see
  `phase-20-skills-docs.md` §5.)
- `docs/site/` — VitePress (or similar) docs site with quickstart, API
  reference, scene catalog, theme guide, examples.
- `.github/workflows/pages.yml` — CI to build and deploy on push.
- The §19 hook in `drift-audit.sh` activates: a user-facing surface
  change touching `pptx/` or `scene/` requires a matching `docs/site/`
  / `skills/` update in the same PR.
**Acceptance criteria:**
- Docs site builds cleanly and deploys to GitHub Pages.
- Each shipped node has a docs page with a runnable example.
- Each skill has a passing smoke run.

#### Phase 21 — v0.1.0 release prep

**Subsystem:** repo / release
**RFC sections:** §5, §24
**Deps:** Phase 20.
**What lands:**
- `CHANGELOG.md` — V1.0.0 (or v0.1.0 — semver track) section.
- `RELEASING.md` — release procedure.
- `docs/V2-BACKLOG.md` — consolidated V2 deferrals (RFC §24).
- A GitHub Release with attached examples (a "Hello, pptx-go" sample
  output deck).
**Acceptance criteria:**
- `git tag v0.1.0`-able with a green release dry-run.
- Sample example renders the canonical Galici-style deck end-to-end.

### Wave 8 — Post-V1 engine extensions

Caller-driven engine mechanisms requested by the product built on pptx-go,
specified in `DECKARD-PRODUCT-REQUIREMENTS.md` (R1–R7). Each is additive and
deterministic; the engine stays unopinionated (D-026) — it provides the
mechanism, the caller supplies the values.

#### Phase 22 — content-aware text height

**Subsystem:** scene
**RFC sections:** §10.2
**Deps:** Phase 13 (alignment + metrics), Phase 05–08 (scene spine).
**What lands:**
- `scene/metrics.go` — `wrappedLines`, a deterministic
  `ceil(naturalWidth / availableWidth)` line-count estimate built on the
  existing pinned char-width model.
- `scene/render.go` — `preferredHeight`/`nodesHeight` become content-aware
  (`Prose`, `List`, `Heading`, `Quote`, `Callout`, `Table`); the overflow
  `LayoutWarning` now fires on real wrapped overflow.
**Acceptance criteria:**
- A paragraph that wraps to N lines is allotted ≥ N line-heights; the next
  stacked node does not overlap it.
- A slide whose wrapped content exceeds the body region emits the overflow
  warning (it did not before).
- Single-line content is byte-identical; determinism holds under N workers.
**Note (the one intentional break):** this phase changes layout for multi-line
text by design (less overlap, truthful overflow). Single-line content is
unaffected; there are no byte-golden snapshots to regenerate (determinism is
proven by parallel≡sequential equality).

#### Phase 23 — grow-to-fit

**Subsystem:** scene
**RFC sections:** §10.2
**Deps:** Phase 13 (alignment), Phase 22 (content-aware height).
**What lands:**
- `scene/align.go` — `VAlignFill`, a new opt-in body-stack vertical alignment.
- `scene/render.go` — `alignedStackIn` distributes leftover body height to the
  flexible nodes (`Grid`, `TwoColumn`, `Card`, `CardSection`, `Table`, `Chart`,
  `Image`) so they grow to fill the frame; fixed leaves stay at preferred
  height. No container renderer changes — the geometry engine already honors a
  taller box.
**Acceptance criteria:**
- A heading + grid under `VAlignFill` pins the heading at top and grows the grid
  to the bottom margin; a taller slot yields proportionally taller cells.
- Two flexible nodes share the slack proportional to preferred height.
- Every non-fill mode is byte-identical; determinism holds under N workers.

#### Phase 24 — slide chrome

**Subsystem:** scene
**RFC sections:** §10.2, §10.6
**Deps:** Phase 06 (leaves), Phase 11 (Image/asset resolution).
**What lands:**
- `scene/scene.go` — a `Chrome` struct on `Scene` (brand slot + page total +
  `Enabled`) and `SceneSlide.Section` / `.PageNumber` fields.
- `scene/chrome.go` — opt-in chrome drawn outside a shrunk body region: a top
  section eyebrow + hairline rule, and a bottom footer with a brand slot
  (text or image asset) and an `N / total` page number. Native shapes reusing
  existing tokens; a brand image forces sequential composition.
**Acceptance criteria:**
- A chrome-enabled deck renders a consistent footer page number + per-slide
  section eyebrow outside the body box; the body region shrinks to make room.
- Page total and per-slide number auto-derive and are overridable.
- Chrome disabled is byte-identical; determinism holds under N workers.

#### Phase 25 — rich card visuals

**Subsystem:** scene
**RFC sections:** §11.2
**Deps:** Phase 14 (Card), Phase 13 (token-alpha).
**What lands:**
- `scene/nodes.go` — three additive `Card` fields: `HeaderFill *ColorRole`
  (colored header band, body keeps `Fill`), `StatusDot *ColorRole` (top-right
  dot), `Watermark string` (large faint label behind the body).
- `scene/render_card.go` — renders the band (rounded rect to the header bottom),
  the dot (ellipse), and the watermark (low-opacity `TokenColorAlpha` run), all
  reusing existing tokens.
**Acceptance criteria:**
- A card with all three set renders all three; each is omitted when unset.
- A bare card is byte-identical; determinism holds under N workers.

#### Phase 26 — column join

**Subsystem:** scene
**RFC sections:** §11.2
**Deps:** Phase 07 (containers).
**What lands (R5 sub-units a+b):**
- `scene/nodes.go` — a `ColumnJoin` enum (`JoinNone`/`JoinBadge`/`JoinArrow`)
  and `TwoColumn.Join` / `TwoColumn.JoinLabel`.
- `scene/render_container.go` — a centered inter-column element on the seam: a
  "VS"-style accent badge (ellipse + label) or a connector arrow.
**Acceptance criteria:**
- A two-column with a center badge renders the badge on the seam; with a
  connector renders the arrow; `JoinNone` is byte-identical; determinism holds.

*(R5 sub-unit (c), the row-labeled bento grid, lands as a separate phase.)*

#### Phase 27 — bento grid

**Subsystem:** scene
**RFC sections:** §11.2, §10.4
**Deps:** Phase 07 (containers), Phase 22/23 (height/flex).
**What lands (R5 sub-unit c):**
- `scene/nodes.go` — a new `Bento` container node (`Bento{Columns, Rows}`,
  `BentoRow{Label, Cells}`, `BentoCell{Span, Node}`), wired through every node
  switch (policy, validation, render, flex, asset/icon/image/decoration walks).
- `scene/render_bento.go` — rows with an optional left label and cells of
  variable column span on a shared column grid (absolute spans align columns).
**Acceptance criteria:**
- A row-labeled bento renders labels + span-aligned cells; an unlabeled bento
  reserves no gutter; Stage-1 rejects malformed bento; the catalog has 21 kinds
  and the every-node round-trip covers `Bento`; determinism holds.

*(Completes R5.)*

#### Phase 28 — stat node

**Subsystem:** scene
**RFC sections:** §11.1, §10.4
**Deps:** Phase 06 (leaves), Phase 07 (Grid).
**What lands (R6):**
- `scene/nodes.go` — a `Stat` leaf node (`Value`, `Label`, `Delta string`,
  `DeltaTone`) for a hero number with a label and an optional directional delta.
- `scene/render_stat.go` — value at display scale + label + a delta colored by
  tone (`ColorSuccess`/`ColorError`/`TextMuted`). A `Grid` of `Stat`s is a
  metric/pricing strip.
**Acceptance criteria:**
- A stat renders value + label (+ toned delta); a Grid of stats renders a strip;
  Stage-1 rejects an empty value; the catalog has 22 kinds; determinism holds.

#### Phase 29 — resolved colors

**Subsystem:** scene
**RFC sections:** §10.1, §13.3
**Deps:** Phase 05–06 (Stats), the `VariantDark` derived palette.
**What lands (R7, completes Wave 8):**
- `scene/scene.go` — a `SlideColors{SlideID, Canvas, Surface, PrimaryText}` type
  and a `Stats.Colors []SlideColors` field, captured per slide from the theme the
  slide actually rendered with (the derived dark palette for `VariantDark`).
- No contrast logic in the engine — it exposes the resolved RGBs; the caller
  computes its own ratios/thresholds.
**Acceptance criteria:**
- After `Render`, a caller reads one scene-ordered `SlideColors` per slide; a
  dark slide's colors are the dark palette and differ from a light slide's; the
  rendered bytes are unchanged; deterministic.

*(Completes Wave 8 — R1–R7.)*

---

### Wave 9 — Typography & type system

The R9 (`DECKARD-PRODUCT-REQUIREMENTS.md`) engine units. Phases 30–34 land the
"designed type" foundation — letter-spacing (R9.3, D-060), line-height (R9.4,
D-061), case transform (R9.11, D-062), display-face role (R9.2, D-063), and
per-face width metrics (R9.5, D-064). Phase 35 closes the font cluster's gating
unit; R9.6–R9.8 (fallback stack, emphasis-as-italic, weight-aware embedding)
follow.

#### Phase 35 — font-embedding pipeline

**Subsystem:** pptx — Layer 1 builder (font embedding)
**RFC sections:** §7.6
**Deps:** Phase 33 (display-face role, D-063); the `EmbedFont`/`FontSource`
mechanism (D-019).
**What lands (R9.1, engine half — D-059):**
- `pptx.WithFontEmbedding()` — an opt-in save-time pass that walks every slide's
  runs, collects the distinct `(family, bold, italic)` faces in stable sorted
  order, and `EmbedFont`s each via the registered `FontSource`.
- `internal/ooxml/slide.SlidePart.UsedFontFaces()` — the codec-side run walk
  (shape + table-cell text bodies).
- `internal/ooxml/presentation.PresentationPart.HasEmbeddedFace()` — dedup vs a
  manual `EmbedFont`.
- Warn-don't-fail on a missing face; byte-identical when off / no source;
  deterministic part order regardless of worker count.
**Acceptance criteria:**
- A themed deck with a `FontSource` + `WithFontEmbedding` embeds every used face;
  two saves are byte-identical; the flag off (or no source) is byte-identical to
  the prior output; a manual `EmbedFont` is not duplicated; a missing face warns
  and the Save still succeeds.

#### Phase 36 — font fallback chain

**Subsystem:** pptx — Layer 1 builder (theme/typography + font embedding)
**RFC sections:** §7.6
**Deps:** Phase 35 (font-embedding pass, D-065); the `FontSource` mechanism (D-019).
**What lands (R9.6, engine half — D-059):**
- `pptx.FontSpec.Fallback []string` — a per-role ordered substitute chain.
- A save-time `resolveFontFallbacks` pass: when a `FontSource` cannot resolve a
  role's primary family, the run's single-valued `a:latin` is rewritten to the
  first family in `[Family] + Fallback` the source resolves (a controlled
  near-match instead of a host default).
- `internal/ooxml/slide.SlidePart.RewriteFontFaces()` — the codec-side rewrite.
- Self-gated (no `FontSource` or no declared chain ⇒ byte-identical); the
  availability oracle is the `FontSource`; deterministic + idempotent across saves.
**Acceptance criteria:**
- A deck whose primary face the source cannot resolve renders in the declared
  fallback (not the host default); the primary wins when the source resolves it;
  an empty chain or no `FontSource` is byte-identical; two saves are
  byte-identical; with embedding on, the resolved fallback face is what is
  embedded.

#### Phase 37 — italic-aware font fallback (emphasis-as-italic-display)

**Subsystem:** pptx — Layer 1 builder (theme/typography + font embedding)
**RFC sections:** §7.6
**Deps:** Phase 33 (display family, D-063), Phase 35 (embedding, D-065),
Phase 36 (fallback, D-066).
**What lands (R9.7, engine half — D-059):**
- Italic-aware fallback: `resolveFontFallbacks` resolves per `(family, italic)`,
  probing the italic cut for italic runs and the regular cut for upright ones, so
  an italic emphasis run whose family lacks an italic cut falls back to an
  italic-capable face (not a faux-italic) while its upright runs keep the primary.
- `internal/ooxml/slide.SlidePart.RewriteFontFaces` generalized to a resolver
  callback `func(typeface string, bold, italic bool) string`.
- Fix: `<p:font>` in `embeddedFontLst` was emitted bare (`<font>`) — invalid
  OOXML PowerPoint cannot bind; `font` added to the `RestoreNamespaces` prefix
  map. The display-italic guarantee itself was already delivered by D-063+D-065
  (now covered by a verification test).
**Acceptance criteria:**
- An italic run of a regular-only family falls back to an italic-capable face;
  upright runs keep the primary; an italic display run embeds the display italic
  cut; the embedded `<p:font>` carries the `p:` prefix; byte-identical when
  unused; deterministic.

#### Phase 38 — weight-aware font embedding

**Subsystem:** pptx — Layer 1 builder (font embedding)
**RFC sections:** §7.6
**Deps:** Phase 35 (embedding, D-065).
**What lands (R9.8, engine half — D-059; final R9 unit, R9.12 → V2):**
- Per-run resolved weight tracking: `slide.XTextProperties.Weight` (`xml:"-"`,
  never serialized) set by `toProps`; `slide.FontFace.Weight` populated by
  `UsedFontFaces` (inferred from the bold bit when 0).
- `autoEmbedFonts` keys on `(family, weight, italic)` and embeds the actual
  weight file nearest each OOXML bucket's nominal (400/700), so a medium (500)
  role ships the medium file; colliding weights coalesce per bucket (logged).
**Acceptance criteria:**
- A medium-weight role embeds the medium file (the provider is asked for the
  resolved weight); a single-weight deck embeds one file; colliding weights
  coalesce to the nearest-nominal winner; byte-identical when unused;
  deterministic. (Embedding one file per OOXML bucket, not per numeric weight —
  D-068.)

---

### Wave 10 — Content fit & density

The R10 (`DECKARD-PRODUCT-REQUIREMENTS.md`) engine units: pro slides pack dense
content compactly and fill the frame without overflowing or floating. Opens with
the two CRITICAL off-slide/overlap fixes (R10.1 card-header height, R10.2
fit-to-region compression).

#### Phase 39 — content-aware card header height

**Subsystem:** scene — Layer 2 renderer (card chrome / layout)
**RFC sections:** §10.1, §12.1
**Deps:** brief 09 (`wrappedLines`), the D-054 card chrome.
**What lands (R10.1, CRITICAL · engine):**
- `cardHeaderColumnW` + `cardHeaderRowHeights` size the card eyebrow/title (and
  the D-054 header band + the body region top) to `wrappedLines × per-row`, so a
  header that wraps to N lines no longer collides with the body.
- Single-line headers are byte-identical to the legacy fixed advance; the slot
  estimate (`cardChromeEst`) parity is deferred to R10.10.
**Acceptance criteria:**
- A long header in a 1/3-width card advances the body top below the wrapped header
  bottom (no overlap) and sizes the header band to the wrapped height; a
  single-line header is byte-identical; deterministic.

#### Phase 40 — fit-to-region compression

**Subsystem:** scene — Layer 2 renderer (body-stack layout)
**RFC sections:** §10, §10.2
**Deps:** Phase 13 (`alignedStackIn`), Phase 23 (`VAlignFill`, D-052), Phase 39
(D-070), brief 23.
**What lands (R10.2, CRITICAL · engine):**
- A new opt-in `VAlignFit` body-stack mode (set via `SceneSlide.Content.Vertical`).
  When the stack overflows its region, a deterministic `fitCompress` pass floors
  the inter-node gap toward `SpaceXS` then proportionally scales slot heights
  toward a pinned `sMin=0.60` ratio, so the last node lands inside the frame
  instead of clipping off-slide.
- Byte-identical when off or when content already fits; the card-padding /
  type-scale sub-steps are deferred to R10.7 / R10.5, container-internal fit to
  R10.3 / R10.4.
**Acceptance criteria:**
- For a stack overflowing by up to ~25%, `VAlignFit` keeps the last node bottom
  ≤ region bottom using only the pinned steps; fitting content is byte-identical
  to `VAlignTop`; deterministic at any worker count; the overflow warning fires
  iff content still overflows after compression.

#### Phase 41 — content-weighted bento rows

**Subsystem:** scene — Layer 2 renderer (Bento container)
**RFC sections:** §11.2, §10
**Deps:** Phase 27 (Bento, D-056), Phase 40 (D-071), brief 24.
**What lands (R10.3, HIGH · engine):**
- A new opt-in `Bento.WeightedRows` flag. When set, each bento row sizes to its
  content's preferred height (the per-row max cell height at span widths) instead
  of an equal `(box.H − gaps)/nRows` band, clamped by a single basis-point scale
  so `Σ rows + gaps ≤ box.H` — a dense row no longer shares equal height with a
  sparse one.
- `bentoGeometry` refactored to factor out `bentoColumns`/`cellWidth` and return
  per-row Y/H; the equal-row default is byte-identical. Gutter labels anchor-
  middle within their actual row height. Grid analog + estimator parity deferred.
**Acceptance criteria:**
- A bento with a 1-line row and a 4-line row in weighted mode renders the dense
  row taller and fits the region; equal-mode and single-density bentos are
  byte-identical; deterministic at any worker count.

#### Phase 42 — card body vertical distribution

**Subsystem:** scene — Layer 2 renderer (Card)
**RFC sections:** §11.2, §10
**Deps:** Phase 13 (`alignedStackIn`), Phase 14 (Card), Phase 40 (D-071), brief 25.
**What lands (R10.4, HIGH · engine):**
- A new opt-in `Card.BodyVAlign VAlign`. The card's vertical body routes through
  the existing `alignedStackIn` (center / bottom / justify / fill / fit) on the
  card body box instead of the top-anchored `stackIn`, so secondary content can
  pin to the card bottom or fill the frame — no more dead space in tall cards.
- The zero value (`VAlignTop`) is byte-identical to today (the alignment engine
  already matches `stackIn` for the zero Alignment). Card only; CardSection
  deferred.
**Acceptance criteria:**
- `BodyVAlign=Bottom` pins the last body node's bottom to the card body bottom;
  `Justify` spreads inter-item gaps; `Top` is byte-identical; deterministic.

#### Phase 43 — display text shrink-to-fit

**Subsystem:** scene — Layer 2 renderer (+ a `pptx` run override)
**RFC sections:** §8.4, §10.2
**Deps:** Phase 22 (`naturalWidth`), Phase 28 (Stat), Phase 40 (D-071), brief 26.
**What lands (R10.5, HIGH · engine):**
- An opt-in `AutoFit bool` on the display nodes (`Hero`, `Stat`, `Heading`) and a
  new per-run `RunStyle.FontScale` multiplier. When a display run's estimated
  `naturalWidth` exceeds its box, a pure `fitScale` quantizes `boxW/naturalWidth`
  down to a 0.025 step (floored at 0.60) and emits the reduced `a:rPr/@sz`, so a
  too-wide title or price fits one line instead of wrapping.
- Never upscales; `FontScale=0` / `AutoFit=false` / already-fitting text are
  byte-identical; the scale keeps the role size token as source of truth (P2) and
  round-trips via `Run.FontSize()`.
**Acceptance criteria:**
- An over-wide Stat/price value fits its column at a font ≥ ratioMin×base;
  fitting text and AutoFit-off are byte-identical; a scaled run round-trips its
  size; deterministic.

#### Phase 44 — fill cap (no over-grow)

**Subsystem:** scene — Layer 2 renderer (body-stack fill)
**RFC sections:** §10
**Deps:** Phase 23 (`VAlignFill`, D-052), Phase 40 (D-071), brief 27.
**What lands (R10.6, HIGH · engine):**
- A new opt-in `VAlignFillCapped` body-stack mode. Like `VAlignFill` but each
  flexible node grows by at most a pinned factor of its preferred height
  (`fillGrowthMaxBP = 10000` → ≤ 1× added), so a sparse node can't balloon; the
  leftover slack becomes balanced spacing (even top margin + widened inter-node
  gaps, `residual/(n+1)`).
- Uncapped `VAlignFill` keeps calling the unchanged `distributeFill` — byte-
  identical.
**Acceptance criteria:**
- A sparse+dense capped-fill stack grows the sparse node ≤ its cap and shows the
  residual as even spacing (not one ballooned node); uncapped fill is byte-
  identical; deterministic and within the box.

#### Phase 45 — density-aware card padding

**Subsystem:** scene — Layer 2 renderer (Card)
**RFC sections:** §11.2, §7
**Deps:** Phase 14 (Card), Phase 39 (D-070), brief 28.
**What lands (R10.7, MED · engine):**
- An additive `Card.PaddingScale int` — a basis-point multiplier on the
  size-resolved interior padding (0/10000 = unchanged), floored at a pinned
  `SpaceXS` `padMin`. A tighter scale shrinks the inset and grows the card body so
  a dense card reclaims interior space; resolves through theme spacing tokens (no
  literals). The three padding sites route through a new `cardPaddingFor`.
- Zero/default reproduces the SM/MD/LG output byte-for-byte; auto-tighten-in-fit
  is deferred.
**Acceptance criteria:**
- A tighter scale measurably reduces the inset and grows the body (≥ padMin); an
  extreme scale floors at padMin; default is byte-identical; deterministic.

#### Phase 46 — balanced vertical rhythm

**Subsystem:** scene — Layer 2 renderer (body-stack alignment)
**RFC sections:** §10
**Deps:** Phase 13 (alignment), Phase 44 (D-075), brief 29.
**What lands (R10.8, MED · engine):**
- A new opt-in `VAlignBalanced` body-stack mode. It distributes a sparse stack's
  slack as an even rhythm — `unit = slack/(n+1)` into a top margin and widened
  inter-node gaps — with an optical-center upward bias (top margin = 85% of an
  even unit), so a sparse cover/closing reads balanced instead of clustered with a
  large void. Distinct from `VAlignJustify` (no margins) and `VAlignCenter` (fixed
  gaps).
- `VAlignTop`/`VAlignCenter`/`VAlignJustify` are untouched; per-node gap weighting
  stays the caller's (D-026).
**Acceptance criteria:**
- A 3-node sparse stack under balanced mode has a non-zero top margin + widened
  gaps that distribute the slack (no single void), sits above geometric center,
  and stays in the box; Top/Center byte-identical; deterministic.

#### Phase 47 — list bullet indent density

**Subsystem:** scene — Layer 2 renderer (+ a `pptx` paragraph option)
**RFC sections:** §8.4, §11.1
**Deps:** Phase 03 (text builder), Phase 11 (List), brief 30.
**What lands (R10.9, MED · engine):**
- A new `pptx.ParagraphOpts.BulletIndent` (the bullet hanging indent; 0 = the
  default 0.5") and a scene `List.Indent` preset (`IndentNormal`/`IndentTight`).
  `IndentTight` tightens the marker-to-text offset to `In(0.25)` so lists read
  dense instead of loose, consistently across items and levels.
- Pinned presets (no token); `IndentNormal`/`BulletIndent=0` byte-identical; the
  emitted `marL`/`indent` round-trip.
**Acceptance criteria:**
- A tight list shows a smaller, consistent marker-to-text offset; the default is
  byte-identical; the emitted indent round-trips; deterministic.

#### Phase 48 — estimate/actual parity

**Subsystem:** scene — Layer 2 renderer (slot estimators)
**RFC sections:** §10.2
**Deps:** Phase 22 (`preferredHeight`), Phase 39 (D-070), Phase 41 (D-072), brief 31.
**What lands (R10.10, HIGH · engine):**
- The Card/CardSection `preferredHeight` becomes wrapped-header-aware
  (`cardChromeEst` + the extra eyebrow/title lines at the header column width), and
  the Bento estimate measures each cell at its actual span width instead of the
  unit width — so the overflow warning and the fit pass operate on accurate
  numbers. Closes the `cardChromeEst` parity deferred by R10.1.
- Single-line headers (increment 0) and span-1 bento cells (span width = unit
  width) are byte-identical; the card-header helpers are refactored to theme-taking
  free functions with method wrappers. Card body inset parity deferred.
**Acceptance criteria:**
- A multi-line-header card's estimate grows by the wrapped increment; a wide-span
  bento cell's estimate ≤ the unit-width one; single-line/span-1 byte-identical;
  the overflow warning fires iff the composed content exceeds the region;
  deterministic.

---

### Wave 11 — Rendering robustness

The R11 (`DECKARD-PRODUCT-REQUIREMENTS.md`) engine units: every component renders
correctly under ANY content — wrapped headers grow their band, all card/container
chrome text is variant-aware (auto-contrast), and pills / dots / watermarks /
badges never collide or overflow the slide safe area.

#### Phase 49 — card header content-aware height (verify-and-close)

**Subsystem:** scene — Layer 2 renderer (card chrome / layout)
**RFC sections:** §10.1, §12.1
**Deps:** Phase 39 (R10.1 / D-070), Phase 48 (R10.10 / D-079), brief 32.
**What lands (R11.1, CRITICAL · engine):**
- A verify-and-close: the wrapped-aware card-header geometry
  (`cardHeaderColumnWOf` / `cardHeaderRowHeights`, shared by `cardHeaderBottom` and
  `renderCardChrome`) already shipped in R10.1/D-070 with estimator parity in
  R10.10/D-079, so R11.1 needs only its named acceptance golden, not a
  reimplementation (D-081).
- A combinatorial acceptance test: a deliberately long multi-line header swept
  across all `CardSize × CardLayout` combos asserting `body.Y >= header band
  bottom` and band containment; single-line headers byte-identical. No renderer
  change.
**Acceptance criteria:**
- For a long header in a narrow card, across `{MD, SM, LG} × {Default, IconTop}`,
  body top ≥ `cardHeaderBottom`; the `HeaderFill` band bottom equals
  `cardHeaderBottom`; the header wraps to ≥ 2 lines; single-line `titleH ==
  cardTitleRowH`; deterministic.

#### Phase 50 — card-text auto-contrast

**Subsystem:** scene — Layer 2 renderer (card / container chrome)
**RFC sections:** §7.1, §7.4, §12.1
**Deps:** Phase 25 (D-054), Phase 29 (D-058), brief 33.
**What lands (R11.2, CRITICAL · engine):**
- A deterministic auto-contrast mechanism `onCardSurface(bg)`: pinned sRGB relative
  luminance (a 256-entry integer table built once at init) returns a light text
  token on a dark surface and nil (the dark default) on a light one. Wired into the
  card header / eyebrow / pill, the TwoColumn join-badge label, and the Stat value,
  so chrome text is legible on any fill or slide variant — fixing black-on-dark
  headers and same-hue eyebrows.
- The eyebrow keeps its accent tint only when it clears 4.5:1 against its surface.
  Light-surface cards are byte-identical (nil → the inherited default); the default
  accent on a white card clears the check, so the common eyebrow is unchanged. A
  mechanism the caller overrides via an explicit `Color` (D-026); reconciles the
  D-058 no-contrast-logic tension (D-082).
**Acceptance criteria:**
- Chrome text clears ≥ 4.5:1 against any light/dark/brand surface; light cards are
  byte-identical (no explicit header color); dark fills/variants flip the header to
  a light color; the eyebrow drops a same-hue accent; deterministic across worker
  counts.

---

## 4. Post-V1 backlog

See `RFC-001-pptx-go.md §24` for the full backlog. Headline items: native
`c:chart`, third-party PPTX read fidelity, animations/transitions,
SmartArt-equivalents.
