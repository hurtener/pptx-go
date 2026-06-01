package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/ooxml/presentation"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// Template ingestion — FromTemplate (RFC §13.1)
// ============================================================================
//
// A "brand kit" is a .pptx template: a populated theme + ≥1 master with
// layouts (RFC §13.4). The caller opens it like any deck — its theme and
// masters are extracted on open (see loadPresentationPart) — then seeds a new
// presentation from it:
//
//	brand, err := pptx.OpenStream("brand-template.pptx")
//	if err != nil { return err }
//	defer brand.Close()
//	pres := pptx.New(pptx.FromTemplate(brand))   // theme + masters + layouts adopted
//
// Ingestion copies the template's parts wholesale (theme, masters, layouts, and
// the auxiliary parts PowerPoint expects) by cloning its package, then strips
// any slides so the new deck starts empty. Cloning preserves the template's
// already-valid relationship graph, so ingestion never hand-rewires masters and
// can't reintroduce the repair-prompt class of bug. (D-NNN — wholesale copy.)

// FromTemplate returns a New option that seeds the presentation from brand: its
// theme becomes the active theme, and its masters + layouts are adopted (and
// reachable via Masters()). A nil brand is ignored (the deck falls back to the
// default scaffold). The brand presentation is not retained or mutated.
func FromTemplate(brand *Presentation) Option {
	return func(p *Presentation) { p.template = brand }
}

// adoptTemplate replaces the presentation's seeded scaffold with the template's
// theme/master/layout/auxiliary parts, by cloning the template package and
// stripping its slides. Called by New (in place of seedScaffold) when a template
// was supplied. On any failure the caller falls back to the default scaffold so
// New never yields a broken deck.
func (p *Presentation) adoptTemplate() error {
	src := p.template
	if src == nil || src.pkg == nil {
		return fmt.Errorf("pptx: FromTemplate: nil template")
	}

	src.mu.RLock()
	clone := src.pkg.Clone()
	srcTheme := src.theme
	src.mu.RUnlock()
	if clone == nil {
		return fmt.Errorf("pptx: FromTemplate: clone template package")
	}
	p.pkg = clone

	// Re-parse presentation.xml from the cloned package.
	p.presentationPart = presentation.NewPresentationPart()
	presPart := p.pkg.GetPartByRelType(opc.RelTypeOfficeDocument)
	if presPart == nil {
		return fmt.Errorf("pptx: FromTemplate: template has no presentation part")
	}
	if err := p.presentationPart.FromXML(presPart.Blob()); err != nil {
		return fmt.Errorf("pptx: FromTemplate: parse presentation.xml: %w", err)
	}

	// Start empty: drop any slides the template carried (a template is usually
	// slide-free, but a sample slide must not leak into the caller's deck).
	p.clearTemplateSlides()

	// Adopt the theme (the template's, extracted on its open) and the read-only
	// master/layout registry.
	if srcTheme != nil {
		p.theme = srcTheme.Clone()
	}
	p.masters = buildMasterRegistry(p.pkg)

	// New slides start fresh; media/section state is empty.
	p.slideCounter = 0
	p.sections = nil
	return nil
}

// clearTemplateSlides removes every slide part + presentation→slide relationship
// + sldIdLst entry from the (cloned) package, leaving the masters, layouts,
// theme and auxiliary parts intact. Caller is a constructor (value not yet
// shared).
func (p *Presentation) clearTemplateSlides() {
	presPart := p.pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))

	// Collect slide relationship ids before mutating the model.
	var relIDs []string
	for i := 0; i < int(p.presentationPart.SlideCount()); i++ {
		if rid, err := p.presentationPart.SlideRelID(i); err == nil && rid != "" {
			relIDs = append(relIDs, rid)
		}
	}
	for _, rid := range relIDs {
		if presPart != nil {
			if rel := presPart.Relationships().Get(rid); rel != nil && rel.TargetURI() != nil {
				_ = p.pkg.RemovePart(rel.TargetURI())
			}
			_ = presPart.RemoveRelationship(rid)
		}
	}
	// Remove from the end so indices stay valid.
	for i := int(p.presentationPart.SlideCount()) - 1; i >= 0; i-- {
		_ = p.presentationPart.RemoveSlide(i)
	}
	p.slides = nil
}
