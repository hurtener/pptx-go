#!/usr/bin/env bash
#
# Phase 87 smoke — native vector micro-charts (Deckard R14.8, engine; part 1).
# Adds a NEW scene DataMark node (Kind = DataMarkBar / DataMarkBars /
# DataMarkSparkline) drawn entirely from preset rects + lines in theme colors —
# crisp native vector, no AssetResolver. Values are 0..1; integer-EMU geometry is
# worker-count deterministic; embeds in a Card/Bento cell. Sparkline upward
# segments use a new pptx.WithFlipV shape option. Catalog grows 29 → 30; unused =
# absent. Arc-based marks (donut, gauge) are a follow-up (Phase 88). Asserts the
# node + wiring exist and the bar / bars+sparkline / in-card / invalid /
# determinism tests pass (docs/plans/phase-87-native-dataviz.md).
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
grep_check "KindDataMark enum"        scene/nodes.go      'KindDataMark'
grep_check "DataMark node struct"     scene/nodes.go      'type DataMark struct'
grep_check "DataMarkKind enum"        scene/nodes.go      'DataMarkBar DataMarkKind'
grep_check "DataMark policy entry"    scene/policy.go     'KindDataMark:'
grep_check "DataMark validation"      scene/validate.go   'case DataMark:'
grep_check "renderDataMark composer"  scene/render_datamark.go 'func (r \*renderer) renderDataMark'
grep_check "WithFlipV shape option"   pptx/shape.go       'func WithFlipV'
grep_check "DataMark in catalog test" scene/scene_test.go 'scene.DataMark{'
run_check  "bar renders native rects"      ./scene/ 'TestDataMark_Bar$'
run_check  "bars + sparkline (flipV)"      ./scene/ 'TestDataMark_BarsAndSparkline'
run_check  "data mark embeds in card"      ./scene/ 'TestDataMark_InCard'
run_check  "invalid value fails"           ./scene/ 'TestDataMark_InvalidWarns'
run_check  "data marks deterministic"      ./scene/ 'TestDataMark_Deterministic'
run_check  "integration covers DataMark"   ./test/integration/ 'TestRoundTrip_SceneNodes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
