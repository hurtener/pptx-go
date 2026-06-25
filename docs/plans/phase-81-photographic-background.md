# Phase 81 — photographic-imagery background class (scrim + duotone)

**Subsystem:** `scene` (Background) + `pptx` (builder Image)
**RFC sections:** §10.1 (backward-compat), §10.2 (degrade to warning), §11 (asset path), §7.1 (token colors)
**Deps:** Wave-13 background work (D-104..D-114); brief 64.
**Status:** Done

---

## 1. Goal

Reach the photographic slide class: add a slide-background **scrim** (a
darkening/tinting overlay that guarantees text legibility over a photo or busy
background) and a **duotone** two-tone recolor of a photographic background, both
additive and byte-identical when unused.

## 2. Why now

Wave 14 generalizes beyond the one sample deck (`docs/plans/README.md`); the
photo-driven class — full-bleed covers with text overlays, on-brand-tinted
imagery — is the most common professional class still unreachable. Engine req
R14.1 (HIGH · both; engine half per D-059). The full-bleed photo background
itself already exists (`BackgroundAsset`); this phase adds the two missing atoms.

## 3. RFC sections implemented

- `RFC §10.1` — a nil `Scrim` / nil `Duotone` is byte-identical.
- `RFC §10.2` — the asset/duotone path degrades to a warning, never panics.
- `RFC §11` — composes the existing `AssetResolver` background pic path.
- `RFC §7.1` — scrim/duotone colors are theme tokens (P2).

## 4. Brief findings incorporated

- `docs/research/64-photographic-background.md` — *"a scrim is a general overlay,
  not a kind"* → `Background.Scrim *Scrim` drawn after any successful base fill.
- `64` — *"`renderBackground` needs a tiny refactor"* → `drawBackgroundFill`
  returns whether a fill was drawn; scrim overlays only then. Fill emission
  unchanged → byte-identical.
- `64` — *"duotone is a real OOXML blip effect → a builder addition (P1)"* →
  `(*Image).SetDuotone` + `XBlip.Duotone`; `duotone` registered in
  `restorenamespaces`.
- `64` — *"G6 is structural + a read accessor"* → `(*Image).Duotone()` accessor +
  write → reopen → re-write test.
- `64` — *"image-as-card-fill is a distinct, larger mechanism"* → deferred to
  Phase 82 (§5 departure).

## 5. Findings I'm departing from

- **Image-as-card/cell/column fill** (R14.1 spec clause b) is split out to a
  follow-up phase (**Phase 82**, R14.1 cont.): it needs a `blipFill`-on-shape
  builder `Fill` and asset wiring into the card chrome path — cleanly separable
  from the slide-background class. Documented as a §4.3 split (D-116).
- **Uniform cover-fit** (no distortion at any aspect) needs the image's pixel
  dimensions, which §7 forbids parsing; the `Fit` type already defers aspect-aware
  cover/contain to V1.x. A matching-aspect full-bleed photo is exact under the
  existing stretch; uniform cover stays deferred.

## 6. Decisions referenced

- `D-059` — engine extension; engine half of a `both` req.
- `D-026` — engine exposes a mechanism (scrim color/opacity, duotone roles); the
  soul applies it to meet a contrast target.
- `D-114` — the thin `*Image` setter pattern (`SetCornerRadius`/`SetElevation`).
- `D-061` — the `restorenamespaces` registration gotcha for new elements.
- `D-116` (new) — files the scrim + duotone atoms and the R14.1 split.

## 7. Architecture

Scene: `Background.Scrim *Scrim{Color ColorRole; Opacity int; Gradient bool;
GradientAngle int}` + `Background.Duotone *Duotone{Shadow, Highlight ColorRole}`.
`renderBackground` → `drawBackgroundFill` (the existing switch, now returning
`drawn bool`); when a fill was drawn and `Scrim != nil`, `renderScrim` overlays a
full-slide rect — `SolidFill(TokenColorAlpha(Color, Opacity))` or, for `Gradient`,
`LinearGradient(angle, {0: Color@0}, {1: Color@Opacity})` (angle 0 → 90°). In the
`BackgroundAsset` case, a non-nil `Duotone` calls `img.SetDuotone(TokenColor(
Shadow), TokenColor(Highlight))`.

Builder: `(*Image).SetDuotone(shadow, highlight Color)` resolves both against the
active theme and sets `XBlip.Duotone = <a:duotone><a:srgbClr/><a:srgbClr/>`; a nil
tone is a no-op (byte-identical). `(*Image).Duotone() (shadow, highlight RGB, ok)`
is the read inverse. `"duotone"` registered in `restorenamespaces`.

