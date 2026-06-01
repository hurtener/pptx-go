#!/usr/bin/env bash
#
# Phase 10 smoke — Frame chrome: curated device frames + WithFrameExtension
# (docs/plans/phase-10-frame-chrome.md §13).
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

# 2. Each curated frame places the image inside the bezel interior (criterion 1).
run_check "curated frames render image inside interior" ./scene/ 'CuratedFrame|FrameInterior'
# 3. A caller-extended frame renders via WithFrameExtension (criterion 2).
run_check "WithFrameExtension renders a caller frame" ./scene/ 'FrameExtension|WithFrameExtension'
# 4. FrameNone renders no bezel — back-compat (criterion 3).
run_check "FrameNone renders no bezel (back-compat)" ./scene/ 'FrameNone|UnframedImage'
# 5. Unknown FrameName is a Stage-1 validation error (criterion 4).
run_check "unknown FrameName fails Stage-1 validation" ./scene/ 'UnknownFrame|FrameValidation'
# 6. assets/frames recipes: interior within region + deterministic (criterion 1).
run_check "frame recipes: interior within region, deterministic" ./assets/frames/ 'Recipe|Interior|Determinis'
# 7. Framed-image deck round-trips + byte-identical + conformant (criterion 5).
run_check "framed-image deck round-trip + determinism + conformance" ./test/integration/ 'FrameImage|FramedImage'

echo
echo "phase-10 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
