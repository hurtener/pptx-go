#!/usr/bin/env bash
#
# Phase 11 smoke — Image node crop/fit (scene IR) + media seam consolidation
# (docs/plans/phase-11-image-node.md §13).
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

# 2. A non-zero Crop emits an srcRect; zero crop emits none (criterion 1).
run_check "scene Image crop emits srcRect" ./scene/ 'RenderImage_Crop|RenderImage_NoCrop'
# 3. FitNone omits the stretch fill; FitFill keeps it (criterion 2).
run_check "scene Image fit toggles stretch" ./scene/ 'RenderImage_Fit'
# 4. Crop + fit compose with a frame (criterion 3).
run_check "crop/fit compose with a frame" ./scene/ 'RenderImage_CropWithFrame'
# 5. Out-of-range / over-crop fails Stage-1 validation (criterion 4).
run_check "invalid crop fails Stage-1 validation" ./scene/ 'RenderImage_InvalidCrop'
# 6. Same asset twice → one media part at the scene seam (criterion 5).
run_check "scene-seam image dedup" ./scene/ 'RenderImage_SceneDedup'
# 7. Composite (frame+crop+fit+alt) round-trip + byte-identical + conformant (criterion 6).
run_check "composite image deck round-trip + determinism" ./test/integration/ 'CropFitComposite'

echo
echo "phase-11 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
