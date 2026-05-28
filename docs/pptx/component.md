# Component - Component System

The component system provides reusable rendering building blocks with support for custom components and the composite pattern.

## Core Interface

### Component

All building blocks that can be rendered onto a slide must implement this interface.

```go
type Component interface {
    // Render renders the component onto the slide.
    // ctx: provides the context and resource access the component needs.
    // Returns an error if rendering fails.
    Render(ctx *SlideContext) error
}
```

## Built-in Components

### ShapeComponent

A shape component — the most fundamental component type, directly wrapping an XSp.

```go
type ShapeComponent struct {
    // Has unexported fields.
}
```

#### Constructor

```go
func NewShapeComponent(sp *parts.XSp, x, y int) *ShapeComponent
```

**Parameters:**
- `sp`: the shape object
- `x, y`: position (in EMU)

#### Methods

```go
// Render implements the Component interface
func (sc *ShapeComponent) Render(ctx *SlideContext) error

// Name implements the ComponentWithName interface
func (sc *ShapeComponent) Name() string

// SetName sets the component name
func (sc *ShapeComponent) SetName(name string)

// Position implements the ComponentWithPosition interface
func (sc *ShapeComponent) Position() (x, y int)

// SetPosition implements the ComponentWithPosition interface
func (sc *ShapeComponent) SetPosition(x, y int)

// Bounds implements the ComponentWithSize interface
func (sc *ShapeComponent) Bounds() (x, y, cx, cy int)
```

### CompositeComponent

A composite component that combines multiple components into one.

```go
type CompositeComponent struct {
    // Has unexported fields.
}
```

#### Constructor

```go
func NewCompositeComponent(name string, components ...Component) *CompositeComponent
```

**Example:**

```go
// Create a composite component
header := NewCompositeComponent("header",
    &TitleComponent{Text: "Title"},
    &SubtitleComponent{Text: "Subtitle"},
)
```

#### Methods

```go
// Render implements the Component interface
func (cc *CompositeComponent) Render(ctx *SlideContext) error

// Add adds a child component
func (cc *CompositeComponent) Add(c Component)

// Components returns all child components
func (cc *CompositeComponent) Components() []Component

// Name implements the ComponentWithName interface
func (cc *CompositeComponent) Name() string
```

### ConditionalComponent

A conditional component that renders based on a condition.

```go
type ConditionalComponent struct {
    // Has unexported fields.
}
```

#### Constructor

```go
func NewConditionalComponent(condition func() bool, ifComponent, elseComponent Component) *ConditionalComponent
```

**Parameters:**
- `condition`: the condition function
- `ifComponent`: the component rendered when the condition is true
- `elseComponent`: the component rendered when the condition is false (may be nil)

**Example:**

```go
// Display different components based on a condition
condComp := NewConditionalComponent(
    func() bool { return showAdvanced },
    &AdvancedComponent{},
    &SimpleComponent{},
)
```

#### Methods

```go
// Render implements the Component interface
func (cc *ConditionalComponent) Render(ctx *SlideContext) error
```

### RepeatedComponent

A repeated component that renders a component once for each item in a data slice.

```go
type RepeatedComponent struct {
    // Has unexported fields.
}
```

#### Constructor

```go
func NewRepeatedComponent(count int, template func(index int) Component) *RepeatedComponent
```

**Parameters:**
- `count`: number of repetitions
- `template`: component template function that receives an index and returns a component

**Example:**

```go
// Create a list component
list := NewRepeatedComponent(5, func(index int) Component {
    return &ListItemComponent{
        Text:  fmt.Sprintf("Item %d", index+1),
        Index: index,
    }
})
```

#### Methods

```go
// Render implements the Component interface
func (rc *RepeatedComponent) Render(ctx *SlideContext) error
```

### FuncComponent

A functional component that wraps an ordinary function as a Component.

```go
type FuncComponent func(ctx *SlideContext) error
```

**Example:**

```go
// Create a functional component
myFunc := FuncComponent(func(ctx *SlideContext) error {
    // Render directly using the context
    ctx.AppendShape(&parts.XSp{
        // ... shape properties
    })
    return nil
})

// Add to a slide
slide.AddComponent(myFunc)
```

#### Methods

```go
// Render implements the Component interface
func (fc FuncComponent) Render(ctx *SlideContext) error
```

