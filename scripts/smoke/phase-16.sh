#!/usr/bin/env bash
#
# Phase 16 smoke — CodeBlock raster path (image + caption + language badge).
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

badge_built() { grep -rq "\.Language" scene/render_code_block.go 2>/dev/null; }

# 2. Language badge wired (the set-but-unused field renders).
if badge_built; then
    if go test ./scene/ -run 'CodeBlockBadge|CodeBlockLanguage' >/dev/null 2>&1; then
        ok "code_block language badge renders; empty language = no badge"
    else
        fail "code_block language badge renders" "badge tests failed"
    fi
else
    skip "code_block language badge renders" "language badge not built yet"
fi

# 3. Raster + caption path intact.
if go test ./scene/ -run 'CodeBlock' >/dev/null 2>&1; then
    ok "code_block raster + caption render"
else
    fail "code_block raster + caption render" "code_block tests failed"
fi

# 4. Render is byte-identical workers=1 vs N.
if badge_built; then
    if go test ./scene/ -run 'CodeBlockParallel' >/dev/null 2>&1; then
        ok "code_block render is byte-identical workers=1 vs N"
    else
        fail "code_block render is byte-identical workers=1 vs N" "parallel test failed"
    fi
else
    skip "code_block render is byte-identical workers=1 vs N" "not built yet"
fi

echo
echo "phase-16 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
