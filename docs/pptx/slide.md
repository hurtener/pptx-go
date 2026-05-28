# Slide - Slides

`Slide` is the high-level slide object, providing methods for adding text, images, tables, shapes, and other elements.

## Type Definition

```go
type Slide struct {
    // Has unexported fields.
}
```

## Basic Information

### Index

Returns the slide index (zero-based).

```go
func (s *Slide) Index() int
```

### Layout

Returns the current layout name.

```go
func (s *Slide) Layout() string
```

### SetLayout

Sets the slide layout.

```go
func (s *Slide) SetLayout(layoutName string) bool
```

**Parameters:**
- `layoutName`: layout name (e.g. "blank", "title", "titleAndContent")

**Returns:**
- Whether the layout was set successfully

### SlideSize

Returns the slide size (in px).

```go
func (s *Slide) SlideSize() (cx, cy int)
```

### SlideSizeEMU

Returns the slide size in EMU (advanced usage).

```go
func (s *Slide) SlideSizeEMU() (cx, cy int)
```

### Part

Returns the underlying SlidePart.

```go
func (s *Slide) Part() *parts.SlidePart
```

### PartURI

Returns the part URI.

```go
func (s *Slide) PartURI() *opc.PackURI
```

## Text Operations

### AddTextBox

Adds a text box.

```go
func (s *Slide) AddTextBox(x, y, cx, cy int, text string) *parts.XSp
```

**Parameters:**
- `x, y`: position (in px)
- `cx, cy`: size (in px)
- `text`: text content

**Returns:**
- The shape object `*parts.XSp`

**Example:**

```go
// Add a title
title := slide.AddTextBox(100, 50, 600, 60, "Presentation Title")

// Add body text
body := slide.AddTextBox(100, 150, 600, 400, "Body content here...")
```

## Shape Operations

### AddRectangle

Adds a rectangle.

```go
func (s *Slide) AddRectangle(x, y, cx, cy int) *parts.XSp
```

**Example:**

```go
rect := slide.AddRectangle(100, 100, 200, 150)
```

### AddEllipse

Adds an ellipse.

```go
func (s *Slide) AddEllipse(x, y, cx, cy int) *parts.XSp
```

**Example:**

```go
ellipse := slide.AddEllipse(100, 100, 200, 150)
```

### AddRoundRect

Adds a rounded rectangle.

```go
func (s *Slide) AddRoundRect(x, y, cx, cy int) *parts.XSp
```

**Example:**

```go
roundRect := slide.AddRoundRect(100, 100, 200, 150)
```

### AddAutoShape

Adds an auto shape.

```go
func (s *Slide) AddAutoShape(x, y, cx, cy int, presetID string) *parts.XSp
```

**Parameters:**
- `presetID`: preset shape type (e.g. "rectangle", "ellipse", "roundRect")

**Example:**

```go
// Add a rectangle
shape1 := slide.AddAutoShape(100, 100, 200, 150, "rectangle")

// Add an ellipse
shape2 := slide.AddAutoShape(100, 100, 200, 150, "ellipse")

// Add a rounded rectangle
shape3 := slide.AddAutoShape(100, 100, 200, 150, "roundRect")
```

## Image Operations

### AddPicture

Adds a picture.

```go
func (s *Slide) AddPicture(x, y, cx, cy int, imageRId string) *parts.XPicture
```

**Parameters:**
- `x, y`: position (in px)
- `cx, cy`: size (in px)
- `imageRId`: image relationship ID

**Example:**

```go
// Obtain the rId first
rId := slide.AddImageRel("media/image1.png")
pic := slide.AddPicture(100, 100, 400, 300, rId)
```

### AddPictureFromBytes

Adds a picture from byte data, automatically handling media asset addition and relationship ID assignment.

```go
func (s *Slide) AddPictureFromBytes(x, y, cx, cy int, fileName string, data []byte) (*parts.XPicture, error)
```

**Example:**

```go
data, _ := os.ReadFile("logo.png")
pic, err := slide.AddPictureFromBytes(100, 100, 200, 150, "logo.png", data)
if err != nil {
    panic(err)
}
```

