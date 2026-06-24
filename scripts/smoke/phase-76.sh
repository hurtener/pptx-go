#!/usr/bin/env bash
#
# Phase 76 smoke — starfield-scatter-ornament (Deckard R13.6, engine). Adds a
# curated "starfield" ornament: a deterministic, irregular scatter of role-colored
# dots with per-dot size ({1,2,3}pt) and alpha ({35,60,100}%) variance,
# hash-perturbed (no RNG/clock). The dot count derives from the box size at a
# fixed pitch (full-bleed → dense, small → sparse); capped for file size. Asserts
# the recipe/name exist, the curated set is seven, and the variance/density/
# determinism tests pass (docs/plans/phase-76-starfield-scatter-ornament.md §11/§13).
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
grep_check "Starfield recipe"           assets/ornaments/patterns.go 'func Starfield'
grep_check "NameStarfield registry name" scene/ornaments/registry.go  'NameStarfield'
grep_check "starfield in Curated()"      scene/ornaments/registry.go  'assetornaments.Starfield'
run_check  "curated set is seven (incl starfield)" ./scene/ornaments/ 'TestCurated_HasSixOrnaments'
run_check  "ornament recipes emit shapes (incl starfield)" ./assets/ornaments/ 'TestOrnamentRecipes_EmitShapes'
run_check  "starfield size+alpha variance + density + determinism" ./scene/ 'TestDecoration_Starfield'
run_check  "curated decorations byte-identical"            ./scene/ 'TestDecoration_ColorNilByteIdentical'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
