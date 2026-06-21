#!/usr/bin/env bash
#
# Phase 26 smoke — TwoColumn column join (Deckard R5 a+b). Verifies the optional
# centered seam element — a "VS"-style badge (JoinBadge + JoinLabel) or a
# connector arrow (JoinArrow) — renders, is omitted (byte-identical) when
# JoinNone, and is deterministic (docs/plans/phase-26-column-join.md §11/§13).

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

# 2. JoinBadge renders an ellipse + label (criterion 1).
run_check "center badge renders on the seam"         ./scene/ 'TestColumnJoin_Badge'
# 3. JoinArrow renders a connector arrow (criterion 2).
run_check "inter-column connector renders"           ./scene/ 'TestColumnJoin_Arrow'
# 4. JoinNone is byte-identical (criterion 3).
run_check "JoinNone is byte-identical"               ./scene/ 'TestColumnJoin_NoneByteIdentical'
# 5. Join render is deterministic across workers (criterion 4).
run_check "join render is deterministic"             ./scene/ 'TestColumnJoin_Deterministic'

echo
echo "phase-26 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