### AddPictureFromFile

Adds a picture from a file.

```go
func (s *Slide) AddPictureFromFile(x, y, cx, cy int, path string) (*parts.XPicture, error)
```

**Example:**

```go
pic, err := slide.AddPictureFromFile(100, 100, 400, 300, "photo.png")
if err != nil {
    panic(err)
}
```

## Table Operations

### AddTable

Adds a table.

```go
func (s *Slide) AddTable(x, y, cx, cy, rows, cols int) *parts.XGraphicFrame
```

**Parameters:**
- `x, y`: position (in px)
- `cx, cy`: size (in px)
- `rows, cols`: number of rows and columns

**Returns:**
- The graphic frame object `*parts.XGraphicFrame`

**Example:**

```go
// Add a table with 3 rows and 4 columns
table := slide.AddTable(100, 100, 600, 400, 3, 4)

// Set cell content
slide.SetTableCellText(table, 0, 0, "Name")
slide.SetTableCellText(table, 0, 1, "Age")
slide.SetTableCellText(table, 1, 0, "Alice")
slide.SetTableCellText(table, 1, 1, "25")
```

### SetTableCellText

Sets the text in a table cell.

```go
func (s *Slide) SetTableCellText(gf *parts.XGraphicFrame, row, col int, text string)
```

**Parameters:**
- `gf`: the graphic frame object
- `row, col`: row and column indexes (zero-based)
- `text`: text content

## Component Operations

### AddComponent

Adds a component to the slide.

```go
func (s *Slide) AddComponent(c Component) error
```

**Parameters:**
- `c`: a component that implements the `Component` interface

**Example:**

```go
// Add a custom component
err := slide.AddComponent(&MyCustomComponent{
    Text: "Hello",
    X:    100,
    Y:    100,
})
if err != nil {
    panic(err)
}
```

### AddComponents

Adds multiple components at once.

```go
func (s *Slide) AddComponents(components ...Component) error
```

**Example:**

```go
err := slide.AddComponents(
    &TitleComponent{Text: "Title"},
    &BodyComponent{Text: "Body"},
)
```

### NewContext

Creates a slide context (for manual component rendering).

```go
func (s *Slide) NewContext() *SlideContext
```

### Builder

Returns the slide builder.

```go
func (s *Slide) Builder() *SlideBuilder
```

## Relationship Management

### AddImageRel

Adds an image relationship.

```go
func (s *Slide) AddImageRel(targetURI string) string
```

**Returns:**
- Relationship ID (rId)

### AddMediaRel

Adds a media relationship.

```go
func (s *Slide) AddMediaRel(targetURI string) string
```

### AddChartRel

Adds a chart relationship.

```go
func (s *Slide) AddChartRel(targetURI string) string
```

### GetImageRId

Returns the rId for an image relationship, adding it if it does not exist.

```go
func (s *Slide) GetImageRId(targetURI string) string
```

### HasImage

Checks whether an image relationship already exists.

```go
func (s *Slide) HasImage(targetURI string) bool
```

## Boundary Checking

### CheckBoundary

Checks the boundary of an element.

```go
func (s *Slide) CheckBoundary(x, y, cx, cy int) BoundaryCheckResult
```

**Parameters:**
- `x, y`: top-left coordinates of the element (px)
- `cx, cy`: width and height of the element (px)

**Returns:**
- A boundary check result containing overflow information and visibility state

**Example:**

```go
result := slide.CheckBoundary(100, 100, 200, 150)
switch result.Status {
case pptx.BoundaryStatusInside:
    fmt.Println("Fully within the boundary")
case pptx.BoundaryStatusPartial:
    fmt.Println("Partially outside the boundary")
case pptx.BoundaryStatusOutside:
    fmt.Println("Completely outside the boundary")
}
```

### IsInsideBoundary

Checks whether an element is completely within the boundary.

```go
func (s *Slide) IsInsideBoundary(x, y, cx, cy int) bool
```

### IsVisible

Checks whether any part of an element is visible.

```go
func (s *Slide) IsVisible(x, y, cx, cy int) bool
```

### Viewport

