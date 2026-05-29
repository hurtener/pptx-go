package presentation

import (
	"encoding/xml"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hurtener/pptx-go/internal/ooxml"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// SlideIDStart is the starting value for slide IDs.
const SlideIDStart = 256

// SlideSize holds the slide dimensions.
type SlideSize struct {
	Cx int // width in EMU (English Metric Units)
	Cy int // height in EMU
}

// StandardSlideSizes contains the standard slide size presets.
var StandardSlideSizes = struct {
	// 16:9 widescreen (12192000 x 6858000 EMU)
	Wide16x9 SlideSize
	// 4:3 standard (9144000 x 6858000 EMU)
	Standard4x3 SlideSize
}{
	Wide16x9:    SlideSize{Cx: 12192000, Cy: 6858000},
	Standard4x3: SlideSize{Cx: 9144000, Cy: 6858000},
}

// PresentationPart corresponds to /ppt/presentation.xml.
// It is the logical root of the entire PPTX.
type PresentationPart struct {
	uri *opc.PackURI

	// Slide management
	slideIDs    []uint32 // allocated slide ID list
	slideIDNext uint32   // next available slide ID (atomic)
	slideCount  int32    // current slide count (atomic)

	// Master and layout management
	slideMasterIDs []string // master rId list
	slideLayoutIDs []string // layout rId list (one-to-one with slides)

	// Global properties
	slideSize     SlideSize // slide dimensions
	notesMasterID string    // notes master rId
	themeID       string    // theme rId

	// Embedded fonts (<p:embeddedFontLst>) recorded via AddEmbeddedFont.
	embeddedFonts []EmbeddedFontEntry

	// Sections (<p14:sectionLst> in extLst). Emitted last, after the core
	// CT_Presentation children (D-021, RFC §8.7).
	sections []SectionEntry

	mu sync.RWMutex
}

// SectionEntry is one named slide grouping: a display name, a unique {GUID}
// id, and the slide IDs (the sldId/@id values) it contains in order.
type SectionEntry struct {
	Name     string
	ID       string // braced GUID, e.g. {00000000-0000-0000-0000-000000000001}
	SlideIDs []uint32
}

// NewPresentationPart creates a new presentation part.
func NewPresentationPart() *PresentationPart {
	return &PresentationPart{
		uri:            opc.NewPackURI("/ppt/presentation.xml"),
		slideIDs:       make([]uint32, 0),
		slideMasterIDs: make([]string, 0),
		slideLayoutIDs: make([]string, 0),
		slideSize:      StandardSlideSizes.Wide16x9,
		slideIDNext:    SlideIDStart,
	}
}

// NewPresentationPartWithSize creates a presentation part with the given slide size.
func NewPresentationPartWithSize(size SlideSize) *PresentationPart {
	p := NewPresentationPart()
	p.slideSize = size
	return p
}

// PartURI returns the part URI.
func (p *PresentationPart) PartURI() *opc.PackURI {
	return p.uri
}

// SlideSize returns the slide dimensions.
func (p *PresentationPart) SlideSize() SlideSize {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.slideSize
}

// SetSlideSize sets the slide dimensions.
func (p *PresentationPart) SetSlideSize(size SlideSize) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.slideSize = size
}

// SlideCount returns the number of slides.
func (p *PresentationPart) SlideCount() int32 {
	return atomic.LoadInt32(&p.slideCount)
}

// SlideIDAt returns the slide ID at the given index.
func (p *PresentationPart) SlideIDAt(index int) (uint32, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if index < 0 || index >= len(p.slideIDs) {
		return 0, fmt.Errorf("slide index out of range: %d", index)
	}
	return p.slideIDs[index], nil
}

// SlideIDs returns all slide IDs.
func (p *PresentationPart) SlideIDs() []uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := make([]uint32, len(p.slideIDs))
	copy(ids, p.slideIDs)
	return ids
}

