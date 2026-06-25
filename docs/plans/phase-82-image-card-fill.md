# Phase 82 — image-as-card-fill (cover-fit photo surface)

**Subsystem:** `pptx` (builder shape fill) + `scene` (Card) + `internal/ooxml`
**RFC sections:** §8.2/§8.3 (shape fills), §10.1 (backward-compat), §10.2 (degrade), §11 (asset path), §7 (no pixel parse)
**Deps:** Phase 81 (R14.1 part 1); briefs 64, 65.
**Status:** Done

---

## 1. Goal

Fill a Card surface with a cover-fit photo instead of a solid/gradient color —
the image-as-fill atom of the photographic class — via a general builder
mechanism (`WithImageFill`) plus the Card wiring. Additive: byte-identical when
unused.

## 2. Why now

Completes R14.1 (HIGH · both; engine half) after Phase 81's background scrim +
duotone. Image-filled cards are core to photo-driven decks (cover panels, product
tiles); they were the deferred clause of R14.1 (§4.3 split, D-116).

## 3. RFC sections implemented

- `RFC §8.2/§8.3` — a shape gains an image (blip) fill alongside solid/gradient.
- `RFC §10.1` — `ImageFill == ""` / nil `WithImageFill` is byte-identical.
- `RFC §10.2` — an unresolvable `ImageFill` warns and falls back; no panic.
- `RFC §11` — composes the existing `AssetResolver` + media-part path.
- `RFC §7` — cover-fit reads only the format-header dims (D-046), never pixels.

## 4. Brief findings incorporated

- `docs/research/65-image-card-fill.md` — *"a shape image fill is `<a:blipFill>`
  inside `spPr` — a new builder `Fill` capability"* → `WithImageFill` option +
  `XShapeProperties.BlipFill`.
- `65` — *"namespace gotcha: `blipFill` is `p:` in a pic but `a:` in `spPr`"* →
  `prefixFor` context rule + a test asserting the literal `<a:blipFill` bytes.
- `65` — *"cover-fit is computable without parsing pixels"* → `coverSrcRect` via
  `image.DecodeConfig` (center-crop `srcRect`, integer permille).
- `65` — *"determinism: the Card becomes asset-bearing"* → `nodeUsesAssets(Card)`
  includes `ImageFill != ""`.
- `65` — *"policy stays `HasAsset:false`"* → field named `ImageFill` (not
  `AssetID`); `KindCard` unchanged.

## 5. Findings I'm departing from

- **Bento cell / column image fill** deferred to a follow-up (the `WithImageFill`
  mechanism is general; the cell/column surfaces would route through it the same
  way). Card is the primary surface and satisfies R14.1's acceptance (an
  image-filled card cover-fits without distortion at any aspect). §4.3.

## 6. Decisions referenced

- `D-116` — R14.1 part 1 (scrim + duotone) + the split that queued this.
- `D-046` — reading the image dimension header is permitted (not pixel data).
- `D-026` — image fill is a caller-driven mechanism; the soul decides when a
  surface is a photo.
- `D-117` (new) — files the image-fill mechanism + the Card wiring.

## 7. Architecture

