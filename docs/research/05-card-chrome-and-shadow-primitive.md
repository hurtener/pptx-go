# Brief 05 — card chrome, the elevation/shadow primitive & icon consumption

**Subsystem:** scene (composites) + pptx (builder shadow primitive) + Curated assets (icon consumption)
**Authored:** 2026-06-02
**Motivating phase:** Phase 14 — Card + CardSection (RFC §11.2, §12, D-026)

## 1. Question

How should pptx-go render the `card` and `card_section` scene nodes as
**native PPTX chrome** — a rounded-rect background with an accent stripe,
an optional icon / eyebrow / header / header-pill, then a body laid out by
`body_layout` — across all v4 card knobs (`fill`, `border_style`, `size`,
`elevation`, `body_layout`, `layout`, `header_pill`), composing the public
builder only (P1)?

Three sub-questions force decisions the phase plan must land:

1. **Elevation.** The Theme resolves an `Elevation` (blur / offset / color /
   alpha) per role, but the builder has **no shadow primitive** — every shape
   emits an empty `<a:effectLst/>` and the resolved elevation is dropped. What
   does rendering a Raised / Elevated card actually require?
2. **Icon placement.** A Card's optional icon is the **first scene node that
   places a curated icon**. The icon registry (`scene/icons`:
   `Curated`/`With`/`Lookup`/`Names`) and `pptx.Slide.AddIcon` both ship
   already (Phase 12), but `cfg.icons` was intentionally **not** stored in
   `renderConfig` to avoid a write-only field. How is the registry wired into
   compose and an unknown icon name failed fast?
3. **IR expansion.** The shipped `Card` lacks `Eyebrow`, `Icon`, `HeaderPill`,
   `BorderStyle`, `Size`, `Layout`. How is the IR grown without breaking the
   byte-identical idempotency invariant (D-035) for existing scenes?

## 2. Prior art surveyed

- **The container composers** (`scene/render_container.go`, `scene/layout`):
  `two_column` / `grid` emit **no shape of their own** — they subdivide the
  slot via the `scene/layout` geometry engine and recurse `renderNode` into
  each sub-slot. A `card_section` reuses exactly this recursion for its body.
- **The leaf composers** (`scene/render_leaves.go`): `callout` and `chip`
  already build "rounded rect + text" chrome via `pptx.AddShape(ShapeRoundRect,
  …, WithFill, WithRadius)` — the card's background and header-pill are the
  same primitive at a larger scale.
- **The curated-asset registry pattern** (frames D-038, icons D-040, ornaments
  D-041): `Curated()` + immutable `With()` overlay + `Lookup`/`Names`, built
  once per `Render` into `renderConfig`, read-only during the parallel compose,
  with a `WithXExtension` RenderOption and a Stage-1 **closed-name** ref check
  (`validateFrameRefs` / `validateOrnamentRefs`) that fails an unknown name
  *before* compose. Icons already have the registry + the registration-time
  SVG validation; only the **consumption** half (store in `renderConfig`,
  `validateIconRefs`, resolve name→bytes→`AddIcon`) is missing.
- **The builder shape surface** (`pptx/shape.go`, `internal/ooxml/slide`):
  `AddShape` resolves fill/line/radius/rotation against the active theme at
  call time. `XShapeProperties` (slide_types.go:222) carries
  `xfrm / prstGeom|custGeom / solidFill|gradFill|noFill / ln` — and **no
  effect list**. The theme's `effectStyleLst` slots are empty `<a:effectLst/>`
  placeholders (`scaffold_assets.go`), unrelated to a per-shape shadow.
- **The gradient precedent** (D-041, brief 04): when ornaments needed real
  glows, V1 **built** the `<a:gradFill>` primitive (wire types in
  `internal/ooxml/slide/gradient.go`, `restorenamespaces` entries, builder
  fill API, round-trip golden) rather than approximate it. The shadow
  primitive is the identical situation one effect over.