Returns the slide viewport.

```go
func (s *Slide) Viewport() *SlideViewport
```

## Color Handling

### ResolveColor

Resolves a color (supports names, hex, RGB, and theme colors).

```go
func (s *Slide) ResolveColor(color string) Color
```

**Example:**

```go
// Resolve a hex color
c1 := slide.ResolveColor("#FF0000")

// Resolve a theme color
c2 := slide.ResolveColor("accent1")

// Resolve an RGB color
c3 := slide.ResolveColor("rgb(255, 0, 0)")
```

### ValidateColor

Validates a color.

```go
func (s *Slide) ValidateColor(color string) ColorValidationResult
```

---

# SlideBuilder - Slide Builder

`SlideBuilder` provides slide building capabilities, primarily for low-level operations.

## Type Definition

```go
type SlideBuilder struct {
    // Has unexported fields.
}
```

## Constructor

### NewSlideBuilder

Creates a slide builder.

```go
func NewSlideBuilder(slide *parts.SlidePart) *SlideBuilder
```

## Shape Operations

### AddAutoShape

Adds an auto shape to the slide.

```go
func (b *SlideBuilder) AddAutoShape(x, y, cx, cy int, presetID string) *parts.XSp
```

**Note:** Uses EMU units.

### AddTextBox

Adds a text box to the slide.

```go
func (b *SlideBuilder) AddTextBox(x, y, cx, cy int, text string) *parts.XSp
```

**Note:** Uses EMU units.

### AddPicture

Adds a picture to the slide.

```go
func (b *SlideBuilder) AddPicture(x, y, cx, cy int, imageRId string) *parts.XPicture
```

**Note:** Uses EMU units.

### AddTable

Adds a table to the slide.

```go
func (b *SlideBuilder) AddTable(x, y, cx, cy, rows, cols int) *parts.XGraphicFrame
```

**Note:** Uses EMU units.

## Relationship Management

### AddImage

Adds an image relationship and returns the rId.

```go
func (b *SlideBuilder) AddImage(targetURI string) string
```

### AddMedia

Adds a media relationship and returns the rId.

```go
func (b *SlideBuilder) AddMedia(targetURI string) string
```

### AddChart

Adds a chart relationship and returns the rId.

```go
func (b *SlideBuilder) AddChart(targetURI string) string
```

### GetImageRId

Returns the rId for an image relationship, adding it if it does not exist.

```go
func (b *SlideBuilder) GetImageRId(targetURI string) string
```

### GetChartRId

Returns the rId for a chart relationship, adding it if it does not exist.

```go
func (b *SlideBuilder) GetChartRId(targetURI string) string
```

### GetMediaRId

Returns the rId for a media relationship, adding it if it does not exist.

```go
func (b *SlideBuilder) GetMediaRId(targetURI string) string
```

### HasImage

Checks whether an image relationship already exists.

```go
func (b *SlideBuilder) HasImage(targetURI string) bool
```

### HasMedia

Checks whether a media relationship already exists.

```go
func (b *SlideBuilder) HasMedia(targetURI string) bool
```

### GetRelationshipURI

Returns the target URI for a given rId.

```go
func (b *SlideBuilder) GetRelationshipURI(rId string) string
```

## Helper Methods

### GetOrAddPicture

Adds a picture to the slide and returns an XPicture, automatically handling the image relationship ID.

```go
func (b *SlideBuilder) GetOrAddPicture(x, y, cx, cy int, imageURI string) *parts.XPicture
```

### SetTableCellText

Sets the text in a table cell.

```go
func (b *SlideBuilder) SetTableCellText(gf *parts.XGraphicFrame, row, col int, text string)
```

### Slide

Returns the underlying SlidePart.

```go
func (b *SlideBuilder) Slide() *parts.SlidePart
```

---

# SlideContext - Slide Rendering Context

`SlideContext` provides the resources and capabilities needed for component rendering.

## Type Definition

```go
type SlideContext struct {
    // Has unexported fields.
}
```

## Constructor

### NewSlideContext

Creates a slide context.

```go
func NewSlideContext(s *Slide) *SlideContext
```

