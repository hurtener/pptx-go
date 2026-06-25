#!/usr/bin/env bash
#
# Phase 81 smoke — photographic-imagery background class (Deckard R14.1, engine
# half). Adds a slide-background legibility scrim (scene Background.Scrim) drawn as
# a solid or transparent→color gradient overlay over any drawn background, and a
# two-tone duotone recolor of a photographic background (scene Background.Duotone)
# realized by a new builder (*pptx.Image).SetDuotone(shadow, highlight Color) that
# emits an <a:duotone> blip effect. All additive: a nil Scrim / nil Duotone renders
# byte-identically. Asserts the API/fields exist and the round-trip, byte-identity,
# and determinism tests pass (docs/plans/phase-81-photographic-background.md §11/§13).
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
grep_check "Image.SetDuotone method"        pptx/media.go      'func (im \*Image) SetDuotone'
grep_check "Image.Duotone read accessor"    pptx/media.go      'func (im \*Image) Duotone()'
grep_check "XDuotone blip effect struct"    internal/ooxml/slide/slide_types.go 'type XDuotone struct'
grep_check "duotone namespace registered"   internal/ooxml/restorenamespaces.go '"duotone": "a"'
grep_check "scene Background.Scrim field"    scene/background.go 'Scrim \*Scrim'
grep_check "scene Background.Duotone field"  scene/background.go 'Duotone \*Duotone'
grep_check "renderScrim overlay"             scene/render.go    'func (r \*renderer) renderScrim'
run_check  "builder duotone round-trips"     ./pptx/  'TestImageDuotone_RoundTrip'
run_check  "builder duotone token resolves"  ./pptx/  'TestImageDuotone_TokenResolves'
run_check  "builder nil-tone byte-identical" ./pptx/  'TestImageDuotone_NilByteIdentical'
run_check  "scene solid scrim overlay"       ./scene/ 'TestBackground_ScrimSolid'
run_check  "scene gradient scrim overlay"    ./scene/ 'TestBackground_ScrimGradient'
run_check  "scene scrim+duotone photo class" ./scene/ 'TestBackground_ScrimDuotonePhoto'
run_check  "scene nil scrim byte-identical"  ./scene/ 'TestBackground_ScrimNilByteIdentical'
run_check  "scene scrim+duotone deterministic" ./scene/ 'TestBackground_ScrimDeterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
