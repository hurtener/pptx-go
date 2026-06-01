// Package pptx provides a high-level API for authoring PPTX files.
package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// SlideBuilder — high-level slide construction helper
// ============================================================================
//
// Provides a convenient API for building slide content:
//   - adding text boxes
//   - adding auto shapes
//   - adding images
//   - adding tables
//   - managing media relationships
// ============================================================================

// SlideBuilder wraps a SlidePart and provides helper methods for building
// slide content.
type SlideBuilder struct {
	slide *slide.SlidePart
}

// NewSlideBuilder creates a SlideBuilder for the given SlidePart.
func NewSlideBuilder(slide *slide.SlidePart) *SlideBuilder {
	return &SlideBuilder{
		slide: slide,
	}
}

// Slide returns the underlying SlidePart.
func (b *SlideBuilder) Slide() *slide.SlidePart {
	return b.slide
}

// ============================================================================
// Shape addition methods
// ============================================================================

// AddTextBox adds a text box to the slide.
// x, y, cx, cy are the position and size in EMU; text is the initial content.
func (b *SlideBuilder) AddTextBox(x, y, cx, cy int, text string) *slide.XSp {
	sp := &slide.XSp{
		NonVisual: slide.XNonVisualDrawingShape{
			CNvPr: &slide.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("TextBox %d", b.slide.CurrentShapeID()),
			},
			CNvSpPr: &slide.XNvSpPr{},
		},
		ShapeProperties: &slide.XShapeProperties{
			Transform2D: &slide.XTransform2D{
				Offset: &slide.XOv2DrOffset{X: x, Y: y},
				Extent: &slide.XOv2DrExtent{Cx: cx, Cy: cy},
			},
			PresetGeom: &slide.XPresetGeometry{Prst: "rect", AvLst: &slide.XAvLst{}},
		},
		TextBody: &slide.XTextBody{
			BodyPr:   &slide.XBodyPr{},
			LstStyle: &slide.XTextParagraphList{},
			Paragraphs: []slide.XTextParagraph{
				{Content: []any{&slide.XTextRun{Text: text}}},
			},
		},
	}

	b.slide.AppendShapeChild(sp)
	return sp
}

// AddAutoShape adds an auto shape to the slide.
// x, y, cx, cy are the position and size in EMU.
// presetID is the preset shape type (e.g. "rectangle", "ellipse", "roundRect").
func (b *SlideBuilder) AddAutoShape(x, y, cx, cy int, presetID string) *slide.XSp {
	sp := &slide.XSp{
		NonVisual: slide.XNonVisualDrawingShape{
			CNvPr: &slide.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("%s %d", presetID, b.slide.CurrentShapeID()),
			},
			CNvSpPr: &slide.XNvSpPr{},
		},
		ShapeProperties: &slide.XShapeProperties{
			Transform2D: &slide.XTransform2D{
				Offset: &slide.XOv2DrOffset{X: x, Y: y},
				Extent: &slide.XOv2DrExtent{Cx: cx, Cy: cy},
			},
			PresetGeom: &slide.XPresetGeometry{Prst: presetID, AvLst: &slide.XAvLst{}},
		},
	}

	b.slide.AppendShapeChild(sp)
	return sp
}

// AddCustomShape adds a shape with custom path geometry (an icon glyph).
// x, y, cx, cy are the position and size in EMU; geom is the translated
// <a:custGeom>. The caller applies fill/line to the returned shape's properties.
func (b *SlideBuilder) AddCustomShape(x, y, cx, cy int, geom *slide.XCustomGeometry) *slide.XSp {
	sp := &slide.XSp{
		NonVisual: slide.XNonVisualDrawingShape{
			CNvPr: &slide.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("Icon %d", b.slide.CurrentShapeID()),
			},
			CNvSpPr: &slide.XNvSpPr{},
		},
		ShapeProperties: &slide.XShapeProperties{
			Transform2D: &slide.XTransform2D{
				Offset: &slide.XOv2DrOffset{X: x, Y: y},
				Extent: &slide.XOv2DrExtent{Cx: cx, Cy: cy},
			},
			CustomGeom: geom,
		},
	}

	b.slide.AppendShapeChild(sp)
	return sp
}

// AddPicture adds an image to the slide.
// x, y, cx, cy are the position and size in EMU; imageRId is the image
// relationship ID.
func (b *SlideBuilder) AddPicture(x, y, cx, cy int, imageRId string) *slide.XPicture {
	pic := &slide.XPicture{
		NonVisual: slide.XNonVisualDrawingPic{
			CNvPr: &slide.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("Picture %d", b.slide.CurrentShapeID()),
			},
			CNvPicPr: &slide.XNvPicPr{},
		},
		BlipFill: &slide.XBlipFillProperties{
			Blip: &slide.XBlip{
				Embed: imageRId,
			},
			Stretch: &slide.XStretchProperties{FillRect: &slide.XFillRectProperties{}},
		},
		ShapeProperties: &slide.XShapeProperties{
			Transform2D: &slide.XTransform2D{
				Offset: &slide.XOv2DrOffset{X: x, Y: y},
				Extent: &slide.XOv2DrExtent{Cx: cx, Cy: cy},
			},
			// A picture needs a geometry to define the region the blip fills;
			// without it many renderers (Quick Look, Keynote, LibreOffice) draw
			// nothing. PowerPoint always emits <a:prstGeom prst="rect">.
			PresetGeom: &slide.XPresetGeometry{Prst: "rect", AvLst: &slide.XAvLst{}},
		},
	}

	b.slide.AppendShapeChild(pic)
	return pic
}

