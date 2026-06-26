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
- `22-card-header-height.md` — make the card header row wrapped-title-aware: a
  shared `cardHeaderColumnW`/`cardHeaderRowHeights` sizes the eyebrow/title (and
  the D-054 band + body Y) to `wrappedLines × per-row` so a long header no longer
  overlaps the body; single-line byte-identical; estimator parity deferred to
  R10.10. (R10.1, CRITICAL.)
- `23-fit-to-region-compression.md` — opt-in `VAlignFit`: when the body stack
  overflows its region, a deterministic `fitCompress` pass floors the inter-node
  gap toward `SpaceXS` then proportionally scales slot heights toward a pinned
  `sMin=0.60` ratio, so the last node lands inside the frame instead of clipping
  off-slide; byte-identical when off or when content fits; the card-padding /
  type-scale sub-steps are deferred to R10.7 / R10.5. (R10.2, CRITICAL.)
- `24-content-weighted-bento-rows.md` — opt-in `Bento.WeightedRows`: size each
  bento row to its content's preferred height (per-row max cell height at span
  widths), clamped by a single basis-point scale so `Σ rows + gaps ≤ box.H`, so a
  dense row no longer shares equal height with a sparse one; default equal rows
  byte-identical; the `bentoGeometry` refactor factors out `bentoColumns`/
  `cellWidth` and returns per-row Y/H. Grid analog + estimator parity deferred.
  (R10.3, HIGH.)
- `25-card-body-vertical-distribution.md` — opt-in `Card.BodyVAlign VAlign`: route
  the card's vertical body through the existing `alignedStackIn` (center/bottom/
  justify/fill/fit) on the card body box instead of top-anchored `stackIn`, so
  secondary content can pin to the bottom or fill the card; zero value (Top) is
  byte-identical (the alignment engine already matches `stackIn` for the zero
  Alignment). Card only; CardSection deferred. (R10.4, HIGH.)
- `26-display-text-shrink-to-fit.md` — opt-in `AutoFit` on the display nodes
  (`Hero`/`Stat`/`Heading`): a pure `fitScale(natW, boxW)` quantizes `boxW/natW`
  down to a 0.025 step, floored at a 0.60 ratio, and a new per-run
  `RunStyle.FontScale` multiplier emits the reduced `@sz` so a too-wide title or
  price fits one line; never upscales; default OFF byte-identical; the scale keeps
  the role size token as source of truth (P2). (R10.5, HIGH.)
- `27-fill-cap-no-overgrow.md` — opt-in `VAlignFillCapped`: a bounded
  `distributeFill` that grows each flexible node by at most `growthMax×preferredH`
  so a sparse node can't balloon; the residual slack becomes balanced spacing
  (even top margin + widened inter-node gaps, `residual/(n+1)`), reusing the
  Justify/Fit offset-and-gap mechanism; uncapped `VAlignFill` byte-identical.
  (R10.6, HIGH.)
- `28-density-aware-card-padding.md` — additive `Card.PaddingScale` (basis-point
  multiplier on the size-resolved card padding, default unchanged) floored at a
  pinned `SpaceXS` `padMin`; a tighter scale shrinks the interior inset and grows
  the body, token-resolved (P2, no literals); zero/default byte-identical;
  auto-tighten-in-fit deferred. (R10.7, MED.)
- `29-balanced-vertical-rhythm.md` — opt-in `VAlignBalanced`: distribute a sparse
  stack's slack into an even rhythm — `unit = slack/(n+1)` into a top margin and
  widened inter-node gaps — with an optical-center upward bias (`top = unit ×
  0.85`), so a sparse cover/closing reads balanced instead of clustered-plus-void;
  distinct from Justify (no margins) and Center (fixed gaps); `VAlignTop`/`Center`
  byte-identical; per-node gap weighting stays the caller's. (R10.8, MED.)
