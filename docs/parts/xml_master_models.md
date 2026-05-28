# XML Master Models — Interface Documentation

> XML struct definitions used when parsing masters and layouts

---

## Base Types

### XMLOffset

Offset struct, corresponding to XML: `<a:off x="..." y="..."/>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `X` | `x` | `int64` | X coordinate |
| `Y` | `y` | `int64` | Y coordinate |

#### Methods

```go
func (o *XMLOffset) IsValid() bool  // Returns true if valid (x and y attributes must be present)
func (o *XMLOffset) IsZero() bool   // Returns true if zero value
```

---

### XMLExtents

Extents struct, corresponding to XML: `<a:ext cx="..." cy="..."/>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Cx` | `cx` | `int64` | Width (EMU units) |
| `Cy` | `cy` | `int64` | Height (EMU units) |

#### Methods

```go
func (e *XMLExtents) IsValid() bool  // Returns true if valid (OpenXML spec requires cx and cy to be positive)
func (e *XMLExtents) IsZero() bool   // Returns true if zero value
```

---

### XMLTransform

2D transform struct, corresponding to XML: `<a:xfrm>...</a:xfrm>`

| Field | Type | Description |
|------|------|------|
| `Off` | `*XMLOffset` | Position offset |
| `Ext` | `*XMLExtents` | Size extents |

---

### XMLPlaceholder

Placeholder struct, corresponding to XML: `<p:ph type="..." idx="..."/>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Type` | `type` | `string` | Placeholder type |
| `Idx` | `idx` | `string` | Placeholder index |

---

## Non-Visual Properties

### XMLCNvPr

Common non-visual properties, corresponding to XML: `<p:cNvPr id="..." name="..."/>`

| Field | XML Attribute | Type |
|------|----------|------|
| `ID` | `id` | `int` |
| `Name` | `name` | `string` |

---

### XMLNvPr

Non-visual properties, corresponding to XML: `<p:nvPr>...</p:nvPr>`

| Field | Type | Description |
|------|------|------|
| `Ph` | `*XMLPlaceholder` | Placeholder definition (if present) |

---

### XMLNvSpPr

Non-visual shape properties, corresponding to XML: `<p:nvSpPr>...</p:nvSpPr>`

| Field | Type | Description |
|------|------|------|
| `CNvPr` | `*XMLCNvPr` | Common non-visual properties |
| `NvPr` | `*XMLNvPr` | Non-visual properties |

---

### XMLSpPr

Visual shape properties, corresponding to XML: `<p:spPr>...</p:spPr>`

| Field | Type | Description |
|------|------|------|
| `Xfrm` | `*XMLTransform` | Transform information |

---

## Group Shapes

### XMLNvGrpSpPr

Non-visual group properties, corresponding to XML: `<p:nvGrpSpPr>...</p:nvGrpSpPr>`

| Field | Type |
|------|------|
| `CNvPr` | `*XMLCNvPr` |
| `CNvGrpSpPr` | `*XMLCNvGrpSpPr` |

---

### XMLCNvGrpSpPr

Group shape common non-visual properties, corresponding to XML: `<p:cNvGrpSpPr>...</p:cNvGrpSpPr>`

---

### XMLGrpSpPr

Group shape properties, corresponding to XML: `<p:grpSpPr>...</p:grpSpPr>`

| Field | Type | Description |
|------|------|------|
| `Xfrm` | `*XMLTransform` | Group transform |

---

### XMLGroupShape

Group shape, corresponding to XML: `<p:grpSp>...</p:grpSp>`

| Field | Type | Description |
|------|------|------|
| `NvGrpSpPr` | `*XMLNvGrpSpPr` | Non-visual group properties |
| `GrpSpPr` | `*XMLGrpSpPr` | Group shape properties |
| `Shapes` | `[]XMLShape` | Child shape list |

---

## Background

### XMLBackground

Background struct, corresponding to XML: `<p:bg>...</p:bg>`

| Field | Type | Description |
|------|------|------|
| `BgPr` | `*XMLBackgroundPr` | Background properties |
| `BgRef` | `*XMLBackgroundRef` | Background reference |

---

### XMLBackgroundRef

Background reference, corresponding to XML: `<p:bgRef idx="..."><a:schemeClr val="..."/>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Idx` | `idx` | `string` | Background index |
| `Clr` | `schemeClr` | `*XMLSchemeColor` | Theme color |

---

### XMLBackgroundPr

Background properties, corresponding to XML: `<p:bgPr>...</p:bgPr>`

| Field | Type | Description |
|------|------|------|
| `Fill` | `*XMLFillProperties` | Fill properties |

---

## Fill

### XMLFillProperties

Fill properties (union type), corresponding to XML: `<a:solidFill>` / `<a:gradFill>` / `<a:blipFill>`, etc.

| Field | XML Element | Type |
|------|----------|------|
| `SolidFill` | `a:solidFill` | `*XMLSolidFill` |
| `GradFill` | `a:gradFill` | `*XMLGradFill` |
| `BlipFill` | `a:blipFill` | `*XMLBlipFill` |
| `NoFill` | `a:noFill` | `*struct{}` |

---

### XMLSolidFill

Solid fill, corresponding to XML: `<a:solidFill>...</a:solidFill>`

| Field | Type | Description |
|------|------|------|
| `SrgbClr` | `*XMLSRgbColor` | RGB color |
| `SchemeClr` | `*XMLSchemeColor` | Theme color |

---

### XMLSRgbColor

RGB color, corresponding to XML: `<a:srgbClr val="..."/>`

| Field | XML Attribute | Type |
|------|----------|------|
| `Val` | `val` | `string` |

---

### XMLSchemeColor

Theme color, corresponding to XML: `<a:schemeClr val="..."/>`

| Field | XML Attribute | Type |
|------|----------|------|
| `Val` | `val` | `string` |

---

### XMLGradFill

Gradient fill, corresponding to XML: `<a:gradFill>...</a:gradFill>`

| Field | Type | Description |
|------|------|------|
| `GsLst` | `*XMLGradientStopList` | Gradient stop list |
| `Lin` | `*XMLLinearGradient` | Linear gradient |

---

### XMLGradientStopList

Gradient stop list, corresponding to XML: `<a:gsLst>...</a:gsLst>`

| Field | Type | Description |
|------|------|------|
| `Stops` | `[]XMLGradientStop` | Gradient stops |

---

### XMLGradientStop

Gradient stop, corresponding to XML: `<a:gs pos="...">...</a:gs>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Pos` | `pos` | `int64` | Position |
| `SolidFill` | `a:solidFill` | `*XMLSolidFill` | Color |

---

### XMLLinearGradient

Linear gradient, corresponding to XML: `<a:lin ang="..." scaled="..."/>`

| Field | XML Attribute | Type |
|------|----------|------|
| `Ang` | `ang` | `int64` |
| `Scaled` | `scaled` | `bool` |

---

### XMLBlipFill

Picture fill, corresponding to XML: `<a:blipFill>...</a:blipFill>`

| Field | Type | Description |
|------|------|------|
| `Blip` | `*XMLBlip` | Picture reference |

---

### XMLBlip

Picture reference, corresponding to XML: `<a:blip r:embed="..."/>`

| Field | XML Attribute | Type |
|------|----------|------|
| `Embed` | `r:embed` | `string` |

---

## Shapes

### XMLShape

Shape struct, corresponding to XML: `<p:sp>...</p:sp>`

| Field | Type | Description |
|------|------|------|
| `NvSpPr` | `*XMLNvSpPr` | Non-visual shape properties |
| `SpPr` | `*XMLSpPr` | Visual shape properties |

---

### XMLShapeTree

Shape tree struct, corresponding to XML: `<p:spTree>...</p:spTree>`

| Field | Type | Description |
|------|------|------|
| `NvGrpSpPr` | `*XMLNvGrpSpPr` | Non-visual group properties |
| `GrpSpPr` | `*XMLGrpSpPr` | Group shape properties |
| `Shapes` | `[]XMLShape` | Shape list |
| `GroupShapes` | `[]XMLGroupShape` | Group shape list |

---

## Slide Structures

### XMLCommonSlideData

Common slide data, corresponding to XML: `<p:cSld>...</p:cSld>`

| Field | Type | Description |
|------|------|------|
| `Bg` | `*XMLBackground` | Background |
| `SpTree` | `*XMLShapeTree` | Shape tree |

---

### XMLSlideLayout

Slide layout, corresponding to XML: `<p:sldLayout>...</p:sldLayout>`

| Field | Type | Description |
|------|------|------|
| `XmlnsA` | `string` | DrawingML namespace |
| `XmlnsR` | `string` | Relationships namespace |
| `XmlnsP` | `string` | PresentationML namespace |
| `CSld` | `*XMLCommonSlideData` | Common slide data |

---

### XMLSlideMaster

Slide master, corresponding to XML: `<p:sldMaster>...</p:sldMaster>`

| Field | Type | Description |
|------|------|------|
| `XmlnsA` | `string` | DrawingML namespace |
| `XmlnsR` | `string` | Relationships namespace |
| `XmlnsP` | `string` | PresentationML namespace |
| `CSld` | `*XMLCommonSlideData` | Common slide data |

---

## Lines

### XLineProperties

Line properties, corresponding to XML: `<a:ln w="...">`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `Width` | `w` | `int` | Line width |
| `SolidFill` | `a:solidFill` | `*XPresetFill` | Fill |

---

### XPresetFill

Preset fill, corresponding to XML: `<a:solidFill>` or `<a:schemeClr>`

| Field | Type | Description |
|------|------|------|
| `SrgbClr` | `*XSrgbClr` | RGB color |
| `SchemeClr` | `*XSchemeClr` | Theme color |

---

## Text Properties

### XTextProperties

Text properties, corresponding to `<a:rPr>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `FontSize` | `sz` | `int` | Font size (hundredths of a point) |
| `Bold` | `b` | `bool` | Bold |
| `Italic` | `i` | `bool` | Italic |
| `Underline` | `u` | `string` | Underline |
| `FontFace` | `typeface` | `string` | Font face |
| `Color` | `solidFill` | `string` | Color |

