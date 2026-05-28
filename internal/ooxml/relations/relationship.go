package relations

import (
	"encoding/xml"

	"github.com/hurtener/pptx-go/internal/ooxml"
)

// ============================================================================
// OpenXML Relationships XML structs - correspond to *.rels files
// ============================================================================
//
// Relationship file locations:
//   - package level: /_rels/.rels
//   - slide:         /ppt/slides/_rels/slide1.xml.rels
//   - master:        /ppt/slideMasters/_rels/slideMaster1.xml.rels
//
// Namespace: http://schemas.openxmlformats.org/package/2006/relationships
// ============================================================================

// XMLRelationships is the root element of a .rels file.
// XML: <Relationships xmlns="...">...</Relationships>
type XMLRelationships struct {
	XMLName       xml.Name          `xml:"Relationships"`
	Xmlns         string            `xml:"xmlns,attr,omitempty"`
	Relationships []XMLRelationship `xml:"Relationship"`
}

// XMLRelationship represents a single relationship entry.
// XML: <Relationship Id="rId1" Type="..." Target="..."/>
type XMLRelationship struct {
	ID         string `xml:"Id,attr"`                   // relationship ID (e.g. rId1, rId2)
	Type       string `xml:"Type,attr"`                 // relationship type URI
	Target     string `xml:"Target,attr"`               // target path (relative or absolute)
	TargetMode string `xml:"TargetMode,attr,omitempty"` // Internal (default) or External
}

// ============================================================================
// Constants
// ============================================================================

const (
	// NamespaceRelationships is the OPC relationships namespace URI.
	NamespaceRelationships = "http://schemas.openxmlformats.org/package/2006/relationships"

	// target mode values
	TargetModeInternal = "Internal"
	TargetModeExternal = "External"

	// common relationship type URIs
	RelTypeImage       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	RelTypeHyperlink   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
	RelTypeSlide       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide"
	RelTypeSlideLayout = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout"
	RelTypeSlideMaster = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster"
	RelTypeTheme       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"
	RelTypeNotesSlide  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide"
	RelTypeComments    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/comments"
	RelTypeChart       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart"
	RelTypeTable       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/table"
	RelTypeMedia       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/video"
	RelTypeAudio       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/audio"
)

// ============================================================================
// Constructors
// ============================================================================

// NewXMLRelationships creates a relationships collection with the default namespace.
func NewXMLRelationships() *XMLRelationships {
	return &XMLRelationships{
		Xmlns:         NamespaceRelationships,
		Relationships: make([]XMLRelationship, 0),
	}
}

// NewXMLRelationship creates a single relationship.
func NewXMLRelationship(id, relType, target string) XMLRelationship {
	return XMLRelationship{
		ID:     id,
		Type:   relType,
		Target: target,
	}
}

// NewXMLRelationshipExternal creates an external relationship.
func NewXMLRelationshipExternal(id, relType, target string) XMLRelationship {
	return XMLRelationship{
		ID:         id,
		Type:       relType,
		Target:     target,
		TargetMode: TargetModeExternal,
	}
}

// ============================================================================
// Helper methods
// ============================================================================

// Add appends a relationship to the collection.
func (rs *XMLRelationships) Add(rel XMLRelationship) {
	rs.Relationships = append(rs.Relationships, rel)
}

// AddNew creates and appends a new relationship.
func (rs *XMLRelationships) AddNew(id, relType, target string) {
	rs.Add(NewXMLRelationship(id, relType, target))
}

// GetByID returns the relationship with the given ID, or nil if not found.
func (rs *XMLRelationships) GetByID(id string) *XMLRelationship {
	for i := range rs.Relationships {
		if rs.Relationships[i].ID == id {
			return &rs.Relationships[i]
		}
	}
	return nil
}

// GetByType returns all relationships of the given type.
func (rs *XMLRelationships) GetByType(relType string) []XMLRelationship {
	var result []XMLRelationship
	for _, rel := range rs.Relationships {
		if rel.Type == relType {
			result = append(result, rel)
		}
	}
	return result
}

// GetByTarget returns the relationship with the given target path, or nil if not found.
func (rs *XMLRelationships) GetByTarget(target string) *XMLRelationship {
	for i := range rs.Relationships {
		if rs.Relationships[i].Target == target {
			return &rs.Relationships[i]
		}
	}
	return nil
}

// Count returns the number of relationships.
func (rs *XMLRelationships) Count() int {
	return len(rs.Relationships)
}

// IsExternal reports whether this is an external relationship.
func (r *XMLRelationship) IsExternal() bool {
	return r.TargetMode == TargetModeExternal
}

// ============================================================================
// XML serialization / deserialization
// ============================================================================

// ToXML serializes the relationships collection to XML bytes.
func (rs *XMLRelationships) ToXML() ([]byte, error) {
	output, err := xml.MarshalIndent(rs, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(ooxml.XMLDeclaration), output...), nil
}

// FromXML deserializes a relationships collection from XML bytes.
func (rs *XMLRelationships) FromXML(data []byte) error {
	if err := xml.Unmarshal(data, rs); err != nil {
		return err
	}
	return nil
}

// ParseRelationships parses a relationships collection from XML bytes.
func ParseRelationships(data []byte) (*XMLRelationships, error) {
	var rs XMLRelationships
	if err := rs.FromXML(data); err != nil {
		return nil, err
	}
	return &rs, nil
}
