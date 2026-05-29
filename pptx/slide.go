// Package pptx provides a high-level interface for working with PPTX files.
package pptx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
	"github.com/hurtener/pptx-go/utils"
)

// ============================================================================
// Slide - high-level slide wrapper
// ============================================================================
//
// Slide is a high-level wrapper around the underlying slide.SlidePart. It
// provides:
//  1. Convenient content-addition methods (text boxes, pictures, tables, etc.)
//  2. Bidirectional association with Presentation
//  3. Automatic media-resource management
//
// Units:
//   - Position and size parameters on the content-addition methods are EMU
//     (English Metric Units, OOXML's canonical integer length, 1 inch =
//     914400 EMU). Compute them with the pptx converters — pptx.In, pptx.Cm,
//     pptx.Pt, pptx.Px — or a pptx.Box.
//   - The coordinate origin (0, 0) is the top-left corner of the slide.
//   - Coordinates may be negative; placing elements outside the canvas
//     boundary is legal in the drawing model.
//
// (The SlideViewport / boundary-check helpers below remain px-based; they are
// a placement utility, not the shape-emission path.)
//
// Example:
//
//	s := pres.AddSlide()
//	s.AddTextBox(int(pptx.In(1)), int(pptx.In(1)), int(pptx.In(5)), int(pptx.In(0.5)), "Hello World")
//	s.AddRectangle(914400, 914400, 2743200, 1371600) // 1in, 1in, 3in, 1.5in
//
// ============================================================================

// ============================================================================
// EMU constants - unit conversion ratios (fixed)
// ============================================================================
//
// Conversion ratio: 1 px = 9525 EMU (based on 96 DPI).
// This ratio is constant and does not change with slide size.
// The coordinate origin (0, 0) is the top-left corner of the slide.

const (
	// EMUsPerPixel is the number of EMUs per pixel at 96 DPI.
	// 1 inch = 914400 EMU, 1 inch = 96 pixels (96 DPI),
	// therefore 1 px = 914400 / 96 = 9525 EMU.
	EMUsPerPixel = 914400 / 96 // = 9525
)

// ============================================================================
// Standard slide sizes - in px (based on 96 DPI)
// ============================================================================
//
// These sizes help callers understand the canvas dimensions.
// Callers are responsible for adjusting coordinates to fit the chosen size.
//
// Note: px coordinates may be negative; elements may be placed outside the
// canvas boundary.

// SlideSize represents a slide's dimensions.
type SlideSize struct {
	Width  int // width (px)
	Height int // height (px)
}

// Standard slide size variables.
var (
	// SlideSize16x9 is the widescreen slide size (16:9).
	// Width: 1280 px (13.333 inches), Height: 720 px (7.5 inches).
	SlideSize16x9 = SlideSize{Width: 1280, Height: 720}

	// SlideSize4x3 is the standard slide size (4:3).
	// Width: 960 px (10 inches), Height: 720 px (7.5 inches).
	SlideSize4x3 = SlideSize{Width: 960, Height: 720}

	// SlideSize16x10 is the wide slide size (16:10).
	// Width: 1280 px (13.333 inches), Height: 800 px (8.333 inches).
	SlideSize16x10 = SlideSize{Width: 1280, Height: 800}
)

// ============================================================================
// Boundary marker system - viewport checking
// ============================================================================
//
// Boundary markers indicate the position of an element relative to the slide
// viewport:
//   - Elements are never blocked from being placed outside the viewport.
//   - The result is available for downstream processing.
//   - Batch boundary detection is supported.

// BoundaryStatus describes the boundary state of an element.
type BoundaryStatus int

const (
	// BoundaryStatusInside means the element is fully within the viewport.
	BoundaryStatusInside BoundaryStatus = iota
	// BoundaryStatusPartial means the element partially overflows the viewport.
	BoundaryStatusPartial
	// BoundaryStatusOutside means the element is fully outside the viewport.
	BoundaryStatusOutside
	// BoundaryStatusOverflowRight means the element overflows the right edge.
	BoundaryStatusOverflowRight
	// BoundaryStatusOverflowLeft means the element overflows the left edge.
	BoundaryStatusOverflowLeft
	// BoundaryStatusOverflowTop means the element overflows the top edge.
	BoundaryStatusOverflowTop
	// BoundaryStatusOverflowBottom means the element overflows the bottom edge.
	BoundaryStatusOverflowBottom
)

