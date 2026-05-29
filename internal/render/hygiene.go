// Package render holds builder-internal rendering helpers that sit below the
// public pptx API but above the OOXML wire types.
//
// hygiene.go is the always-on repair-prompt hygiene pass (D-020): PowerPoint
// shows a "this file has been repaired" prompt on certain OOXML quirks that are
// mechanically harmless but spook recipients. Emitting XML that opens cleanly
// is correctness, not preference, so this pass runs unconditionally on every
// emitted part — there is no caller-facing switch to disable it. The trigger
// list is documented in docs/design/HYGIENE.md and grows as new triggers
// surface in the wild (each addition is a documented fix + a list entry in one
// PR, never a silent change).
package render

import "bytes"

// utf8BOM is the UTF-8 byte-order mark. A BOM before the XML declaration is a
// repair trigger; OPC parts are UTF-8 and never need one.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Sanitize applies the repair-prompt hygiene pass to one emitted XML part and
// returns the cleaned bytes. It is conservative by design: each rule targets a
// single documented trigger and leaves everything else byte-for-byte intact, so
// XML that carries no trigger is returned unchanged. Idempotent.
//
// Triggers handled (docs/design/HYGIENE.md):
//
//   - H1  a leading UTF-8 BOM before the XML declaration
//   - H2  empty xml:lang / lang attributes (lang="") — PowerPoint rejects them
func Sanitize(xml []byte) []byte {
	out := xml

	// H1 — strip a leading UTF-8 BOM.
	out = bytes.TrimPrefix(out, utf8BOM)

	// H2 — drop empty DrawingML lang attributes (`<a:rPr lang="">`). An empty
	// lang is a known repair trigger; removing it lets the run inherit the
	// document language.
	if bytes.Contains(out, langEmptyDQ) || bytes.Contains(out, langEmptySQ) {
		out = bytes.ReplaceAll(out, langEmptyDQ, nil)
		out = bytes.ReplaceAll(out, langEmptySQ, nil)
	}

	return out
}

// Empty-lang trigger patterns (both quote styles). The leading space is part of
// the match so the surrounding attribute spacing stays well-formed after the
// attribute is removed.
var (
	langEmptyDQ = []byte(` lang=""`)
	langEmptySQ = []byte(` lang=''`)
)
