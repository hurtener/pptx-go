// Package pptx provides a high-level API for authoring PPTX files.
package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/opc"
	"github.com/hurtener/pptx-go/parts"
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
	slide *parts.SlidePart
}

// NewSlideBuilder creates a SlideBuilder for the given SlidePart.
func NewSlideBuilder(slide *parts.SlidePart) *SlideBuilder {
	return &SlideBuilder{
		slide: slide,
	}
}

// Slide returns the underlying SlidePart.
func (b *SlideBuilder) Slide() *parts.SlidePart {
	return b.slide
}

// ============================================================================
// Shape addition methods
// ============================================================================

// AddTextBox adds a text box to the slide.
// x, y, cx, cy are the position and size in EMU; text is the initial content.
func (b *SlideBuilder) AddTextBox(x, y, cx, cy int, text string) *parts.XSp {
	sp := &parts.XSp{
		NonVisual: parts.XNonVisualDrawingShape{
			CNvPr: &parts.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("TextBox %d", b.slide.CurrentShapeID()),
			},
			CNvSpPr: &parts.XNvSpPr{},
		},
		ShapeProperties: &parts.XShapeProperties{
			Transform2D: &parts.XTransform2D{
				Offset: &parts.XOv2DrOffset{X: x, Y: y},
				Extent: &parts.XOv2DrExtent{Cx: cx, Cy: cy},
			},
		},
		TextBody: &parts.XTextBody{
			BodyPr:   &parts.XBodyPr{},
			LstStyle: &parts.XTextParagraphList{},
			Paragraphs: []parts.XTextParagraph{
				{
					TextRuns: []parts.XTextRun{
						{Text: text},
					},
				},
			},
		},
	}

	b.slide.AppendShapeChild(sp)
	return sp
}

// AddAutoShape adds an auto shape to the slide.
// x, y, cx, cy are the position and size in EMU.
// presetID is the preset shape type (e.g. "rectangle", "ellipse", "roundRect").
func (b *SlideBuilder) AddAutoShape(x, y, cx, cy int, presetID string) *parts.XSp {
	sp := &parts.XSp{
		NonVisual: parts.XNonVisualDrawingShape{
			CNvPr: &parts.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("%s %d", presetID, b.slide.CurrentShapeID()),
			},
			CNvSpPr: &parts.XNvSpPr{},
		},
		ShapeProperties: &parts.XShapeProperties{
			Transform2D: &parts.XTransform2D{
				Offset: &parts.XOv2DrOffset{X: x, Y: y},
				Extent: &parts.XOv2DrExtent{Cx: cx, Cy: cy},
			},
		},
		ShapePreset: presetID,
	}

	b.slide.AppendShapeChild(sp)
	return sp
}

// AddPicture adds an image to the slide.
// x, y, cx, cy are the position and size in EMU; imageRId is the image
// relationship ID.
func (b *SlideBuilder) AddPicture(x, y, cx, cy int, imageRId string) *parts.XPicture {
	pic := &parts.XPicture{
		NonVisual: parts.XNonVisualDrawingPic{
			CNvPr: &parts.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("Picture %d", b.slide.CurrentShapeID()),
			},
			CNvPicPr: &parts.XNvPicPr{},
		},
		BlipFill: &parts.XBlipFillProperties{
			Blip: &parts.XBlip{
				Embed: imageRId,
			},
			Stretch: &parts.XStretchProperties{},
		},
		ShapeProperties: &parts.XShapeProperties{
			Transform2D: &parts.XTransform2D{
				Offset: &parts.XOv2DrOffset{X: x, Y: y},
				Extent: &parts.XOv2DrExtent{Cx: cx, Cy: cy},
			},
		},
	}

	b.slide.AppendShapeChild(pic)
	return pic
}

// AddTable adds a table to the slide.
// x, y, cx, cy are the position and size in EMU; rows and cols are the table
// dimensions.
func (b *SlideBuilder) AddTable(x, y, cx, cy, rows, cols int) *parts.XGraphicFrame {
	// compute per-column width
	cellW := cx / cols

	// build the table grid
	gridCols := make([]parts.XTableColumn, cols)
	for i := range gridCols {
		gridCols[i] = parts.XTableColumn{W: cellW}
	}

	// build the table rows
	tableRows := make([]parts.XTableRow, rows)
	for r := range tableRows {
		cells := make([]parts.XTableCell, cols)
		for c := range cells {
			cells[c] = parts.XTableCell{
				TextBody: &parts.XTextBody{
					BodyPr:   &parts.XBodyPr{},
					LstStyle: &parts.XTextParagraphList{},
					Paragraphs: []parts.XTextParagraph{
						{TextRuns: []parts.XTextRun{{Text: ""}}},
					},
				},
			}
		}
		tableRows[r] = parts.XTableRow{GridSpan: 1, Cells: cells}
	}

	table := parts.XTable{
		Grid: &parts.XTableGrid{GridCols: gridCols},
		Rows: tableRows,
	}

	gf := &parts.XGraphicFrame{
		NonVisual: parts.XNonVisualGraphicFrame{
			CNvPr: &parts.XNvCxnSpPr{
				ID:   int(b.slide.AllocateShapeID()),
				Name: fmt.Sprintf("Table %d", b.slide.CurrentShapeID()),
			},
			CNvGraphicFramePr: &parts.XNvGraphicFramePr{},
		},
		Graphic: &parts.XGraphic{
			Table: &table,
		},
		Transform2D: &parts.XTransform2D{
			Offset: &parts.XOv2DrOffset{X: x, Y: y},
			Extent: &parts.XOv2DrExtent{Cx: cx, Cy: cy},
		},
	}

	b.slide.AppendShapeChild(gf)
	return gf
}

// SetTableCellText sets the text of the cell at (row, col) in the given table.
func (b *SlideBuilder) SetTableCellText(gf *parts.XGraphicFrame, row, col int, text string) {
	if gf == nil || gf.Graphic == nil || gf.Graphic.Table == nil {
		return
	}
	table := gf.Graphic.Table
	if row < 0 || row >= len(table.Rows) || col < 0 || col >= len(table.Rows[row].Cells) {
		return
	}
	table.Rows[row].Cells[col].TextBody.Paragraphs[0].TextRuns[0].Text = text
}

// GetOrAddPicture adds an image to the slide by URI and returns its XPicture.
// The image relationship is created automatically.
func (b *SlideBuilder) GetOrAddPicture(x, y, cx, cy int, imageURI string) *parts.XPicture {
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
