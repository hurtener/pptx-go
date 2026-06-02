// Package pptx provides a high-level interface for working with PPTX files.
// It is the primary entry point for both human developers and AI callers.
package pptx

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hurtener/pptx-go/internal/ooxml/presentation"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
	"github.com/hurtener/pptx-go/internal/render"
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

	// masters is the read-only master/layout registry built when a deck is
	// opened or seeded from a template (Phase 09); nil for a blank New() deck.
	masters []*Master

	// template, when set by FromTemplate, is the brand kit whose theme + masters
	// + layouts seed this presentation (adopted in New before any slide is added).
	template *Presentation

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

	// logger is an optional structured logger (nil = no logs). Set via
	// WithLogger; the builder emits a write-boundary event (RFC §18).
	logger *slog.Logger

	// sections are the named slide groupings (D-021), emitted into
	// presentation.xml's extLst at write time.
	sections []*Section

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

	// Seed from a brand-kit template when one was supplied (RFC §13.1); else
	// initialize a fresh package + the default scaffold. Template adoption falls
	// back to the scaffold on any failure so New never yields a broken deck.
	if pres.template != nil {
		if err := pres.adoptTemplate(); err == nil {
			pres.template = nil // don't retain the source after adoption
			return pres
		}
		pres.template = nil
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

	// Rebuild the high-level slide and section models from the package so an
	// opened deck can be inspected, edited, and re-saved losslessly (G6).
	if err := p.repopulateSlides(); err != nil {
		return err
	}
	p.repopulateSections()
	p.seedMediaCounter()

	// Extract the deck's theme and master/layout registry so an opened deck can
	// serve as a brand kit (RFC §13.1): brand.Theme() returns its theme and
	// brand.Masters() lists its layouts. Both are best-effort — a deck without a
	// theme part keeps DefaultTheme; an unreadable master contributes nothing.
	if t, err := loadThemeFromPackage(p.pkg); err == nil && t != nil {
		p.theme = t
	}
	p.masters = buildMasterRegistry(p.pkg)

	return nil
}

// seedMediaCounter advances the media manager's file counter past any imageN
// parts already in the package, so images added to an opened deck get fresh
// names instead of colliding with existing media.
func (p *Presentation) seedMediaCounter() {
	var max int64
	for _, part := range p.pkg.AllParts() {
		uri := part.PartURI().URI()
		if !strings.HasPrefix(uri, "/ppt/media/image") {
			continue
		}
		base := path.Base(uri)
		name := strings.TrimSuffix(base, path.Ext(base)) // imageN
		var n int64
		if _, err := fmt.Sscanf(name, "image%d", &n); err == nil && n > max {
			max = n
		}
	}
	p.mediaManager.SeedFileCounter(max)
}

// repopulateSlides rebuilds p.slides from the package after an Open, in
// sldIdLst order. Each slide part is parsed back into a slide.SlidePart, its
// relationships are loaded so its rId counter is initialized (future image/
// notes adds won't collide), its layout id is recovered, and its shape-id
// allocator is advanced past the existing shapes. Caller holds p.mu (or is a
// constructor before the value is shared).
func (p *Presentation) repopulateSlides() error {
	presPart := p.pkg.GetPartByRelType(opc.RelTypeOfficeDocument)
	if presPart == nil {
		return nil
	}

	n := int(p.presentationPart.SlideCount())
	p.slides = make([]*Slide, 0, n)
	maxNum := 0
	for i := 0; i < n; i++ {
		relID, err := p.presentationPart.SlideRelID(i)
		if err != nil || relID == "" {
			continue
		}
		rel := presPart.Relationships().Get(relID)
		if rel == nil {
			continue // dangling reference — skip rather than fail
		}
		slideURI := rel.TargetURI()
		opcPart := p.pkg.GetPart(slideURI)
		if opcPart == nil {
			continue
		}

		sp := slide.NewSlidePartWithURI(slideURI)
		if err := sp.FromXML(opcPart.Blob()); err != nil {
			return fmt.Errorf("parse slide %s: %w", slideURI.URI(), err)
		}
		// Load the slide's relationships so the rId counter resumes past the
		// existing rIds, and recover the layout id.
		if rb, relErr := opcPart.RelationshipsBlob(); relErr == nil && rb != nil {
			_ = sp.Relationships().FromXML(rb)
			if lr := sp.Relationships().GetByType(opc.RelTypeSlideLayout); len(lr) > 0 {
				sp.SetLayoutRId(lr[0].RID())
			}
		}
		sp.SetShapeIDStart(maxShapeID(sp) + 1)

		num := slideNumFromURI(slideURI.URI())
		if num > maxNum {
			maxNum = num
		}
		s := &Slide{
			presentation: p,
			part:         sp,
			builder:      NewSlideBuilder(sp),
			mediaManager: p.mediaManager,
			index:        len(p.slides),
			num:          num,
		}
		s.shapeIDCounter.Store(1)
		p.slides = append(p.slides, s)
	}

	// Ensure new slides get a unique file number even after removals.
	if int(p.slideCounter) < maxNum {
		p.slideCounter = int32(maxNum)
	}
	return nil
}