- `30-list-bullet-indent-density.md` — a per-paragraph `ParagraphOpts.BulletIndent`
  override (builder) + a `List.Indent` preset (`IndentNormal`/`IndentTight`)
  plumbed through `renderList`: `IndentTight` tightens the bullet hanging indent to
  `In(0.25)` (vs the `0.5"` default) so lists read dense instead of loose; pinned
  presets (no token), default byte-identical, emitted `marL`/`indent` round-trip.
  (R10.9, MED.)
- `31-estimate-actual-parity.md` — align the slot estimators with the composers:
  the Card/CardSection `preferredHeight` adds a wrapped-header increment
  (`cardChromeEst` + extra eyebrow/title lines, baseline kept so single-line is
  byte-identical) and the Bento estimate measures each cell at its actual span
  width (span-1 byte-identical, wide-span no longer over-counts), so overflow
  detection / the fit pass are trustworthy; closes the R10.1-deferred
  `cardChromeEst` parity. (R10.10, HIGH.)
- `32-card-header-robustness-verify.md` — verify-and-close of R11.1: the
  wrapped-aware card header geometry (`cardHeaderColumnWOf`/`cardHeaderRowHeights`
  shared by `cardHeaderBottom` + `renderCardChrome`) already shipped in R10.1/D-070
  (estimator parity in R10.10/D-079), so R11.1 needs only the acceptance golden — a
  long multi-line header swept across all `CardSize × CardLayout` combos asserting
  `body.Y >= header band bottom` and band containment, single-line byte-identical.
  No renderer change. (R11.1, CRITICAL.)
- `33-card-text-auto-contrast.md` — a deterministic auto-contrast mechanism
  (`onCardSurface(bg)`): pinned sRGB relative luminance (a 256-entry integer table
  built once at init) picks a light text token on a dark surface and nil (the dark
  default, byte-identical) on a light one; the eyebrow keeps its accent only when it
  clears 4.5:1, else falls back. Wired into card header/eyebrow/pill, the join-badge
  label, and the Stat value, so chrome text is legible on any fill/variant. Resolves
  the D-058 tension under D-026 (a mechanism the caller overrides, not a policy).
  (R11.2, CRITICAL.)
- `34-container-slide-bounds-clamp.md` — a `clampToSafeArea(box)` guard at the
  `renderBento`/`renderGrid`/`renderCard` entry caps a container's box so its bottom
  never exceeds the slide safe area (`= bodyRegion()`: slide − margins − chrome
  bands), warning once on clamp. Fires only on strict overflow → fitting layouts
  (and `VAlignFill` to the region bottom) are byte-identical; a pure integer cap →
  deterministic. Defense-in-depth complementary to the opt-in `VAlignFit` (which
  reflows; this caps). (R11.3, CRITICAL.)
- `35-content-region-reserves-chrome.md` — verify-and-close of R11.4: the chrome-
  aware `bodyRegion()` already reserves the eyebrow + footer bands (D-053) and the
  body stack lays out inside it; the overflow hole that defeated it is closed by the
  R11.3 clamp (D-083). The close is the acceptance — the chrome-on body region is
  disjoint from both bands (recomputed from the chrome constants), chrome-off is the
  plain margin box (byte-identical), and a clamped container stays above the footer.
  No renderer change. (R11.4, HIGH.)
- `36-header-pill-fit-to-label.md` — a shared `cardPillWidthOf(theme, pill, innerW)`
  = clamp(`naturalWidth(pill@TypeCaption)` + 2·padX, min, innerW) sizes a card header
  pill to its label, called from both `cardHeaderColumnWOf` (the reservation) and
  `renderCardChrome` (the drawn pill) so they never drift; a clamped label is
  shrunk to one line via the R10.5 `fitScale`/`FontScale`. Fixes a long pill
  ("CUSTOMIZABLE") wrapping inside a fixed `In(1.0)` chip. Pure integer →
  deterministic. (R11.5, HIGH.)
