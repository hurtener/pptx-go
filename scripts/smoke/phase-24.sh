#!/usr/bin/env bash
#
# Phase 24 smoke — slide chrome (Deckard R3). Verifies opt-in chrome (top section
# eyebrow + bottom footer brand + N/total page number) drawn outside a shrunk
# body region, with page numbering derivation, brand text/asset, byte-identity
# when disabled, and determinism (docs/plans/phase-24-slide-chrome.md §11/§13).

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then
		skip "$desc" "not yet landed"
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then
		ok "$desc"
	else
		fail "$desc" "test failed"
	fi
}

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 2. Footer page number + per-slide section eyebrow render (criterion 1).
run_check "chrome renders footer + eyebrow"          ./scene/ 'TestChrome_FooterAndEyebrow'
# 3. Body region shrinks to make room for chrome (criterion 2).
run_check "chrome shrinks the body region"           ./scene/ 'TestBodyRegion_ChromeShrinks'
# 4. Total + page-number derivation and override (criterion 3).
run_check "page number / total derivation"           ./scene/ 'TestChrome_PageNumberOverrideAndTotal|TestChromeTotalFor'
# 5. Brand asset resolves; an unresolved one warns, not fails (criterion 4).
run_check "brand asset resolves and warns"           ./scene/ 'TestChrome_BrandAssetResolvesAndWarns'
# 6. Chrome disabled is byte-identical (criterion 5).
run_check "chrome disabled is byte-identical"        ./scene/ 'TestChrome_DisabledByteIdentical'
# 7. Chrome render is deterministic across workers (criterion 6).
run_check "chrome render is deterministic"           ./scene/ 'TestChrome_Deterministic'

echo
echo "phase-24 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
