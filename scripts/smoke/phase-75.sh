#!/usr/bin/env bash
#
# Phase 75 smoke — text-number-watermark-decoration (Deckard R13.9, engine). Adds a
# DecorationText kind + Decoration.Text / FontSize so a slide can carry an
# oversized, low-opacity ghost number/word behind the body. Reuses the
# Card.Watermark text-alpha pattern, Decoration.Color (D-107), and
# RunStyle.FontScale (>1 grows) for size. Native, decorative, LayerBackground,
# byte-identical when unused. Asserts the kind/fields/render-case exist and the
# watermark/validation/determinism tests pass
# (docs/plans/phase-75-text-watermark-decoration.md §11/§13).
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
grep_check "DecorationText kind"          scene/nodes.go           'DecorationText'
grep_check "Decoration.Text field"        scene/nodes.go           'Text string'
grep_check "Decoration.FontSize field"    scene/nodes.go           'FontSize float64'
grep_check "renderDecoration text case"   scene/render_decoration.go 'case DecorationText'
grep_check "DecorationText validation"    scene/validate.go        'text decoration requires text'
run_check  "text watermark run + low alpha" ./scene/ 'TestDecoration_TextWatermark$'
run_check  "empty text fails validation"    ./scene/ 'TestDecoration_TextWatermarkEmpty'
run_check  "text watermark deterministic"   ./scene/ 'TestDecoration_TextWatermarkDeterministic'
run_check  "curated decorations byte-ident" ./scene/ 'TestDecoration_ColorNilByteIdentical'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
