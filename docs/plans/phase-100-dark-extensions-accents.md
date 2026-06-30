# Phase 100 — dark-variant accents & extensions (verify-and-close)

**Subsystem:** `scene` (VariantDark derivation) — acceptance only
**RFC sections:** §7.1 (color roles), §13.3 (theme variants)
**Deps:** Phase 97 (D-135 `DarkColors` overlay — the mechanism); brief 83
**Status:** Done

---

## 1. Goal

Confirm — and lock with acceptance goldens — that a VariantDark slide re-resolves
accent surfaces, accent text, and the engine's neutral borders to dark-variant
values rather than inheriting light-theme values, and record that the
border/accentSoft *extension tokens* are Deckard's product half.

## 2. Why now

R8.7 in priority order after the foundational R8.3 (Phase 97) it builds on. It is
MED and, on the engine side, a **verify-and-close**: the per-variant override
mechanism (`Theme.DarkColors`) and the engine's SurfaceAlt-based borders already
satisfy R8.7; the missing piece is the acceptance proof and the engine/product
decision (`DECKARD-PRODUCT-REQUIREMENTS.md` R8.7; D-059). Mirrors the R11.1 /
Phase-49 verify-and-close pattern (acceptance golden, no renderer change).

## 3. RFC sections implemented

- `RFC §7.1` / `§13.3` — accent/semantic/text roles and the engine's neutral
  borders re-resolve per VariantDark via the `DarkColors` overlay; no new code.

## 4. Brief findings incorporated

- `docs/research/83-dark-extensions-and-accents.md` — *the Phase-97 `DarkColors`
  overlay is already general (re-resolves any role per dark variant) and the
  engine's borders use `ColorSurfaceAlt` which `darkThemeFrom` dark-resolves* →
  this phase adds acceptance goldens for those behaviors, no production change.
- `docs/research/83-dark-extensions-and-accents.md` — *the border / accentSoft
  extension tokens + the derived-dark-hairline default are Deckard's* → recorded
  in D-138; not built in the engine.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-135` — soul-driven dark palette (`DarkColors`) — the override seam R8.7
  relies on; its `Surfaces`/`Text` maps re-resolve accent/semantic/text per
  variant.
- `D-026` / `D-059` — the engine preserves the brand accent on dark by default
  and lets the soul override (taste-free); the extension tokens + derived hairline
  are Deckard's product half.
- New decision **D-138** filed in this PR.

## 7. Architecture

No production change. The acceptance goldens exercise the existing
`darkThemeFrom` + `DarkColors` overlay (D-135) through the scene render:

```text
darkThemeFrom(base):
  pinned dark canvas/surface/surfaceAlt + dark text   (borders use ColorSurfaceAlt → dark)
  overlay base.DarkColors.Surfaces / .Text            (re-resolves ColorAccent / TextAccent / any role)
```

## 8. Files added or changed

```text
scene/render_dark_extensions_test.go   # NEW — R8.7 acceptance goldens
scripts/smoke/phase-100.sh             # NEW — phase smoke
docs/research/83-dark-extensions-and-accents.md   # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 83
docs/plans/phase-100-dark-extensions-accents.md   # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 15 phase entry
docs/decisions.md                      # CHANGED — adds D-138
docs/design/THEME.md                   # CHANGED — clarify accent/semantic roles are dark-overridable
skills/define-a-theme/SKILL.md         # CHANGED — clarify dark accent override
```

No public API change (verify-and-close).

## 9. Public API surface

None — no new or changed exported symbol.

## 10. Risks

- **R1 — the "no production change" claim hides a real gap** — **Mitigation:** the
  four acceptance goldens fail if any R8.7 behavior is absent (dark border
  resolution, accent/text override, byte-identity); they pass against the current
  engine, proving the mechanism is complete.

## 11. Acceptance criteria

1. A dark card's neutral border resolves to the dark `ColorSurfaceAlt` (#374151),
   never the light value (#F1F3F5).
2. `DarkColors.Surfaces[ColorAccent]` re-tints a dark accent border distinct from
   the light accent; without it, the brand light accent is preserved on dark.
3. `DarkColors.Text[TextAccent]` re-tints dark accent text.
4. A dark slide with no `DarkColors` is byte-identical to the default theme.
5. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; test-only addition |

## 13. Smoke check

`scripts/smoke/phase-100.sh` verifies the four acceptance tests pass and the
brief/decision exist.

## 14. Tests

- **Unit / round-trip golden:** `scene` black-box acceptance goldens (dark border
  resolution, accent + text override, byte-identity). No new production code.
- **Integration / Fuzz / Benchmark:** none.

## 15. Vocabulary added

none (the `Dark palette` term, D-135, already covers the mechanism).

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-100.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (no new term).
- [x] Decision entries added (D-138).
- [x] Docs site / skill clarified.
