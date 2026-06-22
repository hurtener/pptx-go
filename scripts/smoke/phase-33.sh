#!/usr/bin/env bash
#
# Phase 33 smoke — display-face-role (Deckard R9.2). Verifies Theme.DisplayFont /
# WithDisplayFont put a distinct face on TypeDisplay (heading/body unchanged),
# order-independent with WithFonts, byte-identical when omitted, and that a
# display run renders with the display typeface
# (docs/plans/phase-33-display-font.md §11/§13).

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

run_check "DisplayFont resolves on TypeDisplay only"  ./pptx/ 'TestDisplayFont_Resolution|TestDisplayFont_OrderIndependent'
run_check "omitted DisplayFont inherits HeadingFont"  ./pptx/ 'TestDisplayFont_OmittedInheritsHeading'
run_check "display run renders with display typeface" ./pptx/ 'TestDisplayFont_Render'

echo
echo "phase-33 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
