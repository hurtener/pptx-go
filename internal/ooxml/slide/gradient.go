package slide

// Gradient fill wire types (<a:gradFill>, RFC §8.3, D-041). A gradient is a list
// of color stops plus a direction: <a:lin> (linear, by angle) or <a:path
// path="circle"> + <a:fillToRect> (radial / focal). Stops reuse XSrgbClr, so an
// alpha stop (a glow's transparent edge) is expressed by the srgbClr's <a:alpha>
// child. These are plain struct-tag types — element names are emitted bare and
// ooxml.RestoreNamespaces re-applies the a: prefix.

// XGradientFill is <a:gradFill>: a stop list and one direction.
type XGradientFill struct {
	XMLName struct{}          `xml:"gradFill"`
	GsLst   XGradientStopList `xml:"gsLst"`
	Lin     *XLinearGradient  `xml:"lin,omitempty"`
	Path    *XPathGradient    `xml:"path,omitempty"`
}

// XGradientStopList is <a:gsLst>.
type XGradientStopList struct {
	Gs []XGradientStop `xml:"gs"`
}

// XGradientStop is <a:gs pos="0..100000">, a color at a position.
type XGradientStop struct {
	Pos     int       `xml:"pos,attr"`
	SrgbClr *XSrgbClr `xml:"srgbClr,omitempty"`
}

// XLinearGradient is <a:lin ang="(1/60000)°">.
type XLinearGradient struct {
	Ang int `xml:"ang,attr"`
}

// XPathGradient is <a:path path="circle|rect|shape"> with an optional focal
// rectangle.
type XPathGradient struct {
	Path       string       `xml:"path,attr"`
	FillToRect *XFillToRect `xml:"fillToRect,omitempty"`
}

// XFillToRect is <a:fillToRect l="" t="" r="" b="">, the gradient's focal region
// (1/1000 % insets from each edge).
type XFillToRect struct {
	L int `xml:"l,attr,omitempty"`
	T int `xml:"t,attr,omitempty"`
	R int `xml:"r,attr,omitempty"`
	B int `xml:"b,attr,omitempty"`
}