- `37-chrome-element-anti-collision.md` — when a card has both a header pill and a
  status dot (both top-right anchored), place the dot left of the pill (`dotX =
  pillX − gap − dotSz`, floored at `innerX`) so their boxes are disjoint; inert /
  byte-identical when only one is set. Disjointness by construction (shares
  `cardPillWidthOf`). (R11.6, HIGH.)
- `38-join-badge-fit-to-label.md` — the TwoColumn join badge grows to contain its
  label: `badgeSz = clamp(naturalWidth(label@TypeBodySmall) + 2·padX, In(0.62),
  In(1.5))`, then `fitScale`/`FontScale` shrinks a label that still overflows the
  cap. Fixes "One agent" breaking mid-word in the fixed `In(0.62)` ellipse; "vs"
  keeps the base (byte-identical). (R11.7, HIGH.)
- `39-stat-value-overflow-guard.md` — a pinned role ladder for the Stat value
  (`statValueFit`): walk `[TypeDisplay, TypeH1, TypeH2]` and pick the first that fits
  one line, else the floor + a `fitScale` shrink — so a wide value ("$4,000+") never
  wraps and crowds the caption. Gated on `AutoFit` (D-074 opt-in) → AutoFit-off and
  AutoFit-on-fitting are byte-identical; the existing AutoFit tests stay green.
  Stack-height clamp deferred (needs `slideID`). (R11.8, HIGH.)
- `40-bento-rowlabel-gutter-fit.md` — a shared `bentoGutterWidthOf(theme, v)` =
  clamp(max `naturalWidth(label@TypeCaption)` + 2·padX, In(0.8), In(1.6)) sizes the
  bento row-label gutter to its widest label, used by both `bentoColumns` (layout)
  and the `preferredHeight` Bento estimate (parity). Fixes "Control plane" wrapping
  in the fixed `In(1.2)` gutter; `theme` threaded into the bento geometry. (R11.9,
  MED.)
- `41-list-bullet-hanging-indent.md` — make the `IndentTight` bullet hanging indent
  proportional: `listTightIndent() = listTightIndentBase × bodySize / 14`, anchored
  byte-identical to `In(0.25)` at the default 14 pt and scaling with the body size.
  The mechanism (`BulletIndent`/`IndentTight`) is from R10.9/D-078; the list start Y
  respects the grown card header via R10.1. (R11.10, MED.)
- `42-decoration-watermark-anticollision.md` — verify-and-close of R11.11: the card
  watermark is already emitted before the body (z-order behind) at ~13% opacity, and
  background decorations are z-order-behind too (D-054). R11.11's acceptance is an OR
  (residual region OR behind-at-low-alpha); the engine takes the second branch, so
  the close is the acceptance test (z-order-behind, low-alpha, inert-when-unset), not
  the optional residual-region restriction. (R11.11, LOW.)
- `43-adversarial-content-fit-fixtures.md` — a reusable torture harness rendering
  every component × {light, dark} under hostile content; invariant (2) parses every
  `<a:off>`/`<a:ext>` and asserts on-canvas (no recorder needed), (1)/(3)/(4) via the
  fit/contrast helpers. The suite surfaced an off-canvas card-body-leaf overflow,
  fixed by generalizing the R11.3 safe-area clamp to `renderNode` (exempting
  full-slide overlays; subsumes the three container clamps). (R11.12, HIGH.)
- `44-prim-cta-button.md` — a new `Button` leaf IR node (the first R12 primitive):
  content-fit `RadiusFull` pill, `ButtonTone` → `ColorRole` fill (ghost = `NoFill` +
  accent hairline), pinned `ButtonSize` height/padding scale, middle-anchored bold
  `TypeBody` label flanked by native custGeom leading/trailing icons, `fitScale`
  tail, `Align` offset within the box. Presentational only (no hyperlink — static
  deck). Reuses the header-pill geometry, the icon registry + `walkIconRefs`
  validation, and `onCardSurface` auto-contrast. Additive ⇒ byte-identical when
  unused. (R12.1, CRITICAL.)
