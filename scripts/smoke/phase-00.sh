#!/usr/bin/env bash
#
# Phase 00 smoke — module rename and hygiene scaffolding.
# Verifies the acceptance criteria in docs/plans/phase-00-foundation.md §11.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

MODULE="github.com/hurtener/pptx-go"

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "CGO_ENABLED=0 go build ./... failed"
fi

# 2. AGENTS.md == CLAUDE.md.
if diff -q AGENTS.md CLAUDE.md >/dev/null 2>&1; then
	ok "AGENTS.md == CLAUDE.md"
else
	fail "AGENTS.md == CLAUDE.md" "files differ"
fi

# 3. No stale upstream module path outside RFC / decisions log.
# Needle assembled from fragments so this script never contains the literal.
needle='Muprprpr/''Go-pptx'
stale=$(git ls-files -z 2>/dev/null \
	| grep -zvE '^(RFC-001-pptx-go\.md|docs/decisions\.md)$' \
	| xargs -0 grep -lI "$needle" 2>/dev/null || true)
if [ -z "$stale" ]; then
	ok "no stale upstream module path outside RFC/decisions log"
else
	fail "no stale upstream module path" "$(echo "$stale" | tr '\n' ' ')"
fi

# 4. Apache-2.0 LICENSE + preserved upstream MIT.
if grep -q "Apache License" LICENSE 2>/dev/null && grep -q "MIT License" LICENSE.upstream 2>/dev/null; then
	ok "Apache-2.0 LICENSE present; upstream MIT preserved at LICENSE.upstream"
else
	fail "license files" "expected Apache-2.0 LICENSE and MIT LICENSE.upstream"
fi

# 5. coveragecheck tests pass and the gate CLI runs.
if go test ./internal/coveragecheck/ >/dev/null 2>&1; then
	prof=$(mktemp)
	if go test -covermode=atomic -coverprofile="$prof" ./internal/coveragecheck/ >/dev/null 2>&1 \
		&& go run ./internal/coveragecheck/cmd/coveragecheck -profile="$prof" -config=internal/coveragecheck/coverage.json >/dev/null 2>&1; then
		ok "internal/coveragecheck tests pass and the gate CLI runs"
	else
		fail "coverage gate" "gate CLI returned non-zero"
	fi
	rm -f "$prof"
else
	fail "internal/coveragecheck tests" "go test failed"
fi

# 6. go.mod declares the canonical module path.
if head -1 go.mod 2>/dev/null | grep -q "module $MODULE"; then
	ok "go.mod declares $MODULE"
else
	fail "go.mod module path" "expected 'module $MODULE'"
fi

echo
echo "phase-00 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
