#!/usr/bin/env bash
#
# Phase 60 smoke — adversarial content-fit fixtures (Deckard R11.12). Renders every
# component under hostile content and asserts the structural invariants: every box on
# the canvas, header band ≤ body top, fit-required text on one line, chrome contrast
# passes — and the suite is deterministic
# (docs/plans/phase-60-adversarial-content-fit-fixtures.md §11/§13).
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
run_check "every box stays on the slide canvas"         ./scene/ 'TestAdversarial_AllBoxesOnCanvas'
run_check "header band stays above the body"            ./scene/ 'TestAdversarial_HeaderBandBelowBody'
run_check "fit-required chrome text is one line"        ./scene/ 'TestAdversarial_FitTextOneLine'
run_check "chrome text clears the contrast minimum"     ./scene/ 'TestAdversarial_ContrastPasses'
run_check "adversarial fixture renders deterministically" ./scene/ 'TestAdversarial_Deterministic'
echo
echo "phase-60 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
