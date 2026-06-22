#!/usr/bin/env bash
#
# Phase 59 smoke — decoration/watermark anti-collision, R11.11 verify-and-close
# (Deckard R11.11). Verifies the card watermark is emitted before the body (z-order
# behind) at low ~13% opacity, and is inert when unset
# (docs/plans/phase-59-decoration-watermark-anticollision.md §11/§13).
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
run_check "watermark emitted behind the body (z-order)"  ./scene/ 'TestWatermark_BehindBody'
run_check "watermark carries a low ~13% alpha"           ./scene/ 'TestWatermark_LowAlpha'
run_check "no watermark set emits no alpha run (inert)"  ./scene/ 'TestWatermark_OmittedWhenUnset'
echo
echo "phase-59 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
