#!/usr/bin/env bash
#
# Phase 40 smoke — fit-to-region compression (Deckard R10.2). Verifies the opt-in
# VAlignFit mode compresses an over-full body stack inside its region (gaps then
# slot heights, toward pinned floors), stays byte-identical to VAlignTop when the
# content already fits, and renders deterministically across worker counts
# (docs/plans/phase-40-fit-to-region-compression.md §11/§13).
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
run_check "VAlignFit fits an over-full stack inside the region" ./scene/ 'TestFitCompress_PlacementFitsWithinRegion'
run_check "VAlignFit byte-identical to VAlignTop when fitting"  ./scene/ 'TestFitCompress_ByteIdenticalWhenFits'
run_check "fit-to-region compression stays deterministic"      ./scene/ 'TestRenderDeterministic_VAlignFit'
echo
echo "phase-40 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
