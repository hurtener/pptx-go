#!/usr/bin/env bash
#
# Phase 78 smoke — gradient-mesh-background (Deckard R13.4, engine). Adds a
# BackgroundMesh kind + MeshGlow{Anchor; Color pptx.ColorRole; Radius pptx.EMU;
# Alpha int} + Background.Mesh []MeshGlow: a base canvas fill + N low-alpha
# caller-anchored radial glows pooled over it (the cover mesh wash), reusing
# pptx.RadialGradient. An empty Mesh emits no shapes (absent config); a new kind
# is byte-identical when unused. Asserts the kind/type/field/render-case exist and
# the mesh/empty/determinism tests pass
# (docs/plans/phase-78-gradient-mesh-background.md §11/§13).
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
grep_check "BackgroundMesh kind"          scene/background.go 'BackgroundMesh'
grep_check "MeshGlow type"                scene/background.go 'type MeshGlow struct'
grep_check "Background.Mesh field"        scene/background.go 'Mesh \[\]MeshGlow'
grep_check "renderBackground mesh case"   scene/render.go     'case BackgroundMesh'
run_check  "mesh emits base + N glows"    ./scene/ 'TestBackground_Mesh$'
run_check  "empty mesh byte-identical"    ./scene/ 'TestBackground_MeshEmpty'
run_check  "mesh deterministic"           ./scene/ 'TestBackground_MeshDeterministic'
run_check  "BackgroundMesh String == mesh" ./scene/ 'TestBackgroundKind_MeshString'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
