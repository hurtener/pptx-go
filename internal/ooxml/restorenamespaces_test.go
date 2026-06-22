package ooxml

import (
	"strings"
	"testing"
)

func restore(t *testing.T, bare string) string {
	t.Helper()
	out, err := RestoreNamespaces([]byte(bare))
	if err != nil {
		t.Fatalf("RestoreNamespaces(%q): %v", bare, err)
	}
	return string(out)
}

func TestRestorePrefixesElements(t *testing.T) {
	bare := `<sld><clrMapOvr/><cSld><spTree><sp><spPr><xfrm><off x="1" y="2"/><ext cx="3" cy="4"/></xfrm><solidFill><srgbClr val="FF0000"/></solidFill></spPr></sp></spTree></cSld></sld>`
	got := restore(t, bare)

	for _, want := range []string{
		`<p:sld `, `<p:clrMapOvr/>`, `<p:cSld>`, `<p:spTree>`, `<p:sp>`, `<p:spPr>`,
		`<a:xfrm>`, `<a:off x="1" y="2"/>`, `<a:ext cx="3" cy="4"/>`,
		`<a:solidFill>`, `<a:srgbClr val="FF0000"/>`,
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"`,
		`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	// xmlns:r must NOT be declared — no relationship attr in this fragment.
	if strings.Contains(got, "xmlns:r=") {
		t.Errorf("declared xmlns:r though unused:\n%s", got)
	}
}

// TestRestoreEmbeddedFontList proves the <p:embeddedFontLst> face elements —
// including the <p:font> typeface child — are all p:-prefixed (a bare <font> is
// invalid OOXML and PowerPoint cannot bind the embedded face). R9.1/R9.7 fix.
func TestRestoreEmbeddedFontList(t *testing.T) {
	bare := `<embeddedFontLst><embeddedFont><font typeface="Cardo"/><regular rid="rId6"/><bold rid="rId7"/><italic rid="rId8"/><boldItalic rid="rId9"/></embeddedFont></embeddedFontLst>`
	got := restore(t, bare)
	for _, want := range []string{
		`<p:embeddedFontLst `, `<p:embeddedFont>`, `<p:font typeface="Cardo"/>`,
		`<p:regular r:id="rId6"/>`, `<p:bold r:id="rId7"/>`,
		`<p:italic r:id="rId8"/>`, `<p:boldItalic r:id="rId9"/>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, `<font `) {
		t.Errorf("bare <font> survived (must be p:font):\n%s", got)
	}
}

// TestRestoreLineSpacing proves the paragraph line-spacing elements (D-061) are
// a:-prefixed — a bare <lnSpc>/<spcPct> is invalid OOXML.
func TestRestoreLineSpacing(t *testing.T) {
	bare := `<pPr><lnSpc><spcPct val="102000"/></lnSpc></pPr>`
	got := restore(t, bare)
	for _, want := range []string{`<a:pPr `, `<a:lnSpc>`, `<a:spcPct val="102000"/>`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRestoreRelationshipAttr(t *testing.T) {
	bare := `<presentation><sldIdLst><sldId id="256" rid="rId1"></sldId></sldIdLst></presentation>`
	got := restore(t, bare)
	for _, want := range []string{
		`<p:presentation `, `<p:sldIdLst>`, `<p:sldId id="256" r:id="rId1"`,
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, ` rid=`) {
		t.Errorf("bare rid= survived (should be r:id):\n%s", got)
	}
}

// TestRestoreIsStripInverse proves Restore is the faithful inverse of
// StripNamespacePrefixes: stripping a restored document yields the bare input
// (modulo self-closing normalization, which we neutralize by stripping both).
func TestRestoreIsStripInverse(t *testing.T) {
	bare := `<presentation><sldIdLst><sldId id="256" rid="rId1"/></sldIdLst></presentation>`
	restored := restore(t, bare)
	stripped, err := StripNamespacePrefixes([]byte(restored))
	if err != nil {
		t.Fatal(err)
	}
	// Normalize the original bare through Strip too (self-close/format parity).
	wantNorm, err := StripNamespacePrefixes([]byte(bare))
	if err != nil {
		t.Fatal(err)
	}
	if string(stripped) != string(wantNorm) {
		t.Errorf("strip(restore(bare)) != strip(bare)\n got: %s\nwant: %s", stripped, wantNorm)
	}
}

func TestRestoreDropsMarshaledXmlns(t *testing.T) {
	// A marshaler may emit xmlns attrs on the root; this pass owns them and
	// must not duplicate.
	bare := `<sld xmlns:a="x" xmlns:p="y" xmlns:r="z"><cSld/></sld>`
	got := restore(t, bare)
	if strings.Count(got, "xmlns:p=") != 1 {
		t.Errorf("xmlns:p not declared exactly once:\n%s", got)
	}
	if strings.Contains(got, `="x"`) || strings.Contains(got, `="y"`) {
		t.Errorf("stale marshaled xmlns survived:\n%s", got)
	}
}

func TestRestoreUnknownElementStaysBare(t *testing.T) {
	got := restore(t, `<sld><madeUpElement/></sld>`)
	if !strings.Contains(got, "<madeUpElement/>") {
		t.Errorf("unknown element should stay bare:\n%s", got)
	}
}
