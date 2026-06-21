#!/usr/bin/env bash
#
# Phase 30 smoke — letter-spacing (tracking) token (Deckard R9.3, opens Wave 9).
# Verifies FontSpec.Tracking / RunStyle.Tracking emit OOXML a:rPr/@spc, are
# byte-identical when zero, and round-trip via Run.Tracking()
# (docs/plans/phase-30-type-tracking.md §11/§13).

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then
		skip "$desc" "not yet landed"
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then
		ok "$desc"
	else
		fail "$desc" "test failed"
	fi
}

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 2. Tracking emits a:rPr/@spc (signed); role-level path works (criteria 1, 3).
run_check "tracking emits a:rPr/@spc"               ./pptx/ 'TestTracking_EmitsSpc|TestTracking_RoleLevel'
# 3. Zero tracking is byte-identical (criterion 2).
run_check "zero tracking is byte-identical"         ./pptx/ 'TestTracking_ZeroByteIdentical'
# 4. Tracking round-trips via Run.Tracking() (criterion 4).
run_check "tracking round-trips"                    ./pptx/ 'TestTracking_RoundTrips'

echo
echo "phase-30 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
