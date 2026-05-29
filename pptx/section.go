package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/ooxml/presentation"
)

// ============================================================================
// Sections — named slide groupings (RFC §8.7, D-021)
// ============================================================================
//
// PowerPoint groups slides into named sections (shown in the slide sorter and
// the outline) via the p14:sectionLst extension on the presentation. pptx-go
// exposes them as a first-class builder primitive. A slide belongs to at most
// one section; slides left unassigned fall into an implicit leading "Default
// Section" so the emitted section list covers every slide (PowerPoint requires
// each slide to appear in exactly one section once any section exists).
//
// This is distinct from the scene IR's section_divider (a full-bleed chapter
// slide), which may or may not live inside a pptx.Section.

// DefaultSectionName is the name of the implicit section that holds slides not
// assigned to any caller-created section.
const DefaultSectionName = "Default Section"

// Section is a named, ordered grouping of slides. Create one with
// Presentation.AddSection and assign slides with Include.
type Section struct {
	name string
	// slideIndexes are the zero-based indexes (into Presentation.slides) of the
	// member slides, in insertion order.
	slideIndexes []int
}

// Name returns the section's display name.
func (sec *Section) Name() string { return sec.name }

// AddSection appends a new, empty section and returns it. Sections are emitted
// in creation order.
func (p *Presentation) AddSection(name string) *Section {
	p.mu.Lock()
	defer p.mu.Unlock()
	sec := &Section{name: name}
	p.sections = append(p.sections, sec)
	return sec
}

// Sections returns the presentation's sections in creation order.
func (p *Presentation) Sections() []*Section {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Section, len(p.sections))
	copy(out, p.sections)
	return out
}

// Include assigns a slide to the section (idempotent). A slide assigned to more
// than one section is emitted under the first that includes it.
func (sec *Section) Include(s *Slide) {
	if sec == nil || s == nil {
		return
	}
	for _, i := range sec.slideIndexes {
		if i == s.index {
			return
		}
	}
	sec.slideIndexes = append(sec.slideIndexes, s.index)
}

// removeSlideIndex drops a removed slide's index from the section and shifts
// every higher index down by one, keeping membership consistent after
// RemoveSlide.
func (sec *Section) removeSlideIndex(removed int) {
	kept := sec.slideIndexes[:0]
	for _, idx := range sec.slideIndexes {
		switch {
		case idx == removed:
			// drop it
		case idx > removed:
			kept = append(kept, idx-1)
		default:
			kept = append(kept, idx)
		}
	}
	sec.slideIndexes = kept
}

// syncSections resolves the builder's sections into presentation-part section
// entries (slide indexes → slide IDs) and records them for emission. It runs
// before syncPresentationPart. Callers hold p.mu.
func (p *Presentation) syncSections() {
	if len(p.sections) == 0 {
		p.presentationPart.SetSections(nil)
		return
	}

	slideIDs := p.presentationPart.SlideIDs() // parallel to p.slides order
	idFor := func(idx int) (uint32, bool) {
		if idx < 0 || idx >= len(slideIDs) {
			return 0, false
		}
		return slideIDs[idx], true
	}

	assigned := make(map[int]bool)
	var entries []presentation.SectionEntry

	for i, sec := range p.sections {
		entry := presentation.SectionEntry{
			Name: sec.name,
			ID:   sectionGUID(i + 1),
		}
		for _, idx := range sec.slideIndexes {
			if assigned[idx] {
				continue // a slide lives in only the first section that claims it
			}
			id, ok := idFor(idx)
			if !ok {
				continue
			}
			assigned[idx] = true
			entry.SlideIDs = append(entry.SlideIDs, id)
		}
		entries = append(entries, entry)
	}

	// Cover any unassigned slides with a leading implicit default section so the
	// section list spans every slide.
	var defaults []uint32
	for idx := range p.slides {
		if assigned[idx] {
			continue
		}
		if id, ok := idFor(idx); ok {
			defaults = append(defaults, id)
		}
	}
	if len(defaults) > 0 {
		def := presentation.SectionEntry{
			Name:     DefaultSectionName,
			ID:       sectionGUID(0),
			SlideIDs: defaults,
		}
		entries = append([]presentation.SectionEntry{def}, entries...)
	}

	p.presentationPart.SetSections(entries)
}

// sectionGUID builds a deterministic, unique braced GUID for section n. Section
// GUIDs only need to be unique within the file; deriving them from a counter
// keeps emitted decks byte-stable for golden tests.
func sectionGUID(n int) string {
	return fmt.Sprintf("{00000000-0000-0000-0000-%012X}", n)
}
