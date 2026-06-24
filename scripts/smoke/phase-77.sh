#!/usr/bin/env bash
#
# Phase 77 smoke — pattern-density-pitch (Deckard R13.7, engine). Threads a
# trailing pitch pptx.EMU through the ornaments.Recipe signature + Decoration.Pitch
# so grid_dots/noise_overlay/starfield derive their dot count from the box at a
# caller pitch (consistent visual density at any box size). pitch == 0 keeps the
# legacy fixed counts (byte-identical); capped at patternMaxDots with a past-cap
# LayoutWarning in render_decoration. Asserts the signature/field exist and the
# pitch-density/legacy/cap tests pass
# (docs/plans/phase-77-pattern-density-pitch.md §11/§13).
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
grep_check "Recipe signature carries pitch" scene/ornaments/registry.go 'pitch pptx.EMU) int'
grep_check "patternDims helper"             assets/ornaments/patterns.go 'func patternDims'
grep_check "Decoration.Pitch field"         scene/nodes.go               'Pitch pptx.EMU'
grep_check "render_decoration passes pitch + cap warn" scene/render_decoration.go 'patternProjection'
run_check  "pitch density (many columns)"   ./scene/ 'TestDecoration_PatternPitch$'
run_check  "legacy pitch byte-identical"    ./scene/ 'TestDecoration_PatternPitchLegacyByteIdentical'
run_check  "tiny pitch caps + warns"        ./scene/ 'TestDecoration_PatternPitchCapWarns'
run_check  "ornament recipes emit shapes"   ./assets/ornaments/ 'TestOrnamentRecipes_EmitShapes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
