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

// UnmarshalXML reconstructs the shape tree from its child elements, preserving
// order. Names are bare here because FromXML strips namespace prefixes before
// decoding.
func (xst *XSpTree) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	xst.Children = xst.Children[:0]
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
