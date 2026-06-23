#!/usr/bin/env bash
#
# Phase 69 smoke — prim-attribution-lockup (Deckard R12.9). Adds a Lockup scene IR node: a
# caption + a small partner logo (an asset pic OR a curated icon) composed as one centered
# inline group. Asserts the icon path is media-free, the asset path registers a pic,
# neither/both/unknown-icon fail validation, and the catalog count covers it
# (docs/plans/phase-69-prim-attribution-lockup.md §11/§13).
set -uo pipefail
cd "$(dirname "$0")/../.."
OK=0; FAIL=0; SKIP=0
ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }
run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then skip "$desc" "not yet landed"
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then ok "$desc"
	else fail "$desc" "test failed"; fi
}
grep_check() {
	local desc="$1" file="$2" pat="$3"
	if [ ! -f "$file" ]; then skip "$desc" "missing $file"
	elif grep -q "$pat" "$file"; then ok "$desc"
	else fail "$desc" "pattern not found: $pat"; fi
}
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
grep_check "KindLockup in the node catalog"             scene/nodes.go   'KindLockup'
grep_check "Lockup + AssetSide structs"                 scene/nodes.go   'type Lockup struct'
grep_check "KindLockup policy (HasAsset)"               scene/policy.go  'KindLockup'
grep_check "Lockup Stage-1 validation"                  scene/validate.go 'case Lockup'
grep_check "Lockup composer exists"                     scene/render_lockup.go 'func (r \*renderer) renderLockup'
grep_check "lockup asset use classification"            scene/render.go  'v.AssetID != ""'
grep_check "lockup icon Stage-1 validated"              scene/render_card.go 'lockup'
run_check  "icon lockup is media-free"                  ./scene/ 'TestLockup_IconPath'
run_check  "asset lockup registers a pic"               ./scene/ 'TestLockup_AssetPath'
run_check  "neither/both asset+icon fails validation"   ./scene/ 'TestLockup_Validation'
run_check  "unknown lockup icon fails Stage-1"          ./scene/ 'TestLockup_UnknownIconFails'
run_check  "logo height default + asset classification" ./scene/ 'TestLockupLogoH'
run_check  "lockup render is deterministic"             ./scene/ 'TestLockup_Deterministic'
run_check  "catalog covers KindLockup (28 kinds)"       ./scene/ 'TestCatalog_KindsDistinct'
run_check  "integration round-trips every node"         ./test/integration/ 'TestRoundTrip_SceneNodes'
echo
echo "phase-69 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
