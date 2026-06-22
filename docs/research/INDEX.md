# Research briefs — index

> Subsystem → research-brief reverse index. A phase plan that cites no
> brief in its **Brief findings incorporated** section is a drift signal
> (`CLAUDE.md §16`).
>
> A research brief is a `docs/research/NN-slug.md` file authored *before*
> a phase plan to investigate prior art, capture domain knowledge, and
> tee up the design decisions that will land in the RFC or the phase
> plan. Briefs are **context, not decisions** — they inform; the RFC and
> phase plans decide.

---

## Format

Each entry below is keyed by **subsystem** (matching `RFC §3.3`) and
lists the briefs that inform work in that subsystem.

```text
### <subsystem>
- `NN-slug.md` — one-line summary
- `NN-slug.md` — one-line summary
```

A brief may cross-cut multiple subsystems; list it under each.

---

## Subsystems

### internal/opc — OPC package layer

*(no briefs yet — Phase 01 may add one investigating OOXML transitional
vs strict profile edge cases at the OPC layer)*

### internal/ooxml — OOXML codec layer

- `01-master-layout-theme-ingestion.md` — theme1.xml color/font scheme +
  master/layout inheritance, and the read paths template ingestion depends on.

*(candidates: chart wire-format survey for V2, table XML shape, theme XML
compatibility across PowerPoint versions)*

### pptx — Layer 1 builder

- `01-master-layout-theme-ingestion.md` — `LoadTheme`/`FromTemplate` strategy:
  copy template parts wholesale, extract the `Theme`, map `LayoutKind` to
  named layouts.
- `05-card-chrome-and-shadow-primitive.md` — the `outerShdw` builder shadow
  primitive (`WithShadow` / `WithElevation`) elevation needs, mirroring the
  D-041 gradient build.
- `07-chart-image-shape.md` — the `pptx.ChartPlaceholder` slot helper and the
  §7-boundary clarification that reading image dimension headers
  (`image.DecodeConfig`) is permitted (decoding pixel data is not), enabling the
  chart aspect-ratio warning.
- `08-roundtrip-read.md` — the read-side public model (`Slide.Shapes()` + read
  accessors on the builder types) needed to reconstruct an opened deck per
  RFC §16, backed by pure `X*`→public mappings over the already-parsed tree.
- `18-font-embedding-pipeline.md` — an opt-in save-time pass (`WithFontEmbedding`)
  that walks every slide's runs, collects the distinct `(family, bold, italic)`
  faces in stable sorted order, and `EmbedFont`s each via the registered
  `FontSource`; warn-don't-fail on a missing face, idempotent vs manual
  `EmbedFont`, byte-identical when off / no source. (R9.1, engine half.)

*(candidates: rich-text auto-fit modes in OOXML practice, table merged-cell
semantics)*

### scene — Layer 2 renderer

- `05-card-chrome-and-shadow-primitive.md` — `card` / `card_section` as native
  chrome (rounded rect + accent stripe + icon/eyebrow/header/pill + body);
  the builder shadow primitive elevation needs (`outerShdw`), wiring the icon
  registry into compose (closing the Phase-12 deferral), and the additive
  Card IR expansion.
