#!/usr/bin/env bash
#
# Phase 43 smoke — display text shrink-to-fit (Deckard R10.5). Verifies the opt-in
# AutoFit shrinks an over-wide display run to fit its box (at or above the pinned
# ratio floor), keeps fitting/AutoFit-off content byte-identical, round-trips a
# scaled run's size, and renders deterministically
# (docs/plans/phase-43-display-text-shrink-to-fit.md §11/§13).
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
run_check "over-wide AutoFit value fits at >= ratio floor" ./scene/ 'TestFitScale_OverflowFitsAtOrAboveFloor'
run_check "AutoFit-off / fitting text byte-identical"      ./scene/ 'TestAutoFit_OffByteIdentical'
run_check "scaled run round-trips its size"                ./pptx/  'TestRunFontScale_RoundTrip'
run_check "AutoFit render stays deterministic"             ./scene/ 'TestAutoFit_Deterministic'
echo
echo "phase-43 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
