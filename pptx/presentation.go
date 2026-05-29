// Package pptx provides a high-level interface for working with PPTX files.
// It is the primary entry point for both human developers and AI callers.
package pptx

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/hurtener/pptx-go/internal/ooxml/presentation"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// Presentation - top-level facade
// ============================================================================
//
// Presentation is the primary entry point for both human developers and AI
// callers. It wraps the underlying opc.Package and provides high-level
// business methods.
//
// Core responsibilities:
//  1. Lazily load and clone a thread-safe copy from a template.
//  2. Create a dedicated MediaManager per instance (prevents concurrent image
//     cross-contamination).
//  3. Automatically manage slide ID registration and .rels routing.
//  4. Provide streaming output (supports HTTP response writers).
//
// Example:
//
//	pres := pptx.New()
//	slide1 := pres.AddSlide()
//	slide1.AddTextBox(100, 100, 500, 50, "Hello World")
//	pres.Save("output.pptx")
//
// ============================================================================

// Presentation is the top-level PPTX facade.
type Presentation struct {
	// pkg is the underlying OPC package.
	pkg *opc.Package

	// presentationPart wraps presentation.xml.
	presentationPart *presentation.PresentationPart

	// slides is the ordered list of slides.
	slides []*Slide

	// mediaManager is instance-specific to prevent concurrent cross-contamination.
	mediaManager *MediaManager

	// masterManager manages slide masters.
	masterManager *MasterManager

	// masterCache holds master/layout information loaded from a template.
	masterCache *MasterCache

	// slideCounter generates slide file names (slide1.xml, slide2.xml, …).
	slideCounter int32

	// chartCounter generates chart file names (chart1.xml, chart2.xml, …).
	chartCounter int32

	// relCounter generates relationship IDs.
	relCounter int32

	// fontSource resolves font bytes for EmbedFont (nil = no source). D-019.
	fontSource FontSource

	// theme is the active theme driving token resolution (default
	// DefaultTheme). Set via WithTheme or SetTheme.
	theme *Theme

	// fontCounter generates embedded font part names (font1.fntdata, …).
	fontCounter int32

	// mu guards concurrent access.
	mu sync.RWMutex
}

// ============================================================================
// Constructors
// ============================================================================

// New creates a blank presentation. With no options it is a 16:9 widescreen
// deck themed with DefaultTheme; pass options (WithFormat, WithFontSource,
// WithTheme) to configure it (RFC §8.1).
func New(opts ...Option) *Presentation {
	pres := &Presentation{
		pkg:              opc.NewPackage(),
		presentationPart: presentation.NewPresentationPart(),
		slides:           make([]*Slide, 0),
		mediaManager:     NewMediaManager(),
		masterManager:    NewMasterManager(),
		theme:            DefaultTheme(),
		slideCounter:     0,
		relCounter:       0,
	}

	// Apply caller options (format, font source, theme) before any emission.
	for _, opt := range opts {
		if opt != nil {
			opt(pres)
		}
	}

	// Initialize the package structure.
	pres.initPackage()

	// Seed a complete scaffold (master + blank layout + theme) so the deck is
	// valid the moment it is created (Phase 03 A2; RFC §8.7).
	pres.seedScaffold()

	return pres
}

// NewWithTemplate creates a presentation from the named template
// (e.g. TemplateDefault, TemplateBlank).
func NewWithTemplate(name TemplateType) (*Presentation, error) {
	// Load the template from the template manager.
	pkg, err := globalTemplateManager.LoadTemplate(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	pres := &Presentation{
		pkg:           pkg,
		slides:        make([]*Slide, 0),
		mediaManager:  NewMediaManager(),
		masterManager: NewMasterManager(),
		slideCounter:  0,
		relCounter:    0,
	}

	// Parse presentation.xml from the package.
	if err := pres.loadPresentationPart(); err != nil {
		return nil, fmt.Errorf("failed to parse presentation part: %w", err)
	}

	// Retrieve the master cache.
	pres.masterCache = globalTemplateManager.GetMasterCache()

	return pres, nil
}

// NewFromBytes creates a presentation from raw PPTX bytes.
func NewFromBytes(data []byte) (*Presentation, error) {
	reader := bytes.NewReader(data)
	pkg, err := opc.Open(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PPTX data: %w", err)
	}

	pres := &Presentation{
		pkg:           pkg,
		slides:        make([]*Slide, 0),
		mediaManager:  NewMediaManager(),
		masterManager: NewMasterManager(),
		slideCounter:  0,
		relCounter:    0,
	}

	if err := pres.loadPresentationPart(); err != nil {
		return nil, fmt.Errorf("failed to parse presentation part: %w", err)
	}

	return pres, nil
}

// NewFromFile creates a presentation from a PPTX file path.
func NewFromFile(path string) (*Presentation, error) {
	pkg, err := opc.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open PPTX file: %w", err)
	}

	pres := &Presentation{
		pkg:           pkg,
		slides:        make([]*Slide, 0),
		mediaManager:  NewMediaManager(),
		masterManager: NewMasterManager(),
		slideCounter:  0,
		relCounter:    0,
	}

	if err := pres.loadPresentationPart(); err != nil {
		return nil, fmt.Errorf("failed to parse presentation part: %w", err)
	}

	return pres, nil
}

