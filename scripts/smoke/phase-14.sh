#!/usr/bin/env bash
#
# Phase 14 smoke — Card + CardSection + the builder elevation/shadow primitive.
#
# Spot-checks the phase-14 acceptance criteria mechanically (CLAUDE.md §16); it
# does NOT re-implement the test suite. Each criterion prints exactly one of:
#   OK:   <criterion>
#   SKIP: <criterion> — <reason>     (surface not built yet)
#   FAIL: <criterion> — <details>
#
# Delivered in two PRs (D-043): PR#1 lands the shadow primitive (criteria 1-3
# go OK), card criteria 4-9 stay SKIP; PR#2 flips them to OK. A phase is done
# only when OK >= count(criteria) and FAIL == 0.

set -uo pipefail
cd "$(dirname "$0")/../.."

OK=0
FAIL=0
SKIP=0

ok()   { printf 'OK:   %s\n' "$1"; OK=$((OK + 1)); }
fail() { printf 'FAIL: %s — %s\n' "$1" "$2"; FAIL=$((FAIL + 1)); }
skip() { printf 'SKIP: %s — %s\n' "$1" "$2"; SKIP=$((SKIP + 1)); }

# --- criteria -------------------------------------------------------------

# 1. Library builds CGo-free.
if CGO_ENABLED=0 go build ./... 2>/dev/null; then
    ok "library builds CGo-free"
else
    fail "library builds CGo-free" "go build failed"
fi

# 2. The elevation/shadow primitive exists and emits outerShdw (PR#1).
if grep -rq "func WithElevation" pptx/ 2>/dev/null && grep -rq "outerShdw\|XOuterShadow" internal/ooxml/slide/ 2>/dev/null; then
    if go test ./test/pptx/ -run 'WithElevation|WithShadow' >/dev/null 2>&1; then
        ok "WithElevation emits a round-tripping outerShdw"
    else
        fail "WithElevation emits a round-tripping outerShdw" "shadow tests failed"
    fi
else
    skip "WithElevation emits a round-tripping outerShdw" "shadow primitive not built yet (PR#1)"
fi

# 3. A no-shadow shape emits no effectLst (byte-identical guard) (PR#1).
if grep -rq "func WithElevation" pptx/ 2>/dev/null; then
    if go test ./test/pptx/ -run 'ShadowOmittedWhenFlat' >/dev/null 2>&1; then
        ok "no-shadow / flat shape emits no effectLst"
    else
        fail "no-shadow / flat shape emits no effectLst" "byte-identical guard test failed"
    fi
else
    skip "no-shadow / flat shape emits no effectLst" "shadow primitive not built yet (PR#1)"
fi

# 4. Card renders chrome (background + stripe + header) (PR#2).
if grep -rq "func.*renderCard\b" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'Card' >/dev/null 2>&1; then
        ok "Card renders chrome (background + stripe + header)"
    else
        fail "Card renders chrome (background + stripe + header)" "card render tests failed"
    fi
else
    skip "Card renders chrome (background + stripe + header)" "card composer not built yet (PR#2)"
fi

# 5. Each card knob renders (PR#2).
if grep -rq "BorderStyle\|CardSize\|CardLayout" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'CardKnob|CardVariant' >/dev/null 2>&1; then
        ok "each card knob renders (fill/border/size/elevation/body_layout/layout/pill/eyebrow)"
    else
        fail "each card knob renders" "knob coverage test failed"
    fi
else
    skip "each card knob renders" "card knobs not built yet (PR#2)"
fi

# 6. A Card with a curated icon places the icon; unknown name fails Stage-1 (PR#2).
if grep -rq "validateIconRefs" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'Icon' >/dev/null 2>&1; then
        ok "Card icon resolves; unknown icon name fails Stage-1"
    else
        fail "Card icon resolves; unknown icon name fails Stage-1" "icon consumption test failed"
    fi
else
    skip "Card icon resolves; unknown icon name fails Stage-1" "icon wiring not built yet (PR#2)"
fi

# 7. A card_section of cards renders (card-of-cards) (PR#2).
if grep -rq "func.*renderCardSection" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'CardSection' >/dev/null 2>&1; then
        ok "card_section of cards renders (card-of-cards)"
    else
        fail "card_section of cards renders (card-of-cards)" "card_section test failed"
    fi
else
    skip "card_section of cards renders (card-of-cards)" "card_section composer not built yet (PR#2)"
fi

# 8. A card_section with a code_block renders native chrome + a pic (PR#2).
if grep -rq "func.*renderCardSection" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'CardSectionMixed|CardSectionCode' >/dev/null 2>&1; then
        ok "card_section + code_block renders native chrome + a pic"
    else
        fail "card_section + code_block renders native chrome + a pic" "mixed-policy test failed"
    fi
else
    skip "card_section + code_block renders native chrome + a pic" "card_section composer not built yet (PR#2)"
fi

# 9. Render is byte-identical at workers=1 vs N for a card scene (PR#2).
if grep -rq "func.*renderCard\b" scene/ 2>/dev/null; then
    if go test ./scene/ -run 'CardParallel|CardIdempot' >/dev/null 2>&1; then
        ok "card scene render is byte-identical workers=1 vs N"
    else
        fail "card scene render is byte-identical workers=1 vs N" "parallel-equivalence test failed"
    fi
else
    skip "card scene render is byte-identical workers=1 vs N" "card composer not built yet (PR#2)"
fi

# --------------------------------------------------------------------------

echo
echo "phase-14 smoke: ${OK} OK, ${FAIL} FAIL, ${SKIP} SKIP"
[ "$FAIL" -eq 0 ]
