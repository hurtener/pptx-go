#!/usr/bin/env bash
#
# Phase 15 smoke — Flow (step pills + connectors).
#
# Spot-checks the phase-15 acceptance criteria mechanically (CLAUDE.md §16).
# Each criterion prints exactly one of:
#   OK:   <criterion>
#   SKIP: <criterion> — <reason>     (surface not built yet)
#   FAIL: <criterion> — <details>
#
# A phase is done only when OK >= count(criteria) and FAIL == 0.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# --- criteria -------------------------------------------------------------

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... 2>/dev/null; then
    ok "library builds CGo-free"
else
    fail "library builds CGo-free" "go build failed"
fi

flow_built() { grep -rq "func.*renderFlow\b" scene/ 2>/dev/null; }

# 2. A flow renders pills + connectors (horizontal arrow).
if flow_built; then
    if go test ./scene/ -run 'Flow' >/dev/null 2>&1; then
        ok "flow renders step pills + connectors"
    else
        fail "flow renders step pills + connectors" "flow render tests failed"
    fi
else
    skip "flow renders step pills + connectors" "flow composer not built yet"
fi

# 3. Connector kinds (cycle return-arrow, arrow_dashed, plus, vertical rotation).
if grep -rq "ConnectorKind\|ConnectorCycle" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'FlowConnector|FlowCycle|FlowVertical|FlowDashed|FlowPlus' >/dev/null 2>&1; then
        ok "connector kinds render (cycle / arrow_dashed / plus / vertical)"
    else
        fail "connector kinds render" "connector tests failed"
    fi
else
    skip "connector kinds render" "connector kinds not built yet"
fi

# 4. A flow step icon resolves; an unknown name fails Stage-1.
if grep -rq "FlowStep" scene/nodes.go 2>/dev/null && grep -q "Icon" <(grep -A6 "type FlowStep" scene/nodes.go 2>/dev/null); then
    if go test ./scene/ -run 'FlowIcon' >/dev/null 2>&1; then
        ok "flow step icon resolves; unknown name fails Stage-1"
    else
        fail "flow step icon resolves; unknown name fails Stage-1" "flow icon test failed"
    fi
else
    skip "flow step icon resolves; unknown name fails Stage-1" "flow step icon not built yet"
fi

# 5. Flow render is byte-identical workers=1 vs N.
if flow_built; then
    if go test ./scene/ -run 'FlowParallel|FlowIdempot' >/dev/null 2>&1; then
        ok "flow render is byte-identical workers=1 vs N"
    else
        fail "flow render is byte-identical workers=1 vs N" "parallel-equivalence test failed"
    fi
else
    skip "flow render is byte-identical workers=1 vs N" "flow composer not built yet"
fi

# --------------------------------------------------------------------------

echo
echo "phase-15 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
