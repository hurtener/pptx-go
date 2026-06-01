package slide

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/hurtener/pptx-go/utils"
)

// ============================================================================
// Slide master / layout XML parser
// ============================================================================
//
// Converts nested XML structs into clean, read-only domain models.
// ============================================================================

// ParseLayout parses a slide layout XML, extracting all placeholder
// positions and types.
func ParseLayout(xmlData []byte) (*SlideLayoutData, error) {
	var xmlLayout XMLSlideLayout
	if err := xml.Unmarshal(xmlData, &xmlLayout); err != nil {
		return nil, fmt.Errorf("failed to parse layout XML: %w", err)
	}

	if xmlLayout.CSld == nil || xmlLayout.CSld.SpTree == nil {
		return nil, fmt.Errorf("layout is missing required nodes p:cSld or p:spTree")
	}

	// extract placeholders
	placeholders := extractPlaceholders(xmlLayout.CSld.SpTree)

	// extract background
	var background *Background
	if xmlLayout.CSld.Bg != nil {
		background = parseBackground(xmlLayout.CSld.Bg)
	}

	name := ""
	if xmlLayout.CSld != nil {
		name = xmlLayout.CSld.Name
	}

	return &SlideLayoutData{
		id:           generateLayoutID(),
		name:         name,
		layoutType:   layoutTypeFromAttr(xmlLayout.Type),
		background:   background,
		placeholders: placeholders,
	}, nil
}

// layoutTypeFromAttr maps the OOXML sldLayout@type attribute to the internal
// SlideLayoutType enum. Unknown or absent types map to SlideLayoutBlank.
func layoutTypeFromAttr(t string) SlideLayoutType {
	switch t {
	case "title":
		return SlideLayoutTitle
	case "obj", "tx":
		return SlideLayoutTitleAndContent
	case "twoObj", "twoTxTwoObj":
		return SlideLayoutTwoContent
	case "fourObj", "vertTx", "clipArtAndVertTx":
		return SlideLayoutComparison
	case "titleOnly":
		return SlideLayoutTitleOnly
	case "objOnly", "objTx", "picTx":
		return SlideLayoutObject
	case "secHead":
		return SlideLayoutTitle
	default:
		return SlideLayoutBlank
	}
}

// ParseMaster parses a slide master XML, extracting all placeholder definitions.
func ParseMaster(xmlData []byte) (*SlideMasterData, error) {
	var xmlMaster XMLSlideMaster
	if err := xml.Unmarshal(xmlData, &xmlMaster); err != nil {
		return nil, fmt.Errorf("failed to parse master XML: %w", err)
	}

	if xmlMaster.CSld == nil || xmlMaster.CSld.SpTree == nil {
		return nil, fmt.Errorf("master is missing required nodes p:cSld or p:spTree")
	}

	// extract placeholders
	placeholders := extractPlaceholders(xmlMaster.CSld.SpTree)

	// extract background
	var background *Background
	if xmlMaster.CSld.Bg != nil {
		background = parseBackground(xmlMaster.CSld.Bg)
	}

	name := ""
	if xmlMaster.CSld != nil {
		name = xmlMaster.CSld.Name
	}

	return &SlideMasterData{
		id:           generateMasterID(),
		name:         name,
		background:   background,
		placeholders: placeholders,
	}, nil
}

// ============================================================================
// Placeholder extraction
// ============================================================================

// extractPlaceholders extracts all placeholders from a shape tree.
func extractPlaceholders(spTree *XMLShapeTree) map[string]*Placeholder {
	placeholders := make(map[string]*Placeholder)

	// extract from regular shapes
	for _, shape := range spTree.Shapes {
		ph := extractPlaceholderFromShape(shape)
		if ph != nil {
			placeholders[ph.id] = ph
		}
	}

	// recurse into group shapes
	for _, grpSp := range spTree.GroupShapes {
		extractPlaceholdersFromGroup(&grpSp, placeholders)
	}

	return placeholders
}

// extractPlaceholdersFromGroup recursively extracts placeholders from a group shape.
func extractPlaceholdersFromGroup(grpSp *XMLGroupShape, placeholders map[string]*Placeholder) {
	for _, shape := range grpSp.Shapes {
		ph := extractPlaceholderFromShape(shape)
		if ph != nil {
			placeholders[ph.id] = ph
		}
	}
}

// extractPlaceholderFromShape extracts placeholder information from a single shape.
func extractPlaceholderFromShape(shape XMLShape) *Placeholder {
	// skip shapes that have no placeholder element
	if shape.NvSpPr == nil || shape.NvSpPr.NvPr == nil || shape.NvSpPr.NvPr.Ph == nil {
		return nil
	}

	xmlPh := shape.NvSpPr.NvPr.Ph

	// parse placeholder type
	phType := parsePlaceholderType(xmlPh.Type)

	// generate placeholder ID
	phID := generatePlaceholderID(xmlPh.Idx, shape.NvSpPr.CNvPr)

	// extract position and size
	var x, y, cx, cy int64
	if shape.SpPr != nil && shape.SpPr.Xfrm != nil {
		if shape.SpPr.Xfrm.Off != nil {
			x = shape.SpPr.Xfrm.Off.X
			y = shape.SpPr.Xfrm.Off.Y
		}
		if shape.SpPr.Xfrm.Ext != nil {
			cx = shape.SpPr.Xfrm.Ext.Cx
			cy = shape.SpPr.Xfrm.Ext.Cy
		}
	}

	return &Placeholder{
		id:              phID,
		placeholderType: phType,
		x:               x,
		y:               y,
		cx:              cx,
		cy:              cy,
	}
}

// ============================================================================
// Background parsing
// ============================================================================

