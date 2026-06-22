#!/usr/bin/env bash
#
# Phase 37 smoke — italic-aware font fallback (Deckard R9.7). Verifies an italic
# run whose family lacks an italic cut falls back to an italic-capable face while
# its upright runs keep the primary; an italic display run embeds the display
# italic cut (the D-063+D-065 guarantee); and the embedded <p:font> element is
# p:-prefixed (docs/plans/phase-37-emphasis-italic-display.md §11/§13).
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
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
run_check "italic run falls back, upright keeps primary" ./pptx/ 'TestEmphasisItalicFallback'
run_check "italic display run embeds the italic cut"     ./pptx/ 'TestEmphasisDisplayItalicEmbedded'
run_check "codec rewrite is italic-aware"                ./internal/ooxml/slide/ 'TestRewriteFontFacesItalicAware'
run_check "embedded <p:font> is p:-prefixed"             ./internal/ooxml/ 'TestRestoreEmbeddedFontList'
run_check "embedded face element p:-prefixed (deck)"     ./pptx/ 'TestAutoEmbedShipsUsedFaces'
echo
echo "phase-37 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
