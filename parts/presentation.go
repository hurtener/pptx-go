package parts

import (
	"encoding/xml"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/hurtener/pptx-go/opc"
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

	mu sync.RWMutex
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

// allocateSlideID atomically allocates a new slide ID.
func (p *PresentationPart) allocateSlideID() uint32 {
	return atomic.AddUint32(&p.slideIDNext, 1)
}

// AddSlide adds a slide.
// layoutRId is the relationship ID of the associated layout; slidePart is the actual slide part.
func (p *PresentationPart) AddSlide(layoutRId string, slidePart *SlidePart) error {
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
type XPresentation struct {
	XMLName xml.Name `xml:"presentation"`

	// Compatibility settings
	Compatibility *XCompatibility `xml:"compatSpt,omitempty"`

	// Slide size (required)
	SldSz *XSldSz `xml:"sldSz"`

	// Notes size
	NotesSz *XSldSz `xml:"notesSz,omitempty"`

	// Slide ID list (required)
	SldIdLst *XSldIdLst `xml:"sldIdLst"`

	// Master ID list
	SldMasterIdLst *XSldMasterIdLst `xml:"sldMasterIdLst,omitempty"`

	// Notes master ID list
	NotesMasterIdLst *XSldMasterIdLst `xml:"notesMasterIdLst,omitempty"`

	// Print settings
	PrintSettings *XPrintSettings `xml:"printSettings,omitempty"`
}

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
		// Master IDs start at 1.
		for i, rId := range p.slideMasterIDs {
			xp.SldMasterIdLst.SldMasterIds[i] = XSldMasterId{
				Id:  uint32(i + 1),
				RId: rId,
			}
		}
	}

	output, err := xml.Marshal(&xp)
	if err != nil {
		return nil, err
	}
	return append([]byte(XMLDeclaration), output...), nil
}

// FromXML deserializes a PresentationPart from XML.
func (p *PresentationPart) FromXML(data []byte) error {
	// Strip namespace prefixes for compatibility with Go's xml.Unmarshal.
	cleanData, err := StripNamespacePrefixes(data)
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
