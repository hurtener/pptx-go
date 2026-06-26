# Phase 96 — multi-archetype conformance corpus

**Subsystem:** `scene` (test corpus)
**RFC sections:** §16 (testing), §10 (invariants)
**Deps:** D-092 (adversarial harness); brief 79.
**Status:** Done

## 1. Goal
Add a standing corpus of professional deck archetypes, rendered across variants and
checked against the structural invariants — the generalizable proof beyond the one
sample deck.

## 2. Why now
Wave 14 CRITICAL meta req R14.19; the engine's coverage must be measured per class.

## 3-6. RFC/brief/decisions
RFC §16 (tests), §10 (on-canvas/determinism). Brief 79. Decisions D-092 (adversarial
harness), D-026, D-132 (corpus), D-133 (Wave-14 deferrals).

## 7. Architecture
`corpusArchetypes(variant)` returns one fixture slide per archetype class (cover/
section/agenda/comparison-matrix/pricing/timeline/org-chart/quote/photo-cover/
logo-wall/chart/dashboard/process/quadrant/closing); `corpusScene()` renders both
light + dark. Three tests assert: every box on-canvas (the R11.12 regex check
reused), OOXML conformance, byte-identical re-render across worker counts. Test-only;
RTL/CJK deferred (R14.15 → V2).

## 8. Files
scene/render_corpus_test.go (NEW), scripts/smoke/phase-96.sh, docs/research/79 +
INDEX, this plan, README, glossary, decisions D-132 + D-133.

## 9. Public API
None (test-only).

## 10-11. Risks/acceptance
A benign chart aspect advisory is expected (not asserted as a failure). Accept: all
corpus archetypes render on-canvas across light/dark; conformant; byte-identical;
adding an archetype is a one-fixture addition.

## 12-14. Coverage/smoke/tests
scripts/smoke/phase-96.sh. Tests: TestCorpus_AllBoxesOnCanvas / _Conformant /
_Deterministic over the 2-variant archetype corpus.
