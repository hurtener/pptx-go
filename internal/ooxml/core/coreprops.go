package core

import (
	"encoding/xml"

	"github.com/hurtener/pptx-go/internal/ooxml"
)

// ============================================================================
// Core Properties XML struct - corresponds to /docProps/core.xml
// ============================================================================
//
// OpenXML core properties are based on the Dublin Core metadata standard.
// Namespaces:
//   - cp:      http://schemas.openxmlformats.org/package/2006/metadata/core-properties
//   - dc:      http://purl.org/dc/elements/1.1/
//   - dcterms: http://purl.org/dc/terms/
//   - xsi:     http://www.w3.org/2001/XMLSchema-instance
//
// ============================================================================

// XMLCoreProperties is the XML struct for /docProps/core.xml.
// It carries document metadata (title, author, created/modified times, etc.).
type XMLCoreProperties struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties coreProperties"`

	// namespace declarations (for serialization)
	XmlnsCp      string `xml:"xmlns:cp,attr,omitempty"`
	XmlnsDc      string `xml:"xmlns:dc,attr,omitempty"`
	XmlnsDcterms string `xml:"xmlns:dcterms,attr,omitempty"`
	XmlnsXsi     string `xml:"xmlns:xsi,attr,omitempty"`

	// Dublin Core elements (dc: namespace -> http://purl.org/dc/elements/1.1/)
	Title       string `xml:"http://purl.org/dc/elements/1.1/ title,omitempty"`
	Creator     string `xml:"http://purl.org/dc/elements/1.1/ creator,omitempty"`
	Subject     string `xml:"http://purl.org/dc/elements/1.1/ subject,omitempty"`
	Description string `xml:"http://purl.org/dc/elements/1.1/ description,omitempty"`

	// Dublin Core Terms elements (dcterms: namespace -> http://purl.org/dc/terms/)
	Created  *XMLW3CDTFDate `xml:"http://purl.org/dc/terms/ created,omitempty"`
	Modified *XMLW3CDTFDate `xml:"http://purl.org/dc/terms/ modified,omitempty"`

	// Core properties extension (cp: namespace -> http://schemas.openxmlformats.org/package/2006/metadata/core-properties)
	Keywords       string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties keywords,omitempty"`
	LastModifiedBy string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties lastModifiedBy,omitempty"`
	Revision       string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties revision,omitempty"`
	Category       string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties category,omitempty"`
	ContentType    string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties contentType,omitempty"`
	Version        string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties version,omitempty"`
	Identifier     string `xml:"http://schemas.openxmlformats.org/package/2006/metadata/core-properties identifier,omitempty"`
	Language       string `xml:"http://purl.org/dc/elements/1.1/ language,omitempty"`
}

// XMLW3CDTFDate represents a W3CDTF-format date element.
// Example XML: <dcterms:created xsi:type="dcterms:W3CDTF">...</dcterms:created>
// W3CDTF format: YYYY-MM-DDThh:mm:ssZ
type XMLW3CDTFDate struct {
	Type  string `xml:"xsi:type,attr,omitempty"`
	Value string `xml:",chardata"`
}

// ============================================================================
// Constants
// ============================================================================

const (
	// namespace URIs
	NamespaceCoreProperties  = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
	NamespaceDublinCore      = "http://purl.org/dc/elements/1.1/"
	NamespaceDublinCoreTerms = "http://purl.org/dc/terms/"
	NamespaceXMLSchema       = "http://www.w3.org/2001/XMLSchema-instance"

	// W3CDTF type identifier
	W3CDTFType = "dcterms:W3CDTF"
)

// ooxml.XMLDeclaration is defined in xmlutils.go.

// ============================================================================
// Constructor
// ============================================================================

// NewXMLCoreProperties creates a core properties struct with default namespace declarations.
func NewXMLCoreProperties() *XMLCoreProperties {
	return &XMLCoreProperties{
		XmlnsCp:      NamespaceCoreProperties,
		XmlnsDc:      NamespaceDublinCore,
		XmlnsDcterms: NamespaceDublinCoreTerms,
		XmlnsXsi:     NamespaceXMLSchema,
	}
}

// ============================================================================
// Helper methods
// ============================================================================

// SetCreated sets the created timestamp.
func (cp *XMLCoreProperties) SetCreated(value string) {
	cp.Created = &XMLW3CDTFDate{
		Type:  W3CDTFType,
		Value: value,
	}
}

// SetModified sets the last-modified timestamp.
func (cp *XMLCoreProperties) SetModified(value string) {
	cp.Modified = &XMLW3CDTFDate{
		Type:  W3CDTFType,
		Value: value,
	}
}

// GetCreated returns the created timestamp value.
func (cp *XMLCoreProperties) GetCreated() string {
	if cp.Created == nil {
		return ""
	}
	return cp.Created.Value
}

// GetModified returns the last-modified timestamp value.
func (cp *XMLCoreProperties) GetModified() string {
	if cp.Modified == nil {
		return ""
	}
	return cp.Modified.Value
}

// ToXML serializes the core properties to XML bytes.
func (cp *XMLCoreProperties) ToXML() ([]byte, error) {
	output, err := xml.MarshalIndent(cp, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(ooxml.XMLDeclaration), output...), nil
}

// FromXML deserializes core properties from XML bytes.
func (cp *XMLCoreProperties) FromXML(data []byte) error {
	if err := xml.Unmarshal(data, cp); err != nil {
		return err
	}
	return nil
}

// ParseCoreProperties parses core properties from XML bytes.
func ParseCoreProperties(data []byte) (*XMLCoreProperties, error) {
	var cp XMLCoreProperties
	if err := cp.FromXML(data); err != nil {
		return nil, err
	}
	return &cp, nil
}

// ParseCoreProps is a short alias for ParseCoreProperties.
//
// Deprecated: use ParseCoreProperties or FromXML instead.
func ParseCoreProps(data []byte) (*XMLCoreProperties, error) {
	return ParseCoreProperties(data)
}
