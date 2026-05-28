# Package pptx - High-Level PPTX Interface

## Overview

The `pptx` package provides a high-level interface for working with PPTX files, serving as the primary entry point for both human developers and AI consumers. It encapsulates the complexity of the underlying OPC package and parts layer, exposing a clean and easy-to-use API.

## Core Design Philosophy

- **High-level abstraction**: Business-oriented API that hides OOXML specification details
- **Composability**: Reusable component system designed for composition and extension
- **Type safety**: Strongly typed interfaces that reduce runtime errors
- **Concurrency safety**: Core structures support concurrent access

## Main Modules

### 1. [Presentation](presentation.md) - Presentations
The top-level facade for a presentation, providing core functionality such as creation, saving, and import/export.

### 2. [Slide](slide.md) - Slides
Slide manipulation interface supporting the addition of text, images, tables, shapes, and other elements.

### 3. [Component](component.md) - Component System
Reusable rendering components with support for custom components and the composite pattern.

### 4. [Color](color.md) - Color System
A complete color handling solution supporting RGB, theme colors, transparency, and more.

### 5. [Media](media.md) - Media Management
Management and deduplication of media assets including images, video, and audio.

### 6. [Template](template.md) - Template System
Template loading, caching, and management with support for embedded templates.

## Quick Start

### Installation

```bash
go get github.com/hurtener/pptx-go/pptx
```

### Creating a Presentation

```go
package main

import (
    "github.com/hurtener/pptx-go/pptx"
)

func main() {
    // Create a blank presentation
    pres := pptx.New()

    // Add a slide
    slide := pres.AddSlide()

    // Add a text box
    slide.AddTextBox(100, 100, 400, 50, "Hello, World!")

    // Save the file
    pres.Save("output.pptx")
}
```

### Creating from a Template

```go
// Use the default template
pres, err := pptx.NewWithTemplate(pptx.TemplateDefault)
if err != nil {
    panic(err)
}

// Use the blank template
pres, err := pptx.NewWithTemplate(pptx.TemplateBlank)
```

### Adding an Image

```go
slide := pres.AddSlide()

// Add an image from a file
pic, err := slide.AddPictureFromFile(100, 100, 400, 300, "image.png")
if err != nil {
    panic(err)
}

// Add an image from bytes
data, _ := os.ReadFile("logo.png")
pic, err = slide.AddPictureFromBytes(100, 100, 200, 150, "logo.png", data)
```

### Adding a Table

```go
slide := pres.AddSlide()

// Add a 3x4 table
table := slide.AddTable(100, 100, 600, 400, 3, 4)

// Set cell text
slide.SetTableCellText(table, 0, 0, "Header 1")
slide.SetTableCellText(table, 0, 1, "Header 2")
```

### Using the Component System

```go
// Create a custom component
type TitleComponent struct {
    Text string
    X, Y int
}

func (t *TitleComponent) Render(ctx *pptx.SlideContext) error {
    // Use SlideContext capabilities to render the component
    return nil
}

// Add the component to a slide
slide.AddComponent(&TitleComponent{
    Text: "My Title",
    X:    100,
    Y:    50,
})
```

## Units

This package primarily uses two units of measurement:

1. **Pixels (px)**: Most high-level APIs use pixels, based on 96 DPI
2. **EMU**: The unit used internally by OOXML (1 px = 9525 EMU)

Conversion functions:
- `PxToEMU(px int) int` - pixels to EMU
- `EMUToPx(emu int) int` - EMU to pixels

## Standard Slide Sizes

```go
// 16:9 widescreen (1280 x 720 px)
pptx.SlideSize16x9

// 4:3 standard (960 x 720 px)
pptx.SlideSize4x3

// 16:10 wide (1280 x 800 px)
pptx.SlideSize16x10
```

## Architecture Layers

```
┌─────────────────────────────────────┐
│         pptx (High-Level API)        │
│  Presentation, Slide, Component     │
├─────────────────────────────────────┤
│       slide (Business Building)     │
│        SlideBuilder, Helper         │
├─────────────────────────────────────┤
│         parts (XML Structure)       │
│    SlidePart, MasterPart, etc.      │
├─────────────────────────────────────┤
│          opc (Low-Level Package)    │
│      Package, Relationships         │
└─────────────────────────────────────┘
```

## Further Documentation

- [Presentation](presentation.md)
- [Slide](slide.md)
- [Component System](component.md)
- [Color System](color.md)
- [Media Management](media.md)
- [Template System](template.md)