// String returns the string representation of the boundary status.
func (bs BoundaryStatus) String() string {
	switch bs {
	case BoundaryStatusInside:
		return "Inside"
	case BoundaryStatusPartial:
		return "Partial"
	case BoundaryStatusOutside:
		return "Outside"
	case BoundaryStatusOverflowRight:
		return "OverflowRight"
	case BoundaryStatusOverflowLeft:
		return "OverflowLeft"
	case BoundaryStatusOverflowTop:
		return "OverflowTop"
	case BoundaryStatusOverflowBottom:
		return "OverflowBottom"
	default:
		return "Unknown"
	}
}

// BoundaryCheckResult holds the result of a boundary check.
type BoundaryCheckResult struct {
	// Status is the boundary state.
	Status BoundaryStatus
	// ElementRect is the element rectangle (x, y, cx, cy in px).
	ElementRect Rect
	// ViewportRect is the viewport rectangle (0, 0, width, height in px).
	ViewportRect Rect
	// OverflowX is the overflow amount along X (positive = right overflow, negative = left overflow).
	OverflowX int
	// OverflowY is the overflow amount along Y (positive = bottom overflow, negative = top overflow).
	OverflowY int
	// IsVisible indicates whether any part of the element is within the viewport.
	IsVisible bool
}

// Rect represents a rectangular region.
type Rect struct {
	X, Y   int // top-left corner (px)
	Cx, Cy int // width and height (px)
}

// SlideViewport represents the slide viewport.
type SlideViewport struct {
	// Width is the viewport width (px).
	Width int
	// Height is the viewport height (px).
	Height int
	// SizeName is the optional standard size name.
	SizeName string
}

// NewSlideViewport creates a slide viewport with the given dimensions.
func NewSlideViewport(width, height int) *SlideViewport {
	return &SlideViewport{
		Width:  width,
		Height: height,
	}
}

// NewSlideViewportFromSize creates a viewport from a SlideSize.
func NewSlideViewportFromSize(size SlideSize) *SlideViewport {
	return &SlideViewport{
		Width:  size.Width,
		Height: size.Height,
	}
}

// Rect returns the viewport as a Rect.
func (vp *SlideViewport) Rect() Rect {
	return Rect{X: 0, Y: 0, Cx: vp.Width, Cy: vp.Height}
}

// CheckBoundary checks whether an element is within the viewport.
// x, y are the top-left coordinates (px); cx, cy are the dimensions (px).
func (vp *SlideViewport) CheckBoundary(x, y, cx, cy int) BoundaryCheckResult {
	elementRect := Rect{X: x, Y: y, Cx: cx, Cy: cy}
	viewportRect := vp.Rect()

	result := BoundaryCheckResult{
		ElementRect:  elementRect,
		ViewportRect: viewportRect,
		IsVisible:    true,
	}

	// Compute element bottom-right corner.
	elementRight := x + cx
	elementBottom := y + cy

	// Compute overflow amounts.
	// Positive = right overflow; negative = left overflow.
	result.OverflowX = elementRight - vp.Width
	if x < 0 {
		result.OverflowX = x - 0 // negative indicates left overflow
	}

	// Positive = bottom overflow; negative = top overflow.
	result.OverflowY = elementBottom - vp.Height
	if y < 0 {
		result.OverflowY = y - 0 // negative indicates top overflow
	}

	// Determine visibility (at least partially within viewport).
	result.IsVisible = !(elementRight <= 0 || x >= vp.Width ||
		elementBottom <= 0 || y >= vp.Height)

	// Determine boundary status.
	if x >= 0 && y >= 0 && elementRight <= vp.Width && elementBottom <= vp.Height {
		// Fully within the viewport.
		result.Status = BoundaryStatusInside
	} else if elementRight <= 0 || x >= vp.Width || elementBottom <= 0 || y >= vp.Height {
		// Fully outside the viewport.
		result.Status = BoundaryStatusOutside
		result.IsVisible = false
	} else {
		// Partially outside; direction-specific status not set here.
		result.Status = BoundaryStatusPartial
	}

	return result
}

// CheckRect checks whether a Rect is within the viewport.
func (vp *SlideViewport) CheckRect(rect Rect) BoundaryCheckResult {
	return vp.CheckBoundary(rect.X, rect.Y, rect.Cx, rect.Cy)
}

// IsInside reports whether an element is fully within the viewport.
func (vp *SlideViewport) IsInside(x, y, cx, cy int) bool {
	return vp.CheckBoundary(x, y, cx, cy).Status == BoundaryStatusInside
}

// IsVisible reports whether any part of an element is within the viewport.
func (vp *SlideViewport) IsVisible(x, y, cx, cy int) bool {
	return vp.CheckBoundary(x, y, cx, cy).IsVisible
}

