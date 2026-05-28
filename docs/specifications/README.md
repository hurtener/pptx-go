# Vendored OOXML / OPC specifications

Spec snapshots pptx-go implements against, pinned by edition + date
(`CLAUDE.md §10`). A codec change motivated by a spec re-read updates the
relevant snapshot and the codec goldens in the same PR.

## ISO/IEC 29500 transitional XSDs (schema-validity layer — D-031)

The schema-conformance layer (`scripts/validate-schema.sh`, run in CI)
validates emitted part XML against the **transitional** profile schemas with
`xmllint`. The schemas are **not yet vendored** — until they are, the layer
SKIPs.

To vendor them:

1. Obtain the ECMA-376 (1st edition / ISO 29500 transitional) XSD bundle from
   the ECMA-376 download (Part 4 / the `OfficeOpenXML-XMLSchema-Transitional`
   set).
2. Place the schema files under `docs/specifications/ooxml-transitional/`,
   named by convention so the validator can find them:
   - `pml.xsd` — PresentationML (presentation.xml, slides, masters, layouts)
   - `dml.xsd` — DrawingML (theme, shapes)
   - `sml.xsd`, `wml.xsd` — Spreadsheet/Word (not used by pptx-go)
   - plus the shared imports the above reference (`shared-*.xsd`, etc.).
3. Pin the edition + date here:

   | Schema set | Edition | Date pinned |
   |---|---|---|
   | ISO/IEC 29500 transitional | _(fill in)_ | _(fill in)_ |

4. Run `scripts/validate-schema.sh` — it will switch from SKIP to validating.

Schema validation is a layer, not the whole story: PowerPoint is both
stricter and looser than the schema in places. Known divergences are
annotated rather than chased to 100% (D-031). The office-app open proxy
(LibreOffice headless) and the manual PowerPoint check cover what the schema
cannot.
