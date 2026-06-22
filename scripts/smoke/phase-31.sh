#!/usr/bin/env bash
#
# Phase 31 smoke — line-height / leading token (Deckard R9.4). Verifies
# FontSpec.LineHeight / ParagraphOpts.LineHeight emit OOXML a:pPr/a:lnSpc/a:spcPct
# (a:-prefixed), are byte-identical when 0/100, round-trip via Paragraph.
# LineHeight(), and that the scene applies a role's leading to its paragraphs
# (docs/plans/phase-31-line-height.md §11/§13).

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

# 2. Line-height emits a:lnSpc/a:spcPct (criterion 1).
run_check "line-height emits a:lnSpc/a:spcPct"      ./pptx/ 'TestLineHeight_EmitsLnSpc'
# 3. Zero / single line-height is byte-identical (criterion 2).
run_check "zero/single line-height byte-identical"  ./pptx/ 'TestLineHeight_ZeroAndSingleByteIdentical'
# 4. Line-height round-trips via Paragraph.LineHeight() (criterion 3).
run_check "line-height round-trips"                 ./pptx/ 'TestLineHeight_RoundTrips'
# 5. Scene applies a role's leading to its paragraphs; default theme unchanged.
run_check "scene applies role line-height"          ./scene/ 'TestSceneLineHeight_RoleDriven'

echo
echo "phase-31 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
