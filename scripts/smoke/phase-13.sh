#!/usr/bin/env bash
#
# Phase 13 smoke — Ornaments + Decoration + builder foundations (gradients,
# rotation, token-alpha, core metadata, WithLogger). Two-PR phase
# (docs/plans/phase-13-ornaments-decoration.md §13); checks SKIP until each
# surface lands.

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

# --- PR #1: builder foundations + carried fixes ---------------------------
run_check "gradient fill round-trip (linear + radial)" ./test/pptx/ 'Gradient'
run_check "WithRotation sets xfrm rot + round-trips"    ./test/pptx/ 'Rotation'
run_check "TokenColorAlpha emits token + alpha"         ./test/pptx/ 'TokenAlpha|TokenColorAlpha'
run_check "SetMetadata writes core.xml + round-trips"   ./test/pptx/ 'Metadata|SetMetadata'
run_check "scene Meta reaches core.xml"                 ./scene/ 'MetaCoreProps|Meta'
run_check "pptx.WithLogger emits a save event"          ./test/pptx/ 'WithLogger|LoggerEmits'

# --- PR #2: ornaments + Decoration ----------------------------------------
run_check "each curated ornament renders at its anchor" ./scene/ 'Ornament|Decoration'
run_check "bleed decoration uses negative offsets"      ./scene/ 'Bleed'
run_check "layer z-order honored (bg before / fg after)" ./scene/ 'LayerZOrder|DecorationLayer'
run_check "WithOrnamentExtension renders a caller ornament" ./scene/ 'OrnamentExtension'
run_check "ornament recipes deterministic"              ./assets/ornaments/ 'Recipe|Determinis'

echo
echo "phase-13 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
