# Brief 40 — bento-rowlabel-gutter-fit

**Subsystem:** scene — Layer 2 renderer (Bento container)
**Authored:** 2026-06-22
**Motivating phase:** Phase 57 — bento-rowlabel-gutter-fit (R11.9, MED · engine)

## 1. Question

The bento row-label gutter is a fixed `bentoGutterW = In(1.2)`; "Control plane"
wraps awkwardly to two lines in it and "The core" sits near the footer (recreation
slide 6). The gutter width is unrelated to the actual label widths. How can it size
to the widest label, clamped to a sane range, deterministically?

## 2. Prior art surveyed

- **`scene/render_bento.go bentoColumns`** — sets `gutterW = bentoGutterW` (fixed)
  iff any row is labeled; shared by `bentoGeometry` and `bentoWeightedRowHeights`.
- **`scene/render.go preferredHeight` (Bento case)** — subtracts the same fixed
  `bentoGutterW` from the content width to estimate the unit column width.
- **`scene/metrics.go naturalWidth`** — the deterministic label-width estimator.
- DECKARD R11.9 spec: `gutterW = clamp(max over rows of naturalWidth(label @
  TypeCaption, Bold) + padding, minGutter, maxGutter)` instead of the fixed value;
  byte-identical when all labels fit the old 1.2".

## 3. Findings

- **One shared fit function, used by layout and the estimator.** Extract
  `bentoGutterWidthOf(theme, v) = clamp(max naturalWidth(label @ TypeCaption) +
  2·bentoGutterPadX, bentoGutterMinW, bentoGutterMaxW)` (0 when unlabeled) and call
  it from both `bentoColumns` (the drawn gutter) and the `preferredHeight` Bento case
  (the slot estimate), so the layout and the estimate use the same gutter — closing
  the parity the fixed constant left.
- **Thread `theme` into `bentoColumns` / `bentoGeometry`.** `naturalWidth` needs the
  theme; the two free functions gain a `theme *pptx.Theme` parameter (the white-box
  tests already construct a `pptx.DefaultTheme()`, the renderer passes `r.theme`).
- **Bounds.** `bentoGutterMinW = In(0.8)` (a short label still gets a usable gutter),
  `bentoGutterMaxW = In(1.6)` (a long label caps rather than starving the cells),
  `bentoGutterPadX = In(0.1)` each side. The base `In(1.2)` is replaced — most
  realistic labels land between the bounds.
- **Not byte-identical, by design.** The gutter changes for most label sets (a 1-char
  "A" → `In(0.8)`, "Control plane" → its fitted width), and the unit column width
  changes with it. R11.9 does not require byte-identity; determinism holds (pure
  integer `naturalWidth`), and the existing bento tests assert gutter-presence, span
  ratios, and equal-row heights — all gutter-width-independent — so they pass.
- **Vertical clipping.** The labels render anchor-middle in their row height (already
  the case); the fit addresses horizontal wrapping/clipping. A multi-line label still
  fits its row via the existing anchor-middle frame.

## 4. Recommendation

Add `bentoGutterWidthOf` + `bentoGutterMinW`/`MaxW`/`PadX`; route `bentoColumns` and
the `preferredHeight` Bento estimate through it; thread `theme` into the bento
geometry functions. White-box tests: the gutter fits/clamps per label width, and the
geometry reserves exactly `bentoGutterWidthOf`. The existing bento determinism tests
(labeled rows) cover the rendered path. D-089 records the fit-to-label gutter.

## 5. Open questions

- A multi-line (wrapped) row label growing the row height is out of scope; the fit
  keeps a single-line label inside the gutter, which is the reported case.
