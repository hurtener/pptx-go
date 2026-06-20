# Phase 24 — slide chrome

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §10.2 (body region / slot assignment), §10.6 (asset resolution)
**Deps:** Phase 06 (leaf composers), Phase 11 (Image/asset resolution). External:
none.
**Status:** In progress

---

## 1. Goal

Add opt-in per-slide chrome — a top section eyebrow + hairline rule and a bottom
footer with a brand slot and an `N / total` page number — drawn outside a shrunk
body region, so a deck reads "designed" without the caller hand-placing
furniture on every slide.

## 2. Why now

Third unit of **Wave 8 — post-V1 engine extensions**
(`DECKARD-PRODUCT-REQUIREMENTS.md` R3, MEDIUM), picked up after R1/R2 per the
one-requirement-per-PR cadence. The product's reference decks carry consistent
chrome on every slide; the engine has no chrome concept, so the caller can't
produce it. R3 adds the *mechanism* (a chrome toggle + slots); the caller drives
the taste (brand, section names, whether to enable it at all).

## 3. RFC sections implemented

- `RFC §10.2` — the body region is the slot space; this phase shrinks it when
  chrome is enabled and draws the bands in the reclaimed margin, so chrome never
  overlaps content.
- `RFC §10.6` — the optional brand *asset* resolves through the existing
  `AssetResolver`; an unresolved brand degrades to a `LayoutWarning`
  (warn-don't-fail).

## 4. Brief findings incorporated

- `docs/research/11-slide-chrome.md` — *chrome belongs outside the body region,
  and the body must shrink* → `bodyRegion` shrinks by fixed band heights when
  `Chrome.Enabled`; overlap is structurally impossible.
- `docs/research/11-slide-chrome.md` — *the field split writes itself* →
  `Scene.Chrome` (brand slot + total + `Enabled`); `SceneSlide.Section` +
  `SceneSlide.PageNumber`.
- `docs/research/11-slide-chrome.md` — *auto-derive but overridable numbering* →
  `Total` defaults to `len(Slides)`, page number defaults to scene position
  (1-based); both overridable.
- `docs/research/11-slide-chrome.md` — *brand is text-or-asset; only the asset
  registers media* → brand text renders in parallel; a brand asset forces
  sequential composition for stable part numbering.
- `docs/research/11-slide-chrome.md` — *tokens, not literals* → chrome uses
  `TextMuted` + `ColorSurfaceAlt`; **no new token**, so no `THEME.md` entry.

## 5. Findings I'm departing from

None. The brief's open-questions (per-slide opt-out, custom page-number format,
extra footer slots, master-placeholder chrome) are explicitly deferred there.

## 6. Decisions referenced

- `D-026` — *Engine, not product.* Chrome is a caller-driven mechanism: the
  engine draws the bands it's told to and composes `N / total`, but invents no
  brand, no section names, and no decision that a deck "needs" chrome.
- This plan files **D-053 — opt-in slide chrome** in `docs/decisions.md`.

## 7. Architecture

```text
scene/scene.go    Chrome struct (Enabled, Brand, BrandAsset, Total)   NEW type
                  Scene.Chrome Chrome                                  NEW field
                  SceneSlide.Section string, .PageNumber int           NEW fields

scene/render.go   renderer: + chrome Chrome, chromeTotal int, slideIndex int
                  composeOne(ps, sl, idx)  — threads the scene index
                  composeSlide: renderChrome(ps, sl) after the node loop
                  bodyRegion: shrink top by eyebrow band, bottom by footer band
                              when chrome.Enabled

scene/chrome.go   renderChrome / renderChromeBrand  — native shapes (+ optional
                  brand image); pinned EMU band constants                NEW file
                  chromeTotalFor(Scene), chromeNeedsSerial(Scene)        helpers

scene/scene.go    Render: base carries Chrome + chromeTotal; the free/serial
                  split also forces serial when chromeNeedsSerial (brand asset).
```

Chrome is drawn **after** the body nodes so the page number stays visible even
over a full-bleed background or section divider; it occupies the margin the body
region vacated, so it never collides with content.

## 8. Files added or changed

```text
scene/scene.go                       # CHANGED — Chrome type, Scene.Chrome, SceneSlide.Section/PageNumber, Render wiring
scene/render.go                      # CHANGED — renderer chrome fields, composeOne idx, composeSlide call, bodyRegion shrink
scene/chrome.go                      # NEW — renderChrome + brand slot + band constants + helpers
scene/render_chrome_test.go          # NEW — eyebrow/footer present, body shrink, disabled identity, total, brand, determinism
scripts/smoke/phase-24.sh            # NEW — phase smoke
docs/research/11-slide-chrome.md     # NEW — informing brief
docs/research/INDEX.md               # CHANGED — registers brief 11
docs/plans/phase-24-slide-chrome.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — adds Phase 24 to Wave 8
docs/decisions.md                    # CHANGED — adds D-053
docs/glossary.md                     # CHANGED — adds "Slide chrome", "Section eyebrow", "Footer page number"
docs/site/guide/scene.md             # CHANGED — chrome section (§19)
skills/compose-a-scene/SKILL.md      # CHANGED — chrome fields (§19)
```