```text
Background{Kind: BackgroundAsset, AssetID, Duotone:{Accent,Canvas}, Scrim:{Surface, 50000, Gradient}}
  → pic (full bleed) + <a:blip><a:duotone><a:srgbClr val=accent/><a:srgbClr val=FFFFFF/>
  → scrim rect: <a:gradFill> transparent → surface@50000
Background{Kind: BackgroundColor, Color} (no Scrim/Duotone) → unchanged (byte-identical)
```

## 8. Files added or changed

```text
internal/ooxml/slide/slide_types.go   # CHANGED — XBlip.Duotone + XDuotone struct
internal/ooxml/restorenamespaces.go    # CHANGED — register "duotone": "a"
pptx/media.go                          # CHANGED — (*Image).SetDuotone + Duotone()
pptx/media_duotone_test.go             # NEW — duotone round-trip / token / nil byte-identical
scene/background.go                    # CHANGED — Scrim, Duotone types + Background fields
scene/render.go                        # CHANGED — drawBackgroundFill + renderScrim + duotone wiring
scene/render_background_photo_test.go  # NEW — scrim solid/gradient, photo class, nil byte-ident, determinism
scene/render_adversarial_test.go       # CHANGED — scrim slide in the torture fixture
scripts/smoke/phase-81.sh              # NEW — phase smoke
docs/research/64-photographic-background.md  # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 64
docs/plans/phase-81-photographic-background.md  # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 14 section + Phase 81 detail
docs/design/THEME.md                   # CHANGED — scrim + duotone mechanism notes
docs/glossary.md                       # CHANGED — scrim, duotone terms
docs/decisions.md                      # CHANGED — adds D-116
docs/site/reference/scene.md           # CHANGED — Background.Scrim / Duotone
docs/site/reference/pptx.md            # CHANGED — Image.SetDuotone / Duotone
skills/compose-a-scene/SKILL.md        # CHANGED — Background scrim/duotone
```

## 9. Public API surface

```go
// pptx
func (im *Image) SetDuotone(shadow, highlight Color) *Image
func (im *Image) Duotone() (shadow, highlight RGB, ok bool)
// scene
type Scrim struct { Color pptx.ColorRole; Opacity int; Gradient bool; GradientAngle int }
type Duotone struct { Shadow, Highlight pptx.ColorRole }
// Background gains: Scrim *Scrim; Duotone *Duotone
```

Additive; no break.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** nil Scrim/Duotone self-gate; a
  byte-identity test pins a plain color background and a nil-tone image.
- **R2 — invalid OOXML (bare `duotone`).** **Mitigation:** registered in
  `restorenamespaces`; the round-trip test reopens + re-writes and asserts the
  `<a:duotone>` + both colors survive.
- **R3 — determinism.** **Mitigation:** the scrim+duotone photo path is asserted
  byte-identical across 1 vs 8 workers.

## 11. Acceptance criteria

1. A solid scrim over a background draws a second full-slide rect carrying the
   scrim alpha; a gradient scrim emits a linear-gradient overlay.
2. A full-bleed photo background with a duotone emits `<a:duotone>` with both
   resolved colors and survives write → `pptx.Open` → re-write (G6).
3. A scrim+duotone photo with text records no warnings and is byte-identical
   across worker counts.
4. A background with a nil Scrim + nil Duotone is byte-identical to the
   pre-Phase-81 build.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | builder duotone method |
| `scene` | 80% | default |
| `internal/ooxml/slide` | 85% | codec struct |

## 13. Smoke check

`scripts/smoke/phase-81.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `(*Image).SetDuotone` / `Duotone()` present; `XDuotone` + `duotone` registered.
3. `OK:` scene `Background.Scrim` / `Duotone` fields + `renderScrim`.
4. `OK:` builder duotone round-trip / token-resolve / nil byte-identical tests.
5. `OK:` scene solid/gradient scrim, photo class, nil byte-identical, determinism tests.

## 14. Tests

- **Round-trip golden (`pptx`):** a duotone image survives write → reopen →
  re-write with `<a:duotone>` + both colors; token colors resolve to the active
  palette; a nil tone is byte-identical.
- **Black-box (`scene_test`):** solid + gradient scrim overlays; a scrim+duotone
  photo (asset resolver) emits the pic + `<a:duotone>` with no warnings; a nil
  Scrim is byte-identical; the path is worker-count deterministic.
- **Adversarial:** a gradient scrim slide added to the torture fixture.
- **Integration / Fuzz:** no (no new node; the asset path is already integration-
  covered).
