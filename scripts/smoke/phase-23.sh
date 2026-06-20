#!/usr/bin/env bash
#
# Phase 23 smoke — vertical fill / grow-to-fit (Deckard R2). Verifies VAlignFill
# pins fixed leaves at the top and grows the flexible nodes to fill the frame,
# the distribution is proportional and deterministic, and non-fill modes stay
# byte-identical (docs/plans/phase-23-grow-to-fit.md §11/§13).

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

# 2. The flexible node set is correct (containers + Image/Chart; not CodeBlock).
run_check "isFlexible covers the intended set"        ./scene/ '^TestIsFlexible$'
# 3. VAlignFill grows the grid to the bottom margin; slot grows past preferred (criteria 1, 2).
run_check "fill grows grid to bottom margin"          ./scene/ 'TestVAlignFill_GridFillsToBottom'
# 4. Two flexible nodes share slack proportionally; remainder to the last (criterion 3).
run_check "fill distributes proportionally"           ./scene/ 'TestDistributeFill_Proportional'
# 5. VAlignFill with no flexible node lays out identically to VAlignTop (criterion 4).
run_check "no-flex fill matches top-align"            ./scene/ 'TestVAlignFill_NoFlexMatchesTop'
# 6. Fill composes with R1: content that overflows still warns (slack <= 0).
run_check "overflow still warns under fill"           ./scene/ 'TestVAlignFill_OverflowStillWarns'
# 7. A VAlignFill deck renders byte-identically across worker counts (criterion 6).
run_check "fill render is deterministic"             ./scene/ 'TestRenderDeterministic_VAlignFill'

echo
echo "phase-23 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
