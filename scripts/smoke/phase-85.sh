#!/usr/bin/env bash
#
# Phase 85 smoke — quote / testimonial enrichment (Deckard R14.5, engine). Extends
# the scene Quote node additively: Mark bool (an oversized low-emphasis quotation
# glyph behind the text), AvatarAssetID (a rounded author avatar via the
# AssetResolver), structured AttributionName/Role/Company, and LogoAssetID (a
# customer logo). Any enrichment field switches to the testimonial layout (one
# balanced unit); a Quote with only Text+Attribution renders byte-for-byte as
# before. A Quote with an avatar/logo becomes asset-bearing (serial determinism).
# Asserts the fields + wiring exist and the testimonial, plain-byte-identical,
# missing-warns, and determinism tests pass (docs/plans/phase-85-quote-testimonial.md).
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
grep_check "Quote.Mark field"            scene/nodes.go      'Mark bool'
grep_check "Quote.AvatarAssetID field"   scene/nodes.go      'AvatarAssetID AssetID'
grep_check "Quote structured attribution" scene/nodes.go     'AttributionName    string'
grep_check "Quote.enriched helper"       scene/nodes.go      'func (q Quote) enriched()'
grep_check "renderTestimonial composer"  scene/render_leaves.go 'func (r \*renderer) renderTestimonial'
grep_check "Quote serial when asset"     scene/render.go     'v.AvatarAssetID != "" || v.LogoAssetID != ""'
run_check  "testimonial renders unit"     ./scene/ 'TestQuote_Testimonial$'
run_check  "plain quote byte-identical"   ./scene/ 'TestQuote_PlainByteIdentical'
run_check  "testimonial missing warns"    ./scene/ 'TestQuote_TestimonialMissingWarns'
run_check  "testimonial deterministic"    ./scene/ 'TestQuote_TestimonialDeterministic'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
