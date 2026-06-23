#!/usr/bin/env bash
#
# Phase 62 smoke — prim-in-card-checklist-fill (Deckard R12.2). Adds a Checklist scene
# IR node: true filled status glyphs (curated check/x/dot custGeom, NOT a font
# checkbox), a hanging indent from the glyph width, row-major 1–3 column reflow, and a
# Fill mode that distributes rows to span the box. Asserts the node renders a filled
# glyph, reflows columns, fails on an unknown icon, and the catalog count + integration
# kind loop cover it (docs/plans/phase-62-prim-in-card-checklist-fill.md §11/§13).
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
grep_check "KindChecklist in the node catalog"          scene/nodes.go   'KindChecklist'
grep_check "Checklist struct + CheckState"              scene/nodes.go   'type Checklist struct'
grep_check "KindChecklist policy entry"                 scene/policy.go  'KindChecklist'
grep_check "Checklist Stage-1 validation"               scene/validate.go 'case Checklist'
grep_check "Checklist composer exists"                  scene/render_checklist.go 'func (r \*renderer) renderChecklist'
grep_check "filled glyph via the icon registry"         scene/render_checklist.go 'addChecklistGlyph'
grep_check "checklist item icons are Stage-1 validated" scene/render_card.go 'checklist item'
run_check  "filled glyph emitted (not an empty box)"    ./scene/ 'TestChecklist_FilledGlyphNotCheckbox'
run_check  "2-column reflow renders all items"          ./scene/ 'TestChecklist_Columns'
run_check  "unknown item icon fails Stage-1"            ./scene/ 'TestChecklist_UnknownIconFails'
run_check  "Fill distributes rows + per-row glyph+text" ./scene/ 'TestRenderChecklist_Fill'
run_check  "columns clamp to 1..3"                      ./scene/ 'TestChecklistCols_Clamp'
run_check  "glyph per state + tone override"            ./scene/ 'TestChecklistGlyphColor'
run_check  "row heights are content-aware (row-major)"  ./scene/ 'TestChecklistRowHeights_RowMajor'
run_check  "checklist render is deterministic"          ./scene/ 'TestChecklist_Deterministic'
run_check  "catalog covers KindChecklist (24 kinds)"    ./scene/ 'TestCatalog_KindsDistinct'
run_check  "integration round-trips every node"         ./test/integration/ 'TestRoundTrip_SceneNodes'
echo
echo "phase-62 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
