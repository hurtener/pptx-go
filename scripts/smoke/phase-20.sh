#!/usr/bin/env bash
#
# Phase 20 smoke — agent skills + published docs site.
#
# Delivered across PRs: PR#1 plan + reconcile (this skeleton); PR#2 the eight
# skills + runnable examples + the scene/doc.go fix; PR#3 the VitePress docs
# site + pages.yml; PR#4 the §19 drift-hook activation. Criteria flip to OK as
# the PRs land. Each criterion prints OK / SKIP / FAIL; done when OK >= count
# and FAIL == 0.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# The eight skills CLAUDE.md §19 mandates.
SKILLS=(
  scaffold-a-presentation
  define-a-theme
  load-a-brand-template
  compose-a-scene
  embed-a-chart-raster
  embed-a-code-block-raster
  extend-the-icon-set
  register-an-asset
)

# 1. Library builds CGo-free (regression guard).
if CGO_ENABLED=0 go build ./... 2>/dev/null; then
    ok "library builds CGo-free"
else
    fail "library builds CGo-free" "go build failed"
fi

# 2. The eight skills exist with name:/description: frontmatter (PR#2).
if [ -d skills ]; then
    missing=""
    for s in "${SKILLS[@]}"; do
        f="skills/${s}/SKILL.md"
        if [ ! -f "$f" ] || ! grep -q '^name:' "$f" || ! grep -q '^description:' "$f"; then
            missing="${missing} ${s}"
        fi
    done
    if [ -z "$missing" ]; then
        ok "eight agent skills present with valid frontmatter"
    else
        fail "eight agent skills present with valid frontmatter" "missing/invalid:${missing}"
    fi
else
    skip "eight agent skills present with valid frontmatter" "skills/ not added yet (PR#2)"
fi

# 3. Every example program builds and runs clean (the skill smoke) (PR#2).
if [ -d examples ]; then
    failed=""
    for d in examples/*/; do
        [ -f "${d}main.go" ] || continue
        if ! go build "./${d}..." >/dev/null 2>&1; then
            failed="${failed} ${d}(build)"
            continue
        fi
        if ! go run "./${d}" >/dev/null 2>&1; then
            failed="${failed} ${d}(run)"
        fi
    done
    if [ -z "$failed" ]; then
        ok "all examples build and run clean"
    else
        fail "all examples build and run clean" "failed:${failed}"
    fi
else
    skip "all examples build and run clean" "examples/ not added yet (PR#2)"
fi

# 4. The docs site builds (PR#3). The Go preflight does not install Node deps, so
#    this only runs when vitepress is already installed locally; otherwise it
#    SKIPs — the authoritative build runs in .github/workflows/pages.yml on every
#    PR that touches docs/site. This mirrors the codec schema smoke's
#    optional-tool pattern (never FAIL on a missing dev toolchain).
if [ -d docs/site ]; then
    if [ -x docs/site/node_modules/.bin/vitepress ]; then
        if (cd docs/site && node_modules/.bin/vitepress build >/dev/null 2>&1); then
            ok "docs site builds (vitepress)"
        else
            fail "docs site builds (vitepress)" "vitepress build failed"
        fi
    else
        skip "docs site builds (vitepress)" "vitepress not installed; built by pages.yml CI"
    fi
else
    skip "docs site builds (vitepress)" "docs/site/ not added yet (PR#3)"
fi

# 5. The §19 drift hook is active (no longer an inert SKIP) (PR#4).
if grep -q "§19" scripts/drift-audit.sh 2>/dev/null && \
   ! grep -q "inert pre-Phase 20" scripts/drift-audit.sh 2>/dev/null; then
    ok "§19 user-facing-vocabulary drift hook is active"
else
    skip "§19 user-facing-vocabulary drift hook is active" "hook still inert (PR#4)"
fi

echo
echo "phase-20 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
