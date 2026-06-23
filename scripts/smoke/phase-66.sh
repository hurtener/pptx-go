#!/usr/bin/env bash
#
# Phase 66 smoke — prim-inter-column-connectors (Deckard R12.4). Adds Grid.Connectors: a
# connector glyph drawn in the gutter between two adjacent columns, reusing the Flow
# renderConnector + a new ConnectorBiArrow. Asserts the glyphs render, the bi-arrow emits
# a leftRightArrow, connectors are additive, and bad indices fail validation
# (docs/plans/phase-66-prim-inter-column-connectors.md §11/§13).
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
grep_check "ConnectorBiArrow kind"                      scene/nodes.go   'ConnectorBiArrow'
grep_check "GridConnector + Grid.Connectors"            scene/nodes.go   'type GridConnector struct'
grep_check "ConnectorBiArrow render case"               scene/render_flow.go 'leftRightArrow'
grep_check "grid connector composer"                    scene/render_container.go 'func (r \*renderer) renderGridConnectors'
grep_check "grid connector validation"                  scene/validate.go 'grid connector'
run_check  "connectors render gutter glyphs + label"    ./scene/ 'TestGridConnectors_RenderGlyphs'
run_check  "connectors are additive"                    ./scene/ 'TestGridConnectors_Additive'
run_check  "bad connector indices fail validation"      ./scene/ 'TestGridConnectors_Validation'
run_check  "connector render is deterministic"          ./scene/ 'TestGridConnectors_Deterministic'
echo
echo "phase-66 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
