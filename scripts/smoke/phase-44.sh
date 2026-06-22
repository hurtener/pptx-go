#!/usr/bin/env bash
#
# Phase 44 smoke — fill cap / no over-grow (Deckard R10.6). Verifies the opt-in
# VAlignFillCapped bounds each flexible node's growth, turns the leftover slack
# into balanced spacing within the box, and renders deterministically across
# worker counts (docs/plans/phase-44-fill-cap-no-overgrow.md §11/§13).
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
run_check "capped growth bounds a sparse node + residual"  ./scene/ 'TestDistributeFillCapped_BoundsAndResidual'
run_check "capped fill spaces evenly within the box"       ./scene/ 'TestFillCapped_EvenSpacingWithinBox'
run_check "capped fill render stays deterministic"         ./scene/ 'TestRenderDeterministic_VAlignFillCapped'
echo
echo "phase-44 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
