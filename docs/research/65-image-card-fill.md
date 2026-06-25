# Brief 65 — Image-as-card-fill: cover-fit photo surface (R14.1, part 2)

> Informs Phase 82 (Wave 14). Engine req R14.1
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · both — engine half, part 2; D-059).

## 1. Motivating phase

R14.1's photographic class has three engine atoms: a full-bleed photo background
(already present), a legibility scrim + photo duotone (Phase 81, D-116), and
**image-as-fill** — filling a Card/panel surface with a cover-fit photo instead of
a solid color. Phase 82 lands the last one as a general builder mechanism plus the
Card surface wiring.

## 2. Subsystem / files

- `internal/ooxml/slide/slide_types.go` — `XShapeProperties` (the shape's fill
  choice) + `XBlipFillProperties` (already used by the picture path).
- `internal/ooxml/restorenamespaces.go` — the context-sensitive prefix map.
- `pptx/shape.go` — `AddShape` + the `ShapeOption` set.
- `pptx/media.go` — `addImagePart`, `ImageSource`, the image dim reader pattern.
- `scene/render_card.go` — `cardChrome` + `renderCardChrome` (the surface fill).
- `scene/nodes.go` / `scene/render.go` — the `Card` node + `nodeUsesAssets`.

## 3. Findings

- **A shape image fill is `<a:blipFill>` inside `spPr` — a new builder `Fill`
  capability (P1).** `XShapeProperties` carries `solidFill`/`gradFill`/`noFill`
  today; adding a `BlipFill *XBlipFillProperties` field (the same struct the pic
  uses) realizes a blip surface fill. `applyFill` has no slide handle (it can't
  register an image part), so the cleanest seam is a `WithImageFill(src
  ImageSource)` `ShapeOption`: `AddShape` (which *does* hold the slide) resolves
  the source, registers the part via `addImagePart`, and sets `BlipFill`. It wins
  over `WithFill` (clears solid/gradient/noFill).
- **Namespace gotcha: `blipFill` is `p:` in a pic but `a:` in `spPr`.** The
  `restorenamespaces` map keys on local name; `"blipFill" → "p"` (for the pic).
  Inside `spPr` it must be `a:blipFill`. `prefixFor` is already context-sensitive
  (it special-cases `txBody` by parent `tc`); add the same rule: `blipFill` under
  parent `spPr` → `a`. Without it the shape fill emits bare-prefixed `p:blipFill`
  = invalid OOXML (the reader matches by local name and would *hide* the bug — so
  the test must assert the literal `<a:blipFill` bytes, the Phase-67 lesson).
- **Cover-fit is computable without parsing pixels.** True "cover" (fill the box,
  no distortion, crop the overflow) needs the image aspect; `image.DecodeConfig`
  reads the dimension header only (not pixel data — §7/D-046; the chart composer
  already does this). A center-crop `<a:srcRect>` (integer thousandths-of-a-
  percent on the overflowing axis) realizes cover deterministically. An unreadable
  header falls back to a plain stretch (best effort).
- **Card wiring reuses the existing surface AddShape.** `renderCardChrome` already
  draws the surface rounded-rect with `WithRadius(RadiusLG) + WithFill`; appending
  `WithImageFill` when set replaces the fill while the `RadiusLG` geometry still
  clips the photo's corners. The scene resolves `Card.ImageFill` (an `AssetID`)
  via `r.resolve` up front; a miss warns and falls back to the solid fill.
- **Determinism: the Card becomes asset-bearing.** `nodeUsesAssets(Card)` must
  return true when `ImageFill != ""` so the slide composes in the sequential pass
  and the media part numbering stays deterministic (the existing asset-serial
  contract). `walkImages` is only for frame-ref validation (Image nodes) — no
  change needed.
- **Policy stays `HasAsset:false`.** `TestPolicy_MatchesStructs` ties `HasAsset`
  to a field literally named `AssetID`. The card still renders as native chrome
  (not a pic) — the photo is a *fill*, not the node — so the field is named
  `ImageFill` and `KindCard` keeps `HasAsset:false`. Consistent.
- **G6 is structural.** The emitted `<a:blipFill>` (+ `srcRect`) survives a write
  → `pptx.Open` → re-write; the round-trip test asserts it persists — the same
  structural proof used for duotone/framing (no new read accessor).

## 4. Recommendations

- Codec: `XShapeProperties.BlipFill`; `prefixFor` rule `blipFill@spPr → a`.
- Builder: `WithImageFill(src ImageSource)` + a `coverSrcRect(data, box)` helper
  (its own file with the image decoders).
- Scene: `Card.ImageFill AssetID`; `cardChrome.imageFill`; `renderCardChrome`
  appends `WithImageFill`; `renderCard` resolves + warns; `nodeUsesAssets(Card)`
  includes `ImageFill != ""`.
- Tests: builder emit (asserts literal `<a:blipFill`, not `p:`), cover-crop
  (wide → l/r, tall → t/b, match → none), nil byte-identical, round-trip; scene
  card resolves + emits, missing warns, empty byte-identical, determinism.
  THEME.md note, glossary, compose-a-scene skill, docs/site pptx.md. D-117.

## 5. Open questions

- Bento cell / column image fill → a follow-up (the `WithImageFill` mechanism is
  ready; the cell/column surfaces would route through it the same way).
- A scrim over an image-filled card (legibility for header text on the photo) →
  the caller composes `Card.Backdrop`/`HeaderFill` today; a card-level scrim knob
  is a possible follow-up.