// ============================================================================
// Initialization
// ============================================================================

// initPackage sets up the OPC package structure.
func (p *Presentation) initPackage() {
	// Add presentation.xml.
	uri := opc.NewPackURI("/ppt/presentation.xml")
	blob, _ := p.presentationPart.ToXML()
	part := opc.NewPart(uri, opc.ContentTypePresentation, blob)
	_ = p.pkg.AddPart(part)

	// Add the package-level relationship pointing to presentation.xml.
	_, _ = p.pkg.AddRelationship(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)
}

// loadPresentationPart parses presentation.xml from the package.
func (p *Presentation) loadPresentationPart() error {
	// Locate the presentation part via its relationship type.
	part := p.pkg.GetPartByRelType(opc.RelTypeOfficeDocument)
	if part == nil {
		// Not found; create a new empty part.
		p.presentationPart = presentation.NewPresentationPart()
		return nil
	}

	// Parse the XML.
	p.presentationPart = presentation.NewPresentationPart()
	if err := p.presentationPart.FromXML(part.Blob()); err != nil {
		return err
	}

	// Synchronize the slide counter.
	p.slideCounter = int32(p.presentationPart.SlideCount())

	return nil
}

// ============================================================================
// Slide management - core methods
// ============================================================================

// AddSlide appends a new slide to the presentation.
// layout is an optional layout name (e.g. "title", "blank", "titleAndContent").
// If no layout is specified, a blank layout is used.
//
// Internally this method:
//   - Registers a new ID in presentation.xml's <p:sldIdLst>.
//   - Allocates a .rels route.
//   - Returns a high-level *Slide.
func (p *Presentation) AddSlide(layout ...string) *Slide {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Allocate a slide number.
	slideNum := int(atomic.AddInt32(&p.slideCounter, 1))

	// Create the slide part.
	slidePart := slide.NewSlidePart(slideNum)

	// Resolve layout.
	layoutRId := ""
	if len(layout) > 0 && layout[0] != "" {
		if p.masterCache != nil {
			if layoutData, ok := p.masterCache.GetLayoutByName(layout[0]); ok {
				// Create the layout relationship.
				layoutRId = p.allocateRelID()
				// TODO: add layout relationship to the slide
				_ = layoutData
			}
		}
	}
	slidePart.SetLayoutRId(layoutRId)

	// Set the slide URI.
	slideURI := opc.NewPackURI(fmt.Sprintf("/ppt/slides/slide%d.xml", slideNum))
	slidePart.SetURI(slideURI)

	// Add the part to the package.
	slideBlob, _ := slidePart.ToXML()
	slidePartOPC := opc.NewPart(slideURI, opc.ContentTypeSlide, slideBlob)
	_ = p.pkg.AddPart(slidePartOPC)

	// Wire presentation→slide and slide→layout relationships; the returned
	// relationship id is what <p:sldId r:id="…"> must carry (Phase 03 A2).
	slideRId := p.relateSlide(slidePartOPC)

	// Register with PresentationPart (auto-assigns a slide ID; slideRId is the
	// presentation→slide relationship that <p:sldId> references).
	_ = p.presentationPart.AddSlide(slideRId, slidePart)

	// Build the high-level Slide object.
	s := &Slide{
		presentation: p,
		part:         slidePart,
		builder:      NewSlideBuilder(slidePart),
		mediaManager: p.mediaManager,
		index:        len(p.slides),
	}
	// Initialize the atomic counter (OOXML spec: shapeId starts at 2; 1 is
	// reserved for the root node).
	s.shapeIDCounter.Store(1)

	p.slides = append(p.slides, s)

	return s
}

