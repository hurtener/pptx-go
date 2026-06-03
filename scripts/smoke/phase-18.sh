#!/usr/bin/env bash
#
# Phase 18 smoke — round-trip read of self-authored decks (the navigable model).
#
# Delivered across 4 PRs (D-047): PR#1 shapes+props, PR#2 text, PR#3
# tables+images, PR#4 comprehensive test. Criteria flip to OK as the PRs land.
# Each criterion prints OK / SKIP / FAIL; done when OK >= count and FAIL == 0.

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

shapes_read() { grep -rq "func (s \*Slide) Shapes()" pptx/ 2>/dev/null; }

# 2. Slide.Shapes() reconstructs shape geometry/fill/line/shadow (PR#1).
if shapes_read; then
    if go test ./pptx/ ./test/pptx/ -run 'ShapeRead|ReadShape|RoundTripShape' >/dev/null 2>&1; then
        ok "Slide.Shapes() reconstructs shape geometry/fill/line/shadow"
    else
        fail "Slide.Shapes() reconstructs shape geometry/fill/line/shadow" "read tests failed"
    fi
else
    skip "Slide.Shapes() reconstructs shape geometry/fill/line/shadow" "read API not built yet (PR#1)"
fi

# 3. Text read round-trips runs/styles/links (PR#2).
if grep -rq "func (.*TextFrame) Paragraphs()" pptx/ 2>/dev/null; then
    if go test ./pptx/ ./test/pptx/ -run 'TextRead|ReadText' >/dev/null 2>&1; then
        ok "text read round-trips runs/styles/links"
    else
        fail "text read round-trips runs/styles/links" "text read tests failed"
    fi
else
    skip "text read round-trips runs/styles/links" "text read not built yet (PR#2)"
fi

# 4. Table + image read round-trip (PR#3).
if grep -rq "func (sh \*Shape) Table()" pptx/ 2>/dev/null; then
    if go test ./pptx/ ./test/pptx/ -run 'TableRead|ImageRead' >/dev/null 2>&1; then
        ok "table + image read round-trip"
    else
        fail "table + image read round-trip" "table/image read tests failed"
    fi
else
    skip "table + image read round-trip" "table/image read not built yet (PR#3)"
fi

# 5. Comprehensive round-trip test (PR#4).
if [ -f test/integration/roundtrip_test.go ]; then
    if go test ./test/integration/ -run 'RoundTrip' >/dev/null 2>&1; then
        ok "comprehensive round-trip test passes (every primitive + IR node)"
    else
        fail "comprehensive round-trip test passes" "roundtrip_test failed"
    fi
else
    skip "comprehensive round-trip test passes" "roundtrip_test.go not added yet (PR#4)"
fi

echo
echo "phase-18 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
