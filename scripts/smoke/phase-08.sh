#!/usr/bin/env bash
#
# Phase 08 smoke — Table (docs/plans/phase-08-table.md §13).

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

if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

run_check "merged-cell round-trip" ./test/pptx/ 'TableMerge|MergedCell'
run_check "banded-table alternating fills" ./test/pptx/ 'TableBand'
run_check "graphic frame emits p:xfrm" ./internal/ooxml/slide/ 'GraphicFrameXfrm|TableXfrm'
run_check "scene table + caption render" ./scene/ 'RenderTable|TableCaption'

echo
echo "phase-08 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
