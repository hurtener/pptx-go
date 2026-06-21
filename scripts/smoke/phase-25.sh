#!/usr/bin/env bash
#
# Phase 25 smoke — rich card visuals (Deckard R4). Verifies the three additive
# Card visuals — colored header band (HeaderFill), top-right status dot
# (StatusDot), and ghosted watermark (Watermark) — each opt-in, omitted when
# unset, byte-identical when none are set, and deterministic
# (docs/plans/phase-25-rich-card-visuals.md §11/§13).

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

# 2. Header band + status dot + watermark all render (criterion 1).
run_check "header band + status dot + watermark"     ./scene/ 'TestCardRichVisuals_AllThree'
# 3. Each visual is omitted when its field is unset (criterion 2).
run_check "visuals omitted when unset"               ./scene/ 'TestCardRichVisuals_OmittedWhenUnset|TestCardRichVisuals_IndividuallyOptional'
# 4. A bare card stays byte-identical (criterion 3): the header-row constant
#    extraction is value-identical (TestCardHeaderConstants), plus the existing
#    parallel/render idempotency guards.
run_check "bare card byte-identical / round-trips"   ./scene/ 'TestCardHeaderConstants|TestCardParallel|TestRenderCard$'
# 5. Rich-card render is deterministic across workers (criterion 4).
run_check "rich-card render is deterministic"        ./scene/ 'TestCardRichVisuals_Deterministic'

echo
echo "phase-25 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