// AddSlideAt inserts a new slide at the specified index.
func (p *Presentation) AddSlideAt(index int, layout ...string) (*Slide, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index > len(p.slides) {
		return nil, fmt.Errorf("index %d out of range [0, %d]", index, len(p.slides))
	}

	// Create the slide.
	slideNum := int(atomic.AddInt32(&p.slideCounter, 1))
	slidePart := slide.NewSlidePart(slideNum)

	// Resolve layout.
	layoutRId := ""
	if len(layout) > 0 && layout[0] != "" && p.masterCache != nil {
		if layoutData, ok := p.masterCache.GetLayoutByName(layout[0]); ok {
			_ = layoutData // TODO: set layout relationship
		}
	}
	slidePart.SetLayoutRId(layoutRId)

	// Set URI.
	slideURI := opc.NewPackURI(fmt.Sprintf("/ppt/slides/slide%d.xml", slideNum))
	slidePart.SetURI(slideURI)

	// Add to package.
	slideBlob, _ := slidePart.ToXML()
	slidePartOPC := opc.NewPart(slideURI, opc.ContentTypeSlide, slideBlob)
	_ = p.pkg.AddPart(slidePartOPC)

	// Wire presentation→slide and slide→layout relationships.
	slideRId := p.relateSlide(slidePartOPC)

	// Register with PresentationPart.
	_ = p.presentationPart.AddSlide(slideRId, slidePart)

	// Build the high-level object.
	s := &Slide{
		presentation: p,
		part:         slidePart,
		builder:      NewSlideBuilder(slidePart),
		mediaManager: p.mediaManager,
		index:        index,
	}
	// Initialize the atomic counter (OOXML spec: shapeId starts at 2; 1 is
	// reserved for the root node).
	s.shapeIDCounter.Store(1)

	// Insert at the specified position.
	p.slides = append(p.slides[:index], append([]*Slide{s}, p.slides[index:]...)...)

	// Update indexes for subsequent slides.
	for i := index + 1; i < len(p.slides); i++ {
		p.slides[i].index = i
	}

	return s, nil
}

// GetSlide returns the slide at the given index.
func (p *Presentation) GetSlide(index int) (*Slide, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if index < 0 || index >= len(p.slides) {
		return nil, fmt.Errorf("index %d out of range [0, %d)", index, len(p.slides))
	}

	return p.slides[index], nil
}

// RemoveSlide removes the slide at the given index.
func (p *Presentation) RemoveSlide(index int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index >= len(p.slides) {
		return fmt.Errorf("index %d out of range [0, %d)", index, len(p.slides))
	}

	// Get the slide to remove.
	s := p.slides[index]

	// Remove the part from the package.
	_ = p.pkg.RemovePart(s.part.PartURI())

	// Remove from presentation.xml.
	_ = p.presentationPart.RemoveSlide(index)

	// Remove from the slice.
	p.slides = append(p.slides[:index], p.slides[index+1:]...)

	// Update indexes for subsequent slides.
	for i := index; i < len(p.slides); i++ {
		p.slides[i].index = i
	}

	return nil
}

// SlideCount returns the number of slides.
func (p *Presentation) SlideCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.slides)
}

// Slides returns a copy of the slide list.
func (p *Presentation) Slides() []*Slide {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*Slide, len(p.slides))
	copy(result, p.slides)
	return result
}

// ============================================================================
// Save methods
// ============================================================================

// Save serializes the presentation and writes it to a file.
func (p *Presentation) Save(path string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Sync all slides to the package.
	if err := p.syncSlides(); err != nil {
		return err
	}

	// Sync presentation.xml.
	if err := p.syncPresentationPart(); err != nil {
		return err
	}

	// Write to file.
	return p.pkg.SaveFile(path)
}

// Write serializes the presentation and writes it to an io.Writer.
// This is suitable for high-concurrency streaming output such as HTTP responses.
func (p *Presentation) Write(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Sync all slides to the package.
	if err := p.syncSlides(); err != nil {
		return err
	}

	// Sync presentation.xml.
	if err := p.syncPresentationPart(); err != nil {
		return err
	}

	// Write to the writer.
	return p.pkg.Save(w)
}

// WriteToBytes serializes the presentation and returns it as a byte slice.
func (p *Presentation) WriteToBytes() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Sync all slides to the package.
	if err := p.syncSlides(); err != nil {
		return nil, err
	}

	// Sync presentation.xml.
	if err := p.syncPresentationPart(); err != nil {
		return nil, err
	}

	// Write to a byte slice.
	return p.pkg.SaveToBytes()
}

// ============================================================================
// Sync helpers
// ============================================================================

