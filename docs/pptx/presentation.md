# Presentation - Presentations

`Presentation` is the top-level facade for a PPTX presentation, providing core functionality such as creation, saving, and import/export.

## Type Definition

```go
type Presentation struct {
    // Has unexported fields.
}
```

## Constructors

### New

Creates a blank presentation using the default template (16:9 widescreen).

```go
func New() *Presentation
```

**Example:**

```go
pres := pptx.New()
```

### NewFromFile

Creates a presentation from a file.

```go
func NewFromFile(path string) (*Presentation, error)
```

**Example:**

```go
pres, err := pptx.NewFromFile("template.pptx")
if err != nil {
    panic(err)
}
```

### NewFromBytes

Creates a presentation from byte data.

```go
func NewFromBytes(data []byte) (*Presentation, error)
```

**Example:**

```go
data, _ := os.ReadFile("template.pptx")
pres, err := pptx.NewFromBytes(data)
if err != nil {
    panic(err)
}
```

### NewWithTemplate

Creates a presentation from a template.

```go
func NewWithTemplate(name TemplateType) (*Presentation, error)
```

**Parameters:**
- `name`: template name (e.g. `TemplateDefault`, `TemplateBlank`)

**Example:**

```go
// Use the default template
pres, err := pptx.NewWithTemplate(pptx.TemplateDefault)

// Use the blank template
pres, err := pptx.NewWithTemplate(pptx.TemplateBlank)

// Use the widescreen template
pres, err := pptx.NewWithTemplate(pptx.TemplateWide)

// Use the standard template (4:3)
pres, err := pptx.NewWithTemplate(pptx.TemplateStandard)
```

## Slide Management

### AddSlide

Adds a new slide.

```go
func (p *Presentation) AddSlide(layout ...string) *Slide
```

**Parameters:**
- `layout`: optional layout name (e.g. "title", "blank", "titleAndContent"); if omitted, a blank layout is used

**Returns:**
- The newly created `*Slide`

**Example:**

```go
// Add a blank slide
slide := pres.AddSlide()

// Add a title slide
slide := pres.AddSlide("title")

// Add a title-and-content slide
slide := pres.AddSlide("titleAndContent")
```

### AddSlideAt

Inserts a slide at the specified position.

```go
func (p *Presentation) AddSlideAt(index int, layout ...string) (*Slide, error)
```

**Parameters:**
- `index`: insertion position (0 is the beginning)
- `layout`: optional layout name

**Example:**

```go
// Insert a slide at position 2
slide, err := pres.AddSlideAt(1)
if err != nil {
    panic(err)
}
```

### GetSlide

Returns the slide at the specified index.

```go
func (p *Presentation) GetSlide(index int) (*Slide, error)
```

**Example:**

```go
slide, err := pres.GetSlide(0) // get the first slide
if err != nil {
    panic(err)
}
```

### RemoveSlide

Removes the slide at the specified index.

```go
func (p *Presentation) RemoveSlide(index int) error
```

**Example:**

```go
err := pres.RemoveSlide(0) // delete the first slide
if err != nil {
    panic(err)
}
```

### Slides

Returns all slides.

```go
func (p *Presentation) Slides() []*Slide
```

### SlideCount

Returns the number of slides.

```go
func (p *Presentation) SlideCount() int
```

**Example:**

```go
count := pres.SlideCount()
fmt.Printf("Total slides: %d\n", count)
```

## Size Settings

### SlideSize

Returns the current slide size.

```go
func (p *Presentation) SlideSize() (int, int)
```

**Returns:**
- Width and height in px

### SetSlideSize

Sets the slide size.

```go
func (p *Presentation) SetSlideSize(cx, cy int)
```

**Parameters:**
- `cx`: width in px
- `cy`: height in px

**Example:**

```go
pres.SetSlideSize(1280, 720) // set to 16:9 widescreen
```

### SetSlideSizeStandard

Sets a standard slide size.

```go
func (p *Presentation) SetSlideSizeStandard(name string)
```

**Example:**

```go
pres.SetSlideSizeStandard("16:9")
pres.SetSlideSizeStandard("4:3")
```

## Output Methods

### Save

Saves the presentation to a file.

```go
func (p *Presentation) Save(path string) error
```

**Example:**

```go
err := pres.Save("output.pptx")
if err != nil {
    panic(err)
}
```

### Write

Writes the presentation to an `io.Writer`.

```go
func (p *Presentation) Write(w io.Writer) error
```

**Usage:** Suitable for high-concurrency streaming output (e.g. HTTP responses).

**Example:**

```go
// HTTP response example
http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
    pres := pptx.New()
    // ... build the presentation

    w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
    w.Header().Set("Content-Disposition", "attachment; filename=output.pptx")
    pres.Write(w)
})
```

### WriteToBytes

Writes the presentation to a byte slice.

```go
func (p *Presentation) WriteToBytes() ([]byte, error)
```

**Example:**

```go
data, err := pres.WriteToBytes()
if err != nil {
    panic(err)
}
// Can be saved to a database or sent over the network
```

## Other Methods

### Clone

Clones the presentation, returning a fully independent copy.

```go
func (p *Presentation) Clone() (*Presentation, error)
```

**Usage:** Creates a copy of the presentation for modification without affecting the original.

**Example:**

```go
presCopy, err := pres.Clone()
if err != nil {
    panic(err)
}
presCopy.AddSlide() // modifying the copy does not affect the original pres
```

### Close

Closes the presentation and releases resources.

```go
func (p *Presentation) Close() error
```

### Package

Returns the underlying OPC package (advanced usage).

```go
func (p *Presentation) Package() *opc.Package
```

### PresentationPart

Returns the presentation part.

```go
func (p *Presentation) PresentationPart() *parts.PresentationPart
```

### MasterCache

Returns the master cache.

```go
func (p *Presentation) MasterCache() *MasterCache
```

### MediaManager

Returns the media manager.

```go
func (p *Presentation) MediaManager() *MediaManager
```

## Complete Example

```go
package main

import (
    "github.com/hurtener/pptx-go/pptx"
)

func main() {
    // Create a presentation
    pres := pptx.New()

    // Set the slide size
    pres.SetSlideSizeStandard("16:9")

    // Add a title slide
    titleSlide := pres.AddSlide("title")
    titleSlide.AddTextBox(100, 100, 400, 50, "Presentation Title")

    // Add a content slide
    contentSlide := pres.AddSlide("titleAndContent")
    contentSlide.AddTextBox(100, 100, 400, 50, "Chapter 1")
    contentSlide.AddTextBox(100, 200, 600, 300, "Body content goes here...")

    // Add a slide with an image
    imageSlide := pres.AddSlide()
    imageSlide.AddPictureFromFile(100, 100, 400, 300, "photo.png")

    // Save the file
    err := pres.Save("demo.pptx")
    if err != nil {
        panic(err)
    }
}
```
