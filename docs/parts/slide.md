# Slide Module — Interface Documentation

> Corresponds to `/ppt/slides/slideN.xml`; contains XML structures for slides, layouts, shapes, text, images, tables, and more

---

## Enum Types

### SlideLayoutType

Slide layout type, corresponding to `slideLayoutN.xml`.

| Constant | Value | Description |
|------|-----|------|
| `SlideLayoutBlank` | `0` | Blank layout |
| `SlideLayoutTitle` | `1` | Title layout |
| `SlideLayoutTitleAndContent` | `2` | Title and content layout |
| `SlideLayoutTwoContent` | `3` | Two-column content layout |
| `SlideLayoutComparison` | `4` | Comparison layout |
| `SlideLayoutTitleOnly` | `5` | Title-only layout |
| `SlideLayoutBlankVertical` | `6` | Blank vertical layout |
| `SlideLayoutObject` | `7` | Object layout |
| `SlideLayoutPictureAndCaption` | `8` | Picture and caption layout |

---

## Relationship Type Constants

```go
const (
    RelationshipTypeImage       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
    RelationshipTypeMedia      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/video"
    RelationshipTypeChart      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/chart"
    RelationshipTypeSlideLayout = "http://schemas.openxmlformats.org/presentationml/2006/relationships/slideLayout"
    RelationshipTypeSlideMaster = "http://schemas.openxmlformats.org/presentationml/2006/relationships/slideMaster"
    RelationshipTypeTable      = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/table"
)
```

---

## SlidePart

Slide part, corresponding to `/ppt/slides/slideN.xml`.

### Create

```go
func NewSlidePart(id int) *SlidePart
func NewSlidePartWithURI(uri *opc.PackURI) *SlidePart
```

### URI Methods

```go
func (s *SlidePart) PartURI() *opc.PackURI
func (s *SlidePart) SetURI(uri *opc.PackURI)
```

### Layout / Master Association

```go
func (s *SlidePart) LayoutRId() string
func (s *SlidePart) SetLayoutRId(rId string)
func (s *SlidePart) MasterRId() string
func (s *SlidePart) SetMasterRId(rId string)
```

### Relationship Management

```go
func (s *SlidePart) Relationships() *SlideRelationships
func (s *SlidePart) AddImage(targetURI string) string
func (s *SlidePart) AddMedia(targetURI string) string
func (s *SlidePart) AddChart(targetURI string) string
func (s *SlidePart) GetRelationshipURI(rId string) string
func (s *SlidePart) HasImage(targetURI string) bool
func (s *SlidePart) HasMedia(targetURI string) bool
func (s *SlidePart) GetImageRId(targetURI string) string
func (s *SlidePart) GetMediaRId(targetURI string) string
func (s *SlidePart) GetChartRId(targetURI string) string
func (s *SlidePart) GetOrAddPicture(x, y, cx, cy int, imageURI string) *XPicture
```

### Shape ID Management

```go
func (s *SlidePart) Allocator() *ShapeIDAllocator
func (s *SlidePart) NextShapeID() uint32
func (s *SlidePart) AllocateShapeID() uint32
func (s *SlidePart) AllocateShapeIDBatch(count int) []uint32
func (s *SlidePart) AllocateShapeIDWithOffset(offset uint32) uint32
func (s *SlidePart) PeekNextShapeID() uint32
func (s *SlidePart) CurrentShapeID() uint32
func (s *SlidePart) ResetShapeID()
func (s *SlidePart) SetShapeIDStart(startID uint32)
func (s *SlidePart) ShapeIDCount() uint32
```

### Adding Shapes

```go
func (s *SlidePart) AddShape(shape any)
func (s *SlidePart) AddTextBox(x, y, cx, cy int, text string) *XSp
func (s *SlidePart) AddAutoShape(x, y, cx, cy int, presetID string) *XSp
func (s *SlidePart) AddPicture(x, y, cx, cy int, imageRId string) *XPicture
func (s *SlidePart) AddTable(x, y, cx, cy, rows, cols int) *XGraphicFrame
func (s *SlidePart) SetTableCellText(gf *XGraphicFrame, row, col int, text string)
```

### XML Serialization

```go
func (s *SlidePart) ToXML() ([]byte, error)
func (s *SlidePart) FromXML(data []byte) error
```

