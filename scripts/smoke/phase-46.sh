#!/usr/bin/env bash
#
# Phase 46 smoke — balanced vertical rhythm (Deckard R10.8). Verifies the opt-in
# VAlignBalanced distributes a sparse stack's slack as an even rhythm (top margin
# + widened gaps) within the box, keeps Top/Center byte-identical, and renders
# deterministically (docs/plans/phase-46-balanced-vertical-rhythm.md §11/§13).
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
run_check "balanced distributes even rhythm within the box" ./scene/ 'TestBalanced_EvenRhythmWithinBox'
run_check "Top/Center byte-identical under balanced"        ./scene/ 'TestBalanced_TopCenterByteIdentical'
run_check "balanced render stays deterministic"             ./scene/ 'TestRenderDeterministic_VAlignBalanced'
echo
echo "phase-46 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
