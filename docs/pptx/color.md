# Color - Color System

The color system provides a complete color handling solution with support for RGB, theme colors, transparency, and more.

## Type Definitions

### Color

Color structure.

```go
type Color struct {
    // Type is the color type
    Type ColorType
    // RGB hex value (6 characters, e.g. "FF0000", no # prefix)
    RGB string
    // Scheme is the theme color name (e.g. "accent1")
    Scheme string
    // Alpha is the transparency value (0-100000, OOXML standard)
    // 100000 = 100% opaque, 0 = fully transparent
    Alpha int
    // IsValid indicates whether the color is valid
    IsValid bool
}
```

### ColorType

Color type.

```go
type ColorType int

const (
    // ColorTypeRGB is an RGB color
    ColorTypeRGB ColorType = iota
    // ColorTypeScheme is a theme color
    ColorTypeScheme
    // ColorTypeInvalid is an invalid color
    ColorTypeInvalid
)
```

## Color Constants

### Alpha Constants

Alpha constants (OOXML range: 0-100000).

```go
const (
    // AlphaOpaque is fully opaque
    AlphaOpaque = 100000
    // AlphaTransparent is fully transparent
    AlphaTransparent = 0
    // AlphaDefault is the default transparency (100%)
    AlphaDefault = 100000
)
```

### Theme Color Names

```go
const (
    // Background colors
    SchemeBg1 = "bg1" // Background color 1 (typically white or light)
    SchemeBg2 = "bg2" // Background color 2
    SchemeFg1 = "fg1" // Foreground/text color 1 (typically black or dark)
    SchemeFg2 = "fg2" // Foreground/text color 2

    // Accent colors
    SchemeAccent1 = "accent1" // Accent color 1
    SchemeAccent2 = "accent2" // Accent color 2
    SchemeAccent3 = "accent3" // Accent color 3
    SchemeAccent4 = "accent4" // Accent color 4
    SchemeAccent5 = "accent5" // Accent color 5
    SchemeAccent6 = "accent6" // Accent color 6

    // Hyperlink colors
    SchemeHlink    = "hlink"    // Hyperlink color
    SchemeFolHlink = "folHlink" // Followed hyperlink color

    // Special colors
    SchemePhClr = "phClr" // Placeholder color
    SchemeTx1   = "tx1"   // Text color 1
    SchemeTx2   = "tx2"   // Text color 2
)
```

### Preset Color Constants

