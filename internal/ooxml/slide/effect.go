package slide

// Effect wire types (<a:effectLst>, RFC §8, D-043). V1 ships a single effect —
// the outer drop shadow (<a:outerShdw>) — which realizes the Theme's Elevation
// token on a shape. These are plain struct-tag types: element names are emitted
// bare and ooxml.RestoreNamespaces re-applies the a: prefix.

// XEffectList is <a:effectLst>: the ordered effect container on a shape's
// properties. V1 carries at most an outer shadow; the element is omitted
// entirely for a flat (no-shadow) shape.
type XEffectList struct {
	XMLName   struct{}      `xml:"effectLst"`
	OuterShdw *XOuterShadow `xml:"outerShdw,omitempty"`
}

// XOuterShadow is <a:outerShdw>: a drop shadow cast outside the shape. The
// blur radius and offset distance are EMU; the direction is an angle in
// 1/60000° (0 = east, 5400000 = straight down). rotWithShape="0" keeps the
// shadow fixed when the shape rotates (V1 cards are unrotated; this is the
// stable default). The shadow color carries its alpha as an <a:srgbClr> child.
type XOuterShadow struct {
	BlurRad      int       `xml:"blurRad,attr,omitempty"`
	Dist         int       `xml:"dist,attr,omitempty"`
	Dir          int       `xml:"dir,attr,omitempty"`
	RotWithShape int       `xml:"rotWithShape,attr"`
	SrgbClr      *XSrgbClr `xml:"srgbClr,omitempty"`
}
