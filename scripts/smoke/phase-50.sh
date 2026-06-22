#!/usr/bin/env bash
#
# Phase 50 smoke — card-text auto-contrast (Deckard R11.2). Verifies the
# onCardSurface mechanism: chrome text on any surface clears the contrast minimum,
# light cards stay byte-identical, dark cards / dark variants flip the header to a
# light color, the eyebrow drops a same-hue accent, and the path is deterministic
# across worker counts (docs/plans/phase-50-card-text-auto-contrast.md §11/§13).
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
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
run_check "every surface clears the contrast minimum"      ./scene/ 'TestOnCardSurface_ContrastGuarantee'
run_check "light surface yields nil (byte-identical)"      ./scene/ 'TestOnCardSurface_LightIsNil'
run_check "accent legible on white, illegible same-hue"    ./scene/ 'TestAccentLegible'
run_check "dark-variant card header is white"              ./scene/ 'TestCardHeader_AutoContrast_DarkVariant'
run_check "light card header has no explicit color"       ./scene/ 'TestCardHeader_AutoContrast_LightByteIdentical'
run_check "eyebrow drops a same-hue accent band"           ./scene/ 'TestCardEyebrow_AccentFallback'
run_check "auto-contrast deterministic across workers"     ./scene/ 'TestRenderDeterministic_AutoContrast'
echo
echo "phase-50 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