// SlideMasterIDs returns all master relationship IDs.
func (p *PresentationPart) SlideMasterIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	ids := make([]string, len(p.slideMasterIDs))
	copy(ids, p.slideMasterIDs)
	return ids
}

// AddSlideMaster adds a master by its relationship ID.
func (p *PresentationPart) AddSlideMaster(rId string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.slideMasterIDs = append(p.slideMasterIDs, rId)
}

// SetNotesMaster records the notes-master relationship ID, emitted in
// <p:notesMasterIdLst> (D-022).
func (p *PresentationPart) SetNotesMaster(rId string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.notesMasterID = rId
}

// SetSections sets the slide-grouping sections emitted in extLst (D-021). A nil
// or empty slice emits no section list.
func (p *PresentationPart) SetSections(secs []SectionEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sections = secs
}

// Sections returns the parsed/assigned sections.
func (p *PresentationPart) Sections() []SectionEntry {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]SectionEntry, len(p.sections))
	copy(out, p.sections)
	return out
}

// allocateSlideID atomically allocates a new slide ID.
func (p *PresentationPart) allocateSlideID() uint32 {
	return atomic.AddUint32(&p.slideIDNext, 1)
}

// AddSlide adds a slide.
// layoutRId is the relationship ID of the associated layout; slidePart is the actual slide part.
func (p *PresentationPart) AddSlide(layoutRId string, slidePart *slide.SlidePart) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Allocate a slide ID.
	slideID := p.allocateSlideID()

	// Update internal state.
	p.slideIDs = append(p.slideIDs, slideID)
	p.slideLayoutIDs = append(p.slideLayoutIDs, layoutRId)
	atomic.AddInt32(&p.slideCount, 1)

	return nil
}

// InsertSlide inserts a slide at the given index, allocating a new slide ID and
// recording its presentation→slide relationship id at the same position, so the
// emitted <p:sldIdLst> order matches the builder's slide order.
func (p *PresentationPart) InsertSlide(index int, rId string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index > len(p.slideIDs) {
		return fmt.Errorf("slide index out of range: %d", index)
	}

	slideID := p.allocateSlideID()
	p.slideIDs = append(p.slideIDs[:index], append([]uint32{slideID}, p.slideIDs[index:]...)...)
	p.slideLayoutIDs = append(p.slideLayoutIDs[:index], append([]string{rId}, p.slideLayoutIDs[index:]...)...)
	atomic.AddInt32(&p.slideCount, 1)
	return nil
}

// SlideRelID returns the presentation→slide relationship id recorded for the
// slide at the given index (the value <p:sldId r:id> carries).
func (p *PresentationPart) SlideRelID(index int) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if index < 0 || index >= len(p.slideLayoutIDs) {
		return "", fmt.Errorf("slide index out of range: %d", index)
	}
	return p.slideLayoutIDs[index], nil
}

// RemoveSlide removes the slide at the given index.
func (p *PresentationPart) RemoveSlide(index int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index >= len(p.slideIDs) {
		return fmt.Errorf("slide index out of range: %d", index)
	}

	// Remove from slices.
	p.slideIDs = append(p.slideIDs[:index], p.slideIDs[index+1:]...)
	p.slideLayoutIDs = append(p.slideLayoutIDs[:index], p.slideLayoutIDs[index+1:]...)
	atomic.AddInt32(&p.slideCount, -1)

	return nil
}

