# Vendored OOXML / OPC specifications

Spec snapshots pptx-go implements against, pinned by edition + date
(`CLAUDE.md §10`). A codec change motivated by a spec re-read updates the
relevant snapshot and the codec goldens in the same PR.

## ISO/IEC 29500 transitional XSDs (schema-validity layer — D-031)

The schema-conformance layer (`scripts/validate-schema.sh`, run in CI)
validates emitted part XML against the **transitional** profile schemas with
`xmllint`. The schemas are **vendored** under
`docs/specifications/ooxml-transitional/` (the layer validates, no longer SKIPs).

Pinned edition:

| Schema set | Edition | Date pinned |
|---|---|---|
| ISO/IEC 29500-4 transitional | ISO/IEC 29500-4:2016 | 2026-05-30 |

The validator maps part types to top-level schemas:
- `pml.xsd` — PresentationML: `presentation.xml`, slides, masters, layouts,
  notes masters, notes slides.
- `dml-main.xsd` — DrawingML: `theme*.xml`.
- the `shared-*.xsd` / `dml-*.xsd` / `vml-*.xsd` / `sml.xsd` / `wml.xsd` /
  `xml.xsd` files are the imports those reference (kept so imports resolve).

`scripts/validate-schema.sh` (no argument) emits the full-surface showcase deck
(`_gen/genshowcase`) and validates every part — this catches namespace/ordering/
element bugs the structural OPC gate (`internal/conformance`, layer 1) cannot.
Two real PowerPoint "repair" bugs were found this way: a table cell emitting
`<p:txBody>` (must be `<a:txBody>`) and a `<p:sldMasterId>` inside the notes-
master list (must be `<p:notesMasterId>`).

To re-pin after a spec re-read: replace the `.xsd` files, update the table
above, and re-run the validator (D-017).

Schema validation is a layer, not the whole story: PowerPoint is both
stricter and looser than the schema in places. Known divergences are
annotated rather than chased to 100% (D-031). The office-app open proxy
(LibreOffice headless) and the manual PowerPoint check cover what the schema
cannot.