> **Notes:**
> - `ToXML` uses `XMLWriterPool` for efficient serialization, producing standard OOXML output with namespace prefixes.
> - `FromXML` calls `StripNamespacePrefixes` internally to handle namespace issues. See [xmlutils.md](xmlutils.md).

---

## SlideLayoutPart

Slide layout part, corresponding to `/ppt/slideLayouts/slideLayoutN.xml`.

### Create

```go
func NewSlideLayoutPart(id int) *SlideLayoutPart
```

### Methods

```go
func (s *SlideLayoutPart) PartURI() *opc.PackURI
func (s *SlideLayoutPart) LayoutType() SlideLayoutType
func (s *SlideLayoutPart) SetLayoutType(t SlideLayoutType)
func (s *SlideLayoutPart) MasterRId() string
func (s *SlideLayoutPart) SetMasterRId(rId string)
```

---

## SlideRelationships

Page-level relationship manager; maintains rId mappings for images, charts, layouts, and more.

### Create

```go
func NewSlideRelationships() *SlideRelationships
```

### Add Relationships

```go
func (sr *SlideRelationships) AddImageRel(targetURI string) string
func (sr *SlideRelationships) AddMediaRel(targetURI string) string
func (sr *SlideRelationships) AddChartRel(targetURI string) string
func (sr *SlideRelationships) AddTableRel(targetURI string) string
```

### Query Relationships

```go
func (sr *SlideRelationships) ImageRels() map[string]string
func (sr *SlideRelationships) MediaRels() map[string]string
func (sr *SlideRelationships) ChartRels() map[string]string
func (sr *SlideRelationships) TableRels() map[string]string
func (sr *SlideRelationships) LayoutRId() string
func (sr *SlideRelationships) SetLayoutRId(rId string)
func (sr *SlideRelationships) MasterRId() string
func (sr *SlideRelationships) SetMasterRId(rId string)
func (sr *SlideRelationships) GetImageRelByURI(targetURI string) string
func (sr *SlideRelationships) GetMediaRelByURI(targetURI string) string
func (sr *SlideRelationships) RelationshipCount() int
```

### Serialization

```go
func (sr *SlideRelationships) ToRelationshipsXML() ([]byte, error)
```

---

## ShapeIDAllocator

Shape ID allocator (single-threaded use).

### Create

```go
func NewShapeIDAllocator(reservedID uint32) *ShapeIDAllocator
func NewShapeIDAllocatorWithMax(reservedID, maxID uint32) *ShapeIDAllocator
```

### Allocation Methods

```go
func (a *ShapeIDAllocator) Next() uint32                    // Allocate the next ID
func (a *ShapeIDAllocator) NextBatch(count int) []uint32    // Allocate a batch of IDs
func (a *ShapeIDAllocator) Peek() uint32                    // Peek at the next ID (without allocating)
func (a *ShapeIDAllocator) Current() uint32                 // Return the current ID
func (a *ShapeIDAllocator) Reset()                          // Reset
func (a *ShapeIDAllocator) ResetFrom(startID uint32)        // Reset from a specific ID
func (a *ShapeIDAllocator) SetReserved(reservedID uint32)   // Set the reserved starting ID
func (a *ShapeIDAllocator) Remaining() uint32               // Remaining allocatable count
func (a *ShapeIDAllocator) IsExhausted() bool               // Check whether exhausted
func (a *ShapeIDAllocator) UsedCount() uint32               // Number of IDs used
```

---

## ShapeIDAllocatorSync

Thread-safe shape ID allocator.

### Create

```go
func NewShapeIDAllocatorSync(reservedID uint32) *ShapeIDAllocatorSync
func NewShapeIDAllocatorSyncWithMax(reservedID, maxID uint32) *ShapeIDAllocatorSync
```

### Allocation Methods

```go
func (a *ShapeIDAllocatorSync) Next() uint32
func (a *ShapeIDAllocatorSync) NextBatch(count int) []uint32
func (a *ShapeIDAllocatorSync) TryNext() (uint32, bool)  // Try to allocate; returns false on failure
func (a *ShapeIDAllocatorSync) Peek() uint32
func (a *ShapeIDAllocatorSync) Reset()
func (a *ShapeIDAllocatorSync) ResetFrom(startID uint32)
```

