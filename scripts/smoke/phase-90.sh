#!/usr/bin/env bash
#
# Phase 90 smoke — logo wall / customer grid (Deckard R14.7, engine). NEW scene
# LogoWall node: an N-up grid of logo assets, each contained (not cropped) +
# centered in its cell, optionally recolored to a uniform tone (mono/brand via
# the duotone blip effect) so a mixed set reads as one cohesive wall; optional
# caption. Asset-bearing; a missing logo warns + is skipped. Catalog 31 -> 32.
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
grep_check "KindLogoWall enum"        scene/nodes.go      'KindLogoWall'
grep_check "LogoWall node struct"     scene/nodes.go      'type LogoWall struct'
grep_check "LogoToneKind enum"        scene/nodes.go      'LogoToneMono'
grep_check "LogoWall policy entry"    scene/policy.go     'KindLogoWall:'
grep_check "LogoWall validation"      scene/validate.go   'case LogoWall:'
grep_check "renderLogoWall composer"  scene/render_logo_wall.go 'func (r \*renderer) renderLogoWall'
grep_check "LogoWall in catalog test" scene/scene_test.go 'scene.LogoWall{'
run_check  "logo wall renders pics + duotone" ./scene/ 'TestLogoWall$'
run_check  "none tone = no recolor"           ./scene/ 'TestLogoWall_NoneTone'
run_check  "missing logo warns"               ./scene/ 'TestLogoWall_MissingWarns'
run_check  "logo wall deterministic"          ./scene/ 'TestLogoWall_Deterministic'
run_check  "empty logo wall fails"            ./scene/ 'TestLogoWall_EmptyFails'
run_check  "integration covers LogoWall"      ./test/integration/ 'TestRoundTrip_SceneNodes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
