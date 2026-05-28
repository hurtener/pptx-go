package parts

// ============================================================================
// App Properties XML type definitions - corresponds to /docProps/app.xml
// ============================================================================
//
// Application properties follow the OpenXML specification.
// Namespace: http://schemas.openxmlformats.org/officeDocument/2006/extended-properties
// File location: /docProps/app.xml
//
// ============================================================================

import "encoding/xml"

// XMLAppProps is the XML structure for application properties (/docProps/app.xml).
type XMLAppProps struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/officeDocument/2006/extended-properties Properties"`

	// namespace declarations (for serialization)
	XmlnsProp string `xml:"xmlns,attr,omitempty"`
	XmlnsVt   string `xml:"xmlns:vt,attr,omitempty"`

	// application information
	Application string `xml:"Application,omitempty"` // application name
	AppVersion  string `xml:"AppVersion,omitempty"`  // application version

	Security string `xml:"DocSecurity,omitempty"` // document security level

	// document statistics
	TotalTime    *int  `xml:"TotalTime,omitempty"`    // total editing time (minutes)
	Words        *int  `xml:"Words,omitempty"`        // word count
	Characters   *int  `xml:"Characters,omitempty"`   // character count
	Pages        *int  `xml:"Pages,omitempty"`        // page count
	Paragraphs   *int  `xml:"Paragraphs,omitempty"`   // paragraph count
	Slides       *int  `xml:"Slides,omitempty"`       // slide count
	Notes        *int  `xml:"Notes,omitempty"`        // notes count
	HiddenSlides *int  `xml:"HiddenSlides,omitempty"` // hidden slide count
	MMClips      *int  `xml:"MMClips,omitempty"`      // multimedia clip count
	ScaleCrop    *bool `xml:"ScaleCrop,omitempty"`    // scale-crop flag

	// organization info
	Company string `xml:"Company,omitempty"` // company name
	Manager string `xml:"Manager,omitempty"` // manager name

	// hyperlink info
	HyperlinkBase     string `xml:"HyperlinkBase,omitempty"`     // hyperlink base URL
	LinksUpToDate     *bool  `xml:"LinksUpToDate,omitempty"`     // whether links are up to date
	HyperlinksChanged *bool  `xml:"HyperlinksChanged,omitempty"` // whether hyperlinks changed
	SharedDoc         *bool  `xml:"SharedDoc,omitempty"`         // shared document flag

	// heading pairs and part titles (raw InnerXML preserves the original structure)
	HeadingPairs  *XMLHeadingPairs  `xml:"HeadingPairs,omitempty"`
	TitlesOfParts *XMLTitlesOfParts `xml:"TitlesOfParts,omitempty"`

	// template info
	Template string `xml:"Template,omitempty"` // document template

	// other properties
	PresentationFormat string `xml:"PresentationFormat,omitempty"` // presentation format
	LineSketches       *bool  `xml:"LineSketches,omitempty"`       // line sketches flag
}

// XMLHeadingPairs holds the HeadingPairs element.
// The structure is complex; raw XML is preserved via InnerXML.
// Namespace: same as the parent Properties element (extended-properties).
type XMLHeadingPairs struct {
	XMLName  xml.Name `xml:"HeadingPairs"`
	InnerXML string   `xml:",innerxml"` // preserves the original XML content
}

// XMLTitlesOfParts holds the TitlesOfParts element.
// Namespace: same as the parent Properties element (extended-properties).
type XMLTitlesOfParts struct {
	XMLName  xml.Name `xml:"TitlesOfParts"`
	InnerXML string   `xml:",innerxml"` // preserves the original XML content
}

// ============================================================================
// Default constants
// ============================================================================

// DefaultAppProps is the default application properties template.
var DefaultAppProps = &XMLAppProps{
	Application: "Microsoft Office PowerPoint",
	AppVersion:  "15.0000",
	Company:     "",
	Manager:     "",
}

// ============================================================================
// Namespace constants
// ============================================================================

const (
	// NamespaceExtendedProperties is the extended properties namespace URI.
	NamespaceExtendedProperties = "http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"
	// NamespaceDocPropsVTypes is the document property value types namespace URI.
	NamespaceDocPropsVTypes = "http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"
)
