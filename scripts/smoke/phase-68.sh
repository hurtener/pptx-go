#!/usr/bin/env bash
#
# Phase 68 smoke — prim-spanning-column-bridge (Deckard R12.8). Extends TwoColumn.Join with
# a JoinPosition: a horizontal accent bracket (spanning line + 2 stubs + a content-fit label
# pill) across the top/bottom of both columns, reserving a band so it spans above/below the
# content. Asserts the bridge renders, the seam default is byte-identical, the label stays
# one intact run, and a bad position fails validation
# (docs/plans/phase-68-prim-spanning-column-bridge.md §11/§13).
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
grep_check "JoinPosition in nodes"                      scene/nodes.go   'type JoinPosition int'
grep_check "TwoColumn.JoinPosition field"               scene/nodes.go   'JoinPosition JoinPosition'
grep_check "bridge band reserve"                        scene/render_container.go 'bridgeBandH'
grep_check "column-bridge composer"                     scene/render_container.go 'func (r \*renderer) renderColumnBridge'
grep_check "join position validation"                   scene/validate.go 'join position'
run_check  "top bridge renders label pill + content"    ./scene/ 'TestColumnBridge_TopRenders'
run_check  "bridge label stays one intact run"          ./scene/ 'TestColumnBridge_LabelIntact'
run_check  "bridge reserves a band"                     ./scene/ 'TestTwoColumn_BridgeReservesBand'
run_check  "per-config bridge shape counts"             ./scene/ 'TestColumnBridge_ShapeCount'
run_check  "out-of-range position fails validation"     ./scene/ 'TestColumnBridge_Validation'
run_check  "bridge render is deterministic"             ./scene/ 'TestColumnBridge_Deterministic'
echo
echo "phase-68 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
