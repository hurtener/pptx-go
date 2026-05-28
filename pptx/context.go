// Package pptx provides a high-level API for authoring PPTX files.
package pptx

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/hurtener/pptx-go/internal/ooxml/chart"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// SlideContext — component rendering context
// ============================================================================
//
// SlideContext is the "delegate" passed to a component, giving it limited but
// privileged access to slide-level resources.
//
// Core responsibilities:
// 1. Allocate collision-free shape IDs.
// 2. Manage media resources (images, audio, video).
// 3. Manage chart XML and return the corresponding relationship ID.
// 4. Mount the XML structures produced by the component onto the slide.
//
// Usage example:
//
//	func (c *MyComponent) Render(ctx *SlideContext) error {
//		// 1. allocate a shape ID
//		id := ctx.NextShapeID()
//
//		// 2. add an image and obtain its relationship ID
//		rId, err := ctx.AddMedia(imageBytes, "image.png")
//		if err != nil {
//			return err
//		}
//
//		// 3. build the shape XML
//		sp := &slide.XSp{...}
//
//		// 4. mount it onto the slide
//		ctx.AppendShape(sp)
//		return nil
//	}
//
// ============================================================================

// SlideContext carries the resources and capabilities a component needs to
// render itself onto a slide.
type SlideContext struct {
	// the slide this context belongs to
	slide *Slide

	// shape IDs allocated via this context
	allocatedIDs map[uint32]bool

	// relationship IDs allocated via this context
	allocatedRIDs map[string]bool

	// concurrency guard
	mu sync.RWMutex
}

// NewSlideContext creates a SlideContext for the given slide.
func NewSlideContext(s *Slide) *SlideContext {
	return &SlideContext{
		slide:         s,
		allocatedIDs:  make(map[uint32]bool),
		allocatedRIDs: make(map[string]bool),
	}
}

// ============================================================================
// Shape ID management
// ============================================================================

// NextShapeID allocates and returns the next collision-free shape ID.
// Safe for concurrent use; uses an atomic increment.
func (ctx *SlideContext) NextShapeID() uint32 {
	// Add(1) atomically increments the counter and returns the new value.
	return ctx.slide.shapeIDCounter.Add(1)
}

// CurrentShapeID returns the most recently allocated shape ID.
func (ctx *SlideContext) CurrentShapeID() uint32 {
	return ctx.slide.shapeIDCounter.Load()
}

// AllocateShapeIDBatch allocates a batch of shape IDs at once.
func (ctx *SlideContext) AllocateShapeIDBatch(count int) []uint32 {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ids := ctx.slide.part.AllocateShapeIDBatch(count)
	for _, id := range ids {
		ctx.allocatedIDs[id] = true
	}
	return ids
}

// IsShapeIDAllocated reports whether the given shape ID has been allocated
// through this context.
func (ctx *SlideContext) IsShapeIDAllocated(id uint32) bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.allocatedIDs[id]
}

// ============================================================================
// Media resource management
// ============================================================================

// AddMedia adds a media resource (image, audio, or video) to the slide.
// data is the raw bytes and fileName is used to infer the MIME type.
// Returns the relationship ID and any error.
func (ctx *SlideContext) AddMedia(data []byte, fileName string) (string, error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// add the media via MediaManager
	_, resource := ctx.slide.mediaManager.AddMediaAuto(fileName, data)
	if resource == nil {
		return "", fmt.Errorf("failed to add media resource")
	}

	// obtain the target URI
	targetURI := resource.Target()

	// register the relationship on the slide
	slideRID := ctx.slide.part.AddImageRel(targetURI)
	ctx.allocatedRIDs[slideRID] = true

	return slideRID, nil
}

// AddMediaWithMIME adds a media resource with an explicit MIME type.
func (ctx *SlideContext) AddMediaWithMIME(data []byte, fileName, mimeType string) (string, error) {
	// currently delegates to AddMedia; can be extended to honor mimeType
	return ctx.AddMedia(data, fileName)
}

// AddImage adds an image resource. It is an alias for AddMedia with clearer
// semantics.
func (ctx *SlideContext) AddImage(data []byte, fileName string) (string, error) {
	return ctx.AddMedia(data, fileName)
}

// AddVideo adds a video resource.
func (ctx *SlideContext) AddVideo(data []byte, fileName string) (string, error) {
	return ctx.AddMedia(data, fileName)
}

// AddAudio adds an audio resource.
func (ctx *SlideContext) AddAudio(data []byte, fileName string) (string, error) {
	return ctx.AddMedia(data, fileName)
}

// ============================================================================
// Chart management
// ============================================================================

