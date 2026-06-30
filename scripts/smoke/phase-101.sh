#!/usr/bin/env bash
#
# Phase 101 smoke — contrast-aware accent-text mechanism (Deckard R8.6, MED · both,
# engine). A pure exported scene.LegibleTextOn(fg, bg, minRatioX10) nudges an
# accent hue-preserving (lighten on a dark bg / darken on a light one) until it
# clears a target WCAG contrast ratio, reusing the existing contrast math. A caller
# mechanism (D-026): no auto-apply, no render change → byte-identical; the soul
# derives per-variant TextAccent with it and stores via WithDarkText.
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
grep_check() {
	local desc="$1" file="$2" pat="$3"
	if [ ! -f "$file" ]; then skip "$desc" "missing $file"
	elif grep -q "$pat" "$file"; then ok "$desc"
	else fail "$desc" "pattern not found: $pat"; fi
}
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
grep_check "LegibleTextOn exported mechanism"        scene/contrast.go 'func LegibleTextOn'
run_check  "clears target / lighten + darken paths"  ./scene/ 'TestLegibleTextOn_ClearsTargetPerBackground'
run_check  "already-legible unchanged"               ./scene/ 'TestLegibleTextOn_AlreadyLegibleUnchanged'
run_check  "pure + malformed fail-safe"              ./scene/ 'TestLegibleTextOn_Pure|TestLegibleTextOn_MalformedFailsSafe'
run_check  "hue preserved + target honored"          ./scene/ 'TestLegibleTextOn_DarkenPreservesHue|TestLegibleTextOn_DisplayTargetWeaker'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
