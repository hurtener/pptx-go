#!/usr/bin/env bash
#
# Phase 03 smoke — Builder spine (docs/plans/phase-03-builder-spine.md §13).
#
# Built incrementally per chunk (A0→A4, B, C). Criteria whose surface a later
# chunk delivers SKIP until then; what is built must pass. A0 hardened the
# validity harness; A1 rebuilds the emission onto one path (xml.Marshal +
# ooxml.RestoreNamespaces, D-032).

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... >/dev/null 2>&1; then
	ok "library builds CGo-free"
else
	fail "library builds CGo-free" "go build failed"
fi

# 2. A1 (D-032): one emission path — the hand-rolled slide XMLWriter is gone.
if ! grep -rq "globalXMLWriterPool\|func NewXMLWriter\b" internal/ooxml/slide/; then
	ok "A1: hand-rolled slide XMLWriter retired (D-032)"
else
	fail "A1: hand-rolled writer retired" "XMLWriter machinery still present in internal/ooxml/slide"
fi

# 3. A1: slide emission round-trips losslessly (codec + builder halves).
if go test ./internal/ooxml/slide/ -run 'RoundTrip|ToXMLStructure' >/dev/null 2>&1 &&
	go test ./test/parts/ -run 'MarshalFullPage' >/dev/null 2>&1; then
	ok "A1: slide round-trip + structure tests pass"
else
	fail "A1: slide round-trip" "go test (round-trip / structure) failed"
fi

# 4. A1: emitted parts carry a namespaced root (presentation.xml + slides).
NSCHECK="$(
	go run ./_gen/genrefdeck /tmp/pptx-go-smoke-03.pptx >/dev/null 2>&1 &&
		unzip -p /tmp/pptx-go-smoke-03.pptx ppt/presentation.xml 2>/dev/null |
		grep -c '<p:presentation '
)"
if [ "${NSCHECK:-0}" -ge 1 ]; then
	ok "A1: presentation.xml emits a namespaced root <p:presentation>"
else
	fail "A1: namespaced presentation root" "presentation.xml is not namespaced"
fi
rm -f /tmp/pptx-go-smoke-03.pptx

# 5. A2: New() emits a complete deck that passes the full conformance gate
#    (presentation + slides + master + blank layout + theme, all wired).
if go test ./test/integration/ -run Conformance_BuilderOutput >/dev/null 2>&1; then
	ok "A2: New() passes the full-deck conformance gate"
else
	fail "A2: full-deck conformance gate" "TestConformance_BuilderOutput failed"
fi

# 6. A3: EMU coordinates (no px scaling) + New(opts) + WithFormat + Theme.
if go test ./test/pptx/ -run 'New_DefaultFormat|WithFormat|WithTheme|WithFontSource' >/dev/null 2>&1; then
	ok "A3: EMU coords + New(opts)/WithFormat/WithTheme/WithFontSource"
else
	fail "A3: EMU Box API + options" "go test ./test/pptx/ (options set) failed"
fi

# B: Color interface + Fill/Line + AddShape + theme-swap (D-033).
if go test ./test/pptx/ -run 'AddShape' >/dev/null 2>&1 &&
	go test ./internal/ooxml/slide/ -run 'FillRoundTrip' >/dev/null 2>&1; then
	ok "B: Color interface + Fill/Line + AddShape + theme-swap (D-033)"
else
	fail "B: Color/Fill/Line + AddShape" "shape/fill or fill round-trip tests failed"
fi
# C: media / sections / speaker notes / streaming (RFC §8.6–8.8, §17.2).
if go test ./test/pptx/ -run 'AddImage|Sections|SpeakerNotes|SaveStream|OpenStream' >/dev/null 2>&1 &&
	go test ./internal/ooxml/slide/ -run 'PictureMediaRoundTrip' >/dev/null 2>&1; then
	ok "C: media + sections + speaker notes + streaming round-trip"
else
	fail "C: media/sections/notes/streaming" "Chunk C tests failed"
fi

# 7. A4: always-on repair-prompt hygiene pass (D-020).
if go test ./internal/render/ >/dev/null 2>&1 && [ -f docs/design/HYGIENE.md ]; then
	ok "A4: repair-prompt hygiene pass on every write (D-020)"
else
	fail "A4: repair-prompt hygiene pass" "internal/render tests or HYGIENE.md missing"
fi

echo
echo "phase-03 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
