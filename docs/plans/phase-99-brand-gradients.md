# Phase 99 — named brand gradients

**Subsystem:** `pptx` (Theme + gradient spec) + `scene` (Background resolution)
**RFC sections:** §7.1 (theme tokens), §10.2 (background rendering / degrade)
**Deps:** Phase 97 (D-135 slot-less theme field pattern); brief 82
**Status:** Done

---

## 1. Goal

A theme can register named brand gradients (e.g. `hero`, `heroDark`) as explicit
stop lists with an angle and a linear/radial flag, and a slide can request one by
name — so a brand's signature wash is reusable, deterministic, and correct on
light or dark variants; a deck that names no gradient is byte-identical.

## 2. Why now

Third phase of Wave 15. R8.5 is HIGH: the reference deck's deep-navy radial hero
wash is unreachable today because gradients are ad-hoc role pairs resolved
through the active theme, with no way to name, store, or reuse a brand gradient
(`DECKARD-PRODUCT-REQUIREMENTS.md` R8.5). D-059 puts the named-gradient mechanism
here; the soul defining the brand gradients from a brand source is Deckard's.

## 3. RFC sections implemented

- `RFC §7.1` — adds named brand gradients to the theme token surface.
- `RFC §10.2` — background rendering: a missing / invalid named gradient degrades
  to a `LayoutWarning` + skip, never a panic.

## 4. Brief findings incorporated

- `docs/research/82-brand-gradients.md` — *`pptx.GradientStop.Color` is already a
  `pptx.Color` (RGB or token), so a named gradient's stops are just
  `[]pptx.GradientStop` and variant handling is automatic* → `GradientSpec.Stops
  []GradientStop`; an RGB stop pins an exact hue across variants, a `TokenColor`
  stop follows the active/dark theme with no special-casing.
- `docs/research/82-brand-gradients.md` — *the named spec owns its linear/radial
  choice via `Radial bool`, so one `GradientName` field feeds either primitive* →
  `drawNamedGradient` picks `RadialGradient` vs `LinearGradient(spec.Angle)`.
- `docs/research/82-brand-gradients.md` — *resolution slots in front of the
  existing path, byte-identically; miss/invalid → warn + skip* → `GradientName !=
  ""` supersedes `Stops`/legacy in the `BackgroundGradient` case; empty name runs
  the existing path unchanged.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-135` / `D-136` — slot-less theme fields (dark palette, accent palette)
  consumed only by the scene renderer; the named-gradient map follows the same
  pattern (resolved fill round-trips, the map does not).
- `D-105` — multi-stop background gradient — the `backgroundGradientStops`
  invariant (2..8 ascending in [0,1]) the named spec's stops reuse.
- `D-026` / `D-059` — the engine stores + resolves named gradients; the soul
  defines them from a brand source (Deckard's half).
- New decision **D-137** filed in this PR.

## 7. Architecture

```text
pptx.GradientSpec{ Stops []GradientStop; Angle int; Radial bool }   // pptx/fill.go
pptx.Theme
  + Gradients map[string]GradientSpec
  + WithGradient(name, spec) ThemeOption
  + Gradient(name) (GradientSpec, bool)
  Clone()  // deep-copies the map + each spec's stop slice

scene.Background
  + GradientName string   // BackgroundGradient: names a theme gradient; supersedes Stops/legacy

scene drawBackgroundFill / case BackgroundGradient:
  if GradientName != "" → drawNamedGradient(name):
      spec, ok := theme.Gradient(name)          // miss → warn + skip
      validGradientStopPositions(spec.Stops)     // invalid → warn + skip
      spec.Radial ? RadialGradient(stops) : LinearGradient(spec.Angle, stops)
  else → existing Stops / legacy 2-role path (byte-identical)