// repopulateSections rebuilds p.sections from the parsed presentation part,
// mapping each section's slide IDs back to slide indices.
func (p *Presentation) repopulateSections() {
	parsed := p.presentationPart.Sections()
	if len(parsed) == 0 {
		return
	}
	ids := p.presentationPart.SlideIDs()
	idToIndex := make(map[uint32]int, len(ids))
	for i, id := range ids {
		idToIndex[id] = i
	}
	p.sections = nil
	for _, se := range parsed {
		sec := &Section{name: se.Name}
		for _, sid := range se.SlideIDs {
			if idx, ok := idToIndex[sid]; ok {
				sec.slideIndexes = append(sec.slideIndexes, idx)
			}
		}
		p.sections = append(p.sections, sec)
	}
}

// maxShapeID returns the largest shape (cNvPr) id used in the slide's shape
// tree, or 1 (the reserved root id) when there are no shapes.
func maxShapeID(sp *slide.SlidePart) uint32 {
	var max uint32 = 1
	for _, child := range sp.SpTree().Children {
		id := 0
		switch c := child.(type) {
		case *slide.XSp:
			if c.NonVisual.CNvPr != nil {
				id = c.NonVisual.CNvPr.ID
			}
		case *slide.XPicture:
			if c.NonVisual.CNvPr != nil {
				id = c.NonVisual.CNvPr.ID
			}
		case *slide.XGraphicFrame:
			if c.NonVisual.CNvPr != nil {
				id = c.NonVisual.CNvPr.ID
			}
		}
		if id > 0 && uint32(id) > max {
			max = uint32(id)
		}
	}
	return max
}