// XPresentation is the complete XML structure for presentation.xml.
// Note: XML tags use unprefixed names because namespace prefixes are stripped before parsing.
// Field order matches the CT_Presentation schema sequence — PowerPoint is
// order-sensitive: sldMasterIdLst, notesMasterIdLst, sldIdLst, sldSz, notesSz,
// embeddedFontLst.
type XPresentation struct {
	XMLName xml.Name `xml:"presentation"`

	// Master ID list (first child)
	SldMasterIdLst *XSldMasterIdLst `xml:"sldMasterIdLst,omitempty"`

	// Notes master ID list
	NotesMasterIdLst *XSldMasterIdLst `xml:"notesMasterIdLst,omitempty"`

	// Slide ID list
	SldIdLst *XSldIdLst `xml:"sldIdLst,omitempty"`

	// Slide size (required)
	SldSz *XSldSz `xml:"sldSz"`

	// Notes size
	NotesSz *XSldSz `xml:"notesSz,omitempty"`

	// Embedded font list (<p:embeddedFontLst>)
	EmbeddedFontLst *XEmbeddedFontList `xml:"embeddedFontLst,omitempty"`

	// Extension list (last child). Only populated on read: the section list
	// lives under a p14: namespace that the single-table RestoreNamespaces pass
	// can't emit (the local name "sldId" collides with the top-level slide
	// list), so it is injected as a literal fragment on write and parsed back
	// here on read.
	ExtLst *XExtLst `xml:"extLst,omitempty"`
}

// XExtLst is the presentation extension list (read-side capture).
type XExtLst struct {
	Exts []XExt `xml:"ext"`
}

// XExt is one extension; only the section list is recognized.
type XExt struct {
	URI        string       `xml:"uri,attr"`
	SectionLst *XSectionLst `xml:"sectionLst,omitempty"`
}

// XSectionLst is the p14 section list (prefix stripped on read).
type XSectionLst struct {
	Sections []XSection `xml:"section"`
}

// XSection is one named section grouping a set of slide IDs.
type XSection struct {
	Name     string           `xml:"name,attr"`
	ID       string           `xml:"id,attr"`
	SldIdLst XSectionSldIdLst `xml:"sldIdLst"`
}

// XSectionSldIdLst is a section's slide-id list.
type XSectionSldIdLst struct {
	SldIds []XSectionSldId `xml:"sldId"`
}

// XSectionSldId references a slide by its sldId/@id value.
type XSectionSldId struct {
	ID uint32 `xml:"id,attr"`
}

// SectionLstExtURI is the registered extension URI for the p14 section list.
const SectionLstExtURI = "{521415D9-36F7-43E2-AB2F-B90AF26B5E84}"

// p14Namespace is the PowerPoint 2010 main namespace carrying the section list.
const p14Namespace = "http://schemas.microsoft.com/office/powerpoint/2010/main"

// XCompatibility holds compatibility settings.
type XCompatibility struct {
	CompatMode string `xml:"compatMode,attr,omitempty"`
}

// XSldSz holds slide dimensions.
type XSldSz struct {
	Cx int `xml:"cx,attr"`
	Cy int `xml:"cy,attr"`
}

// XSldIdLst holds the slide ID list.
type XSldIdLst struct {
	SldIds []XSldId `xml:"sldId"`
}

// XSldId holds a single slide ID entry.
type XSldId struct {
	Id  uint32 `xml:"id,attr"`
	RId string `xml:"rid,attr"`
}

// XSldMasterIdLst holds the master ID list.
type XSldMasterIdLst struct {
	SldMasterIds []XSldMasterId `xml:"sldMasterId"`
}

// XSldMasterId holds a single master ID entry.
type XSldMasterId struct {
	Id  uint32 `xml:"id,attr"`
	RId string `xml:"rid,attr"`
}

// XPrintSettings holds print settings.
type XPrintSettings struct {
	OutputOptions *XOutputOptions `xml:"outputOptions,omitempty"`
}

// XOutputOptions holds output options.
type XOutputOptions struct {
	UsePrintFml     *bool `xml:"usePrintFml,attr,omitempty"`
	CloneLinkedObjs *bool `xml:"cloneLinkedObjs,attr,omitempty"`
}