## Extension Interfaces

### ComponentWithName

A component with a name.

```go
type ComponentWithName interface {
    Component
    // Name returns the component name (used for debugging and logging)
    Name() string
}
```

### ComponentWithPosition

A positionable component.

```go
type ComponentWithPosition interface {
    Component
    // SetPosition sets the component position (in EMU)
    SetPosition(x, y int)
    // Position returns the component position
    Position() (x, y int)
}
```

### ComponentWithSize

A component with size information.

```go
type ComponentWithSize interface {
    Component
    // Bounds returns the component's bounding box (x, y, cx, cy in EMU)
    Bounds() (x, y, cx, cy int)
}
```

### ComponentWithSizeSetter

A resizable component.

```go
type ComponentWithSizeSetter interface {
    Component
    // SetSize sets the component size (in EMU)
    SetSize(cx, cy int)
    // Size returns the component size
    Size() (cx, cy int)
}
```

## Component List

### ComponentList

A component list for managing components in bulk.

```go
type ComponentList []Component
```

#### Methods

```go
// Add appends a component to the list
func (cl *ComponentList) Add(c Component)

// Count returns the number of components
func (cl ComponentList) Count() int

// RenderAll renders all components
func (cl ComponentList) RenderAll(ctx *SlideContext) error
```

**Example:**

```go
var list ComponentList
list.Add(&TitleComponent{Text: "Title"})
list.Add(&BodyComponent{Text: "Body"})

fmt.Printf("Component count: %d\n", list.Count())

// Render all components
ctx := slide.NewContext()
list.RenderAll(ctx)
```

## Error Handling

### ComponentRenderError

A component rendering error.

```go
type ComponentRenderError struct {
    Index      int       // component index
    Component  Component // the component object
    Underlying error     // the underlying error
}
```

#### Methods

```go
// Error implements the error interface
func (e *ComponentRenderError) Error() string

// Unwrap returns the underlying error
func (e *ComponentRenderError) Unwrap() error
```

## Custom Component Examples

### Basic Custom Component

```go
// TitleComponent is a title component
type TitleComponent struct {
    Text string
    X, Y int
    FontSize int
    Color    string
}

func (t *TitleComponent) Render(ctx *SlideContext) error {
    // Defaults
    if t.FontSize == 0 {
        t.FontSize = 44
    }
    if t.Color == "" {
        t.Color = "000000"
    }

    // Create a text shape
    sp := &parts.XSp{
        NvSpPr: &parts.XNvSpPr{
            CNvPr: &parts.XCNvPr{
                ID:   ctx.NextShapeID(),
                Name: "Title",
            },
        },
        SpPr: &parts.XSpPr{
            Xfrm: &parts.XXfrm{
                Off: &parts.XPoint{
                    X: ctx.PxToEMU(t.X),
                    Y: ctx.PxToEMU(t.Y),
                },
                Ext: &parts.XSize{
                    Cx: ctx.PxToEMU(600),
                    Cy: ctx.PxToEMU(60),
                },
            },
        },
        TxBody: &parts.XTxBody{
            BodyPr: &parts.XBodyPr{
                Wrap: "square",
            },
            P: []*parts.XP{
                {
                    R: []*parts.XR{
                        {
                            T:  t.Text,
                            RPr: &parts.XRPr{
                                Sz:     t.FontSize * 100,
                                SolidFill: &parts.XSolidFill{
                                    SrgbClr: &parts.XSrgbClr{
                                        Val: t.Color,
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    ctx.AppendShape(sp)
    return nil
}
```

### Component with an Image

```go
// ImageComponent is an image component
type ImageComponent struct {
    X, Y     int
    Cx, Cy   int
    ImagePath string
}

func (img *ImageComponent) Render(ctx *SlideContext) error {
    // Read the image file
    data, err := os.ReadFile(img.ImagePath)
    if err != nil {
        return err
    }

    // Add the media asset
    rId, err := ctx.AddImage(data, filepath.Base(img.ImagePath))
    if err != nil {
        return err
    }

    // Create an image shape
    pic := &parts.XPicture{
        NvPicPr: &parts.XNvPicPr{
            CNvPr: &parts.XCNvPr{
                ID:   ctx.NextShapeID(),
                Name: "Picture",
            },
        },
        BlipFill: &parts.XBlipFill{
            Blip: &parts.XBlip{
                Embed: rId,
            },
        },
        SpPr: &parts.XSpPr{
            Xfrm: &parts.XXfrm{
                Off: &parts.XPoint{
                    X: ctx.PxToEMU(img.X),
                    Y: ctx.PxToEMU(img.Y),
                },
                Ext: &parts.XSize{
                    Cx: ctx.PxToEMU(img.Cx),
                    Cy: ctx.PxToEMU(img.Cy),
                },
            },
        },
    }

    ctx.AppendShape(pic)
    return nil
}
```

