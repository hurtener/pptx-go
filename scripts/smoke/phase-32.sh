#!/usr/bin/env bash
#
# Phase 32 smoke — case-transform token (Deckard R9.11). Verifies FontSpec.Case /
# RunStyle.Case emit OOXML a:rPr/@cap (all/small), are byte-identical when none,
# round-trip via Run.Case() while preserving the run text, and that the role-level
# path emits on an override-free run (docs/plans/phase-32-text-case.md §11/§13).

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

# 2. Case emits a:rPr/@cap; role-level path works (criteria 1, role-level).
run_check "case emits a:rPr/@cap"                   ./pptx/ 'TestCase_EmitsCap|TestCase_RoleLevel'
# 3. No case is byte-identical (criterion 2).
run_check "no case is byte-identical"               ./pptx/ 'TestCase_NoneByteIdentical'
# 4. Case round-trips and preserves the run text (criterion 3, G6).
run_check "case round-trips, text preserved"        ./pptx/ 'TestCase_RoundTripsAndPreservesText'

echo
echo "phase-32 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
