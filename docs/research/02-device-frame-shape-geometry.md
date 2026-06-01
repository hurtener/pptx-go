# Brief 02 — device-frame shape geometry

**Subsystem:** Curated assets (icons, ornaments, frames)
**Authored:** 2026-06-01
**Motivating phase:** Phase 10 — Frame chrome (RFC §14.3, §14.4)

## 1. Question

How should pptx-go draw the four V1 device frames — `browser`, `phone`,
`desktop`, `laptop` — as **native PPTX shapes** (no rasters), such that:

- the bezel reads as the intended device at slide scale,
- the recipe yields an **interior region** the renderer drops an image into,
- the geometry is **deterministic** (byte-identical re-render — D-035),
- all visible color flows through **Theme tokens** (P2), and
- the recipe **composes the public `pptx` builder only** (P1) — no raw OOXML.

The companion question for §14.4: how is a frame **referenced from the IR**
and how does a **caller-extended** frame slot into the same path?

## 2. Prior art surveyed

- **The existing leaf composers** (`scene/render_leaves.go`) — the
  established pattern for "emit a few native shapes into a `pptx.Box`":
  `renderCallout` (accent bar + inset text), `renderChip` (roundRect +
  centered label), `renderArrow` (preset-geometry arrow). A frame recipe is
  the same shape, one level up: a cluster of `AddShape` calls bounded by a
  region.
- **The builder shape surface** (`pptx/shape.go`): `AddShape(geom, box,
  …opts)` with `ShapeRect`, `ShapeRoundRect`, `ShapeEllipse`; `WithFill`,
  `WithLine`, `WithRadius(RadiusRole)`. `WithRadius` resolves a theme radius
  token against the active theme and converts it to OOXML's `roundRect`
  adjust fraction — exactly what a rounded device corner needs. Fills and
  lines resolve token roles against the presentation theme **inside**
  `AddShape` (`s.activeTheme()`), so a recipe that passes `TokenColor(role)`
  needs no theme handle of its own.
- **The Decoration node** (`scene/nodes.go`): the IR's existing
  curated-vs-extension precedent. `Decoration` carries both a
  `DecorationKind` enum *and* a free-form `Preset string` curated-ornament
  name. The curated set is closed; the string selects within it (and, in
  Phase 13, within a caller-extended registry). Frames face the identical
  enum-vs-name tension and should resolve it the same way.
- **The macOS/Chrome "window chrome" idiom** and stock device mockups
  (Apple device frames, Chrome devtools device toolbar, the `browserframe`/
  `deviceframe` families): a browser is a rounded window + a title/toolbar
  strip + three "traffic-light" dots + a content viewport; a phone is a
  tall rounded slab + a notch/status strip + a home indicator + a viewport;
  desktop is a monitor panel + a neck + a foot; laptop is a screen panel +
  a trapezoidal/└ base. None of these need curves beyond rounded rectangles
  and ellipses — i.e. all expressible with `ShapeRect`/`ShapeRoundRect`/
  `ShapeEllipse` preset geometry. No custom path is required in V1 (the
  SVG-path translator is Phase 12).
- **EMU determinism note** (D-035): integer EMU arithmetic on the region is
  deterministic; the only non-determinism risks are map iteration and
  wall-clock — neither appears in a pure geometry function.

## 3. Findings

- **F1 — Rounded rectangles + ellipses suffice.** Every V1 frame is
  expressible with `ShapeRect`, `ShapeRoundRect` (+ `WithRadius`), and
  `ShapeEllipse`. No preset-geometry beyond the curated subset already in
  `pptx/shape.go`, and no custom path, is needed. This keeps Phase 10 free
  of the Phase 12 SVG-path translator.
- **F2 — The recipe returns the interior, the renderer inserts the image.**
  RFC §14.3: "The recipe is positioned by the renderer; the image is
  inserted into the interior region." The clean seam is a recipe that emits
  the bezel shapes into a region and **returns the interior `pptx.Box`**;
  the renderer (`render_image.go`) then resolves the asset and calls
  `AddImage` into that interior. Asset resolution and alt-text stay
  centralized in the renderer; the recipe is pure geometry. The recipe also
  reports how many bezel shapes it emitted so `Stats.Shapes` stays honest.
