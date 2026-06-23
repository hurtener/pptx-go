#!/usr/bin/env bash
#
# Phase 64 smoke — prim-callout-banner (Deckard R12.6). Adds a Banner scene IR node: a
# full-width RadiusLG filled strip with a leading icon + bold lead + body and optional
# right-aligned Trailing children (Stat/Button). Asserts the node renders a filled strip,
# embeds a trailing button, defaults the fill to accent, fails on an unknown icon, and
# the catalog count + integration kind loop cover it
# (docs/plans/phase-64-prim-callout-banner.md §11/§13).
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
grep_check "KindBanner in the node catalog"             scene/nodes.go   'KindBanner'
grep_check "Banner struct"                              scene/nodes.go   'type Banner struct'
grep_check "KindBanner policy entry"                    scene/policy.go  'KindBanner'
grep_check "Banner Stage-1 validation (Trailing)"       scene/validate.go 'case Banner'
grep_check "Banner composer exists"                     scene/render_banner.go 'func (r \*renderer) renderBanner'
grep_check "Banner icons walked (Icon + Trailing)"      scene/render_card.go 'case Banner'
grep_check "Banner Trailing walked for images"          scene/render_image.go 'case Banner'
run_check  "banner renders a filled strip + text"       ./scene/ 'TestBanner_FilledStrip'
run_check  "embedded trailing button renders"           ./scene/ 'TestBanner_TrailingButton'
run_check  "default fill is accent (not invisible)"     ./scene/ 'TestBanner_DefaultFillAccent'
run_check  "unknown banner/child icon fails Stage-1"    ./scene/ 'TestBanner_UnknownIconFails'
run_check  "fill default + text auto-contrast"          ./scene/ 'TestBannerFillRole'
run_check  "trailing reserves a right region"           ./scene/ 'TestBannerRegions'
run_check  "banner render is deterministic"             ./scene/ 'TestBanner_Deterministic'
run_check  "catalog covers KindBanner (26 kinds)"       ./scene/ 'TestCatalog_KindsDistinct'
run_check  "integration round-trips every node"         ./test/integration/ 'TestRoundTrip_SceneNodes'
echo
echo "phase-64 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
