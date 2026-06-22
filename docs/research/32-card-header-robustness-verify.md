# Brief 32 — card-header-content-aware-height (R11.1 verify-and-close)

**Subsystem:** scene — Layer 2 renderer (card chrome / layout)
**Authored:** 2026-06-22
**Motivating phase:** Phase 49 — R11.1 card-header-content-aware-height (CRITICAL · engine)

## 1. Question

R11.1 (DECKARD R11) demands that a card's header band height and body-start Y grow
with the *actual* wrapped line count of the eyebrow + title — at any title length,
any inner width, any size/layout — so a long header never spills past its band onto
the body. Is this already satisfied by the Wave-10 work, and if so, what is the
minimal close: a verification + the R11.1 acceptance golden, or new mechanism?

## 2. Prior art surveyed

- **R10.1 / D-070 (Phase 39)** already replaced the fixed single-line header
  geometry with wrapped-aware row heights:
  - `cardHeaderColumnWOf(theme, box, c)` (`scene/render_card.go`) — the true header
    text column width: `innerW − iconShift − pillReserve`, the width at which the
    eyebrow and title wrap.
  - `cardHeaderRowHeights(box, c)` — `(eyebrowH, titleH)` each `per-row-const ×
    wrappedLines(text, role, headerW, theme)`; single-line yields exactly the
    legacy fixed row (byte-identical).
  - `cardHeaderBottom(box, c)` and `renderCardChrome` **both** route through
    `cardHeaderRowHeights`, so the band drawn (in `renderCardChrome`), the body Y,
    and the computed bottom never drift — exactly R11.1's "MUST consume the
    identical headerW/line-count computation" requirement.
- **R10.10 / D-079 (Phase 48)** closed the estimator side: `cardHeaderExtraHeight`
  feeds `preferredHeight` the wrapped header lines, so the slot estimate matches
  the composed chrome.
- Existing white-box tests (`render_card_internal_test.go`): `TestCardHeaderBottom_
  WrappedTitle` and `TestCardBodyBelowWrappedHeader` already prove the advance and
  the no-overlap property — but only for `CardSizeMD` / `CardLayoutDefault`, a
  single combo.

## 3. Findings

- **R11.1 is already implemented by D-070.** The spec text of R11.1 ("replace the
  fixed `hH := cardTitleRowH` with `cardTitleRowH × wrappedLines(...)`", "extract a
  shared `headerRowHeights`", "the two functions MUST consume the identical
  computation") describes verbatim what `cardHeaderRowHeights` + the two shared call
  sites already do. No reimplementation is warranted — doing so would be churn.
- **The acceptance gap is test coverage, not mechanism.** R11.1's acceptance is
  explicit: "A golden test with a deliberately long multi-line header asserts
  `body.Y >= bandBottom` **across all CardSize/CardLayout combinations**; single-line
  headers stay byte-identical." The existing tests cover one combo. The close is the
  full sweep: `{CardSizeMD, SM, LG} × {CardLayoutDefault, IconTop}` (6 combos), each
  asserting (a) the rendered body box top ≥ the wrapped header bottom, (b) the D-054
  header band, when enabled, fully contains the header (band bottom == header
  bottom), and (c) a single-line header is byte-identical to the legacy fixed
  advance.
- **Icon-top interaction.** `CardLayoutIconTop` adds `cardIconSz + gapSM` before the
  header rows in *both* `cardHeaderBottom` and `renderCardChrome`, so the property
  holds there too — the sweep proves it rather than assuming it.
- **No new public API, no new token, no OOXML change.** This phase is a verification
  + an acceptance golden + a decision entry recording that R11.1 is closed by D-070
  (and D-079 for the estimator). Per §17 this is the right shape for a
  verify-and-close: prove the invariant under the full combinatorial content the
  requirement names, then record the closure so it is not re-litigated.

## 4. Recommendation

Ship Phase 49 as: (1) `TestCardBodyBelowWrappedHeader_AllCombos` — the
CardSize×CardLayout sweep asserting `body.Y >= cardHeaderBottom` and band
containment for a long header, plus single-line byte-identical; (2) a smoke check;
(3) D-081 recording R11.1 = closed-by-D-070/D-079 with the acceptance golden as the
evidence. No code change to the renderer.
