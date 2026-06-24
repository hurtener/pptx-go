# Brief 61 — Gradient-mesh background (R13.4)

> Informs Phase 78 (Wave 13). Engine req R13.4
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059). Builds on the
> radial fill (Phase 72 / D-106) and the role-colored glows (Phase 73 / D-107).

## 1. Motivating phase

The reference cover and light content slides carry a soft "mesh" glow — diffuse
colored light pooling at one or two corners over the paper, not a single
straight gradient. The current single linear/radial full-bleed fill cannot
produce the multi-pool look. Phase 78 adds a composable mesh: several very-low-
alpha radial glows pooled at caller-chosen anchors over the canvas.

## 2. Subsystem / files

- `scene/background.go` — `BackgroundKind` + the `Background` struct.
- `scene/render.go` — `renderBackground`.
- `pptx/fill.go` — `RadialGradient` (already exists; the glow primitive).
- The scene `Anchor` type — `Anchor.Point(box)` resolves a slide point (used by
  `decorationBox`).

## 3. Findings

- **A `BackgroundMesh` kind is the convenience the req asks for.** Add it to
  `BackgroundKind` (last) and a `Background.Mesh []MeshGlow` field
  (`{Anchor; Color pptx.ColorRole; Radius pptx.EMU; Alpha int}`). Each glow is a
  radial-gradient ellipse centered on its anchor (`Anchor.Point(full)`), sized
  `2*Radius`, fading from `TokenColorAlpha(Color, Alpha)` (center) to alpha 0
  (edge). Drawn in slice order → deterministic.
- **Base canvas under the glows.** A mesh reads "over the paper", so draw a base
  `SolidFill(TokenColor(bg.Color))` first (`bg.Color` zero = `ColorCanvas` = the
  paper/dark canvas), then the glows on top. This makes the kind self-contained
  (a cover = paper + pooled glows) without a second Background kind.
- **Empty `Mesh` → nothing (the absent-config case).** With no glows configured,
  `BackgroundMesh` emits no shapes on a light slide (and the dark canvas on a
  dark variant, matching `BackgroundNone`) — so "absent config → no shapes".
- **Reuses the role-color + alpha tokens (P2).** The glow color is a surface
  role; a theme swap re-paints the mesh. `Alpha` is the OOXML opacity the soul
  keeps subtle (R13.13).
- **Parallel-safe, deterministic.** Not asset-bearing (`slideUsesAssets` stays
  false); fixed-order integer-EMU geometry.
- **No OOXML / `restorenamespaces` change.** Same `<a:gradFill>`/`<a:path>`
  radial ellipses the glow ornaments already emit.
- **Byte-identical when unused.** A new kind appended last; existing kinds
  unchanged.

## 4. Recommendations

- Add `BackgroundMesh` kind + `MeshGlow{Anchor; Color pptx.ColorRole; Radius
  pptx.EMU; Alpha int}` + `Background.Mesh []MeshGlow`.
- `renderBackground` `case BackgroundMesh`: empty `Mesh` → None behavior; else a
  base canvas fill + one radial-gradient ellipse per glow (`Radius > 0`).
- Tests: a 2-glow mesh emits a base rect + 2 distinct-anchor radial fills; an
  empty mesh emits nothing (light); determinism. THEME.md note, glossary,
  compose-a-scene skill, docs/site Background. D-112.

## 5. Open questions

- A curated `gradient_mesh` *ornament* (vs the Background kind) was the spec's
  alternative; the Background kind is cleaner for a full-slide wash and composes
  the base canvas. The role-colored glow ornaments (Phase 73) already cover the
  per-glow decoration case for callers who want foreground pools.
- Per-glow shape (ellipse vs soft-edged blob) — an ellipse radial is the
  deterministic primitive; true blur is not an OOXML fill mode.
