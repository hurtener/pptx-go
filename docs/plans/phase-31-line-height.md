# Phase 31 — line height

**Subsystem:** pptx — Layer 1 builder + scene renderer (`RFC §3.3`, §7, §9)
**RFC sections:** §7 (theme tokens), §9 (rich text / paragraphs)
**Deps:** Phase 02 (theme/FontSpec), Phase 04 (rich text), Phase 30 (the
type-detail-token pattern). External: none.
**Status:** Done

---

## 1. Goal

Add a per-type-role **line-height** (leading) token so display headlines set
tight and body stays readable: `FontSpec.LineHeight` (percent of single) applied
to a node's paragraphs and emitted as OOXML `a:pPr/a:lnSpc/a:spcPct`, with a
`ParagraphOpts.LineHeight` builder override — additive, round-trip clean,
byte-identical when 0/100.

## 2. Why now

Second Wave 9 unit (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.4, HIGH · engine; D-059).
Multi-line display headlines need tight leading to read as compact editorial
blocks; the engine had no leading control at all. Sibling of Phase 30 tracking.

## 3. RFC sections implemented

- `RFC §7` — extends the resolved type scale (`FontSpec`) with a line-height.
- `RFC §9` — paragraph line spacing on `ParagraphOpts`, applied by the scene per
  node role and round-tripped via the read model (G6).

## 4. Brief findings incorporated

- `docs/research/17-type-detail-tokens.md` — the type-detail-token family
  (tracking, line-height, case): line-height is *paragraph*-level (`a:lnSpc`),
  distinct from run-level tracking, and also informs the wrap estimator. This
  phase delivers the paragraph mechanism + scene role-application; the estimator
  feed is the brief's noted later refinement.

## 5. Findings I'm departing from

The brief notes line-height should "feed the wrap/preferredHeight estimator." That
estimator feed is **deferred** (documented in §16 + D-061): the current per-line
height is a fixed constant, not leading-derived, so making it leading-aware is a
model rework folded with R9.5 / R10.10. This phase delivers the *visual* leading;
the estimator-accuracy refinement follows.

## 6. Decisions referenced

- `D-059` — Wave-2 engine scope. `D-060` — sibling tracking token (the pattern).
- `G6` — round-trip fidelity (line-height round-trips).
- Files **D-061 — line-height token** in `docs/decisions.md`.

## 7. Architecture

```text
pptx/theme.go        FontSpec += LineHeight float64   (percent of single; 0/100 = none)
pptx/text.go         ParagraphOpts += LineHeight; AddParagraph emits a:lnSpc;
                     (*Paragraph).LineHeight() read accessor
internal/ooxml/slide XParaProps += LnSpc *XLnSpc (first child); XLnSpc/XSpcPct
internal/ooxml       RestoreNamespaces: lnSpc/spcPct → a: prefix (write-path fix)
scene/render_leaves  lineH(role) helper + leaf renderers set ParagraphOpts.LineHeight
                     from the node's base role (plainPara routes Hero/section/attr)
```

## 8. Files added or changed

```text
pptx/theme.go                          # CHANGED — FontSpec.LineHeight
pptx/text.go                           # CHANGED — ParagraphOpts.LineHeight, AddParagraph emit, Paragraph.LineHeight()
internal/ooxml/slide/slide_types.go    # CHANGED — XParaProps.LnSpc + XLnSpc/XSpcPct
internal/ooxml/restorenamespaces.go    # CHANGED — lnSpc/spcPct a: prefix
scene/render_leaves.go                 # CHANGED — lineH helper + per-role LineHeight on leaves
pptx/text_lineheight_test.go           # NEW — emit, byte-identity, round-trip
scene/render_lineheight_test.go        # NEW — scene role-driven leading + default byte-identity
scripts/smoke/phase-31.sh              # NEW — phase smoke
docs/plans/phase-31-line-height.md     # NEW — this plan
docs/decisions.md                      # CHANGED — adds D-061
docs/glossary.md                       # CHANGED — adds "Line height"
docs/design/THEME.md                   # CHANGED — line-height token
docs/site/guide/theme.md               # CHANGED — line-height note (§19)
skills/define-a-theme/SKILL.md         # CHANGED — FontSpec.LineHeight (§19)
```

## 9. Public API surface

```go
// pptx
type FontSpec struct { /* … */ LineHeight float64 } // percent of single; 0/100 = none
type ParagraphOpts struct { /* … */ LineHeight float64 }
func (p *Paragraph) LineHeight() float64 // read accessor
```

No new node, no new token *family* beyond the typography taxonomy entry (P2).

## 10. Risks

- **R1 — invalid OOXML (bare element).** `lnSpc`/`spcPct` are new *elements* (vs
  tracking's attribute); they must carry the `a:` prefix. **Mitigation:**
  registered in `RestoreNamespaces`; the emit test asserts `<a:lnSpc>` /
  `<a:spcPct>` in the actual bytes (this caught a bare-element bug).
- **R2 — byte-identity.** **Mitigation:** 0/100 emit nothing; the default theme
  sets no line-height; a test asserts byte-identity + no `lnSpc`.
- **R3 — schema child order.** `lnSpc` must precede the bullet children in `pPr`.
  **Mitigation:** `LnSpc` is declared as the first child field; conformance/
  schema validation in preflight covers it.

## 11. Acceptance criteria

1. A paragraph with a non-single line-height emits `a:pPr/a:lnSpc/a:spcPct`
   (`a:`-prefixed, value in 1/1000 percent).
2. 0/100 line-height emits nothing and is byte-identical to today.
3. A paragraph's line-height round-trips via `Paragraph.LineHeight()`.
4. A theme whose role declares a line-height tightens that role's scene
   paragraphs; the default theme is byte-identical.
5. `make coverage` ≥ band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default new builder API |
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-31.sh` — builds; line-height emits `a:lnSpc/a:spcPct`;
0/100 byte-identical; round-trips; scene applies a role's leading.

## 14. Tests

- **Unit / golden:** `pptx` — emit (`a:`-prefixed), 0/100 byte-identity,
  round-trip via `LineHeight()`.
- **Scene:** role-driven leading emits `spcPct`; default theme no `lnSpc`.
- **Round-trip golden:** the round-trip test (G6).

## 15. Vocabulary added

- `Line height` — per-role paragraph leading (`FontSpec.LineHeight`, percent),
  emitted as `a:pPr/a:lnSpc/a:spcPct`.

## 16. Plan deviations encountered during implementation

- **Estimator feed deferred** (see §5 / D-061): leading does not yet scale
  `preferredHeight` — folded with the R9.5/R10.10 estimator rework.
- **`RestoreNamespaces` fix:** added — the new `lnSpc`/`spcPct` elements were
  emitting without the `a:` prefix (invalid OOXML); registering them is part of
  this phase.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for `pptx` + `scene`.
- [x] `scripts/smoke/phase-31.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] `make lint` clean.
- [x] Glossary + `docs/design/THEME.md` updated.
- [x] Decision entry D-061 added.
- [x] Docs site + `define-a-theme` skill updated (§19).
