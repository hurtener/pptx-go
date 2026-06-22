# Phase 50 — card-text auto-contrast (R11.2)

**Subsystem:** scene — Layer 2 renderer (card / container chrome)
**RFC sections:** §7.1, §7.4, §12.1
**Deps:** Phase 25 (D-054 card visuals), Phase 29 (D-058 resolved colors), brief 33.
**Status:** Done

---

## 1. Goal

Make every card/container chrome text run legible against whatever surface it sits
on — a dark card fill, a dark-variant slide, or a same-hue header band — via a
deterministic, pinned auto-contrast mechanism the caller can still override, with
the common light-surface card byte-identical to today.

## 2. Why now

R11.2 is the second CRITICAL of Wave 11 and the most-reported robustness bug class
(recreation slides 2, 3, 7: black-on-dark headers, invisible same-hue eyebrows). It
is the first *new mechanism* of the wave. It tensions with D-058 (the engine ships
no contrast logic), so it needs a brief (33) and a decision (D-082) reconciling the
two under D-026 (mechanism, not policy).

## 3. RFC sections implemented

- `RFC §7.1`, `§7.4` — token taxonomy + resolution: the mechanism resolves a
  surface `ColorRole` to RGB and picks the contrast-correct text token.
- `RFC §12.1` — card chrome legibility (the header / eyebrow / pill runs).

## 4. Brief findings incorporated

- `docs/research/33-card-text-auto-contrast.md`:
  - "nil on a light surface is the byte-identical lever" → `onCardSurface` returns
    nil (leave `Color` unset, inherit the dark default) on light surfaces and a
    light token only on dark ones; light cards are byte-identical by construction.
  - "sRGB relative luminance, pinned, integer per call" → a 256-entry `srgbLinear`
    table built once at init; per-call luminance is pure integer (worker-count
    independent).
  - "the dark threshold is the black/white crossover (L ≈ 0.179)" → pinned
    `darkSurfaceLumaMax = 17912`, which guarantees both branches clear ~4.58:1.
  - "eyebrow keeps the accent only when it clears 4.5:1" → `accentLegible`; the
    default accent on a white card passes (5.17:1) so the common eyebrow is
    byte-identical.
  - "apply where the surface is known" → card chrome (header/eyebrow/pill),
    join-badge; Stat value against the slide variant surface; `TextMuted` labels
    left as a deliberate mid-gray.

## 5. Findings I'm departing from

*"none"* — but one brief recommendation is scoped to a follow-up: threading a
container surface into leaf renderers (so a Stat inside a strongly-colored card
contrasts against the card fill, not the slide surface). The brief flags it as a
documented limitation; this phase ships the slide-surface estimate.

## 6. Decisions referenced

- `D-026` — engine, not product. The auto-contrast is a *mechanism* the caller
  drives / overrides (an explicit `Color` wins), not a legibility *policy*.
- `D-058` — resolved-color exposure with no contrast logic. D-082 records the
  reconciliation: a pinned, overridable token picker is not opinionated taste.
- `D-054` — the card visuals (`HeaderFill`/eyebrow/pill) whose runs gain contrast.
- `D-082` — **new** — the auto-contrast mechanism and its byte-identical guarantee.

## 7. Architecture

```text
scene/contrast.go (NEW)
  srgbLinear[256]            built once at init (math.Pow), integer-scaled
  relLuminance(RGB) int      WCAG sRGB relative luminance, [0,100000]
  contrastRatioT10(a,b) int  WCAG contrast ratio ×10
  (r) onCardSurface(bg) Color   light token on dark surface, nil on light
  (r) accentLegible(bg) bool    accent clears 4.5:1 against bg?

renderCardChrome  → header title / pill → onCardSurface; eyebrow → accent-then-onCardSurface
renderColumnJoin  → join label → onCardSurface(ColorAccent)  (nil → TextPrimary)
renderStat        → value → onCardSurface(ColorCanvas)
```

## 8. Files added or changed

