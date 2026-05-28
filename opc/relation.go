package opc

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// Relationship represents a relationship between two parts.
type Relationship struct {
	rID        string   // Relationship ID (rId1, rId2, …)
	relType    string   // Relationship type
	target     *PackURI // Target URI
	targetMode string   // Internal or External
	isExternal bool     // Whether the target is external
	source     *PackURI // Source part URI (used for resolving relative paths)
}

// NewRelationship creates a new relationship.
func NewRelationship(rID, relType, targetURI string, isExternal bool, source *PackURI) *Relationship {
	rel := &Relationship{
		rID:        rID,
		relType:    relType,
		isExternal: isExternal,
		source:     source,
	}

	if isExternal {
		rel.targetMode = "External"
		rel.target = &PackURI{uri: targetURI}
	} else {
		rel.targetMode = "Internal"
		// For internal relationships, resolve relative paths using the source URI's directory.
		if source != nil && !strings.HasPrefix(targetURI, "/") {
			// Relative path: resolve against the source's directory.
			rel.target = source.Join(targetURI)
		} else {
			// Absolute path: use directly.
			rel.target = NewPackURI(targetURI)
		}
	}

	return rel
}

// RID returns the relationship ID.
func (r *Relationship) RID() string {
	return r.rID
}

// Type returns the relationship type.
func (r *Relationship) Type() string {
	return r.relType
}

// TargetURI returns the target URI.
func (r *Relationship) TargetURI() *PackURI {
	return r.target
}

// TargetRef returns the target reference (relative or absolute).
// If a source part is set, returns the relative path from the source to the target.
func (r *Relationship) TargetRef() string {
	if r.isExternal {
		return r.target.URI()
	}
	if r.source != nil {
		return r.target.RelPathFrom(r.source)
	}
	return r.target.URI()
}

// IsExternal reports whether this is an external relationship.
func (r *Relationship) IsExternal() bool {
	return r.isExternal
}

// TargetMode returns the target mode string.
func (r *Relationship) TargetMode() string {
	return r.targetMode
}

// SourceURI returns the source URI.
func (r *Relationship) SourceURI() *PackURI {
	return r.source
}

// SetSource sets the source URI.
func (r *Relationship) SetSource(source *PackURI) {
	r.source = source
}

// Equals reports whether two relationships are equal.
func (r *Relationship) Equals(other *Relationship) bool {
	if other == nil {
		return false
	}
	return r.rID == other.rID && r.relType == other.relType && r.target.Equals(other.target)
}

// Relationships is an ordered collection of relationships.
type Relationships struct {
	relationships map[string]*Relationship
	order         []string // Preserves insertion order.
	mu            sync.RWMutex
	sourceURI     *PackURI     // The source part this collection belongs to.
	rIDCounter    atomic.Int32 // Atomic counter for generating rIDs.
}

// NewRelationships creates a new, empty Relationships collection.
func NewRelationships(sourceURI *PackURI) *Relationships {
	return &Relationships{
		relationships: make(map[string]*Relationship),
		order:         make([]string, 0),
		sourceURI:     sourceURI,
	}
}

// Add adds a relationship to the collection.
func (rs *Relationships) Add(rel *Relationship) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if _, exists := rs.relationships[rel.RID()]; exists {
		return fmt.Errorf("relationship with rID %s already exists", rel.RID())
	}

	// Set the source URI if not already set.
	if rel.SourceURI() == nil && rs.sourceURI != nil {
		rel.SetSource(rs.sourceURI)
	}

	rs.relationships[rel.RID()] = rel
	rs.order = append(rs.order, rel.RID())
	return nil
}

