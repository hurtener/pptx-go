#!/usr/bin/env bash
#
# Phase 35 smoke — automatic font-embedding pass (Deckard R9.1). Verifies
# WithFontEmbedding embeds the distinct used faces via the FontSource, is
# deterministic + byte-identical when off, idempotent vs manual EmbedFont, and
# warn-don't-fail on a missing face
# (docs/plans/phase-35-font-embedding-pipeline.md §11/§13).
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
run_check "embeds the distinct used faces"            ./pptx/ 'TestAutoEmbedShipsUsedFaces'
run_check "two embedded saves are byte-identical"     ./pptx/ 'TestAutoEmbedDeterministic'
run_check "flag off is byte-identical to baseline"    ./pptx/ 'TestAutoEmbedOffByteIdentical'
run_check "idempotent vs a manual EmbedFont"          ./pptx/ 'TestAutoEmbedIdempotentWithManual'
run_check "warn-don't-fail on a missing face"         ./pptx/ 'TestAutoEmbedWarnsOnMissing'
run_check "codec collects distinct used faces"        ./internal/ooxml/slide/ 'TestUsedFontFaces'
run_check "presentation dedup accessor"               ./internal/ooxml/presentation/ 'TestHasEmbeddedFace'
echo
echo "phase-35 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
