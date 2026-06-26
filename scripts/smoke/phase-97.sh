#!/usr/bin/env bash
#
# Phase 97 smoke — soul-driven dark palette (Deckard R8.3, CRITICAL · both, engine).
# A theme can carry an optional DarkPalette (Theme.DarkColors) so the scene
# renderer's VariantDark derivation overlays soul-driven surface/text overrides
# over its pinned Tailwind-gray default — a brand renders its own deep dark side,
# and a theme that sets no dark palette is byte-identical to today.
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
grep_check "DarkPalette type + Theme.DarkColors" pptx/theme.go 'type DarkPalette struct'
grep_check "WithDarkSurface / WithDarkText options" pptx/theme.go 'func WithDarkSurface'
grep_check "darkThemeFrom overlays base.DarkColors" scene/background.go 'base.DarkColors'
grep_check "Clone deep-copies DarkColors" pptx/theme.go 'if t.DarkColors != nil'
run_check  "soul dark palette resolves + emits"  ./scene/ 'TestDarkPalette_SoulDrivenColors'
run_check  "nil DarkColors byte-identical"       ./scene/ 'TestDarkPalette_NilByteIdentical'
run_check  "dark-palette determinism"            ./scene/ 'TestDarkPalette_Determinism'
run_check  "darkThemeFrom overlay/fallback"      ./scene/ 'TestDarkThemeFrom_'
run_check  "Clone deep-copy + concurrent reuse"  ./pptx/  'TestCloneDarkColorsIndependence|TestDarkColorsConcurrentReuse'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