## Shape ID Management

### NextShapeID

Allocates the next shape ID, guaranteed to be conflict-free (thread-safe).

```go
func (ctx *SlideContext) NextShapeID() uint32
```

### CurrentShapeID

Returns the current shape ID (the most recently allocated one).

```go
func (ctx *SlideContext) CurrentShapeID() uint32
```

### AllocateShapeIDBatch

Allocates a batch of shape IDs.

```go
func (ctx *SlideContext) AllocateShapeIDBatch(count int) []uint32
```

### IsShapeIDAllocated

Checks whether a shape ID has already been allocated.

```go
func (ctx *SlideContext) IsShapeIDAllocated(id uint32) bool
```

## Shape Appending

### AppendShape

Appends a shape to the slide.

```go
func (ctx *SlideContext) AppendShape(shape interface{})
```

**Parameters:**
- `shape`: a shape struct (`*parts.XSp`, `*parts.XPicture`, `*parts.XGraphicFrame`, etc.)

**Example:**

```go
sp := &parts.XSp{
    // ... set shape properties
}
ctx.AppendShape(sp)
```

### AppendShapes

Appends multiple shapes at once.

```go
func (ctx *SlideContext) AppendShapes(shapes ...interface{})
```

## Adding Media

### AddMedia

Adds a media asset (image, audio, or video).

```go
func (ctx *SlideContext) AddMedia(data []byte, fileName string) (string, error)
```

**Returns:**
- The relationship ID and an error

### AddImage

Adds an image asset (alias for AddMedia).

```go
func (ctx *SlideContext) AddImage(data []byte, fileName string) (string, error)
```

### AddAudio

Adds an audio asset.

```go
func (ctx *SlideContext) AddAudio(data []byte, fileName string) (string, error)
```

### AddVideo

Adds a video asset.

```go
func (ctx *SlideContext) AddVideo(data []byte, fileName string) (string, error)
```

### AddMediaWithMIME

Adds a media asset with an explicit MIME type.

```go
func (ctx *SlideContext) AddMediaWithMIME(data []byte, fileName, mimeType string) (string, error)
```

## Adding Charts

### AddChart

Adds a chart using a template.

```go
func (ctx *SlideContext) AddChart(chartType parts.ChartType, data map[string]interface{}) (string, error)
```

**Parameters:**
- `chartType`: chart type
- `data`: chart data

**Returns:**
- The relationship ID and an error

### AddChartXML

Adds a chart from raw XML.

```go
func (ctx *SlideContext) AddChartXML(chartXML []byte) (string, error)
```

## Relationship Management

### AddImageRel

Adds an image relationship.

```go
func (ctx *SlideContext) AddImageRel(targetURI string) string
```

### AddMediaRel

Adds a media relationship.

```go
func (ctx *SlideContext) AddMediaRel(targetURI string) string
```

### AddChartRel

Adds a chart relationship.

```go
func (ctx *SlideContext) AddChartRel(targetURI string) string
```

### HasRelationship

Checks whether a relationship exists.

```go
func (ctx *SlideContext) HasRelationship(rID string) bool
```

## Unit Conversion

### PxToEMU

Converts pixels to EMU (based on 96 DPI).

```go
func (ctx *SlideContext) PxToEMU(px int) int
```

### EMUToPx

Converts EMU to pixels (based on 96 DPI).

```go
func (ctx *SlideContext) EMUToPx(emu int) int
```

## Other Methods

### SlideIndex

Returns the slide index.

```go
func (ctx *SlideContext) SlideIndex() int
```

### SlideSize

Returns the slide size (cx, cy in EMU).

```go
func (ctx *SlideContext) SlideSize() (cx, cy int)
```

### SlidePart

Returns the underlying SlidePart (advanced usage).

```go
func (ctx *SlideContext) SlidePart() *parts.SlidePart
```

### Presentation

Returns the parent presentation (advanced usage).

```go
func (ctx *SlideContext) Presentation() *Presentation
```

### RenderComponents

Renders multiple components at once.

```go
func (ctx *SlideContext) RenderComponents(components ...Component) error
```
