package parts

import "encoding/xml"

// ============================================================================
// OpenXML base structs for encoding/xml deserialization
// ============================================================================
//
// This file contains the lowest-level XML struct definitions required for
// parsing masters and layouts.
// These structs can be composed by higher-level structs (bottom-up principle).
// ============================================================================

// XMLOffset represents a position offset.
// Corresponds to XML: <a:off x="..." y="..."/>
// Expresses the X/Y position of a shape or object (in EMU).
type XMLOffset struct {
	X int64 `xml:"x,attr"`
	Y int64 `xml:"y,attr"`
}

// IsZero reports whether the offset is a zero value (unset or missing attributes).
func (o *XMLOffset) IsZero() bool {
	return o.X == 0 && o.Y == 0
}

// IsValid reports whether the offset is valid.
// Note: coordinates of (0, 0) are technically a valid position.
func (o *XMLOffset) IsValid() bool {
	return o != nil
}

// XMLExtents represents dimensions.
// Corresponds to XML: <a:ext cx="..." cy="..."/>
// Expresses the width/height of a shape or object (in EMU).
type XMLExtents struct {
	Cx int64 `xml:"cx,attr"`
	Cy int64 `xml:"cy,attr"`
}

// IsZero reports whether the extents are a zero value (unset or missing attributes).
func (e *XMLExtents) IsZero() bool {
	return e.Cx == 0 && e.Cy == 0
}

// IsValid reports whether the extents are valid (cx and cy must be positive per the OpenXML spec).
// Zero or negative extents typically indicate an invalid shape.
func (e *XMLExtents) IsValid() bool {
	return e != nil && e.Cx > 0 && e.Cy > 0
}

// XMLTransform represents a 2D transform.
// Corresponds to XML: <a:xfrm>...</a:xfrm>
// Contains position offset and size information.
// Note: Go's xml package handles namespace resolution automatically, so local names are used.
type XMLTransform struct {
	Off *XMLOffset  `xml:"off,omitempty"`
	Ext *XMLExtents `xml:"ext,omitempty"`
}

// XMLPlaceholder represents a placeholder element.
// Corresponds to XML: <p:ph type="..." idx="..."/>
// Marks the type of a fillable region in a master or layout.
type XMLPlaceholder struct {
	Type string `xml:"type,attr,omitempty"`
	Idx  string `xml:"idx,attr,omitempty"`
}

// ============================================================================
// Intermediate wrapper structs
// ============================================================================