- **OOXML outer shadow** (ECMA-376 §20.1.8.45, `CT_OuterShadowEffect`):
  `<a:effectLst><a:outerShdw blurRad="" dist="" dir="" rotWithShape="0">
  <a:srgbClr val=""><a:alpha val=""/></a:srgbClr></a:outerShdw></a:effectLst>`.
  `blurRad`/`dist` are EMU; `dir` is an angle in 1/60000° (0 = east, 5400000 =
  south/down). The `effectLst` element sits **after `ln`** in `CT_ShapeProperties`'s
  ordered content model — so the new field appends after `Line` in
  `XShapeProperties`. A drop shadow under a card is `dir≈5400000` (straight
  down) with the theme `OffsetX/OffsetY` resolved to a `dist`+`dir` pair (or,
  simpler and lossless, emit `dist` from the offset magnitude and `dir` from
  its angle).
- **pengui-slides v4.16 card** — the knob set Phase 14 reproduces 1:1
  (RFC §16 mapping table, line ~1379): `fill` (surface role), `border_style`
  (none / solid / accent), `size` (sm / md / lg → padding + min height),
  `elevation` (flat / raised / elevated), `body_layout` (vertical / horizontal
  child stacking), `layout` (header arrangement: where the icon/eyebrow/title/
  pill sit relative to the body), and `header_pill` (a small pill badge in the
  header row, e.g. a status tag). The card body holds **leaves**; the
  `card_section` body holds **containers** (grid / two_column / nested cards).

## 3. Findings

- **F1 — A card emits no genuinely new OOXML capability except the shadow.**
  Everything else (rounded rect, accent stripe, text frames, header pill,
  icon, nested layout) composes existing builder calls. Per P1 the only new
  *builder capability* Phase 14 needs is the shadow primitive — and only
  because elevation is a real visual property the theme already tokenizes.
- **F2 — The shadow is a real builder primitive, not a card-local hack.**
  `outerShdw` belongs on `pptx` (`WithShadow(Elevation)` / `WithElevation(role)`
  as a `ShapeOption`), with wire types in `internal/ooxml/slide`, two
  `restorenamespaces` entries (`effectLst`, `outerShdw` → `a`), and a
  round-trip golden (G6). This keeps P3 intact (raw XML only in
  `internal/ooxml`) and makes elevation reusable by any future node, not just
  cards. It mirrors D-041 exactly.
- **F3 — `dir`/`dist` vs `OffsetX`/`OffsetY`.** The theme `Elevation` carries
  cartesian offsets; `outerShdw` wants polar (`dist` magnitude + `dir` angle).
  The default elevations are straight-down (`OffsetX=0`), so `dist=OffsetY`,
  `dir=5400000`. For a general offset, `dist=round(hypot(dx,dy))`,
  `dir=round(atan2(dy,dx)·60000/(π/180) mod 360°)`. Determinism (D-035) holds:
  integer EMU in, integer 1/60000° out, no float in the serialized bytes
  beyond a rounded integer.
- **F4 — Icon consumption closes a documented Phase-12 deferral.** Store the
  built icon registry in `renderConfig` (`cfg.icons`, curated ∪ `iconExt`),
  add `validateIconRefs(s, cfg.icons)` to Stage-1 (mirroring
  `validateOrnamentRefs`), and resolve a card's `Icon` name → bytes →
  `ps.AddIcon(svg, box, …)` at compose. An unknown icon name fails the render
  before any slide composes — the same closed-name contract as frames/ornaments.
- **F5 — Icons are native, so a plain card stays media-free.** `AddIcon`
  emits a `custGeom` path shape, not a `pic` — it registers **no global
  media**. So a Card whose body is text + icon is parallel-safe and must
  **not** be classified asset-bearing. Only a Card / CardSection whose *body*
  contains an `Image` / `CodeBlock` is media-bearing; the existing
  `nodeUsesAssets` container recursion already models this — extend it to
  recurse `Card.Body` and `CardSection.Body` (treating the card chrome itself
  as media-free).
