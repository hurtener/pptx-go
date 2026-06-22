#!/usr/bin/env bash
#
# Phase 58 smoke — proportional list bullet indent (Deckard R11.10). Verifies the
# IndentTight hanging indent is byte-identical In(0.25) at the default 14pt body,
# scales with the body type size, and stays tighter than the 0.5" default
# (docs/plans/phase-58-list-bullet-hanging-indent.md §11/§13).
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
run_check "tight indent is In(0.25) at 14pt (anchor)"   ./scene/ 'TestListTightIndent_AnchorByteIdentical'
run_check "tight indent scales with the body size"      ./scene/ 'TestListTightIndent_Proportional'
run_check "tight indent stays below the 0.5\" default"  ./scene/ 'TestListTightIndent_GapTight'
run_check "IndentTight emits In(0.25) marL (R10.9 guard)" ./scene/ 'TestListIndent_TightSmallerOffset'
echo
echo "phase-58 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
