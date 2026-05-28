// Package drawing is a placeholder for DrawingML wire types (shapes, fills,
// geometries, text bodies) per RFC §6.2.
//
// Phase 01 keeps the DrawingML types in internal/ooxml/slide because they
// are used only by the slide family today and share the slide package's
// XMLWriter serialization base (D-028). They migrate here — with the
// serialization base moving to a shared helper — when the builder
// (Phase 03+) or the SVG→OOXML translator (Phase 12) first needs them
// outside the slide family.
package drawing