- `45-prim-in-card-checklist-fill.md` — a new `Checklist` leaf IR node: true filled
  status glyphs (`check`/`x`/`dot` curated custGeom via `AddIcon`, not a font
  checkbox — fixing the empty-square bug), a hanging indent from the glyph width,
  row-major reflow into 1–3 columns, per-state glyph color with an optional
  `*ColorRole` `GlyphTone` override (D-054 pattern), and a `Fill` mode that
  distributes inter-row slack so a short list spans the card (added to `isFlexible`).
  Reuses the icon registry + `walkIconRefs`, `wrappedLines`, and the VAlignJustify
  slack math. Additive ⇒ byte-identical when unused. (R12.2, CRITICAL.)
- `46-prim-chip-row-group.md` — a new `ChipRow` leaf IR node: a greedy left-to-right
  wrap of content-fit chip pills (reusing the single-`Chip` pill + the button
  content-fit width), an optional leading `TypeCaption` label on line 0, per-line
  `HAlign` offset, and a shared two-pass packer (`chipRowLines`) feeding both the
  renderer and `preferredHeight`. `Wrap` is the engine mechanism (zero = one line;
  the product sets it true). Pinned metrics, `ChipTone` token colors. Additive ⇒
  byte-identical when unused. (R12.5, HIGH.)
- `47-prim-callout-banner.md` — a new `Banner` IR node (with `Trailing []SlideNode`
  children): a full-width `RadiusLG` filled strip (fill defaults to accent), a leading
  icon + bold lead + body with auto-contrast text, and an optional right region
  stacking the trailing `Stat`/`Button`. Recurses like a container in every walk
  (validate / walkIconRefs / walkImages / nodeUsesAssets / collectKinds). Distinct
  from the side-bar `Callout`. Additive ⇒ byte-identical when unused. (R12.6, HIGH.)
- `48-prim-ribbon-corner-badge.md` — a `Card.Ribbon *Ribbon` field extension (not a new
  node): a pinned emphasis badge outside the header text flow. `RibbonTopBar` reserves
  a band (`ribbonReserveOf` threaded through `cardHeaderBottom` / `renderCardChrome` /
  `cardHeaderExtraHeight` so the body shifts down); `RibbonCornerStar` is the curated
  star glyph; `RibbonCornerTL/TR` are content-fit corner text tabs (the diagonal
  rotated-band variant deferred — no builder text rotation). nil ⇒ byte-identical.
  (R12.3, HIGH.)
- `49-prim-inter-column-connectors.md` — a `Grid.Connectors []GridConnector{Between
  [2]int; Kind; Label}` field extension: a connector glyph drawn in the gutter between
  two adjacent columns (derived from the cell boxes), reusing `render_flow`'s
  `renderConnector` + a new `ConnectorBiArrow` (`leftRightArrow`/`upDownArrow`).
  Adjacency validated at Stage-1; empty ⇒ byte-identical. (R12.4, HIGH.)
- `50-prim-icon-label-rows.md` — a new `IconRows` leaf IR node: a vertical stack of
  `[icon | label | optional right-aligned meta]` rows with an optional `RowPill`
  `SurfaceAlt` frame and a `Fill` mode (added to `isFlexible`). Mirrors the Checklist
  row engine (content-aware heights, slack distribution, `walkIconRefs` per-row icon);
  `GlyphColor` defaults to accent. Additive ⇒ byte-identical. (R12.7, MED.)
- `51-prim-spanning-column-bridge.md` — a `TwoColumn.JoinPosition` field extension
  (`JoinSeam`/`JoinTopBridge`/`JoinBottomBridge`): a horizontal accent bracket (spanning
  line + two end stubs + a content-fit centered label pill, no mid-word wrap) across the
  top/bottom of both columns, reserving a band (ribbon-style) so it spans above/below the
  content. `JoinSeam` (zero) is byte-identical to the D-055 seam element. (R12.8, MED.)