Preset color constants (6-character hex, no # prefix).

```go
var (
    // Basic colors
    ColorBlack   = RGBColor("000000")
    ColorWhite   = RGBColor("FFFFFF")
    ColorRed     = RGBColor("FF0000")
    ColorGreen   = RGBColor("00FF00")
    ColorBlue    = RGBColor("0000FF")
    ColorYellow  = RGBColor("FFFF00")
    ColorCyan    = RGBColor("00FFFF")
    ColorMagenta = RGBColor("FF00FF")

    // Transparent color
    ColorTransparent = Color{Type: ColorTypeRGB, RGB: "000000", Alpha: AlphaTransparent, IsValid: true}

    // Common UI colors
    ColorGray      = RGBColor("808080")
    ColorLightGray = RGBColor("C0C0C0")
    ColorDarkGray  = RGBColor("404040")
    ColorOrange    = RGBColor("FFA500")
    ColorPurple    = RGBColor("800080")
    ColorPink      = RGBColor("FFC0CB")
    ColorBrown     = RGBColor("A52A2A")
    ColorNavy      = RGBColor("000080")
    ColorTeal      = RGBColor("008080")
    ColorOlive     = RGBColor("808000")
    ColorMaroon    = RGBColor("800000")
    ColorLime      = RGBColor("00FF00")
    ColorAqua      = RGBColor("00FFFF")
    ColorSilver    = RGBColor("C0C0C0")
    ColorGold      = RGBColor("FFD700")
)
```

## Constructors

### RGBColor

Creates an RGB color (no alpha).

```go
func RGBColor(hex string) Color
```

**Parameters:**
- `hex`: 6-character hex RGB (no # prefix)

**Example:**

```go
red := pptx.RGBColor("FF0000")
blue := pptx.RGBColor("0000FF")
```

### RGBAColor

Creates an RGB color with an alpha value.

```go
func RGBAColor(hex string, alpha int) Color
```

**Parameters:**
- `hex`: 6-character hex RGB
- `alpha`: 0-100000 (OOXML standard)

**Example:**

```go
// Semi-transparent red
semiRed := pptx.RGBAColor("FF0000", 50000) // 50% transparent
```

### SchemeColor

Creates a theme color.

```go
func SchemeColor(name string) Color
```

**Example:**

```go
// Use a theme accent color
accent1 := pptx.SchemeColor(pptx.SchemeAccent1)
bgColor := pptx.SchemeColor(pptx.SchemeBg1)
```

### ColorFromRGB

Creates a color from RGB values.

```go
func ColorFromRGB(r, g, b int) Color
```

**Parameters:**
- `r, g, b`: 0-255

**Example:**

```go
red := pptx.ColorFromRGB(255, 0, 0)
```

### ColorFromRGBA

Creates a color from RGBA values.

```go
func ColorFromRGBA(r, g, b, alpha int) Color
```

**Parameters:**
- `r, g, b`: 0-255
- `alpha`: 0-100000 (OOXML standard)

**Example:**

```go
semiRed := pptx.ColorFromRGBA(255, 0, 0, 50000) // 50% transparent
```

### ParseColor

Parses a color string; supports multiple formats.

```go
func ParseColor(s string) Color
```

**Supported formats:**
- `#FF0000` or `FF0000` (6-character hex)
- `#FF0000FF` (8-character hex, last two characters are alpha)
- `rgb(255, 0, 0)` (RGB)
- `rgba(255, 0, 0, 0.5)` (RGBA, alpha 0-1)
- `accent1`, `bg1`, and other theme color names

**Example:**

```go
// Hex
c1 := pptx.ParseColor("#FF0000")
c2 := pptx.ParseColor("FF0000")

// RGB
c3 := pptx.ParseColor("rgb(255, 0, 0)")

// RGBA
c4 := pptx.ParseColor("rgba(255, 0, 0, 0.5)")

// Theme color
c5 := pptx.ParseColor("accent1")
```

## Color Methods

### WithAlpha

Sets the alpha value and returns a new color.

```go
func (c Color) WithAlpha(alpha int) Color
```

**Parameters:**
- `alpha`: 0-100000 (OOXML standard)

**Example:**

```go
red := pptx.RGBColor("FF0000")
semiRed := red.WithAlpha(50000) // 50% transparent
```

### WithAlphaPercent

Sets the alpha value as a percentage and returns a new color.

```go
func (c Color) WithAlphaPercent(percent float64) Color
```

**Parameters:**
- `percent`: 0-100

**Example:**

```go
red := pptx.RGBColor("FF0000")
semiRed := red.WithAlphaPercent(50) // 50% transparent
```

### AlphaPercent

Returns the alpha value as a percentage (0-100).

```go
func (c Color) AlphaPercent() float64
```

### AlphaValue

Returns the alpha value in OOXML format (0-100000).

```go
func (c Color) AlphaValue() int
```

**Usage:** For `<a:alpha val="50000"/>`

### ToRGB

Converts to an RGB hex string (6 characters, no # prefix).

```go
func (c Color) ToRGB() string
```

**Usage:** Conforms to the OOXML specification: `<a:srgbClr val="FF0000"/>`

### ToHex

Converts to a 6-character hex string with a # prefix (for display).

```go
func (c Color) ToHex() string
```

**Example:**

```go
red := pptx.RGBColor("FF0000")
fmt.Println(red.ToHex()) // Output: #FF0000
```

### ToHexA

Converts to an 8-character hex string with a # prefix (includes alpha).

```go
func (c Color) ToHexA() string
```

### ToScheme

Converts to a theme color name.

```go
func (c Color) ToScheme() string
```

### RGBComponents

Returns the RGB components (r, g, b).

```go
func (c Color) RGBComponents() (r, g, b int, ok bool)
```

**Example:**

```go
red := pptx.RGBColor("FF0000")
r, g, b, ok := red.RGBComponents()
// r=255, g=0, b=0, ok=true
```

### String

Returns a string representation of the color.

```go
func (c Color) String() string
```

## Helper Functions

### IsSchemeColor

Checks whether a string is a valid theme color name.

```go
func IsSchemeColor(name string) bool
```

**Example:**

```go
pptx.IsSchemeColor("accent1") // true
pptx.IsSchemeColor("FF0000")  // false
```

### ValidateColor

Validates a color string.

```go
func ValidateColor(s string) ColorValidationResult
```

**Returns:**
- `ColorValidationResult` — the validation result

**Example:**

```go
result := pptx.ValidateColor("#FF0000")
if result.IsValid {
    fmt.Println("Valid color:", result.Color.ToHex())
} else {
    fmt.Println("Invalid color:", result.Message)
}
```

## ColorValidationResult

Color validation result.

```go
type ColorValidationResult struct {
    // IsValid indicates whether the color is valid
    IsValid bool
    // Color is the parsed color
    Color Color
    // Original is the original input string
    Original string
    // Message is the validation message
    Message string
}
```

## ColorMap - Color Mapping Table

A color map maps color names to actual color values.

```go
type ColorMap struct {
    // Has unexported fields.
}
```

### Constructors

```go
// NewColorMap creates a new color map
func NewColorMap() *ColorMap

// DefaultColorMap returns the default color map
func DefaultColorMap() *ColorMap
```

### Methods

```go
// Set adds a color mapping
func (cm *ColorMap) Set(name string, color Color)

// Get retrieves a color mapping
func (cm *ColorMap) Get(name string) (Color, bool)

// All returns all color mappings
func (cm *ColorMap) All() map[string]Color

// Resolve resolves a color (supports names, hex, RGB, and theme colors)
func (cm *ColorMap) Resolve(s string) Color
```

**Example:**

```go
cm := pptx.NewColorMap()

// Set custom colors
cm.Set("primary", pptx.RGBColor("007AFF"))
cm.Set("secondary", pptx.RGBColor("5856D6"))

// Get a color
if primary, ok := cm.Get("primary"); ok {
    fmt.Println("Primary color:", primary.ToHex())
}

// Resolve a color (supports multiple formats)
c1 := cm.Resolve("primary")      // custom name
c2 := cm.Resolve("#FF0000")      // hex
c3 := cm.Resolve("accent1")      // theme color
```

## Usage Examples

### Basic Color Usage

```go
// Create a presentation
pres := pptx.New()
slide := pres.AddSlide()

// Use preset colors
red := pptx.ColorRed
blue := pptx.ColorBlue
green := pptx.ColorGreen

// Use a custom color
custom := pptx.RGBColor("007AFF")
```

### Handling Transparency

```go
// Create semi-transparent colors
semiRed := pptx.RGBColor("FF0000").WithAlpha(50000)     // 50% transparent
semiBlue := pptx.ColorBlue.WithAlphaPercent(30)         // 30% transparent

// Fully transparent
transparent := pptx.ColorTransparent
```

### Theme Colors

```go
// Use theme colors
accent1 := pptx.SchemeColor(pptx.SchemeAccent1)
accent2 := pptx.SchemeColor(pptx.SchemeAccent2)
bgColor := pptx.SchemeColor(pptx.SchemeBg1)

// Theme colors also support transparency
semiAccent1 := accent1.WithAlphaPercent(50)
```

### Color Parsing

```go
// Parse colors from strings
colors := []string{
    "#FF0000",
    "rgb(0, 255, 0)",
    "rgba(0, 0, 255, 0.5)",
    "accent1",
    "bg1",
}

for _, s := range colors {
    c := pptx.ParseColor(s)
    fmt.Printf("%s -> %s (alpha: %.0f%%)\n",
        s, c.ToHex(), c.AlphaPercent())
}
```

### Using the Color Map

```go
// Create a brand color map
brandColors := pptx.NewColorMap()
brandColors.Set("primary", pptx.RGBColor("007AFF"))
brandColors.Set("secondary", pptx.RGBColor("5856D6"))
brandColors.Set("success", pptx.RGBColor("34C759"))
brandColors.Set("warning", pptx.RGBColor("FF9500"))
brandColors.Set("danger", pptx.RGBColor("FF3B30"))

// Use a brand color
primary := brandColors.Resolve("primary")
```

### Using Colors in Shapes

```go
// Add a shape with color
rect := slide.AddRectangle(100, 100, 200, 150)

// Set fill color
rect.SpPr.SolidFill = &parts.XSolidFill{
    SrgbClr: &parts.XSrgbClr{
        Val: pptx.ColorRed.ToRGB(),
    },
}

// Set border color
rect.SpPr.Ln = &parts.XLn{
    SolidFill: &parts.XSolidFill{
        SrgbClr: &parts.XSrgbClr{
            Val: pptx.ColorBlack.ToRGB(),
        },
    },
}
```

### Using Colors in Text

```go
// Add colored text
textBox := slide.AddTextBox(100, 100, 400, 50, "Hello World")

// Set text color
textBox.TxBody.P[0].R[0].RPr.SolidFill = &parts.XSolidFill{
    SrgbClr: &parts.XSrgbClr{
        Val: pptx.ColorBlue.ToRGB(),
    },
}
```

## Unit Conversion

### PxToEMU / EMUToPx

Conversion between pixels and EMU.

```go
// PxToEMU converts pixels to EMU (based on 96 DPI)
func PxToEMU(px int) int

// EMUToPx converts EMU to pixels (based on 96 DPI)
func EMUToPx(emu int) int
```

**Example:**

```go
emu := pptx.PxToEMU(100)  // 914400 / 96 * 100 = 952500
px := pptx.EMUToPx(952500) // 100
```

### EMUsPerPixel Constant

Number of EMUs per pixel (96 DPI).

```go
const (
    // EMUsPerPixel is the number of EMUs per pixel (96 DPI)
    // 1 inch = 914400 EMU
    // 1 inch = 96 pixels (96 DPI)
    // Therefore 1 px = 914400 / 96 = 9525 EMU
    EMUsPerPixel = 914400 / 96 // = 9525
)
```
