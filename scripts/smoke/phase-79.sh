#!/usr/bin/env bash
#
# Phase 79 smoke — focal-glow-behind-region (Deckard R13.10, engine). Adds
# Card.Backdrop *Decoration drawn behind the card's computed box (before its fill)
# via the existing renderDecoration with the card box as the region — a focal halo
# that tracks the card across any layout. A center-anchored, bleeding radial_glow
# becomes a halo behind the card. nil = byte-identical. Asserts the field/wiring
# exist and the z-order/nil tests pass
# (docs/plans/phase-79-focal-glow-backdrop.md §11/§13).
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
grep_check "Card.Backdrop field"          scene/nodes.go       'Backdrop \*Decoration'
grep_check "renderCard draws Backdrop"    scene/render_card.go 'v.Backdrop != nil'
run_check  "backdrop glow before card fill (z-order)" ./scene/ 'TestCardBackdrop$'
run_check  "nil backdrop byte-identical"  ./scene/ 'TestCardBackdrop_NilByteIdentical'
run_check  "card chrome still renders"    ./scene/ 'TestRenderCard$'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
