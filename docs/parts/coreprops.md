# Core Properties — Interface Documentation

> Corresponds to `/docProps/core.xml`, based on the Dublin Core metadata standard

## Namespaces

| Constant | Value |
|------|-----|
| `NamespaceCoreProperties` | `http://schemas.openxmlformats.org/package/2006/metadata/core-properties` |
| `NamespaceDublinCore` | `http://purl.org/dc/elements/1.1/` |
| `NamespaceDublinCoreTerms` | `http://purl.org/dc/terms/` |
| `NamespaceXMLSchema` | `http://www.w3.org/2001/XMLSchema-instance` |

## Structs

### XMLCoreProperties

Core properties XML struct corresponding to the `core.xml` file.

| Field | XML Path | Type | Description |
|------|----------|------|------|
| `Title` | `dc:title` | `string` | Document title |
| `Creator` | `dc:creator` | `string` | Creator |
| `Subject` | `dc:subject` | `string` | Subject |
| `Description` | `dc:description` | `string` | Description |
| `Created` | `dcterms:created` | `*XMLW3CDTFDate` | Creation time |
| `Modified` | `dcterms:modified` | `*XMLW3CDTFDate` | Modification time |
| `Keywords` | `cp:keywords` | `string` | Keywords |
| `LastModifiedBy` | `cp:lastModifiedBy` | `string` | Last modified by |
| `Revision` | `cp:revision` | `string` | Revision number |
| `Category` | `cp:category` | `string` | Category |
| `ContentType` | `cp:contentType` | `string` | Content type |
| `Version` | `cp:version` | `string` | Version |
| `Identifier` | `cp:identifier` | `string` | Identifier |
| `Language` | `dc:language` | `string` | Language |

### XMLW3CDTFDate

W3CDTF-format date element, corresponding to `<dcterms:created xsi:type="dcterms:W3CDTF">`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Type` | `xsi:type` | `string` | Type identifier, fixed as `dcterms:W3CDTF` |
| `Value` | chardata | `string` | Date value, format: `YYYY-MM-DDThh:mm:ssZ` |

## Constructors

### NewXMLCoreProperties

```go
func NewXMLCoreProperties() *XMLCoreProperties
```

Creates a core properties struct with default namespaces.

## Helper Methods

### SetCreated

```go
func (cp *XMLCoreProperties) SetCreated(value string)
```

Sets the creation time in W3CDTF format.

### SetModified

```go
func (cp *XMLCoreProperties) SetModified(value string)
```

Sets the modification time in W3CDTF format.

### GetCreated

```go
func (cp *XMLCoreProperties) GetCreated() string
```

Returns the creation time value.

### GetModified

```go
func (cp *XMLCoreProperties) GetModified() string
```

Returns the modification time value.

### ToXML

```go
func (cp *XMLCoreProperties) ToXML() ([]byte, error)
```

Serializes the core properties to XML bytes.

### ParseCoreProperties

```go
func ParseCoreProperties(data []byte) (*XMLCoreProperties, error)
```

Parses core properties from XML bytes.

### ParseCoreProps

```go
func ParseCoreProps(data []byte) (*XMLCoreProperties, error)
```

Shorthand alias for `ParseCoreProperties`.
