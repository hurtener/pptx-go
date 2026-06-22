#!/usr/bin/env bash
#
# Phase 42 smoke — card body vertical distribution (Deckard R10.4). Verifies the
# opt-in Card.BodyVAlign distributes the card body (bottom pins the last node),
# keeps the top-anchored default byte-identical, and renders deterministically
# across worker counts
# (docs/plans/phase-42-card-body-vertical-distribution.md §11/§13).
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
run_check "BodyVAlign=Bottom pins the last body node"   ./scene/ 'TestCardBodyVAlign_BottomPinsLastNode'
run_check "BodyVAlign=Top byte-identical to stackIn"    ./scene/ 'TestCardBodyVAlign_TopByteIdentical'
run_check "card BodyVAlign render stays deterministic"  ./scene/ 'TestCardBodyVAlign_Deterministic'
echo
echo "phase-42 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
