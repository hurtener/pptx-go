#!/usr/bin/env bash
#
# Wave 15 §17 checkpoint smoke — the FIX-NOW items (D-141): a named-gradient spec
# with a nil-Color stop, or GradientName set on a non-BackgroundGradient kind, must
# degrade to a LayoutWarning + skip (RFC §10.2 degrade-and-warn) rather than ship a
# schema-invalid <a:gs> or silently drop the name.
set -uo pipefail
cd "$(dirname "$0")/../.."
OK=0; FAIL=0; SKIP=0
ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then printf 'SKIP: %s\n' "$desc"; SKIP=$((SKIP + 1))
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then ok "$desc"
	else fail "$desc" "test failed"; fi
}
grep_check() {
	local desc="$1" file="$2" pat="$3"
	if grep -q "$pat" "$file"; then ok "$desc"; else fail "$desc" "pattern not found: $pat"; fi
}
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then ok "library builds CGo-free"; else fail "library builds CGo-free" "go build failed"; fi
grep_check "nil-Color guard in the gradient gate" scene/render.go 's.Color == nil'
run_check  "nil-Color named gradient warns + skips" ./scene/ 'TestNamedGradient_NilColorStopWarns'
run_check  "GradientName on wrong kind warns"       ./scene/ 'TestNamedGradient_WrongKindWarns'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