- `52-prim-attribution-lockup.md` — a new `Lockup` leaf IR node (asset-bearing): a
  caption + a small partner logo composed as one centered inline group. The mark is an
  `AssetID` (resolved via the AssetResolver → a pic, `nodeUsesAssets` true → serial part
  numbering) OR an `Icon` (media-free); exactly one is set. `AssetSide` orders the pair;
  `MaxHeight` bounds the (square — §7) logo. `policy {HasAsset:true}` like Decoration.
  Additive ⇒ byte-identical. (R12.9, LOW.)

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
- `53-tinted-paper-canvas.md` — a `ColorPaper` surface role (appended to the
  `ColorRole` iota, default `FFFFFF` = canvas so byte-identical) for a faintly
  tinted off-white "paper" canvas; a role *without* a theme1.xml slot (keeps
  its default on read-back, like `TextMuted`), but its resolved background RGB
  round-trips. (R13.1, engine half.)
- `54-multistop-background-gradient.md` — extend the scene `Background` from a
  fixed `Gradient [2]ColorRole` to `Stops []GradientStop` (2..8 ascending in
  `[0,1]`); `pptx.LinearGradient` is already variadic, so it's a scene-side
  field extension. Empty `Stops` → legacy 2-role path (byte-identical); invalid
  stops → `LayoutWarning` + skip (D-026); the slice makes `Background`
  non-comparable. (R13.3, engine; foundation for R13.2 radial.)
- `55-radial-vignette-background.md` — add a `BackgroundRadial` kind feeding the
  already-variadic `pptx.RadialGradient` (centered 50%-inset focal — a
  spotlight/vignette); reuses the Phase-71 stops resolver (multi-stop or legacy
  2-role). Focal-offset knob deferred (center-only, documented). Byte-identical
  when unused. (R13.2, engine.)
- `56-decoration-color-role.md` — thread a `role pptx.ColorRole` through the
  `ornaments.Recipe` signature (a v0.x public break) + all 6 curated recipes, and
  add `Decoration.Color *pptx.ColorRole` (nil = `ColorAccent`, byte-identical —
  the D-054 pointer pattern). Lets textures/glows be neutral grey, inverse-white,
  or any brand role. (R13.5, engine; `Decoration.Palette` multi-hue → R13.6.)
- `57-surface-fill-gradient.md` — add `Card.FillGradient *GradientFill` (`{From,
  To pptx.ColorRole; Angle int}`) → `pptx.LinearGradient` for a 2-stop card
  surface depth shift; nil = solid `Fill` (byte-identical). Auto-tint of `To` is
  the soul's (D-026); CardSection/Bento/Container deferred. (R13.8, engine.)
- `58-text-watermark-decoration.md` — a `DecorationText` kind + `Decoration.Text`
  / `FontSize` for a slide-level oversized low-opacity ghost number/word behind
  the body, reusing the `Card.Watermark` text-alpha pattern, `Decoration.Color`
  (D-107), and `RunStyle.FontScale` (>1 grows) for size. Byte-identical when
  unused. (R13.9, engine.)
- `59-starfield-scatter-ornament.md` — a curated `starfield` ornament: a
  box-derived lattice (count scales with the box at a fixed pitch — no caller
  param, since the `Recipe` signature is fixed) perturbed by a deterministic
  integer hash, with per-dot size + alpha variance and the D-107 role color.
  Capped for file size; multi-hue `Decoration.Palette` deferred. (R13.6, engine.)
- `60-pattern-density-pitch.md` — thread a trailing `pitch pptx.EMU` through the
  `Recipe` signature (a 2nd v0.x break) + `Decoration.Pitch`, so `grid_dots` /
  `noise_overlay` / `starfield` derive their count from the box at a caller pitch
  (consistent visual density at any box size). `pitch == 0` = the legacy fixed
  count (byte-identical); capped with a past-cap `LayoutWarning`. (R13.7, engine.)
