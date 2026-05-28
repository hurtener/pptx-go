# Relationship Resolution

Parts within an OPC package reference one another through relationships. This document describes the relationship resolution mechanism and how to use it.

## Overview

In an OPC package, relationships define references between parts:

```
presentation.xml ──relationship──> slides/slide1.xml
                   │
                   └──relationship──> slideMasters/slideMaster1.xml
```

Relationships are stored in `.rels` files under the `_rels` directory. Target paths may be either:
- **Absolute paths**: beginning with `/`, e.g. `/ppt/slides/slide1.xml`
- **Relative paths**: relative to the source part, e.g. `slides/slide1.xml`

## Relative Path Resolution

### Problem

When a relationship target uses a relative path (e.g. `slides/slide1.xml`), it must be resolved to an absolute path based on the location of the source part.

```xml
<!-- presentation.xml.rels -->
<Relationship Id="rId2" Type="..." Target="slides/slide1.xml"/>
```

For `/ppt/presentation.xml`, the target should resolve to `/ppt/slides/slide1.xml`.

### Solution

The `NewRelationship` function handles relative path resolution automatically:

```go
func NewRelationship(rID, relType, targetURI string, isExternal bool, source *PackURI) *Relationship {
    // ...
    if source != nil && !strings.HasPrefix(targetURI, "/") {
        // Relative path: resolve using the source's directory
        rel.target = source.Join(targetURI)
    } else {
        // Absolute path: create directly
        rel.target = NewPackURI(targetURI)
    }
    // ...
}
```

### PackURI.Join Method

The `Join` method resolves a relative path against the URI's directory:

```go
source := opc.NewPackURI("/ppt/presentation.xml")

source.Join("slides/slide1.xml")        // → /ppt/slides/slide1.xml
source.Join("../theme/theme1.xml")      // → /theme/theme1.xml
source.Join("./slides/slide1.xml")      // → /ppt/slides/slide1.xml
source.Join("/ppt/slides/slide1.xml")   // → /ppt/slides/slide1.xml (absolute path unchanged)
```

## Relationship API

### Creating Relationships

```go
// Create via the Part method
rel, err := part.AddRelationship(relType, targetURI, isExternal)

// Create directly
rel := opc.NewRelationship(rID, relType, targetURI, isExternal, sourceURI)
```

### Resolving Relationships

```go
// Resolve through Package to get the target part
targetPart := pkg.ResolveRelationship(sourcePart, relType)

// Get the relationship target URI from a Part
targetURI := sourcePart.GetRelatedPart(rID)
```

### Relationship Properties

```go
rel.RID()         // Relationship ID, e.g. "rId1"
rel.Type()        // Relationship type URI
rel.TargetURI()   // Absolute path of the target (after resolution)
rel.TargetRef()   // Relative reference for the target (used in serialization)
rel.IsExternal()  // Whether this is an external relationship
rel.SourceURI()   // Source part URI
```

## Common Relationship Types

| Source Part | Target Part | Relationship Type |
|-------------|-------------|------------------|
| Package | Presentation | `officeDocument` |
| Presentation | Slide | `slide` |
| Presentation | SlideMaster | `slideMaster` |
| Slide | SlideLayout | `slideLayout` |
| SlideLayout | SlideMaster | `slideMaster` |
| SlideMaster | Theme | `theme` |
| Slide | Image | `image` |

## Example: A Complete Relationship Chain

```go
pkg := opc.NewPackage()

// 1. Create parts
presPart, _ := pkg.CreatePart(opc.NewPackURI("/ppt/presentation.xml"), contentType, data)
slidePart, _ := pkg.CreatePart(opc.NewPackURI("/ppt/slides/slide1.xml"), contentType, slideData)

// 2. Add relationship (using relative path)
presPart.AddRelationship(
    "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide",
    "slides/slide1.xml",  // relative path
    false,
)

// 3. Resolve relationship
resolved := pkg.ResolveRelationship(presPart, "http://.../slide")
// resolved == slidePart
```

## Notes

1. **Relative path format**: use forward slashes `/`; do not use backslashes.
2. **Path resolution**: relative paths are resolved against the directory of the source part.
3. **Serialization**: `TargetRef()` returns a relative path suitable for writing to XML.
4. **External relationships**: set `isExternal` to `true` for external targets.
