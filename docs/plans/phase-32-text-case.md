# Phase 32 — text case

**Subsystem:** pptx — Layer 1 builder (theme/typography)
**RFC sections:** §7 (theme tokens), §9 (rich text / runs)
**Deps:** Phase 02 (theme/FontSpec), Phase 04 (rich text), Phase 30 (tracking —
the run-attribute token pattern). External: none.
**Status:** Done

---

## 1. Goal

Add a per-type-role **case transform** token (`FontSpec.Case`: none / upper /
small-caps) rendered via OOXML `a:rPr/@cap`, so a soul can uppercase eyebrows/
labels by theme (not by pre-uppercasing copy) — additive, round-trip clean, the
run text preserved, byte-identical when none.

## 2. Why now

Third Wave 9 unit (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.11, LOW · engine; D-059).
Pairs with Phase 30 tracking to produce the canonical tracked-caps eyebrow.
Run-level like tracking — flows automatically through `toProps`.

## 3. RFC sections implemented

- `RFC §7` — extends the resolved type scale (`FontSpec`) with a case transform.
- `RFC §9` — an optional per-run case override on `RunStyle`, round-tripped via
  the read model (G6); the run text is preserved (cap is a display attribute).

## 4. Brief findings incorporated

- `docs/research/17-type-detail-tokens.md` — the type-detail-token family; case
  is the run-level transform. Chosen path: emit the OOXML `cap` attribute (text
  preserved, round-trips) over rewriting run text — the deterministic, lossless
  option.

## 5. Findings I'm departing from

None. Lower/title case (not OOXML `cap` values) would need a text rewrite and are
deferred. The engine does **not** seed `DefaultTheme().TypeCaption` to upper
(that's the soul's choice, D-026) — keeping the engine default byte-identical.

## 6. Decisions referenced

- `D-059` — Wave-2 engine scope. `D-060` — sibling tracking token (same pattern).
- `G6` — round-trip fidelity. Files **D-062 — case token**.

## 7. Architecture

```text
pptx/theme.go        FontSpec += Case TextCase; type TextCase (None/Upper/SmallCaps)
                     + capAttr() / textCaseFromCap() helpers
pptx/text.go         RunStyle += Case *TextCase (nil = inherit role); (*Run).Case()
internal/ooxml/slide XTextProperties += Cap string (a:rPr/@cap, omitempty)
pptx/text_layout.go  toProps: emit cap from effective case when not CaseNone
```

`cap` is an attribute on the already-prefixed `rPr` element ⇒ no
`RestoreNamespaces` change (contrast D-061's `lnSpc` element). Run-level ⇒ no
scene change (flows via `toProps`).

## 8. Files added or changed

```text
pptx/theme.go                        # CHANGED — FontSpec.Case + TextCase + helpers
pptx/text.go                         # CHANGED — RunStyle.Case + Run.Case()
pptx/text_layout.go                  # CHANGED — toProps emits a:rPr/@cap
internal/ooxml/slide/slide_types.go  # CHANGED — XTextProperties.Cap
pptx/text_case_test.go               # NEW — emit, byte-identity, round-trip+text-preserved, role-level
scripts/smoke/phase-32.sh            # NEW — phase smoke
docs/plans/phase-32-text-case.md     # NEW — this plan
docs/decisions.md                    # CHANGED — adds D-062
docs/glossary.md                     # CHANGED — adds "Case (type)"
docs/design/THEME.md                 # CHANGED — case token
docs/site/guide/theme.md             # CHANGED — case note (§19)
skills/define-a-theme/SKILL.md       # CHANGED — FontSpec.Case (§19)
```

## 9. Public API surface

```go
// pptx
type TextCase int
const ( CaseNone TextCase = iota; CaseUpper; CaseSmallCaps )

type FontSpec struct { /* … */ Case TextCase }
type RunStyle struct { /* … */ Case *TextCase } // nil = inherit role
func (r *Run) Case() TextCase
```

## 10. Risks

- **R1 — text rewrite vs display attr.** Chose the `cap` attribute so the run
  text is preserved and round-trips. **Mitigation:** a test asserts `Run.Text()`
  is unchanged after a `CaseUpper` round-trip.
- **R2 — byte-identity.** `cap` omitempty + emitted only for non-`CaseNone`; a
  test asserts no `cap` attr / byte-identity for the default.
- **R3 — builtin shadowing.** `cap` is a Go builtin; the local var is named
  `capv`/`capRole` to avoid the `predeclared` lint.

## 11. Acceptance criteria

1. A role/run with `CaseUpper`/`CaseSmallCaps` emits `a:rPr cap="all"`/`"small"`.
2. `CaseNone` (role zero, override nil) emits no `cap` and is byte-identical.
3. A cased run round-trips via `Run.Case()` and its `Text()` stays original-case.
4. A `RunStyle.Case` override wins over the role; the role-level path emits on an
   override-free run.
5. `make coverage` ≥ band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default new builder API |

## 13. Smoke check

`scripts/smoke/phase-32.sh` — builds; case emits `a:rPr/@cap` (+ role-level);
no case byte-identical; round-trips with text preserved.

## 14. Tests

- **Unit / golden:** `pptx` — emit `cap` (all/small), byte-identity, role-level.
- **Round-trip golden:** `Run.Case()` + `Run.Text()` preserved (G6).

## 15. Vocabulary added

- `Case (type)` — per-role case transform (`FontSpec.Case`) via `a:rPr/@cap`.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for `pptx`.
- [x] `scripts/smoke/phase-32.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] `make lint` clean.
- [x] Glossary + `docs/design/THEME.md` updated.
- [x] Decision entry D-062 added.
- [x] Docs site + `define-a-theme` skill updated (§19).
