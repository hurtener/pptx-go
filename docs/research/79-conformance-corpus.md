# Brief 79 — Multi-archetype conformance corpus (R14.19)

> Informs Phase 96 (Wave 14). Engine half of req R14.19
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, CRITICAL · both; D-059).

## 1. Motivating phase
Generalization was asserted against ONE reference investor deck. Nothing measured
whether the engine reproduces the OTHER professional classes. R14.19 adds a
standing corpus of archetype fixtures checked against the adversarial invariants,
so any class that breaks is caught before ship.

## 2. Findings
- **A test-only deliverable.** No production change — a new `corpusArchetypes(
  variant)` fixture (one slide per class: cover, section, agenda, comparison-matrix,
  pricing, timeline/roadmap, org-chart, quote, photo-cover, logo-wall, chart,
  dashboard, process funnel/cycle, quadrant, closing) rendered across the light AND
  dark variants, plus tests asserting the standing invariants:
  - every emitted box lies within the slide canvas (the R11.12 on-canvas regex
    check, reused),
  - the deck is OOXML-conformant (`conformance.ValidateBytes`),
  - re-render is byte-identical across worker counts.
- **Composes the wave's nodes.** Each archetype is a realistic composition of the
  Wave-14 + earlier nodes, so a regression in any class (off-canvas, non-conformant,
  non-deterministic) fails CI. Adding an archetype is a one-fixture addition.
- **RTL/CJK deferred.** R14.15 layout-direction is not implemented (deferred to
  V2), so the corpus asserts the LTR archetypes; the RTL/CJK variants are noted as
  a follow-up when R14.15 lands.

## 3. Recommendations
- `render_corpus_test.go` with `corpusArchetypes`/`corpusScene` + the three
  invariant tests, wired into the standard `go test`/preflight gate. D-132. Also
  record the Wave-14 deferrals (R14.6/14/15/16/18) in D-133.

## 4. Open questions
- Per-archetype contrast assertions (beyond the engine's auto-contrast mechanism)
  + RTL/CJK fixtures → land with R14.15.
