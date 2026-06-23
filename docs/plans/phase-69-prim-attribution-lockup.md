# Phase 69 — prim attribution lockup

**Subsystem:** scene — Layer 2 renderer (new IR leaf node, asset-bearing)
**RFC sections:** §11.1, §12, §14.3 (asset path)
**Deps:** Phase 10 (Image asset path), Phase 61 (icon glyph); brief 52.
**Status:** Done

---

## 1. Goal

Add a `Lockup` node — a caption paired with a small partner logo (an asset or an icon)
composed as one centered inline unit — for the "powered by / in partnership with"
attribution mark a branded deck places on its cover and closing.

## 2. Why now

R12.9 is the last Wave-12 primitive (R12.10 is product, skipped per D-059). The recreation
dropped the partner logo and rendered the caption as plain text. See `docs/plans/README.md`
Wave 12.

## 3. RFC sections implemented

- `RFC §11.1` — extends the leaf catalog with `Lockup`.
- `RFC §12` — policy `{HasAsset:true}` (asset → pic; icon → native, like Decoration).
- `RFC §14.3` — the asset resolve + `AddImage` path (the icon variant is media-free).

## 4. Brief findings incorporated

- `docs/research/52-prim-attribution-lockup.md` — *"a leaf with an asset OR an icon;
  `nodeUsesAssets` is `AssetID != ""`; a centered inline group; square logo (§7)"* → the
  composer and wiring.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-039` — image crop/fit — the asset `AddImage` path. `D-026` — engine not product.
  New: `D-102` — prim-attribution-lockup (this PR).

## 7. Architecture

```text
Lockup{Caption, AssetID | Icon, AssetSide, MaxHeight, Align}
  nodeUsesAssets = AssetID != ""  (asset → serial part numbering; icon → parallel)
  logoH = MaxHeight | default ; logoW = logoH (square — no pixel aspect, §7)
  group = [caption | gap | logo] (LeadCaption) | [logo | gap | caption] (TrailCaption)
  centered/aligned within the box; caption TypeCaption muted; logo via AddImage | AddIcon
```

## 8. Files added or changed

```text
scene/nodes.go             # CHANGED — KindLockup + String; AssetSide; Lockup struct
scene/policy.go            # CHANGED — KindLockup {HasAsset:true}
scene/validate.go          # CHANGED — Lockup case (asset XOR icon, side, height)
scene/render.go            # CHANGED — dispatch + preferredHeight + nodeUsesAssets(AssetID) + nodeEffectiveHAlign
scene/render_card.go       # CHANGED — walkIconRefs case Lockup
scene/render_lockup.go     # NEW — lockup composer (asset + icon paths)
scene/render_lockup_test.go ; render_lockup_render_test.go # NEW — white/black-box
scene/scene_test.go        # CHANGED — allNodes + catalog 27 -> 28
scene/render_adversarial_test.go ; test/integration/roundtrip_test.go # CHANGED
scripts/smoke/phase-69.sh  # NEW
docs/research/52-...md + INDEX.md ; docs/plans/phase-69-...md + README.md
docs/glossary.md ; docs/design/THEME.md ; docs/site/catalog/text-leaves.md ; skills/compose-a-scene/SKILL.md ; docs/decisions.md (D-102)
```

## 9. Public API surface

```go
// scene
type AssetSide int
const ( LeadCaption AssetSide = iota; TrailCaption )
type Lockup struct { Caption string; AssetID AssetID; Icon string; AssetSide AssetSide; MaxHeight pptx.EMU; Align HAlign }
func (Lockup) NodeKind() NodeKind // KindLockup
const KindLockup NodeKind = ... // appended after KindIconRows
```

## 10. Risks

- **R1 — asset part-numbering nondeterminism.** **Mitigation:** `nodeUsesAssets` true for
  the asset variant → it composes serially; a determinism test + the everyNodeScene
  round-trip cover it.
- **R2 — icon not validated.** **Mitigation:** `walkIconRefs case Lockup`; a test asserts
  an unknown icon fails `Render`.
- **R3 — ambiguous mark source.** **Mitigation:** validation requires exactly one of
  `AssetID`/`Icon`; a test covers neither/both.
- **R4 — catalog/kind drift.** **Mitigation:** count 27 → 28 and the integration loop →
  `KindLockup`.

## 11. Acceptance criteria

1. The caption and logo render as one tight centered unit (logo height-bounded, not
   stretched beyond its box); the group centers as a whole.
2. An icon-based lockup is media-free and byte-identical across workers; an asset lockup
   registers one asset.
3. A lockup with neither or both of asset/icon, or an unknown icon, fails Stage-1.
4. A deck without a `Lockup` is byte-identical; catalog is 28; the kind loop covers it.
5. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-69.sh` greps the new surface (KindLockup, the struct, the composer,
policy/validate/walkIconRefs entries, catalog 28) and runs the white/black-box tests; SKIPs
gracefully before the code exists. OK ≥ the acceptance-criteria count, FAIL = 0.
