# Brief 64 — Photographic background class: scrim + duotone (R14.1)

> Informs Phase 81 (Wave 14). Engine req R14.1
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · both — engine half; D-059).

## 1. Motivating phase

Most professional brand / investor / marketing decks lead with full-bleed
photography: image-with-text-overlay covers and image-filled cards. The engine
already places a full-bleed picture background (`BackgroundAsset`, D-026 asset
path) but offers no **scrim** (a darkening/tinting overlay that guarantees text
legibility over a busy photo) and no **duotone** (a two-tone brand recolor of a
photo). Phase 81 adds those two engine atoms — the photographic class's
legibility + on-brand-tint mechanisms — building directly on the Wave-13
background work.

## 2. Subsystem / files

- `scene/background.go` — the `Background` struct + the existing background kinds.
- `scene/render.go` — `renderBackground` (the `BackgroundAsset` pic path +
  `backgroundGradientStopsFor`).
- `pptx/media.go` — `*Image` (the picture handle) + its `Set*` methods.
- `internal/ooxml/slide/slide_types.go` — `XBlip` (carries the optional
  `<a:alphaModFix>` already; the duotone effect is a sibling).
- `internal/ooxml/restorenamespaces.go` — the bare-element → `a:` prefix map.

## 3. Findings

- **The asset background already exists.** `BackgroundAsset` resolves an `AssetID`
  through the `AssetResolver` and draws a full-bleed pic. So the photographic
  *background* is reachable today; the gaps are the overlay and the recolor.
- **A scrim is a general overlay, not a kind.** Drawing a full-slide rect *after*
  the base fill — solid `SolidFill(TokenColorAlpha(color, opacity))` or a
  transparent→color `LinearGradient` — is a pure, deterministic mechanism that
  works over *any* drawn background (color, gradient, mesh, photo). Modeling it as
  `Background.Scrim *Scrim` (nil = no overlay) keeps it byte-identical when unused
  and avoids a kind explosion. The engine draws it; the soul picks color/opacity
  to meet its contrast target (D-026 — mechanism, not taste).
- **`renderBackground` needs a tiny refactor for "draw the scrim after a
  successful fill".** Splitting the switch into a `drawBackgroundFill` that
  returns whether a fill was drawn lets the caller overlay the scrim only when
  there is something to darken (never over an empty light slide). The fill
  emission is unchanged → byte-identical.
- **Duotone is a real OOXML blip effect → a builder addition (P1).**
  `<a:blip><a:duotone><a:srgbClr/><a:srgbClr/></a:duotone></a:blip>` recolors a
  picture's shadows → first color, highlights → second. `XBlip` already carries
  `<a:alphaModFix>`; adding an optional `Duotone *XDuotone` sibling + a
  `(*Image).SetDuotone(shadow, highlight Color)` is the same thin-wrapper shape as
  D-114's `SetCornerRadius`/`SetElevation`. The colors are resolved against the
  active theme at call time (P2) and emitted as literal `srgbClr` (the token
  pattern the fill path already uses).
- **`duotone` is a new element → register it.** `srgbClr`/`blip`/`alphaModFix`
  are already in the `restorenamespaces` map; `duotone` must be added or it emits
  bare (invalid OOXML). The children `srgbClr` are already registered. (Same
  gotcha as D-061's `lnSpc`.)
- **G6 is structural + a read accessor.** The emitted `<a:duotone>` survives a
  write → `pptx.Open` → re-write losslessly; a `(*Image).Duotone() (shadow,
  highlight RGB, ok bool)` read accessor proves the inverse and pins it.
- **Cover-fit (uniform) needs pixel dims we don't parse.** True uniform cover (no
  distortion at any aspect) needs the image's dimensions; §7 forbids parsing pixel
  data and the `Fit` type already defers aspect-aware cover/contain to V1.x. A
  full-bleed photo on a matching-aspect slide is exact under the existing stretch;
  uniform cover stays deferred — consistent with the existing `Fit` posture.
- **Image-as-card-fill is a distinct, larger mechanism.** Filling a *card*
  surface with a photo needs a `blipFill` inside a shape's `spPr` (a new builder
  `Fill`), not a separate pic, plus asset resolution wired into the card chrome
  path. It is cleanly separable from the slide-background class and is queued as a
  follow-up phase (Phase 82) under R14.1 (§4.3 split).

## 4. Recommendations

- Scene: `Background.Scrim *Scrim{Color ColorRole; Opacity int; Gradient bool;
  GradientAngle int}` + `Background.Duotone *Duotone{Shadow, Highlight ColorRole}`;
  `renderScrim` overlay + duotone applied in the `BackgroundAsset` case.
- Builder: `(*Image).SetDuotone(shadow, highlight Color)` + `Duotone()` accessor;
  `XBlip.Duotone *XDuotone`; register `"duotone": "a"`.
- Tests: builder duotone round-trip + token-resolve + nil byte-identical; scene
  solid/gradient scrim + scrim+duotone photo (asset resolver) + nil byte-identical
  + worker-count determinism; extend the adversarial fixture with a scrim slide.
  THEME.md scrim/duotone mechanism notes (colors are tokens — P2), glossary,
  compose-a-scene skill, docs/site. D-116.

## 5. Open questions

- Image-as-card/cell/column fill → Phase 82 (R14.1 cont.; `blipFill`-on-shape).
- Uniform cover-fit (aspect-aware) → V1.x with the `Fit` work (needs dims, §7).
- A non-centered scrim gradient focal / radial scrim → caller can already
  approximate with `GradientAngle`; richer focal control deferred.
