#!/usr/bin/env bash
#
# Phase 04 smoke — Rich text model (docs/plans/phase-04-rich-text.md §13).
#
# Built incrementally per chunk (A: core model + styles; B: bullets, inline
# code, hyperlinks, notes-as-TextFrame). Criteria whose surface a later chunk
# delivers SKIP until then; what is built must pass.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# run_check SKIPs when no test matches the pattern yet (surface not built),
# otherwise runs it and reports OK/FAIL. A pattern matching zero tests would
# otherwise make `go test -run` exit 0 — a false OK.
run_check() {
	local desc="$1" pkg="$2" pat="$3"
	if ! go test "$pkg" -list "$pat" 2>/dev/null | grep -qE '^Test'; then
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

# A: core model + styled-run round-trip.
run_check "A: TextFrame + styled-run round-trip" ./test/pptx/ 'TextFrame|StyledRun'

# B: inline code, hyperlinks, bullets, token color, notes-as-TextFrame.
run_check "B: inline-code styling (D-013)" ./test/pptx/ 'InlineCode'
run_check "B: hyperlink run + relationship" ./test/pptx/ 'Hyperlink'
run_check "B: bullet / numbered / checklist paragraphs" ./test/pptx/ 'Bullet|Checklist|Numbered'
run_check "B: token run color honors theme swap" ./test/pptx/ 'RunColor|TextToken'
run_check "B: SpeakerNotes() *TextFrame round-trip" ./test/pptx/ 'NotesTextFrame|SpeakerNotesFrame'

echo
echo "phase-04 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
