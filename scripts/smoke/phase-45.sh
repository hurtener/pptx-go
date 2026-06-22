#!/usr/bin/env bash
#
# Phase 45 smoke — density-aware card padding (Deckard R10.7). Verifies the
# additive Card.PaddingScale tightens a card's interior inset (growing the body),
# floors at a pinned minimum, and is byte-identical at the default
# (docs/plans/phase-45-density-aware-card-padding.md §11/§13).
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
run_check "tighter scale reduces the inset + grows the body" ./scene/ 'TestCardPaddingScale_TighterReducesInset'
run_check "default padding byte-identical"                   ./scene/ 'TestCardPaddingScale_DefaultByteIdentical'
run_check "extreme scale floors at padMin"                   ./scene/ 'TestCardPaddingScale_FloorsAtMin'
echo
echo "phase-45 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