- `61-gradient-mesh-background.md` — a `BackgroundMesh` kind + `Background.Mesh
  []MeshGlow` (`{Anchor; Color ColorRole; Radius EMU; Alpha int}`): a base canvas
  fill + N low-alpha caller-anchored radial glows pooled over it (the cover mesh
  look), reusing `pptx.RadialGradient`. Empty mesh → no shapes; byte-identical
  when unused. (R13.4, engine.)
- `62-focal-glow-backdrop.md` — `Card.Backdrop *Decoration` drawn behind the
  card's computed box (before its fill) via the existing `renderDecoration` with
  the card box as the region — a focal halo that tracks the card across any
  layout. nil = byte-identical. (R13.10, engine.)
- `63-image-framing.md` — `(*Image).SetCornerRadius(RadiusRole)` /
  `SetElevation(ElevationRole)` builder methods (thin wrappers over the existing
  `applyCornerRadius`/`applyShadow` on the pic spPr) + scene `Image.CornerRadius`
  / `Elevation`; zero tokens self-gate (byte-identical), structural G6 round-trip.
  (R13.11, engine; `DecorationAsset` framing deferred.)
- `64-photographic-background.md` — the photographic-imagery class: a slide
  background legibility `Scrim` (solid or transparent→color gradient overlay drawn
  after any base fill) + a `Duotone` two-tone recolor of a photo background,
  realized by a new builder `(*Image).SetDuotone(shadow, highlight Color)` emitting
  an `<a:duotone>` blip effect (registered in `restorenamespaces`). nil = byte-
  identical; structural G6 round-trip + read accessor. (R14.1, engine half;
  image-as-card-fill → Phase 82, uniform cover-fit → V1.x.)
