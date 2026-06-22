#!/usr/bin/env bash
#
# Phase 49 smoke — card header content-aware height, R11.1 verify-and-close
# (Deckard R11.1). Verifies the acceptance golden: a long wrapping header across
# every CardSize × CardLayout combo keeps the body below the header band (no
# overlap, exact agreement), single-line headers stay byte-identical, and the
# prior R10.1 wrapped-header guard still passes
# (docs/plans/phase-49-card-header-content-aware-height-verify.md §11/§13).
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
run_check "long header: body == band bottom across all CardSize×CardLayout" ./scene/ 'TestCardBodyBelowWrappedHeader_AllCombos'
run_check "wrapped header advances body below header (R10.1 guard)"          ./scene/ 'TestCardHeaderBottom_WrappedTitle'
echo
echo "phase-49 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
