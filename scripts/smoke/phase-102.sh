#!/usr/bin/env bash
#
# Phase 102 smoke — soul→engine roundtrip verification (Deckard R8.10, MED · both,
# engine; the Wave-15 capstone). Extends the Stats.Colors per-slide hook (D-058)
# with resolved SurfaceAlt/Accent/AccentAlt/TextAccent (kept comparable), captured
# from the variant theme so they dark-resolve; a fidelity test proves a full brand
# soul (light+dark+multi-accent+gradient) round-trips to the rendered colors.
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
grep_check "SlideColors gains resolved accent roles" scene/scene.go 'TextAccent  pptx.RGB'
run_check  "soul fidelity light + dark roundtrip"    ./scene/ 'TestSoulFidelity_LightAndDark'
run_check  "deliberate mismatch caught"              ./scene/ 'TestSoulFidelity_MismatchFails'
run_check  "extended SlideColors deterministic"      ./scene/ 'TestSoulFidelity_Deterministic'
run_check  "existing per-slide colors unaffected"    ./scene/ 'TestStatsColors_LightMatchesTheme|TestStatsColors_DarkPalette'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
