# Phase 73 — decoration color role

**Subsystem:** `scene` (Layer 2 — decoration) + `assets/ornaments` (recipes)
**RFC sections:** §14.2 (ornaments), §7.1 (token color), §10.1 (backward-compat)
**Deps:** none; brief 56.
**Status:** Done

---

## 1. Goal

Let a `Decoration` choose its color (any surface role) instead of being locked
to the accent, so textures/glows/shapes can be neutral grey, inverse-white, or
any brand role — defaulting to accent (byte-identical).

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); R13.5 is the CRITICAL unblock —
without it the engine can only render accent-colored decoration, so neutral
paper grain and white starfields (the reference's look) are unreachable, and
later R13 reqs (starfield R13.6, mesh R13.4, watermark R13.9) all need a
caller-chosen decoration color. Engine req R13.5 (CRITICAL · engine; D-059).

## 3. RFC sections implemented

- `RFC §14.2` — ornament recipes gain a color parameter.
- `RFC §7.1` — decoration color is a surface token role (P2).
- `RFC §10.1` — unset color = accent = byte-identical.

## 4. Brief findings incorporated

- `docs/research/56-decoration-color-role.md` — *"the recipe color is
  centralized in `accent()`/glow stops"* → thread one `role` through; rename
  `accent` → `roleFill`.
- `56` — *"the `Recipe` signature must change (a v0.x public break)"* → add a
  trailing `role pptx.ColorRole`; `OrnamentRecipe` alias follows; CHANGELOG note.
- `56` — *"`Decoration.Color` must be a pointer (zero ColorRole = ColorCanvas)"*
  → `Color *pptx.ColorRole`, nil = `ColorAccent` (D-054 pattern); §4.3 deviation.
- `56` — *"byte-identity via nil → ColorAccent"* → `render_decoration` resolves
  the role and passes it; unset = today.

## 5. Findings I'm departing from

- `56` notes the spec's literal `Decoration.Color pptx.ColorRole (zero =
  ColorAccent)`. **Departing because** `ColorRole`'s zero is `ColorCanvas`, a
  real color — a value field could not mean "accent by default" without
  remapping zero, which would make `ColorCanvas` decoration impossible. Use
  `*pptx.ColorRole` (nil = accent), the same resolution D-054 reached.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-054` — the `*ColorRole` nil-default pattern.
- `D-041`/`D-005` — ornament recipe mechanism.
- `D-107` (new) — files the decoration color role + the `Recipe` signature break.

## 7. Architecture

The `ornaments.Recipe` gains a trailing `role pptx.ColorRole`; the `accent`
helper becomes `roleFill(role, alpha)` and every curated recipe fills with the
supplied role (solid fills and glow gradient stops). `Decoration.Color
*pptx.ColorRole` (nil = `ColorAccent`) is resolved in `render_decoration` and
threaded to the recipe. `scene.OrnamentRecipe` (the alias) and
`WithOrnamentExtension` follow the new signature.

```text
Decoration{Preset: "grid_dots", Color: &white}  // white dot grid
  render_decoration: role = *Color (else ColorAccent)
  recipe(ps, box, alpha, rotation, role) → roleFill(role, alpha) → srgbClr=white
Decoration{Preset: "grid_dots"}  (Color nil) → role = ColorAccent → byte-identical
```

## 8. Files added or changed

```text
assets/ornaments/ornaments.go    # CHANGED — accent → roleFill(role, alpha)
assets/ornaments/glows.go        # CHANGED — RadialGlow/GlowRing take role; stops use it
assets/ornaments/patterns.go     # CHANGED — GridDots/NoiseOverlay take role
assets/ornaments/accents.go      # CHANGED — CornerBracket/ChevronArrow take role
assets/ornaments/*_test.go       # CHANGED — call recipes with a role arg
scene/ornaments/registry.go      # CHANGED — Recipe signature + doc
scene/ornaments/registry_test.go # CHANGED — test recipes with role
scene/nodes.go                   # CHANGED — Decoration.Color *pptx.ColorRole
scene/render_decoration.go       # CHANGED — resolve role, pass to recipe
scene/render_decoration_test.go  # CHANGED/NEW — color override + byte-identical default
scripts/smoke/phase-73.sh        # NEW — phase smoke
docs/research/56-decoration-color-role.md  # NEW — brief
docs/research/INDEX.md           # CHANGED — registers brief 56
docs/plans/phase-73-decoration-color-role.md  # NEW — this plan
docs/plans/README.md             # CHANGED — Phase 73 detail
docs/design/THEME.md             # CHANGED — decoration color mechanism note
docs/glossary.md                 # CHANGED — decoration color term
docs/decisions.md                # CHANGED — adds D-107
CHANGELOG.md                     # CHANGED — OrnamentRecipe signature break
docs/site/reference/scene.md     # CHANGED — Decoration.Color + OrnamentRecipe
skills/compose-a-scene/SKILL.md  # CHANGED — decoration color
```

## 9. Public API surface

```go
// scene/ornaments (and the scene.OrnamentRecipe alias)
type Recipe func(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64, role pptx.ColorRole) int // +role
// scene
// Decoration gains: Color *pptx.ColorRole // nil = ColorAccent
```

**Break:** `OrnamentRecipe`/`Recipe` gains a parameter — caller ornament
extensions must add the `role` arg (they may ignore it). v0.x; CHANGELOG noted.

## 10. Risks

- **R1 — byte-identity drift.** **Mitigation:** nil `Color` → `ColorAccent` →
  `roleFill(ColorAccent, alpha)` == old `accent(alpha)`; a byte-identity test
  over all 6 presets pins it.
- **R2 — caller-extension break surprises.** **Mitigation:** documented in
  CHANGELOG + D-107; the positional param is mechanical to adopt.

## 11. Acceptance criteria

1. A decoration with `Color = &ColorTextInverse` (white) emits a different `srgbClr` than the accent default for the same preset.
2. A decoration with `Color == nil` is byte-identical to the pre-Phase-73 build, for every curated preset.
3. The `role` threads through all 6 curated recipes (solid + glow).
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |
| `assets/ornaments` | 100% | curated recipes stay fully covered |

## 13. Smoke check

`scripts/smoke/phase-73.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `Recipe` signature carries `role pptx.ColorRole`.
3. `OK:` `roleFill` helper + `Decoration.Color` field.
4. `OK:` color-override test passes (white ≠ accent).
5. `OK:` nil-color byte-identical test passes.

## 14. Tests

- **Unit (assets/ornaments):** recipes called with a role render shapes
  (existing tests adapted to the new signature).
- **Black-box (`scene_test`):** a white decoration's slide XML differs from the
  accent default; a nil-`Color` decoration is byte-identical per preset.
- **Determinism:** color is a pure fill role; covered by byte-identity.
- **Integration / Fuzz:** no.
