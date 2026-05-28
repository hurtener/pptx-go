#!/usr/bin/env bash
#
# Smoke skeleton — copy to scripts/smoke/phase-NN.sh and fill in (CLAUDE.md §16).
#
# A smoke script spot-checks user-visible acceptance criteria mechanically;
# it does NOT re-implement the test suite. Each criterion prints exactly one
# of:
#   OK:   <criterion>
#   SKIP: <criterion> — <reason>     (surface not built yet)
#   FAIL: <criterion> — <details>
#
# A phase is done only when OK >= count(criteria) and FAIL == 0
# (master plan §1.4 / §1.7).

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# --- criteria -------------------------------------------------------------
# Example:
# if CGO_ENABLED=0 go build ./...; then
#     ok "library builds CGo-free"
# else
#     fail "library builds CGo-free" "go build failed"
# fi
# --------------------------------------------------------------------------

echo
echo "phase-NN smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
