#!/usr/bin/env bash
#
# Phase 94 smoke — image / diagram annotations (Deckard R14.17, engine; LOW). An
# additive Image.Annotations *ImageAnnotations overlays numbered pins (at
# fractional 0..1 coords of the image box, optional leader-line caption) and
# highlight boxes, drawn as native shapes over the picture. Empty = byte-identical.
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
grep_check "Image.Annotations field"     scene/nodes.go    'Annotations \*ImageAnnotations'
grep_check "ImagePin struct"             scene/nodes.go    'type ImagePin struct'
grep_check "ImageHighlight struct"       scene/nodes.go    'type ImageHighlight struct'
grep_check "renderImageAnnotations"      scene/render_image.go 'func (r \*renderer) renderImageAnnotations'
grep_check "annotation validation"       scene/validate.go 'image pin'
run_check  "annotations render pins+caption" ./scene/ 'TestImageAnnotations$'
run_check  "invalid pin coord fails"         ./scene/ 'TestImageAnnotations_InvalidWarns'
run_check  "nil annotations byte-identical"  ./scene/ 'TestImageAnnotations_NilByteIdentical'
run_check  "annotations deterministic"       ./scene/ 'TestImageAnnotations_Deterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
