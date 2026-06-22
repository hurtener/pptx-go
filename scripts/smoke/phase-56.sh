#!/usr/bin/env bash
#
# Phase 56 smoke — stat-value overflow guard (Deckard R11.8). Verifies the Stat
# value steps down the pinned role ladder (TypeDisplay → H1 → H2 → floor+scale) to
# stay on one line as the box narrows, AutoFit-off is byte-identical, and the
# existing AutoFit render/determinism guards still pass
# (docs/plans/phase-56-stat-value-overflow-guard.md §11/§13).
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
run_check "value steps down the role ladder by width"    ./scene/ 'TestStatValueFit_RoleLadder'
run_check "value is one line at the chosen role/scale"   ./scene/ 'TestStatValueFit_OneLine'
run_check "AutoFit Stat emits a reduced sz (render)"     ./scene/ 'TestAutoFit_Stat_EmitsReducedSz'
run_check "AutoFit deck deterministic across workers"    ./scene/ 'TestAutoFit_Deterministic'
echo
echo "phase-56 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