---

## Tables

### XTableGrid

Table grid, corresponding to XML: `<a:tblGrid>`

| Field | Type | Description |
|------|------|------|
| `GridCols` | `[]XTableColumn` | Column definitions |

---

### XTableColumn

Table column, corresponding to XML: `<a:gridCol w="..."/>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `W` | `w` | `int` | Column width |

---

## Miscellaneous

### XFillRectProperties

Fill rectangle properties, corresponding to XML: `<a:fillRect/>`

### XStretchProperties

Stretch fill properties, corresponding to XML: `<a:stretch><a:fillRect/></a:stretch>`

| Field | Type | Description |
|------|------|------|
| `FillRect` | `*XFillRectProperties` | Fill rectangle |

### XGraphic

Graphic, corresponding to XML: `<a:graphic>`

| Field | Type | Description |
|------|------|------|
| `Table` | `*XTable` | Table |

### XNonVisualGraphicFrame

Graphic frame non-visual properties, corresponding to `<p:nvGraphicFramePr>`

| Field | Type | Description |
|------|------|------|
| `CNvPr` | `*XNvCxnSpPr` | Common non-visual properties |
| `CNvGraphicFramePr` | `*XNvGraphicFramePr` | Graphic frame non-visual properties |

### XNvGraphicFramePr

Graphic frame non-visual properties, corresponding to `<p:cNvGraphicFramePr>`

| Field | Type | Description |
|------|------|------|
| `CNvPr` | `*XNvPr` | Non-visual properties |

### XNvCxnSpPr

Connector shape non-visual properties, corresponding to `<p:cNvCxnSpPr>`

| Field | XML Attribute | Type |
|------|----------|------|
| `ID` | `id` | `int` |
| `Name` | `name` | `string` |

### XNvPicPr

Picture non-visual properties, corresponding to `<p:cNvPicPr>`

| Field | Type | Description |
|------|------|------|
| `CNvPr` | `*XNvPr` | Non-visual properties |

### XTextParagraphList

Text paragraph list, corresponding to `<a:lstStyle/>`

### XOutputOptions

Output options, corresponding to `<p:outputOptions>`

| Field | XML Attribute | Type | Description |
|------|----------|------|------|
| `UsePrintFml` | `usePrintFml` | `*bool` | Use print format |
| `CloneLinkedObjs` | `cloneLinkedObjs` | `*bool` | Clone linked objects |