// syncSlides serializes all slides into the OPC package.
func (p *Presentation) syncSlides() error {
	for _, s := range p.slides {
		blob, err := s.part.ToXML()
		if err != nil {
			return fmt.Errorf("failed to serialize slide %d: %w", s.index+1, err)
		}

		// Update or create the part.
		uri := s.part.PartURI()
		existingPart := p.pkg.GetPart(uri)
		if existingPart != nil {
			existingPart.SetBlob(blob)
		} else {
			part := opc.NewPart(uri, opc.ContentTypeSlide, blob)
			_ = p.pkg.AddPart(part)
		}
	}

	return nil
}

// syncPresentationPart serializes presentation.xml into the OPC package.
func (p *Presentation) syncPresentationPart() error {
	blob, err := p.presentationPart.ToXML()
	if err != nil {
		return fmt.Errorf("failed to serialize presentation.xml: %w", err)
	}

	uri := opc.NewPackURI("/ppt/presentation.xml")
	existingPart := p.pkg.GetPart(uri)
	if existingPart != nil {
		existingPart.SetBlob(blob)
	} else {
		part := opc.NewPart(uri, opc.ContentTypePresentation, blob)
		_ = p.pkg.AddPart(part)
	}

	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// allocateRelID allocates a new relationship ID.
func (p *Presentation) allocateRelID() string {
	id := atomic.AddInt32(&p.relCounter, 1)
	return fmt.Sprintf("rId%d", id)
}

// Package returns the underlying OPC package (advanced usage).
func (p *Presentation) Package() *opc.Package {
	return p.pkg
}

// PresentationPart returns the presentation part.
func (p *Presentation) PresentationPart() *presentation.PresentationPart {
	return p.presentationPart
}

// MediaManager returns the media manager.
func (p *Presentation) MediaManager() *MediaManager {
	return p.mediaManager
}

// MasterCache returns the master cache.
func (p *Presentation) MasterCache() *MasterCache {
	return p.masterCache
}

// Theme returns the active theme (never nil; DefaultTheme by default).
func (p *Presentation) Theme() *Theme {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.theme == nil {
		return DefaultTheme()
	}
	return p.theme
}

// SetTheme sets the active theme used for token resolution. A nil theme is
// ignored.
func (p *Presentation) SetTheme(t *Theme) {
	if t == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.theme = t
}

// SetSlideSize sets the slide dimensions (in EMU).
func (p *Presentation) SetSlideSize(cx, cy int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.presentationPart.SetSlideSize(presentation.SlideSize{Cx: cx, Cy: cy})
}

// SetSlideSizeStandard sets the slide size to a named standard (e.g. "widescreen").
func (p *Presentation) SetSlideSizeStandard(name string) {
	size := presentation.NewSlideSizeFromStandard(name)
	p.SetSlideSize(size.Cx, size.Cy)
}

// SlideSize returns the current slide dimensions (in EMU).
func (p *Presentation) SlideSize() (int, int) {
	size := p.presentationPart.SlideSize()
	return size.Cx, size.Cy
}

// Close releases all resources held by the presentation.
func (p *Presentation) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.slides = nil
	p.mediaManager = nil
	p.masterCache = nil

	return p.pkg.Close()
}

// ============================================================================
// Clone
// ============================================================================

// Clone returns a fully independent deep copy of the presentation.
func (p *Presentation) Clone() (*Presentation, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Clone the underlying package.
	newPkg := p.pkg.Clone()

	// Create the new presentation.
	newPres := &Presentation{
		pkg:           newPkg,
		slides:        make([]*Slide, len(p.slides)),
		mediaManager:  NewMediaManager(),
		masterManager: p.masterManager,
		masterCache:   p.masterCache,
		theme:         p.theme,
		slideCounter:  p.slideCounter,
		relCounter:    p.relCounter,
	}

	// Clone the presentation part.
	newPres.presentationPart = presentation.NewPresentationPart()
	presPartData, err := p.presentationPart.ToXML()
	if err != nil {
		return nil, err
	}
	if err := newPres.presentationPart.FromXML(presPartData); err != nil {
		return nil, err
	}

	// Clone each slide.
	for i, s := range p.slides {
		newSlidePart := slide.NewSlidePartWithURI(s.part.PartURI())
		slideData, err := s.part.ToXML()
		if err != nil {
			return nil, err
		}
		if err := newSlidePart.FromXML(slideData); err != nil {
			return nil, err
		}

		newPres.slides[i] = &Slide{
			presentation: newPres,
			part:         newSlidePart,
			builder:      NewSlideBuilder(newSlidePart),
			mediaManager: newPres.mediaManager,
			index:        i,
		}
	}

	return newPres, nil
}