```

## 8. Files added or changed

```text
pptx/fill.go                           # CHANGED — GradientSpec type
pptx/theme.go                          # CHANGED — Theme.Gradients, WithGradient, Gradient accessor, Clone deep-copy
pptx/theme_test.go                     # CHANGED — option + accessor + Clone + -race concurrent reuse
scene/background.go                    # CHANGED — Background.GradientName field
scene/render.go                        # CHANGED — drawNamedGradient + validGradientStopPositions; BackgroundGradient case
scene/render_gradient_named_test.go    # NEW — brand wash / linear vs radial / miss+invalid warns / empty byte-identical / determinism
scripts/smoke/phase-99.sh              # NEW — phase smoke
docs/research/82-brand-gradients.md    # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 82
docs/plans/phase-99-brand-gradients.md # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 15 phase entry
docs/decisions.md                      # CHANGED — adds D-137
docs/glossary.md                       # CHANGED — Brand gradient
docs/design/THEME.md                   # CHANGED — brand gradient mechanism
docs/site/reference/pptx.md            # CHANGED — GradientSpec / WithGradient surface
docs/site/reference/scene.md           # CHANGED — Background.GradientName (if Background is documented there)
skills/define-a-theme/SKILL.md         # CHANGED — named gradient section
skills/compose-a-scene/SKILL.md        # CHANGED — Background.GradientName field
```

## 9. Public API surface

```go
// pptx
type GradientSpec struct {
    Stops  []GradientStop // each Color is RGB (variant-independent) or TokenColor (follows theme)
    Angle  int            // linear angle, degrees CW from +x; ignored when Radial
    Radial bool
}
// Theme gains: Gradients map[string]GradientSpec
func WithGradient(name string, spec GradientSpec) ThemeOption
func (t *Theme) Gradient(name string) (GradientSpec, bool)

// scene
// Background gains: GradientName string
```

No prior public surface breaks (additive types/fields/options).

## 10. Risks

- **R1 — byte-identity regression** — the `BackgroundGradient` case must be
  unchanged when `GradientName == ""`. **Mitigation:** the named branch is gated
  on `GradientName != ""`; the existing gradient goldens pass, and a guard
  asserts a legacy gradient is byte-identical even when the theme registers an
  unused named gradient.
- **R2 — shallow clone aliases the map/stops** — **Mitigation:** `Clone()`
  deep-copies the map and each spec's stop slice; a `-race` concurrent-reuse test
  proves it.
- **R3 — invalid / missing named gradient panics** — **Mitigation:** miss and
  invalid-stops both `r.warn` + return false (skip the fill), proven by tests.

## 11. Acceptance criteria

1. A named radial brand gradient renders its exact RGB stop hues into the slide
   as a circular `gradFill`; a non-radial named gradient renders as `<a:lin>` at
   the spec's angle.
2. Requesting an unregistered name, or a spec with invalid stops, records a
   `LayoutWarning` and skips the fill (no panic).
3. A gradient background with no `GradientName` is byte-identical whether or not
   the theme registers unrelated named gradients.
4. `Theme.Clone()` deep-copies `Gradients` (race-safe); a named-gradient deck is
   byte-identical across worker counts.
5. `make coverage` shows `pptx` and `scene` ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default |
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-99.sh` verifies:

1. `OK:` `pptx.GradientSpec` + `Theme.Gradients` + `WithGradient` exist.
2. `OK:` `Background.GradientName` exists (grep `scene/background.go`).
3. `OK:` `drawNamedGradient` resolves the named spec (grep `scene/render.go`).
4. `OK:` the brand-wash / warn / byte-identical / determinism tests pass.

## 14. Tests

- **Unit:** `pptx` (option + accessor + Clone + `-race`).
- **Round-trip golden:** the existing gradient goldens pin the no-name path; the
  new black-box tests assert the named radial/linear fills reach the bytes,
  miss/invalid warn, and the deck is deterministic.
- **Integration:** no new cross-subsystem seam.
- **Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `Brand gradient` — a named gradient registered on a theme
  (`Theme.Gradients` / `WithGradient`) and requested by a scene
  `Background.GradientName`; resolved to a linear/radial fill, slot-less.

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-99.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-137).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (define-a-theme, compose-a-scene).
