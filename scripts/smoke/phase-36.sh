#!/usr/bin/env bash
#
# Phase 36 smoke — font fallback chain (Deckard R9.6). Verifies a role's
# declared FontSpec.Fallback is realized at write time (the run's a:latin is
# rewritten to the first source-resolvable face when the primary is
# unavailable), the primary wins when available, byte-identical when unused, and
# deterministic/idempotent (docs/plans/phase-36-font-fallback-stack.md §11/§13).
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
run_check "substitutes an unavailable primary"    ./pptx/ 'TestFontFallbackSubstitutesUnavailablePrimary'
run_check "primary wins when available"           ./pptx/ 'TestFontFallbackPrimaryWinsWhenAvailable'
run_check "byte-identical when unused"            ./pptx/ 'TestFontFallbackByteIdenticalWhenUnused'
run_check "deterministic + idempotent"            ./pptx/ 'TestFontFallbackDeterministicIdempotent'
run_check "embeds the resolved fallback face"     ./pptx/ 'TestFontFallbackEmbedsResolvedFace'
run_check "codec rewrites matching faces"         ./internal/ooxml/slide/ 'TestRewriteFontFaces'
echo
echo "phase-36 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
