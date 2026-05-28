package opc

import (
	"encoding/xml"
	"fmt"
	"io"
	"sync"
)

// Part represents a single part within a package.
type Part struct {
	uri           *PackURI
	contentType   string
	blob          []byte // Exclusively owned data copy (mutable).
	sharedBlob    []byte // Shared read-only data (zero-copy).
	relationships *Relationships
	dirty         bool // Whether the part has been modified.
	immutable     bool // Whether the part is an immutable resource (uses sharedBlob).
	mu            sync.RWMutex
}

// NewPart creates a new part.
func NewPart(uri *PackURI, contentType string, blob []byte) *Part {
	return &Part{
		uri:           uri,
		contentType:   contentType,
		blob:          blob,
		relationships: NewRelationships(uri),
		dirty:         true,
		immutable:     false,
	}
}

// NewSharedPart creates a part that shares its data (zero-copy, for immutable resources).
// The caller must guarantee that sharedBlob will not be modified for the lifetime of the Part.
func NewSharedPart(uri *PackURI, contentType string, sharedBlob []byte) *Part {
	return &Part{
		uri:           uri,
		contentType:   contentType,
		sharedBlob:    sharedBlob,
		relationships: NewRelationships(uri),
		dirty:         false,
		immutable:     true,
	}
}

// NewPartFromReader creates a part by reading all content from r.
func NewPartFromReader(uri *PackURI, contentType string, r io.Reader) (*Part, error) {
	blob, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read part content: %w", err)
	}
	return NewPart(uri, contentType, blob), nil
}

// PartURI returns the part URI.
func (p *Part) PartURI() *PackURI {
	return p.uri
}

// ContentType returns the content type.
func (p *Part) ContentType() string {
	return p.contentType
}

// SetContentType sets the content type.
func (p *Part) SetContentType(ct string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.contentType = ct
	p.dirty = true
}

// Blob returns the raw content.
// For immutable resources it returns the shared read-only slice; otherwise it returns the owned copy.
func (p *Part) Blob() []byte {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.immutable && p.sharedBlob != nil {
		return p.sharedBlob
	}
	return p.blob
}

// SetBlob sets the content, triggering a copy-on-write by releasing the shared reference.
func (p *Part) SetBlob(blob []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sharedBlob = nil  // Release shared reference.
	p.immutable = false // Mark as mutable.
	p.blob = blob
	p.dirty = true
}

// SetBlobFromReader sets the content from a Reader.
func (p *Part) SetBlobFromReader(r io.Reader) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	blob, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read content: %w", err)
	}
	p.blob = blob
	p.dirty = true
	return nil
}

// Reader returns an io.Reader over the part content.
func (p *Part) Reader() io.Reader {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return NewBytesReader(p.blob)
}

// Size returns the content size in bytes.
func (p *Part) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.blob)
}

// Relationships returns the relationship collection for this part.
func (p *Part) Relationships() *Relationships {
	return p.relationships
}

// AddRelationship adds a relationship to this part.
func (p *Part) AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error) {
	rel, err := p.relationships.AddNew(relType, targetURI, isExternal)
	if err != nil {
		return nil, err
	}
	p.dirty = true
	return rel, nil
}

// RemoveRelationship removes a relationship from this part.
func (p *Part) RemoveRelationship(rID string) error {
	err := p.relationships.Remove(rID)
	if err != nil {
		return err
	}
	p.dirty = true
	return nil
}

// GetRelatedPart returns the target URI of the relationship with the given rID.
// A Package context is required to resolve the URI to an actual Part.
func (p *Part) GetRelatedPart(rID string) *PackURI {
	rel := p.relationships.Get(rID)
	if rel == nil {
		return nil
	}
	return rel.TargetURI()
}

// IsDirty reports whether the part has been modified.
func (p *Part) IsDirty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dirty
}

// SetDirty sets the dirty flag.
func (p *Part) SetDirty(dirty bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dirty = dirty
}

// LoadRelationships loads relationships from XML data.
func (p *Part) LoadRelationships(data []byte) error {
	return p.relationships.FromXML(data)
}

// RelationshipsBlob returns the serialised XML of the relationships.
func (p *Part) RelationshipsBlob() ([]byte, error) {
	if p.relationships.Count() == 0 {
		return nil, nil
	}
	return p.relationships.ToXML()
}

// HasRelationships reports whether the part has any relationships.
func (p *Part) HasRelationships() bool {
	return p.relationships.Count() > 0
}

// RelationshipsURI returns the URI of the relationships file for this part.
func (p *Part) RelationshipsURI() *PackURI {
	return p.uri.RelationshipsURI()
}

// Clone returns a deep copy of the part (suitable for mutable content).
func (p *Part) Clone() *Part {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var blobCopy []byte
	if p.blob != nil {
		blobCopy = make([]byte, len(p.blob))
		copy(blobCopy, p.blob)
	} else if p.sharedBlob != nil {
		// If the original was shared, a deep copy produces an independent slice.
		blobCopy = make([]byte, len(p.sharedBlob))
		copy(blobCopy, p.sharedBlob)
	}

	return &Part{
		uri:           p.uri.Clone(),
		contentType:   p.contentType,
		blob:          blobCopy,
		relationships: p.relationships.Clone(),
		dirty:         p.dirty,
		immutable:     false, // Clone becomes mutable.
	}
}

