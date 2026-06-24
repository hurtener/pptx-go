#!/usr/bin/env bash
#
# Phase 70 smoke — tinted-paper-canvas (Deckard R13.1, engine half). Adds a
# ColorPaper surface token (appended last to the ColorRole iota) + a WithPaper
# theme option for a faintly tinted off-white "paper" canvas, defaulting to
# ColorCanvas (white) so it is byte-identical until a theme overrides the tint.
# Asserts the token exists, the default == canvas, WithPaper sets it, and a
# ColorPaper background's resolved RGB survives a round-trip
# (docs/plans/phase-70-tinted-paper-canvas.md §11/§13).
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
grep_check "ColorPaper in the ColorRole iota"   pptx/theme.go 'ColorPaper'
grep_check "WithPaper theme option"             pptx/theme.go 'func WithPaper'
grep_check "ColorPaper default in DefaultTheme" pptx/theme.go 'ColorPaper:      "FFFFFF"'
run_check  "ColorPaper default resolves to canvas" ./pptx/ 'TestColorPaperDefaultIsCanvas'
run_check  "WithPaper sets + Clone carries the tint" ./pptx/ 'TestWithPaper'
run_check  "every role (incl ColorPaper) resolves"   ./pptx/ 'TestDefaultThemeResolvesEveryRole'
run_check  "ColorPaper background round-trips"       ./scene/ 'TestBackground_PaperRoundTrip'
run_check  "default ColorPaper bg byte-identical"    ./scene/ 'TestBackground_PaperDefaultByteIdentical'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
