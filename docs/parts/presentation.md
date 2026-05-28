# Presentation Module — Interface Documentation

> Corresponds to `/ppt/presentation.xml`, the logical root node of the entire PPTX

---

## Constants

```go
SlideIDStart = 256  // Starting value for slide IDs
```

---

## Data Structures

### SlideSize

Slide size in EMU (English Metric Units).

| Field | Type | Description |
|------|------|------|
| `Cx` | `int` | Width |
| `Cy` | `int` | Height |

### StandardSlideSizes

Standard slide sizes.

| Field | Type | Dimensions |
|------|------|------|
| `Wide16x9` | `SlideSize` | 12192000 x 6858000 EMU |
| `Standard4x3` | `SlideSize` | 9144000 x 6858000 EMU |

### PresentationPart

Presentation part, corresponding to `/ppt/presentation.xml`.

| Field | Type | Description |
|------|------|------|
| `uri` | `*opc.PackURI` | Part URI |
| `slideIDs` | `[]uint32` | List of allocated slide IDs |
| `slideIDNext` | `uint32` | Next slide ID to allocate |
| `slideCount` | `int32` | Current slide count |
| `slideMasterIDs` | `[]string` | List of master rIds |
| `slideLayoutIDs` | `[]string` | List of layout rIds |
| `slideSize` | `SlideSize` | Slide size |
| `notesMasterID` | `string` | Notes master rId |
| `themeID` | `string` | Theme rId |

---

## XML Struct Types

### XPresentation

Complete XML structure corresponding to `presentation.xml`.

| Field | XML Path | Type | Description |
|------|----------|------|------|
| `XmlnsA` | `xmlns:a` | `string` | DrawingML namespace |
| `XmlnsR` | `xmlns:r` | `string` | Relationships namespace |
| `XmlnsP` | `xmlns:p` | `string` | PresentationML namespace |
| `Compatibility` | `p:compatSpt` | `*XCompatibility` | Compatibility settings |
| `SldSz` | `p:sldSz` | `*XSldSz` | Slide size |
| `NotesSz` | `p:notesSz` | `*XSldSz` | Notes size |
| `SldIdLst` | `p:sldIdLst` | `*XSldIdLst` | Slide ID list |
| `SldMasterIdLst` | `p:sldMasterIdLst` | `*XSldMasterIdLst` | Master ID list |
| `NotesMasterIdLst` | `p:notesMasterIdLst` | `*XSldMasterIdLst` | Notes master ID list |
| `PrintSettings` | `p:printSettings` | `*XPrintSettings` | Print settings |

### XCompatibility

Compatibility settings, corresponding to `p:compatSpt`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `CompatMode` | `compatMode` | `string` | Compatibility mode |

### XSldSz

Slide size, corresponding to `p:sldSz`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Cx` | `cx` | `int` | Width |
| `Cy` | `cy` | `int` | Height |

### XSldIdLst

Slide ID list, corresponding to `p:sldIdLst`.

| Field | Type | Description |
|------|------|------|
| `SldIds` | `[]XSldId` | List of slide IDs |

### XSldId

Single slide ID, corresponding to `p:sldId`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Id` | `id` | `uint32` | Slide ID |
| `RId` | `rid` | `string` | Relationship ID |

> **Note**: The XML attribute `r:id` is converted to `rid` during parsing (Go xml tag: `xml:"rid,attr"`).

### XSldMasterIdLst

Master ID list, corresponding to `p:sldMasterIdLst`.

| Field | Type | Description |
|------|------|------|
| `SldMasterIds` | `[]XSldMasterId` | List of master IDs |

### XSldMasterId

Single master ID, corresponding to `p:sldMasterId`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Id` | `id` | `uint32` | Master ID |
| `RId` | `rid` | `string` | Relationship ID |

> **Note**: The XML attribute `r:id` is converted to `rid` during parsing.

### XPrintSettings

Print settings, corresponding to `p:printSettings`.

| Field | Type | Description |
|------|------|------|
| `OutputOptions` | `*XOutputOptions` | Output options |

### XOutputOptions

Output options, corresponding to `p:outputOptions`.

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `UsePrintFml` | `usePrintFml` | `*bool` | Use print format |
| `CloneLinkedObjs` | `cloneLinkedObjs` | `*bool` | Clone linked objects |

---

## Constructors

### NewPresentationPart

```go
func NewPresentationPart() *PresentationPart
```

Creates a presentation part using the default 16:9 widescreen size.

### NewPresentationPartWithSize

```go
func NewPresentationPartWithSize(size SlideSize) *PresentationPart
```

Creates a presentation part with the specified size.

---

## PresentationPart Methods

### URI Methods

```go
func (p *PresentationPart) PartURI() *opc.PackURI
```

### Size Methods

```go
func (p *PresentationPart) SlideSize() SlideSize
func (p *PresentationPart) SetSlideSize(size SlideSize)
```

### Slide Management

```go
func (p *PresentationPart) SlideCount() int32
func (p *PresentationPart) SlideIDAt(index int) (uint32, error)
func (p *PresentationPart) SlideIDs() []uint32
func (p *PresentationPart) AddSlide(layoutRId string, slidePart *SlidePart) error
func (p *PresentationPart) RemoveSlide(index int) error
```

### Master Management

```go
func (p *PresentationPart) SlideMasterIDs() []string
func (p *PresentationPart) AddSlideMaster(rId string)
```

### XML Serialization / Deserialization

```go
func (p *PresentationPart) ToXML() ([]byte, error)
func (p *PresentationPart) FromXML(data []byte) error
```

> **Note**: `FromXML` calls `StripNamespacePrefixes` internally to handle namespace prefix issues. See [xmlutils.md](xmlutils.md).

---

## Helper Functions

### NewSlideSizeFromStandard

```go
func NewSlideSizeFromStandard(name string) SlideSize
```

Creates a SlideSize from a standard size name.

| Argument | Description |
|------|------|
| `"16:9"`, `"wide"`, `"widescreen"` | Returns Wide16x9 |
| `"4:3"`, `"standard"` | Returns Standard4x3 |
| Any other value | Returns Wide16x9 by default |

---

## EMU Unit Conversions

### EMUFromPoints / PointsFromEMU

```go
func EMUFromPoints(points float64) int
func PointsFromEMU(emu int) float64
```

Converts between points and EMU (1 pt = 12700 EMU).

### EMUFromInches / InchesFromEMU

```go
func EMUFromInches(inches float64) int
func InchesFromEMU(emu int) float64
```

Converts between inches and EMU (1 inch = 914400 EMU).

### EMUFromMM / MMFromEMU

```go
func EMUFromMM(mm float64) int
func MMFromEMU(emu int) float64
```

Converts between millimeters and EMU (1 mm = 36000 EMU).