- `06-flow-step-pipeline.md` — `flow` as native step pills + connector glyphs
  (`arrow` / `arrow_dashed` / `cycle` / `plus`) composing preset shapes (no new
  builder API — the RFC's unbuilt `AddConnector` is not needed); flow-level
  connector kind, the `arrow_dashed` geometry wrinkle, a lighter dedicated
  step-pill, and the additive Flow IR (`Connector`, optional step `Icon`).
- `07-chart-image-shape.md` — `chart` as the code_block raster path minus the
  badge plus an aspect-ratio `LayoutWarning`; aspect detection options
  (`image.DecodeConfig` header read vs caller field) and the `ChartPlaceholder`
  slot helper.
- `09-text-height-metrics.md` — deterministic wrapped-line-count estimation
  (`ceil(naturalWidth / availableWidth)`) feeding a content-aware
  `preferredHeight`, so stacked nodes stop overlapping and overflow is reported
  truthfully, while single-line content stays byte-identical.
- `10-grow-to-fit.md` — `VAlignFill`: after fixed leaves take preferred height,
  distribute the leftover body height to the flexible nodes (containers +
  Image/Chart) so they grow to consume the frame; the geometry engine already
  honors the taller box, so no container renderer changes.
- `11-slide-chrome.md` — opt-in per-slide chrome (top section eyebrow + hairline,
  bottom brand slot + `N / total` page number) drawn outside a shrunk body
  region; driven by `Scene.Chrome` + `SceneSlide` fields, native shapes reusing
  existing tokens, chrome-off byte-identical.
- `12-rich-card-visuals.md` — additive `Card` visuals: a colored header band
  (`HeaderFill`), a top-right status dot (`StatusDot`), and a ghosted watermark
  label (`Watermark`); optional colors are `*ColorRole` (nil = omit), watermark
  is true low-opacity text via `TokenColorAlpha`, byte-identical when unset.
- `13-column-join.md` — a centered inter-column element on `TwoColumn`: a
  `ColumnJoin` enum (`JoinNone`/`JoinBadge`/`JoinArrow`) + `JoinLabel` draws a
  "VS"-style badge or a connector arrow on the column seam; byte-identical when
  `JoinNone`. (R5 sub-units a+b; the bento grid (c) is a separate phase.)
- `14-bento-grid.md` — a new `Bento` IR node: rows with an optional left label
  and cells of variable column span against a shared column grid (absolute spans
  keep columns aligned); the new-node wiring checklist + a `cellNodes()` helper.
  (R5 sub-unit c, completing R5.)
- `15-stat-node.md` — a new `Stat` leaf node (display-scale value + label +
  optional toned delta) reusing the `Hero` text idiom and existing
  success/error/muted tokens; a `Grid` of `Stat`s is a metric/pricing strip. The
  new-leaf wiring checklist (no `walk*`/`isFlexible`). (R6.)
- `16-resolved-colors.md` — expose per-slide resolved canvas/surface/primary-text
  colors in `Stats.Colors`, captured from the per-slide theme after compose (the
  derived dark palette for `VariantDark`), so a caller computes its own contrast;
  no contrast logic in the engine. (R7, final Wave 8 unit.)

*(candidates: layout-engine survey (CSS grid analogues expressible in EMU),
scene IR JSON wire form compatibility with pengui-slides v4)*

### Theme & tokens

- `17-type-detail-tokens.md` — per-role letter-spacing (tracking) on `FontSpec`
  (+ a `RunStyle` override) emitted as OOXML `a:rPr/@spc`, round-trip clean and
  byte-identical when zero; the run-attribute starting point for the R9 type-detail
  tokens (line-height, case follow). (R9.3.)
- `19-font-fallback-stack.md` — a per-role `FontSpec.Fallback []string` realized
  at write time: a registered `FontSource` is the availability oracle, and a run's
  single-valued `a:latin` is rewritten to the first resolvable family in
  `[Family] + Fallback`; byte-identical when unused, deterministic, idempotent.
  (R9.6, engine half.)
- `20-emphasis-italic-display.md` — the display-italic guarantee is already
  delivered (D-063 family + D-065 embeds the italic cut); the incremental engine
  work is italic-aware fallback (resolve per `(family, italic)`, so an italic run
  lacking an italic cut falls back rather than faux-italicizing a sans) + the
  `<p:font>` embedded-list prefix fix. (R9.7, engine half.)
- `21-weight-aware-embedding.md` — track the resolved numeric weight per run
  (`XTextProperties.Weight`, `xml:"-"`) so the embedding pass ships the actual
  weight file per OOXML bucket (nearest-nominal when weights collide) instead of a
  synthetic 400/700; embeds one file per bucket (the 4-cut limit). (R9.8, engine.)
- `01-master-layout-theme-ingestion.md` — how a brand kit's color scheme,
  `clrMap` indirection, and font scheme map onto pptx-go's token roles.

*(candidates: token taxonomy comparison with design systems (Tailwind, Radix,
Material))*

### Curated assets (icons, ornaments, frames)

- `02-device-frame-shape-geometry.md` — drawing the four V1 device frames
  (browser/phone/desktop/laptop) as native rounded-rect + ellipse shapes;
  the recipe→interior seam, token-only bezel color, and the enum-vs-named
  frame-reference reconciliation for §14.4 extension.
- `03-svg-path-to-ooxml-translator.md` — rendering curated icons as native
  `custGeom` path shapes: the supported SVG `d` subset (M/L/H/V/C/Q/Z, no
  arcs), the viewBox→path coordinate mapping, registration-time constraint
  validation, and why the set is lucide-*style* (filled single paths), not
  lucide's stroke-based data.
- `04-preset-ornament-recipes.md` — the six V1 ornaments as native shape
  recipes; the builder primitives they need first (gradient fills for glows,
  `WithRotation`, token-alpha), the Decoration IR expansion
  (offset/bleed/opacity/rotation/size), and the layer z-order.
- `05-card-chrome-and-shadow-primitive.md` — wiring the curated icon registry
  into compose (the first node to *place* an icon: card), closing the Phase-12
  consumption deferral with a `validateIconRefs` closed-name Stage-1 check.

### Charts

*(no briefs yet — V2 will warrant briefs on `c:chart` XML survey by
chart type and PowerPoint Online vs Desktop divergences)*

### Streaming & performance

*(no briefs yet — candidates: concurrent rendering scaling on M-class
Apple Silicon vs x86_64, zip-streaming costs vs in-memory)*

### Read & round-trip

- `08-roundtrip-read.md` — what `pptx.Open` reconstructs today (hybrid: high-
  level structure rebuilt, slide shapes preserved as opaque OOXML) vs the RFC
  §16 mandate (reconstruct the navigable Go model — `Slides()[0].Shapes()[0]`);
  the extend-the-builder-types read model and the split-by-primitive delivery.

*(candidates: PowerPoint output variance (PowerPoint vs PowerPoint Online vs
Office for Mac), Keynote-to-PPTX export quirks)*

---

## Authoring a brief

A research brief is a Markdown file under `docs/research/` named
`NN-slug.md` where `NN` is the next available two-digit number. Brief
structure:

```markdown
# Brief NN — <slug>

**Subsystem:** <RFC §3.3 subsystem>
**Authored:** <YYYY-MM-DD>
**Motivating phase:** <Phase NN — slug> (or "RFC-level investigation")

## 1. Question
What this brief investigates.

## 2. Prior art surveyed
Specs, libraries, papers, decks consulted.

## 3. Findings
What we learned. Bullet-point. Each finding is something a phase plan
can incorporate or reject.

## 4. Recommendations
Suggested directions for the motivating phase. Recommendations are
inputs to the phase plan; the plan decides.

## 5. Open questions
What we *didn't* answer, with a note about which phase / RFC change
should pick it up.
```

Briefs are authored before the phase plan, listed in this INDEX under
the relevant subsystem, and cited by the phase plan's **Brief findings
incorporated** section.

A brief is not a phase plan. A brief makes recommendations; the phase
plan binds. When a phase plan departs from a brief's recommendation, the
**Findings I'm departing from** section names the brief and the
rationale.

---

*Add new briefs to the subsystem section above and to a chronological
list below (V1.x — for now the index above is the canonical view).*
