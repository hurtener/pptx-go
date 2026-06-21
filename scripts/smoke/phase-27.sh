#!/usr/bin/env bash
#
# Phase 27 smoke — row-labeled bento grid (Deckard R5 c). Verifies the new Bento
# container node: per-row left labels + variable-span cells on a shared column
# grid, Stage-1 validation, the catalog/round-trip coverage, and determinism
# (docs/plans/phase-27-bento-grid.md §11/§13).

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

# 2. Span geometry aligns columns; gutter reserved only when labelled (criteria 1, 2).
run_check "bento span widths + gutter geometry"      ./scene/ 'TestBentoGeometry'
# 3. Labels + cells render (criterion 1).
run_check "bento renders labels and cells"           ./scene/ 'TestBento_RendersLabelsAndCells'
# 4. Stage-1 rejects malformed bento (criterion 3).
run_check "bento Stage-1 validation"                 ./scene/ 'TestBento_Validation'
# 5. Catalog has 21 kinds; round-trip covers Bento (criterion 5).
run_check "catalog + policy cover Bento"             ./scene/ 'TestCatalog_KindsDistinct|TestPolicy_MatchesStructs'
run_check "every-node round-trip covers Bento"       ./test/integration/ 'TestRoundTrip_SceneNodes'
# 6. Bento render is deterministic across workers (criterion 6).
run_check "bento render is deterministic"            ./scene/ 'TestBento_Deterministic'

echo
echo "phase-27 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
