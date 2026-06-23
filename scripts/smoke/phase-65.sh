#!/usr/bin/env bash
#
# Phase 65 smoke — prim-ribbon-corner-badge (Deckard R12.3). Adds a Card.Ribbon field: a
# pinned emphasis badge outside the header text flow. RibbonTopBar reserves a band so the
# body shifts down; RibbonCornerStar is a star glyph; RibbonCornerTL/TR are corner text
# tabs. Asserts the top bar shifts the body, the star emits a glyph, a ribbon-free card is
# byte-identical, and an out-of-range position fails validation
# (docs/plans/phase-65-prim-ribbon-corner-badge.md §11/§13).
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
grep_check "Ribbon + RibbonPos in nodes"                scene/nodes.go   'type Ribbon struct'
grep_check "Card.Ribbon field"                          scene/nodes.go   'Ribbon \*Ribbon'
grep_check "top-bar band reserve"                       scene/render_card.go 'func ribbonReserveOf'
grep_check "ribbon composer"                            scene/render_card.go 'func (r \*renderer) renderCardRibbon'
grep_check "ribbon position validation"                 scene/validate.go 'ribbon position'
run_check  "top bar shifts the body down"               ./scene/ 'TestRibbonTopBarShiftsBody'
run_check  "reserve is top-bar only"                    ./scene/ 'TestRibbonReserve'
run_check  "color + auto-contrast text resolution"      ./scene/ 'TestRibbonColors'
run_check  "per-position shape counts"                  ./scene/ 'TestRenderCardRibbon_ShapeCount'
run_check  "top bar renders its label"                  ./scene/ 'TestRibbon_TopBarText'
run_check  "corner star emits a glyph"                  ./scene/ 'TestRibbon_CornerStar'
run_check  "ribbon-free card is additive/stable"        ./scene/ 'TestRibbon_NilDistinct'
run_check  "out-of-range position fails validation"     ./scene/ 'TestRibbon_PositionValidation'
run_check  "ribbon render is deterministic"             ./scene/ 'TestRibbon_Deterministic'
echo
echo "phase-65 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