- **F6 — IR growth is additive with zero-value = current behavior.** New
  fields (`Eyebrow string`, `Icon string`, `HeaderPill string`,
  `BorderStyle BorderStyle`, `Size CardSize`, `Layout CardLayout`) default to
  empty/zero, which must reproduce today's output byte-for-byte for an
  existing `Card{Header, Body, BodyLayout, Fill, Outline, Elevation}`. New
  enums are re-exported into the scene package (like `Anchor`/`Crop`/`Fit`) so
  the IR reads `scene.BorderSolid`, not `pptx.…`. Note `Outline bool` and a
  new `BorderStyle` enum overlap — the plan must reconcile (keep `Outline` as
  the zero-state shorthand, or fold it into `BorderStyle` with `BorderNone`
  as zero; the latter is cleaner but the former preserves the field).
- **F7 — `card_section` is a card whose body is containers.** The only
  structural difference from `card` is what its body accepts and that it's a
  top-level node. Chrome rendering is shared; the body layout differs (a
  card stacks leaves per `body_layout`; a card_section recurses grids /
  two_columns / cards through the existing layout engine). One shared chrome
  helper, two body strategies.
- **F8 — Card internal geometry.** A card's slot decomposes into: (1) the
  background rounded-rect (full slot, `fill` + `border_style` + `elevation`
  shadow); (2) a thin accent stripe along one edge (a second rounded/plain
  rect in the accent token); (3) a header row inset by the card padding
  (icon □ + eyebrow/title stack + right-aligned header-pill); (4) the body
  region below the header, inset by padding, laid out by `body_layout`.
  Padding scales with `Size`. All integer EMU → deterministic.

## 4. Recommendations

(Recommendations are inputs; the phase plan binds.)

- **R1 — Build the shadow primitive (PR#1).** Add `XOuterShadow` /
  `XEffectList` wire types to `internal/ooxml/slide`, append `EffectList
  *XEffectList` to `XShapeProperties` after `Line`, register `effectLst` /
  `outerShdw` in `restorenamespaces`, and expose `pptx.WithShadow(Elevation)`
  + `pptx.WithElevation(role ElevationRole)` `ShapeOption`s that resolve the
  token at `AddShape` time. Ship a round-trip golden and a smoke check. This
  is the D-041 pattern.
- **R2 — Deliver as a split.** PR#1 = the builder shadow primitive (self-
  contained, independently mergeable). PR#2 = `card` / `card_section` + icon
  wiring, consuming the primitive. One plan, two clearly-scoped PRs — as
  Phase 13 split the gradient primitive from the ornaments.
- **R3 — Wire icons exactly like ornaments.** `cfg.icons = icons.Curated()`
  overlaid with `iconExt`, built in `Render` before compose;
  `validateIconRefs` in Stage-1; name→bytes→`AddIcon` at compose. Reuse the
  registration-time SVG validation already in place.
- **R4 — Grow the Card IR additively** with `Eyebrow` / `Icon` /
  `HeaderPill` / `BorderStyle` / `Size` / `Layout`, zero-values preserving
  current output; re-export the new enums into `scene`. Reconcile `Outline`
  vs `BorderStyle` explicitly in the plan + a D-NNN.
- **R5 — Extend `nodeUsesAssets`** to recurse `Card.Body` /
  `CardSection.Body` so an image-bearing card renders sequentially while a
  plain (text+icon) card stays parallel — preserving idempotency and
  parallelism both.
- **R6 — Treat elevation as a mechanism, not a token addition.** The shadow
  reuses the existing `Elevation` token; no new theme token is added, so
  `docs/design/THEME.md` gets a **mechanism note** (no new taxonomy entry),
  per the gradient/rotation precedent.

## 5. Open questions

- **Q1 — `layout` (header arrangement) value set.** pengui-slides v4 cards
  vary header placement (icon-left vs icon-top, title-only vs eyebrow+title).
  The exact enum set Phase 14 ships vs defers is a plan decision; default to
  the variants the Galici / Databricks reference decks actually use and defer
  the rest (note the deferral — no silent cap, RFC §11.3 growth area).
- **Q2 — `border_style` accent variant.** Whether `border_style` includes an
  "accent" (themed) border distinct from a neutral solid border, or whether
  that's expressed via the accent stripe alone. Plan decides; record in the
  D-NNN.
- **Q3 — Shadow on rotated shapes.** `outerShdw` has `rotWithShape`; cards are
  not rotated in V1, so default `rotWithShape="0"`. If a future node wants a
  rotating shadow, revisit — out of scope for Phase 14.
