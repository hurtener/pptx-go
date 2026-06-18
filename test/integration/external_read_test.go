// external_read_test is the Phase 19 PR#2 capstone: it proves the best-effort
// external-deck read posture (RFC §16; D-048). pptx-go does not promise
// round-trip fidelity for decks it did not author, but it must:
//
//   - never panic while opening a malformed or unfamiliar deck (criterion 4),
//   - surface every degradation via Presentation.ReadWarnings (criteria 1, 2),
//   - pass unmodeled parts through unchanged on re-save (criterion 3; D-035).
//
// The corpus is synthesised here rather than vendored as binary .pptx fixtures:
// each variant is a pptx-go-authored base deck whose ZIP is mutated to look like
// a third-party export (group shapes, mc:AlternateContent, foreign namespaces,
// missing/dangling parts, truncated XML). Generating the corpus in-test keeps it
// fully reviewable and stdlib-only, and exercises the same internal/opc +
// encoding/xml read seam a real external deck would (CLAUDE.md §17). This is a
// documented deviation from the plan's §8 "hand-authored .pptx" listing.
package integration

import (
	"archive/zip"
	"bytes"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// extBox is a reusable placement box for the external-read corpus base deck.
var extBox = pptx.Box{X: 914400, Y: 914400, W: 1828800, H: 914400}

// authoredBaseDeck builds a minimal one-slide deck and returns its bytes. Its
// slide carries exactly one navigable shape, so a corpus variant that only drops
// an unrecognized element can still assert the authored shape survives.
func authoredBaseDeck(t *testing.T) []byte {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()
	s.AddShape(pptx.ShapeRect, extBox, pptx.WithFill(pptx.SolidFill(pptx.RGB("2563EB"))))
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// unzip reads every ZIP entry into a name→bytes map.
func unzip(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	m := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		m[f.Name] = b
	}
	return m
}

// rezip writes the parts back into a ZIP in a deterministic order.
func rezip(t *testing.T, m map[string][]byte) []byte {
	t.Helper()
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatalf("zip create %s: %v", n, err)
		}
		if _, err := w.Write(m[n]); err != nil {
			t.Fatalf("zip write %s: %v", n, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

const slide1 = "ppt/slides/slide1.xml"

// injectIntoSpTree splices xmlFragment in just before the slide's </p:spTree>.
func injectIntoSpTree(t *testing.T, parts map[string][]byte, xmlFragment string) {
	t.Helper()
	xml := string(parts[slide1])
	if !strings.Contains(xml, "</p:spTree>") {
		t.Fatalf("base slide XML missing </p:spTree>:\n%s", xml)
	}
	parts[slide1] = []byte(strings.Replace(xml, "</p:spTree>", xmlFragment+"</p:spTree>", 1))
}

// wantWarning describes one ReadWarning a corpus variant must surface.
type wantWarning struct {
	kind           pptx.ReadWarningKind
	element        string // empty = don't assert the element
	part           string // empty = don't assert the part URI
	detailContains string // empty = don't assert the detail substring
}

// corpusEntry is one synthetic external-style deck plus its expected degradation.
type corpusEntry struct {
	name string
	// mutate turns a freshly-unzipped authored deck into an external-style one.
	mutate func(t *testing.T, parts map[string][]byte)
	// want is the warning the open must surface (nil = no warning is required).
	want *wantWarning
	// expectClean asserts the open surfaces no warnings at all (a well-formed
	// but unfamiliar deck, or an unrelated extra part).
	expectClean bool
	// keepsAuthoredShape is true when the authored rectangle should still read
	// back (i.e. the mutation only dropped an unknown element, not the slide).
	keepsAuthoredShape bool
}

// externalCorpus is the table of synthetic external-style decks.
func externalCorpus() []corpusEntry {
	return []corpusEntry{
		{
			name: "group_shape",
			mutate: func(t *testing.T, parts map[string][]byte) {
				injectIntoSpTree(t, parts,
					`<p:grpSp><p:nvGrpSpPr><p:cNvPr id="42" name="grp"/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:grpSp>`)
			},
			want:               &wantWarning{kind: pptx.WarnDroppedElement, element: "grpSp"},
			keepsAuthoredShape: true,
		},
		{
			name: "alternate_content",
			mutate: func(t *testing.T, parts map[string][]byte) {
				injectIntoSpTree(t, parts,
					`<mc:AlternateContent xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006"><mc:Choice Requires="a14"><p:sp><p:nvSpPr><p:cNvPr id="7" name="x"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr><p:spPr/></p:sp></mc:Choice><mc:Fallback/></mc:AlternateContent>`)
			},
			want:               &wantWarning{kind: pptx.WarnDroppedElement, element: "AlternateContent"},
			keepsAuthoredShape: true,
		},
		{
			name: "foreign_namespace_element",
			mutate: func(t *testing.T, parts map[string][]byte) {
				injectIntoSpTree(t, parts,
					`<x:custom xmlns:x="urn:example:foreign"><x:inner attr="v">payload</x:inner></x:custom>`)
			},
			want:               &wantWarning{kind: pptx.WarnDroppedElement, element: "custom"},
			keepsAuthoredShape: true,
		},
		{
			name: "empty_sptree",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// A slide whose shape tree carries only the required group props
				// and no shapes — valid, must read without panic or warning.
				parts[slide1] = []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
					`<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">` +
					`<p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld>` +
					`<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sld>`)
			},
			expectClean: true,
		},
		{
			name: "master_bgref_without_color",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// A slide master whose <p:bgRef> carries no color child — a
				// malformed background the master parser must not nil-deref on
				// (the codebase guards xmlBg.BgRef.Clr; D-048).
				const master = "ppt/slideMasters/slideMaster1.xml"
				xml := string(parts[master])
				out := strings.Replace(xml,
					`<p:bgRef idx="1001"><a:schemeClr val="bg1"/></p:bgRef>`,
					`<p:bgRef idx="1001"></p:bgRef>`, 1)
				if out == xml {
					t.Fatalf("expected to strip the bgRef color child from the master:\n%s", xml)
				}
				parts[master] = []byte(out)
			},
			expectClean:        true,
			keepsAuthoredShape: true,
		},
		{
			name: "missing_referenced_part",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// presentation.xml still references rId5 → slides/slide1.xml, but
				// the part itself is gone from the package.
				delete(parts, slide1)
				delete(parts, "ppt/slides/_rels/slide1.xml.rels")
			},
			want: &wantWarning{kind: pptx.WarnUnreadablePart, part: "/ppt/slides/slide1.xml", detailContains: "missing from the package"},
		},
		{
			name: "dangling_slide_relationship",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// Drop the slide's relationship entry while presentation.xml still
				// lists the sldId pointing at rId5 — a dangling reference.
				rels := string(parts["ppt/_rels/presentation.xml.rels"])
				out := dropRelationship(rels, "rId5")
				if out == rels {
					t.Fatalf("expected to drop rId5 from presentation rels:\n%s", rels)
				}
				parts["ppt/_rels/presentation.xml.rels"] = []byte(out)
			},
			want: &wantWarning{kind: pptx.WarnUnreadablePart, detailContains: "has no target"},
		},
		{
			name: "malformed_slide_xml",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// Truncated, unbalanced XML — the parse fails; the open degrades to
				// a warn-and-skip rather than a hard error or panic.
				parts[slide1] = []byte(`<?xml version="1.0"?><p:sld><p:cSld><p:spTree><p:sp`)
			},
			want: &wantWarning{kind: pptx.WarnUnreadablePart, part: "/ppt/slides/slide1.xml", detailContains: "could not be parsed"},
		},
		{
			name: "text_body_field",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// A slide-number field <a:fld> inside a shape's text body — common in
				// real decks, unmodeled by pptx-go, dropped at parse. It must surface
				// as a nested dropped-element warning, not vanish silently.
				injectIntoSpTree(t, parts,
					`<p:sp><p:nvSpPr><p:cNvPr id="9" name="tb"/><p:cNvSpPr/><p:nvPr/></p:nvSpPr><p:spPr/>`+
						`<p:txBody><a:bodyPr/><a:p><a:fld id="{F}" type="slidenum"><a:t>3</a:t></a:fld></a:p></p:txBody></p:sp>`)
			},
			want: &wantWarning{kind: pptx.WarnDroppedElement, element: "fld", part: "/ppt/slides/slide1.xml", detailContains: "text-body element"},
		},
		{
			name: "malformed_theme",
			mutate: func(t *testing.T, parts map[string][]byte) {
				// A theme part that exists but cannot be parsed: the deck keeps the
				// default theme and warns, rather than failing the open.
				parts["ppt/theme/theme1.xml"] = []byte(`<a:theme><a:themeElements`)
			},
			want:               &wantWarning{kind: pptx.WarnUnreadablePart, part: "/ppt/theme/theme1.xml", detailContains: "theme could not be parsed"},
			keepsAuthoredShape: true,
		},
		{
			name: "stray_custom_part",
			mutate: func(t *testing.T, parts map[string][]byte) {
				parts["customXml/item1.xml"] = []byte(`<?xml version="1.0"?><root>external-unmodeled</root>`)
				ct := string(parts["[Content_Types].xml"])
				parts["[Content_Types].xml"] = []byte(strings.Replace(ct, "</Types>",
					`<Override PartName="/customXml/item1.xml" ContentType="application/xml"/></Types>`, 1))
			},
			expectClean:        true,
			keepsAuthoredShape: true, // unrelated part; slide is untouched
		},
	}
}

