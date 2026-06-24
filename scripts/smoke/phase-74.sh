#!/usr/bin/env bash
#
# Phase 74 smoke — surface-fill-gradient (Deckard R13.8, engine). Adds a
# GradientFill{From,To pptx.ColorRole; Angle int} type + Card.FillGradient
# *GradientFill so a card surface can carry a 2-stop top-to-bottom depth shift via
# pptx.LinearGradient. nil = the solid Fill (byte-identical). Both stops are
# explicit token roles (a darker-To auto-tint is the soul's, D-026). Asserts the
# type/field/render-branch exist and the gradient/solid/determinism tests pass
# (docs/plans/phase-74-surface-fill-gradient.md §11/§13).
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
grep_check "GradientFill type"            scene/nodes.go       'type GradientFill struct'
grep_check "Card.FillGradient field"      scene/nodes.go       'FillGradient \*GradientFill'
grep_check "cardChrome.fillGradient"      scene/render_card.go 'fillGradient \*GradientFill'
grep_check "renderCardChrome gradient branch" scene/render_card.go 'c.fillGradient != nil'
run_check  "gradient card emits gradFill" ./scene/ 'TestCardFillGradient$'
run_check  "solid card byte-identical"    ./scene/ 'TestCardFillGradient_NilByteIdentical'
run_check  "gradient card deterministic"  ./scene/ 'TestCardFillGradient_Deterministic'
run_check  "card chrome still renders"    ./scene/ 'TestRenderCard$'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
