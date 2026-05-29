#!/usr/bin/env bash
#
# Phase 06 smoke — Leaf-node rendering (docs/plans/phase-06-leaf-rendering.md §13).

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

run_check "one-of-each leaf renders + conformance" ./scene/ 'RenderLeaves|EachLeaf'
run_check "code_block renders image + caption" ./scene/ 'CodeBlock'
run_check "9pt text stays 9pt (no boosting)" ./scene/ 'NoBoost|NinePt'
run_check "scene→PPTX→reopen round-trip" ./scene/ 'RenderRoundTrip'

echo
echo "phase-06 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
