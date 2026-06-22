#!/usr/bin/env bash
#
# Phase 54 smoke — chrome-element anti-collision (Deckard R11.6). Verifies a status
# dot shifts left of the header pill when both are set (disjoint top-right chrome),
# and a dot-only card is byte-identical (no shift)
# (docs/plans/phase-54-chrome-element-anti-collision.md §11/§13).
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
run_check "status dot shifts left of the header pill"  ./scene/ 'TestStatusDot_AntiCollision'
run_check "dot-only card keeps the corner placement"   ./scene/ 'TestStatusDot_ByteIdentical_NoPill'
echo
echo "phase-54 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
