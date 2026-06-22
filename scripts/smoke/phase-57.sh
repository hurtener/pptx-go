#!/usr/bin/env bash
#
# Phase 57 smoke — bento row-label gutter fit (Deckard R11.9). Verifies the gutter
# sizes to the widest row label (clamped to a min/max), the geometry reserves
# exactly that width, and labeled bentos render deterministically
# (docs/plans/phase-57-bento-rowlabel-gutter-fit.md §11/§13).
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
run_check "gutter fits/clamps to the widest label"      ./scene/ 'TestBentoGutterWidth_FitsLabels'
run_check "geometry reserves the fitted gutter width"   ./scene/ 'TestBentoGutter_GeometryUsesFit'
run_check "labeled bento renders deterministically"     ./scene/ 'TestBento_Deterministic'
echo
echo "phase-57 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
