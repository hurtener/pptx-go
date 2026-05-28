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
a phase plan lands; gaps are allowed (a planned-then-cancelled phase
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
suite; it spot-checks user-visible behaviour (the binary builds, the
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
```

Each wave ends with a checkpoint audit (`CLAUDE.md §17`). V1.0.0 ships at
the end of Wave 7.

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
  name with no behaviour change).
- `make test` passes (upstream tests still green).
- `make preflight` passes.
- `make check-mirror` passes (`AGENTS.md == CLAUDE.md`).
- `.github/workflows/ci.yml` green on a sample PR.

---

### Wave 1 — Theme + Builder spine

#### Phase 01 — OPC + OOXML reorg

**Subsystem:** internal/opc, internal/ooxml
**RFC sections:** §6
**Deps:** Phase 00.
**What lands:**
- Move upstream `opc/` → `internal/opc/` (rename only; no behaviour change).
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
  different OOXML colours.
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
  options. Product behaviour lives in callers (D-026).
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
- `pptx.LoadTheme(path)` — extract Theme from a `.pptx` template.
- `pptx.FromTemplate(brand)` — option for `pptx.New` to seed
  presentation from a template (masters + theme + default layouts
  copied).
- `pptx/master.go` — `Master`, `Layout`, `LayoutMap`.
- `scene.WithTheme`, `scene.WithLayoutMap` — render options.
**Acceptance criteria:**
- Loading a PowerPoint-emitted template's theme produces a `Theme`
  whose `Resolve(ColorAccent)` returns the template's accent.
- A scene rendered with `scene.WithTheme(brandTheme)` uses the brand's
  colours.

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
- Foreground vs background layer ordering is honoured.

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
**What lands:**
- Parsers gracefully handle unrecognized OOXML elements (preserved as
  opaque `RawShape` / `RawPart` carriers).
- Documented degradation modes when external-deck features don't map
  to the builder model.
**Acceptance criteria:**
- A library of PowerPoint-authored sample decks loads without panic.
- Unsupported elements surface in a `ReadWarnings` slice.

---

### Wave 7 — Docs, skills, release

#### Phase 20 — Agent skills + published docs site

**Subsystem:** docs/site, skills
**RFC sections:** `CLAUDE.md §19`
**Deps:** Wave 6 complete.
**What lands:**
- `skills/` — initial SKILL.md set: "scaffold a presentation", "load a
  brand theme", "compose a scene", "extend the icon set",
  "rasterize and embed a code block", "rasterize and embed a chart".
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

---

## 4. Post-V1 backlog

See `RFC-001-pptx-go.md §24` for the full backlog. Headline items: native
`c:chart`, third-party PPTX read fidelity, animations/transitions,
SmartArt-equivalents.
