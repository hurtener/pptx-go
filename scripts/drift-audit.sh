#!/usr/bin/env bash
#
# drift-audit.sh — mechanical design-coherence checks (CLAUDE.md §4, §13, §19).
#
# Each check is independent and either PASSes, FAILs, or SKIPs (a check whose
# target surface does not yet exist skips gracefully). The script exits
# non-zero if any check FAILs. Checks grow as phases land; this is the
# Phase 00 baseline.

set -uo pipefail
cd "$(dirname "$0")/.."

PASS=0
FAIL=0
SKIP=0

pass() { printf 'OK:   %s\n' "$1"; PASS=$((PASS + 1)); }
fail() { printf 'FAIL: %s\n' "$1"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# Tracked Go files, NUL-delimited, reused across checks.
go_files() { git ls-files -z '*.go'; }

# ---------------------------------------------------------------------------
# 1. AGENTS.md and CLAUDE.md are verbatim identical (CLAUDE.md §18).
# ---------------------------------------------------------------------------
if [ -f AGENTS.md ] && [ -f CLAUDE.md ]; then
	if diff -q AGENTS.md CLAUDE.md >/dev/null; then
		pass "AGENTS.md == CLAUDE.md (mirror)"
	else
		fail "AGENTS.md and CLAUDE.md diverge — re-mirror them"
	fi
else
	skip "mirror check" "AGENTS.md / CLAUDE.md missing"
fi

# ---------------------------------------------------------------------------
# 2. Module path is canonical. The pre-fork upstream name must not reappear
#    in tracked source/config; historical mentions live only in the RFC and
#    the decisions log.
# ---------------------------------------------------------------------------
# The needle is assembled from two fragments so this detector script (and
# the smoke script) never contains the contiguous literal — otherwise the
# grep below would flag itself once these files are tracked.
needle='Muprprpr/''Go-pptx'
stale=$(git ls-files -z | grep -zvE '^(RFC-001-pptx-go\.md|docs/decisions\.md)$' \
	| xargs -0 grep -lI "$needle" 2>/dev/null || true)
if [ -z "$stale" ]; then
	pass "no stale upstream module path outside RFC/decisions log"
else
	fail "stale upstream module path found in: $(echo "$stale" | tr '\n' ' ')"
fi

# ---------------------------------------------------------------------------
# 3. P3 — raw OOXML / encoding/xml types are isolated to the OOXML + OPC
#    layers. Only internal/ooxml (wire types) and internal/opc (OPC plumbing:
#    content-types, relationships) may import encoding/xml; nothing above the
#    internal wall (pptx, scene, …) may. Test files are exempt.
# ---------------------------------------------------------------------------
offenders=$(go_files \
	| { grep -zvE '^internal/(ooxml|opc)/' || true; } \
	| { grep -zvE '_test\.go$' || true; } \
	| xargs -0 grep -lE '"encoding/xml"' 2>/dev/null || true)
if [ -z "$offenders" ]; then
	pass "encoding/xml confined to internal/ooxml and internal/opc (P3)"
else
	fail "encoding/xml imported outside the OOXML/OPC layers (P3): $(echo "$offenders" | tr '\n' ' ')"
fi

# ---------------------------------------------------------------------------
# 3b. The old top-level opc/ and parts/ packages are fully relocated.
# ---------------------------------------------------------------------------
if [ -d opc ] || [ -d parts ]; then
	fail "old opc/ or parts/ package directories still present (Phase 01 relocates them)"
else
	pass "opc/ and parts/ relocated under internal/"
fi

# ---------------------------------------------------------------------------
# 4. P1 — pptx must not import scene (the dependency is one-way).
# ---------------------------------------------------------------------------
if [ -d pptx ]; then
	offenders=$(go_files \
		| { grep -zlE '^pptx/' || true; } \
		| xargs -0 grep -lE '"github.com/hurtener/pptx-go/scene' 2>/dev/null || true)
	if [ -z "$offenders" ]; then
		pass "pptx/ does not import scene/ (P1)"
	else
		fail "pptx/ imports scene/: $(echo "$offenders" | tr '\n' ' ')"
	fi
else
	skip "P1 layering check" "pptx/ not yet present"
fi

# ---------------------------------------------------------------------------
# 5. §19 — contributor phase vocabulary must not leak into user-facing docs.
#    Inert until the published docs site / skills exist (Phase 20).
# ---------------------------------------------------------------------------
if [ -d docs/site ] || [ -d skills ]; then
	paths=$(git ls-files 'README.md' 'CHANGELOG.md' 'docs/site/**/*.md' 'examples/**/README.md' 2>/dev/null || true)
	leak=""
	for p in $paths; do
		if grep -qiE '\bphase[ -]?[0-9]' "$p" 2>/dev/null; then
			leak="$leak $p"
		fi
	done
	if [ -z "$leak" ]; then
		pass "no phase vocabulary in user-facing docs (§19)"
	else
		fail "phase vocabulary leaked into:$leak"
	fi
else
	skip "§19 user-facing vocabulary check" "docs/site and skills not yet present (inert pre-Phase 20)"
fi

# ---------------------------------------------------------------------------
echo
echo "drift-audit: ${PASS} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
