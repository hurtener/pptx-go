package opc

import (
	"encoding/xml"
	"fmt"
	"strings"
	"sync"
)

// ContentTypes represents the contents of [Content_Types].xml,
// which defines the content type of every part in the package.
type ContentTypes struct {
	defaults  map[string]string // extension -> content type
	overrides map[string]string // URI -> content type
	mu        sync.RWMutex
}

// NewContentTypes creates a new ContentTypes definition populated with the default mappings.
func NewContentTypes() *ContentTypes {
	ct := &ContentTypes{
		defaults:  make(map[string]string),
		overrides: make(map[string]string),
	}
	// Seed with the built-in default content types.
	for ext, ctType := range DefaultContentTypes {
		ct.defaults[ext] = ctType
	}
	return ct
}

// AddDefault adds a default content type mapping for the given extension.
func (ct *ContentTypes) AddDefault(extension, contentType string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.defaults[extension] = contentType
}

// AddOverride adds a content type override for a specific URI.
func (ct *ContentTypes) AddOverride(uri *PackURI, contentType string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.overrides[uri.URI()] = contentType
}

// GetContentType returns the content type for the given URI.
// Overrides take precedence over defaults.
func (ct *ContentTypes) GetContentType(uri *PackURI) string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	// Check overrides first.
	if ct, ok := ct.overrides[uri.URI()]; ok {
		return ct
	}

	// Fall back to defaults by extension.
	ext := uri.Extension()
	if ct, ok := ct.defaults[ext]; ok {
		return ct
	}

	return ContentTypeDefault
}

// GetDefault returns the default content type for the given extension.
func (ct *ContentTypes) GetDefault(extension string) string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.defaults[extension]
}

// GetOverride returns the content type override for the given URI.
func (ct *ContentTypes) GetOverride(uri *PackURI) string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.overrides[uri.URI()]
}

// RemoveOverride removes the content type override for the given URI.
func (ct *ContentTypes) RemoveOverride(uri *PackURI) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	delete(ct.overrides, uri.URI())
}

// Defaults returns a snapshot of all default content type mappings.
func (ct *ContentTypes) Defaults() map[string]string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make(map[string]string, len(ct.defaults))
	for k, v := range ct.defaults {
		result[k] = v
	}
	return result
}

// Overrides returns a snapshot of all content type overrides.
func (ct *ContentTypes) Overrides() map[string]string {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	result := make(map[string]string, len(ct.overrides))
	for k, v := range ct.overrides {
		result[k] = v
	}
	return result
}

// ===== XML serialization =====

// XContentTypes is the root element for XML serialization of content types.
type XContentTypes struct {
	XMLName   xml.Name    `xml:"Types"`
	Xmlns     string      `xml:"xmlns,attr"`
	Defaults  []XDefault  `xml:"Default"`
	Overrides []XOverride `xml:"Override"`
}

// XDefault is the XML representation of a default content type entry.
type XDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// XOverride is the XML representation of a content type override entry.
type XOverride struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// FromXML parses content types from XML data.
func (ct *ContentTypes) FromXML(data []byte) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	var xct XContentTypes
	if err := xml.Unmarshal(data, &xct); err != nil {
		return fmt.Errorf("failed to unmarshal content types: %w", err)
	}

	ct.defaults = make(map[string]string)
	ct.overrides = make(map[string]string)

	for _, d := range xct.Defaults {
		ct.defaults[d.Extension] = d.ContentType
	}

	for _, o := range xct.Overrides {
		// PartName is typically an absolute path such as /ppt/presentation.xml.
		ct.overrides[o.PartName] = o.ContentType
	}

	return nil
}

// ToXML serializes the content types to XML.
func (ct *ContentTypes) ToXML() ([]byte, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	xct := XContentTypes{
		Xmlns: NamespaceOPCPackage,
	}

	for ext, ctType := range ct.defaults {
		xct.Defaults = append(xct.Defaults, XDefault{
			Extension:   strings.TrimPrefix(ext, "."),
			ContentType: ctType,
		})
	}

	for uri, ctType := range ct.overrides {
		xct.Overrides = append(xct.Overrides, XOverride{
			PartName:    uri,
			ContentType: ctType,
		})
	}

	output, err := xml.MarshalIndent(xct, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content types: %w", err)
	}

	return append([]byte(XMLDeclaration), output...), nil
}
