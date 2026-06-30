# Brief 84 — Contrast-aware accent-text mechanism (R8.6)

> Informs Phase 101 (Wave 15 — theme/soul engine bits). Engine side of the
> `both`-tagged requirement R8.6 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
> MED · both; D-059 puts the contrast-nudge *mechanism* here, the soul's
> per-variant accent-text *derivation* in Deckard's `refine.go`).

## 1. Motivating phase

R8.6 wants any brand accent used as emphasis text to stay legible on whatever
surface it sits on (jade on navy, deeper jade on cream), derived per variant.
Deckard's `derivedTextAccent` (refine.go) does a crude 0.78 channel scale that is
neither perceptual nor contrast-checked, and only for the light theme. The engine
already ships per-variant `TextAccent` override (`DarkColors.Text`, D-135) and a
binary `accentLegible` check, but it has **no deterministic primitive that nudges
an accent color until it clears a target contrast ratio**. Phase 101 adds that
mechanism so the soul (or any caller) can derive a legible accent text color
reproducibly.

## 2. Subsystem / files

- `scene/contrast.go` — the WCAG machinery: `relLuminance`, `contrastRatioT10`,
  the precomputed `srgbLinear` table, `accentMinContrastT10`, `parseHexRGB`,
  `onCardSurface` / `onSurfaceRGB`, `accentLegible`. The new helper reuses all of
  it.
- `pptx/theme.go` — `DarkColors.Text` (D-135) is where a soul stores the derived
  per-variant accent text; the engine consumes it (already wired, Phase 100).

## 3. Findings

- **The engine already has the WCAG contrast math** (D-082): integer
  `relLuminance` (precomputed sRGB-linear table) + `contrastRatioT10` + the
  black/white crossover `darkSurfaceLumaMax`. A contrast-nudge helper is a thin
  layer over these — no new math, fully deterministic (integer steps).
- **The missing atom is a graded nudge, not the binary check.** `accentLegible`
  only answers pass/fail, and the eyebrow falls back to white (`onCardSurface`)
  when an accent fails. R8.6 wants the accent *moved* — lightened on a dark
  surface, darkened on a light one — until it clears the ratio, preserving hue.
  That graded primitive is what's absent.
- **Hue-preserving nudge = blend toward black/white.** Scaling all channels
  toward black (`c·(1−k)`) preserves hue exactly; blending toward white
  (`c + (255−c)·k`) preserves the dominant hue while desaturating — the standard,
  acceptable "preserve hue" for legibility. Direction is chosen from the
  background: a dark surface (`relLuminance(bg) < darkSurfaceLumaMax`) lightens
  the accent toward white, a light surface darkens toward black — moving the
  accent *away* from the background's luminance to maximize contrast.
- **It must be a pure caller mechanism, byte-identical (D-026).** The engine
  should **not** auto-apply the nudge in the render path — that would change the
  existing R11.2 eyebrow fallback (white → nudged accent) and break goldens, and
  it would impose taste. The engine exposes the primitive; the soul decides which
  accents to derive (per variant) and applies the result via
  `DarkColors.Text[TextAccent]` (already consumed, Phase 100). So no render path
  changes → byte-identical.
- **Already-legible returns unchanged; malformed fails safe.** If the accent
  already clears the ratio, return it verbatim (so a caller that runs everything
  through the helper gets byte-identical results for the common case). A
  malformed hex returns the input unchanged (fail-safe, mirroring
  `relLuminance`'s malformed → 100000 convention).

## 4. Recommendations

- Add a pure exported `scene.LegibleTextOn(fg, bg pptx.RGB, minRatioX10 int)
  pptx.RGB` in `scene/contrast.go`: returns `fg` if it already clears
  `minRatioX10` (×10; e.g. 45 = 4.5:1) against `bg`, else nudges `fg`
  hue-preserving (toward white on a dark `bg`, toward black on a light `bg`) in
  fixed integer steps until it clears the ratio or reaches the nearest endpoint;
  malformed input → `fg` unchanged. Reuses `relLuminance` / `contrastRatioT10` /
  `darkSurfaceLumaMax`.
- Document it as a **mechanism** (D-026), the graded color analog of
  `onCardSurface`: the engine computes a legible accent deterministically; the
  soul drives *when* to use it and stores the per-variant result on the theme via
  `WithDarkText`. No auto-apply, no render-path change → byte-identical.
- Tests: clears the target on a navy bg (lightened) and a cream bg (darkened),
  the two results differ and each meets the ratio; an already-legible accent
  returns unchanged; the function is pure (same inputs → same hex); a malformed
  input returns unchanged.
- Light §19: a THEME.md note + define-a-theme skill snippet showing a soul
  deriving a per-variant legible `TextAccent` with `LegibleTextOn` + `WithDarkText`.

## 5. Open questions

- Should the engine auto-apply contrast-aware accent text on default decks? No —
  that imposes taste and breaks byte-identity (D-026); the existing
  `accentLegible` + `onCardSurface` fallback stays. The nudge is a caller tool.
- Should the helper live in `pptx` (lower layer) instead of `scene`? No — the
  WCAG machinery already lives in `scene/contrast.go`; exposing it there keeps a
  single source of truth that the engine's own `onCardSurface` shares. Deckard
  imports `scene` already.
- Per-variant *derivation* (which accents, against which canvas) is the soul's
  product half (D-059) — the engine ships the primitive only.
