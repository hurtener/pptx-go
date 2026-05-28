# XML Utility Functions

This module provides XML processing utility functions, primarily for resolving compatibility issues between Go's standard library `encoding/xml` and namespace-prefixed XML.

## Overview

Office Open XML (OOXML) files use XML elements and attributes with namespace prefixes, for example:

```xml
<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:sldIdLst>
    <p:sldId id="256" r:id="rId2"/>
  </p:sldIdLst>
</p:presentation>
```

Go's `xml.Unmarshal` cannot process this format correctly because it:
1. Cannot match element names with prefixes (e.g. `<p:presentation>`)
2. Cannot match attribute names with prefixes (e.g. `r:id`)

## Core Function

### StripNamespacePrefixes

```go
func StripNamespacePrefixes(data []byte) ([]byte, error)
```

Processes XML data by removing namespace prefixes to make it compatible with Go's `xml.Unmarshal`.

**Transformation rules:**

| Original XML | After transformation |
|----------|--------|
| `<p:presentation>` | `<presentation>` |
| `<a:solidFill>` | `<solidFill>` |
| `r:id="rId1"` | `rid="rId1"` |
| `xmlns:p="..."` | (removed) |

**Usage example:**

```go
// Raw XML data read from a PPTX file
rawXML := slidePart.Blob()

// Strip namespace prefixes
cleanXML, err := parts.StripNamespacePrefixes(rawXML)
if err != nil {
    return err
}

// Can now be parsed correctly
var slide XSlide
if err := xml.Unmarshal(cleanXML, &slide); err != nil {
    return err
}
```

## XML Struct Tag Conventions

After applying `StripNamespacePrefixes`, struct tags should follow these conventions:

### Element Names
- No prefix: `xml:"presentation"` not `xml:"p:presentation"`
- No namespace URI: `xml:"spTree"` not `xml:"http://... spTree"`

### Attribute Names
- Merge prefix into attribute name: `xml:"rid,attr"` not `xml:"r:id,attr"`
- Examples: `r:id` → `rid`, `r:embed` → `rembed`

```go
// Correct tag format
type XSldId struct {
    Id  uint32 `xml:"id,attr"`    // matches id="256"
    RId string `xml:"rid,attr"`   // matches the transformed rid="rId2"
}

// Incorrect tag format (causes parsing failure)
type XSldId struct {
    Id  uint32 `xml:"id,attr"`
    RId string `xml:"r:id,attr"`  // cannot match rid
}
```

## Constants

### XMLDeclaration

```go
const XMLDeclaration = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
```

Standard XML declaration header for all XML files in an OPC package.

## Internal Implementation

### Processing Flow

1. **Collect namespace mappings**: Iterate over all `xmlns` attributes and build a URI → prefix mapping.
2. **Transform element names**: Remove prefixes from element names (e.g. `p:presentation` → `presentation`).
3. **Transform attribute names**: Merge prefixes into attribute names (e.g. `r:id` → `rid`).
4. **Remove xmlns declarations**: Delete all namespace declaration attributes.

### Code Example

```go
func stripNamespacePrefixes(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    decoder := xml.NewDecoder(bytes.NewReader(data))
    nsToPrefix := make(map[string]string)

    for {
        token, err := decoder.Token()
        if err == io.EOF {
            break
        }
        // ... handle each token type
    }
    return buf.Bytes(), nil
}
```

## Notes

1. **Performance**: This function copies the entire XML data; it may have memory overhead for large files.
2. **Information retained**: Namespace URI information is discarded, but prefix information is preserved in attribute names.
3. **One-way transformation**: This function is intended for reading (deserialization) only; full namespaces are used when writing.
