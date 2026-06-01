#!/usr/bin/env bash
#
# Phase 12 smoke — Curated icons: SVG→custGeom translator + icon registry
# (docs/plans/phase-12-curated-icons.md §13).
#
# SKIPs gracefully until the surface lands; FAIL blocks the merge.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# run_check SKIPs when no test matches the pattern yet (surface not built).
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

# 2. AddIcon emits a custGeom path shape with accent fill (criterion 1).
run_check "AddIcon emits custGeom + accent fill" ./test/pptx/ 'AddIcon_EmitsCustGeom'
# 3. custGeom shape round-trips + byte-identical (criteria 2/6).
run_check "icon custGeom round-trip + determinism" ./test/pptx/ 'AddIcon_RoundTrip'
# 4. custGeom wire types marshal + round-trip (criterion 2).
run_check "custGeom wire round-trip" ./internal/ooxml/slide/ 'CustomGeometry'
# 5. SVG translator: subset + rejections + determinism (criteria 1/4).
run_check "SVG translator subset + rejections" ./internal/render/ 'Translate'
# 6. WithIconExtension valid registers; invalid fails at registration (criteria 3/4).
run_check "WithIconExtension valid + invalid-at-registration" ./scene/ 'WithIconExtension|ValidateIcon'
# 7. Every embedded curated icon translates (criterion 5).
run_check "every curated icon translates" ./assets/icons/ 'AllTranslate'

echo
echo "phase-12 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