// slideNumFromURI extracts N from a /ppt/slides/slideN.xml pack URI (0 if it
// doesn't match the convention).
func slideNumFromURI(uri string) int {
	base := path.Base(uri)
	var n int
	if _, err := fmt.Sscanf(base, "slide%d.xml", &n); err == nil {
		return n
	}
	return 0
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

	// Resolve the slide's layout: a named layout from the template registry
	// (Phase 09), else the default blank layout.
	layoutURI := p.resolveLayoutURI(layout...)

	// Set the slide URI.
	slideURI := opc.NewPackURI(fmt.Sprintf("/ppt/slides/slide%d.xml", slideNum))
	slidePart.SetURI(slideURI)

	// Add the part to the package.
	slideBlob, _ := slidePart.ToXML()
	slidePartOPC := opc.NewPart(slideURI, opc.ContentTypeSlide, slideBlob)
	_ = p.pkg.AddPart(slidePartOPC)

	// Wire presentation→slide and slide→layout relationships; the returned
	// relationship id is what <p:sldId r:id="…"> must carry (Phase 03 A2).
	slideRId := p.relateSlide(slidePart, slidePartOPC, layoutURI)

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
		num:          slideNum,
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

	// Resolve the slide's layout (Phase 09): a named template layout, else blank.
	layoutURI := p.resolveLayoutURI(layout...)

	// Set URI.
	slideURI := opc.NewPackURI(fmt.Sprintf("/ppt/slides/slide%d.xml", slideNum))
	slidePart.SetURI(slideURI)

	// Add to package.
	slideBlob, _ := slidePart.ToXML()
	slidePartOPC := opc.NewPart(slideURI, opc.ContentTypeSlide, slideBlob)
	_ = p.pkg.AddPart(slidePartOPC)

	// Wire presentation→slide and slide→layout relationships.
	slideRId := p.relateSlide(slidePart, slidePartOPC, layoutURI)

	// Register with PresentationPart at the same position, so the emitted
	// sldIdLst order matches the builder's slide order.
	_ = p.presentationPart.InsertSlide(index, slideRId)

	// Build the high-level object.
	s := &Slide{
		presentation: p,
		part:         slidePart,
		builder:      NewSlideBuilder(slidePart),
		mediaManager: p.mediaManager,
		index:        index,
		num:          slideNum,
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

	// Capture the presentation→slide relationship id before mutating the model
	// so we can drop the now-orphaned relationship (else it dangles and fails
	// conformance).
	slideRId, _ := p.presentationPart.SlideRelID(index)

	// Remove the slide part from the package.
	_ = p.pkg.RemovePart(s.part.PartURI())

	// Drop the presentation→slide relationship.
	if slideRId != "" {
		if presPart := p.pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml")); presPart != nil {
			_ = presPart.RemoveRelationship(slideRId)
		}
	}

	// Remove the slide's notes part, if any (RemovePart is a no-op when absent).
	_ = p.pkg.RemovePart(opc.NewPackURI(fmt.Sprintf("/ppt/notesSlides/notesSlide%d.xml", s.num)))

	// Remove from presentation.xml.
	_ = p.presentationPart.RemoveSlide(index)

	// Remove from the slice.
	p.slides = append(p.slides[:index], p.slides[index+1:]...)

	// Update indexes for subsequent slides.
	for i := index; i < len(p.slides); i++ {
		p.slides[i].index = i
	}

	// Keep section membership consistent: drop the removed index and shift
	// higher indexes down.
	for _, sec := range p.sections {
		sec.removeSlideIndex(index)
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

	if err := p.prepareForWrite(); err != nil {
		return err
	}
	return p.pkg.SaveFile(path)
}

// Write serializes the presentation and writes it to an io.Writer.
// This is suitable for high-concurrency streaming output such as HTTP responses.
func (p *Presentation) Write(w io.Writer) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if err := p.prepareForWrite(); err != nil {
		return err
	}
	return p.pkg.Save(w)
}

// WriteToBytes serializes the presentation and returns it as a byte slice.
func (p *Presentation) WriteToBytes() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if err := p.prepareForWrite(); err != nil {
		return nil, err
	}
	return p.pkg.SaveToBytes()
}

// prepareForWrite syncs every in-memory builder structure into the OPC package
// and runs the always-on hygiene pass — the shared body of every write path
// (Save / Write / WriteToBytes / SaveStream). Ordering matters: slides and their
// media/notes parts are materialized first, sections resolve slide IDs, and
// presentation.xml is serialized last so it reflects notes-master and section
// state. Callers hold p.mu.
func (p *Presentation) prepareForWrite() error {
	// Notes first: it creates notesSlide parts and adds the slide→notesSlide
	// relationship to each slide's relationship set, which syncSlides then
	// mirrors onto the package part.
	if err := p.syncNotes(); err != nil {
		return err
	}
	if err := p.syncSlides(); err != nil {
		return err
	}
	p.syncMedia()
	p.syncSections()
	if err := p.syncPresentationPart(); err != nil {
		return err
	}
	// Repair-prompt hygiene on every emitted part (D-020).
	p.applyHygiene()
	if p.logger != nil {
		p.logger.Debug("pptx: prepared deck for write", "slides", len(p.slides), "sections", len(p.sections))
	}
	return nil
}

// ============================================================================
// Sync helpers
// ============================================================================

