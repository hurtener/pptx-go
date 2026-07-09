#!/usr/bin/env bash
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
run_check "Grid Fill under VAlignTop fills body region" ./scene/ 'TestGridFillFillsBodyRegion'
run_check "Bento Fill under VAlignTop fills body region" ./scene/ 'TestBentoFillFillsBodyRegion'
run_check "Fill=false is byte-identical" ./scene/ 'TestGridBentoFillFalseByteIdentical'
run_check "worker-count determinism with Fill" ./scene/ 'TestRenderDeterministic_GridBentoFill'
if grep -q 'Fill bool' skills/compose-a-scene/SKILL.md && grep -q 'Fill (Grid/Bento)' docs/glossary.md; then ok "docs/skill updated for new field"; else fail "docs/skill updated for new field" "missing glossary or skill update"; fi
echo
echo "phase-104 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
