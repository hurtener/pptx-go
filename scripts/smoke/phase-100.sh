#!/usr/bin/env bash
#
# Phase 100 smoke — dark-variant accents & extensions (Deckard R8.7, MED · both,
# engine verify-and-close). The Phase-97 DarkColors overlay already re-resolves
# accent/semantic/text roles per dark variant, and the engine's neutral card
# borders use ColorSurfaceAlt which darkThemeFrom dark-resolves — so a dark slide
# carries dark borders/accents, not light-theme values. The border/accentSoft
# extension tokens + the derived-dark-hairline default are Deckard's product half.
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
run_check  "dark border resolves to dark SurfaceAlt" ./scene/ 'TestDarkExtensions_DefaultBorderDarkResolves'
run_check  "dark accent border overridable"          ./scene/ 'TestDarkExtensions_AccentBorderOverridable'
run_check  "dark accent text overridable"            ./scene/ 'TestDarkExtensions_AccentTextOverridable'
run_check  "no-DarkColors dark slide byte-identical" ./scene/ 'TestDarkExtensions_NilByteIdentical'
grep_check "R8.7 brief present"                       docs/research/83-dark-extensions-and-accents.md 'R8.7'
grep_check "D-138 decision present"                   docs/decisions.md 'D-138'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