// AddNew creates and adds a new relationship, allocating an ID atomically.
func (rs *Relationships) AddNew(relType, targetURI string, isExternal bool) (*Relationship, error) {
	rID := rs.allocateRID()
	rel := NewRelationship(rID, relType, targetURI, isExternal, rs.sourceURI)
	err := rs.Add(rel)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

// Get returns the relationship with the given rID.
func (rs *Relationships) Get(rID string) *Relationship {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.relationships[rID]
}

// GetByType returns all relationships of the given type.
func (rs *Relationships) GetByType(relType string) []*Relationship {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	var result []*Relationship
	for _, rID := range rs.order {
		if rel := rs.relationships[rID]; rel.Type() == relType {
			result = append(result, rel)
		}
	}
	return result
}

// GetByTarget returns the relationship whose target matches the given URI.
func (rs *Relationships) GetByTarget(targetURI *PackURI) *Relationship {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	for _, rel := range rs.relationships {
		if rel.TargetURI().Equals(targetURI) {
			return rel
		}
	}
	return nil
}

// Remove removes the relationship with the given rID.
func (rs *Relationships) Remove(rID string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if _, exists := rs.relationships[rID]; !exists {
		return fmt.Errorf("relationship with rID %s not found", rID)
	}

	delete(rs.relationships, rID)

	// Remove from the order slice.
	for i, id := range rs.order {
		if id == rID {
			rs.order = append(rs.order[:i], rs.order[i+1:]...)
			break
		}
	}
	return nil
}

// Contains reports whether a relationship with the given rID exists.
func (rs *Relationships) Contains(rID string) bool {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	_, exists := rs.relationships[rID]
	return exists
}

// All returns all relationships in insertion order.
func (rs *Relationships) All() []*Relationship {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	result := make([]*Relationship, 0, len(rs.order))
	for _, rID := range rs.order {
		result = append(result, rs.relationships[rID])
	}
	return result
}

// Count returns the number of relationships.
func (rs *Relationships) Count() int {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return len(rs.relationships)
}

// NextRID returns the next relationship ID without consuming it (idempotent preview).
// Multiple calls return the same value until AddNew actually uses it.
func (rs *Relationships) NextRID() string {
	// Peek at the current value + 1 without incrementing the counter.
	nextID := rs.rIDCounter.Load() + 1
	return fmt.Sprintf("rId%d", nextID)
}

// allocateRID allocates and returns the next relationship ID.
// Used by AddNew and similar methods that need an actual ID.
// The atomic Add returns the new (post-increment) value, e.g. counter=0 → Add(1) returns 1.
func (rs *Relationships) allocateRID() string {
	return fmt.Sprintf("rId%d", rs.rIDCounter.Add(1))
}

// InitRIDCounter initialises the rID counter from the existing relationships,
// ensuring newly allocated IDs do not clash with existing ones.
// Call this after loading relationships from XML.
func (rs *Relationships) InitRIDCounter() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.initRIDCounterLocked()
}

// SetSourceURI sets the source URI and propagates it to all relationships.
func (rs *Relationships) SetSourceURI(sourceURI *PackURI) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.sourceURI = sourceURI
	// Update the source on all existing relationships.
	for _, rel := range rs.relationships {
		rel.SetSource(sourceURI)
	}
}

// SourceURI returns the source URI.
func (rs *Relationships) SourceURI() *PackURI {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.sourceURI
}

// Clone returns a deep copy of the relationship collection.
func (rs *Relationships) Clone() *Relationships {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	newRs := NewRelationships(rs.sourceURI)
	for _, rID := range rs.order {
		rel := rs.relationships[rID]
		newRel := NewRelationship(rel.RID(), rel.Type(), rel.TargetURI().URI(), rel.IsExternal(), rel.SourceURI())
		newRs.relationships[rID] = newRel
		newRs.order = append(newRs.order, rID)
	}
	// Copy the current value of the atomic counter.
	newRs.rIDCounter.Store(rs.rIDCounter.Load())
	return newRs
}

// XML structs for serialising relationships.

// XRelationships is the XML root element for relationships.
type XRelationships struct {
	XMLName       xml.Name        `xml:"Relationships"`
	Xmlns         string          `xml:"xmlns,attr"`
	Relationships []XRelationship `xml:"Relationship"`
}

