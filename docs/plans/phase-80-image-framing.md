# Phase 80 — rounded + shadow image framing

**Subsystem:** `pptx` (builder Image) + `scene` (Image node)
**RFC sections:** §7.1 (radius/elevation tokens), §11.1 (image), §10.1 (backward-compat)
**Deps:** none; brief 63.
**Status:** Done

---

## 1. Goal

Add corner-radius and elevation (drop-shadow) options to the scene `Image` node,
resolved from the theme's radius/elevation tokens, so image finish matches the
rounded, elevated card finish — byte-identical when unset.

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); plain rectangular pics lose the
rounded/elevated finish the cards have. Engine req R13.11 (MED · engine; D-059).
Closes the non-product R13 set.

## 3. RFC sections implemented

- `RFC §7.1` — the image gains radius/elevation token options (P2).
- `RFC §11.1` — the `Image` node's picture framing.
- `RFC §10.1` — zero radius/elevation = byte-identical.

## 4. Brief findings incorporated

- `docs/research/63-image-framing.md` — *"`applyCornerRadius`/`applyShadow`
  already exist for the pic spPr type"* → thin `*Image` wrappers.
- `63` — *"self-gating zero values keep byte-identity"* → `renderImage` calls
  both setters; zero tokens no-op.
- `63` — *"G6 round-trip is structural"* → assert `roundRect`/`outerShdw` survive
  write → reopen → re-write; no new read accessor.
- `63` — *"DecorationAsset deferred"* → `Image` only this phase.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-043` — the elevation/shadow mechanism (`applyShadow`).
- `D-114` (new) — files the image framing methods + fields.

## 7. Architecture

Builder: `(*Image).SetCornerRadius(RadiusRole)` sets the pic geometry to
`roundRect` and calls `applyCornerRadius(spPr, ResolveRadius(role), picBox(spPr))`
(returns early for `RadiusNone` = rectangular); `(*Image).SetElevation(
ElevationRole)` calls `applyShadow(spPr, ResolveElevation(role))` (no-op for
`ElevationFlat`). A `picBox` helper reconstructs the box from the pic transform.
Scene: `Image.CornerRadius RadiusRole` + `Image.Elevation ElevationRole`;
`renderImage` calls both setters (self-gating at zero). No new OOXML.

```text
Image{AssetID: …, CornerRadius: RadiusMD, Elevation: ElevationRaised}
  renderImage → img.SetCornerRadius(RadiusMD) → <a:prstGeom prst="roundRect"><a:avLst><a:gd .../>
              → img.SetElevation(ElevationRaised) → <a:effectLst><a:outerShdw …>
Image{AssetID: …}  (zero tokens) → rectangular, no shadow (byte-identical)
```

## 8. Files added or changed

```text
pptx/media.go              # CHANGED — (*Image).SetCornerRadius / SetElevation + picBox
pptx/media_test.go         # CHANGED/NEW — image radius+elevation round-trip golden
scene/nodes.go             # CHANGED — Image.CornerRadius / Elevation fields
scene/render_image.go      # CHANGED — renderImage applies the two setters
scene/render_image_test.go # CHANGED/NEW — scene image framing emits roundRect+outerShdw; zero byte-identical
scripts/smoke/phase-80.sh  # NEW — phase smoke
docs/research/63-image-framing.md  # NEW — brief
docs/research/INDEX.md     # CHANGED — registers brief 63
docs/plans/phase-80-image-framing.md  # NEW — this plan
docs/plans/README.md       # CHANGED — Phase 80 detail
docs/design/THEME.md       # CHANGED — image framing mechanism note
docs/glossary.md           # CHANGED — image framing term
docs/decisions.md          # CHANGED — adds D-114
docs/site/reference/pptx.md  # CHANGED — Image.SetCornerRadius/SetElevation
docs/site/reference/scene.md # CHANGED — Image.CornerRadius/Elevation
skills/compose-a-scene/SKILL.md  # CHANGED — Image framing
skills/register-an-asset/SKILL.md  # CHANGED — Image framing options (if it documents Image)
```

## 9. Public API surface

```go
// pptx
func (im *Image) SetCornerRadius(role RadiusRole) *Image // roundRect-clip the picture
func (im *Image) SetElevation(role ElevationRole) *Image // drop shadow on the picture
// scene
// Image gains: CornerRadius RadiusRole; Elevation ElevationRole
```

Additive; no break.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** zero tokens self-gate (early return /
  flat no-op); a byte-identity test pins a plain image.
- **R2 — round-trip loss.** **Mitigation:** a write → reopen → re-write test
  asserts `roundRect` + `outerShdw` persist.

## 11. Acceptance criteria

1. A builder image with a radius + elevation token round-trips: after write → `pptx.Open` → re-write, the pic carries `prst="roundRect"` and `<a:outerShdw>`.
2. A scene `Image` with `CornerRadius` + `Elevation` set emits both on its pic.
3. A scene `Image` with zero radius/elevation is byte-identical to the pre-Phase-80 build.
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | builder image methods |
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-80.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `(*Image).SetCornerRadius` / `SetElevation` present.
3. `OK:` `Image.CornerRadius` / `Elevation` fields + render wiring.
4. `OK:` framing round-trip test.
5. `OK:` zero-token byte-identical test.

## 14. Tests

- **Round-trip golden (`pptx`):** an image with radius+elevation survives write →
  reopen → re-write with `roundRect` + `outerShdw`.
- **Black-box (`scene_test`):** a scene `Image` with the tokens emits both; a
  zero-token image byte-identical.
- **Integration / Fuzz:** no.
