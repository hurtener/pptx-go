# Brief 22 — content-aware card header height

**Subsystem:** scene — Layer 2 renderer (card chrome / layout)
**Authored:** 2026-06-22
**Motivating phase:** Phase 39 — wrapped-title-aware card header height

## 1. Question

A card's header row is advanced by a **fixed** per-row height
(`cardEyebrowRowH` / `cardTitleRowH`) in both `cardHeaderBottom` (the body-Y
computation) and `renderCardChrome` (the emitted header text frame + the D-054
header band). When a header (or eyebrow) wraps to two lines in a narrow card, the
body region still begins at the *single-line* header bottom — so the second
header line collides with the body list/prose. How can the body region and the
header band size to the **actual wrapped** header height, deterministically, with
single-line cards staying byte-identical?

This is `DECKARD-PRODUCT-REQUIREMENTS.md` R10.1 (`content-aware-card-header-height`,
CRITICAL · engine) — one of the two off-slide/overlap bugs that open Wave 10.

## 2. Prior art surveyed

- **`scene/render_card.go`.** `cardHeaderBottom` and `renderCardChrome` already
  share the row-height *constants* (extracted in the Wave-8 checkpoint precisely so
  they never drift), but both add a single fixed row regardless of wrap. The
  header text column width is computed only inside `renderCardChrome` (innerW
  minus the icon-left shift and the header-pill reservation) — `cardHeaderBottom`
  doesn't compute it at all.
- **`scene/metrics.go::wrappedLines(rt, base, avail, theme)`** (brief 09) — the
  deterministic line-count estimator (`ceil(naturalWidth / avail)`), already used
  by `preferredHeight` for Prose/List/Callout/table cells. A plain string wraps as
  `RichText{{Text: s}}`. This is exactly the primitive R10.1 needs.
- **D-054 header band.** `renderCardChrome` already sizes the band to
  `cardHeaderBottom(box,c) - box.Y`, so fixing `cardHeaderBottom` fixes the band
  for free.
- **R10.10 (estimate-actual-parity).** Separately owns making `preferredHeight`'s
  `cardChromeEst` (a fixed ~1.2") wrapped-header-aware. R10.1's acceptance is the
  *rendered* geometry (no overlap); the slot-estimate parity is R10.10.

## 3. Findings

- **The two paths must wrap at the same width.** The fix is a shared header-column
  width: `cardHeaderColumnW(box,c)` = innerW − (icon-left shift) − (pill
  reservation), computed identically for both the body-Y path and the emit path.
  A second shared helper `cardHeaderRowHeights(box,c)` returns
  `eyebrowH = cardEyebrowRowH × wrappedLines(eyebrow, TypeCaption, headerW)` and
  `titleH = cardTitleRowH × wrappedLines(header, TypeH3, headerW)`.
- **`cardHeaderBottom` consumes the heights** instead of adding fixed rows;
  `renderCardChrome` emits the eyebrow/title text frames at those heights and
  advances by them. Same helper → no drift; the D-054 band (sized off
  `cardHeaderBottom`) tracks automatically.
- **Single-line is byte-identical.** `wrappedLines` returns 1 for text that fits,
  so `titleH == cardTitleRowH` and `eyebrowH == cardEyebrowRowH` exactly — the
  emitted boxes, band, and body Y are unchanged for every card whose header fits
  one line at its column width. Only wrapping headers change (the bug fix).
- **Deterministic.** `wrappedLines` is pure integer `ceil` division over pinned
  EMU; no map iteration, no worker dependence.
- **Estimator parity is deferred to R10.10.** This phase fixes the *composed*
  geometry (the CRITICAL overlap). `preferredHeight`'s `cardChromeEst` stays a
  fixed estimate until R10.10 wires it to the wrapped `cardHeaderBottom` — the
  same split D-061 used (visual first, estimator parity follows).

## 4. Recommendations

1. `scene/render_card.go`: add `cardHeaderColumnW(box,c)` and
   `cardHeaderRowHeights(box,c)`; route `cardHeaderBottom` + `renderCardChrome`
   through them.
2. White-box tests: a long header in a 1/3-width card advances the body bottom by
   ≥2 title rows; the body region top ≥ the wrapped header bottom (no overlap);
   single-line is byte-identical to the legacy fixed advance.
3. Smoke + a determinism guard; note `cardChromeEst` parity is R10.10.

## 5. Open questions

- **Per-line height vs true leading.** The per-line advance stays the fixed
  `cardTitleRowH`/`cardEyebrowRowH` constant (not D-061 leading-derived); R9.4
  deferred leading-aware line height to the estimator rework (R9.5/R10.10), and
  this phase preserves that — the multiplier is the wrapped *count*.
- **Icon-top layout.** The header column for icon-top is the full innerW (the icon
  sits above, not beside), already handled by `cardHeaderColumnW`'s `c.layout`
  check.
