#!/usr/bin/env bash
#
# Phase 67 smoke — prim-icon-label-rows (Deckard R12.7). Adds an IconRows scene IR node: a
# vertical stack of [icon | label | optional right-aligned meta] rows with an optional
# RowPill frame and a Fill mode. Also carries a codec fix (read-side entity escaping in
# StripNamespacePrefixes) surfaced by the round-trip test. Asserts the rows render, the
# pill frames, an unknown icon fails validation, special chars round-trip, and the catalog
# count covers it (docs/plans/phase-67-prim-icon-label-rows.md §11/§13).
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
grep_check "KindIconRows in the node catalog"           scene/nodes.go   'KindIconRows'
grep_check "IconRows + IconRow + RowTone structs"       scene/nodes.go   'type IconRows struct'
grep_check "KindIconRows policy entry"                  scene/policy.go  'KindIconRows'
grep_check "IconRows Stage-1 validation"                scene/validate.go 'case IconRows'
grep_check "IconRows composer exists"                   scene/render_iconrows.go 'func (r \*renderer) renderIconRows'
grep_check "icon-row icons are Stage-1 validated"       scene/render_card.go 'icon row'
grep_check "read-side entity escaping fix"              internal/ooxml/xmlutils.go 'EscapeText'
run_check  "icon-rows render icons + labels + meta"     ./scene/ 'TestIconRows_RendersRows'
run_check  "RowPill frames the row"                     ./scene/ 'TestIconRows_RowPillFrame'
run_check  "unknown row icon fails Stage-1"             ./scene/ 'TestIconRows_UnknownIconFails'
run_check  "glyph color defaults to accent"            ./scene/ 'TestIconRowsGlyphColor'
run_check  "row heights are content-aware"             ./scene/ 'TestIconRowsRowHeights'
run_check  "icon-rows render is deterministic"         ./scene/ 'TestIconRows_Deterministic'
run_check  "special chars round-trip (& < >)"          ./internal/ooxml/ 'TestStripNamespacePrefixes_EscapesEntities'
run_check  "catalog covers KindIconRows (27 kinds)"     ./scene/ 'TestCatalog_KindsDistinct'
run_check  "integration round-trips every node"         ./test/integration/ 'TestRoundTrip_SceneNodes'
echo
echo "phase-67 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