// Slide is the high-level slide object.
type Slide struct {
	// presentation is the owning Presentation.
	presentation *Presentation

	// part is the underlying slide part.
	part *slide.SlidePart

	// builder is the slide builder.
	builder *SlideBuilder

	// mediaManager is shared with the owning Presentation.
	mediaManager *MediaManager

	// index is the zero-based slide index (position in the deck).
	index int

	// num is the slide's file number (slideN.xml). Stable across reordering, so
	// dependent parts (e.g. notesSlideN.xml) keep a consistent name.
	num int

	// notes is the speaker-notes rich-text frame (D-022), lazily created by
	// SpeakerNotes; nil means the slide has no notes.
	notes *TextFrame

	// shapeIDCounter is a lock-free atomic counter for allocating unique shape IDs.
	shapeIDCounter atomic.Uint32
}

// ============================================================================
// Basic accessors
// ============================================================================

// Index returns the zero-based slide index.
func (s *Slide) Index() int {
	return s.index
}

// Part returns the underlying SlidePart.
func (s *Slide) Part() *slide.SlidePart {
	return s.part
}

// Builder returns the slide builder.
func (s *Slide) Builder() *SlideBuilder {
	return s.builder
}

// PartURI returns the part URI.
func (s *Slide) PartURI() *opc.PackURI {
	return s.part.PartURI()
}

// ============================================================================
// Viewport and boundary checking
// ============================================================================

// Viewport returns the slide viewport.
func (s *Slide) Viewport() *SlideViewport {
	cx, cy := s.SlideSize()
	return NewSlideViewport(cx, cy)
}

// CheckBoundary checks whether an element is within the slide viewport.
// x, y are the top-left coordinates (px); cx, cy are the dimensions (px).
// Returns a BoundaryCheckResult with overflow and visibility information.
func (s *Slide) CheckBoundary(x, y, cx, cy int) BoundaryCheckResult {
	return s.Viewport().CheckBoundary(x, y, cx, cy)
}

// IsInsideBoundary reports whether an element is fully within the slide viewport.
func (s *Slide) IsInsideBoundary(x, y, cx, cy int) bool {
	return s.Viewport().IsInside(x, y, cx, cy)
}

// IsVisible reports whether any part of an element is within the slide viewport.
func (s *Slide) IsVisible(x, y, cx, cy int) bool {
	return s.Viewport().IsVisible(x, y, cx, cy)
}

// ============================================================================
// Component system
// ============================================================================

// AddComponent adds a component to the slide.
// Any value implementing the Component interface (text, picture, chart, etc.)
// is accepted. A SlideContext is created internally and c.Render(ctx) is called.
func (s *Slide) AddComponent(c Component) error {
	ctx := NewSlideContext(s)
	return c.Render(ctx)
}

// AddComponents adds multiple components to the slide in one call.
func (s *Slide) AddComponents(components ...Component) error {
	ctx := NewSlideContext(s)
	return ctx.RenderComponents(components...)
}

// NewContext creates a SlideContext for manual component rendering.
func (s *Slide) NewContext() *SlideContext {
	return NewSlideContext(s)
}

// ============================================================================
// Text methods - default unit: EMU
// ============================================================================

// AddTextBox adds a text box to the slide.
// x, y are the position (EMU); cx, cy are the size (EMU); text is the content.
func (s *Slide) AddTextBox(x, y, cx, cy int, text string) *slide.XSp {
	return s.builder.AddTextBox(
		x, y,
		cx, cy,
		text,
	)
}

// ============================================================================
// Shape methods - default unit: EMU
// ============================================================================

// AddAutoShape adds an auto shape to the slide.
// x, y are the position (EMU); cx, cy are the size (EMU).
// presetID is the preset shape type (e.g. "rectangle", "ellipse", "roundRect").
func (s *Slide) AddAutoShape(x, y, cx, cy int, presetID string) *slide.XSp {
	return s.builder.AddAutoShape(
		x, y,
		cx, cy,
		presetID,
	)
}

// AddRectangle adds a rectangle to the slide.
func (s *Slide) AddRectangle(x, y, cx, cy int) *slide.XSp {
	return s.AddAutoShape(x, y, cx, cy, "rect")
}

// AddEllipse adds an ellipse to the slide.
func (s *Slide) AddEllipse(x, y, cx, cy int) *slide.XSp {
	return s.AddAutoShape(x, y, cx, cy, "ellipse")
}

// AddRoundRect adds a rounded rectangle to the slide.
func (s *Slide) AddRoundRect(x, y, cx, cy int) *slide.XSp {
	return s.AddAutoShape(x, y, cx, cy, "roundRect")
}

// ============================================================================
// Picture methods - default unit: EMU
// ============================================================================