// AddTable adds a table to the slide.
// x, y, cx, cy are the position and size in EMU; rows and cols are the table
// dimensions.
func (b *SlideBuilder) AddTable(x, y, cx, cy, rows, cols int) *slide.XGraphicFrame {
	// compute per-column width
	cellW := cx / cols

	// build the table grid
	gridCols := make([]slide.XTableColumn, cols)
	for i := range gridCols {
		gridCols[i] = slide.XTableColumn{W: cellW}
	}

	// compute per-row height
	rowH := cy / rows

	// build the table rows
	tableRows := make([]slide.XTableRow, rows)
	for r := range tableRows {
		cells := make([]slide.XTableCell, cols)
		for c := range cells {
			cells[c] = slide.XTableCell{
				TextBody: &slide.XTextBody{
					BodyPr:   &slide.XBodyPr{},
					LstStyle: &slide.XTextParagraphList{},
					Paragraphs: []slide.XTextParagraph{
						{Content: []any{&slide.XTextRun{Text: ""}}},
					},
				},
			}
		}
		tableRows[r] = slide.XTableRow{H: rowH, Cells: cells}
	}

	table := slide.XTable{
		Grid: &slide.XTableGrid{GridCols: gridCols},
		Rows: tableRows,
	}

	gf := &slide.XGraphicFrame{
		NonVisual: slide.XNonVisualGraphicFrame{
			CNvPr: &slide.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("Table %d", b.slide.CurrentShapeID()),
			},
			CNvGraphicFramePr: &slide.XNvGraphicFramePr{},
		},
		Transform2D: &slide.XTransform2D{
			Offset: &slide.XOv2DrOffset{X: x, Y: y},
			Extent: &slide.XOv2DrExtent{Cx: cx, Cy: cy},
		},
		Graphic: &slide.XGraphic{
			GraphicData: &slide.XGraphicData{
				URI:   slide.TableGraphicDataURI,
				Table: &table,
			},
		},
	}

	b.slide.AppendShapeChild(gf)
	return gf
}

// SetTableCellText sets the text of the cell at (row, col) in the given table.
func (b *SlideBuilder) SetTableCellText(gf *slide.XGraphicFrame, row, col int, text string) {
	if gf == nil || gf.Graphic == nil || gf.Graphic.GraphicData == nil || gf.Graphic.GraphicData.Table == nil {
		return
	}
	table := gf.Graphic.GraphicData.Table
	if row < 0 || row >= len(table.Rows) || col < 0 || col >= len(table.Rows[row].Cells) {
		return
	}
	para := &table.Rows[row].Cells[col].TextBody.Paragraphs[0]
	if len(para.Content) == 0 {
		para.Content = []any{&slide.XTextRun{Text: text}}
		return
	}
	if run, ok := para.Content[0].(*slide.XTextRun); ok {
		run.Text = text
	}
}

// GetOrAddPicture adds an image to the slide by URI and returns its XPicture.
// The image relationship is created automatically.
func (b *SlideBuilder) GetOrAddPicture(x, y, cx, cy int, imageURI string) *slide.XPicture {
	rId := b.GetImageRId(imageURI)
	return b.AddPicture(x, y, cx, cy, rId)
}

// ============================================================================
// Relationship management
// ============================================================================

// AddImage adds an image relationship and returns its relationship ID.
func (b *SlideBuilder) AddImage(targetURI string) string {
	return b.slide.AddImageRel(targetURI)
}

// AddMedia adds a media relationship and returns its relationship ID.
func (b *SlideBuilder) AddMedia(targetURI string) string {
	return b.slide.AddMediaRel(targetURI)
}

// AddChart adds a chart relationship and returns its relationship ID.
func (b *SlideBuilder) AddChart(targetURI string) string {
	return b.slide.AddChartRel(targetURI)
}

// HasImage reports whether an image relationship for targetURI already exists.
func (b *SlideBuilder) HasImage(targetURI string) bool {
	rels := b.slide.Relationships()
	target := opc.NewPackURI(targetURI)
	return rels.GetByTarget(target) != nil
}

// HasMedia reports whether a media relationship for targetURI already exists.
func (b *SlideBuilder) HasMedia(targetURI string) bool {
	rels := b.slide.Relationships()
	target := opc.NewPackURI(targetURI)
	return rels.GetByTarget(target) != nil
}

// GetImageRId returns the relationship ID for the given image URI, creating the
// relationship if it does not yet exist.
func (b *SlideBuilder) GetImageRId(targetURI string) string {
	return b.slide.AddImageRel(targetURI)
}

// GetMediaRId returns the relationship ID for the given media URI, creating the
// relationship if it does not yet exist.
func (b *SlideBuilder) GetMediaRId(targetURI string) string {
	return b.slide.AddMediaRel(targetURI)
}

// GetChartRId returns the relationship ID for the given chart URI, creating the
// relationship if it does not yet exist.
func (b *SlideBuilder) GetChartRId(targetURI string) string {
	return b.slide.AddChartRel(targetURI)
}

// GetRelationshipURI returns the target URI for the relationship with the given
// ID, or an empty string if it does not exist.
func (b *SlideBuilder) GetRelationshipURI(rId string) string {
	rels := b.slide.Relationships()
	rel := rels.Get(rId)
	if rel != nil {
		return rel.TargetURI().URI()
	}
	return ""
}
