package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// RestoreNamespaces is the write-side inverse of StripNamespacePrefixes
// (D-032). Codecs marshal their structs with bare element/attribute names via
// encoding/xml (which serializes attributes correctly, unlike the retired
// hand-rolled writer); this pass re-applies the canonical OOXML prefixes
// (p:/a:/r:) per element and declares the used namespaces on the root, so the
// emitted XML is what PowerPoint expects.
//
// It is self-configuring: it prefixes each element via elementPrefix, maps the
// relationship attributes (rid→r:id, rembed→r:embed, rlink→r:link) back, drops
// any xmlns attributes the marshaler emitted (this pass owns them), and
// declares xmlns:<prefix> on the root for exactly the prefixes used.
func RestoreNamespaces(data []byte) ([]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var toks []xml.Token
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("restore namespaces: %w", err)
		}
		toks = append(toks, xml.CopyToken(t))
	}

	// First pass: which prefixes does the document use?
	used := map[string]bool{}
	for _, t := range toks {
		if se, ok := t.(xml.StartElement); ok {
			if p := elementPrefix(se.Name.Local); p != "" {
				used[p] = true
			}
			for _, a := range se.Attr {
				if p := attrPrefix(a.Name.Local); p != "" {
					used[p] = true
				}
			}
		}
	}

	var b bytes.Buffer
	depth := 0
	var parents []string // open element local names, for context-sensitive prefixing
	for i := 0; i < len(toks); i++ {
		switch t := toks[i].(type) {
		case xml.StartElement:
			name := prefixed(prefixFor(t.Name.Local, parents), elementLocal(t.Name.Local))
			b.WriteByte('<')
			b.WriteString(name)
			if depth == 0 {
				// Declare exactly the namespaces used, deterministically.
				for _, p := range sortedPrefixes(used) {
					fmt.Fprintf(&b, ` xmlns:%s="%s"`, p, nsURI[p])
				}
			}
			for _, a := range t.Attr {
				if strings.EqualFold(a.Name.Local, "xmlns") || strings.EqualFold(a.Name.Space, "xmlns") {
					continue // this pass owns namespace declarations
				}
				an := prefixed(attrPrefix(a.Name.Local), attrLocal(a.Name.Local))
				fmt.Fprintf(&b, ` %s="%s"`, an, escapeAttr(a.Value))
			}
			// Self-close when the next token closes this element with no content
			// (no child → nothing pushed onto the parent stack).
			if i+1 < len(toks) {
				if ee, ok := toks[i+1].(xml.EndElement); ok && ee.Name.Local == t.Name.Local {
					b.WriteString("/>")
					i++ // consume the matching EndElement
					continue
				}
			}
			b.WriteByte('>')
			parents = append(parents, t.Name.Local)
			depth++
		case xml.EndElement:
			if len(parents) > 0 {
				parents = parents[:len(parents)-1] // pop before prefixing so the parent is the new top
			}
			b.WriteString("</")
			b.WriteString(prefixed(prefixFor(t.Name.Local, parents), elementLocal(t.Name.Local)))
			b.WriteByte('>')
			depth--
		case xml.CharData:
			_ = xml.EscapeText(&b, t)
		case xml.Comment:
			b.WriteString("<!--")
			b.Write(t)
			b.WriteString("-->")
		}
	}
	return b.Bytes(), nil
}

func prefixed(prefix, local string) string {
	if prefix == "" {
		return local
	}
	return prefix + ":" + local
}

// prefixFor returns an element's prefix, accounting for the few elements whose
// namespace depends on context. "txBody" is <a:txBody> inside a table cell
// (<a:tc>) but <p:txBody> inside a shape — the same local name, different
// namespace by parent (the context-free table would otherwise emit p:txBody in
// a cell, which PowerPoint rejects).
func prefixFor(local string, parents []string) string {
	if local == "txBody" && len(parents) > 0 && parents[len(parents)-1] == "tc" {
		return "a"
	}
	return elementPrefix(local)
}

// elementLocal rewrites a sentinel element local name to its real output name
// (the element analogue of attrLocal). "pxfrm" lets the graphic frame's
// transform emit as <p:xfrm> while a shape's transform stays <a:xfrm> — the two
// share the local name "xfrm" but need different prefixes, which the
// context-free pass can't otherwise distinguish.
func elementLocal(name string) string {
	if name == "pxfrm" {
		return "xfrm"
	}
	return name
}

