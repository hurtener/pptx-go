# Brief 43 — adversarial-content-fit-fixtures

**Subsystem:** scene — Layer 2 renderer (test harness + safe-area clamp)
**Authored:** 2026-06-22
**Motivating phase:** Phase 60 — adversarial-content-fit-fixtures (R11.12, HIGH · both)

## 1. Question

The recreation's overlaps reproduced only under real, long, or dark content; the
existing tests pass because their fixtures use short, light, single-line content.
There is no torture suite proving each component renders correctly under hostile
content. How can the engine ship a reusable, deterministic acceptance harness that
renders every component with adversarial content and asserts the structural
invariants — and what does it surface?

## 2. Prior art surveyed

- The per-component Wave-11 guards: header/body disjoint (Phase 49), safe-area clamp
  (Phase 51), pill / join / stat fit (Phases 53/55/56), auto-contrast (Phase 50) —
  each tested in isolation with one hostile case.
- **`scene/render.go renderNode`** — the single dispatch point for every node; the
  R11.3 clamp currently lives in the three *container* composers
  (`renderBento`/`renderGrid`/`renderCard`), not at `renderNode`.
- **D-059 scope:** R11.12 is tagged `both`; the engine side is the harness + fixtures
  + the invariant assertions (Deckard's authoring-loop side operates on its own
  packages and is out of this repo).
- DECKARD R11.12 spec: a reusable harness rendering every component under adversarial
  content, asserting (1) header band ≤ body top, (2) every box within the safe area,
  (3) fit-required text is one line, (4) chrome text passes the luminance-contrast
  check; deterministic; part of the CI gate.

## 3. Findings

- **A box recorder is unnecessary — the emitted XML already carries every box.** The
  on-canvas invariant (2) is asserted by parsing every `<a:off>`/`<a:ext>` pair from
  each rendered slide and checking it lies within the `12192000 × 6858000` canvas.
  This needs no test-only sink (the spec's optional approach) and cannot perturb byte
  output. It is the strongest catch: any component that reintroduces a fixed-size
  assumption and draws off-slide fails here.
- **The suite surfaced a real bug: card-body leaf overflow.** A `Grid` of cards whose
  bodies hold a tall hostile `List` placed a body leaf *below the card and off the
  slide canvas* (bottom 7.28 M EMU vs the 6.858 M canvas). The R11.3 clamp clamps the
  *container* box but not a leaf the over-full card body pushes past it. Per
  `CLAUDE.md §17`, fix it in this PR.
- **The fix generalizes the R11.3 clamp to `renderNode`.** Clamp every *content*
  node's box to the safe area at the single dispatch point, exempting the full-slide
  overlays (`Decoration`, which may bleed off-canvas by design, and `SectionDivider`).
  This subsumes the three per-container clamps (which are removed, leaving one clamp
  point and one warning) and additionally caps an over-full card body / stack leaf.
  Byte-identical when the box already fits (the clamp is a no-op); pure integer →
  deterministic; the existing Phase-51 clamp tests still pass (the bento still warns,
  the helper is unchanged).
- **Invariants (1)/(3)/(4) are asserted white-box on the same hostile inputs** via the
  engine helpers (`cardHeaderBottom`/`renderCardChrome`, `cardPillWidthOf` +
  `statValueFit` + `fitScale`, `onCardSurface` + `contrastRatioT10`), so the harness is
  self-contained rather than only re-pointing at the per-component tests.

## 4. Recommendation

Ship `render_adversarial_test.go` (the hostile fixture + the on-canvas parse +
determinism) and `render_adversarial_invariants_test.go` (header/body, fit-one-line,
contrast). Generalize the R11.3 clamp to `renderNode` (exempting overlays) and remove
the three redundant container clamps. D-092 records the harness and the leaf-overflow
fix.

## 5. Open questions

- Card-body content that genuinely exceeds the card is now *clamped* (box on-canvas)
  and *warned* (the overflow `LayoutWarning`); legibly compressing it remains the
  opt-in `VAlignFit` / `Card.BodyVAlign` path (D-026 — the product drives density).
