#!/usr/bin/env bash
#
# Phase 91 smoke — footnotes / sources / disclaimers (Deckard R14.12, engine).
# SceneSlide.Footnotes []RichText renders in a reserved bottom band (above the
# chrome footer) in the muted role; the body region shrinks to reserve it so
# footnotes never overlap the body or the page-number footer. A scene
# RunStyle.Superscript marks footnote references on figures/stats. Lines past a
# region cap are dropped + warned. Empty = byte-identical.
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
grep_check "SceneSlide.Footnotes field"  scene/scene.go    'Footnotes \[\]RichText'
grep_check "RunStyle.Superscript field"  scene/richtext.go 'Superscript bool'
grep_check "footnote band reservation"   scene/render.go   'footnoteBandHeight'
grep_check "renderFootnotes composer"    scene/render.go   'func (r \*renderer) renderFootnotes'
run_check  "footnotes render + superscript"  ./scene/ 'TestFootnotes$'
run_check  "empty footnotes byte-identical"  ./scene/ 'TestFootnotes_EmptyByteIdentical'
run_check  "footnote cap warns"              ./scene/ 'TestFootnotes_CapWarns'
run_check  "footnotes deterministic"         ./scene/ 'TestFootnotes_Deterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
