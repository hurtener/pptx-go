#!/usr/bin/env bash
#
# Phase 72 smoke — radial-vignette-background (Deckard R13.2, engine). Adds a
# BackgroundRadial kind (appended last) that emits a center-out radial fill via
# pptx.RadialGradient (centered 50%-inset focal), reusing the Phase-71 stops via
# a shared backgroundGradientStopsFor resolver (multi-stop or legacy 2-role). The
# linear BackgroundGradient case is refactored through the same resolver
# (byte-identical). Center-only focal; the focal-offset knob is deferred (D-106).
# Asserts the kind/render path exist and the radial render/round-trip/validation
# tests pass (docs/plans/phase-72-radial-vignette-background.md §11/§13).
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
grep_check "BackgroundRadial kind"               scene/background.go 'BackgroundRadial'
grep_check "BackgroundRadial String case"        scene/background.go 'return "radial"'
grep_check "radial render case"                  scene/render.go     'pptx.RadialGradient(stops...)'
grep_check "shared stops resolver"               scene/render.go     'func backgroundGradientStopsFor'
run_check  "radial render + round-trip (Radial)"  ./scene/ 'TestBackground_Radial'
run_check  "radial invalid stops warn + skip"     ./scene/ 'TestBackground_RadialInvalidStopsWarn'
run_check  "radial deterministic"                 ./scene/ 'TestBackground_RadialDeterministic'
run_check  "BackgroundRadial String == radial"    ./scene/ 'TestBackgroundKind_RadialString'
run_check  "legacy 2-role linear byte-ident"      ./scene/ 'TestBackground_LegacyGradientByteIdentical'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
