#!/usr/bin/env bash
#
# Phase 61 smoke — prim-cta-button (Deckard R12.1). Adds a Button scene IR node: a
# content-fit RadiusFull pill, tone→token fill (ghost = outline), pinned size scale,
# bold TypeBody label flanked by native custGeom icons, presentational only. Asserts
# the node renders, composes in a card body, fails on an unknown icon, and the catalog
# count + integration kind loop cover it
# (docs/plans/phase-61-prim-cta-button.md §11/§13).
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
grep_check "KindButton in the node catalog"             scene/nodes.go   'KindButton'
grep_check "Button struct + tone/size enums"            scene/nodes.go   'type Button struct'
grep_check "KindButton policy entry"                    scene/policy.go  'KindButton'
grep_check "Button Stage-1 validation"                  scene/validate.go 'case Button'
grep_check "Button render dispatch"                     scene/render.go  'renderButton'
grep_check "Button composer exists"                     scene/render_button.go 'func (r \*renderer) renderButton'
grep_check "Button icons are Stage-1 validated"         scene/render_card.go 'button leading'
run_check  "button renders as a pill + label + icon"    ./scene/ 'TestButton_RendersPill'
run_check  "button composes inside a card body"         ./scene/ 'TestButton_InCardBody'
run_check  "ghost button is an outline"                 ./scene/ 'TestButton_GhostIsOutline'
run_check  "unknown button icon fails Stage-1"          ./scene/ 'TestButton_UnknownIconFails'
run_check  "button width is content-fit + clamped"      ./scene/ 'TestButtonWidthOf_ContentFit'
run_check  "button size scale is monotonic"             ./scene/ 'TestButtonMetrics_SizeScale'
run_check  "button render is deterministic"             ./scene/ 'TestButton_Deterministic'
run_check  "catalog covers KindButton (23 kinds)"       ./scene/ 'TestCatalog_KindsDistinct'
run_check  "integration round-trips every node"         ./test/integration/ 'TestRoundTrip_SceneNodes'
echo
echo "phase-61 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
