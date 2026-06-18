package slide

import "encoding/xml"

// The shape tree (<p:spTree>) holds a heterogeneous, ordered list of children
// — shapes, pictures and graphic frames — which Go's struct-tag marshaling
// cannot express (a single field cannot carry mixed element types in order).
// A custom MarshalXML/UnmarshalXML pair keeps that ordering on write and
// reconstructs it on read, so a self-authored slide round-trips losslessly
// (D-032). Elements are emitted bare; ooxml.RestoreNamespaces re-applies the
// p:/a: prefixes.

// xNvGrpSpPrOut is the write shape of <p:nvGrpSpPr> (cNvPr, cNvGrpSpPr, nvPr).
type xNvGrpSpPrOut struct {
	XMLName    struct{}   `xml:"nvGrpSpPr"`
	CNvPr      xCNvPrOut  `xml:"cNvPr"`
	CNvGrpSpPr XEmptyElem `xml:"cNvGrpSpPr"`
	NvPr       XEmptyElem `xml:"nvPr"`
}

// xCNvPrOut is the non-visual drawing props of the group (id + name required).
type xCNvPrOut struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// xGrpSpPrOut is the (empty) group shape properties <p:grpSpPr/>.
type xGrpSpPrOut struct {
	XMLName struct{}      `xml:"grpSpPr"`
	Xfrm    *XTransform2D `xml:"xfrm,omitempty"`
}

// MarshalXML serializes the shape tree: the group's non-visual + shape
// properties, then each ordered child via its own struct tags.
func (xst *XSpTree) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start = xml.StartElement{Name: xml.Name{Local: "spTree"}}
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	id, name := 1, "Layout"
	if xst.NonVisual.CNvPr != nil {
		id = xst.NonVisual.CNvPr.ID
		name = xst.NonVisual.CNvPr.Name
	}
	if err := e.Encode(xNvGrpSpPrOut{CNvPr: xCNvPrOut{ID: id, Name: name}}); err != nil {
		return err
	}

	grp := xGrpSpPrOut{}
	if xst.GroupShapeProperties != nil {
		grp.Xfrm = xst.GroupShapeProperties.Xfrm
	}
	if err := e.Encode(grp); err != nil {
		return err
	}

	for _, child := range xst.Children {
		if err := e.Encode(child); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// xGraphicFrameOut is the write shape of <p:graphicFrame>. Its transform is
// emitted under the sentinel local name "pxfrm" so RestoreNamespaces renames it
// to <p:xfrm> — the graphic frame's PresentationML transform, distinct from a
// shape's <a:xfrm>. (Read keeps XGraphicFrame's "xfrm" tag; Go's unmarshal is
// context-aware, so round-trip is unaffected.)
type xGraphicFrameOut struct {
	XMLName     struct{}               `xml:"graphicFrame"`
	NonVisual   XNonVisualGraphicFrame `xml:"nvGraphicFramePr"`
	Transform2D *XTransform2D          `xml:"pxfrm,omitempty"`
	Graphic     *XGraphic              `xml:"graphic,omitempty"`
}

// MarshalXML serializes a graphic frame with a PresentationML p:xfrm transform.
func (gf *XGraphicFrame) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	return e.Encode(xGraphicFrameOut{
		NonVisual:   gf.NonVisual,
		Transform2D: gf.Transform2D,
		Graphic:     gf.Graphic,
	})
}

// Runs returns the paragraph's text runs in document order (breaks excluded).
func (p *XTextParagraph) Runs() []*XTextRun {
	var out []*XTextRun
	for _, c := range p.Content {
		if r, ok := c.(*XTextRun); ok {
			out = append(out, r)
		}
	}
	return out
}

// MarshalXML serializes a paragraph: an optional <a:pPr> then its ordered run
// and break children. A single field cannot interleave <a:r> and <a:br> in
// document order, so the content is encoded element-by-element.
func (p *XTextParagraph) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start = xml.StartElement{Name: xml.Name{Local: "p"}}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	if p.Pr != nil {
		if err := e.Encode(p.Pr); err != nil {
			return err
		}
	}
	for _, c := range p.Content {
		if err := e.Encode(c); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// UnmarshalXML reconstructs a paragraph, preserving run/break order. Names are
// bare because FromXML strips namespace prefixes before decoding.
func (p *XTextParagraph) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	p.Pr = nil
	p.Content = p.Content[:0]
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pPr":
				var pr XParaProps
				if err := d.DecodeElement(&pr, &t); err != nil {
					return err
				}
				p.Pr = &pr
			case "r":
				var r XTextRun
				if err := d.DecodeElement(&r, &t); err != nil {
					return err
				}
				p.Content = append(p.Content, &r)
			case "br":
				var b XTextBreak
				if err := d.DecodeElement(&b, &t); err != nil {
					return err
				}
				p.Content = append(p.Content, &b)
			default:
				// An unrecognized paragraph child (e.g. <a:fld> field, math) is
				// ignored, but its name is recorded so the reader can surface a
				// nested dropped-element warning (D-048).
				p.dropped = append(p.dropped, t.Name.Local)
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}

// UnmarshalXML reconstructs the shape tree from its child elements, preserving
// order. Names are bare here because FromXML strips namespace prefixes before
// decoding.
func (xst *XSpTree) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	xst.Children = xst.Children[:0]
	xst.dropped = xst.dropped[:0]
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "nvGrpSpPr":
				var nv nvGrpSpPr
				if err := d.DecodeElement(&nv, &t); err != nil {
					return err
				}
				xst.NonVisual = nv
			case "grpSpPr":
				var grp XGroupShapeProperties
				if err := d.DecodeElement(&grp, &t); err != nil {
					return err
				}
				if grp.Xfrm != nil {
					xst.GroupShapeProperties = &grp
				}
			case "sp":
				var sp XSp
				if err := d.DecodeElement(&sp, &t); err != nil {
					return err
				}
				xst.Children = append(xst.Children, &sp)
			case "pic":
				var pic XPicture
				if err := d.DecodeElement(&pic, &t); err != nil {
					return err
				}
				xst.Children = append(xst.Children, &pic)
			case "graphicFrame":
				var gf XGraphicFrame
				if err := d.DecodeElement(&gf, &t); err != nil {
					return err
				}
				xst.Children = append(xst.Children, &gf)
			default:
				// Best-effort external read (RFC §16, D-048): an unrecognized
				// shape-tree child (group shape, mc:AlternateContent, …) is
				// ignored, but its name is recorded so the reader can surface a
				// dropped-element warning.
				xst.dropped = append(xst.dropped, t.Name.Local)
				if err := d.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}
