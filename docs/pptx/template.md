# Template - Template System

The template system provides template loading, caching, and management, with support for loading templates from the filesystem, embedded resources, and other sources.

## TemplateType

Template type.

```go
type TemplateType string
```

### Predefined Template Types

```go
const (
    // TemplateBlank is the blank template
    TemplateBlank TemplateType = "blank.pptx"
    // TemplateDefault is the default template (16:9 widescreen)
    TemplateDefault TemplateType = "default.pptx"
    // TemplateWide is the widescreen template
    TemplateWide TemplateType = "wide.pptx"
    // TemplateStandard is the standard template (4:3)
    TemplateStandard TemplateType = "standard.pptx"
)
```

## TemplateManager

The template manager handles lazy loading, caching, and cloning of templates.

```go
type TemplateManager struct {
    // Has unexported fields.
}
```

### Constructors

#### NewTemplateManager

Creates a new template manager.

```go
func NewTemplateManager() *TemplateManager
```

#### NewTemplateManagerWithDir

Creates a template manager with a template directory.

```go
func NewTemplateManagerWithDir(dir string) *TemplateManager
```

**Parameters:**
- `dir`: path to the template file directory

### Template Loading

#### LoadDefault

Loads the default template.

```go
func (tm *TemplateManager) LoadDefault() (*opc.Package, error)
```

#### LoadTemplate

Loads the specified template.

```go
func (tm *TemplateManager) LoadTemplate(name TemplateType) (*opc.Package, error)
```

**Note:** If the template is already cached, a cloned copy is returned directly; otherwise an attempt is made to load it from the filesystem.

**Example:**

```go
tm := pptx.NewTemplateManager()
pkg, err := tm.LoadTemplate(pptx.TemplateDefault)
if err != nil {
    panic(err)
}
```

### Template Registration

#### RegisterTemplate

Registers a template from a file path.

```go
func (tm *TemplateManager) RegisterTemplate(name TemplateType, path string) error
```

**Example:**

```go
tm := pptx.NewTemplateManager()
err := tm.RegisterTemplate("custom", "/path/to/custom.pptx")
if err != nil {
    panic(err)
}
```

#### RegisterTemplateFromBytes

Registers a template from byte data.

```go
func (tm *TemplateManager) RegisterTemplateFromBytes(name TemplateType, data []byte) error
```

**Example:**

```go
data, _ := os.ReadFile("custom.pptx")
err := tm.RegisterTemplateFromBytes("custom", data)
```

#### RegisterTemplateFromFS

Registers a template from a filesystem.

```go
func (tm *TemplateManager) RegisterTemplateFromFS(fsys fs.FS, name TemplateType, path string) error
```

**Example:**

```go
// Register from an embedded filesystem
//go:embed templates/*.pptx
var templateFS embed.FS

err := tm.RegisterTemplateFromFS(templateFS, "custom", "templates/custom.pptx")
```

### Configuration Methods

#### SetDefaultTemplate

Sets the default template.

```go
func (tm *TemplateManager) SetDefaultTemplate(name TemplateType)
```

#### SetTemplateDir

Sets the template directory.

```go
func (tm *TemplateManager) SetTemplateDir(dir string)
```

### Cache Management

#### ClearCache

Clears the template cache.

```go
func (tm *TemplateManager) ClearCache()
```

#### HasTemplate

Checks whether a template has been loaded.

```go
func (tm *TemplateManager) HasTemplate(name TemplateType) bool
```

#### GetMasterCache

Returns the master cache.

```go
func (tm *TemplateManager) GetMasterCache() *MasterCache
```

---

## EmbeddedTemplateManager

An embedded template manager that creates templates programmatically.

```go
type EmbeddedTemplateManager struct {
    // Has unexported fields.
}
```

### Getting the Global Manager

```go
func GetEmbeddedTemplateManager() *EmbeddedTemplateManager
```

### Methods

#### Init

Initializes the embedded templates (runs only once).

```go
func (etm *EmbeddedTemplateManager) Init() error
```

#### HasTemplate

Checks whether a template exists.

```go
func (etm *EmbeddedTemplateManager) HasTemplate(name TemplateType) bool
```

#### GetTemplate

Returns a template (returns a cloned copy).

```go
func (etm *EmbeddedTemplateManager) GetTemplate(name TemplateType) (*opc.Package, error)
```

#### GetDefaultTemplate

Returns the default template.

```go
func (etm *EmbeddedTemplateManager) GetDefaultTemplate() (*opc.Package, error)
```

---

## TemplateBuilder

