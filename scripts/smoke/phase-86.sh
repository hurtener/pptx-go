#!/usr/bin/env bash
#
# Phase 86 smoke — number / currency / percent / locale format (Deckard R14.13,
# engine half). Adds a deterministic scene.NumberFormat{Decimals, GroupSep,
# DecimalSep, CurrencySymbol, SymbolAfter, Percent, Compact, CompactThreshold,
# Prefix, Suffix} + FormatNumber(v, f), and a typed numeric path on Stat
# (Number *float64 + Format *NumberFormat) that formats deterministically and then
# applies the existing AutoFit shrink-to-fit so a value like "$4,000+" stays on one
# line (the slide-09 wrap regression fix). Raw-string Stat.Value is unaffected
# (byte-identical). Asserts the API + wiring exist and the format + Stat-path tests
# pass (docs/plans/phase-86-locale-format.md).
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
grep_check "NumberFormat type"        scene/numberformat.go 'type NumberFormat struct'
grep_check "FormatNumber func"        scene/numberformat.go 'func FormatNumber'
grep_check "Stat.Number field"        scene/nodes.go        'Number \*float64'
grep_check "Stat.displayValue helper" scene/nodes.go        'func (s Stat) displayValue()'
grep_check "renderStat uses formatted value" scene/render_stat.go 'v.displayValue()'
run_check  "FormatNumber locale/currency/percent/compact" ./scene/ 'TestFormatNumber'
run_check  "Stat numeric path renders formatted"          ./scene/ 'TestStat_NumberPath'
printf '\n%d OK, %d FAIL, %d SKIP\n' "$OK" "$FAIL" "$SKIP"
[ "$FAIL" -eq 0 ]