// parseBackground parses a background XML element into a Background struct.
func parseBackground(xmlBg *XMLBackground) *Background {
	// handle <p:bgRef> background reference (theme color reference)
	if xmlBg.BgRef != nil {
		return &Background{
			backgroundType: BackgroundTypeThemeColor,
			solidColorRGB:  xmlBg.BgRef.Clr.Val, // theme color name, e.g. "bg1"
			opacity:        1.0,
		}
	}

	// handle <p:bgPr> background properties
	if xmlBg.BgPr == nil || xmlBg.BgPr.Fill == nil {
		return nil
	}

	fill := xmlBg.BgPr.Fill

	// solid fill
	if fill.SolidFill != nil {
		rgb := extractColorFromSolidFill(fill.SolidFill)
		if rgb != "" {
			return &Background{
				backgroundType: BackgroundTypeSolidColor,
				solidColorRGB:  rgb,
				opacity:        1.0,
			}
		}
	}

	// gradient fill
	if fill.GradFill != nil {
		return parseGradientBackground(fill.GradFill)
	}

	// picture fill
	if fill.BlipFill != nil && fill.BlipFill.Blip != nil {
		return &Background{
			backgroundType: BackgroundTypePicture,
			pictureRId:     fill.BlipFill.Blip.Embed,
			opacity:        1.0,
		}
	}

	// no fill
	if fill.NoFill != nil {
		return &Background{
			backgroundType: BackgroundTypeNone,
			opacity:        1.0,
		}
	}

	return nil
}

// parseGradientBackground parses a gradient fill into a Background struct.
func parseGradientBackground(gradFill *XMLGradFill) *Background {
	bg := &Background{
		backgroundType: BackgroundTypeGradient,
		opacity:        1.0,
	}

	// parse gradient angle
	if gradFill.Lin != nil {
		bg.gradientAngle = int32(gradFill.Lin.Ang / 60000) // convert to degrees
	}

	// parse gradient stops
	if gradFill.GsLst != nil {
		bg.gradientColors = make([]GradientStop, 0, len(gradFill.GsLst.Stops))
		for _, stop := range gradFill.GsLst.Stops {
			rgb := ""
			if stop.SolidFill != nil {
				rgb = extractColorFromSolidFill(stop.SolidFill)
			}
			if rgb != "" {
				bg.gradientColors = append(bg.gradientColors, GradientStop{
					position: float32(stop.Pos) / 100000.0, // convert to 0.0–1.0
					colorRGB: rgb,
				})
			}
		}
	}

	return bg
}

// extractColorFromSolidFill extracts an RGB hex value from a solid fill element.
func extractColorFromSolidFill(solidFill *XMLSolidFill) string {
	if solidFill.SrgbClr != nil {
		return solidFill.SrgbClr.Val
	}
	if solidFill.SchemeClr != nil {
		// theme colors require a lookup table; return empty for now
		// can be extended to return theme color names in the future
		return ""
	}
	return ""
}

// ============================================================================
// Placeholder type conversion
// ============================================================================

// parsePlaceholderType converts an XML type string to a PlaceholderType enum value.
func parsePlaceholderType(typeStr string) PlaceholderType {
	switch typeStr {
	case "title":
		return PlaceholderTypeTitle
	case "body":
		return PlaceholderTypeBody
	case "ctrTitle":
		return PlaceholderTypeCenterTitle
	case "subTitle":
		return PlaceholderTypeSubTitle
	case "dt":
		return PlaceholderTypeDateTime
	case "sldNum":
		return PlaceholderTypeSlideNumber
	case "ftr":
		return PlaceholderTypeFooter
	case "hdr":
		return PlaceholderTypeHeader
	case "obj":
		return PlaceholderTypeObject
	case "chart":
		return PlaceholderTypeChart
	case "tbl":
		return PlaceholderTypeTable
	case "clipArt":
		return PlaceholderTypeClipArt
	case "dgm":
		return PlaceholderTypeOrgChart
	case "media":
		return PlaceholderTypeMedia
	case "sldImg":
		return PlaceholderTypeSlideImage
	case "pic":
		return PlaceholderTypePicture
	default:
		return PlaceholderTypeNone
	}
}

// ============================================================================
// ID generation
// ============================================================================

var (
	layoutIDCounter int64
	masterIDCounter int64
)

// generateLayoutID generates a unique layout ID.
func generateLayoutID() string {
	layoutIDCounter++
	return "layout_" + strconv.FormatInt(layoutIDCounter, 10)
}

// generateMasterID generates a unique master ID.
func generateMasterID() string {
	masterIDCounter++
	return "master_" + strconv.FormatInt(masterIDCounter, 10)
}

// generatePlaceholderID generates a placeholder ID.
// Prefers the XML idx attribute; falls back to the cNvPr id.
func generatePlaceholderID(xmlIdx string, cnvPr *XMLCNvPr) string {
	if xmlIdx != "" {
		return "ph_" + xmlIdx
	}
	if cnvPr != nil && cnvPr.ID > 0 {
		return "ph_" + strconv.Itoa(cnvPr.ID)
	}
	// fallback ID
	return "ph_unknown"
}

// ============================================================================
// EMU unit conversion helpers (convenience aliases)
// ============================================================================

// EMUToPixels converts EMU to pixels at 96 DPI.
var EMUToPixels = utils.EMUToPixels

// EMUToPoints converts EMU to points.
var EMUToPoints = utils.EMUToPoints

// EMUToInches converts EMU to inches.
var EMUToInches = utils.EMUToInches

// EMUToCentimeters converts EMU to centimeters.
var EMUToCentimeters = utils.EMUToCentimeters
