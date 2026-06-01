# Phase 11 — Image node + media manager refactor

**Subsystem:** scene (image) + pptx (media)
**RFC sections:** §8.6, §11.1 (image)
**Deps:** Phase 10 (frame chrome — `render_image.go` and the frame interior this
phase composes crop/fit with).
**Status:** Done

---

## 1. Goal

The scene `Image` node can express a per-edge crop and a fill mode, and
`render_image.go` applies them to the composed picture — composing with the
frame interior — through the builder's existing crop/fit mechanism.

## 2. Why now

The master plan (Wave 3) sequences the image node and the "media manager
refactor" as Phase 11, after frame chrome (Phase 10). On inspection, the
**builder half is already delivered**: the media manager (content-hash dedup,
global pool, deterministic ordering), alt text, crop, fit, and MIME sniffing all
shipped with the foundation builder phase (master plan line 254:
*"`pptx/media.go` — Image, ImageSource, ImageFile, ImageBytes, ImageReader,
alt-text, crop, fit"*) and are tested (`test/parts/media_manager_dedup_test.go`,
`test/pptx/media_test.go`). Phase 10 then wired scene asset resolution, alt text,
and frame composition. The single remaining gap is scene-side: the `Image` IR
node could not express **crop** or **fit**, so a scene caller could not drive the
builder's `SetCrop`/`SetFit`. Phase 11 closes that gap and consolidates the
seam's round-trip/dedup coverage at the scene level. (See **D-039** for the
drift reconciliation.)

## 3. RFC sections implemented

- `RFC §8.6` — **Media, dedup, alt text.** Already implemented in the builder
  (dedup pool, `SetAltText`, `SetCrop`, `SetFit`). Phase 11 adds **no builder
  code**; it surfaces the existing crop/fit mechanism through the scene IR. The
  §8.6 example's `FitCover` is **not** in V1 (cover/contain need pixel
  dimensions, forbidden by §7) — consistent with the builder's `FitFill`/
  `FitNone` and D-026; recorded in D-039.
- `RFC §11.1` — **Image node.** The scene `image` node gains the crop/fit
  composition the master plan names for `render_image.go`. Frame chrome (the
  other §11.1 image facet) landed in Phase 10.

## 4. Brief findings incorporated

No informing brief. This phase wires a mechanism that **already exists in the
repository** — the builder's `SetCrop`/`SetFit` (`pptx/media.go`) and their
OOXML mapping (`internal/ooxml/slide`: `XSrcRect`, `XStretchProperties`) — into
the scene IR. There is no prior art to investigate; the design is determined by
the existing builder surface and RFC §8.6/§11.1. (Per `CLAUDE.md §16`, a phase
with no informing brief states so explicitly; this is that statement.)

## 5. Findings I'm departing from

None (no brief). The one **plan deviation** is the media-manager half of the
master-plan scope, reconciled in §6/D-039 rather than departed from a brief.

## 6. Decisions referenced

- **D-039 (NEW, filed in this PR)** — *Phase 11 media work is already delivered;
  scene Image gains crop/fit; no media-manager relocation.* (a) The media
  manager, dedup pool, alt text, crop, fit, and MIME detection shipped with the
  foundation builder phase and are tested, so Phase 11 adds **no builder media
  code**. (b) The master plan's speculative relocation of the dedup pool to
  "`internal/opc` or a new `internal/media`" is **declined**: the manager
  orchestrates `Slide`/`Presentation` state and is correctly placed in `pptx`,
  while the wire type (`MediaResource`) already lives in `internal/ooxml/media`
  (the P3-isolated seam) — relocating working, tested code would be churn with
  no functional gain. (c) The scene `Image` IR gains `Crop` and `Fit` as
  **mechanism exposure** (engine, not product — D-026); `Fit` is limited to
  `FitFill`/`FitNone` (cover/contain forbidden by §7), matching the builder.
