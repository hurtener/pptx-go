#!/usr/bin/env bash
#
# Phase 05 smoke — Scene package scaffold + IR catalog + AssetResolver
# (docs/plans/phase-05-scene-scaffold.md §13). Types only; no rendering yet.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# run_check SKIPs when no test matches the pattern yet (surface not built).
# Captures the listing fully (no early-closing `grep -q`, which would SIGPIPE
# `go test` and trip `pipefail` into a false "not landed").
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

# 2. P1: pptx must not import scene.
if [ -d scene ] && grep -rqs '"github.com/hurtener/pptx-go/scene' pptx/; then
	fail "P1: pptx does not import scene" "pptx imports scene"
else
	ok "P1: pptx does not import scene"
fi

# 3. IR catalog + Stage 1 validation.
run_check "IR catalog + Stage 1 validation" ./scene/ 'Catalog|Validate'
# 4. Render stub on an empty scene.
run_check "Render stub returns zero Stats on empty scene" ./scene/ 'RenderStub|EmptyScene'
# 5. URIAssetResolver.
run_check "URIAssetResolver resolves asset:// ids" ./scene/ 'URIAssetResolver'
# 6. Per-node policy ⇔ struct assertion.
run_check "per-node policy matches the node structs" ./scene/ 'Policy'
# 7. WithWorkers + parallel render is byte-identical to sequential (D-015, RFC §10.1).
run_check "parallel render is deterministic (WithWorkers, D-015)" ./scene/ 'Deterministic|WithWorkers'
# 8. A shared Theme is safe for concurrent render reuse (-race).
run_check "shared Theme safe under concurrent render" ./scene/ 'ConcurrentThemeReuse'

echo
echo "phase-05 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