// syncSlides serializes all slides into the OPC package and mirrors each
// slide's relationships (layout, images, notes) onto its package part, so they
// are emitted. The slide's relationships live canonically on the
// slide.SlidePart; the opc.Part is the serialization vehicle (Phase 03 C —
// closes the relationship seam A2 left open).
func (p *Presentation) syncSlides() error {
	for _, s := range p.slides {
		blob, err := s.part.ToXML()
		if err != nil {
			return fmt.Errorf("failed to serialize slide %d: %w", s.index+1, err)
		}

		// Update or create the part.
		uri := s.part.PartURI()
		existingPart := p.pkg.GetPart(uri)
		if existingPart == nil {
			existingPart = opc.NewPart(uri, opc.ContentTypeSlide, blob)
			_ = p.pkg.AddPart(existingPart)
		} else {
			existingPart.SetBlob(blob)
		}
		mirrorRelationships(existingPart, s.part.Relationships())
	}

	return nil
}

// syncMedia writes every deduplicated media resource registered on the media
// manager into the package as a part (once), so image relationships resolve.
// Content-type coverage follows from the part's content type and the
// extension-default map (the package adds an override where needed).
func (p *Presentation) syncMedia() {
	for _, res := range p.mediaManager.AllGlobalMedia() {
		uri := opc.NewPackURI("/" + res.Target())
		if p.pkg.GetPart(uri) != nil {
			continue
		}
		part := opc.NewPart(uri, res.ContentType(), res.Data())
		_ = p.pkg.AddPart(part)
	}
}

// mirrorRelationships copies every relationship from src onto dst's package part
// preserving its relationship id (so XML references stay valid). It is
// idempotent: a relationship already present on dst is left untouched.
func mirrorRelationships(dst *opc.Part, src *opc.Relationships) {
	if dst == nil || src == nil {
		return
	}
	for _, rel := range src.All() {
		if dst.Relationships().Contains(rel.RID()) {
			continue
		}
		target := ""
		if t := rel.TargetURI(); t != nil {
			target = t.URI()
		}
		nr := opc.NewRelationship(rel.RID(), rel.Type(), target, rel.IsExternal(), dst.PartURI())
		_ = dst.Relationships().Add(nr)
	}
}

// applyHygiene runs the always-on repair-prompt hygiene pass (D-020) over every
// XML part in the package, in place, just before serialization. It has no
// caller-facing switch — emitting OOXML PowerPoint opens without a repair
// prompt is correctness, not preference.
func (p *Presentation) applyHygiene() {
	for _, part := range p.pkg.AllParts() {
		if !strings.Contains(part.ContentType(), "xml") {
			continue
		}
		part.SetBlob(render.Sanitize(part.Blob()))
	}
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
		// Carry the slide's relationships (layout/image/notes) onto the clone so
		// its rId counter resumes correctly and the layout id is preserved.
		if clonedPart := newPkg.GetPart(s.part.PartURI()); clonedPart != nil {
			if rb, relErr := clonedPart.RelationshipsBlob(); relErr == nil && rb != nil {
				_ = newSlidePart.Relationships().FromXML(rb)
				if lr := newSlidePart.Relationships().GetByType(opc.RelTypeSlideLayout); len(lr) > 0 {
					newSlidePart.SetLayoutRId(lr[0].RID())
				}
			}
		}
		newSlidePart.SetShapeIDStart(maxShapeID(newSlidePart) + 1)

		ns := &Slide{
			presentation: newPres,
			part:         newSlidePart,
			builder:      NewSlideBuilder(newSlidePart),
			mediaManager: newPres.mediaManager,
			index:        i,
			num:          s.num,
		}
		// The cloned notes part (if any) survives in the cloned package; a
		// fresh notes frame is rebuilt lazily only if the caller edits notes.
		ns.shapeIDCounter.Store(1)
		newPres.slides[i] = ns
	}

	// Deep-copy the section groupings.
	for _, sec := range p.sections {
		idxCopy := make([]int, len(sec.slideIndexes))
		copy(idxCopy, sec.slideIndexes)
		newPres.sections = append(newPres.sections, &Section{name: sec.name, slideIndexes: idxCopy})
	}

	return newPres, nil
}
