# Phase 61 — prim cta button

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**RFC sections:** §11.1 (leaf node catalog), §12 (per-node rendering policy)
**Deps:** Phase 28 (Stat leaf-node wiring precedent, D-057), Phase 53 (header-pill
fit-to-label `cardPillWidthOf`, D-085), Phase 50 (auto-contrast `onCardSurface`, D-082)
**Status:** Done

---

## 1. Goal

Add a first-class `Button` scene node — a content-fit, tone-filled `RadiusFull` pill
with optional leading/trailing icons — so a deck can place a CTA/action affordance
standalone (a closing slide), inside a card body (a pricing card), or inside a banner.

## 2. Why now

Wave 12 (R12 component primitives) opens with the two CRITICAL primitives a pro deck
cannot do without; R12.1 is first because every sales/investor deck *ends* on a button
and every pricing card prices *against* one, and the IR has no button primitive at all
(no `KindButton`). It is the smallest new-node primitive in R12 (a leaf, no body), so
it also re-establishes the full new-node wiring checklist for the heavier R12 nodes
that follow. See `docs/plans/README.md` Wave 12 and D-059 (engine-scope filter).

## 3. RFC sections implemented

- `RFC §11.1` — extends the leaf-node catalog with `Button` (a presentational action
  affordance). Sibling R12 nodes (Checklist, ChipRow, Banner, …) extend it further in
  later phases.
- `RFC §12` — the per-node rendering policy: `Button` is native (shapes + custGeom
  icons), carries no `AssetID`.

## 4. Brief findings incorporated

- `docs/research/44-prim-cta-button.md` — *"Button is a leaf, content-fit, presentational
  only (no hyperlink)"* → the node carries no children and no `AssetID`; the composer
  uses only existing builder calls (no P1 builder change).
- `docs/research/44-prim-cta-button.md` — *"reuse the header-pill geometry"* → the width
  is `naturalWidth(label) + icons + 2·pad` floored at a circular minimum and clamped to
  the box, mirroring `cardPillWidthOf`, with a `fitScale` tail when the label is clamped.
- `docs/research/44-prim-cta-button.md` — *"Tone → ColorRole with a ghost outline; size →
  a pinned metric"* → tone colors are tokens (P2); the SM/MD/LG height/padding/icon
  scale is a pinned layout metric (no token), documented as such.
- `docs/research/44-prim-cta-button.md` — *"icon names flow through `walkIconRefs`"* →
  `case Button` is added to `walkIconRefs` so both icon fields are Stage-1 validated.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-057` — Stat leaf-node wiring — the leaf checklist this phase follows (no walk\* for
  Stat, but Button additionally carries icons, so it extends `walkIconRefs`).
- `D-085` — header-pill fit-to-label — the content-fit-pill geometry (`naturalWidth + pad`,
  clamp, `fitScale`) the button reuses.
- `D-082` — card-text auto-contrast — `onCardSurface` for the label on a solid fill.
- `D-026` — engine not product — the button is a mechanism (a shape); no default tone
  opinion, no hyperlink/action behavior; the soul-tokened default tone is Deckard's.
- New: `D-094` — prim-cta-button (filed in this PR).

## 7. Architecture

A new leaf node `Button` with a tone enum, a size enum, two icon-name fields, and a
per-node `Align`. The composer lives in `scene/render_button.go`:

```text
Button{Label,Tone,Size,LeadingIcon,TrailingIcon,Align}
  buttonMetrics(Size)           -> {height, padX, gap, iconSz}  (pinned)
  buttonWidthOf(label,icons,..) -> naturalWidth+icons+2*pad, [min, box.W]  (fit)
  renderButton:
    pill box (Align-offset within box.W)
    AddShape(ShapeRoundRect, WithRadius(RadiusFull), tone fill | ghost line)
    [LeadingIcon] AddIcon  | label TypeBody bold middle-anchored (onCardSurface) | [TrailingIcon] AddIcon
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — KindButton + String; ButtonTone/ButtonSize; Button struct
scene/policy.go                      # CHANGED — KindButton policy (native, no asset)
scene/validate.go                    # CHANGED — Button case (non-empty label)
scene/render.go                      # CHANGED — renderNode dispatch + preferredHeight + isFlexible(false) + nodeUsesAssets(false)
scene/render_card.go                 # CHANGED — walkIconRefs case Button (both icon fields)
scene/render_button.go               # NEW — buttonMetrics, buttonWidthOf, renderButton
scene/render_button_test.go          # NEW — white-box: width fit, tone fill, ghost line, size scale, icon validation
scene/scene_test.go                  # CHANGED — allNodes + catalog count 22 -> 23
scene/render_adversarial_test.go     # CHANGED — hostile Button in the torture fixture
scene/render_parallel_test.go        # CHANGED — determinism guard includes a Button
test/integration/roundtrip_test.go   # CHANGED — everyNodeScene + collectKinds + kind-range loop -> KindButton
scripts/smoke/phase-61.sh            # NEW — phase smoke
docs/research/44-prim-cta-button.md  # NEW — brief
docs/research/INDEX.md               # CHANGED — registers brief 44
docs/plans/phase-61-prim-cta-button.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — opens Wave 12, adds Phase 61
docs/glossary.md                     # CHANGED — Button, button tone, button size
docs/design/THEME.md                 # CHANGED — note button tone color mapping (tokens)
docs/site/...                        # CHANGED — scene catalog: Button
skills/compose-a-scene/SKILL.md      # CHANGED — Button node
docs/decisions.md                    # CHANGED — adds D-094
```

## 9. Public API surface

```go
// scene
type ButtonTone int
const ( ButtonPrimary ButtonTone = iota; ButtonAccentAlt; ButtonGhost; ButtonNeutral )

