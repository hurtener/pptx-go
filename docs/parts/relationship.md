# Relationship Module — Interface Documentation

> Corresponds to `*.rels` files; OpenXML relationship definitions

---

## Overview

This module provides two approaches to relationship management:

1. **Simple XML structures** (`XMLRelationships`/`XMLRelationship`): Lightweight XML serialization/deserialization.
2. **OPC relationship system** (`opc.Relationships`): Full package relationship management with path resolution and thread-safe ID allocation.

> **Related documentation**: For relationship path resolution and complete OPC relationship management, see [OPC Relationship Resolution](../opc/relationship_resolution.md).

---

## Overview

Example relationship file locations:
- Package level: `/_rels/.rels`
- Slide: `/ppt/slides/_rels/slide1.xml.rels`
- Master: `/ppt/slideMasters/_rels/slideMaster1.xml.rels`

Namespace: `http://schemas.openxmlformats.org/package/2006/relationships`

---

## Constant Definitions

### Namespace

```go
NamespaceRelationships = "http://schemas.openxmlformats.org/package/2006/relationships"
```

### Target Modes

```go
TargetModeInternal = "Internal"  // Internal target
TargetModeExternal = "External"  // External target
```

### Common Relationship Types

```go
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
```

---

## Data Structures

### XMLRelationships

Relationship collection root node, corresponding to XML: `<Relationships xmlns="...">...</Relationships>`

| Field | Type | Description |
|------|------|------|
| `Xmlns` | `string` | Namespace |
| `Relationships` | `[]XMLRelationship` | List of relationships |

### XMLRelationship

Single relationship, corresponding to XML: `<Relationship Id="rId1" Type="..." Target="..."/>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `ID` | `Id` | `string` | Relationship ID (e.g. `rId1`, `rId2`) |
| `Type` | `Type` | `string` | Relationship type URI |
| `Target` | `Target` | `string` | Target path (relative or absolute) |
| `TargetMode` | `TargetMode` | `string` | `Internal` (default) or `External` |

#### Methods

```go
func (r *XMLRelationship) IsExternal() bool
```

Returns true if this is an external relationship.

---

## Constructors

### NewXMLRelationships

```go
func NewXMLRelationships() *XMLRelationships
```

Creates a relationship collection with the default namespace.

### NewXMLRelationship

```go
func NewXMLRelationship(id, relType, target string) XMLRelationship
```

Creates a single relationship.

### NewXMLRelationshipExternal

```go
func NewXMLRelationshipExternal(id, relType, target string) XMLRelationship
```

Creates an external relationship (`TargetMode=External`).

---

## XMLRelationships Methods

### Add

```go
func (rs *XMLRelationships) Add(rel XMLRelationship)
```

Adds a relationship to the collection.

### AddNew

```go
func (rs *XMLRelationships) AddNew(id, relType, target string)
```

Creates and adds a new relationship.

### GetByID

```go
func (rs *XMLRelationships) GetByID(id string) *XMLRelationship
```

Retrieves a relationship by ID.

### GetByType

```go
func (rs *XMLRelationships) GetByType(relType string) []XMLRelationship
```

Retrieves all relationships of the specified type.

### GetByTarget

```go
func (rs *XMLRelationships) GetByTarget(target string) *XMLRelationship
```

Retrieves a relationship by target path.

### GetByType

```go
func (rs *XMLRelationships) GetByType(relType string) []XMLRelationship
```

Retrieves all relationships of the specified type.

### Count

```go
func (rs *XMLRelationships) Count() int
```

Returns the number of relationships.

---

## XML Serialization / Deserialization

### ToXML

```go
func (rs *XMLRelationships) ToXML() ([]byte, error)
```

Serializes the relationship collection to XML bytes.

### ParseRelationships

```go
func ParseRelationships(data []byte) (*XMLRelationships, error)
```

Parses a relationship collection from XML bytes.