- `D-026` — *Engine, not product.* Crop and fit are mechanisms the caller drives;
  pptx-go reads no pixel data and makes no aspect-aware decision. An over-crop
  is a structural error (Stage-1), not a silently "fixed" image.
- `D-035` — *Byte-identical idempotency.* The default (no crop, `FitFill`)
  re-renders byte-identically to a Phase-10 image (the builder already sets the
  stretch fill by default, so `SetFit(FitFill)` is idempotent).
- `D-038` — *Frame reference.* Crop/fit compose with the frame interior: the
  recipe returns the interior box, and the cropped/fitted image is placed into
  it.

## 7. Architecture

```text
scene.Image{ AssetID, Alt, Frame/FrameName, Crop, Fit }   // Crop, Fit are NEW
         │  (Crop = pptx.Crop, Fit = pptx.Fit — re-exported, like the tokens)
         ▼
renderImage(ps, box, v, slideID):
    interior := frame?(recipe(ps, box)) : box        // Phase 10
    img := ps.AddImage(ImageBytes(bytes, ct), interior)
    img.SetAltText(v.Alt)                            // Phase 10
    if v.Crop != (Crop{}) { img.SetCrop(v.Crop) }    // NEW → <a:srcRect>
    img.SetFit(v.Fit)                                // NEW → <a:stretch> on/off
         ▼
builder (unchanged): SetCrop → BlipFill.SrcRect ; SetFit → BlipFill.Stretch

Validation (Stage 1): an Image's Crop must be in range — each edge in [0,1],
L+R < 1, T+B < 1 — else the composed source rect is degenerate.
```

No builder change. No new package. The scene `Image` node grows two fields and
`render_image.go` two calls; `validate.go` grows a crop-range check.

## 8. Files added or changed

```text
scene/nodes.go                 # CHANGED — Image gains Crop, Fit; Crop/Fit re-export aliases
scene/render_image.go          # CHANGED — apply SetCrop (when non-zero) + SetFit
scene/validate.go              # CHANGED — Stage-1 crop-range check for Image
scene/render_image_test.go     # CHANGED — crop, fit-none, fit-fill, crop+frame, invalid-crop
test/integration/frame_image_test.go  # CHANGED — composite (frame+crop+fit+alt) round-trip + determinism
scripts/smoke/phase-11.sh      # NEW — phase smoke
docs/decisions.md              # CHANGED — adds D-039
docs/glossary.md               # CHANGED — Crop, Fit (scene)
docs/plans/phase-11-image-node.md     # NEW — this plan
```

No `pptx/media.go` change (the media work is already delivered — D-039). No
`docs/design/THEME.md` change (crop/fit are not theme tokens). No user-facing
doc-site / skills updates (inert pre-Phase 20).

## 9. Public API surface

```go
// scene — Image gains crop + fit (mechanism exposure of the builder's, D-039).
type Image struct {
    node
    AssetID   AssetID
    Alt       string
    Frame     FrameKind
    FrameName string
    Crop      Crop // NEW — per-edge fractional crop (0..1); zero value = no crop
    Fit       Fit  // NEW — FitFill (default, stretches) or FitNone
}

// scene — re-exported from the builder so the IR uses the same vocabulary.
type Crop = pptx.Crop // struct{ Left, Top, Right, Bottom float64 }
type Fit  = pptx.Fit
const (
    FitFill = pptx.FitFill // zero value; stretches the image to fill its box
    FitNone = pptx.FitNone // omits the stretch fill
)
```

No prior public surface breaks: both fields are additive and their zero values
(`Crop{}`, `FitFill`) reproduce the Phase-10 behavior exactly.

## 10. Risks

- **R1 — Determinism regression.** Always calling `SetFit` could perturb the
  default output. **Mitigation:** the builder's `AddPicture` already sets the
  stretch fill, so `SetFit(FitFill)` is idempotent; the integration test asserts
  a default-fit framed image is byte-identical to its re-render, and the Phase-10
  determinism test still passes.