// dropRelationship removes the single <Relationship .../> element whose Id is id
// from a .rels document, leaving the rest byte-identical.
func dropRelationship(rels, id string) string {
	marker := `Id="` + id + `"`
	i := strings.Index(rels, marker)
	if i < 0 {
		return rels
	}
	start := strings.LastIndex(rels[:i], "<Relationship")
	if start < 0 {
		return rels
	}
	// Relationships pptx-go emits use an explicit close tag.
	close := strings.Index(rels[i:], "</Relationship>")
	if close < 0 {
		// Fall back to a self-closing element.
		close = strings.Index(rels[i:], "/>")
		if close < 0 {
			return rels
		}
		return rels[:start] + rels[i+close+len("/>"):]
	}
	return rels[:start] + rels[i+close+len("</Relationship>"):]
}

// mustNotPanic runs fn, failing the test (without aborting the table) if it
// panics. The whole point of best-effort external read is that no malformed
// input reaches a panic (D-048; criterion 4).
func mustNotPanic(t *testing.T, name string, fn func()) (panicked bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("[%s] open panicked: %v", name, r)
			panicked = true
		}
	}()
	fn()
	return false
}

// TestExternalRead_CorpusNoPanic is acceptance criterion 4 (and 1/2 per entry):
// every synthetic external-style deck opens without panic, returns a usable
// presentation, and surfaces the expected ReadWarning. Run under -race in CI.
func TestExternalRead_CorpusNoPanic(t *testing.T) {
	for _, e := range externalCorpus() {
		t.Run(e.name, func(t *testing.T) {
			parts := unzip(t, authoredBaseDeck(t))
			e.mutate(t, parts)
			data := rezip(t, parts)

			var re *pptx.Presentation
			var err error
			if mustNotPanic(t, e.name, func() { re, err = pptx.NewFromBytes(data) }) {
				return
			}
			// Best-effort: a third-party deck opens rather than erroring out.
			if err != nil {
				t.Fatalf("[%s] NewFromBytes returned a fatal error (want best-effort open): %v", e.name, err)
			}
			if re == nil {
				t.Fatalf("[%s] NewFromBytes returned nil presentation", e.name)
			}

			ws := re.ReadWarnings()
			if e.want != nil && !hasWarning(ws, *e.want) {
				t.Errorf("[%s] expected warning %+v not surfaced; got %+v", e.name, *e.want, ws)
			}
			if e.expectClean && hasAnyWarning(ws) {
				t.Errorf("[%s] expected a clean open but got warnings: %+v", e.name, ws)
			}
			if e.keepsAuthoredShape {
				if n := len(re.Slides()); n != 1 {
					t.Fatalf("[%s] reopened deck has %d slides, want 1", e.name, n)
				}
				if n := len(re.Slides()[0].Shapes()); n != 1 {
					t.Errorf("[%s] reopened slide has %d navigable shapes, want 1 (authored rect survives drop)", e.name, n)
				}
			}
		})
	}
}

