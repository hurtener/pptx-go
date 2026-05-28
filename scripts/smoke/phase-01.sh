#!/usr/bin/env bash
#
# Phase 01 smoke — OPC + OOXML reorg.
# Verifies the acceptance criteria in docs/plans/phase-01-opc-ooxml-reorg.md §11.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# 1. Old top-level packages are gone; new homes exist.
if [ -d opc ] || [ -d parts ]; then
	fail "opc/ and parts/ relocated" "old directories still present"
elif [ -d internal/opc ] && [ -d internal/ooxml ]; then
	ok "opc/ and parts/ relocated to internal/opc and internal/ooxml"
else
	skip "reorg layout" "internal/opc or internal/ooxml not yet present"
fi

# 2. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 3. No package references the old import paths.
old=$(git ls-files '*.go' 2>/dev/null \
	| xargs grep -lE '"github.com/hurtener/pptx-go/(opc|parts)"' 2>/dev/null || true)
if [ -z "$old" ]; then
	ok "no references to the old opc/ or parts/ import paths"
else
	fail "old import paths remain" "$(echo "$old" | tr '\n' ' ')"
fi

# 4. Each internal/ooxml subpackage builds independently.
if [ -d internal/ooxml ]; then
	bad=""
	for d in internal/ooxml internal/ooxml/*/; do
		[ -d "$d" ] || continue
		ls "$d"*.go >/dev/null 2>&1 || continue
		if ! CGO_ENABLED=0 go build "./${d%/}" >/dev/null 2>&1; then
			bad="$bad ${d%/}"
		fi
	done
	if [ -z "$bad" ]; then
		ok "each internal/ooxml subpackage builds independently"
	else
		fail "internal/ooxml subpackage build" "$bad"
	fi
else
	skip "internal/ooxml subpackage build" "internal/ooxml not yet present"
fi

# 5. Round-trip spot-check integration test passes.
if [ -f test/integration/reorg_roundtrip_test.go ]; then
	if go test ./test/integration/ >/dev/null 2>&1; then
		ok "round-trip spot-check passes"
	else
		fail "round-trip spot-check" "go test ./test/integration/ failed"
	fi
else
	skip "round-trip spot-check" "test/integration/reorg_roundtrip_test.go not yet present"
fi

echo
echo "phase-01 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
