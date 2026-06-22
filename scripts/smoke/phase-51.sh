#!/usr/bin/env bash
#
# Phase 51 smoke — container slide-bounds clamp (Deckard R11.3). Verifies the
# safe-area clamp: an over-tall container is shrunk so its bottom never exceeds the
# slide safe area (with a warning), every emitted bento cell stays inside the safe
# area, a fitting container is byte-identical (no clamp, no warning), and the path
# is deterministic across worker counts
# (docs/plans/phase-51-container-slide-bounds-clamp.md §11/§13).
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
run_check "overflowing box is shrunk + warns"          ./scene/ 'TestClampToSafeArea_ShrinksOverflow'
run_check "fitting box is byte-identical (no warn)"    ./scene/ 'TestClampToSafeArea_FitsByteIdentical'
run_check "clamped bento cells stay in the safe area"  ./scene/ 'TestBentoBoxesWithinSafeArea'
run_check "over-tall container warns at render"        ./scene/ 'TestContainerOverflow_Warns'
run_check "fitting container does not warn"            ./scene/ 'TestContainerFits_NoWarn'
run_check "bounds clamp deterministic across workers"  ./scene/ 'TestRenderDeterministic_BoundsClamp'
echo
echo "phase-51 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
