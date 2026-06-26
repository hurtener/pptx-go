#!/usr/bin/env bash
#
# Phase 88 smoke — native dataviz arcs: donut + gauge (Deckard R14.8, engine;
# part 2, completes R14.8). Adds a builder pptx.AddBlockArc(box, startDeg, sweepDeg,
# innerRatio, opts) — a native annular sector (<a:prstGeom prst="blockArc"> with
# adj1=start, adj2=end, adj3=inner-radius) — and two new DataMarkKind values:
# DataMarkDonut (a single-value ring + centered label, e.g. a 331° arc at 92%) and
# DataMarkGauge (a 270° speedometer). Both compose value + track arcs (no hole,
# no asset). Pure integer-EMU; deterministic. Catalog stays 30 (Kind values, not
# nodes). Asserts the API + the donut/gauge/blockArc round-trip + determinism tests.
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
grep_check "AddBlockArc builder method"   pptx/arc.go         'func (s \*Slide) AddBlockArc'
grep_check "DataMarkDonut kind"           scene/nodes.go      'DataMarkDonut'
grep_check "DataMarkGauge kind"           scene/nodes.go      'DataMarkGauge'
grep_check "renderDataMarkDonut composer" scene/render_datamark.go 'func (r \*renderer) renderDataMarkDonut'
grep_check "renderDataMarkGauge composer" scene/render_datamark.go 'func (r \*renderer) renderDataMarkGauge'
run_check  "donut renders blockArc + label"  ./scene/ 'TestDataMark_Donut$'
run_check  "gauge renders blockArc"           ./scene/ 'TestDataMark_Gauge$'
run_check  "donut full/empty edge cases"      ./scene/ 'TestDataMark_DonutFullAndEmpty'
run_check  "arc marks deterministic"          ./scene/ 'TestDataMark_ArcDeterministic'
run_check  "blockArc builder round-trips"     ./scene/ 'TestBlockArc_Builder'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
