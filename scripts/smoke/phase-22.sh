#!/usr/bin/env bash
#
# Phase 22 smoke — content-aware text height (Deckard R1). Verifies that a
# node's slot grows with its wrapped text, stacked nodes stop overlapping,
# overflow is reported truthfully, and single-line content stays byte-identical
# (docs/plans/phase-22-content-aware-height.md §11/§13).

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

# 2. Wrapped-line estimate: monotonic, ceil, empty/zero/nil fallback (criteria 1, 6).
run_check "wrappedLines estimate is sound"            ./scene/ '^TestWrappedLines'
# 3. Single-line content reduces to the prior fixed heights (criterion 4).
run_check "single-line height byte-identical (fixed)" ./scene/ 'TestPreferredHeight_SingleLineReducesToFixed'
# 4. A wrapped paragraph is allotted >= N line-heights (criterion 1).
run_check "wrapped prose grows to N line-heights"     ./scene/ 'TestPreferredHeight_ProseGrowsWithWrap'
# 5. The node below a multi-line paragraph does not overlap it (criterion 2).
run_check "stacked node below multi-line text"        ./scene/ 'TestLayout_NoOverlapMultiLine'
# 6. Overflow warning fires on wrapped content exceeding the body (criterion 3).
run_check "overflow warning fires on wrapped content" ./scene/ 'TestOverflow_FiresOnWrappedContent'
# 7. A multi-line deck renders byte-identically across worker counts (criterion 5).
run_check "multi-line wrap render is deterministic"   ./scene/ 'TestRenderDeterministic_MultiLineWrap'

echo
echo "phase-22 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