// XRelationship is the XML element for a single relationship.
type XRelationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr,omitempty"`
}

// MarshalXML implements xml.Marshaler.
func (rs *Relationships) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	xrels := XRelationships{
		Xmlns: NamespaceRelationships,
	}

	for _, rID := range rs.order {
		rel := rs.relationships[rID]
		xrel := XRelationship{
			ID:     rel.RID(),
			Type:   rel.Type(),
			Target: rel.TargetRef(),
		}
		if rel.IsExternal() {
			xrel.TargetMode = "External"
		}
		xrels.Relationships = append(xrels.Relationships, xrel)
	}

	return e.Encode(xrels)
}

// UnmarshalXML implements xml.Unmarshaler.
func (rs *Relationships) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	var xrels XRelationships
	if err := d.DecodeElement(&xrels, &start); err != nil {
		return err
	}

	rs.relationships = make(map[string]*Relationship)
	rs.order = make([]string, 0)

	for _, xrel := range xrels.Relationships {
		isExternal := xrel.TargetMode == "External"
		rel := NewRelationship(xrel.ID, xrel.Type, xrel.Target, isExternal, rs.sourceURI)
		rs.relationships[xrel.ID] = rel
		rs.order = append(rs.order, xrel.ID)
	}

	// Initialise the counter so new IDs do not clash with existing ones.
	rs.initRIDCounterLocked()

	return nil
}

// initRIDCounterLocked initialises the rID counter while the lock is already held.
func (rs *Relationships) initRIDCounterLocked() {
	maxNum := int32(0)
	for rID := range rs.relationships {
		if strings.HasPrefix(rID, "rId") {
			var num int
			_, err := fmt.Sscanf(rID, "rId%d", &num)
			if err == nil && int32(num) > maxNum {
				maxNum = int32(num)
			}
		}
	}
	rs.rIDCounter.Store(maxNum)
}

// ToXML serialises the relationship collection to XML.
func (rs *Relationships) ToXML() ([]byte, error) {
	output, err := xml.Marshal(rs)
	if err != nil {
		return nil, err
	}
	return append([]byte(XMLDeclaration), output...), nil
}

// FromXML parses the relationship collection from XML data.
func (rs *Relationships) FromXML(data []byte) error {
	return xml.Unmarshal(data, rs)
}

// Relatable is the interface for parts that can hold relationships.
type Relatable interface {
	// PartURI returns the part's URI.
	PartURI() *PackURI
	// Relationships returns the part's relationship collection.
	Relationships() *Relationships
	// AddRelationship adds a relationship.
	AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error)
}

// RelTypeCollection is a collection of relationships grouped by type.
type RelTypeCollection struct {
	types map[string][]*Relationship
}

// NewRelTypeCollection creates a new, empty RelTypeCollection.
func NewRelTypeCollection() *RelTypeCollection {
	return &RelTypeCollection{
		types: make(map[string][]*Relationship),
	}
}

// Add adds a relationship to the type collection.
func (c *RelTypeCollection) Add(rel *Relationship) {
	c.types[rel.Type()] = append(c.types[rel.Type()], rel)
}

// GetByType returns all relationships of the given type.
func (c *RelTypeCollection) GetByType(relType string) []*Relationship {
	return c.types[relType]
}

// Types returns all distinct relationship types in sorted order.
func (c *RelTypeCollection) Types() []string {
	types := make([]string, 0, len(c.types))
	for t := range c.types {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// ParseRelationshipsFromXML parses a Relationships collection from XML data.
func ParseRelationshipsFromXML(data []byte, sourceURI *PackURI) (*Relationships, error) {
	rels := NewRelationships(sourceURI)
	if err := rels.FromXML(data); err != nil {
		return nil, fmt.Errorf("failed to parse relationships: %w", err)
	}
	return rels, nil
}
