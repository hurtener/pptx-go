#!/usr/bin/env bash
#
# Phase 19 smoke — external-deck read robustness (best-effort).
#
# Delivered across 2 PRs (D-048): PR#1 the ReadWarnings reporting surface +
# dropped-element collection + part pass-through; PR#2 no-panic hardening + the
# synthetic external-deck corpus + fuzz seeds. Criteria flip to OK as the PRs
# land. Each criterion prints OK / SKIP / FAIL; done when OK >= count and
# FAIL == 0.

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

warnings_api() { grep -rq "func (p \*Presentation) ReadWarnings()" pptx/ 2>/dev/null; }

# 2. ReadWarnings surfaces a dropped element; an authored deck reports none (PR#1).
if warnings_api; then
    if go test ./pptx/ ./test/integration/ -run 'ReadWarning|ExternalRead' >/dev/null 2>&1; then
        ok "ReadWarnings surfaces dropped/unrecognized content; authored decks report none"
    else
        fail "ReadWarnings surfaces dropped/unrecognized content" "read-warning tests failed"
    fi
else
    skip "ReadWarnings surfaces dropped/unrecognized content" "ReadWarnings API not built yet (PR#1)"
fi

# 3. An external deck's unmodeled parts round-trip byte-for-byte (PR#1/PR#2).
if grep -rq "PartPassThrough\|PassThrough" test/integration/ pptx/ 2>/dev/null; then
    if go test ./test/integration/ ./pptx/ -run 'PassThrough' >/dev/null 2>&1; then
        ok "external unmodeled parts round-trip byte-for-byte"
    else
        fail "external unmodeled parts round-trip byte-for-byte" "pass-through test failed"
    fi
else
    skip "external unmodeled parts round-trip byte-for-byte" "pass-through test not added yet (PR#1)"
fi

# 4. The synthetic external-deck corpus loads without panic (PR#2).
if [ -f test/integration/external_read_test.go ]; then
    if go test ./test/integration/ -run 'ExternalRead' >/dev/null 2>&1; then
        ok "synthetic external-deck corpus loads without panic"
    else
        fail "synthetic external-deck corpus loads without panic" "external_read_test failed"
    fi
else
    skip "synthetic external-deck corpus loads without panic" "external corpus not added yet (PR#2)"
fi

echo
echo "phase-19 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
