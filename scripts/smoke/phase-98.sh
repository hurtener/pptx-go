#!/usr/bin/env bash
#
# Phase 98 smoke — multi-accent brand palette (Deckard R8.4, HIGH · both, engine).
# An optional Theme.Accents []RGB ordered brand palette that the scene renderer's
# per-element accent cycle reads by index, so a deck renders 4+ coordinated accent
# hues across timeline/funnel/cycle/quadrant/tree/image-pin markers — beyond the
# three accent roles. Empty Accents keeps the pinned five-role cycle (byte-ident).
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
grep_check "Theme.Accents + WithAccents"           pptx/theme.go 'func WithAccents'
grep_check "accentColorAt / accentRGBAt resolvers" scene/render_timeline.go 'func (r \*renderer) accentColorAt'
grep_check "onSurfaceRGB contrast core"            scene/contrast.go 'func (r \*renderer) onSurfaceRGB'
grep_check "cellTextOnColor contrast core"         scene/render_table.go 'func (r \*renderer) cellTextOnColor'
run_check  "brand hues rendered + nil byte-ident"  ./scene/ 'TestMultiAccent_BrandHuesRendered|TestMultiAccent_NilByteIdentical'
run_check  "contrast from a literal hue"           ./scene/ 'TestMultiAccent_ContrastFromHue'
run_check  "multi-accent determinism"              ./scene/ 'TestMultiAccent_Determinism'
run_check  "resolver cycle + pinned fallback"      ./scene/ 'TestAccentResolvers_'
run_check  "WithAccents + Clone + concurrent reuse" ./pptx/  'TestWithAccents|TestAccentsConcurrentReuse'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