// ToXML serializes the PresentationPart to XML.
func (p *PresentationPart) ToXML() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	xp := XPresentation{
		SldSz: &XSldSz{
			Cx: p.slideSize.Cx,
			Cy: p.slideSize.Cy,
		},
		// Notes pages are portrait Letter by convention (CT_Presentation
		// requires notesSz after sldSz when present).
		NotesSz: &XSldSz{Cx: 6858000, Cy: 9144000},
	}

	// Build slide ID list.
	if len(p.slideIDs) > 0 {
		xp.SldIdLst = &XSldIdLst{
			SldIds: make([]XSldId, len(p.slideIDs)),
		}
		for i := range p.slideIDs {
			xp.SldIdLst.SldIds[i] = XSldId{
				Id:  p.slideIDs[i],
				RId: p.slideLayoutIDs[i],
			}
		}
	}

	// Build master ID list.
	if len(p.slideMasterIDs) > 0 {
		xp.SldMasterIdLst = &XSldMasterIdLst{
			SldMasterIds: make([]XSldMasterId, len(p.slideMasterIDs)),
		}
		// CT_SlideMasterIdListEntry/@id is ST_SlideMasterId: PowerPoint requires
		// it to be >= 2147483648.
		for i, rId := range p.slideMasterIDs {
			xp.SldMasterIdLst.SldMasterIds[i] = XSldMasterId{
				Id:  uint32(2147483648 + i),
				RId: rId,
			}
		}
	}

	// Build notes-master ID list (single entry; D-022). Its @id is also an
	// ST_SlideMasterId (>= 2147483648).
	if p.notesMasterID != "" {
		xp.NotesMasterIdLst = &XSldMasterIdLst{
			SldMasterIds: []XSldMasterId{{Id: 2147483648, RId: p.notesMasterID}},
		}
	}

	// Build the embedded font list.
	xp.EmbeddedFontLst = buildEmbeddedFontList(p.embeddedFonts)

	output, err := xml.Marshal(&xp)
	if err != nil {
		return nil, err
	}
	// Bare element names → canonical p:/r: prefixes + root namespace
	// declarations (D-032); without this the root <presentation> carries no
	// namespace and PowerPoint rejects the file.
	restored, err := ooxml.RestoreNamespaces(output)
	if err != nil {
		return nil, fmt.Errorf("restore presentation namespaces: %w", err)
	}
	// Inject the section list (extLst) as a literal p14 fragment — the
	// single-table namespace pass can't emit it (D-021).
	restored = injectSectionLst(restored, p.sections)
	return append([]byte(ooxml.XMLDeclaration), restored...), nil
}

// injectSectionLst appends the p14 section-list extension as the last child of
// <p:presentation>. It returns data unchanged when there are no sections.
func injectSectionLst(data []byte, secs []SectionEntry) []byte {
	if len(secs) == 0 {
		return data
	}
	const close = "</p:presentation>"
	idx := strings.LastIndex(string(data), close)
	if idx < 0 {
		return data
	}

	var b strings.Builder
	b.WriteString(`<p:extLst><p:ext uri="`)
	b.WriteString(SectionLstExtURI)
	b.WriteString(`"><p14:sectionLst xmlns:p14="`)
	b.WriteString(p14Namespace)
	b.WriteString(`">`)
	for _, s := range secs {
		b.WriteString(`<p14:section name="`)
		b.WriteString(escapeXMLAttr(s.Name))
		b.WriteString(`" id="`)
		b.WriteString(s.ID)
		b.WriteString(`"><p14:sldIdLst>`)
		for _, id := range s.SlideIDs {
			fmt.Fprintf(&b, `<p14:sldId id="%d"/>`, id)
		}
		b.WriteString(`</p14:sldIdLst></p14:section>`)
	}
	b.WriteString(`</p14:sectionLst></p:ext></p:extLst>`)

	out := make([]byte, 0, len(data)+b.Len())
	out = append(out, data[:idx]...)
	out = append(out, b.String()...)
	out = append(out, data[idx:]...)
	return out
}

// escapeXMLAttr escapes the five XML metacharacters for an attribute value.
func escapeXMLAttr(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return r.Replace(s)
}

