#!/usr/bin/env bash
#
# Phase 93 smoke — funnel + cycle non-linear process diagrams (Deckard R14.11,
# engine). Two NEW scene nodes: Funnel (a vertical stack of bands tapering in
# width, optional per-stage values) and Cycle (stages placed evenly on a ring with
# directional connector arrows). Branch (1→M) is covered by the Tree node (R14.10).
# Native preset shapes; pure integer-EMU (worker-count deterministic). Catalog
# 33 -> 35.
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
grep_check "KindFunnel/KindCycle enum"  scene/nodes.go      'KindCycle'
grep_check "Funnel node struct"         scene/nodes.go      'type Funnel struct'
grep_check "Cycle node struct"          scene/nodes.go      'type Cycle struct'
grep_check "renderFunnel composer"      scene/render_funnel_cycle.go 'func (r \*renderer) renderFunnel'
grep_check "renderCycle composer"       scene/render_funnel_cycle.go 'func (r \*renderer) renderCycle'
grep_check "Funnel/Cycle policy"        scene/policy.go     'KindFunnel:'
grep_check "Funnel/Cycle validation"    scene/validate.go   'case Funnel:'
grep_check "Funnel in catalog test"     scene/scene_test.go 'scene.Funnel{'
run_check  "funnel tapers + values"        ./scene/ 'TestFunnel$'
run_check  "cycle ring + arrows"           ./scene/ 'TestCycle$'
run_check  "empty funnel/cycle fail"       ./scene/ 'TestFunnelCycle_InvalidWarns'
run_check  "funnel/cycle deterministic"    ./scene/ 'TestFunnelCycle_Deterministic'
run_check  "integration covers Funnel/Cycle" ./test/integration/ 'TestRoundTrip_SceneNodes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
