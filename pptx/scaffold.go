package pptx

import (
	"strings"

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
}

// relateSlide adds the presentation→slide and slide→layout relationships for a
// freshly created slide part, returning the presentation-relative relationship
// id that <p:sldId> must carry.
func (p *Presentation) relateSlide(slidePartOPC *opc.Part) string {
	// slide → layout.
	_, _ = slidePartOPC.AddRelationship(opc.RelTypeSlideLayout, "../slideLayouts/slideLayout1.xml", false)

	// presentation → slide.
	presPart := p.pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))
	if presPart == nil {
		return ""
	}
	target := strings.TrimPrefix(slidePartOPC.PartURI().URI(), "/ppt/")
	rel, _ := presPart.AddRelationship(opc.RelTypeSlide, target, false)
	return rel.RID()
}
