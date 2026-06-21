#!/usr/bin/env bash
#
# Phase 28 smoke — Stat leaf node (Deckard R6). Verifies the new Stat node:
# display-scale value + label + optional toned delta, a Grid of Stats as a metric
# strip, Stage-1 validation, catalog/round-trip coverage, and determinism
# (docs/plans/phase-28-stat-node.md §11/§13).

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

# 2. A Stat renders value + label + toned delta (criterion 1).
run_check "stat renders value/label/delta"           ./scene/ 'TestStat_RendersValueLabelDelta|TestDeltaToneColor'
# 3. A Grid of Stats renders a strip (criterion 2).
run_check "grid of stats renders a strip"            ./scene/ 'TestStat_GridStrip'
# 4. Stage-1 rejects an empty-value Stat (criterion 3).
run_check "stat Stage-1 validation"                  ./scene/ 'TestStat_Validation'
# 5. Stat is fixed; catalog has 22 kinds; round-trip covers Stat (criterion 4).
run_check "stat not flexible / catalog 22"           ./scene/ 'TestStat_NotFlexible|TestCatalog_KindsDistinct'
run_check "every-node round-trip covers Stat"        ./test/integration/ 'TestRoundTrip_SceneNodes'
# 6. Stat render is deterministic across workers (criterion 5).
run_check "stat render is deterministic"             ./scene/ 'TestStat_Deterministic'

echo
echo "phase-28 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
