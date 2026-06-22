#!/usr/bin/env bash
#
# Phase 47 smoke — list bullet indent density (Deckard R10.9). Verifies the opt-in
# List.Indent / ParagraphOpts.BulletIndent tightens the bullet hanging indent
# (smaller, consistent marker-to-text offset), keeps the default byte-identical,
# and round-trips (docs/plans/phase-47-list-bullet-indent-density.md §11/§13).
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
run_check "tight bullet indent emits the smaller marL/indent" ./pptx/  'TestBulletIndent_TightEmits'
run_check "default bullet indent byte-identical"              ./pptx/  'TestBulletIndent_DefaultByteIdentical'
run_check "tight List renders a smaller, consistent offset"   ./scene/ 'TestListIndent_TightSmallerOffset'
run_check "default List byte-identical"                       ./scene/ 'TestListIndent_DefaultByteIdentical'
echo
echo "phase-47 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