```text
scene/contrast.go                  # NEW — onCardSurface, relLuminance, accentLegible, pinned consts
scene/render_card.go               # CHANGED — header/eyebrow/pill auto-contrast against the header surface
scene/render_container.go          # CHANGED — join-badge label auto-contrast against the accent fill
scene/render_stat.go               # CHANGED — Stat value auto-contrast against the slide surface
scene/contrast_test.go             # NEW — white-box mechanism + contrast-guarantee sweep
scene/render_contrast_test.go      # NEW — black-box render: dark white / light byte-identical / eyebrow fallback
scene/render_parallel_test.go      # CHANGED — TestRenderDeterministic_AutoContrast guard
scripts/smoke/phase-50.sh          # NEW — phase smoke
docs/research/33-card-text-auto-contrast.md  # NEW — brief 33
docs/research/INDEX.md             # CHANGED — registers brief 33
docs/plans/phase-50-card-text-auto-contrast.md  # NEW — this plan
docs/plans/README.md               # CHANGED — Wave 11 Phase 50
docs/decisions.md                  # CHANGED — adds D-082
docs/glossary.md                   # CHANGED — adds "auto-contrast" / "relative luminance"
docs/site/...                      # CHANGED — compose-a-scene note on auto-contrast
skills/compose-a-scene/SKILL.md    # CHANGED — auto-contrast note
```

No new exported symbol → no public-API surface change (the mechanism is internal
to the renderer).

## 9. Public API surface

None. `onCardSurface` / `relLuminance` / `accentLegible` are unexported renderer
internals. The user-visible effect is purely a rendering-quality improvement;
callers needing a specific color keep using the existing explicit `Color` path.

## 10. Risks

- **R1 — a contrast proxy that breaks byte-identical.** A linear-luma proxy would
  under-weight blue and flip a legible blue-on-white accent. **Mitigation:** true
  sRGB gamma luminance (verified: `2563EB` on white = 5.17:1 → accent kept →
  byte-identical); the full suite of existing goldens passes unchanged.
- **R2 — leaf nodes don't know their container surface.** A Stat in a colored card
  contrasts against the slide surface, not the card. **Mitigation:** documented as
  a follow-up; the slide-surface estimate fixes the reported (dark-variant) cases
  and is byte-identical on light.
- **R3 — float non-determinism.** **Mitigation:** the only float use is the init
  table build; every per-render decision is integer; a parallel determinism guard
  asserts byte-identical across worker counts.

## 11. Acceptance criteria

1. For a card with any fill and any slide variant, every auto-contrasted chrome run
   (header, pill, Stat value) clears ≥ 4.5:1 against its surface
   (`TestOnCardSurface_ContrastGuarantee` sweeps light/dark/brand surfaces).
2. A light-surface card is byte-identical: `onCardSurface` returns nil, the header
   run carries no explicit color (`TestCardHeader_AutoContrast_LightByteIdentical`),
   and the full existing golden suite passes unchanged.
3. A dark-variant card and a dark-fill card both flip the header to a light color
   (`TestCardHeader_AutoContrast_DarkVariant`, `_DarkFill`).
4. The eyebrow keeps its accent on a light card and drops it on a same-hue band
   (`TestCardEyebrow_AccentFallback`).
5. The path is deterministic across worker counts
   (`TestRenderDeterministic_AutoContrast`).

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; `contrast.go` is unit-covered + render-covered |

## 13. Smoke check

`scripts/smoke/phase-50.sh` runs the seven acceptance tests (contrast guarantee,
nil-on-light, accent legibility, dark-variant white header, light byte-identical,
eyebrow fallback, determinism). All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`contrast_test.go`).
- **Round-trip golden:** n/a — no builder API; the existing scene goldens prove
  byte-identical on light.
- **Integration:** no — same subsystem.
- **Fuzz:** no.
- **Benchmark:** no.

## 15. Vocabulary added

- `auto-contrast` — the engine mechanism (`onCardSurface`) that picks a chrome
  run's text color from the luminance of the surface behind it; opt-out via an
  explicit `Color`; byte-identical on a light surface.
- `relative luminance` — the WCAG sRGB perceptual brightness of a color, the basis
  for the auto-contrast decision.

## 16. Plan deviations encountered during implementation

- Stat value auto-contrast resolves against the slide variant surface
  (`ColorCanvas`), not the (unknown) container surface — leaf nodes don't receive
  their surface through `renderNode`. Documented as a follow-up (brief 33 §5);
  acceptance criterion 1 is stated against the *surface the run is given*.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-50.sh` reports `OK ≥ 7` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-082).
- [x] Docs site updated for user-facing surface changes (auto-contrast note).
- [x] Affected agent skill(s) updated (compose-a-scene note).
