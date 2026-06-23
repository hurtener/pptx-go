#!/usr/bin/env bash
#
# Phase 63 smoke — prim-chip-row-group (Deckard R12.5). Adds a ChipRow scene IR node: a
# greedy left-to-right wrap of content-fit chip pills with an optional leading label and
# per-line alignment. Asserts the node renders real pills (not bullets), wraps, fails on
# an unknown chip icon, and the catalog count + integration kind loop cover it
# (docs/plans/phase-63-prim-chip-row-group.md §11/§13).
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
grep_check "KindChipRow in the node catalog"            scene/nodes.go   'KindChipRow'
grep_check "ChipRow + ChipSpec structs"                 scene/nodes.go   'type ChipRow struct'
grep_check "KindChipRow policy entry"                   scene/policy.go  'KindChipRow'
grep_check "ChipRow Stage-1 validation"                 scene/validate.go 'case ChipRow'
grep_check "ChipRow composer + packer"                  scene/render_chiprow.go 'func chipRowLines'
grep_check "chip icons are Stage-1 validated"           scene/render_card.go 'case ChipRow'
run_check  "chips render as real pills (not bullets)"   ./scene/ 'TestChipRow_RendersPills'
run_check  "wrapping reflows within the width"          ./scene/ 'TestChipRow_Wraps'
run_check  "wrap packer keeps lines within the box"     ./scene/ 'TestChipRowLines_Wrap'
run_check  "leading label rides line 0"                 ./scene/ 'TestChipRowLines_LabelOnLine0'
run_check  "chip width is content-fit"                  ./scene/ 'TestChipWidthOf'
run_check  "unknown chip icon fails Stage-1"            ./scene/ 'TestChipRow_UnknownIconFails'
run_check  "chip row render is deterministic"           ./scene/ 'TestChipRow_Deterministic'
run_check  "catalog covers KindChipRow (25 kinds)"      ./scene/ 'TestCatalog_KindsDistinct'
run_check  "integration round-trips every node"         ./test/integration/ 'TestRoundTrip_SceneNodes'
echo
echo "phase-63 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