// hasWarning reports whether ws contains a warning matching want. Each of
// element / part / detailContains is asserted only when non-empty, so a case can
// pin down as much of the warning as it cares about.
func hasWarning(ws []pptx.ReadWarning, want wantWarning) bool {
	for _, w := range ws {
		if w.Kind != want.kind {
			continue
		}
		if want.element != "" && w.Element != want.element {
			continue
		}
		if want.part != "" && w.Part != want.part {
			continue
		}
		if want.detailContains != "" && !strings.Contains(w.Detail, want.detailContains) {
			continue
		}
		return true
	}
	return false
}

func hasAnyWarning(ws []pptx.ReadWarning) bool { return len(ws) > 0 }

// TestExternalRead_PartPassThrough is acceptance criterion 3 at the integration
// seam: an unmodeled part survives NewFromBytes → WriteToBytes byte-for-byte,
// and its content-type registration is preserved (a re-saved deck with a part
// but no declared content type would be invalid OPC). D-035; D-048.
func TestExternalRead_PartPassThrough(t *testing.T) {
	parts := unzip(t, authoredBaseDeck(t))
	const custom = "customXml/item1.xml"
	payload := []byte(`<?xml version="1.0" encoding="UTF-8"?><root>external-unmodeled-payload</root>`)
	parts[custom] = payload
	ct := string(parts["[Content_Types].xml"])
	parts["[Content_Types].xml"] = []byte(strings.Replace(ct, "</Types>",
		`<Override PartName="/customXml/item1.xml" ContentType="application/xml"/></Types>`, 1))

	re, err := pptx.NewFromBytes(rezip(t, parts))
	if err != nil {
		t.Fatalf("NewFromBytes with unmodeled part: %v", err)
	}
	if w := re.ReadWarnings(); hasAnyWarning(w) {
		t.Errorf("unmodeled part should not raise a read warning; got %+v", w)
	}
	out, err := re.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	outParts := unzip(t, out)
	if got := outParts[custom]; !bytes.Equal(got, payload) {
		t.Errorf("unmodeled part not preserved byte-for-byte:\n got %q\nwant %q", got, payload)
	}
	if got := string(outParts["[Content_Types].xml"]); !strings.Contains(got, "/customXml/item1.xml") {
		t.Errorf("unmodeled part's content-type override dropped on re-save:\n%s", got)
	}
}
