#!/usr/bin/env bash
#
# Phase 38 smoke — weight-aware font embedding (Deckard R9.8). Verifies the
# embedding pass requests the actual resolved weight file per OOXML bucket (a
# medium 500 ships the medium file, not a synthetic 400), a single-weight deck
# embeds one file, and colliding weights coalesce to the nearest-nominal winner
# (docs/plans/phase-38-weight-aware-embedding.md §11/§13).
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
run_check "embeds the actual resolved weight file" ./pptx/ 'TestWeightAwareEmbedsResolvedWeightFile'
run_check "single-weight deck embeds one file"     ./pptx/ 'TestWeightAwareSingleWeightOneFile'
run_check "colliding weights coalesce per bucket"  ./pptx/ 'TestWeightAwareCoalescesBucket'
run_check "codec carries per-run weight"           ./internal/ooxml/slide/ 'TestUsedFontFacesCarriesWeight'
echo
echo "phase-38 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
