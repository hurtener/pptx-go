#!/usr/bin/env bash
#
# Phase 52 smoke — content-region reserves chrome, R11.4 verify-and-close (Deckard
# R11.4). Verifies the chrome-aware body region is disjoint from the eyebrow and
# footer bands, a clamped container stays above the footer, and a chrome-off region
# is the plain margin box (byte-identical)
# (docs/plans/phase-52-content-region-reserves-chrome.md §11/§13).
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
run_check "chromed body region disjoint from both bands" ./scene/ 'TestBodyRegionReservesChrome'
run_check "chrome-off body region is the plain margin box" ./scene/ 'TestBodyRegionChromeOff_ByteIdentical'
run_check "clamped container stays above the footer band" ./scene/ 'TestClampedContainerStaysAboveFooter'
echo
echo "phase-52 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
