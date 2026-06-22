#!/usr/bin/env bash
#
# Phase 55 smoke — join-badge fit-to-label (Deckard R11.7). Verifies the TwoColumn
# join badge grows to contain a multi-word label, a short "vs" keeps the base
# diameter (byte-identical), an over-long label caps at the max, and the path is
# deterministic across worker counts
# (docs/plans/phase-55-join-badge-fit-to-label.md §11/§13).
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
run_check "badge grows to a multi-word label, 'vs' base" ./scene/ 'TestJoinBadge_GrowsToLabel'
run_check "over-long label caps at the max diameter"     ./scene/ 'TestJoinBadge_CapsAndShrinks'
run_check "join badge fit deterministic across workers"  ./scene/ 'TestRenderDeterministic_JoinBadgeFit'
echo
echo "phase-55 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
