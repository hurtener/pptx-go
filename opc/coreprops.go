package opc

import (
	"encoding/xml"
	"fmt"
	"sync"
)

// CoreProperties represents the core properties (Dublin Core metadata) of a package.
type CoreProperties struct {
	title          string // Title
	creator        string // Creator
	subject        string // Subject
	description    string // Description
	keywords       string // Keywords
	created        string // Created timestamp
	modified       string // Modified timestamp
	lastModifiedBy string // Last modified by
	revision       string // Revision number
	category       string // Category
	contentType    string // Content type
	language       string // Language
	mu             sync.RWMutex
}

// --- Getters ---

// Title returns the title.
func (cp *CoreProperties) Title() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.title
}

// Creator returns the creator.
func (cp *CoreProperties) Creator() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.creator
}

// Subject returns the subject.
func (cp *CoreProperties) Subject() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.subject
}

// Description returns the description.
func (cp *CoreProperties) Description() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.description
}

// Keywords returns the keywords.
func (cp *CoreProperties) Keywords() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.keywords
}

// Created returns the created timestamp.
func (cp *CoreProperties) Created() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.created
}

// Modified returns the modified timestamp.
func (cp *CoreProperties) Modified() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.modified
}

// LastModifiedBy returns the name of the last modifier.
func (cp *CoreProperties) LastModifiedBy() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.lastModifiedBy
}

// Revision returns the revision number.
func (cp *CoreProperties) Revision() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.revision
}

// Category returns the category.
func (cp *CoreProperties) Category() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.category
}

// ContentType returns the content type.
func (cp *CoreProperties) ContentType() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.contentType
}

// Language returns the language.
func (cp *CoreProperties) Language() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.language
}

// --- Setters ---

// SetTitle sets the title.
func (cp *CoreProperties) SetTitle(title string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.title = title
}

// SetCreator sets the creator.
func (cp *CoreProperties) SetCreator(creator string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.creator = creator
}

// SetSubject sets the subject.
func (cp *CoreProperties) SetSubject(subject string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.subject = subject
}

// SetDescription sets the description.
func (cp *CoreProperties) SetDescription(description string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.description = description
}

// SetKeywords sets the keywords.
func (cp *CoreProperties) SetKeywords(keywords string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.keywords = keywords
}

// SetCreated sets the created timestamp.
func (cp *CoreProperties) SetCreated(created string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.created = created
}

// SetModified sets the modified timestamp.
func (cp *CoreProperties) SetModified(modified string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.modified = modified
}

// SetLastModifiedBy sets the name of the last modifier.
func (cp *CoreProperties) SetLastModifiedBy(lastModifiedBy string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.lastModifiedBy = lastModifiedBy
}

// SetRevision sets the revision number.
func (cp *CoreProperties) SetRevision(revision string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.revision = revision
}

// SetCategory sets the category.
func (cp *CoreProperties) SetCategory(category string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.category = category
}

// SetContentType sets the content type.
func (cp *CoreProperties) SetContentType(contentType string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.contentType = contentType
}

// SetLanguage sets the language.
func (cp *CoreProperties) SetLanguage(language string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.language = language
}

// ===== XML serialisation =====

// XCoreProperties is the XML representation of core properties.
type XCoreProperties struct {
	XMLName        xml.Name   `xml:"coreProperties"`
	XmlnsDc        string     `xml:"xmlns:dc,attr"`
	XmlnsDcterms   string     `xml:"xmlns:dcterms,attr"`
	XmlnsDcmitype  string     `xml:"xmlns:dcmitype,attr"`
	XmlnsXsi       string     `xml:"xmlns:xsi,attr"`
	XmlnsCore      string     `xml:"xmlns,attr"`
	Title          string     `xml:"dc:title"`
	Creator        string     `xml:"dc:creator"`
	Subject        string     `xml:"dc:subject"`
	Description    string     `xml:"dc:description"`
	Keywords       *XKeywords `xml:"cp:keywords"`
	Created        *XDate     `xml:"dcterms:created"`
	Modified       *XDate     `xml:"dcterms:modified"`
	LastModifiedBy string     `xml:"cp:lastModifiedBy"`
	Revision       string     `xml:"cp:revision"`
	Category       string     `xml:"cp:category"`
	ContentType    string     `xml:"cp:contentType"`
}

// XKeywords is the XML representation of the keywords element.
type XKeywords struct {
	Value string `xml:",chardata"`
}

// XDate is the XML representation of a date element.
type XDate struct {
	Type  string `xml:"xsi:type,attr"`
	Value string `xml:",chardata"`
}

// FromXML parses core properties from XML data.
func (cp *CoreProperties) FromXML(data []byte) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	var xcp XCoreProperties
	if err := xml.Unmarshal(data, &xcp); err != nil {
		return fmt.Errorf("failed to unmarshal core properties: %w", err)
	}

	cp.title = xcp.Title
	cp.creator = xcp.Creator
	cp.subject = xcp.Subject
	cp.description = xcp.Description
	if xcp.Keywords != nil {
		cp.keywords = xcp.Keywords.Value
	}
	if xcp.Created != nil {
		cp.created = xcp.Created.Value
	}
	if xcp.Modified != nil {
		cp.modified = xcp.Modified.Value
	}
	cp.lastModifiedBy = xcp.LastModifiedBy
	cp.revision = xcp.Revision
	cp.category = xcp.Category
	cp.contentType = xcp.ContentType

	return nil
}

// ToXML serialises the core properties to XML.
func (cp *CoreProperties) ToXML() ([]byte, error) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	xcp := XCoreProperties{
		XmlnsDc:        "http://purl.org/dc/elements/1.1/",
		XmlnsDcterms:   "http://purl.org/dc/terms/",
		XmlnsDcmitype:  "http://purl.org/dc/dcmitype/",
		XmlnsXsi:       "http://www.w3.org/2001/XMLSchema-instance",
		XmlnsCore:      "http://schemas.openxmlformats.org/package/2006/metadata/core-properties",
		Title:          cp.title,
		Creator:        cp.creator,
		Subject:        cp.subject,
		Description:    cp.description,
		LastModifiedBy: cp.lastModifiedBy,
		Revision:       cp.revision,
		Category:       cp.category,
		ContentType:    cp.contentType,
	}

	if cp.keywords != "" {
		xcp.Keywords = &XKeywords{Value: cp.keywords}
	}
	if cp.created != "" {
		xcp.Created = &XDate{Type: "dcterms:W3CDTF", Value: cp.created}
	}
	if cp.modified != "" {
		xcp.Modified = &XDate{Type: "dcterms:W3CDTF", Value: cp.modified}
	}

	output, err := xml.MarshalIndent(xcp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal core properties: %w", err)
	}

	return append([]byte(XMLDeclaration), output...), nil
}