// XMLCNvPr represents common non-visual properties.
// Corresponds to XML: <p:cNvPr id="..." name="..."/>
type XMLCNvPr struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:"name,attr,omitempty"`
}

// XMLNvPr represents non-visual properties.
// Corresponds to XML: <p:nvPr>...</p:nvPr>
// Contains a placeholder definition when present.
type XMLNvPr struct {
	Ph *XMLPlaceholder `xml:"ph,omitempty"`
}

// XMLNvSpPr represents non-visual shape properties.
// Corresponds to XML: <p:nvSpPr>...</p:nvSpPr>
// Contains common properties and non-visual properties.
type XMLNvSpPr struct {
	CNvPr *XMLCNvPr `xml:"cNvPr,omitempty"`
	NvPr  *XMLNvPr  `xml:"nvPr,omitempty"`
}

// XMLSpPr represents visual shape properties.
// Corresponds to XML: <p:spPr>...</p:spPr>
// Contains transform information (position and size).
type XMLSpPr struct {
	Xfrm *XMLTransform `xml:"xfrm,omitempty"`
}

// XMLBackground represents a background element.
// Corresponds to XML: <p:bg>...</p:bg>
type XMLBackground struct {
	BgPr  *XMLBackgroundPr  `xml:"bgPr,omitempty"`
	BgRef *XMLBackgroundRef `xml:"bgRef,omitempty"`
}

// XMLBackgroundRef represents a background reference.
// Corresponds to XML: <p:bgRef idx="..."><a:schemeClr val="..."/></p:bgRef>
type XMLBackgroundRef struct {
	Idx string          `xml:"idx,attr,omitempty"`
	Clr *XMLSchemeColor `xml:"schemeClr,omitempty"`
}

// XMLBackgroundPr represents background properties.
// Corresponds to XML: <p:bgPr>...</p:bgPr>
type XMLBackgroundPr struct {
	Fill *XMLFillProperties `xml:",any,omitempty"`
}

// XMLFillProperties is a union type for fill properties.
// Corresponds to XML: <a:solidFill> / <a:gradFill> / <a:blipFill> etc.
type XMLFillProperties struct {
	SolidFill *XMLSolidFill `xml:"a:solidFill,omitempty"`
	GradFill  *XMLGradFill  `xml:"a:gradFill,omitempty"`
	BlipFill  *XMLBlipFill  `xml:"a:blipFill,omitempty"`
	NoFill    *struct{}     `xml:"a:noFill,omitempty"`
}

// XMLSolidFill represents a solid fill.
// Corresponds to XML: <a:solidFill>...</a:solidFill>
type XMLSolidFill struct {
	SrgbClr   *XMLSRgbColor   `xml:"a:srgbClr,omitempty"`
	SchemeClr *XMLSchemeColor `xml:"a:schemeClr,omitempty"`
}

// XMLSRgbColor represents an RGB color.
// Corresponds to XML: <a:srgbClr val="..."/>
type XMLSRgbColor struct {
	Val string `xml:"val,attr,omitempty"`
}

// XMLSchemeColor represents a theme color.
// Corresponds to XML: <a:schemeClr val="..."/>
type XMLSchemeColor struct {
	Val string `xml:"val,attr,omitempty"`
}

// XMLGradFill represents a gradient fill.
// Corresponds to XML: <a:gradFill>...</a:gradFill>
type XMLGradFill struct {
	GsLst *XMLGradientStopList `xml:"a:gsLst,omitempty"`
	Lin   *XMLLinearGradient   `xml:"a:lin,omitempty"`
}

// XMLGradientStopList represents a gradient stop list.
// Corresponds to XML: <a:gsLst>...</a:gsLst>
type XMLGradientStopList struct {
	Stops []XMLGradientStop `xml:"a:gs,omitempty"`
}

// XMLGradientStop represents a gradient color stop.
// Corresponds to XML: <a:gs pos="...">...</a:gs>
type XMLGradientStop struct {
	Pos       int64         `xml:"pos,attr,omitempty"`
	SolidFill *XMLSolidFill `xml:"a:solidFill,omitempty"`
}

// XMLLinearGradient represents a linear gradient.
// Corresponds to XML: <a:lin ang="..." scaled="..."/>
type XMLLinearGradient struct {
	Ang    int64 `xml:"ang,attr,omitempty"`
	Scaled bool  `xml:"scaled,attr,omitempty"`
}

// XMLBlipFill represents a picture fill.
// Corresponds to XML: <a:blipFill>...</a:blipFill>
type XMLBlipFill struct {
	Blip *XMLBlip `xml:"a:blip,omitempty"`
}

// XMLBlip represents a picture reference.
// Corresponds to XML: <a:blip r:embed="..."/>
type XMLBlip struct {
	Embed string `xml:"r:embed,attr,omitempty"`
}

// ============================================================================
// Top-level structs
// ============================================================================

// XMLShape represents a shape element.
// Corresponds to XML: <p:sp>...</p:sp>
// Represents a single shape element on a slide.
type XMLShape struct {
	NvSpPr *XMLNvSpPr `xml:"nvSpPr"`
	SpPr   *XMLSpPr   `xml:"spPr"`
}

// XMLShapeTree represents a shape tree.
// Corresponds to XML: <p:spTree>...</p:spTree>
// A collection of shape elements.
type XMLShapeTree struct {
	NvGrpSpPr   *XMLNvGrpSpPr   `xml:"nvGrpSpPr,omitempty"`
	GrpSpPr     *XMLGrpSpPr     `xml:"grpSpPr,omitempty"`
	Shapes      []XMLShape      `xml:"sp"`
	GroupShapes []XMLGroupShape `xml:"grpSp,omitempty"`
}

// XMLNvGrpSpPr represents non-visual group shape properties.
// Corresponds to XML: <p:nvGrpSpPr>...</p:nvGrpSpPr>
type XMLNvGrpSpPr struct {
	CNvPr      *XMLCNvPr      `xml:"cNvPr,omitempty"`
	CNvGrpSpPr *XMLCNvGrpSpPr `xml:"cNvGrpSpPr,omitempty"`
}

// XMLCNvGrpSpPr represents common non-visual group shape properties.
// Corresponds to XML: <p:cNvGrpSpPr>...</p:cNvGrpSpPr>
type XMLCNvGrpSpPr struct {
	// Usually an empty element; reserved for future extension.
}

// XMLGrpSpPr represents group shape properties.
// Corresponds to XML: <p:grpSpPr>...</p:grpSpPr>
type XMLGrpSpPr struct {
	Xfrm *XMLTransform `xml:"xfrm,omitempty"`
}

// XMLGroupShape represents a group shape.
// Corresponds to XML: <p:grpSp>...</p:grpSp>
type XMLGroupShape struct {
	NvGrpSpPr *XMLNvGrpSpPr `xml:"nvGrpSpPr"`
	GrpSpPr   *XMLGrpSpPr   `xml:"grpSpPr"`
	Shapes    []XMLShape    `xml:"sp"`
}

// XMLCommonSlideData represents common slide data.
// Corresponds to XML: <p:cSld>...</p:cSld>
// Contains background and shape tree.
type XMLCommonSlideData struct {
	Bg     *XMLBackground `xml:"bg,omitempty"`
	SpTree *XMLShapeTree  `xml:"spTree"`
}

// XMLSlideLayout represents a slide layout.
// Corresponds to XML: <p:sldLayout>...</p:sldLayout>
// Root element defining a single layout.
type XMLSlideLayout struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/presentationml/2006/main sldLayout"`

	// Namespace declarations
	XmlnsA string `xml:"xmlns:a,attr,omitempty"`
	XmlnsR string `xml:"xmlns:r,attr,omitempty"`
	XmlnsP string `xml:"xmlns:p,attr,omitempty"`

	CSld *XMLCommonSlideData `xml:"cSld"`
}

// XMLSlideMaster represents a slide master.
// Corresponds to XML: <p:sldMaster>...</p:sldMaster>
// Root element defining a master.
type XMLSlideMaster struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/presentationml/2006/main sldMaster"`

	// Namespace declarations
	XmlnsA string `xml:"xmlns:a,attr,omitempty"`
	XmlnsR string `xml:"xmlns:r,attr,omitempty"`
	XmlnsP string `xml:"xmlns:p,attr,omitempty"`

	CSld *XMLCommonSlideData `xml:"cSld"`
}
