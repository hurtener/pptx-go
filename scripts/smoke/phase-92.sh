#!/usr/bin/env bash
#
# Phase 92 smoke — hierarchy / org-chart / tree (Deckard R14.10, engine). NEW
# scene Tree node: a root with children laid out as a balanced top-down (or
# left-right) tidy tree — leaves spread evenly, internal nodes centered over their
# leaf descendants, parent→child elbow connector edges, soul-styled node cards.
# Pure integer-EMU (worker-count deterministic); depth/breadth past the region
# clamp + warn. Catalog 32 -> 33.
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
grep_check "KindTree enum"           scene/nodes.go      'KindTree'
grep_check "Tree node struct"        scene/nodes.go      'type Tree struct'
grep_check "TreeNode struct"         scene/nodes.go      'type TreeNode struct'
grep_check "Tree policy entry"       scene/policy.go     'KindTree:'
grep_check "Tree validation"         scene/validate.go   'case Tree:'
grep_check "renderTree composer"     scene/render_tree.go 'func (r \*renderer) renderTree'
grep_check "walkTreeIcons recursion" scene/render_card.go 'func walkTreeIcons'
grep_check "Tree in catalog test"    scene/scene_test.go 'scene.Tree{'
run_check  "tree renders cards + edges"  ./scene/ 'TestTree$'
run_check  "left-right tree"             ./scene/ 'TestTree_LeftRight'
run_check  "tree deterministic"          ./scene/ 'TestTree_Deterministic'
run_check  "integration covers Tree"     ./test/integration/ 'TestRoundTrip_SceneNodes'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
