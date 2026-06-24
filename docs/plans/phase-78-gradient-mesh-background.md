# Phase 78 — gradient-mesh background

**Subsystem:** `scene` (Layer 2 — Background)
**RFC sections:** §10.1 (backward-compat), §7.1 (token color)
**Deps:** Phase 72 (radial fill, D-106), Phase 73 (role glows, D-107); brief 61.
**Status:** Done

---

## 1. Goal

Add a `BackgroundMesh` kind so a slide can carry a soft "mesh" wash — a base
canvas plus several low-alpha caller-anchored radial glows pooled over it — for
the diffuse atmospheric light pro covers use, byte-identical when unused.

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); the single linear/radial fill
can't produce the multi-pool mesh look. Engine req R13.4 (HIGH · engine; D-059).
Composes the radial fill (R13.2) and role glows (R13.5).

## 3. RFC sections implemented

- `RFC §10.1` — a new kind appended last; existing slides unaffected.
- `RFC §7.1` — each glow color is a surface token role (P2).

## 4. Brief findings incorporated

- `docs/research/61-gradient-mesh-background.md` — *"a `BackgroundMesh` kind +
  `Mesh []MeshGlow` is the convenience"* → adopted.
- `61` — *"a base canvas under the glows makes it self-contained"* → draw
  `SolidFill(TokenColor(bg.Color))` then the glows.
- `61` — *"empty mesh → nothing (absent config)"* → empty `Mesh` = None behavior.
- `61` — *"reuse `RadialGradient` + role/alpha tokens; deterministic"* → fixed
  slice order, integer-EMU.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-106` / `D-041` — the radial fill the glows reuse.
- `D-107` — role-colored glows (the per-glow color).
- `D-112` (new) — files `BackgroundMesh` + `MeshGlow`.

## 7. Architecture

`BackgroundMesh` joins `BackgroundKind` (last). `MeshGlow{Anchor; Color
pptx.ColorRole; Radius pptx.EMU; Alpha int}` and `Background.Mesh []MeshGlow`.
`renderBackground` `case BackgroundMesh`: empty `Mesh` → the `BackgroundNone`
path; else a base `SolidFill(TokenColor(bg.Color))` (zero = `ColorCanvas`) then,
for each glow with `Radius > 0`, a radial-gradient ellipse centered on
`Anchor.Point(full)` fading `TokenColorAlpha(Color, Alpha)` → alpha 0. Fixed
order → deterministic. No new OOXML.

```text
Background{Kind: BackgroundMesh, Color: ColorPaper, Mesh: []MeshGlow{
    {Anchor: AnchorTopLeft,     Color: ColorAccent,    Radius: In(4), Alpha: 12000},
    {Anchor: AnchorBottomRight, Color: ColorAccentAlt, Radius: In(5), Alpha: 10000}}}
  → paper rect + 2 corner-pooled radial glows over it
```

## 8. Files added or changed

```text
scene/background.go           # CHANGED — BackgroundMesh kind + String + MeshGlow type + Background.Mesh field
scene/render.go               # CHANGED — renderBackground mesh case
scene/background_test.go       # CHANGED/NEW — mesh emits base + N glows; empty = none; determinism
scripts/smoke/phase-78.sh     # NEW — phase smoke
docs/research/61-gradient-mesh-background.md  # NEW — brief
docs/research/INDEX.md        # CHANGED — registers brief 61
docs/plans/phase-78-gradient-mesh-background.md  # NEW — this plan
docs/plans/README.md          # CHANGED — Phase 78 detail
docs/design/THEME.md          # CHANGED — mesh background mechanism note
docs/glossary.md              # CHANGED — mesh background term
docs/decisions.md             # CHANGED — adds D-112
docs/site/reference/scene.md  # CHANGED — BackgroundMesh + MeshGlow
skills/compose-a-scene/SKILL.md  # CHANGED — mesh background
```

## 9. Public API surface

```go
// scene
const BackgroundMesh BackgroundKind = … // appended last; base canvas + N radial glows
type MeshGlow struct { Anchor Anchor; Color pptx.ColorRole; Radius pptx.EMU; Alpha int }
// Background gains: Mesh []MeshGlow
```

`Background` is already non-comparable (the `Stops` slice); the new slice adds no
constraint. No prior surface breaks.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** a new kind appended last; existing
  kinds untouched; empty mesh = None.
- **R2 — a white glow on white reads invisible.** **Mitigation:** caller-set
  colors (opt-in); documented — the soul supplies distinct hues + subtle alpha.

## 11. Acceptance criteria

1. A `BackgroundMesh` with ≥2 glows emits a base canvas rect + one radial-gradient ellipse per glow at the specified anchors (distinguishable pools).
2. An empty `Mesh` emits no shapes on a light slide (absent-config → nothing).
3. A mesh re-renders byte-identically; an unused kind is byte-identical to today.
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-78.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `BackgroundMesh` kind + `MeshGlow` type + `Background.Mesh`.
3. `OK:` `renderBackground` mesh case.
4. `OK:` mesh emits base + glows test.
5. `OK:` empty-mesh + determinism tests.

## 14. Tests

- **Black-box (`scene_test`):** a 2-glow mesh emits a base rect + 2 radial
  `<a:gradFill>` ellipses at distinct positions; an empty mesh emits nothing;
  determinism guard.
- **Integration / Fuzz:** no.