// AddPicture adds a picture to the slide.
// x, y are the position (EMU); cx, cy are the size (EMU).
// imageRId is the relationship ID of the image.
func (s *Slide) AddPicture(x, y, cx, cy int, imageRId string) *slide.XPicture {
	return s.builder.AddPicture(
		x, y,
		cx, cy,
		imageRId,
	)
}

// AddPictureFromBytes adds a picture from raw bytes. The media bytes are
// deduplicated across the deck and written to the package once; the slide
// receives its own image relationship. The extension of fileName selects the
// stored media extension (defaulting to .png).
func (s *Slide) AddPictureFromBytes(x, y, cx, cy int, fileName string, data []byte) (*slide.XPicture, error) {
	ext := strings.ToLower(filepath.Ext(fileName))
	rID := s.addImagePart(data, ext)
	return s.builder.AddPicture(x, y, cx, cy, rID), nil
}

// AddPictureFromFile adds a picture from a file path.
func (s *Slide) AddPictureFromFile(x, y, cx, cy int, path string) (*slide.XPicture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read image file %q: %w", path, err)
	}
	return s.AddPictureFromBytes(x, y, cx, cy, filepath.Base(path), data)
}

// ============================================================================
// Table methods - default unit: EMU
// ============================================================================

// AddTable adds a table to the slide.
// x, y are the position (EMU); cx, cy are the size (EMU).
// rows and cols specify the table dimensions.
func (s *Slide) AddTable(x, y, cx, cy, rows, cols int) *slide.XGraphicFrame {
	return s.builder.AddTable(
		x, y,
		cx, cy,
		rows, cols,
	)
}

// SetTableCellText sets the text content of a table cell.
func (s *Slide) SetTableCellText(gf *slide.XGraphicFrame, row, col int, text string) {
	s.builder.SetTableCellText(gf, row, col, text)
}

// ============================================================================
// Relationship management
// ============================================================================

// AddImageRel adds an image relationship to the slide.
func (s *Slide) AddImageRel(targetURI string) string {
	return s.builder.AddImage(targetURI)
}

// AddMediaRel adds a media relationship to the slide.
func (s *Slide) AddMediaRel(targetURI string) string {
	return s.builder.AddMedia(targetURI)
}

// AddChartRel adds a chart relationship to the slide.
func (s *Slide) AddChartRel(targetURI string) string {
	return s.builder.AddChart(targetURI)
}

// HasImage reports whether a relationship for the given target URI already exists.
func (s *Slide) HasImage(targetURI string) bool {
	return s.builder.HasImage(targetURI)
}

// GetImageRId returns the rId for a target URI, adding it if absent.
func (s *Slide) GetImageRId(targetURI string) string {
	return s.builder.GetImageRId(targetURI)
}

// ============================================================================
// Slide size - default unit: EMU
// ============================================================================

// SlideSize returns the slide dimensions in pixels.
func (s *Slide) SlideSize() (cx, cy int) {
	emuCX, emuCY := s.presentation.SlideSize()
	return EMUToPx(emuCX), EMUToPx(emuCY)
}

// SlideSizeEMU returns the slide dimensions in EMU (advanced usage).
func (s *Slide) SlideSizeEMU() (cx, cy int) {
	return s.presentation.SlideSize()
}

// ============================================================================
// Unit conversion - delegates to the utils package
// ============================================================================

// PxToEMU converts pixels to EMU (at 96 DPI).
//
// Deprecated: shape methods now take EMU directly; use pptx.Px (which returns a
// typed EMU) to convert pixel coordinates.
func PxToEMU(px int) int {
	return int(utils.PixelsToEMU(float64(px)))
}

// EMUToPx converts EMU to pixels (at 96 DPI).
func EMUToPx(emu int) int {
	return int(utils.EMUToPixels(int64(emu)))
}

// ============================================================================
// Layout management
// ============================================================================

// SetLayout sets the slide layout by name
// (e.g. "blank", "title", "titleAndContent").
// Returns true if the layout was found and applied.
func (s *Slide) SetLayout(layoutName string) bool {
	if s.presentation.masterCache == nil {
		return false
	}

	layoutData, ok := s.presentation.masterCache.GetLayoutByName(layoutName)
	if !ok {
		return false
	}

	// Set the layout relationship ID.
	s.part.SetLayoutRId(layoutData.ID())
	return true
}

// Layout returns the name of the current layout.
func (s *Slide) Layout() string {
	layoutRId := s.part.LayoutRId()
	if layoutRId == "" {
		return ""
	}

	// Look up the layout name from the master cache.
	if s.presentation.masterCache != nil {
		if layout, ok := s.presentation.masterCache.GetLayout(layoutRId); ok {
			return layout.Name()
		}
	}

	return layoutRId
}
