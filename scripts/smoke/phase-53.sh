#!/usr/bin/env bash
#
# Phase 53 smoke — header-pill fit-to-label (Deckard R11.5). Verifies the pill sizes
# to its label (naturalWidth + padding, floored and clamped to the inner width), the
# header-column reservation matches the drawn pill width, and the path is
# deterministic across worker counts
# (docs/plans/phase-53-header-pill-fit-to-label.md §11/§13).
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
run_check "pill width tracks its label (fit + clamp)"   ./scene/ 'TestCardPillWidth_FitsLabel'
run_check "header reservation matches the drawn pill"   ./scene/ 'TestCardPillWidth_ReservationMatchesDrawn'
run_check "pill fit deterministic across workers"       ./scene/ 'TestRenderDeterministic_PillFit'
echo
echo "phase-53 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
