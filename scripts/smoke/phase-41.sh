#!/usr/bin/env bash
#
# Phase 41 smoke — content-weighted bento rows (Deckard R10.3). Verifies the
# opt-in Bento.WeightedRows mode sizes dense rows taller than sparse rows while
# fitting the region, keeps the equal-row default byte-identical, and renders
# deterministically across worker counts
# (docs/plans/phase-41-content-weighted-bento-rows.md §11/§13).
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
run_check "weighted rows size dense > sparse and fit" ./scene/ 'TestBentoWeightedRows_DenseTallerAndFits'
run_check "equal-mode bento geometry byte-identical"  ./scene/ 'TestBentoGeometry_EqualModeByteIdentical'
run_check "weighted bento render stays deterministic" ./scene/ 'TestBentoWeighted_Deterministic'
echo
echo "phase-41 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