// CloneShared returns a zero-copy clone of the part (for immutable resources).
// The caller must ensure the part's content will never be modified.
func (p *Part) CloneShared() *Part {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Determine the data source to share.
	var sharedData []byte
	if p.sharedBlob != nil {
		sharedData = p.sharedBlob
	} else {
		sharedData = p.blob
	}

	return &Part{
		uri:           p.uri, // PackURI is immutable — safe to share.
		contentType:   p.contentType,
		sharedBlob:    sharedData,      // zero-copy!
		relationships: p.relationships, // Shared (immutable resources typically have no relationships).
		dirty:         false,
		immutable:     true,
	}
}

// IsImmutable reports whether the part is an immutable resource.
func (p *Part) IsImmutable() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.immutable
}

// SetImmutable marks the part as an immutable resource.
func (p *Part) SetImmutable(immutable bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.immutable = immutable
}

// UnmarshalBlob unmarshals the part's blob as XML into v.
func (p *Part) UnmarshalBlob(v interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return xml.Unmarshal(p.blob, v)
}

// MarshalToBlob marshals v as XML and stores the result in the blob.
func (p *Part) MarshalToBlob(v interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := xml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}
	p.blob = data
	p.dirty = true
	return nil
}

// BytesReader is a simple io.Reader implementation over a byte slice.
type BytesReader struct {
	data []byte
	pos  int
}

// NewBytesReader creates a new BytesReader.
func NewBytesReader(data []byte) *BytesReader {
	return &BytesReader{data: data}
}

// Read implements io.Reader.
func (r *BytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// PartCollection is an ordered collection of parts.
type PartCollection struct {
	parts map[string]*Part
	order []string // Preserves insertion order.
	mu    sync.RWMutex
}

// NewPartCollection creates a new, empty PartCollection.
func NewPartCollection() *PartCollection {
	return &PartCollection{
		parts: make(map[string]*Part),
		order: make([]string, 0),
	}
}

// Add adds a part to the collection.
func (c *PartCollection) Add(part *Part) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	uri := part.PartURI().URI()
	if _, exists := c.parts[uri]; exists {
		return fmt.Errorf("part with URI %s already exists", uri)
	}

	c.parts[uri] = part
	c.order = append(c.order, uri)
	return nil
}

// Get returns the part with the given URI.
func (c *PartCollection) Get(uri *PackURI) *Part {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parts[uri.URI()]
}

// GetByStr returns the part with the given string URI.
func (c *PartCollection) GetByStr(uri string) *Part {
	return c.Get(NewPackURI(uri))
}

// Remove removes the part with the given URI.
func (c *PartCollection) Remove(uri *PackURI) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := uri.URI()
	if _, exists := c.parts[key]; !exists {
		return fmt.Errorf("part with URI %s not found", key)
	}

	delete(c.parts, key)

	// Remove from the order slice.
	for i, u := range c.order {
		if u == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	return nil
}

// Contains reports whether a part with the given URI exists.
func (c *PartCollection) Contains(uri *PackURI) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.parts[uri.URI()]
	return exists
}

// All returns all parts in insertion order.
func (c *PartCollection) All() []*Part {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Part, 0, len(c.order))
	for _, uri := range c.order {
		result = append(result, c.parts[uri])
	}
	return result
}

// Count returns the number of parts.
func (c *PartCollection) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.parts)
}

// URIs returns all part URIs in insertion order.
func (c *PartCollection) URIs() []*PackURI {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*PackURI, 0, len(c.order))
	for _, uri := range c.order {
		result = append(result, NewPackURI(uri))
	}
	return result
}

// GetByType returns all parts with the given content type.
func (c *PartCollection) GetByType(contentType string) []*Part {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*Part
	for _, uri := range c.order {
		if part := c.parts[uri]; part.ContentType() == contentType {
			result = append(result, part)
		}
	}
	return result
}

// Clear removes all parts from the collection.
func (c *PartCollection) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parts = make(map[string]*Part)
	c.order = make([]string, 0)
}

// DirtyParts returns all modified parts.
func (c *PartCollection) DirtyParts() []*Part {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*Part
	for _, uri := range c.order {
		if part := c.parts[uri]; part.IsDirty() {
			result = append(result, part)
		}
	}
	return result
}

// PartFactory is the interface for creating parts.
type PartFactory interface {
	// CreatePart creates a part.
	CreatePart(uri *PackURI, contentType string, blob []byte) (*Part, error)
}

// DefaultPartFactory is the default part factory.
type DefaultPartFactory struct{}

// CreatePart implements PartFactory.
func (f *DefaultPartFactory) CreatePart(uri *PackURI, contentType string, blob []byte) (*Part, error) {
	return NewPart(uri, contentType, blob), nil
}
