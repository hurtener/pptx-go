# Design decisions

pptx-go records its settled design decisions in an append-only log, each
with an immutable ID (`D-NNN`). This page is a curated index of the
decisions that affect how you *use* the library — each with a one-line
note on what it means for you. It is not exhaustive; the
[full decisions log on GitHub](https://github.com/hurtener/pptx-go/blob/main/docs/decisions.md)
is the authoritative text.

## Binding properties (P1–P4)

The four properties every part of the library upholds.

- **P1 — Two layers, one library.** `pptx` (builder) and `scene`
  (renderer) are the only public layers; `scene` composes `pptx`.
  *For you:* pick the imperative builder or the typed scene renderer — and
  anything `scene` does, you can also do directly with `pptx`.
- **P2 — Tokens, not literals.** Visual properties flow through a `Theme`;
  literals are an escape hatch.
  *For you:* author with theme tokens and a theme swap re-renders the same
  input in a new visual language; reach for `RGB`/`Pt` only when you must.
- **P3 — OOXML by isolation.** Raw OOXML wire types stay inside the
  library.
  *For you:* no public API ever hands you an XML struct — you work in
  typed Go domain objects.
- **P4 — No CGo, stdlib-only runtime.** The shipped artifact is pure Go
  with no third-party runtime dependencies.
  *For you:* it cross-compiles anywhere Go does and adds nothing to your
  dependency tree.

## Rendering and content model

- **D-004 — Charts are images in V1.** A chart node renders as an image
  shape, not a native `c:chart`.
  *For you:* supply chart visuals as rendered raster bytes (via an asset);
  native editable charts are a V2 item.
- **D-011 / D-018 — Per-node rendering policy, not per-deck.** Whether a
  node renders as native shapes or as a picture is intrinsic to the node
  type (whether its IR carries an asset field); there is no `Disposition`
  enum or deck-wide mode toggle.
  *For you:* you don't pick a render mode — each node already knows how it
  renders.
- **D-014 — Code blocks are rasters.** A `code_block` renders as a
  caller-supplied raster image in V1.
  *For you:* render syntax-highlighted code to an image yourself and hand
  it in as an asset.
- **D-022 — Speaker notes are V1.** Per-slide speaker notes are a shipped
  builder capability.
  *For you:* set notes with `Slide.SetSpeakerNotes` / a scene slide's
  `Notes` and read them back with `Slide.SpeakerNotes`.
- **D-026 — Engine, not product.** pptx-go converts a typed scene into
  PPTX and nothing else; product behavior (render modes, legibility
  heuristics, markdown ingestion) lives in callers.
  *For you:* the library gives you mechanisms (asset resolution, font
  embedding, theme tokens, slide grouping) — opinions about *what* the deck
  should contain are yours.

## Theme and color

- **D-033 — Apply-time token resolution; theme1.xml emission deferred.**
  `Color` is a sealed interface; the `RGB` type is the literal and tokens
  resolve against the active theme at apply time. Token-driven
  `theme1.xml` emission is a follow-up (resolution to `srgbClr` covers
  V1).
  *For you:* tokens are honored when applied, so a theme swap re-colors the
  same input; a reopened deck surfaces resolved literal colors (the slide
  carries no token to reconstruct).
- **D-037 — Brand-kit ingestion.** `FromTemplate(brand)` clones a brand
  deck's OPC package and strips its slides, preserving theme, masters,
  layouts, and relationships.
  *For you:* start from a corporate template and your deck inherits its
  full look without grafting parts by hand.

## Assets

- **D-024 / D-036 — Asset resolution; warn, don't fail.** Asset IDs are
  free-form (with an `asset://`-URI helper), and every unresolved asset
  degrades to a `LayoutWarning` with the node skipped — there is no
  render-fatal asset path or strict mode in V1.
  *For you:* provide an `AssetResolver`; if you need a missing asset to be
  fatal, inspect `Stats.Warnings` yourself.

## Reading decks

- **D-047 / D-048 — Read model; best-effort external read.** Reopening a
  pptx-go-authored deck reconstructs the same navigable `Shape` model
  (round-trip); reading a third-party deck is best-effort graceful
  degradation — never panic, surface what was dropped, pass unmodeled parts
  through unchanged on re-save.
  *For you:* a deck you wrote reopens losslessly into navigable shapes;
  for foreign decks, check `Presentation.ReadWarnings()` to see what
  degraded (unrecognized *shapes* are lost on re-save; unrecognized *parts*
  are preserved).
- **D-049 — Read security bounds.** The read path enforces a
  caller-configurable per-part memory ceiling (default 100 MB,
  `ErrPartTooLarge`) and rejects zip-slip paths; read constructors accept
  options and log degradation.
  *For you:* opening an untrusted deck is memory-bounded and path-safe by
  default; tune the limit with `WithReadPartLimit` and inject a logger with
  `WithLogger`.
- **D-050 — Notes round-trip.** Speaker notes are reconstructed on open, so
  inspecting and re-saving a deck no longer drops them.
  *For you:* you can open a deck, read its notes, and save without silent
  data loss.

## The full log

The decisions above are the user-relevant subset. For the complete,
authoritative record — including contributor-facing decisions and the full
context/consequences of each entry — see
[`docs/decisions.md` on GitHub](https://github.com/hurtener/pptx-go/blob/main/docs/decisions.md).