// attrLocal maps a stripped relationship-attribute name to its local part
// (rid→id, rembed→embed, rlink→link); other names pass through.
func attrLocal(name string) string {
	switch name {
	case "rid":
		return "id"
	case "rembed":
		return "embed"
	case "rlink":
		return "link"
	}
	return name
}

func attrPrefix(name string) string {
	switch name {
	case "rid", "rembed", "rlink":
		return "r"
	}
	return ""
}

func escapeAttr(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func sortedPrefixes(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for p := range m {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// nsURI maps a prefix to its canonical namespace URI.
var nsURI = map[string]string{
	"p": NSPresentationML,
	"a": NSDrawingML,
	"r": NSRelationships,
}

// elementPrefix returns the canonical OOXML prefix for a bare element local
// name, or "" if the element is not namespaced (left bare). The mapping is the
// one the retired hand-rolled writer encoded, extended to the full element set
// the codecs emit. PresentationML elements are p:, DrawingML elements a:.
func elementPrefix(local string) string {
	if p, ok := elementNS[local]; ok {
		return p
	}
	return ""
}

var elementNS = map[string]string{
	// PresentationML (p:)
	"presentation": "p", "sldSz": "p", "notesSz": "p",
	"sldIdLst": "p", "sldId": "p",
	"sldMasterIdLst": "p", "sldMasterId": "p",
	"notesMasterId": "p", "notesMasterIdLst": "p", "sldLayoutIdLst": "p", "sldLayoutId": "p",
	"embeddedFontLst": "p", "embeddedFont": "p", "font": "p",
	"regular": "p", "bold": "p", "italic": "p", "boldItalic": "p",
	"sld": "p", "sldLayout": "p", "sldMaster": "p",
	"cSld": "p", "clrMap": "p", "clrMapOvr": "p", "bg": "p", "bgPr": "p", "bgRef": "p",
	"masterClrMapping": "a", "overrideClrMapping": "a",
	"spTree": "p", "nvGrpSpPr": "p", "grpSpPr": "p",
	"sp": "p", "nvSpPr": "p", "cNvSpPr": "p", "cNvPr": "p", "nvPr": "p", "ph": "p",
	"cNvGrpSpPr": "p", "spPr": "p", "txBody": "p",
	"pic": "p", "nvPicPr": "p", "cNvPicPr": "p", "blipFill": "p",
	"graphicFrame": "p", "nvGraphicFramePr": "p", "cNvGraphicFramePr": "p",
	"pxfrm":         "p", // graphic-frame transform → p:xfrm (see elementLocal)
	"printSettings": "p", "compatSpt": "p",

	// DrawingML (a:)
	"xfrm": "a", "off": "a", "ext": "a",
	"prstGeom": "a", "custGeom": "a", "avLst": "a", "gd": "a", "rect": "a",
	"gdLst": "a", "pathLst": "a", "path": "a", "pt": "a",
	"moveTo": "a", "lnTo": "a", "cubicBezTo": "a", "quadBezTo": "a", "close": "a",
	"solidFill": "a", "noFill": "a", "gradFill": "a", "pattFill": "a", "grpFill": "a",
	"srgbClr": "a", "schemeClr": "a", "sysClr": "a", "scrgbClr": "a", "prstClr": "a",
	"gs": "a", "gsLst": "a", "lin": "a", "alpha": "a", "lumMod": "a", "lumOff": "a",
	"fillToRect": "a",
	"ln": "a", "prstDash": "a", "lstStyle": "a", "bodyPr": "a",
	"effectLst": "a", "outerShdw": "a",
	"normAutofit": "a", "spAutoFit": "a", "noAutofit": "a", "highlight": "a",
	"p": "a", "pPr": "a", "r": "a", "rPr": "a", "t": "a", "br": "a", "fld": "a",
	"latin": "a", "ea": "a", "cs": "a", "hlinkClick": "a",
	"blip": "a", "stretch": "a", "fillRect": "a", "tile": "a", "srcRect": "a",
	"alphaModFix": "a", "duotone": "a",
	"graphic": "a", "graphicData": "a",
	"tbl": "a", "tblPr": "a", "tblGrid": "a", "gridCol": "a", "tr": "a", "tc": "a",
	"tcPr": "a", "tableStyleId": "a", "lnL": "a", "lnR": "a", "lnT": "a", "lnB": "a",
	"buChar": "a", "buAutoNum": "a", "buNone": "a", "defRPr": "a", "defRgbClrModel": "a",
	"lnSpc": "a", "spcPct": "a", // paragraph line spacing (D-061)
}
