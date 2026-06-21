# Phase 30 — type tracking

**Subsystem:** pptx — Layer 1 builder (`RFC §3.3`, §7 theme/typography)
**RFC sections:** §7 (theme tokens), §9 (rich text / runs)
**Deps:** Phase 02 (theme + FontSpec), Phase 04 (rich text). External: none.
**Status:** In progress

---

## 1. Goal

Add a per-type-role letter-spacing (**tracking**) token to `FontSpec` (with an
optional per-run override), emitted as the OOXML `a:rPr/@spc` attribute, so a
soul can open up eyebrows and tighten display headlines — additive, round-trip
clean, byte-identical when zero.

## 2. Why now

First engine unit of **Wave 9** (typography & type system), the foundational
wave of the R8–R14 professional-bar work (D-059). `DECKARD-PRODUCT-REQUIREMENTS.md`
R9.3 (HIGH · engine): tracked-caps eyebrows are the biggest "designed vs default"
tell and the engine has no tracking at all. Tracking is the cleanest run-attribute
token and sets the pattern for its R9 siblings (line-height, case).

## 3. RFC sections implemented

- `RFC §7` — extends the resolved type scale (`FontSpec`) with a tracking value.
- `RFC §9` — an optional per-run tracking override on `RunStyle`, round-tripped
  through the read model (G6).

## 4. Brief findings incorporated

- `docs/research/17-type-detail-tokens.md` — *tracking belongs on `FontSpec`
  (per role) first* → `FontSpec.Tracking float64`; `RunStyle.Tracking *float64`
  override (nil = inherit role).
- *emit is one attribute, default-off* → `toProps` emits `spc` when non-zero,
  nothing when zero (byte-identical).
- *round-trip is a struct field + a reader* → `XTextProperties.Spc` + a
  `*Run.Tracking()` accessor.
- *deterministic* → `round(pt × 100)`, pure.

## 5. Findings I'm departing from

None. The brief defers line-height (R9.4), case (R9.11), per-face metrics (R9.5),
and a scene-side per-run override to their own phases.

## 6. Decisions referenced

- `D-059` — Wave-2 engine scope (this is an `engine`-tagged unit).
- `D-012`/`D-030` — token resolution; tracking is part of the resolved `FontSpec`.
- `G6` (`CLAUDE.md §6`) — round-trip fidelity; tracking round-trips losslessly.
- Files **D-060 — tracking token** in `docs/decisions.md`.

## 7. Architecture

```text
pptx/theme.go        FontSpec += Tracking float64           (points; 0 = none)
pptx/text.go         RunStyle += Tracking *float64           (nil = inherit role)
                     (*Run).Tracking() float64               (read accessor)
internal/ooxml/slide XTextProperties += Spc int             (a:rPr/@spc, omitempty)
pptx/text_layout.go  toProps: spc = round(effective pt ×100) when non-zero
docs/design/THEME.md tracking token in the typography taxonomy (P2)
```

## 8. Files added or changed

```text
pptx/theme.go                          # CHANGED — FontSpec.Tracking
pptx/text.go                           # CHANGED — RunStyle.Tracking + Run.Tracking()
pptx/text_layout.go                    # CHANGED — emit a:rPr/@spc
internal/ooxml/slide/slide_types.go    # CHANGED — XTextProperties.Spc
pptx/text_tracking_test.go             # NEW — emit, round-trip, override, zero-identity, determinism
scripts/smoke/phase-30.sh              # NEW — phase smoke
docs/research/17-type-detail-tokens.md # NEW — informing brief
docs/research/INDEX.md                 # CHANGED — registers brief 17
docs/plans/phase-30-type-tracking.md   # NEW — this plan
docs/plans/README.md                   # CHANGED — Waves 9–15 map + R8–R14 framing
docs/decisions.md                      # CHANGED — adds D-059 (scope) + D-060 (tracking)
docs/design/THEME.md                   # CHANGED — tracking token taxonomy entry
docs/glossary.md                       # CHANGED — adds "Tracking"
docs/site/guide/theme.md               # CHANGED — tracking note (§19)
skills/define-a-theme/SKILL.md         # CHANGED — FontSpec.Tracking (§19)
```

## 9. Public API surface

```go
// pptx
type FontSpec struct {
    Family   string
    Size     float64
    Weight   int
    Italic   bool
    Tracking float64 // letter-spacing in points (signed); 0 = none
}

type RunStyle struct {
    // ... existing ...
    Tracking *float64 // optional per-run override; nil = inherit the role's tracking
}

func (r *Run) Tracking() float64 // read accessor: the run's resolved tracking (pt)
```

New builder visual property ⇒ a `docs/design/THEME.md` token entry (P2) and a
round-trip golden land in this PR.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** `spc` is `omitempty` and emitted only
  when effective tracking ≠ 0; a test renders a no-tracking deck and asserts no
  `spc` attribute / byte-identity.
- **R2 — round-trip loss.** **Mitigation:** `Spc` on the struct survives parse +
  re-marshal; a round-trip test reopens a tracked run and asserts `Tracking()`.
- **R3 — unit/sign errors.** **Mitigation:** points × 100 (signed), tested for a
  positive and a negative value against the emitted `spc`.

## 11. Acceptance criteria

1. A role/run with non-zero tracking emits `a:rPr` with `spc="<pt×100>"` (signed).
2. A run with no tracking (role 0, override nil) emits no `spc` and the deck is
   byte-identical to today.
3. A `RunStyle.Tracking` override wins over the role's `FontSpec.Tracking`.
4. A tracked run round-trips: reopened, `Run.Tracking()` equals the authored pt.
5. Output is deterministic across worker counts (N/A for a single deck; the value
   is a pure function of inputs).
6. `make coverage` ≥ band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default new-API on the builder (no override) |

## 13. Smoke check

`scripts/smoke/phase-30.sh`:
1. `OK:` library builds CGo-free.
2. `OK:` tracking emits `a:rPr/@spc` (criterion 1).
3. `OK:` zero tracking is byte-identical (criterion 2).
4. `OK:` tracking round-trips via `Run.Tracking()` (criterion 4).

## 14. Tests

- **Unit / golden:** `pptx` — emit `spc` (positive + negative), zero-value
  byte-identity, run-override-wins.
- **Round-trip golden:** a tracked run reopens with the same `Tracking()` (G6).
- **Integration / fuzz / bench:** none.

## 15. Vocabulary added

- `Tracking` — per-type-role letter-spacing (`FontSpec.Tracking`, points, signed),
  emitted as OOXML `a:rPr/@spc`; opens up eyebrows / tightens display.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `pptx`.
- [ ] `scripts/smoke/phase-30.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] `make lint` clean.
- [ ] Glossary + `docs/design/THEME.md` updated.
- [ ] Decision entries D-059 + D-060 added.
- [ ] Docs site + `define-a-theme` skill updated (§19).