Codec: `XShapeProperties.BlipFill *XBlipFillProperties` (the shape fill choice);
`prefixFor` emits `a:blipFill` under `spPr` (vs the pic's `p:blipFill`).

Builder: `WithImageFill(src ImageSource)` ShapeOption → at `AddShape`, resolve the
source, `addImagePart`, set `BlipFill{Blip{embed}, SrcRect: coverSrcRect(bytes,
box), Stretch{fillRect}}`, and clear solid/gradient/noFill. `coverSrcRect` reads
the format-header dims and center-crops the overflowing axis (integer permille);
unreadable → nil (plain stretch). It wins over `WithFill`; the shape geometry
(roundRect) still clips the photo.

Scene: `Card.ImageFill AssetID`; `renderCard` resolves it via `r.resolve` (warn +
fall back on miss), passes `cardChrome.imageFill`; `renderCardChrome` appends
`WithImageFill` to the surface AddShape. `nodeUsesAssets(Card)` returns true when
`ImageFill != ""` so the slide renders serially (deterministic part numbering).

```text
Card{Header, ImageFill: "asset://photo", Fill: ColorSurface}
  renderCard → r.resolve → renderCardChrome:
    AddShape(roundRect, WithRadius(RadiusLG), WithFill(surface), WithImageFill(photo))
      → <p:spPr> … <a:blipFill><a:blip r:embed/><a:srcRect …/><a:stretch><a:fillRect/></a:blipFill>
Card{Header} (no ImageFill) → solid Fill (byte-identical)
```

## 8. Files added or changed

```text
internal/ooxml/slide/slide_types.go    # CHANGED — XShapeProperties.BlipFill
internal/ooxml/restorenamespaces.go     # CHANGED — blipFill@spPr → a:
pptx/shape.go                           # CHANGED — WithImageFill option + AddShape wiring
pptx/imagefill.go                       # NEW — coverSrcRect (cover-fit center-crop)
pptx/imagefill_test.go                  # NEW — emit / cover-crop / nil byte-identical / round-trip
scene/nodes.go                          # CHANGED — Card.ImageFill field
scene/render_card.go                    # CHANGED — cardChrome.imageFill + resolution + chrome wiring
scene/render.go                         # CHANGED — nodeUsesAssets(Card) includes ImageFill
scene/render_card_imagefill_test.go     # NEW — card resolves/emits, missing warns, empty byte-ident, determinism
scripts/smoke/phase-82.sh               # NEW — phase smoke
docs/research/65-image-card-fill.md     # NEW — brief
docs/research/INDEX.md                  # CHANGED — registers brief 65
docs/plans/phase-82-image-card-fill.md  # NEW — this plan
docs/plans/README.md                    # CHANGED — Phase 82 detail
docs/design/THEME.md                    # CHANGED — image-fill mechanism note
docs/glossary.md                        # CHANGED — image fill term
docs/decisions.md                       # CHANGED — adds D-117
docs/site/reference/pptx.md             # CHANGED — WithImageFill
skills/compose-a-scene/SKILL.md         # CHANGED — Card.ImageFill
```

## 9. Public API surface

```go
// pptx
func WithImageFill(src ImageSource) ShapeOption  // cover-fit image surface fill
// scene
// Card gains: ImageFill AssetID
```

Additive; no break.

## 10. Risks

- **R1 — invalid OOXML (`p:blipFill` on a shape).** **Mitigation:** the
  `prefixFor` context rule; a test asserts the literal `<a:blipFill` bytes and
  that no `<p:blipFill>` appears on the shape.
- **R2 — byte-identity.** **Mitigation:** `ImageFill == ""` / nil self-gates; a
  byte-identity test pins a plain card and a nil-source shape.
- **R3 — determinism.** **Mitigation:** `nodeUsesAssets(Card)` makes an
  image-filled card serial; a 1-vs-8-worker test asserts byte-identity.
- **R4 — distortion at non-matching aspect.** **Mitigation:** `coverSrcRect`
  center-crops from the header dims; the cover-crop test covers wide / tall /
  matching.

## 11. Acceptance criteria

1. A shape with `WithImageFill` emits `<a:blipFill>` (a:, not p:) replacing the
   solid fill; it survives write → `pptx.Open` → re-write (G6).
2. Cover-fit center-crops the overflowing axis: a wide image crops left/right, a
   tall image crops top/bottom, a matching aspect emits no `srcRect`.
3. A Card with `ImageFill` fills its surface with the photo, warning-free, the
   asset counted, and worker-count deterministic; an unresolvable ID warns and
   falls back to the solid fill.
4. A Card with `ImageFill == ""` is byte-identical to the pre-Phase-82 build.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | builder image-fill option + cover crop |
| `scene` | 80% | default |
| `internal/ooxml/slide` | 85% | codec struct field |

## 13. Smoke check

`scripts/smoke/phase-82.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `WithImageFill` / `coverSrcRect` / `XShapeProperties.BlipFill` / the
   namespace rule / `Card.ImageFill` present + the chrome + serial wiring.
3. `OK:` builder emit / cover-crop / nil byte-identical tests.
4. `OK:` scene card resolves / missing-warns / empty byte-identical / determinism.

## 14. Tests

- **Round-trip golden (`pptx`):** an image-filled shape survives write → reopen →
  re-write with `<a:blipFill>`; cover-crop wide/tall/match; a nil source is
  byte-identical.
- **Black-box (`scene_test`):** a Card with `ImageFill` resolves a photo and emits
  `<a:blipFill>` (no warnings, asset counted); an unresolvable ID warns; an empty
  `ImageFill` is byte-identical; the path is worker-count deterministic.
- **Integration / Fuzz:** no (no new node; the asset path is already covered).
