# Brief 17 — type-detail-tokens

**Subsystem:** pptx — Layer 1 builder (theme/typography)
**Authored:** 2026-06-21
**Motivating phase:** Phase 30 — letter-spacing (tracking) token

## 1. Question

Pro decks open up eyebrows/labels with wide **tracking** (letter-spacing) and
tighten display headlines with slight negative tracking — the single biggest
"designed vs default" tell. The engine has no tracking anywhere: `pptx.FontSpec`
is `{Family, Size, Weight, Italic}`, `RunStyle` carries none, and the run-property
writer (`pptx/text_layout.go::toProps`) never emits the OOXML `a:rPr/@spc`
attribute. How can a soul set tracking per type role (and a caller override it per
run) — additively, deterministically, byte-identical when unused, and round-trip
clean (G6)?

This is `DECKARD-PRODUCT-REQUIREMENTS.md` R9.3 (`tracking-letterspacing-token`,
HIGH · engine), the first engine unit of the R8–R14 professional-bar work (D-059).

## 2. Prior art surveyed

- **`pptx/text_layout.go::toProps`.** Already maps `RunStyle` + the resolved
  `FontSpec` to `slide.XTextProperties` (the `a:rPr`), emitting size, family,
  bold/italic/underline/strike/baseline/color. Tracking is one more attribute on
  the same struct — a localized addition.
- **OOXML `a:rPr/@spc` (`ST_TextPoint`).** Signed letter-spacing in **1/100 of a
  point**. So tracking in points → `spc = round(points × 100)`. Mirrors how
  `sz` is already emitted (`Size × 100`).
- **The read model (`pptx/text.go` `*Run`).** Reconstructs run props from the
  parsed `rPr` via accessors (`FontSize`, `Bold`, `Baseline`, …). A `Tracking()`
  accessor reading the parsed `spc` is the read inverse — and adding `Spc` to
  `XTextProperties` makes the attribute survive parse + re-marshal (round-trip).
- **D-058 / `Stats.Colors` precedent.** A new resolved-typography value flows
  through the existing resolve path; the soul sets it per role and it reaches the
  bytes — the same shape this token follows via `FontSpec`.
- **P2 (tokens, not literals).** Tracking is a visual property added to the
  builder, so it needs a `docs/design/THEME.md` taxonomy entry in the same PR.

## 3. Findings

- **Tracking belongs on `FontSpec` (per role) first.** The dominant use is
  per-role (wide eyebrows, tight display), set by the resolved type scale — so
  `FontSpec.Tracking float64` (points, signed; 0 = none) is the primary
  mechanism. An optional per-run `RunStyle.Tracking *float64` override (nil =
  inherit the role's value) covers the rare run-level case without forcing every
  run to carry a value.
- **Emit is one attribute, default-off.** In `toProps`, the effective tracking
  (run override else role value) emits `spc = round(pt × 100)` when non-zero;
  zero emits nothing → byte-identical to today.
- **Round-trip is a struct field + a reader.** Add `Spc int` to
  `XTextProperties` (`xml:"spc,attr,omitempty"`); a `*Run.Tracking() float64`
  reader returns `Spc / 100`. A tracked run written then reopened reports the
  same tracking (G6); an untracked run reports 0 and is byte-identical.
- **Deterministic.** `round(pt × 100)` is pure integer output from a pinned
  float; no map iteration, no worker dependence.
- **Scene-side per-run override is deferred.** The engine mechanism (role-level
  `FontSpec.Tracking` + `pptx.RunStyle` override) is what a soul drives; a
  `scene.RunStyle.Tracking` passthrough is a thin follow-on, not needed to
  express the reference's per-role tracked-caps system.

## 4. Recommendations

1. `pptx/theme.go`: `FontSpec.Tracking float64` (points; 0 = none).
2. `pptx/text.go`: `RunStyle.Tracking *float64` (nil = inherit role); a
   `*Run.Tracking()` read accessor.
3. `internal/ooxml/slide`: `XTextProperties.Spc int` (`spc` attr, omitempty).
4. `pptx/text_layout.go::toProps`: emit `spc = round(effectiveTracking × 100)`
   when non-zero.
5. `docs/design/THEME.md`: add the tracking token to the typography taxonomy.
6. Round-trip + determinism + zero-value-byte-identical tests.

## 5. Open questions

- **Unit choice.** Points (×100 → `spc`) is chosen over thousandths-of-em for a
  direct OOXML mapping and round-trip clarity; an em-relative convenience could be
  layered later.
- **Line-height (R9.4) and case (R9.11)** are sibling type-detail tokens but
  land in their own phases — line-height is *paragraph*-level (`a:lnSpc`) and
  also feeds the wrap estimator, and case transforms run text; each is a distinct
  code path. Tracking is the clean run-attribute starting point.
- **Per-face width metrics (R9.5)** interact with tracking (tracking widens
  advance) but the wrap estimator's tracking-awareness is deferred to that phase.