### Composite Component Example

```go
// CardComponent is a card component
type CardComponent struct {
    X, Y       int
    Width      int
    Title      string
    Content    string
    Background string
}

func (c *CardComponent) Render(ctx *SlideContext) error {
    // Create a composite component
    card := NewCompositeComponent("Card")

    // Add background rectangle
    bgColor := c.Background
    if bgColor == "" {
        bgColor = "F5F5F5"
    }

    bgShape := &parts.XSp{
        // ... set background rectangle properties
    }
    card.Add(NewShapeComponent(bgShape, ctx.PxToEMU(c.X), ctx.PxToEMU(c.Y)))

    // Add title
    card.Add(&TitleComponent{
        Text: c.Title,
        X:    c.X + 20,
        Y:    c.Y + 20,
    })

    // Add content
    card.Add(&BodyComponent{
        Text: c.Content,
        X:    c.X + 20,
        Y:    c.Y + 80,
    })

    return card.Render(ctx)
}
```

### Combining Conditional and Repeated Components

```go
// ListComponent is a list component
type ListComponent struct {
    X, Y      int
    ItemWidth int
    Items     []string
    ShowIndex bool
}

func (l *ListComponent) Render(ctx *SlideContext) error {
    // Create a repeated component
    list := NewRepeatedComponent(len(l.Items), func(index int) Component {
        // Create a composite list item
        item := NewCompositeComponent(fmt.Sprintf("Item%d", index))

        // Optional: add an index number
        if l.ShowIndex {
            item.Add(NewConditionalComponent(
                func() bool { return l.ShowIndex },
                &TextComponent{
                    Text: fmt.Sprintf("%d.", index+1),
                    X:    l.X,
                    Y:    l.Y + index*40,
                },
                nil,
            ))
        }

        // Add text
        item.Add(&TextComponent{
            Text: l.Items[index],
            X:    l.X + 30,
            Y:    l.Y + index*40,
        })

        return item
    })

    return list.Render(ctx)
}
```

## Best Practices

### 1. Single Responsibility

Each component should be responsible for exactly one thing:

```go
// Good
type TitleComponent struct { /* ... */ }
type SubtitleComponent struct { /* ... */ }
type BodyComponent struct { /* ... */ }

// Avoid
type EverythingComponent struct {
    Title, Subtitle, Body, Image string
}
```

### 2. Prefer Composition over Inheritance

Go has no inheritance; use composition to build complex components:

```go
// Build a complex component via composition
func NewSlideLayout(title, content string) Component {
    return NewCompositeComponent("SlideLayout",
        &TitleComponent{Text: title},
        &BodyComponent{Text: content},
    )
}
```

### 3. Provide Sensible Defaults

```go
func (t *TextComponent) Render(ctx *SlideContext) error {
    // Set defaults
    if t.FontSize == 0 {
        t.FontSize = 18
    }
    if t.Color == "" {
        t.Color = "000000"
    }
    // ...
}
```

### 4. Error Handling

```go
func (c *ImageComponent) Render(ctx *SlideContext) error {
    if c.ImagePath == "" {
        return fmt.Errorf("image path is required")
    }

    data, err := os.ReadFile(c.ImagePath)
    if err != nil {
        return fmt.Errorf("failed to read image: %w", err)
    }

    // ...
}
```

### 5. Use ComponentList for Bulk Management

```go
func CreateSlide(slide *pptx.Slide) error {
    var components ComponentList

    components.Add(&HeaderComponent{})
    components.Add(&ContentComponent{})
    components.Add(&FooterComponent{})

    ctx := slide.NewContext()
    return components.RenderAll(ctx)
}
```