type ButtonSize int
const ( ButtonMD ButtonSize = iota; ButtonSM; ButtonLG )

type Button struct {
    Label        string
    Tone         ButtonTone
    Size         ButtonSize
    LeadingIcon  string // closed-name curated/extension icon; "" = none
    TrailingIcon string
    Align        HAlign // per-node horizontal alignment override; 0 = inherit slide
}
func (Button) NodeKind() NodeKind // KindButton

const KindButton NodeKind = ... // appended after KindStat
```

## 10. Risks

- **R1 — icon names not Stage-1 validated.** A leaf's icons are easy to miss in
  `walkIconRefs`. **Mitigation:** add `case Button` calling the check for both fields;
  a white-box test asserts an unknown button icon fails `ValidateScene`-time icon
  validation.
- **R2 — byte-identity regression.** A new node must not perturb existing decks.
  **Mitigation:** the node is purely additive (absent ⇒ no new bytes); the determinism
  guard and the integration round-trip cover output stability. (Note: a *rendered*
  button is not byte-identical to anything — there was nothing before; the byte-identity
  obligation is only the unused path.)
- **R3 — catalog-count / kind-loop drift.** **Mitigation:** bump `want 22` → `23` and
  extend the integration loop bound to `KindButton`; both fail loudly if missed.

## 11. Acceptance criteria

1. A standalone `Button` renders as a single `RadiusFull` pill sized to its label (no
   wrap/clip), with a trailing/leading glyph laid out in the same row when set.
2. A `Button` placed last in a card body sits inside the card padding (it is clamped to
   the safe area and laid out by the card body stack like any leaf).
3. Tone resolves to a token fill: Primary/AccentAlt solid + inverse label; Neutral
   `ColorSurfaceAlt`; Ghost `NoFill` + an accent hairline + accent label.
4. An unknown `LeadingIcon`/`TrailingIcon` name fails icon validation at `Render`.
5. A deck that uses no `Button` is byte-identical to the pre-feature render.
6. `scene` catalog count is 23; the integration kind-range loop covers `KindButton`.
7. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-61.sh` greps the source for the new surface (KindButton, Button
struct, ButtonTone/ButtonSize, the render dispatch + composer, the walkIconRefs case,
the policy/validate entries, the catalog count 23) and SKIPs gracefully before the
code exists. It asserts OK ≥ the acceptance-criteria count and FAIL = 0.
