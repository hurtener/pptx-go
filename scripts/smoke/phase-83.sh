#!/usr/bin/env bash
#
# Phase 83 smoke — styled table / comparison matrix (Deckard R14.3, engine half).
# Adds an additive scene Table.Style *TableStyle{HeaderFill, Zebra, HighlightCol,
# RowLabelCol, HeaderGroups []HeaderGroup} so a feature×plan comparison matrix
# renders a header band, zebra striping, a highlighted column, an emphasized
# row-label column, and grouped header spans — all token-driven, controlling every
# cell fill explicitly. A nil Style keeps the plain banded table (byte-identical).
# Cell-value glyphs (check/cross/dot/bar) are intentionally composed with a Bento
# of Checklist/IconRows cells (a native a:tbl cell holds only text, no shapes).
# Asserts the API/fields exist and the emit, grouped-header, byte-identity, and
# determinism tests pass (docs/plans/phase-83-styled-table.md §11/§13).
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
grep_check "scene Table.Style field"        scene/nodes.go    'Style \*TableStyle'
grep_check "TableStyle struct"              scene/nodes.go    'type TableStyle struct'
grep_check "HeaderGroup struct"             scene/nodes.go    'type HeaderGroup struct'
grep_check "renderStyledTable composer"     scene/render_table.go 'func (r \*renderer) renderStyledTable'
grep_check "styleBodyCell zebra/highlight"  scene/render_table.go 'func (r \*renderer) styleBodyCell'
run_check  "styled matrix emits band+zebra"  ./scene/ 'TestTableStyle_Emits'
run_check  "grouped header merges spans"     ./scene/ 'TestTableStyle_HeaderGroups'
run_check  "nil style byte-identical"        ./scene/ 'TestTableStyle_NilByteIdentical'
run_check  "styled table deterministic"      ./scene/ 'TestTableStyle_Deterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
