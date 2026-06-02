# Brief 04 — preset ornament recipes & gradient glows

**Subsystem:** Curated assets (icons, ornaments, frames) + pptx (builder fills/transform)
**Authored:** 2026-06-02
**Motivating phase:** Phase 13 — Curated ornaments + Decoration node (RFC §14.2, §11.1, D-005)

## 1. Question

How should pptx-go render the six V1 preset ornaments (`glow_ring`,
`radial_glow`, `grid_dots`, `corner_bracket`, `chevron_arrow`, `noise_overlay`)
as **native PPTX shapes** composed at render time, in the active accent token,
with caller-controllable opacity / rotation / size / anchor / bleed / layer —
and what builder primitives must land first to make `radial_glow` and
`glow_ring` read as real glows rather than banded approximations?

## 2. Prior art surveyed

- **The frames recipe seam** (`assets/frames`, `scene/frames`, D-038): the
  curated-asset pattern ornaments mirror — a `Recipe` plus a
  `Curated/With/Lookup/Names` registry and a per-render extension overlay.
- **The builder fill/transform surface** (`pptx/fill.go`, `pptx/shape.go`,
  `internal/ooxml/slide`): V1 ships `SolidFill` + `NoFill`, alpha on **literal**
  colors only (`RGBA`), and **no gradient fill** and **no rotation API** — though
  the `XTransform2D.Rotation` wire field exists and `restorenamespaces` already
  maps `gradFill/gs/gsLst/lin`. Gradient *structs* for shape fills are absent
  from `slide_types.go`.
- **OOXML gradient fill** (ECMA-376 §20.1.8.33): `<a:gradFill>` = a `<a:gsLst>`
  of `<a:gs pos="0..100000">` stops, plus either `<a:lin ang="" scaled="">`
  (linear, angle in 1/60000°) or `<a:path path="circle|rect|shape">` with a
  `<a:fillToRect>` (radial / focal). A radial glow is a 2-stop `path="circle"`
  gradient from the accent color (opaque, centre) to the same color at
  `alpha=0` (edge).
- **pengui-slides v4.16 ornaments**: the named set Phase 13 reproduces. They are
  soft decorative layers (glows, dotted textures, brackets, chevrons, grain),
  always tinted from the theme accent, often bled past the slide edge.
- **RFC §10.2 layout order**: decorations are placed by layer —
  `background` first (behind body), `foreground` last (over body) — then
  `section_divider` overrides to full-bleed. This is the z-order the renderer
  must impose; OOXML shape-tree order *is* z-order (first = back).

## 3. Findings

- **F1 — Real gradients unblock honest glows (decision: build them).** A
  `radial_glow`/`glow_ring` rendered as alpha-layered concentric solids bands
  visibly. A `path="circle"` gradient (accent → accent@alpha0) is the correct
  primitive and is a modest, self-contained wire addition. The builder's
  Phase-02 block listed `GradientFill` but it was never built — this closes that
  gap. **Add `XGradientFill` (gsLst + lin | path + fillToRect) to
  `slide_types.go`, a `GradientFill` field on `XShapeProperties`, and a public
  `pptx.LinearGradient` / `pptx.RadialGradient` Fill.** `fillToRect` joins the
  `restorenamespaces` map (`gs/gsLst/lin/gradFill` are already there; `path` was
  added for icons).
- **F2 — Rotation needs a tiny builder option.** `XTransform2D.Rotation` exists;
  expose `WithRotation(deg)` on `AddShape` (and the picture path) — `deg ×
  60000`, normalized. `chevron_arrow` and a caller-rotated asset decoration use
  it. A **multi-shape** ornament cannot be rotated as a unit (the builder has no
  group transform — V2); per-shape rotation suits the rotationally-symmetric
  glows/grid and is acceptable for V1, with the limitation documented.
- **F3 — Opacity needs alpha on token colors.** Decoration `opacity` dims the
  whole ornament; today alpha rides only on `RGBA` *literals*. Add
  `pptx.TokenColorAlpha(role, alpha)` (the token analogue of `RGBA`) so a recipe
  fills with the accent token at a caller opacity (P2 preserved — still a token).
- **F4 — All six are expressible with rect/ellipse/roundRect/chevron + the new
  primitives.** `glow_ring` = an annulus (two concentric radial-gradient
  ellipses, or a thick-lined ellipse). `radial_glow` = one `path="circle"`
  gradient ellipse, bled. `grid_dots` = a deterministic array of small accent
  ellipses. `corner_bracket` = two thin rects forming an L. `chevron_arrow` =
  the existing `ShapeChevron` (rotatable). `noise_overlay` = a deterministic
  sparse low-alpha dot scatter (a grain approximation — no per-pixel noise
  natively; documented). No custom path geometry (icons' `custGeom`) is needed.
- **F5 — Decoration IR must grow.** The shipped `Decoration{Kind, Preset,
  AssetID, Layer, Anchor}` lacks the master-plan fields. Add `Offset` (Position,
  EMU from the anchor), `Size` (the ornament box; zero = a sensible default),
  `Bleed` (bool — allow the box to extend past the slide edge via negative
  offsets, RFC §14.2), `Opacity` (0..1), `Rotation` (degrees). Anchor + offset +
  size + bleed → the placement box; `Bleed` relaxes the on-canvas clamp.
- **F6 — Layer z-order is the renderer's job.** Per RFC §10.2, the renderer
  splits top-level decorations out of the body stack: `background` decorations
  compose **before** body nodes, `foreground` **after**. This mirrors how
  `SectionDivider` is already special-cased in `layout()`. Decorations are
  *overlays* — they do not consume body-stack height.
- **F7 — Preset vs asset disposition (RFC §12).** A `DecorationPreset` renders
  as native shapes (the recipe); a `DecorationAsset` renders as a `pic` with
  bleed-aware offsets (composes the Phase-11 image path + the new rotation).

## 4. Recommendations

1. **Builder foundations first (a separate, smaller change):** gradient fills
   (linear + radial), `WithRotation`, `TokenColorAlpha` — each with a round-trip
   golden. These are general builder primitives, not ornament-specific.
2. **Ornament recipe contract:** `Recipe func(sl *pptx.Slide, box pptx.Box,
   style Style) int` returning the shape count, where `Style{ Color pptx.Color;
   Rotation float64 }` carries the accent token (already alpha-adjusted for
   opacity) and rotation. Curated recipes in `assets/ornaments`; registry in
   `scene/ornaments` mirroring `scene/frames`.
3. **Decoration composition:** add the five IR fields (F5); `render_decoration.go`
   computes the box from anchor+offset+size+bleed, builds the styled color from
   the accent token + opacity, dispatches preset→recipe / asset→pic, and the
   renderer orders by `Layer` (F6).
4. **Glossary + decisions:** `Ornament`, `Bleed`, gradient terms; D-041 (V1
   gradient fills), the carried-forward builder fixes, and the Decoration IR
   expansion.

## 5. Open questions

- **Group-shape rotation.** Rotating a multi-shape ornament as a unit needs a
  builder group transform (absent in V1). Deferred (F2); per-shape rotation is
  the V1 behavior. Revisit if a curated ornament genuinely needs unit rotation.
- **True grain for `noise_overlay`.** Per-pixel noise is a raster; V1
  approximates with a deterministic sparse dot scatter. A caller wanting real
  grain supplies an asset decoration. Flagged, not solved.
- **Pattern fills.** `pattFill` (hatch textures) is namespace-mapped but
  unbuilt; no V1 ornament needs it. Left for a future builder phase.
