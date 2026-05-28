# Master Module — Interface Documentation

> Read-only data structures, parsers, and cache management for slide masters and layouts

## Design Principles

1. All struct fields are read-only (lowercase fields initialised via constructors; uppercase fields are immutable values).
2. Optimised for high-concurrency reads — safe to read without locking.
3. Data is built once at parse time and never modified afterwards.

---

## Enum Types

### PlaceholderType

Placeholder type enum, corresponding to XML: `<p:ph type="...">`

| Constant | Value | XML Type |
|------|-----|----------|
| `PlaceholderTypeNone` | `0` | Unspecified |
| `PlaceholderTypeTitle` | `1` | `title` |
| `PlaceholderTypeBody` | `2` | `body` |
| `PlaceholderTypeCenterTitle` | `3` | `ctrTitle` |
| `PlaceholderTypeSubTitle` | `4` | `subTitle` |
| `PlaceholderTypeDateTime` | `5` | `dt` |
| `PlaceholderTypeSlideNumber` | `6` | `sldNum` |
| `PlaceholderTypeFooter` | `7` | `ftr` |
| `PlaceholderTypeHeader` | `8` | `hdr` |
| `PlaceholderTypeObject` | `9` | `obj` |
| `PlaceholderTypeChart` | `10` | `chart` |
| `PlaceholderTypeTable` | `11` | `tbl` |
| `PlaceholderTypeClipArt` | `12` | `clipArt` |
| `PlaceholderTypeOrgChart` | `13` | `dgm` |
| `PlaceholderTypeMedia` | `14` | `media` |
| `PlaceholderTypeSlideImage` | `15` | `sldImg` |
| `PlaceholderTypePicture` | `16` | `pic` |

#### Methods

```go
func (t PlaceholderType) String() string
```
Returns the string representation of the placeholder type, e.g. `"title"`, `"body"`, etc.

### BackgroundType

Background type enum, corresponding to XML: different child elements under `<p:bg>`.

| Constant | Value | Description |
|------|-----|------|
| `BackgroundTypeNone` | `0` | No background |
| `BackgroundTypeSolidColor` | `1` | Solid color background |
| `BackgroundTypeGradient` | `2` | Gradient background |
| `BackgroundTypePattern` | `3` | Pattern fill |
| `BackgroundTypePicture` | `4` | Picture background |
| `BackgroundTypeThemeColor` | `5` | Theme color background |

### SlideLayoutType

Layout type enum.

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

## Read-Only Data Structures

### TextStyle

Text style, used to define the default font, size, colour, etc. for text in a placeholder.

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `fontName` | `string` | `FontName()` | Font name |
| `fontSize` | `int32` | `FontSize()` | Font size (hundredths of a point; 100 = 1 pt) |
| `bold` | `bool` | `Bold()` | Bold |
| `italic` | `bool` | `Italic()` | Italic |
| `underline` | `bool` | `Underline()` | Underline |
| `colorRGB` | `string` | `ColorRGB()` | Text color (RGB hex, e.g. `"FF0000"`) |

### Placeholder

Placeholder — a fillable region defined in a master/layout. Corresponds to XML: `<p:sp>` with `<p:nvSpPr><p:nvPr><p:ph ...>`

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `id` | `string` | `ID()` | Unique placeholder identifier |
| `placeholderType` | `PlaceholderType` | `Type()` | Placeholder type |
| `x` | `int64` | `X()` | X coordinate (EMU units) |
| `y` | `int64` | `Y()` | Y coordinate (EMU units) |
| `cx` | `int64` | `Cx()` | Width (EMU units) |
| `cy` | `int64` | `Cy()` | Height (EMU units) |
| `rotation` | `int32` | `Rotation()` | Rotation angle (1/60000 of a degree) |
| `defaultStyle` | `*TextStyle` | `DefaultStyle()` | Default text style (may be nil) |

#### Methods

```go
func (p *Placeholder) Bounds() (x, y, cx, cy int64)
```
Returns the bounding rectangle (x, y, cx, cy).

### Background

Background definition, corresponding to XML: `<p:bg>` or `<p:cSld><p:bg>`

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `backgroundType` | `BackgroundType` | `Type()` | Background type |
| `solidColorRGB` | `string` | `SolidColorRGB()` | RGB hex color value (solid backgrounds only) |
| `gradientAngle` | `int32` | `GradientAngle()` | Gradient angle in degrees (gradient backgrounds only) |
| `gradientColors` | `[]GradientStop` | `GradientColors()` | List of gradient stops (gradient backgrounds only) |
| `pictureRId` | `string` | `PictureRId()` | Picture relationship ID (picture backgrounds only) |
| `pictureURI` | `string` | `PictureURI()` | Picture internal URI path (picture backgrounds only) |
| `opacity` | `float32` | `Opacity()` | Opacity (0.0 – 1.0) |

### GradientStop

Gradient stop, corresponding to XML: `<a:gs>`

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `position` | `float32` | `Position()` | Position (0.0 – 1.0) |
| `colorRGB` | `string` | `ColorRGB()` | RGB hex color value |

### SlideLayoutData

Read-only layout data, corresponding to XML: `/ppt/slideLayouts/slideLayoutN.xml`

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `id` | `string` | `ID()` | Unique layout identifier |
| `name` | `string` | `Name()` | Layout name |
| `layoutType` | `SlideLayoutType` | `LayoutType()` | Layout type |
| `background` | `*Background` | `Background()` | Background (may be nil) |
| `masterId` | `string` | `MasterID()` | ID of the owning master |
| `placeholders` | `map[string]*Placeholder` | `Placeholders()` | Placeholder collection |