---

## XML Struct Types

### XSpTree

Shape tree, corresponding to `<p:spTree>`.

```go
func NewXSpTree() *XSpTree
func (xst *XSpTree) WriteXML(xw *XMLWriter) error
```

### XSp

Shape, corresponding to `<p:sp>`.

```go
func (xs *XSp) WriteXML(xw *XMLWriter) error
```

### XPicture

Picture, corresponding to `<p:pic>`.

```go
func (xp *XPicture) WriteXML(xw *XMLWriter) error
```

### XGraphicFrame

Graphic frame, corresponding to `<p:graphicFrame>`.

```go
func (xgf *XGraphicFrame) WriteXML(xw *XMLWriter) error
```

### XTextBody

Text body, corresponding to `<p:txBody>`.

```go
func (xtb *XTextBody) WriteXML(xw *XMLWriter) error
```

### XTextParagraph

Text paragraph, corresponding to `<a:p>`.

```go
func (xtp *XTextParagraph) WriteXML(xw *XMLWriter) error
```

### XTextRun

Text run, corresponding to `<a:r>`.

```go
func (xtr *XTextRun) WriteXML(xw *XMLWriter) error
```

### XTable

Table, corresponding to `<a:tbl>`.

```go
func (xt *XTable) WriteXML(xw *XMLWriter) error
```

### XTableRow

Table row, corresponding to `<a:tr>`.

```go
func (xtr *XTableRow) WriteXML(xw *XMLWriter) error
```

### XTableCell

Table cell, corresponding to `<a:tc>`.

```go
func (xtc *XTableCell) WriteXML(xw *XMLWriter) error
```

### XTransform2D

2D transform, corresponding to `<a:xfrm>`.

```go
func (xt *XTransform2D) WriteXML(xw *XMLWriter) error
```

### XSlide

Slide XML structure, corresponding to `<p:sld>`.

```go
func (xs *XSlide) WriteXML(xw *XMLWriter) error
```

### XCSld

Common slide data, corresponding to `<p:cSld>`. Contains the actual slide content (shape tree, etc.).

```go
type XCSld struct {
    SpTree *XSpTree `xml:"spTree"`  // shape tree
}
```

**XML structure:**

```xml
<p:sld>
  <p:cSld>
    <p:spTree>
      <!-- shape content -->
    </p:spTree>
  </p:cSld>
  <p:clrMapOvr>
    <!-- color map override -->
  </p:clrMapOvr>
</p:sld>
```

**Deserialization example:**

```go
// Read slide XML
data := slidePart.Blob()

// Strip namespace prefixes (required — Go xml.Unmarshal does not support namespaces)
cleanData, err := parts.StripNamespacePrefixes(data)
if err != nil {
    return err
}

// Parse
var xs XSlide
if err := xml.Unmarshal(cleanData, &xs); err != nil {
    return err
}

// Access the shape tree
if xs.CSld != nil && xs.CSld.SpTree != nil {
    for _, child := range xs.CSld.SpTree.Children {
        // process child elements
    }
}
```

### XBlipFillProperties

Picture fill properties, corresponding to `<p:blipFill>`.

```go
func (xbfp *XBlipFillProperties) WriteXML(xw *XMLWriter) error
```

### XBlip

Picture reference, corresponding to `<a:blip r:embed="..."/>`.

### XBodyPr

