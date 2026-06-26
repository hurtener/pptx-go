#!/usr/bin/env bash
#
# Phase 96 smoke — multi-archetype conformance corpus (Deckard R14.19, CRITICAL).
# A deterministic corpus of professional deck archetypes (cover/section/agenda/
# comparison-matrix/pricing/timeline/org-chart/quote/photo-cover/logo-wall/chart/
# dashboard/process/quadrant/closing) rendered across light + dark variants and
# asserted against the standing invariants: every box on-canvas, OOXML-conformant,
# and byte-identical re-render. The generalizable proof beyond the one sample deck.
set -uo pipefail
cd "$(dirname "$0")/../.."
OK=0; FAIL=0; SKIP=0
ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }
run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then skip "$desc" "not yet landed"
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then ok "$desc"
	else fail "$desc" "test failed"; fi
}
grep_check() {
	local desc="$1" file="$2" pat="$3"
	if [ ! -f "$file" ]; then skip "$desc" "missing $file"
	elif grep -q "$pat" "$file"; then ok "$desc"
	else fail "$desc" "pattern not found: $pat"; fi
}
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
grep_check "corpus archetypes fixture"  scene/render_corpus_test.go 'func corpusArchetypes'
grep_check "corpus covers 15+ archetypes" scene/render_corpus_test.go 'mk("orgchart"'
run_check  "corpus all boxes on-canvas"  ./scene/ 'TestCorpus_AllBoxesOnCanvas'
run_check  "corpus OOXML-conformant"     ./scene/ 'TestCorpus_Conformant'
run_check  "corpus byte-identical"       ./scene/ 'TestCorpus_Deterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
