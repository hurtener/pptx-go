package pptx

import (
	"strings"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// Part URIs for the seeded scaffold (Phase 03 A2). A single master + blank
// layout + theme make every New() deck complete and valid (RFC §8.7).
const (
	themeURI       = "/ppt/theme/theme1.xml"
	slideMasterURI = "/ppt/slideMasters/slideMaster1.xml"
	slideLayoutURI = "/ppt/slideLayouts/slideLayout1.xml"
)

// seedScaffold adds the theme, slide master and blank slide layout to the
// package and wires their relationships, so a brand-new presentation already
// satisfies the full-deck conformance gate (master/layout/theme + resolved
// rels) and opens in PowerPoint. AddSlide then only has to add each slide and
// relate it to the layout.
//
// Relationship graph:
//
//	presentation.xml ──slideMaster──▶ slideMaster1.xml
//	slideMaster1.xml ──slideLayout──▶ slideLayout1.xml
//	slideMaster1.xml ──theme───────▶ theme1.xml
//	slideLayout1.xml ──slideMaster─▶ slideMaster1.xml
//	slideN.xml       ──slideLayout──▶ slideLayout1.xml   (AddSlide)
func (p *Presentation) seedScaffold() {
	// 1. Theme.
	themePart := opc.NewPart(opc.NewPackURI(themeURI), opc.ContentTypeTheme, []byte(scaffoldThemeXML))
	_ = p.pkg.AddPart(themePart)

	// 2. Slide layout (blank) → master.
	layoutPart := opc.NewPart(opc.NewPackURI(slideLayoutURI), opc.ContentTypeSlideLayout, []byte(scaffoldSlideLayoutXML))
	_, _ = layoutPart.AddRelationship(opc.RelTypeSlideMaster, "../slideMasters/slideMaster1.xml", false)
	_ = p.pkg.AddPart(layoutPart)

	// 3. Slide master → layout (+ theme). The layout relationship id is
	//    substituted into the master's sldLayoutIdLst.
	masterPart := opc.NewPart(opc.NewPackURI(slideMasterURI), opc.ContentTypeSlideMaster, nil)
	layoutRel, _ := masterPart.AddRelationship(opc.RelTypeSlideLayout, "../slideLayouts/slideLayout1.xml", false)
	_, _ = masterPart.AddRelationship(opc.RelTypeTheme, "../theme/theme1.xml", false)
	masterXML := strings.Replace(scaffoldSlideMasterXML, "%LAYOUT_RID%", layoutRel.RID(), 1)
	masterPart.SetBlob([]byte(masterXML))
	_ = p.pkg.AddPart(masterPart)

	// 4. presentation.xml → master, recorded in <p:sldMasterIdLst>.
	presPart := p.pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))
	if presPart != nil {
		masterRel, _ := presPart.AddRelationship(opc.RelTypeSlideMaster, "slideMasters/slideMaster1.xml", false)
		p.presentationPart.AddSlideMaster(masterRel.RID())
	}

	// 5. Presentation auxiliary parts. PowerPoint expects presProps, viewProps
	//    and tableStyles (a table references a tableStyleId); a deck missing them
	//    opens but prompts to "repair". Related from presentation.xml.
	p.seedPart(opc.NewPackURI("/ppt/presProps.xml"), opc.ContentTypePresProps, scaffoldPresPropsXML)
	p.seedPart(opc.NewPackURI("/ppt/viewProps.xml"), opc.ContentTypeViewProps, scaffoldViewPropsXML)
	p.seedPart(opc.NewPackURI("/ppt/tableStyles.xml"), opc.ContentTypeTableStyles, scaffoldTableStylesXML)
	if presPart != nil {
		_, _ = presPart.AddRelationship(opc.RelTypePresProps, "presProps.xml", false)
		_, _ = presPart.AddRelationship(opc.RelTypeViewProps, "viewProps.xml", false)
		_, _ = presPart.AddRelationship(opc.RelTypeTableStyles, "tableStyles.xml", false)
	}

	// 6. Document properties (core + app), related from the package.
	p.seedPart(opc.NewPackURI("/docProps/core.xml"), opc.ContentTypeCoreProperties, scaffoldCorePropsXML)
	p.seedPart(opc.NewPackURI("/docProps/app.xml"), opc.ContentTypeExtendedProperties, scaffoldAppPropsXML)
	_, _ = p.pkg.AddRelationship(opc.RelTypeCoreProperties, "docProps/core.xml", false)
	_, _ = p.pkg.AddRelationship(opc.RelTypeExtendedProperties, "docProps/app.xml", false)
}

// seedPart adds a hand-authored scaffold part to the package.
func (p *Presentation) seedPart(uri *opc.PackURI, contentType, xml string) {
	_ = p.pkg.AddPart(opc.NewPart(uri, contentType, []byte(xml)))
}

// relateSlide adds the presentation→slide and slide→layout relationships for a
// freshly created slide part, returning the presentation-relative relationship
// id that <p:sldId> must carry. The slide→layout relationship is added to the
// slide part's own relationship set (its single rId namespace, shared with image
// and notes relationships); syncSlides mirrors that set onto the package part.
func (p *Presentation) relateSlide(slidePart *slide.SlidePart, slidePartOPC *opc.Part) string {
	// slide → layout (allocated in the slide's rId namespace).
	_, _ = slidePart.Relationships().AddNew(opc.RelTypeSlideLayout, "../slideLayouts/slideLayout1.xml", false)

	// presentation → slide.
	presPart := p.pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))
	if presPart == nil {
		return ""
	}
	target := strings.TrimPrefix(slidePartOPC.PartURI().URI(), "/ppt/")
	rel, _ := presPart.AddRelationship(opc.RelTypeSlide, target, false)
	return rel.RID()
}
