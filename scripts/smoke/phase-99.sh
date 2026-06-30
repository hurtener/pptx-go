#!/usr/bin/env bash
#
# Phase 99 smoke — named brand gradients (Deckard R8.5, HIGH · both, engine).
# A theme registers named brand gradients (pptx.GradientSpec via WithGradient) that
# a scene Background requests by name (Background.GradientName) and the renderer
# feeds to LinearGradient/RadialGradient. RGB stops pin an exact brand hue across
# variants; empty name = the legacy role-based gradient path (byte-identical).
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
grep_check "GradientSpec type"                 pptx/fill.go 'type GradientSpec struct'
grep_check "Theme.Gradients + WithGradient"    pptx/theme.go 'func WithGradient'
grep_check "Background.GradientName field"      scene/background.go 'GradientName string'
grep_check "drawNamedGradient resolver"        scene/render.go 'func (r \*renderer) drawNamedGradient'
run_check  "brand wash + linear vs radial"     ./scene/ 'TestNamedGradient_BrandWashRendered|TestNamedGradient_LinearAngle'
run_check  "missing + invalid stops warn"      ./scene/ 'TestNamedGradient_MissingWarns|TestNamedGradient_InvalidStopsWarn'
run_check  "empty name byte-identical"         ./scene/ 'TestNamedGradient_EmptyByteIdentical'
run_check  "named-gradient determinism"        ./scene/ 'TestNamedGradient_Determinism'
run_check  "WithGradient + Clone + concurrent"  ./pptx/  'TestWithGradient|TestGradientsConcurrentReuse'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
