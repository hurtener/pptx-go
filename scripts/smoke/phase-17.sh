#!/usr/bin/env bash
#
# Phase 17 smoke — Chart (image-shape V1): pic + caption + aspect warning +
# ChartPlaceholder builder helper.
#
# Each criterion prints exactly one of OK / SKIP / FAIL. A phase is done only
# when OK >= count(criteria) and FAIL == 0 (CLAUDE.md §16).

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... 2>/dev/null; then
    ok "library builds CGo-free"
else
    fail "library builds CGo-free" "go build failed"
fi

# 2. ChartPlaceholder builder helper exists + round-trips.
if grep -rq "func.*ChartPlaceholder" pptx/ 2>/dev/null; then
    if go test ./test/pptx/ -run 'ChartPlaceholder' >/dev/null 2>&1; then
        ok "ChartPlaceholder emits a labeled slot that round-trips"
    else
        fail "ChartPlaceholder emits a labeled slot that round-trips" "tests failed"
    fi
else
    skip "ChartPlaceholder emits a labeled slot that round-trips" "not built yet"
fi

# 3. Chart composer renders pic + caption + aspect warning.
if grep -rq "func.*renderChart" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'Chart' >/dev/null 2>&1; then
        ok "chart renders pic + caption + aspect warning"
    else
        fail "chart renders pic + caption + aspect warning" "chart tests failed"
    fi
else
    skip "chart renders pic + caption + aspect warning" "chart composer not built yet"
fi

# 4. Chart render is byte-identical workers=1 vs N.
if grep -rq "func.*renderChart" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'ChartParallel' >/dev/null 2>&1; then
        ok "chart render is byte-identical workers=1 vs N"
    else
        fail "chart render is byte-identical workers=1 vs N" "parallel test failed"
    fi
else
    skip "chart render is byte-identical workers=1 vs N" "chart composer not built yet"
fi

echo
echo "phase-17 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
