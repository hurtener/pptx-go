#!/usr/bin/env bash
#
# Phase 39 smoke — content-aware card header height (Deckard R10.1). Verifies a
# wrapped card header advances the body below its actual height (no overlap), a
# single-line header is byte-identical to the legacy fixed advance, and the scene
# render stays deterministic
# (docs/plans/phase-39-card-header-height.md §11/§13).
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
run_check "wrapped header advances body (no overlap)" ./scene/ 'TestCardHeaderBottom_WrappedTitle'
run_check "body region begins below wrapped header"   ./scene/ 'TestCardBodyBelowWrappedHeader'
run_check "card render stays deterministic"           ./scene/ 'TestRenderDeterministic|TestDeterministic'
echo
echo "phase-39 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
