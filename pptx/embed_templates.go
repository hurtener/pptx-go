// Package pptx provides a high-level API for authoring PPTX files.
package pptx

import (
	"fmt"
	"sync"

	"github.com/hurtener/pptx-go/opc"
	"github.com/hurtener/pptx-go/parts"
)

// ============================================================================
// Embedded template system
// ============================================================================
//
// Design principles:
// 1. Programmatic generation — templates are built in code; no external files
//    are required at runtime.
// 2. Lazy initialization — sync.Once ensures each template is created at most
//    once.
// 3. Zero dependencies — no external template files are needed at runtime.
// 4. High performance — templates are created once and reused via Clone.
//
// To embed actual .pptx template files instead:
// 1. Place the .pptx files under pptx/templates/.
// 2. Create an embed_fs.go file and add the appropriate //go:embed directive.
// ============================================================================

// EmbeddedTemplateManager creates and caches presentation templates
// programmatically.
type EmbeddedTemplateManager struct {
	// template cache: template name -> *opc.Package
	templates sync.Map

	// initialization guard
	once sync.Once

	// any error captured during initialization
	initErr error

	// cached master/layout data
	masterCache *MasterCache
}

// global EmbeddedTemplateManager instance
var globalEmbeddedTemplates = &EmbeddedTemplateManager{}

// GetEmbeddedTemplateManager returns the global EmbeddedTemplateManager.
func GetEmbeddedTemplateManager() *EmbeddedTemplateManager {
	return globalEmbeddedTemplates
}

// ============================================================================
// Initialization
// ============================================================================

// Init initializes all embedded templates. Only the first call has any effect.
func (etm *EmbeddedTemplateManager) Init() error {
	etm.once.Do(func() {
		etm.initErr = etm.createAllTemplates()
	})
	return etm.initErr
}

// createAllTemplates builds and caches every built-in template.
func (etm *EmbeddedTemplateManager) createAllTemplates() error {
	// default template (16:9 widescreen)
	defaultTmpl := etm.createMinimalTemplate()
	etm.templates.Store(TemplateDefault, defaultTmpl)

	// blank template reuses the default
	etm.templates.Store(TemplateBlank, defaultTmpl)

	// widescreen template (16:9) reuses the default
	etm.templates.Store(TemplateWide, defaultTmpl)

	// standard template (4:3)
	standardTmpl := etm.createStandardTemplate()
	etm.templates.Store(TemplateStandard, standardTmpl)

	return nil
}

// createMinimalTemplate builds a minimal 16:9 widescreen template.
func (etm *EmbeddedTemplateManager) createMinimalTemplate() *opc.Package {
	pkg := opc.NewPackage()

	// create presentation.xml
	presPart := parts.NewPresentationPart()
	presPart.SetSlideSize(parts.SlideSize{Cx: 12192000, Cy: 6858000}) // 16:9 (13.333" x 7.5")

	presURI := opc.NewPackURI("/ppt/presentation.xml")
	presBlob, _ := presPart.ToXML()
	_ = pkg.AddPart(opc.NewPart(presURI, opc.ContentTypePresentation, presBlob))

	// add the package-level relationship
	_, _ = pkg.AddRelationship(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)

	return pkg
}

// createStandardTemplate builds a minimal 4:3 standard template.
func (etm *EmbeddedTemplateManager) createStandardTemplate() *opc.Package {
	pkg := opc.NewPackage()

	// create presentation.xml
	presPart := parts.NewPresentationPart()
	presPart.SetSlideSize(parts.SlideSize{Cx: 9144000, Cy: 6858000}) // 4:3 (10" x 7.5")

	presURI := opc.NewPackURI("/ppt/presentation.xml")
	presBlob, _ := presPart.ToXML()
	_ = pkg.AddPart(opc.NewPart(presURI, opc.ContentTypePresentation, presBlob))

	// add the package-level relationship
	_, _ = pkg.AddRelationship(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)

	return pkg
}

// ============================================================================
// Template retrieval
// ============================================================================

// GetTemplate returns a clone of the named template, initializing it on first
// use.
func (etm *EmbeddedTemplateManager) GetTemplate(name TemplateType) (*opc.Package, error) {
	// ensure templates have been created
	if err := etm.Init(); err != nil {
		return nil, err
	}

	val, ok := etm.templates.Load(name)
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}

	pkg := val.(*opc.Package)
	return pkg.Clone(), nil
}

// GetDefaultTemplate returns a clone of the default template.
func (etm *EmbeddedTemplateManager) GetDefaultTemplate() (*opc.Package, error) {
	return etm.GetTemplate(TemplateDefault)
}

// HasTemplate reports whether the named template exists.
func (etm *EmbeddedTemplateManager) HasTemplate(name TemplateType) bool {
	_, ok := etm.templates.Load(name)
	return ok
}

// ============================================================================
// Package-level convenience functions
// ============================================================================

// GetEmbeddedTemplate returns a clone of the named embedded template.
func GetEmbeddedTemplate(name TemplateType) (*opc.Package, error) {
	return globalEmbeddedTemplates.GetTemplate(name)
}

// GetEmbeddedDefaultTemplate returns a clone of the default embedded template.
func GetEmbeddedDefaultTemplate() (*opc.Package, error) {
	return globalEmbeddedTemplates.GetDefaultTemplate()
}

// InitEmbeddedTemplates initializes all embedded templates.
func InitEmbeddedTemplates() error {
	return globalEmbeddedTemplates.Init()
}