// AddChartXML adds a chart part from raw XML bytes.
// Returns the relationship ID and any error.
//
// This implements the "route C" pattern: the component supplies the chart XML;
// SlideContext writes a ChartPart into the OPC package and returns the rId.
func (ctx *SlideContext) AddChartXML(chartXML []byte) (string, error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// obtain a reference to the Presentation
	pres := ctx.slide.presentation

	// allocate a chart number
	chartNum := int(atomic.AddInt32(&pres.chartCounter, 1))

	// create the chart part
	chartPart := chart.NewChartPart(chartNum)
	chartPart.SetRawXML(chartXML)

	// add the part to the OPC package
	chartURI := chartPart.PartURI()
	chartBlob := []byte(chartPart.Template()) // get the XML content
	part := opc.NewPart(chartURI, opc.ContentTypeChart, chartBlob)
	if err := pres.pkg.AddPart(part); err != nil {
		return "", fmt.Errorf("adding chart part: %w", err)
	}

	// register the relationship on the slide
	chartRelID := ctx.slide.part.AddChartRel(chartURI.RelPathFrom(ctx.slide.part.PartURI()))
	ctx.allocatedRIDs[chartRelID] = true

	return chartRelID, nil
}

// AddChart adds a chart part using a template and the provided data map.
// Returns the relationship ID and any error.
func (ctx *SlideContext) AddChart(chartType chart.ChartType, data map[string]interface{}) (string, error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	pres := ctx.slide.presentation
	chartNum := int(atomic.AddInt32(&pres.chartCounter, 1))

	chartPart := chart.NewChartPartWithType(chartNum, chartType)

	// substitute placeholders
	for key, value := range data {
		chartPart.ReplacePlaceholder(key, fmt.Sprint(value))
	}

	// add the part to the OPC package
	chartURI := chartPart.PartURI()
	chartBlob := []byte(chartPart.Template())
	part := opc.NewPart(chartURI, opc.ContentTypeChart, chartBlob)
	if err := pres.pkg.AddPart(part); err != nil {
		return "", fmt.Errorf("adding chart part: %w", err)
	}

	// register the relationship
	chartRelID := ctx.slide.part.AddChartRel(chartURI.RelPathFrom(ctx.slide.part.PartURI()))
	ctx.allocatedRIDs[chartRelID] = true

	return chartRelID, nil
}

// ============================================================================
// Shape mounting
// ============================================================================

// AppendShape appends a shape to the slide.
// shape may be *slide.XSp, *slide.XPicture, *slide.XGraphicFrame, etc.
func (ctx *SlideContext) AppendShape(shape interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.slide.part.AppendShapeChild(shape)
}

// AppendShapes appends multiple shapes to the slide in one call.
func (ctx *SlideContext) AppendShapes(shapes ...interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	for _, shape := range shapes {
		ctx.slide.part.AppendShapeChild(shape)
	}
}

// ============================================================================
// Relationship management
// ============================================================================

// AddImageRel adds an image relationship and returns its relationship ID.
func (ctx *SlideContext) AddImageRel(targetURI string) string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	rID := ctx.slide.part.AddImageRel(targetURI)
	ctx.allocatedRIDs[rID] = true
	return rID
}

// AddMediaRel adds a media relationship and returns its relationship ID.
func (ctx *SlideContext) AddMediaRel(targetURI string) string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	rID := ctx.slide.part.AddMediaRel(targetURI)
	ctx.allocatedRIDs[rID] = true
	return rID
}

// AddChartRel adds a chart relationship and returns its relationship ID.
func (ctx *SlideContext) AddChartRel(targetURI string) string {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	rID := ctx.slide.part.AddChartRel(targetURI)
	ctx.allocatedRIDs[rID] = true
	return rID
}

// HasRelationship reports whether the given relationship ID was allocated
// through this context.
func (ctx *SlideContext) HasRelationship(rID string) bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.allocatedRIDs[rID]
}

// ============================================================================
// Slide information accessors
// ============================================================================

// SlideIndex returns the zero-based index of the slide.
func (ctx *SlideContext) SlideIndex() int {
	return ctx.slide.Index()
}

// SlideSize returns the slide dimensions in EMU (cx, cy).
func (ctx *SlideContext) SlideSize() (cx, cy int) {
	return ctx.slide.presentation.SlideSize()
}

// SlidePart returns the underlying SlidePart (advanced use).
func (ctx *SlideContext) SlidePart() *slide.SlidePart {
	return ctx.slide.part
}

// Presentation returns the owning Presentation (advanced use).
func (ctx *SlideContext) Presentation() *Presentation {
	return ctx.slide.presentation
}

// ============================================================================
// Unit conversion helpers — default unit: px
// ============================================================================

// PxToEMU converts pixels to EMU at 96 DPI.
func (ctx *SlideContext) PxToEMU(px int) int {
	return PxToEMU(px)
}

// EMUToPx converts EMU to pixels at 96 DPI.
func (ctx *SlideContext) EMUToPx(emu int) int {
	return EMUToPx(emu)
}

// ============================================================================
// Batch operations
// ============================================================================

// RenderComponents renders each component in order, stopping on the first error.
func (ctx *SlideContext) RenderComponents(components ...Component) error {
	for i, c := range components {
		if err := c.Render(ctx); err != nil {
			return &ComponentRenderError{
				Index:      i,
				Component:  c,
				Underlying: err,
			}
		}
	}
	return nil
}
