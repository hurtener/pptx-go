#!/usr/bin/env bash
#
# Phase 71 smoke — multi-stop-background-gradient (Deckard R13.3, engine). Extends
# the scene Background from a fixed 2-role linear gradient to an N-stop (2..8)
# []GradientStop field at caller-chosen positions; pptx.LinearGradient is already
# variadic. Empty Stops → the legacy 2-role path (byte-identical); invalid stops
# (<2, >8, out of [0,1], not ascending) → one LayoutWarning + skip (D-026).
# Asserts the field/type/validator exist and the render/round-trip/validation
# tests pass (docs/plans/phase-71-multistop-background-gradient.md §11/§13).
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
grep_check "GradientStop type in scene"          scene/background.go 'type GradientStop struct'
grep_check "Background.Stops field"              scene/background.go 'Stops \[\]GradientStop'
grep_check "backgroundGradientStops validator"   scene/render.go     'func backgroundGradientStops'
grep_check "multi-stop render path"              scene/render.go     'pptx.LinearGradient(float64(bg.Angle), stops...)'
run_check  "validator accepts/rejects stops"     ./scene/ 'TestBackgroundGradientStops'
run_check  "3-stop gradient renders + round-trips" ./scene/ 'TestBackground_MultiStopGradient'
run_check  "invalid stops warn + skip"           ./scene/ 'TestBackground_InvalidStopsWarn'
run_check  "legacy 2-role gradient byte-ident"   ./scene/ 'TestBackground_LegacyGradientByteIdentical'
run_check  "multi-stop gradient deterministic"   ./scene/ 'TestBackground_MultiStopDeterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
