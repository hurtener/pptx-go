#!/usr/bin/env bash
#
# Phase 82 smoke — image-as-card-fill (Deckard R14.1, engine half, part 2). Adds a
# cover-fit photographic surface fill: a new builder WithImageFill(src ImageSource)
# ShapeOption emits an <a:blipFill> on a shape's spPr (center-cropped to cover via
# srcRect from the image's format-header dims — §7/D-046), and scene Card.ImageFill
# (an AssetID resolved through the AssetResolver) fills a card surface with a photo
# instead of its solid/gradient fill. Additive: "" / nil renders byte-identically.
# Asserts the API/fields exist and the emit, cover-crop, round-trip, byte-identity,
# and determinism tests pass (docs/plans/phase-82-image-card-fill.md §11/§13).
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
grep_check "WithImageFill shape option"     pptx/shape.go      'func WithImageFill'
grep_check "coverSrcRect cover-fit helper"  pptx/imagefill.go  'func coverSrcRect'
grep_check "XShapeProperties.BlipFill field" internal/ooxml/slide/slide_types.go 'BlipFill     \*XBlipFillProperties'
grep_check "blipFill namespace context rule" internal/ooxml/restorenamespaces.go 'local == "blipFill"'
grep_check "scene Card.ImageFill field"      scene/nodes.go    'ImageFill AssetID'
grep_check "cardChrome routes image fill"    scene/render_card.go 'WithImageFill(c.imageFill)'
grep_check "Card image fill = serial asset"  scene/render.go   'v.ImageFill != ""'
run_check  "shape image fill emits a:blipFill" ./pptx/  'TestImageFill_Emits'
run_check  "image fill cover-crop"             ./pptx/  'TestImageFill_CoverCrop'
run_check  "nil image fill byte-identical"     ./pptx/  'TestImageFill_NilNoChange'
run_check  "card image fill resolves photo"    ./scene/ 'TestCardImageFill$'
run_check  "card image fill missing warns"     ./scene/ 'TestCardImageFill_MissingWarns'
run_check  "card empty image fill byte-ident"  ./scene/ 'TestCardImageFill_EmptyByteIdentical'
run_check  "card image fill deterministic"     ./scene/ 'TestCardImageFill_Deterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
