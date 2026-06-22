#!/usr/bin/env bash
#
# Phase 48 smoke — estimate/actual parity (Deckard R10.10). Verifies the Card slot
# estimate grows with a wrapped multi-line header, the Bento estimate measures
# each cell at its actual span width, and single-line / span-1 cases stay
# byte-identical (docs/plans/phase-48-estimate-actual-parity.md §11/§13).
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
run_check "wrapped-header card estimate grows"              ./scene/ 'TestPreferredHeight_WrappedCardGrows'
run_check "single-line card estimate byte-identical"       ./scene/ 'TestPreferredHeight_SingleLineCardUnchanged'
run_check "wide-span bento estimate uses span width"       ./scene/ 'TestPreferredHeight_BentoSpanWidth'
run_check "span-1 bento estimate byte-identical"           ./scene/ 'TestPreferredHeight_BentoSpanOneByteIdentical'
run_check "wrapped-header card overflow warning fires"     ./scene/ 'TestOverflow_WrappedHeaderCardFires'
echo
echo "phase-48 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