- **F3 — Proportional geometry, not spacing tokens.** A device's
  proportions (a phone status bar ≈ 6% of height, a browser toolbar ≈ a
  fixed ~0.32" strip, a monitor stand ≈ 18% of height) are **intrinsic to
  the device silhouette**, not theme spacing. They should be computed as
  ratios/fixed insets of the region, not resolved from `SpaceRole`. This
  means a recipe needs **no `*Theme` parameter** — color/radius tokens
  resolve downstream inside `AddShape`, and geometry is region math. (A
  theme swap restyles a frame's *color*, never its *silhouette* — correct:
  a phone is a phone in every brand.)
- **F4 — All bezel color maps to existing tokens; no new token surfaces.**
  Bezel body → `ColorSurfaceAlt`; chrome strip / status bar → `ColorSurface`;
  device outline stroke → a thin `ColorSurfaceAlt` line; the browser
  traffic-light dots → `ColorError` / `ColorWarning` / `ColorSuccess`
  (semantic, not literal RGB). No frame introduces a visual property the
  theme doesn't already name, so **no `docs/design/THEME.md` token entry is
  required** — frames are composers of existing tokens (P2 holds via reuse).
- **F5 — Reference a frame by name; keep the enum as the ergonomic alias.**
  The `Image` IR already shipped `Frame FrameKind` (closed enum: `FrameNone`
  + four curated). §14.4 extension is **by string name**. Mirroring
  `Decoration` (enum + `Preset string`), the resolution is: the registry is
  keyed by **name**; the four curated `FrameKind` values map to the four
  reserved curated names; a new optional `Image.FrameName string` field
  selects a name (curated **or** caller-registered) and **takes precedence**
  over the enum when set. The enum stays as the zero-import ergonomic path
  for the curated four; the string is the extension seam. Stage-1 validation
  rejects a `FrameName` absent from the (curated ∪ extension) registry.
- **F6 — Extensions are per-render, not global.** §14.4: "Extensions live on
  the `Scene`/`RenderOption`, not on global state." Each `Render` folds its
  `WithFrameExtension(name, recipe)` options over a **copy** of the curated
  registry, producing a read-only registry consulted during the (parallel)
  compose. No process-global mutable map ⇒ concurrent renders with different
  extension sets don't interfere, and the registry being read-only during
  compose preserves D-035 determinism.
- **F7 — Frame chrome adds shapes PowerPoint validates.** A bezel is several
  overlapping rounded rects + ellipses with token fills — all shapes the
  builder already emits and round-trips. The residual risk is purely visual
  (does it *look* like the device?), which automated schema checks can't
  judge; it needs the PowerPoint / Quick Look / Keynote eyeball pass the
  wave already budgets.

## 4. Recommendations

1. **Recipe shape:** `func(sl *pptx.Slide, region pptx.Box) (interior
   pptx.Box, shapes int)` — no theme param (F3), interior + shape count out
   (F2). Define the named type once in `scene/frames`; alias it from the
   `scene` package so the public extension API reads `scene.FrameRecipe`.
2. **Curated recipes in `assets/frames`**, one file per device, importing
   only `pptx`. `scene/frames` wires the four curated names to them and owns
   the registry + lookup + the per-render extension overlay (F6).
3. **IR reference (F5):** add `Image.FrameName string`; `FrameName` wins when
   non-empty, else the `FrameKind` enum selects a curated name (`FrameNone`
   ⇒ no frame). Record this reconciliation as a settled decision.
4. **Stage-1 validation:** reject an `Image` whose resolved frame name is
   absent from the render's registry (closed-name semantics, §14.4). Because
   the registry is render-option-derived, run this check in `Render` (after
   the option-free structural `ValidateScene`), not in `ValidateScene`
   itself.
5. **Geometry, per device (proportions of `region`):**
   - **browser** — rounded window (`RadiusMD`), a top toolbar strip
     (~0.34" or 12% of height, whichever larger), three traffic-light
     ellipses left-aligned in the strip, interior = region minus the strip
     minus a thin inset.
   - **phone** — tall rounded slab (`RadiusLG`), a status strip at top
     (~6% of height), a home-indicator pill near the bottom (`RadiusFull`),
     interior = the central viewport between them, side bezel inset ~4%.
   - **desktop** — monitor panel (rounded), a stand neck + a foot below it
     (~18% of height reserved for the stand), interior = the panel viewport.
   - **laptop** — screen panel on top, a wide base/keyboard deck below
     (~10% of height), interior = the screen viewport.
6. **Determinism:** integer EMU math only; no map iteration in a recipe; no
   wall-clock. The scene determinism tests already in the suite will catch a
   regression.

## 5. Open questions

- **Aspect-ratio handling.** A frame fixes an interior aspect implied by the
  device; the caller's image may not match. Phase 10 stretches the image to
  fill the interior (the builder's default `FitFill`); aspect-aware
  `contain`/letterboxing is **Phase 11**'s image-composition concern
  (crop/fit), not Phase 10's. Flagged for the Phase 10/11 split in the plan.
- **True OOXML group shape.** RFC §14.3 calls a frame "a shape group". The
  builder exposes no public group-shape API in V1; Phase 10 emits the bezel
  as individually-positioned native shapes bounded by the region (visually
  and round-trip-equivalent). Whether to surface a builder
  `AddGroup`/grouping primitive (so a framed image moves as one object in
  PowerPoint) is a candidate for a later builder phase — noted for the V2
  backlog, not Phase 10.
- **Notch vs. status bar realism for `phone`.** A literal notch needs a
  subtractive cutout the preset geometry can't express without a custom
  path (Phase 12). V1 approximates with a status strip; revisit once the
  SVG-path translator lands.
