#!/usr/bin/env bash
#
# Phase 80 smoke — rounded-image-and-shadow-framing (Deckard R13.11, engine). Adds
# (*pptx.Image).SetCornerRadius(RadiusRole) / SetElevation(ElevationRole) builder
# methods (thin wrappers over the existing applyCornerRadius/applyShadow on the pic
# spPr) + scene Image.CornerRadius / Elevation. renderImage applies them; zero
# tokens self-gate (RadiusNone leaves the pic rectangular, ElevationFlat emits no
# shadow) → byte-identical. Asserts the methods/fields exist and the round-trip /
# byte-identical tests pass (docs/plans/phase-80-image-framing.md §11/§13).
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
grep_check "Image.SetCornerRadius method"  pptx/media.go        'func (im \*Image) SetCornerRadius'
grep_check "Image.SetElevation method"     pptx/media.go        'func (im \*Image) SetElevation'
grep_check "scene Image.CornerRadius field" scene/nodes.go      'CornerRadius RadiusRole'
grep_check "renderImage applies framing"   scene/render_image.go 'img.SetCornerRadius'
run_check  "builder image framing round-trips" ./pptx/  'TestImageFraming_RoundTrip'
run_check  "zero-token framing byte-identical"  ./pptx/  'TestImageFraming_ZeroByteIdentical'
run_check  "scene image framing emits roundRect+shadow" ./scene/ 'TestRenderImage_Framing$'
run_check  "scene zero-token framing byte-identical"    ./scene/ 'TestRenderImage_FramingZeroByteIdentical'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
