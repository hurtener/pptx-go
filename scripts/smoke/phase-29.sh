#!/usr/bin/env bash
#
# Phase 29 smoke — resolved per-slide colors (Deckard R7, completes Wave 8).
# Verifies Stats.Colors reports per-slide canvas/surface/primary-text the engine
# rendered with (the derived dark palette for VariantDark), in scene order, and
# deterministically (docs/plans/phase-29-resolved-colors.md §11/§13).

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then
		skip "$desc" "not yet landed"
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then
		ok "$desc"
	else
		fail "$desc" "test failed"
	fi
}

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 2. Stats.Colors: one scene-ordered entry per slide, light matches theme (crit 1, 2).
run_check "per-slide colors in scene order"          ./scene/ 'TestStatsColors_PerSlideSceneOrder|TestStatsColors_LightMatchesTheme'
# 3. Dark-variant colors differ from light + match the derived dark palette (crit 2).
run_check "dark-variant resolved palette exposed"    ./scene/ 'TestStatsColors_DarkPalette'
# 4. Colors is deterministic across workers (crit 3).
run_check "resolved colors are deterministic"        ./scene/ 'TestStatsColors_Deterministic'

echo
echo "phase-29 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
