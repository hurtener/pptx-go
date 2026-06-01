#!/usr/bin/env bash
#
# Phase 09 smoke — Template ingestion: Theme + Masters
# (docs/plans/phase-09-template-ingestion.md §13).
#
# SKIPs gracefully until the surface lands; FAIL blocks the merge.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# run_check SKIPs when no test matches the pattern yet (surface not built).
run_check() {
	local desc="$1" pkg="$2" pat="$3" found
	found="$(go test "$pkg" -list "$pat" 2>/dev/null | grep -E '^Test' || true)"
	if [ -z "$found" ]; then
		skip "$desc" "not yet landed"
	elif go test "$pkg" -run "$pat" >/dev/null 2>&1; then
		ok "$desc"
	else
		fail "$desc" "test failed"
	fi
}

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 2. Opening a brand kit extracts its theme (incl. sysClr) + fonts + layouts (criterion 1).
run_check "brand template theme: accent + fonts + layouts" ./pptx/ 'Open_ExtractsThemeAndLayouts'
# 3. FromTemplate seeds theme + masters; Masters() lists template layouts (criterion 2).
run_check "FromTemplate seeds theme + masters/layouts" ./pptx/ 'FromTemplate'
# 4. scene.WithTheme renders the brand accent into slide XML (criterion 3).
run_check "scene.WithTheme uses brand colors" ./scene/ 'WithTheme'
# 5. scene.WithLayoutMap emits slide→layout rel; unmapped → blank + warning (criterion 4).
run_check "scene.WithLayoutMap maps LayoutKind → layout" ./scene/ 'WithLayoutMap|LayoutMap'
# 6. FromTemplate deck round-trips + byte-identical + conformant (criterion 5).
run_check "FromTemplate round-trip + determinism + conformance" ./test/integration/ 'TemplateIngest|FromTemplate'

echo
echo "phase-09 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
