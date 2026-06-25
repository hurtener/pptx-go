#!/usr/bin/env bash
#
# Phase 84 smoke — timeline / roadmap node (Deckard R14.4, engine). Adds a NEW
# scene Timeline IR node: a horizontal axis with milestones at caller-specified
# proportional positions, optional phase Bands behind the axis, and optional
# swimlanes (Lanes). Markers (accent dots or curated icons), the axis line, and
# staggered labels compose from native preset shapes (no media); integer-EMU
# layout is worker-count deterministic. It is a new node — unused means absent
# (the catalog grows 28 → 29). Asserts the node + wiring exist and the render,
# single-lane, validation, and determinism tests pass
# (docs/plans/phase-84-timeline.md §11/§13).
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
grep_check "KindTimeline enum"          scene/nodes.go      'KindTimeline'
grep_check "Timeline node struct"       scene/nodes.go      'type Timeline struct'
grep_check "Milestone struct"           scene/nodes.go      'type Milestone struct'
grep_check "Timeline policy entry"      scene/policy.go     'KindTimeline:'
grep_check "Timeline validation"        scene/validate.go   'case Timeline:'
grep_check "renderTimeline dispatch"    scene/render.go     'r.renderTimeline'
grep_check "renderTimeline composer"    scene/render_timeline.go 'func (r \*renderer) renderTimeline'
grep_check "walkIconRefs Timeline"      scene/render_card.go 'case Timeline:'
grep_check "Timeline in catalog test"   scene/scene_test.go 'scene.Timeline{'
run_check  "roadmap renders (bands+lanes)" ./scene/ 'TestTimeline_Roadmap'
run_check  "single-lane timeline"          ./scene/ 'TestTimeline_SingleLane'
run_check  "invalid position fails"        ./scene/ 'TestTimeline_InvalidWarns'
run_check  "timeline accent cycle"         ./scene/ 'TestTimelineAccent_Cycle'
run_check  "timeline preferred height"     ./scene/ 'TestTimelinePreferredHeight'
run_check  "timeline deterministic"        ./scene/ 'TestTimeline_Deterministic'
run_check  "integration covers Timeline"   ./test/integration/ 'TestRoundTrip_SceneNodes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
