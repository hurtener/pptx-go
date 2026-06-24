# Brief 63 — Rounded + shadow image framing (R13.11)

> Informs Phase 80 (Wave 13). Engine req R13.11
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase

Reference figures and logo tiles are framed with consistent rounded corners and
a soft drop shadow. Deckard images are plain rectangular pics with no radius or
shadow, so any image-bearing slide loses the rounded, elevated finish the cards
already have. Phase 80 adds corner-radius + elevation options to the scene
`Image` node, resolved from the theme's radius/elevation tokens.

## 2. Subsystem / files

- `pptx/media.go` — `*Image` (the builder picture handle) + its `Set*` methods.
- `pptx/shape.go` — `applyCornerRadius(spPr, radius, box)` and `applyShadow(spPr,
  e)` already realize a roundRect adjust and an `<a:effectLst><a:outerShdw>` on a
  `*slide.XShapeProperties`.
- `scene/nodes.go` / `scene/render_image.go` — the scene `Image` + `renderImage`.

## 3. Findings

- **The builder helpers already exist for shapes.** `applyCornerRadius` and
  `applyShadow` operate on a `*slide.XShapeProperties`, and a picture's
  `pic.ShapeProperties` is the same type (it already carries `Transform2D`,
  `PresetGeom`, `EffectList`). So adding `(*Image).SetCornerRadius(RadiusRole)`
  and `(*Image).SetElevation(ElevationRole)` is a thin wrapper — set the pic
  geometry to `roundRect`, resolve the token via `im.s.activeTheme()`, and call
  the existing helpers.
- **The box comes from the pic transform.** `applyCornerRadius` needs the box;
  reconstruct it from `pic.ShapeProperties.Transform2D.Offset`/`.Extent` (a
  `picBox` helper).
- **Self-gating zero values keep byte-identity.** `ResolveRadius(RadiusNone) ==
  0`, so `SetCornerRadius` returns early (rectangular, no `PresetGeom` change);
  `applyShadow` is a no-op on a flat elevation. So `renderImage` can call both
  unconditionally and a Phase-10 image (zero radius/elevation) is byte-identical.
- **G6 round-trip is structural.** The emitted pic carries `<a:prstGeom
  prst="roundRect">` + `<a:effectLst><a:outerShdw>`; both survive a write →
  `pptx.Open` → re-write cycle losslessly (the round-trip test asserts they
  persist) — the same structural-round-trip proof used for gradient fills, no new
  read accessor required.
- **DecorationAsset deferred.** R13.11 also names `DecorationAsset`; the scene
  `Image` is the primary case and satisfies the acceptance criterion. The new
  builder methods are reusable for a `Decoration` radius/elevation later (§4.3).
- **No OOXML / `restorenamespaces` change.** `prstGeom`/`effectLst`/`outerShdw`
  already emit (the shape path proves it).

## 4. Recommendations

- Add `(*Image).SetCornerRadius(RadiusRole)` + `(*Image).SetElevation(
  ElevationRole)` to the builder (+ a `picBox` helper).
- Add `Image.CornerRadius RadiusRole` + `Image.Elevation ElevationRole` to the
  scene node; `renderImage` calls the two setters (self-gating at zero).
- Tests: a builder round-trip golden (an image with radius+elevation survives
  write → reopen → re-write with `roundRect` + `outerShdw`); a scene Image with
  the tokens emits both; a zero-token image is byte-identical. THEME.md note (the
  radius/elevation tokens already exist — a mechanism), glossary, compose-a-scene
  skill, docs/site. D-114.

## 5. Open questions

- `DecorationAsset` radius/elevation → a follow-up (the builder methods are
  ready).
- Aspect-aware fit interplay (R14.1) is separate; framing is orthogonal to fit.
