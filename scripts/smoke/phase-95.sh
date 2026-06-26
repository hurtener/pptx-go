#!/usr/bin/env bash
#
# Phase 95 smoke — scatter ornament family (Deckard R14.20, engine half). Restates
# the starfield as a parameterized scatter FAMILY: one deterministic hash-of-index
# placement engine (shared with Starfield) parameterized by mark shape — scatter_dot
# / scatter_star / scatter_plus / scatter_ring — so a starfield, dust, confetti, or
# bokeh all draw from one recipe. starfield stays byte-identical (= scatter_dot).
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
grep_check "shared scatter engine"     assets/ornaments/patterns.go 'func scatter('
grep_check "ScatterStar recipe"        assets/ornaments/patterns.go 'func ScatterStar'
grep_check "scatter names registered"  scene/ornaments/registry.go  'NameScatterRing'
grep_check "scatter in Curated"        scene/ornaments/registry.go  'assetornaments.ScatterStar'
grep_check "scatter is pattern preset" scene/render_decoration.go   'NameScatterRing'
run_check  "curated set (11, incl scatter)"  ./scene/ornaments/ 'TestCurated_HasElevenOrnaments'
run_check  "scatter family shapes"           ./scene/ 'TestScatterFamily'
run_check  "starfield == scatter_dot"        ./scene/ 'TestScatter_StarfieldByteIdentical'
run_check  "scatter deterministic"           ./scene/ 'TestScatter_Deterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
