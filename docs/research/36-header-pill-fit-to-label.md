# Brief 36 — header-pill-fit-to-label

**Subsystem:** scene — Layer 2 renderer (card chrome)
**Authored:** 2026-06-22
**Motivating phase:** Phase 53 — header-pill-fit-to-label (R11.5, HIGH · engine)

## 1. Question

A card header pill is drawn at a fixed `pillW = In(1.0)`; a label wider than 1.0"
(e.g. "CUSTOMIZABLE") wraps to two lines inside the rounded chip and overflows it
(recreation slide 5). How can the pill size to its label on a single line, for any
caller-supplied text, deterministically and without drifting from the header-width
reservation?

## 2. Prior art surveyed

- **`scene/render_card.go`** — the pill is drawn in `renderCardChrome` at
  `pillW := pptx.In(1.0)` (clamped to `innerW`), and `cardHeaderColumnWOf` reserves
  the *same* `In(1.0)` from the header text column. The two **must** stay equal or
  the reserved and drawn widths drift.
- **`scene/metrics.go naturalWidth`** — the deterministic one-line text-width
  estimator (pinned char-width model, integer EMU). Already used by the wrapped-
  header geometry.
- **`scene/metrics.go fitScale` + `pptx.RunStyle.FontScale`** (R10.5/D-074) — the
  shrink-to-fit primitive: `fitScale(natW, boxW)` returns a font multiplier (0 when
  it already fits) quantized + floored, emitted as a reduced `@sz`.
- DECKARD R11.5 spec: `pillW = naturalWidth(label @ TypeCaption) + 2·pillPadX`,
  clamped to `innerW`; single-line / no-wrap (or ellipsize); the reserve-space math
  already consumes the new width.

## 3. Findings

- **One shared width function.** Extract `cardPillWidthOf(theme, pill, innerW) =
  clamp(naturalWidth(pill @ TypeCaption) + 2·cardPillPadX, cardPillMinW, innerW)` and
  call it from *both* `cardHeaderColumnWOf` (the reservation) and `renderCardChrome`
  (the drawn pill), so they never drift. A pinned `cardPillPadX = In(0.10)` per side
  (a layout metric, not a token) absorbs the text frame's default inset so the
  measured label fits; a `cardPillMinW = In(0.30) = cardPillH` keeps a one-character
  pill a proper circular chip.
- **Single-line guarantee via `fitScale`.** When the pill is clamped to `innerW` (a
  label too long even at full width), `fitScale(naturalWidth, pillW − 2·pillPadX)`
  shrinks the label to one line; when the pill is *not* clamped, `pillW − 2·pillPadX
  == naturalWidth`, so `fitScale` returns 0 (no shrink). This reuses the R10.5
  primitive instead of inventing ellipsis logic, and keeps the un-clamped pill at
  full font size.
- **Not byte-identical, by design.** Every pill's width changes from the fixed
  `In(1.0)` to its fitted width (and the reserved header column changes with it). The
  R11.5 spec explicitly does not require byte-identity ("byte-identical when the
  measured width rounds to the old value is not required"). Determinism still holds
  (pure integer `naturalWidth`), and the existing pill tests assert presence/shape
  counts, not the fixed width, so they pass unchanged.
- **Reservation symmetry preserved.** `cardHeaderColumnWOf` and the renderer both
  reserve `pillW + gapSM`, so the header title/eyebrow column shrinks by exactly the
  fitted pill width — the wrapped-header geometry (R10.1) stays consistent.

## 4. Recommendation

Add `cardPillWidthOf` + `cardPillPadX`/`cardPillMinW`; route both pill sites through
it; apply `fitScale` to the pill run for the single-line guarantee. Tests: a
white-box width sweep (long label > min and ≤ innerW, sizes to naturalWidth + pad,
short floored, empty → 0, reservation == drawn), and a render-level check that a long
pill label produces a wider-than-`In(1.0)` box and a single-line (scaled when
clamped) run. D-085 records the fit-to-label mechanism.

## 5. Open questions

- Ellipsis vs shrink for a pathologically long label: `fitScale` shrinks to a 0.60
  floor, then the label may still overflow a tiny `innerW`. Acceptable — the realistic
  failure (a 1.0"-fixed pill wrapping a normal label) is fixed; sub-0.60 labels in a
  sub-inch card are out of scope.