- `65-image-card-fill.md` — image-as-card-fill: a builder `WithImageFill(src)`
  `ShapeOption` emits an `<a:blipFill>` on a shape's `spPr` (cover-fit center-crop
  via `<a:srcRect>` from the format-header dims — §7/D-046) + scene `Card.ImageFill`
  (an `AssetID`) fills a card surface with a photo. Namespace context rule:
  `blipFill` under `spPr` → `a:` (vs the pic's `p:`). "" / nil byte-identical;
  policy stays `HasAsset:false` (native chrome, not a pic). (R14.1, engine half,
  part 2; D-117.)
- `66-styled-table-matrix.md` — styled comparison matrix: additive scene
  `Table.Style *TableStyle{HeaderFill, Zebra, HighlightCol, RowLabelCol,
  HeaderGroups}` composes the native-table builder (`SetFill`/`SetBorders`/
  `MergeRight`), controlling every cell fill explicitly (the styled path avoids
  `applyStyling`); nil = the plain banded table (byte-identical). CellKind glyphs
  are *not* a native-`a:tbl` feature (cells hold only text) — composed with a
  `Bento` of `Checklist`/`IconRows` (D-095/D-100). (R14.3, engine half; D-118.)
- `67-timeline-roadmap.md` — a NEW `Timeline` IR node: a horizontal axis with
  milestones at proportional `0..1` positions, optional phase `Bands` behind it,
  and optional swimlanes (`Lanes`). Markers (accent dots / curated icons), the
  axis line, and staggered labels compose from native preset shapes (no media);
  integer-EMU layout is deterministic. `AccentIndex` cycles a pinned token set.
  Catalog 28 → 29; vertical orientation + a date milestone type deferred. (R14.4,
  engine; D-119.)
- `68-quote-testimonial.md` — enrich the `Quote` node additively into a designed
  testimonial: `Mark bool` (oversized low-emphasis quotation glyph behind the
  text), `AvatarAssetID` (a rounded author avatar), structured `AttributionName/
  Role/Company`, and `LogoAssetID` (a customer logo). Any enrichment field
  switches to the testimonial layout; a plain Text+Attribution quote is
  byte-identical. Avatar/logo make the Quote asset-bearing (serial determinism);
  policy stays `HasAsset:false` (native text + pics). (R14.5, engine; D-120.)
- `69-number-locale-format.md` — a deterministic, stdlib-only `scene.NumberFormat`
  + `FormatNumber(v, f)` (grouping, decimal sep, currency symbol/placement,
  percent, compact K/M/B/T, prefix/suffix) and a typed `Stat.Number`/`Format`
  path that formats then shrinks-to-fit (fixes the "$4,000+" wrap). Raw `Value`
  unaffected (byte-identical); a mechanism (D-026), not a visual token. Closes
  R14.2's number-format engine atom (the rest of R14.2 is product, D-004).
  (R14.13, engine half; D-121.)
- `70-native-dataviz.md` — a NEW `DataMark` node: native (no-raster) vector
  micro-charts — a progress `Bar`, a `Bars` group, and a `Sparkline` — from preset
  rects + lines in theme colors; values 0..1, integer-EMU deterministic, embeds in
  a card/cell. A new `pptx.WithFlipV` shape option lets sparkline upward segments
  draw without a negative extent. Catalog 29 → 30. Arc-based marks (donut, gauge)
  → Phase 88 (a `blockArc` adjust-guide builder seam). (R14.8, engine; D-122.)
- `71-dataviz-arcs.md` — completes R14.8: a builder `pptx.AddBlockArc(box, start,
  sweep, innerRatio)` (native `<a:prstGeom prst="blockArc">` ring sector via
  adj1/adj2/adj3) + two new `DataMarkKind` values `DataMarkDonut` (a single-value
  ring + centered label, e.g. a 331° arc at 92%) and `DataMarkGauge` (a 270°
  speedometer). Value + remainder arcs (no hole, no asset); deterministic; catalog
  stays 30. (R14.8, engine; D-123.)
- `72-quadrant-matrix.md` — a NEW `Quadrant` node: a 2x2 positioning map with
  labeled X/Y axes (low/high end captions), optional per-quadrant tint + title,
  and items plotted at (x,y) in [0,1] (origin bottom-left). Axes, dividers, dots,
  and labels are native shapes; labels edge-flip + clamp on-canvas; integer-EMU
  deterministic. Catalog 30 → 31. (R14.9, engine; D-124.)
- `73-logo-wall.md` — a NEW `LogoWall` node: an N-up grid of logo assets, each
  contained (not cropped) + centered via `containBox` (format-header dims,
  §7/D-046), optionally recolored to a uniform tone (mono/brand via the duotone
  seam, D-116) so a mixed set coheres; optional caption. Asset-bearing; missing
  logos warn + skip. Catalog 31 → 32. (R14.7, engine; D-125.)
- `74-footnotes.md` — `SceneSlide.Footnotes []RichText` rendered in a reserved
  bottom band (above the chrome footer) in the muted role; the body region shrinks
  to reserve it so footnotes never overlap the body/footer. A scene
  `RunStyle.Superscript` marks references (maps to the builder's `BaselineRel`).
  Cap → drop + warn; empty = byte-identical. (R14.12, engine half; D-126.)
- `75-hierarchy-tree.md` — a NEW `Tree` node: a root with children as a balanced
  top-down (or left-right) tidy tree — leaves spread evenly, internal nodes
  centered over their leaf descendants, parent→child elbow connector edges (all
  H/V, no flip), soul-styled node cards. Native, deterministic; depth/breadth past
  the region clamp + warn. Catalog 32 → 33. (R14.10, engine; D-127.)
- `76-funnel-cycle.md` — two NEW nodes: `Funnel` (a stepped stack of bands
  tapering in width + optional per-stage values) and `Cycle` (stage cards placed
  evenly on a ring with directional connector arrows — straight lines flip-aware +
  chevron heads rotated to the chord). Branch (1→M) is covered by the `Tree` node.
  Native, deterministic; catalog 33 → 35. (R14.11, engine; D-128.)
- `77-image-annotations.md` — an additive `Image.Annotations` overlay: numbered
  pins at fractional (0..1) coords with optional leader-line captions + highlight
  boxes, drawn as native shapes over the picture. nil = byte-identical; a field,
  not a node. (R14.17, engine; D-130.)

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
