# Repair-prompt hygiene — trigger list

PowerPoint shows a *"PowerPoint found a problem with content … Repair"* prompt
on certain OOXML quirks that are mechanically harmless but alarm recipients.
Emitting XML that opens cleanly is **correctness, not preference** (D-020), so
pptx-go runs an always-on hygiene pass over every emitted part.

- **Where:** `internal/render/hygiene.go` (`render.Sanitize`).
- **When:** unconditionally, on every write path (`Save`, `Write`,
  `WriteToBytes`, and the streaming save), just before serialization.
- **No switch.** There is no caller-facing option to disable it
  (`RepairPromptHygiene` is correctness — D-020). The alternative is "valid
  OOXML that nonetheless looks broken".

The pass is **conservative**: each rule targets one documented trigger and
leaves everything else byte-for-byte intact, so XML carrying no trigger is
returned unchanged. It is idempotent.

## Triggers

| ID | Trigger | Action | Why it triggers repair |
|----|---------|--------|------------------------|
| **H1** | A leading UTF-8 BOM (`EF BB BF`) before `<?xml …?>` | Strip the BOM | A byte-order mark ahead of the XML declaration makes PowerPoint treat the part as malformed. OPC parts are UTF-8 and never need one. |
| **H2** | An empty DrawingML language attribute — `lang=""` (either quote style) | Remove the attribute | An empty `lang` on `<a:rPr>`/`<a:defRPr>`/`<a:endParaRPr>` is rejected; removing it lets the run inherit the document language. (The retired hand-rolled writer once emitted a stray `lang`; this is the safety net.) |

## Adding a trigger

A new repair trigger discovered in the wild is fixed in a **single PR**: add a
rule to `render.Sanitize`, a row to this table, and a golden test asserting the
rule fires on the trigger and touches nothing else (D-020 — never a silent
post-processor change). Keep rules byte-precise (e.g. include the leading space
of an attribute match) so well-formed XML is never corrupted.