- **R2 — Degenerate crop.** `Left+Right ≥ 1` (or top+bottom) yields an empty
  source rect PowerPoint may reject. **Mitigation:** Stage-1 validation rejects
  out-of-range and over-crop before render; a unit test covers it.
- **R3 — Scope drift the other way.** Adding `Fit` when only `FitFill`/`FitNone`
  exist risks implying cover/contain support. **Mitigation:** godoc + D-039
  state cover/contain are forbidden by §7; the enum is the builder's two values,
  re-exported, not a new set.

## 11. Acceptance criteria

1. A scene `Image` with a non-zero `Crop` emits an `<a:srcRect>` with the
   expected per-edge permille; a zero `Crop` emits none.
2. `Fit == FitNone` omits the `<a:stretch>` fill; `Fit == FitFill` (default)
   keeps it.
3. Crop and fit compose with a frame: a framed, cropped, fitted image places the
   cropped picture inside the bezel interior and renders conformantly.
4. An `Image` whose `Crop` is out of range (an edge outside [0,1], or
   `L+R ≥ 1` / `T+B ≥ 1`) fails Stage-1 validation with a clear error.
5. Inserting the same asset twice (same `AssetID`/bytes) writes **one** media
   part (dedup) at the scene seam; both slides reference it.
6. A composite scene image (frame + crop + fit + alt) renders to a conformant
   deck, re-renders **byte-identically** (D-035), and reopens through
   `pptx.NewFromBytes`; `make test -race` and `make coverage` pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for scene packages (unchanged band) |

No new package. The added scene fields/branches are covered by the
`render_image_test.go` additions.

## 13. Smoke check

`scripts/smoke/phase-11.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` crop emits `<a:srcRect>`; zero crop emits none (`scene` test).
3. `OK:` `FitNone` omits stretch; `FitFill` keeps it (`scene` test).
4. `OK:` frame + crop + fit composite renders (`scene` test).
5. `OK:` out-of-range crop fails Stage-1 validation (`scene` test).
6. `OK:` same asset twice → one media part (dedup) at the scene seam
   (`scene` test).
7. `OK:` composite image deck round-trip + byte-identical + conformant
   (`test/integration`).

## 14. Tests

- **Unit:** `scene` — crop srcRect emission, fit stretch on/off, crop+frame
  composite, crop-range validation, scene-seam dedup.
- **Round-trip golden:** scene-level (render → assert wire props → reopen →
  byte-identical). Public model read-back is **Phase 18**, not pulled forward.
- **Integration** (`test/integration/`): yes — extends the Phase-10 framed-image
  seam test with crop + fit; real `internal/opc` write + `encoding/xml` decode +
  temp-file round-trip, `-race`.
- **Fuzz:** none (no new parse/decode surface; the image sniffer is unchanged
  and already in the builder).
- **Benchmark:** none (no hot reusable artifact added).

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Crop` (scene) — per-edge fractional image crop, re-exported from the builder;
  drives the OOXML `srcRect`.
- `Fit` (scene) — image fill mode (`FitFill`/`FitNone`), re-exported from the
  builder; cover/contain are not in V1 (§7).

## 16. Plan deviations encountered during implementation

- **The "media manager refactor" half of the master-plan scope was already
  delivered** (foundation builder phase) — Phase 11 adds no builder media code
  and declines the speculative `internal/media` relocation. Reconciled in
  §6/**D-039** before implementation, per `CLAUDE.md §4.3`.
- *(further deviations appended as implementation happens)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages (scene 90.66%).
- [x] `scripts/smoke/phase-11.sh` reports `OK ≥ 7` and `FAIL = 0` (7 OK, 0
      FAIL).
- [x] Prior phases' smoke scripts still pass (preflight PASS).
- [x] Glossary updated (Crop, Fit).
- [x] Decision entries added (D-039).
- [x] (Phase 20+) Docs site updated — N/A (inert pre-Phase 20).
- [x] (Phase 20+) Affected agent skill(s) updated — N/A (inert pre-Phase 20).