## 9. Public API surface

```go
// scene (scene.go)
type Chrome struct {
    Enabled    bool    // master switch; zero value false = no chrome (byte-identical)
    Brand      string  // footer-left brand text (used when BrandAsset is empty)
    BrandAsset AssetID // footer-left brand image, resolved via the AssetResolver
    Total      int     // page-number denominator; 0 = len(Scene.Slides)
}

type Scene struct {
    // ... existing ...
    Chrome Chrome // optional opt-in slide chrome; zero value = disabled
}

type SceneSlide struct {
    // ... existing ...
    Section    string // chrome: top eyebrow label; empty = no eyebrow on this slide
    PageNumber int    // chrome: the N in "N / total"; 0 = scene position (1-based)
}
```

New public scene surface (a struct + fields) ⇒ a smoke check lands in this PR
(§4.2/§13). No new builder API and no new scene IR node; no round-trip golden is
required (no emitted builder primitive). No new theme token (P2).

## 10. Risks

- **R1 — chrome overlaps content.** **Mitigation:** `bodyRegion` shrinks by the
  band heights whenever chrome is enabled, so the body never extends into the
  bands; a test asserts the shrunk region's top/bottom and a body node's box
  staying within it.
- **R2 — non-determinism via the brand image.** **Mitigation:** a brand asset is
  the only global-media touch; `chromeNeedsSerial` forces the whole deck
  sequential when it is set, so its part number is stable. A determinism test
  covers both brand-text (parallel) and brand-asset (serial) chrome.
- **R3 — backward-compat regression.** **Mitigation:** every new field's zero
  value is inert; `renderChrome` returns immediately and `bodyRegion` does not
  shrink when `Chrome.Enabled` is false. A test asserts a chrome-disabled scene
  is byte-identical to the same scene rendered without the fields set.

## 11. Acceptance criteria

1. A 3-slide deck with `Chrome.Enabled` renders, on each slide, a footer page
   number (`1 / 3`, `2 / 3`, `3 / 3`) and — where `Section` is set — a top
   eyebrow label, both outside the body box.
2. With chrome enabled the body region shrinks (its top is below `bodyMargin`
   and its bottom is above `slideHeight − bodyMargin`), and a body node sits
   inside the shrunk region.
3. `Total` defaults to `len(Slides)` and is overridable; a slide's page number
   defaults to its 1-based scene position and is overridable.
4. A brand text renders a left footer label; a brand asset renders a left footer
   image (and an unresolved brand asset warns, doesn't fail).
5. Chrome disabled (zero `Chrome`) renders byte-identical to today.
6. A chrome deck renders byte-identical across 1 vs N workers — for both brand
   text and brand asset.
7. `make coverage` shows `scene` ≥ its band; `make preflight` passes.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; new branches covered by
`render_chrome_test.go`.

## 13. Smoke check

`scripts/smoke/phase-24.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` chrome renders the page number + eyebrow on each slide (criterion 1).
3. `OK:` body region shrinks under chrome (criterion 2).
4. `OK:` total / page-number derivation + override (criterion 3).
5. `OK:` brand text + brand asset (incl. unresolved warn) (criterion 4).
6. `OK:` chrome disabled is byte-identical (criterion 5).
7. `OK:` chrome render is deterministic across workers (criterion 6).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — `chromeTotalFor`, `bodyRegion` shrink, page-number
  derivation (white-box); black-box render assertions on the emitted slide XML
  (page-number string, eyebrow label present/absent).
- **Round-trip golden:** N/A — no builder primitive / scene node added.
- **Integration** (`test/integration/`): no — internal to `scene`; the brand
  asset reuses the already-integration-tested `AssetResolver` seam.
- **Fuzz:** no — no parse/decode surface.
- **Benchmark:** optional — chrome adds a small fixed per-slide cost; not a gate.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Footer page number` — the `N / total` page indicator chrome draws bottom-right
  on every chrome-enabled slide.
- `Section eyebrow` — the top chrome band: a per-slide section label + hairline
  rule, drawn only when the slide sets a `Section`.
- `Slide chrome` — opt-in recurring per-slide furniture (eyebrow + footer) drawn
  outside a shrunk body region, driven by `Scene.Chrome` and `SceneSlide` fields.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-24.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entry D-053 added.
- [ ] Docs site updated for the chrome surface (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