A template builder for creating PPTX templates from scratch.

```go
type TemplateBuilder struct {
    // Has unexported fields.
}
```

### Constructor

```go
func NewTemplateBuilder() *TemplateBuilder
```

### Methods

#### Package

Returns the underlying OPC package.

```go
func (tb *TemplateBuilder) Package() *opc.Package
```

#### Build

Builds the template and returns the OPC package.

```go
func (tb *TemplateBuilder) Build() *opc.Package
```

#### BuildAndRegister

Builds the template and registers it with the global manager.

```go
func (tb *TemplateBuilder) BuildAndRegister(name TemplateType) error
```

---

## Global Functions

### LoadDefaultTemplate

Loads the default template (using the global manager).

```go
func LoadDefaultTemplate() (*opc.Package, error)
```

### LoadTemplate

Loads the specified template (using the global manager).

```go
func LoadTemplate(name TemplateType) (*opc.Package, error)
```

### RegisterTemplate

Registers a template (using the global manager).

```go
func RegisterTemplate(name TemplateType, path string) error
```

### RegisterTemplateFromBytes

Registers a template from byte data (using the global manager).

```go
func RegisterTemplateFromBytes(name TemplateType, data []byte) error
```

### GetEmbeddedDefaultTemplate

Returns the embedded default template.

```go
func GetEmbeddedDefaultTemplate() (*opc.Package, error)
```

### GetEmbeddedTemplate

Returns an embedded template (using the global manager).

```go
func GetEmbeddedTemplate(name TemplateType) (*opc.Package, error)
```

### InitEmbeddedTemplates

Initializes the embedded templates.

```go
func InitEmbeddedTemplates() error
```

---

## Viewport

### SlideViewport

Slide viewport.

```go
type SlideViewport struct {
    // Width is the viewport width (px)
    Width int
    // Height is the viewport height (px)
    Height int
    // SizeName is the standard size name (optional)
    SizeName string
}
```

### Constructors

```go
func NewSlideViewport(width, height int) *SlideViewport

func NewSlideViewportFromSize(size SlideSize) *SlideViewport
```

### Methods

#### Rect

Returns the viewport rectangle.

```go
func (vp *SlideViewport) Rect() Rect
```

#### CheckBoundary

Checks the boundary of an element.

```go
func (vp *SlideViewport) CheckBoundary(x, y, cx, cy int) BoundaryCheckResult
```

#### CheckRect

Checks the boundary of a rectangle.

```go
func (vp *SlideViewport) CheckRect(rect Rect) BoundaryCheckResult
```

#### IsInside

Checks whether an element is completely within the boundary.

```go
func (vp *SlideViewport) IsInside(x, y, cx, cy int) bool
```

#### IsVisible

Checks whether any part of an element is visible.

```go
func (vp *SlideViewport) IsVisible(x, y, cx, cy int) bool
```

---

## Boundary Checking

### BoundaryStatus

Boundary status.

```go
type BoundaryStatus int
```

**Constants:**

```go
const (
    // BoundaryStatusInside means the element is fully within the boundary
    BoundaryStatusInside BoundaryStatus = iota
    // BoundaryStatusPartial means the element is partially outside the boundary
    BoundaryStatusPartial
    // BoundaryStatusOutside means the element is completely outside the boundary
    BoundaryStatusOutside
    // BoundaryStatusOverflowRight means the element overflows the right edge
    BoundaryStatusOverflowRight
    // BoundaryStatusOverflowLeft means the element overflows the left edge
    BoundaryStatusOverflowLeft
    // BoundaryStatusOverflowTop means the element overflows the top edge
    BoundaryStatusOverflowTop
    // BoundaryStatusOverflowBottom means the element overflows the bottom edge
    BoundaryStatusOverflowBottom
)
```

#### String

Returns a string representation of the boundary status.

```go
func (bs BoundaryStatus) String() string
```

### BoundaryCheckResult

Boundary check result.

```go
type BoundaryCheckResult struct {
    // Status is the boundary status
    Status BoundaryStatus
    // ElementRect is the element rectangle (x, y, cx, cy in px)
    ElementRect Rect
    // ViewportRect is the viewport rectangle (0, 0, width, height in px)
    ViewportRect Rect
    // OverflowX is the overflow amount in the X direction
    // (positive = overflows right edge, negative = overflows left edge)
    OverflowX int
    // OverflowY is the overflow amount in the Y direction
    // (positive = overflows bottom edge, negative = overflows top edge)
    OverflowY int
    // IsVisible indicates whether any part of the element is visible
    // (at least partially within the viewport)
    IsVisible bool
}
```