// FromXML deserializes a PresentationPart from XML.
func (p *PresentationPart) FromXML(data []byte) error {
	// Strip namespace prefixes for compatibility with Go's xml.Unmarshal.
	cleanData, err := ooxml.StripNamespacePrefixes(data)
	if err != nil {
		return fmt.Errorf("failed to clean XML: %w", err)
	}

	var xp XPresentation
	if err := xml.Unmarshal(cleanData, &xp); err != nil {
		return fmt.Errorf("failed to unmarshal presentation XML: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Parse slide dimensions.
	if xp.SldSz != nil {
		p.slideSize = SlideSize{
			Cx: xp.SldSz.Cx,
			Cy: xp.SldSz.Cy,
		}
	}

	// Parse slide list.
	p.slideIDs = make([]uint32, 0)
	p.slideLayoutIDs = make([]string, 0)
	if xp.SldIdLst != nil {
		for _, sldId := range xp.SldIdLst.SldIds {
			p.slideIDs = append(p.slideIDs, sldId.Id)
			p.slideLayoutIDs = append(p.slideLayoutIDs, sldId.RId)
		}
	}

	// Update slide ID counter.
	if len(p.slideIDs) > 0 {
		maxID := p.slideIDs[0]
		for _, id := range p.slideIDs {
			if id > maxID {
				maxID = id
			}
		}
		p.slideIDNext = maxID + 1
	}

	// Parse master list.
	p.slideMasterIDs = make([]string, 0)
	if xp.SldMasterIdLst != nil {
		for _, masterId := range xp.SldMasterIdLst.SldMasterIds {
			p.slideMasterIDs = append(p.slideMasterIDs, masterId.RId)
		}
	}

	// Parse notes-master list (single entry by convention).
	p.notesMasterID = ""
	if xp.NotesMasterIdLst != nil && len(xp.NotesMasterIdLst.SldMasterIds) > 0 {
		p.notesMasterID = xp.NotesMasterIdLst.SldMasterIds[0].RId
	}

	// Parse section list from extLst (D-021).
	p.sections = nil
	if xp.ExtLst != nil {
		for _, ext := range xp.ExtLst.Exts {
			if ext.SectionLst == nil {
				continue
			}
			for _, xs := range ext.SectionLst.Sections {
				se := SectionEntry{Name: xs.Name, ID: xs.ID}
				for _, sid := range xs.SldIdLst.SldIds {
					se.SlideIDs = append(se.SlideIDs, sid.ID)
				}
				p.sections = append(p.sections, se)
			}
		}
	}

	// Update slide count.
	p.slideCount = int32(len(p.slideIDs))

	return nil
}

// Presentation helper functions

// NewSlideSizeFromStandard creates a SlideSize from a standard size name.
func NewSlideSizeFromStandard(name string) SlideSize {
	switch name {
	case "16:9", "wide", "widescreen":
		return StandardSlideSizes.Wide16x9
	case "4:3", "standard":
		return StandardSlideSizes.Standard4x3
	default:
		return StandardSlideSizes.Wide16x9
	}
}

// EMUFromPoints converts points to EMU.
func EMUFromPoints(points float64) int {
	return int(points * 12700)
}

// PointsFromEMU converts EMU to points.
func PointsFromEMU(emu int) float64 {
	return float64(emu) / 12700.0
}

// EMUFromInches converts inches to EMU.
func EMUFromInches(inches float64) int {
	return int(inches * 914400)
}

// InchesFromEMU converts EMU to inches.
func InchesFromEMU(emu int) float64 {
	return float64(emu) / 914400.0
}

// EMUFromMM converts millimeters to EMU.
func EMUFromMM(mm float64) int {
	return int(mm * 36000)
}

// MMFromEMU converts EMU to millimeters.
func MMFromEMU(emu int) float64 {
	return float64(emu) / 36000.0
}
