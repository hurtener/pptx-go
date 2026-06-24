#!/usr/bin/env bash
#
# Phase 73 smoke — decoration-color-role (Deckard R13.5, CRITICAL · engine). Adds
# Decoration.Color *pptx.ColorRole (nil = ColorAccent, byte-identical) and threads
# a role pptx.ColorRole through the ornaments.Recipe signature + all 6 curated
# recipes (the accent() helper becomes roleFill(role, alpha)). Lets textures/glows
# be neutral grey, inverse-white, or any brand role. Asserts the signature/field
# exist and the color-override + nil-byte-identical tests pass
# (docs/plans/phase-73-decoration-color-role.md §11/§13).
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
grep_check "Recipe signature carries role"       scene/ornaments/registry.go 'role pptx.ColorRole'
grep_check "roleFill helper"                     assets/ornaments/ornaments.go 'func roleFill'
grep_check "Decoration.Color field"              scene/nodes.go 'Color \*pptx.ColorRole'
grep_check "render_decoration resolves role"     scene/render_decoration.go 'role := pptx.ColorAccent'
run_check  "color override (white != accent)"    ./scene/ 'TestDecoration_ColorRole'
run_check  "nil color byte-identical per preset" ./scene/ 'TestDecoration_ColorNilByteIdentical'
run_check  "curated ornaments still render"      ./scene/ 'TestDecoration_CuratedOrnaments'
run_check  "ornament recipes emit shapes"        ./assets/ornaments/ 'TestOrnamentRecipes_EmitShapes'
run_check  "ornament extension w/ role"          ./scene/ 'TestDecoration_OrnamentExtension'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