Body properties, corresponding to `<a:bodyPr>`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Wrap` | `wrap` | `string` | Auto wrap |
| `Rotation` | `rot` | `int` | Rotation angle |
| `Vertical` | `vert` | `string` | Vertical direction |
| `Anchor` | `anchor` | `string` | Anchor position |
| `AnchorCtr` | `anchorCtr` | `bool` | Center anchor |

### XClrMap

Color map, corresponding to `<p:clrMap>`.

| Field | XML Attribute | Type |
|------|----------|------|
| `BG1` | `bg1` | `string` |
| `T1` | `t1` | `string` |
| `BG2` | `bg2` | `string` |
| `T2` | `t2` | `string` |
| `Accent1-6` | `accent1-6` | `string` |
| `HLink` | `hlink` | `string` |
| `HLink1` | `hlink1` | `string` |
| `HLink2` | `hlink2` | `string` |
| `FollClr` | `follClr` | `string` |
| `LastClr` | `lastClr` | `string` |

### XSlideRelationships

Slide relationships, corresponding to `_rels/slideN.xml.rels`.

```go
func (xsr *XSlideRelationships) WriteXML(xw *XMLWriter) error
```

---

## XMLWriter

Streaming XML write helper for efficient XML generation.

### Create

```go
func NewXMLWriter(w io.Writer) *XMLWriter
func NewXMLWriterWithIndent(w io.Writer, indentStr string) *XMLWriter
func NewXMLWriterBuffered(cap int) *XMLWriter
```

### Configuration

```go
func (xw *XMLWriter) SetAutoFlush(enable bool)
func (xw *XMLWriter) SetIndent(indentStr string)
func (xw *XMLWriter) SetUseIndent(use bool)
func (xw *XMLWriter) Reset(w io.Writer)
func (xw *XMLWriter) ResetBuffer()
```

### XML Write Methods

```go
func (xw *XMLWriter) Declaration() error
func (xw *XMLWriter) DeclarationWithEncoding(encoding string) error
func (xw *XMLWriter) StartElement(prefix, localName string) error
func (xw *XMLWriter) StartElementNS(prefix, localName, ns string) error
func (xw *XMLWriter) StartElementWithAttrs(prefix, localName string, attrs ...string) error
func (xw *XMLWriter) StartElementNSWithAttrs(prefix, localName, ns string, attrs ...string) error
func (xw *XMLWriter) StartElementRaw(prefix, localName string, attrs ...string) error
func (xw *XMLWriter) EndElement(prefix, localName string) error
func (xw *XMLWriter) EmptyElement(prefix, localName string) error
func (xw *XMLWriter) EmptyElementWithAttrs(prefix, localName string, attrs ...string) error
```

### Content Write Methods

```go
func (xw *XMLWriter) Text(content string) error
func (xw *XMLWriter) TextRaw(content string) error
func (xw *XMLWriter) CharData(data []byte) error
func (xw *XMLWriter) Comment(content string) error
func (xw *XMLWriter) CData(content string) error
func (xw *XMLWriter) ProcessingInstruction(target, data string) error
func (xw *XMLWriter) Newline() error
func (xw *XMLWriter) Raw(content string) error
```

### Indentation Control

```go
func (xw *XMLWriter) Indent()
func (xw *XMLWriter) Dedent()
func (xw *XMLWriter) WithIndent(fn func())
```

### Numeric Write Methods

```go
func (xw *XMLWriter) WriteInt(val int) error
func (xw *XMLWriter) WriteInt64(val int64) error
func (xw *XMLWriter) WriteUint64(val uint64) error
func (xw *XMLWriter) WriteFloat64(val float64, prec int) error
func (xw *XMLWriter) WriteBool(val bool) error
func (xw *XMLWriter) WriteBoolStr(val bool) error
```

### EMU Unit Write Methods

```go
func (xw *XMLWriter) WriteEMUs(val int64) error
func (xw *XMLWriter) WriteEMUsWithUnit(val int64) error
func (xw *XMLWriter) WriteEMUsF(val float64) error
func (xw *XMLWriter) WriteInchesAsEMU(inches float64) error
func (xw *XMLWriter) WriteCentimetersAsEMU(cm float64) error
func (xw *XMLWriter) WriteMillimetersAsEMU(mm float64) error
func (xw *XMLWriter) WritePointsAsEMU(points float64) error
func (xw *XMLWriter) WritePixelsAsEMU(pixels float64) error
func (xw *XMLWriter) WritePercentage(val int) error
```

### Output

```go
func (xw *XMLWriter) Flush() error
func (xw *XMLWriter) Bytes() []byte
func (xw *XMLWriter) String() string
func (xw *XMLWriter) Size() int
func (xw *XMLWriter) Capacity() int
```

---

## XMLWriterPool

XMLWriter object pool for reducing memory allocations.

### Create

```go
func NewXMLWriterPool() *XMLWriterPool
```

### Methods

```go
func (p *XMLWriterPool) Get() *XMLWriter
func (p *XMLWriterPool) Put(xw *XMLWriter)
func (p *XMLWriterPool) GetWithWriter(w io.Writer) *XMLWriter
func (p *XMLWriterPool) GetBuffered() *XMLWriter
```
