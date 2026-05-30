# Research briefs — index

> Subsystem → research-brief reverse index. A phase plan that cites no
> brief in its **Brief findings incorporated** section is a drift signal
> (`CLAUDE.md §16`).
>
> A research brief is a `docs/research/NN-slug.md` file authored *before*
> a phase plan to investigate prior art, capture domain knowledge, and
> tee up the design decisions that will land in the RFC or the phase
> plan. Briefs are **context, not decisions** — they inform; the RFC and
> phase plans decide.

---

## Format

Each entry below is keyed by **subsystem** (matching `RFC §3.3`) and
lists the briefs that inform work in that subsystem.

```text
### <subsystem>
- `NN-slug.md` — one-line summary
- `NN-slug.md` — one-line summary
```

A brief may cross-cut multiple subsystems; list it under each.

---

## Subsystems

### internal/opc — OPC package layer

*(no briefs yet — Phase 01 may add one investigating OOXML transitional
vs strict profile edge cases at the OPC layer)*

### internal/ooxml — OOXML codec layer

- `01-master-layout-theme-ingestion.md` — theme1.xml color/font scheme +
  master/layout inheritance, and the read paths template ingestion depends on.

*(candidates: chart wire-format survey for V2, table XML shape, theme XML
compatibility across PowerPoint versions)*

### pptx — Layer 1 builder

- `01-master-layout-theme-ingestion.md` — `LoadTheme`/`FromTemplate` strategy:
  copy template parts wholesale, extract the `Theme`, map `LayoutKind` to
  named layouts.

*(candidates: rich-text auto-fit modes in OOXML practice, table merged-cell
semantics)*

### scene — Layer 2 renderer

*(no briefs yet — candidates: layout-engine survey (CSS grid analogues
expressible in EMU), text-overflow heuristics, scene IR JSON wire form
compatibility with pengui-slides v4)*

### Theme & tokens

- `01-master-layout-theme-ingestion.md` — how a brand kit's color scheme,
  `clrMap` indirection, and font scheme map onto pptx-go's token roles.

*(candidates: token taxonomy comparison with design systems (Tailwind, Radix,
Material))*

### Curated assets (icons, ornaments, frames)

*(no briefs yet — candidates: lucide-to-OOXML path translator
constraints, preset ornament shape recipes survey, device-frame
shape geometry)*

### Charts

*(no briefs yet — V2 will warrant briefs on `c:chart` XML survey by
chart type and PowerPoint Online vs Desktop divergences)*

### Streaming & performance

*(no briefs yet — candidates: concurrent rendering scaling on M-class
Apple Silicon vs x86_64, zip-streaming costs vs in-memory)*

### Read & round-trip

*(no briefs yet — candidates: PowerPoint output variance (PowerPoint vs
PowerPoint Online vs Office for Mac), Keynote-to-PPTX export quirks)*

---

## Authoring a brief

A research brief is a Markdown file under `docs/research/` named
`NN-slug.md` where `NN` is the next available two-digit number. Brief
structure:

```markdown
# Brief NN — <slug>

**Subsystem:** <RFC §3.3 subsystem>
**Authored:** <YYYY-MM-DD>
**Motivating phase:** <Phase NN — slug> (or "RFC-level investigation")

## 1. Question
What this brief investigates.

## 2. Prior art surveyed
Specs, libraries, papers, decks consulted.

## 3. Findings
What we learned. Bullet-point. Each finding is something a phase plan
can incorporate or reject.

## 4. Recommendations
Suggested directions for the motivating phase. Recommendations are
inputs to the phase plan; the plan decides.

## 5. Open questions
What we *didn't* answer, with a note about which phase / RFC change
should pick it up.
```

Briefs are authored before the phase plan, listed in this INDEX under
the relevant subsystem, and cited by the phase plan's **Brief findings
incorporated** section.

A brief is not a phase plan. A brief makes recommendations; the phase
plan binds. When a phase plan departs from a brief's recommendation, the
**Findings I'm departing from** section names the brief and the
rationale.

---

*Add new briefs to the subsystem section above and to a chronological
list below (V1.x — for now the index above is the canonical view).*
