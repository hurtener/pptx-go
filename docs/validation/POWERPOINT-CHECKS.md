# Manual PowerPoint validation — per-wave reference-deck check

The ground-truth layer of the validity strategy (D-031, layer 4). The three
automated layers — `internal/conformance` (OPC integrity), `xmllint` schema
validation, and the LibreOffice headless open-proxy — approximate what they
can in CI; this is the check that confirms a deck opens cleanly in **real
PowerPoint**.

## Cadence

Once per **wave** (`docs/plans/README.md §1.2`), and any time a change
touches OOXML emission in a way the automated layers can't fully judge.

## Procedure

1. Emit the reference deck:
   ```
   go run ./_gen/genrefdeck test-output/reference.pptx
   ```
   (or use the deck a wave's feature work produces — note which one below.)
2. Open it in PowerPoint (desktop).
3. Confirm:
   - [ ] No "**PowerPoint found a problem with content … Repair**" prompt.
   - [ ] The deck opens to the expected slides.
   - [ ] Shapes / text / theme colors / fonts render as intended.
4. Record the result in the log below.

If PowerPoint shows the repair prompt, the deck is invalid even if the
automated layers passed — capture the repair details, add a reproducing
fixture, and fix before the wave closes (the D-020 hygiene pass exists to
prevent exactly this).

## Results log

| Date | Wave | Deck | PowerPoint version | Repair prompt? | Notes |
|---|---|---|---|---|---|
| _(pending — maintainer)_ | Wave 1 | reference.pptx (Phase 03 A2) | — | — | First complete deck: `New()` seeds master + blank layout + theme and wires every relationship (presentation→master/slide, slide→layout, master→layout/theme); passes OPC conformance + the LibreOffice 2-page render. Awaiting the maintainer's desktop-PowerPoint open (no repair prompt expected). |
