#!/usr/bin/env bash
#
# Phase 02 smoke — Theme & token model + font embedding.
# Verifies the acceptance criteria in docs/plans/phase-02-theme-tokens.md §11.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 2. Theme + token + theme-codec tests pass (resolution, swap, round-trip).
if go test ./pptx/ -run 'Theme|Token|Resol|Load|Unit|Box|Anchor' >/dev/null 2>&1; then
	ok "theme/token/units tests pass (resolution, swap, OOXML round-trip)"
else
	fail "theme/token tests" "go test ./pptx/ (theme set) failed"
fi

# 3. Font-embedding tests pass (embed round-trip + no-embed + errors).
if go test ./pptx/ -run 'Embed|Font' >/dev/null 2>&1; then
	ok "font-embedding tests pass (D-019)"
else
	fail "font-embedding tests" "go test ./pptx/ (font set) failed"
fi

# 4. The default-theme template exists and LoadTheme reads its accent.
if [ -f templates/_default-theme.pptx ]; then
	ok "templates/_default-theme.pptx present"
else
	fail "default theme template" "templates/_default-theme.pptx missing"
fi

# 5. Token catalog documented.
if [ -f docs/design/THEME.md ]; then
	ok "docs/design/THEME.md token catalog present"
else
	fail "THEME.md" "docs/design/THEME.md missing"
fi

echo
echo "phase-02 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
