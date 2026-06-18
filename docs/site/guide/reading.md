# Reading decks

pptx-go reads a `.pptx` file back into the same navigable model the builder
authors. A deck pptx-go wrote round-trips losslessly; a third-party deck opens
best-effort, with every degradation reported so you know what was dropped.

## Opening a deck

| Constructor | Source |
| --- | --- |
| `NewFromBytes(data []byte, opts ...Option) (*Presentation, error)` | An in-memory `.pptx`. |
| `NewFromFile(path string, opts ...Option) (*Presentation, error)` | A file path (eager load). |
| `OpenStream(path string, opts ...Option) (*Presentation, error)` | A file path (lazy per-part streaming load). |

The read constructors accept the same `Option`s as `pptx.New`; two matter on the
read path:

- `WithLogger(*slog.Logger)` — emits a `Warn` event per non-fatal degradation.
- `WithReadPartLimit(int64)` — the per-part decompressed size ceiling (see
  [security bounds](#security-bounds)).

```go
re, err := pptx.NewFromFile("deck.pptx",
	pptx.WithLogger(logger),
	pptx.WithReadPartLimit(50<<20), // 50 MB per part
)
if err != nil {
	log.Fatal(err)
}
```

## Navigating the model

`Slides()` returns the slides; `Slide.Shapes()` returns each shape-tree child as
a `*Shape`. A shape exposes typed read accessors that mirror the authoring API:

```go
for _, sl := range re.Slides() {
	for _, sh := range sl.Shapes() {
		_ = sh.Geometry() // ShapeGeometry preset name ("" for a picture/table/icon)
		_ = sh.Rotation() // clockwise degrees in [0, 360°)
		_ = sh.Fill()     // Fill, or nil if it inherits its style fill
		_ = sh.Line()     // Line outline
		if shadow, ok := sh.Shadow(); ok {
			_ = shadow // an Elevation
		}

		if tf, ok := sh.TextFrame(); ok {
			for _, para := range tf.Paragraphs() {
				for _, run := range para.Runs() {
					_ = run.Text()
				}
			}
		}
		if tbl, ok := sh.Table(); ok {
			_ = tbl.RowCount()
			_ = tbl.Cell(0, 0)
		}
		if img, ok := sh.Image(); ok {
			b, err := img.Bytes() // the embedded bytes, resolved via the relationship
			_, _ = b, err
		}
	}
}
```

The accessors are the read inverse of the authoring API:
`Shape.Geometry/Rotation/Fill/Line/Shadow/Box`, `Shape.TextFrame`,
`Shape.Table`, and `Shape.Image`. Reopened colors surface as resolved literals
(theme tokens are baked to sRGB at write time).

## Round-trip fidelity

Every shape, text run, fill, line, table, and image pptx-go emits reopens into
the same model (G6) — a self-authored deck round-trips losslessly and reports no
read warnings. Speaker notes round-trip too: notes authored with
`SpeakerNotes`/`SetSpeakerNotes` are reconstructed on open, so a read-then-save
cycle preserves them rather than silently dropping them (D-050).

## Best-effort external read

Opening a deck pptx-go did **not** author is best-effort: unrecognized content is
ignored at parse time rather than preserved, but every degradation is reported
via `ReadWarnings()` (D-048). A self-authored deck returns `nil`.

```go
for _, w := range re.ReadWarnings() {
	log.Printf("[%s] %s %s — %s", w.Kind, w.Part, w.Element, w.Detail)
}
```

A `ReadWarning` carries:

```go
type ReadWarning struct {
	Kind    ReadWarningKind // WarnDroppedElement or WarnUnreadablePart
	Part    string          // the part URI, e.g. "/ppt/slides/slide2.xml"
	Element string          // element local-name (for WarnDroppedElement)
	Detail  string          // human-readable context
}
```

- **`WarnDroppedElement`** — an unrecognized element was ignored at parse time
  (e.g. a group shape, `mc:AlternateContent`, or an `<a:fld>` field in a text
  body). The element local-name is in `Element`.
- **`WarnUnreadablePart`** — a referenced part was missing or could not be
  parsed, and was skipped rather than failing the open (a dangling slide
  reference, a malformed slide, an unparseable theme).

Warnings are de-duplicated per `(Kind, Part, Element)` and returned in a stable
order. When a logger is injected, each distinct degradation is also logged as a
`Warn` event.

## Security bounds {#security-bounds}

The read path enforces the §7 memory and path bounds (D-049):

- **Per-part size limit.** A part whose decompressed size exceeds the limit is
  rejected with an error rather than allocated, bounding memory on untrusted
  input. The default is 100 MB; override it with `WithReadPartLimit(n)`, where
  `n <= 0` disables the bound.
- **Zip-slip rejection.** Every part path is validated to stay within the package
  root; an absolute path or one containing a `..` segment is rejected at parse
  time.

These bounds apply to all three read constructors.
