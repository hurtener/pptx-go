#!/usr/bin/env bash
#
# Phase 103 smoke — estimator inter-node gap derives from the theme SpaceMD
# token (Deckard R15.2). Verifies the estimator matches composed geometry, the
# 2-node column no longer overflows the box it was sized for, the existing
# byte-identity properties survive, and determinism + drift guard stay green.
# (docs/plans/phase-103-estimator-gap-token.md §11/§13.)
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
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
run_check "N-node estimated height matches composed (1..4)"  ./scene/ 'TestNodesHeightMatchesComposed'
run_check "2-node column no longer overflows its box"        ./scene/ 'TestTwoColumnPrefitFitsItsBox'
run_check "single-line card wrapped increment is zero"        ./scene/ 'TestPreferredHeight_SingleLineCardUnchanged'
run_check "span-1 bento byte-identity property preserved"     ./scene/ 'TestPreferredHeight_BentoSpanOneByteIdentical'
run_check "worker-count determinism (1 vs 8 workers)"         ./scene/ 'TestRenderDeterministic_ParallelMatchesSequential'
run_check "card-chrome wrapped-header increment intact"       ./scene/ 'TestPreferredHeight_WrappedCardGrows'
run_check "wrapped-header card overflow warning fires"        ./scene/ 'TestOverflow_WrappedHeaderCardFires'
if [ -x ./scripts/drift-audit.sh ] && ./scripts/drift-audit.sh >/dev/null 2>&1; then ok "drift-audit clean"; else fail "drift-audit clean" "scripts/drift-audit.sh reported failures"; fi
if diff -q AGENTS.md CLAUDE.md >/dev/null 2>&1; then ok "AGENTS.md == CLAUDE.md"; else fail "AGENTS.md == CLAUDE.md" "diff between AGENTS.md and CLAUDE.md"; fi
echo
echo "phase-103 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
