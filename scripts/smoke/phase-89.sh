#!/usr/bin/env bash
#
# Phase 89 smoke — 2x2 quadrant / positioning map (Deckard R14.9, engine). NEW
# scene Quadrant node: labeled X/Y axes with low/high end captions, optional
# per-quadrant tint + title, and items plotted at (x,y) 0..1 (origin bottom-left).
# Axes, dividers, item dots, and labels are native shapes; pure integer-EMU
# (worker-count deterministic). Catalog 30 -> 31. Unused = absent.
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
grep_check "KindQuadrant enum"          scene/nodes.go      'KindQuadrant'
grep_check "Quadrant node struct"       scene/nodes.go      'type Quadrant struct'
grep_check "QuadrantItem struct"        scene/nodes.go      'type QuadrantItem struct'
grep_check "Quadrant policy entry"      scene/policy.go     'KindQuadrant:'
grep_check "Quadrant validation"        scene/validate.go   'case Quadrant:'
grep_check "renderQuadrant composer"    scene/render_quadrant.go 'func (r \*renderer) renderQuadrant'
grep_check "Quadrant in catalog test"   scene/scene_test.go 'scene.Quadrant{'
run_check  "quadrant renders axes+dots+tints" ./scene/ 'TestQuadrant$'
run_check  "quadrant invalid coord fails"     ./scene/ 'TestQuadrant_InvalidWarns'
run_check  "quadrant deterministic"           ./scene/ 'TestQuadrant_Deterministic'
run_check  "integration covers Quadrant"      ./test/integration/ 'TestRoundTrip_SceneNodes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