### Rect

Rectangular area.

```go
type Rect struct {
    X, Y   int // top-left coordinates (px)
    Cx, Cy int // width and height (px)
}
```

---

## SlideSize

Slide size.

```go
type SlideSize struct {
    Width  int // width (px)
    Height int // height (px)
}
```

### Preset Sizes

```go
var (
    // SlideSize16x9 is the widescreen slide size (16:9)
    // Width:  1280 px (13.333 inches)
    // Height: 720 px (7.5 inches)
    SlideSize16x9 = SlideSize{Width: 1280, Height: 720}

    // SlideSize4x3 is the standard slide size (4:3)
    // Width:  960 px (10 inches)
    // Height: 720 px (7.5 inches)
    SlideSize4x3 = SlideSize{Width: 960, Height: 720}

    // SlideSize16x10 is the wide slide size (16:10)
    // Width:  1280 px (13.333 inches)
    // Height: 800 px (8.333 inches)
    SlideSize16x10 = SlideSize{Width: 1280, Height: 800}
)
```

---

## Usage Examples

### Using Predefined Templates

```go
// Create a presentation with the default template
pres, err := pptx.NewWithTemplate(pptx.TemplateDefault)
if err != nil {
    panic(err)
}

// Use the blank template
pres, err = pptx.NewWithTemplate(pptx.TemplateBlank)

// Use the widescreen template
pres, err = pptx.NewWithTemplate(pptx.TemplateWide)

// Use the standard template (4:3)
pres, err = pptx.NewWithTemplate(pptx.TemplateStandard)
```

### Registering Custom Templates

```go
// Register from a file
err := pptx.RegisterTemplate("custom", "/path/to/custom.pptx")
if err != nil {
    panic(err)
}

// Register from byte data
data, _ := os.ReadFile("custom.pptx")
err = pptx.RegisterTemplateFromBytes("custom", data)

// Use the custom template
pres, err := pptx.NewWithTemplate("custom")
```

### Using the Template Manager

```go
// Create a template manager
tm := pptx.NewTemplateManagerWithDir("/path/to/templates")

// Register templates
tm.RegisterTemplate("report", "report.pptx")
tm.RegisterTemplate("proposal", "proposal.pptx")

// Load a template
pkg, err := tm.LoadTemplate("report")
if err != nil {
    panic(err)
}

// Check whether a template exists
if tm.HasTemplate("proposal") {
    fmt.Println("Template loaded")
}

// Clear the cache
tm.ClearCache()
```

### Boundary Checking

```go
slide := pres.AddSlide()

// Check element boundary
result := slide.CheckBoundary(100, 100, 200, 150)

switch result.Status {
case pptx.BoundaryStatusInside:
    fmt.Println("Element is fully within the boundary")
case pptx.BoundaryStatusPartial:
    fmt.Printf("Element is partially outside: X=%d, Y=%d\n", result.OverflowX, result.OverflowY)
case pptx.BoundaryStatusOutside:
    fmt.Println("Element is completely outside the boundary")
}

// Quick checks
if slide.IsInsideBoundary(100, 100, 200, 150) {
    fmt.Println("Element is within the boundary")
}

if slide.IsVisible(100, 100, 200, 150) {
    fmt.Println("Element is at least partially visible")
}
```

### Using the Viewport

```go
// Create a viewport
viewport := pptx.NewSlideViewport(1280, 720)

// Check boundary
rect := pptx.Rect{X: 100, Y: 100, Cx: 200, Cy: 150}
result := viewport.CheckRect(rect)

fmt.Printf("Boundary status: %s\n", result.Status)
fmt.Printf("Is visible: %v\n", result.IsVisible)
```

### Loading from Embedded Resources

```go
//go:embed templates/*.pptx
var templateFS embed.FS

func main() {
    // Initialize embedded templates
    err := pptx.InitEmbeddedTemplates()
    if err != nil {
        panic(err)
    }

    // Use the embedded template
    pres, err := pptx.NewWithTemplate(pptx.TemplateDefault)
}
```

### Cloning and Modifying a Template

```go
// Load a template
pres, _ := pptx.NewWithTemplate(pptx.TemplateDefault)

// Clone the presentation
presCopy, err := pres.Clone()
if err != nil {
    panic(err)
}

// Modify the clone
presCopy.AddSlide()

// The original is unaffected
fmt.Printf("Original: %d slides, Clone: %d slides\n", pres.SlideCount(), presCopy.SlideCount())
```
