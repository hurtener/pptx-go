# Phase 101 — contrast-aware accent-text mechanism

**Subsystem:** `scene` (contrast utilities)
**RFC sections:** §7.1 (color roles), §13.3 (variants — per-variant legibility)
**Deps:** Phase 97 (`DarkColors.Text` consumes the derived per-variant accent);
brief 84
**Status:** Done

---

## 1. Goal

Expose a deterministic, hue-preserving WCAG contrast-nudge primitive
(`scene.LegibleTextOn`) so a soul can derive an accent text color that stays
legible on any surface, per variant — without the engine imposing it (byte-identical).

## 2. Why now

R8.6 in priority order after R8.7. It is MED; the engine already ships the
per-variant `TextAccent` override (`DarkColors.Text`, D-135, verified in Phase
100) and the WCAG luminance math (D-082), but lacks the **graded** contrast
primitive the soul's per-variant derivation needs — today the only tool is the
binary `accentLegible` + a white fallback
(`DECKARD-PRODUCT-REQUIREMENTS.md` R8.6). D-059 puts the primitive here; the
derivation policy (which accents, which canvas) is Deckard's `refine.go`.

## 3. RFC sections implemented

- `RFC §7.1` / `§13.3` — a reusable mechanism to keep accent text legible on any
  surface, computed per variant; the soul applies the result via the existing
  per-variant override.

## 4. Brief findings incorporated

- `docs/research/84-contrast-aware-accent-text.md` — *the engine already has the
  WCAG math; the missing atom is a graded nudge over it* → `LegibleTextOn` reuses
  `relLuminance` / `contrastRatioT10` / `darkSurfaceLumaMax`.
- `docs/research/84-contrast-aware-accent-text.md` — *it must be a pure caller
  mechanism, byte-identical (no auto-apply, D-026)* → the engine wires it into no
  render path; the existing `accentLegible` + `onCardSurface` fallback stays.
- `docs/research/84-contrast-aware-accent-text.md` — *hue-preserving nudge =
  blend toward black/white, direction from the background* → lighten on a dark
  surface, darken on a light one (the `onCardSurface` crossover).

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-082` — `onCardSurface` auto-contrast — `LegibleTextOn` is its graded color
  analog, sharing the same WCAG luminance math so the engine and soul agree.
- `D-135` — `DarkColors.Text` — where a soul stores the per-variant accent text
  derived with this primitive; consumed by the renderer (Phase 100).
- `D-026` / `D-059` — the engine exposes the primitive; the soul drives when to
  derive and applies the result. No auto-apply → byte-identical.
- New decision **D-139** filed in this PR.

## 7. Architecture

```text
scene/contrast.go
  + LegibleTextOn(fg, bg pptx.RGB, minRatioX10 int) pptx.RGB   // exported mechanism
      already legible (ratio >= min)  → fg unchanged
      else nudge hue-preserving:  bg dark → lighten toward white
                                  bg light → darken toward black
      integer steps until ratio cleared (or nearest endpoint)
      malformed fg/bg → fg unchanged (fail-safe)
  + channelLuminance / rgbFromChannels (helpers; reuse srgbLinear)
```

No render-path call → all existing output byte-identical.

## 8. Files added or changed

```text
scene/contrast.go                      # CHANGED — LegibleTextOn + channelLuminance + rgbFromChannels
scene/contrast_legible_test.go         # NEW — clears-target / already-legible / pure / malformed / hue / target-honored
scripts/smoke/phase-101.sh             # NEW — phase smoke
docs/research/84-contrast-aware-accent-text.md  # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 84
docs/plans/phase-101-contrast-aware-accent-text.md  # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 15 phase entry
docs/decisions.md                      # CHANGED — adds D-139
docs/glossary.md                       # CHANGED — LegibleTextOn / contrast-aware accent text
docs/design/THEME.md                   # CHANGED — accent-text legibility mechanism note
docs/site/reference/scene.md           # CHANGED — LegibleTextOn surface
skills/define-a-theme/SKILL.md         # CHANGED — derive per-variant accent text snippet
```

## 9. Public API surface

```go
// scene
func LegibleTextOn(fg, bg pptx.RGB, minRatioX10 int) pptx.RGB
```

Additive; no prior surface breaks.

## 10. Risks

- **R1 — accidental render-path coupling** — if a renderer called the helper,
  output would change. **Mitigation:** the helper is wired into no render path;
  the full existing scene suite (the goldens) passes unchanged — the byte-identity
  proof.
- **R2 — non-determinism** — **Mitigation:** integer steps over the precomputed
  `srgbLinear` table; a purity test asserts same-inputs → same-output.

## 11. Acceptance criteria

1. An accent that fails the target on a dark surface lightens until it clears it;
   one that fails on a light surface darkens until it clears it; the two
   derivations of the same accent differ.
2. An accent already clearing the ratio is returned unchanged; a malformed input
   returns unchanged (fail-safe).
3. The function is pure (same inputs → same hex); the darken path preserves hue.
4. A looser large-text target (3:1) nudges no further than the body target (4.5:1).
5. The full existing scene suite passes unchanged (no auto-apply → byte-identical).

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-101.sh` verifies `LegibleTextOn` exists and the acceptance
tests pass.

## 14. Tests

- **Unit:** `scene` white-box (reuses the package contrast math to assert ratios):
  clears-target per background, already-legible unchanged, pure, malformed safe,
  hue-preserving darken, target honored.
- **Round-trip golden:** the existing scene goldens are the byte-identity proof.
- **Integration / Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `LegibleTextOn` — the exported scene mechanism that nudges a foreground color,
  preserving hue, until it clears a target WCAG contrast ratio against a
  background; a caller tool, not auto-applied.

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-101.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-139).
- [x] Docs site / skill updated.
