# Brief 31 — estimate-actual-parity-fit-budget

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 48 — estimate/actual parity (R10.10, HIGH · engine)

## 1. Question

`preferredHeight`'s slot estimates diverge from what the composers actually emit,
so overflow detection and the fit pass operate on wrong numbers: `cardChromeEst`
is a fixed ~1.2" regardless of a wrapped multi-line header, and the bento estimate
measures every cell at the *unit* column width even though a wide-span cell renders
wider (and wraps less). How can the estimators be aligned with the composers —
trustworthy for any content — while single-line / unit-width cases stay
byte-identical?

## 2. Prior art surveyed

- `scene/render.go` `preferredHeight` — the Card/CardSection case adds the fixed
  `cardChromeEst` (1.2") + `nodesHeight(body, avail − 2·cardBodyInsetEst)`; the
  Bento case measures every cell at the unit `cellW` and takes `nRows × maxCell`.
  Phase 22's content-aware nodes (Quote/Callout) use a *baseline + per-extra-line
  increment* shape — the byte-identical-at-single-line pattern.
- `scene/render_card.go` — the wrapped-header geometry (R10.1, D-070):
  `cardHeaderColumnW` (the true header column at which the eyebrow/title wrap),
  `cardHeaderRowHeights` (`per-row × wrappedLines`), `cardHeaderBottom`. These are
  `*renderer` methods that use only `r.theme`.
- `scene/render_bento.go` `cellWidth(span, unitW, gap)` — the actual per-cell span
  width the composer uses.
- D-070's deferral note: "making `preferredHeight`'s `cardChromeEst` wrapped-header
  aware … the slot-size estimate parity follows" → R10.10.
- DECKARD R10.10 spec: align estimators with composers — `cardChromeEst` becomes
  wrapped-header-aware (shared constants); bento/grid preferred-height uses each
  cell's actual span width; card body inset matches `cardPadding`; pinned/integer;
  where an estimate already equals the composed value (single-line), unchanged.

## 3. Findings

- The byte-identical-respecting fix mirrors Phase 22: keep the **baseline**
  (`cardChromeEst`) for single-line and **add** the extra wrapped header/eyebrow
  lines. `extra = (wrappedLines(header) − 1)·cardTitleRowH + (wrappedLines(eyebrow)
  − 1)·cardEyebrowRowH`, measured at the card's true header column width. A
  single-line header gives `extra = 0` → the estimate is unchanged (byte-identical);
  a multi-line header grows the slot to account for the wrapped header (exactly the
  R10.1 deferral). This needs the header column width, so `cardPadding` /
  `cardPaddingFor` / `cardHeaderColumnW` are refactored to theme-taking free
  functions with thin `*renderer` method wrappers (the composer sites and tests are
  unchanged).
- **Bento span width.** Measure each cell at its actual span width `span·unitW +
  estGap·(span−1)` instead of the unit `cellW`. A span-1 cell yields `unitW` →
  byte-identical; a wide-span cell wraps less, so the estimate shrinks to the
  accurate (no-longer-over-counted) value. This removes the documented "wider-span
  cells over-estimate" caveat.
- **Over-estimation is safe, single-line is sacred.** The current `cardChromeEst`
  (1.2") is a loose over-estimate even for single-line; lowering it to the exact
  composed chrome would be more accurate but would change single-line output. The
  acceptance prioritizes *single-line byte-identical*, so the baseline is kept and
  only the wrapped-line increment (the under-counted part that caused real overflow)
  is added. Over-estimation warns conservatively — the safe direction.
- **Card body inset deferred.** Matching `cardBodyInsetEst` (0.20"/side) to the
  actual `cardPadding` (e.g. MD = 0.11") would change `bodyW` → the body wrap count
  → single-line output. Since the chrome + span-width fixes already bring the
  representative nodes within one line-height, the inset estimate is left pinned
  (byte-identical); matching it exactly is a future refinement.

## 4. Recommendations

1. Refactor `cardPadding` / `cardPaddingFor` / `cardHeaderColumnW` to theme-taking
   free functions + `*renderer` method wrappers (no composer/test churn).
2. Card/CardSection `preferredHeight`: `cardChromeEst + extraHeaderLines + body`,
   where `extraHeaderLines` is the wrapped-header/eyebrow increment at the header
   column width (0 for single-line).
3. Bento `preferredHeight`: measure each cell at its actual span width.
4. Tests: a multi-line-header card's estimate exceeds the single-line one by the
   wrapped increment; a wide-span bento cell's estimate ≤ the unit-width one;
   single-line card / span-1 bento byte-identical; the overflow warning fires for a
   genuinely-overflowing wrapped-header card; determinism; smoke `phase-48.sh`.

## 5. Open questions

- **Exact card body inset parity** — deferred (would break single-line
  byte-identity for a second-order accuracy gain).
- **Grid wide-span** — `Grid` cells are single children with no span, so the Grid
  estimate is already unit-width-correct; nothing to change there.