#### Methods

```go
func (l *SlideLayoutData) PlaceholderByID(id string) *Placeholder
func (l *SlideLayoutData) PlaceholderCount() int
func (l *SlideLayoutData) PlaceholderByType(phType PlaceholderType) *Placeholder
func (l *SlideLayoutData) TitlePlaceholder() *Placeholder
func (l *SlideLayoutData) BodyPlaceholder() *Placeholder
```

### SlideMasterData

Read-only master data, corresponding to XML: `/ppt/slideMasters/slideMasterN.xml`

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `id` | `string` | `ID()` | Unique master identifier |
| `name` | `string` | `Name()` | Master name |
| `background` | `*Background` | `Background()` | Background (may be nil) |
| `placeholders` | `map[string]*Placeholder` | `Placeholders()` | Master-level placeholders |
| `layouts` | `[]*SlideLayoutData` | `Layouts()` | List of contained layouts |

#### Methods

```go
func (m *SlideMasterData) PlaceholderByID(id string) *Placeholder
func (m *SlideMasterData) PlaceholderCount() int
func (m *SlideMasterData) LayoutCount() int
func (m *SlideMasterData) LayoutByID(id string) *SlideLayoutData
```

---

## MasterCache

Read-only master/layout cache. All fields are read-only after initialisation; lock-free concurrent access is safe.

### Create

```go
func NewMasterCache() *MasterCache
```

### Initialise

```go
func (c *MasterCache) Init(masters []*SlideMasterData, layouts []*SlideLayoutData)
func (c *MasterCache) InitFunc(initFn func() ([]*SlideMasterData, []*SlideLayoutData))
```

### Read Interface

```go
func (c *MasterCache) GetMaster(masterID string) (*SlideMasterData, bool)
func (c *MasterCache) GetMasterByName(name string) (*SlideMasterData, bool)
func (c *MasterCache) GetLayout(layoutID string) (*SlideLayoutData, bool)
func (c *MasterCache) GetLayoutByName(name string) (*SlideLayoutData, bool)
func (c *MasterCache) GetPlaceholder(layoutID, phType string) (*Placeholder, bool)
func (c *MasterCache) GetPlaceholderByID(layoutID, placeholderID string) (*Placeholder, bool)
func (c *MasterCache) GetMasterPlaceholder(masterID, phType string) (*Placeholder, bool)
```

### Bulk Read

```go
func (c *MasterCache) AllMasters() map[string]*SlideMasterData
func (c *MasterCache) AllLayouts() map[string]*SlideLayoutData
func (c *MasterCache) MasterCount() int
func (c *MasterCache) LayoutCount() int
```

### Helper Methods

```go
func (c *MasterCache) LayoutExists(layoutID string) bool
func (c *MasterCache) MasterExists(masterID string) bool
func (c *MasterCache) ListLayoutIDs() []string
func (c *MasterCache) ListMasterIDs() []string
func (c *MasterCache) ListLayoutNames() []string
```

### Global Cache

```go
func DefaultCache() *MasterCache
func InitDefaultCache(masters []*SlideMasterData, layouts []*SlideLayoutData)
func GetLayout(layoutID string) (*SlideLayoutData, bool)
func GetLayoutByName(name string) (*SlideLayoutData, bool)
func GetMaster(masterID string) (*SlideMasterData, bool)
func GetPlaceholder(layoutID, phType string) (*Placeholder, bool)
```

---

## MasterManager

Master/layout manager (facade pattern), responsible for loading masters and layouts from a ZIP file.

### Create

```go
func NewMasterManager() *MasterManager
func NewMasterManagerWithCache(cache *MasterCache) *MasterManager
```

### Load

```go
func (m *MasterManager) LoadFromZip(zipReader *zip.Reader) error
func (m *MasterManager) LoadFromZipFile(filePath string) error
```

### Accessors

```go
func (m *MasterManager) Cache() *MasterCache
func (m *MasterManager) GetLayout(layoutID string) (*SlideLayoutData, bool)
func (m *MasterManager) GetLayoutByName(name string) (*SlideLayoutData, bool)
func (m *MasterManager) GetMaster(masterID string) (*SlideMasterData, bool)
func (m *MasterManager) GetMasterByName(name string) (*SlideMasterData, bool)
func (m *MasterManager) GetPlaceholder(layoutID, phType string) (*Placeholder, bool)
func (m *MasterManager) AllLayouts() map[string]*SlideLayoutData
func (m *MasterManager) AllMasters() map[string]*SlideMasterData
func (m *MasterManager) LayoutCount() int
func (m *MasterManager) MasterCount() int
func (m *MasterManager) ListLayoutIDs() []string
func (m *MasterManager) ListLayoutNames() []string
```

### Global Manager

```go
func DefaultManager() *MasterManager
func InitDefaultManager(zipReader *zip.Reader) error
func InitDefaultManagerFromFile(filePath string) error
```

---

## XML Parsers

### ParseLayout

```go
func ParseLayout(xmlData []byte) (*SlideLayoutData, error)
```

Parses slide layout XML and returns layout data.

### ParseMaster

```go
func ParseMaster(xmlData []byte) (*SlideMasterData, error)
```

Parses slide master XML and returns master data.

---

## Unit Conversions

```go
var EMUToPixels      = utils.EMUToPixels      // EMU -> pixels (96 DPI)
var EMUToPoints       = utils.EMUToPoints       // EMU -> points
var EMUToInches       = utils.EMUToInches       // EMU -> inches
var EMUToCentimeters  = utils.EMUToCentimeters  // EMU -> centimetres
```
