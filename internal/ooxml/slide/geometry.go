package slide

import (
	"encoding/xml"
	"strconv"
)

// Custom path geometry (<a:custGeom>) — the wire form an icon renders to
// (RFC §14.1, Phase 12). A custom geometry is a list of paths, each an ordered,
// heterogeneous sequence of drawing commands (moveTo / lnTo / cubicBezTo /
// quadBezTo / close) over the path's own w×h coordinate space, which PowerPoint
// scales to the shape's extent. Go's struct-tag marshaling cannot express an
// ordered mixed-element list, so XPath carries a custom Marshal/Unmarshal pair
// (the same technique XSpTree uses for the shape tree). Elements are emitted
// bare; ooxml.RestoreNamespaces re-applies the a: prefix.

// Path command local element names.
const (
	PathMoveTo  = "moveTo"
	PathLnTo    = "lnTo"
	PathCubicTo = "cubicBezTo"
	PathQuadTo  = "quadBezTo"
	PathClose   = "close"
)

// XCustomGeometry is <a:custGeom>: an empty avLst/gdLst followed by the path
// list. Only the path list carries data in V1 (no adjust handles or connection
// sites).
type XCustomGeometry struct {
	XMLName  struct{}  `xml:"custGeom"`
	AvLst    *XAvLst   `xml:"avLst"`
	GdLst    *XGdLst   `xml:"gdLst"`
	PathList XPathList `xml:"pathLst"`
}

// XGdLst is the empty <a:gdLst/> (no shape guides in V1).
type XGdLst struct{}

// XPathList is <a:pathLst> — one or more paths.
type XPathList struct {
	Paths []XPath `xml:"path"`
}

// XPoint is <a:pt x=".." y=".."/> in the path's coordinate space.
type XPoint struct {
	X int64 `xml:"x,attr"`
	Y int64 `xml:"y,attr"`
}

// XPathCommand is one drawing command: its element name (Cmd) plus the points
// it carries (1 for moveTo/lnTo, 2 for quadBezTo, 3 for cubicBezTo, 0 for
// close).
type XPathCommand struct {
	Cmd string
	Pts []XPoint
}

// XPath is <a:path w=".." h=".."> with an ordered command sequence.
type XPath struct {
	W        int64
	H        int64
	Commands []XPathCommand
}

// MarshalXML writes <path w h> with each command and its points in order.
func (p XPath) MarshalXML(e *xml.Encoder, _ xml.StartElement) error {
	start := xml.StartElement{Name: xml.Name{Local: "path"}}
	if p.W != 0 {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "w"}, Value: strconv.FormatInt(p.W, 10)})
	}
	if p.H != 0 {
		start.Attr = append(start.Attr, xml.Attr{Name: xml.Name{Local: "h"}, Value: strconv.FormatInt(p.H, 10)})
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, c := range p.Commands {
		cs := xml.StartElement{Name: xml.Name{Local: c.Cmd}}
		if err := e.EncodeToken(cs); err != nil {
			return err
		}
		for _, pt := range c.Pts {
			pe := xml.StartElement{Name: xml.Name{Local: "pt"}, Attr: []xml.Attr{
				{Name: xml.Name{Local: "x"}, Value: strconv.FormatInt(pt.X, 10)},
				{Name: xml.Name{Local: "y"}, Value: strconv.FormatInt(pt.Y, 10)},
			}}
			if err := e.EncodeToken(pe); err != nil {
				return err
			}
			if err := e.EncodeToken(xml.EndElement{Name: pe.Name}); err != nil {
				return err
			}
		}
		if err := e.EncodeToken(xml.EndElement{Name: cs.Name}); err != nil {
			return err
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// UnmarshalXML reconstructs an XPath (w/h + ordered commands) so a self-authored
// custGeom shape round-trips losslessly (D-032). Element matching ignores the
// namespace (read XML carries the a: URI; write emits bare).
func (p *XPath) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		switch a.Name.Local {
		case "w":
			p.W, _ = strconv.ParseInt(a.Value, 10, 64)
		case "h":
			p.H, _ = strconv.ParseInt(a.Value, 10, 64)
		}
	}
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "pt" { // unexpected stray point
				if err := d.Skip(); err != nil {
					return err
				}
				continue
			}
			cmd := XPathCommand{Cmd: t.Name.Local}
			pts, err := readPoints(d, t.Name.Local)
			if err != nil {
				return err
			}
			cmd.Pts = pts
			p.Commands = append(p.Commands, cmd)
		case xml.EndElement:
			if t.Name.Local == "path" {
				return nil
			}
		}
	}
}

// readPoints consumes the <a:pt/> children of a command element until its end.
func readPoints(d *xml.Decoder, cmdLocal string) ([]XPoint, error) {
	var pts []XPoint
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "pt" {
				var pt XPoint
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "x":
						pt.X, _ = strconv.ParseInt(a.Value, 10, 64)
					case "y":
						pt.Y, _ = strconv.ParseInt(a.Value, 10, 64)
					}
				}
				pts = append(pts, pt)
			}
			if err := d.Skip(); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if t.Name.Local == cmdLocal {
				return pts, nil
			}
		}
	}
}
