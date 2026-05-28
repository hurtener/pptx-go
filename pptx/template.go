// Package pptx provides a high-level API for authoring PPTX files,
// and serves as the primary entry point for both human developers and AI callers.
package pptx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/hurtener/pptx-go/opc"
)

// ============================================================================
// Lazy-loading template system
// ============================================================================
//
// Design principles:
// 1. Deferred loading — templates are loaded only on first use.
// 2. Zero-copy optimization — immutable resources (images, masters, layouts)
//    use zero-copy sharing.
// 3. Thread-safe — sync.Map provides concurrent-safe storage.
// 4. Memory-efficient — each template is loaded once; callers receive Clone()
//    copies for independent modification.
// ============================================================================

// TemplateType identifies a named presentation template.
type TemplateType string

const (
	// TemplateBlank is an empty/blank template.
	TemplateBlank TemplateType = "blank.pptx"
	// TemplateDefault is the default 16:9 widescreen template.
	TemplateDefault TemplateType = "default.pptx"
	// TemplateWide is a widescreen template.
	TemplateWide TemplateType = "wide.pptx"
	// TemplateStandard is a standard 4:3 template.
	TemplateStandard TemplateType = "standard.pptx"
)

// TemplateManager handles lazy loading, caching, and cloning of templates.
type TemplateManager struct {
	// template cache: template name -> *opc.Package
	templates sync.Map

	// default template to use when none is specified
	defaultTemplate TemplateType

	// cached master/layout data
	masterCache *MasterCache

	// optional directory to search for template files
	templateDir string
}

// global default TemplateManager instance
var globalTemplateManager = NewTemplateManager()

// NewTemplateManager creates a new TemplateManager.
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		defaultTemplate: TemplateDefault,
	}
}

// NewTemplateManagerWithDir creates a TemplateManager that searches dir for
// template files.
func NewTemplateManagerWithDir(dir string) *TemplateManager {
	return &TemplateManager{
		defaultTemplate: TemplateDefault,
		templateDir:     dir,
	}
}

// ============================================================================
// Template loading
// ============================================================================

// LoadTemplate returns a clone of the named template, loading and caching it
// from the file system on first use.
func (tm *TemplateManager) LoadTemplate(name TemplateType) (*opc.Package, error) {
	// check cache
	if cached, ok := tm.templates.Load(name); ok {
		pkg := cached.(*opc.Package)
		// return a clone to ensure thread safety
		return pkg.Clone(), nil
	}

	// try to read from the template directory
	data, err := tm.readTemplateFile(string(name))
	if err != nil {
		return nil, fmt.Errorf("loading template %s: %w", name, err)
	}

	// parse into an OPC package
	pkg, err := tm.parseTemplate(data)
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", name, err)
	}

	// cache the original; callers always receive a Clone
	tm.templates.Store(name, pkg)

	return pkg.Clone(), nil
}

// LoadDefault loads the default template.
func (tm *TemplateManager) LoadDefault() (*opc.Package, error) {
	return tm.LoadTemplate(tm.defaultTemplate)
}

// readTemplateFile reads a template file by searching in order:
// 1. The configured templateDir (if set).
// 2. A "templates/" subdirectory of the working directory.
// 3. A "templates/" subdirectory next to the executable.
func (tm *TemplateManager) readTemplateFile(name string) ([]byte, error) {
	searchPaths := []string{}

	if tm.templateDir != "" {
		searchPaths = append(searchPaths, filepath.Join(tm.templateDir, name))
	}

	// working directory
	searchPaths = append(searchPaths, filepath.Join("templates", name))

	// executable directory
	if exePath, err := os.Executable(); err == nil {
		searchPaths = append(searchPaths, filepath.Join(filepath.Dir(exePath), "templates", name))
	}

	for _, path := range searchPaths {
		if data, err := os.ReadFile(path); err == nil {
			return data, nil
		}
	}

	return nil, fmt.Errorf("template file %s not found in any search path", name)
}

// parseTemplate parses raw template bytes into an OPC package and
// populates the master cache if it has not yet been populated.
func (tm *TemplateManager) parseTemplate(data []byte) (*opc.Package, error) {
	// create a ZIP reader
	reader := bytes.NewReader(data)
	zipReader, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("parsing ZIP: %w", err)
	}

	// parse the OPC package
	pkg, err := opc.Open(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("parsing OPC package: %w", err)
	}

	// lazily initialize the master cache
	if tm.masterCache == nil {
		masterMgr := NewMasterManager()
		if err := masterMgr.LoadFromZip(zipReader); err != nil {
			// master loading failure does not prevent the template from being used
		}
		tm.masterCache = masterMgr.Cache()
	}

	return pkg, nil
}

// ============================================================================
// Template registration
// ============================================================================

// RegisterTemplate registers the template at the given file path under name.
func (tm *TemplateManager) RegisterTemplate(name TemplateType, path string) error {
	pkg, err := opc.OpenFile(path)
	if err != nil {
		return fmt.Errorf("opening template file: %w", err)
	}

	tm.templates.Store(name, pkg)
	return nil
}

// RegisterTemplateFromBytes parses and registers a template from raw bytes.
func (tm *TemplateManager) RegisterTemplateFromBytes(name TemplateType, data []byte) error {
	reader := bytes.NewReader(data)
	pkg, err := opc.Open(reader, int64(len(data)))
	if err != nil {
		return fmt.Errorf("parsing template data: %w", err)
	}

	tm.templates.Store(name, pkg)
	return nil
}

// RegisterTemplateFromFS registers a template read from the given fs.FS.
func (tm *TemplateManager) RegisterTemplateFromFS(fsys fs.FS, name TemplateType, path string) error {
	file, err := fsys.Open(path)
	if err != nil {
		return fmt.Errorf("opening template file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("reading template file: %w", err)
	}

	return tm.RegisterTemplateFromBytes(name, data)
}

// ============================================================================
// Helper methods
// ============================================================================

// SetDefaultTemplate sets the template returned by LoadDefault.
func (tm *TemplateManager) SetDefaultTemplate(name TemplateType) {
	tm.defaultTemplate = name
}

// SetTemplateDir sets the directory searched first when loading templates.
func (tm *TemplateManager) SetTemplateDir(dir string) {
	tm.templateDir = dir
}

// GetMasterCache returns the cached master/layout data.
func (tm *TemplateManager) GetMasterCache() *MasterCache {
	return tm.masterCache
}

// HasTemplate reports whether the named template is already cached.
func (tm *TemplateManager) HasTemplate(name TemplateType) bool {
	_, ok := tm.templates.Load(name)
	return ok
}

// ClearCache evicts all cached templates and resets the master cache.
func (tm *TemplateManager) ClearCache() {
	tm.templates = sync.Map{}
	tm.masterCache = nil
}

// ============================================================================
// Package-level convenience functions (use the global manager)
// ============================================================================

// LoadTemplate loads the named template using the global manager.
func LoadTemplate(name TemplateType) (*opc.Package, error) {
	return globalTemplateManager.LoadTemplate(name)
}

// LoadDefaultTemplate loads the default template using the global manager.
func LoadDefaultTemplate() (*opc.Package, error) {
	return globalTemplateManager.LoadDefault()
}

// RegisterTemplate registers a template by file path using the global manager.
func RegisterTemplate(name TemplateType, path string) error {
	return globalTemplateManager.RegisterTemplate(name, path)
}

// RegisterTemplateFromBytes registers a template from raw bytes using the
// global manager.
func RegisterTemplateFromBytes(name TemplateType, data []byte) error {
	return globalTemplateManager.RegisterTemplateFromBytes(name, data)
}

// ============================================================================
// TemplateBuilder — build templates programmatically
// ============================================================================

// TemplateBuilder builds a PPTX template from scratch.
type TemplateBuilder struct {
	pkg *opc.Package
}

// NewTemplateBuilder creates a new TemplateBuilder.
func NewTemplateBuilder() *TemplateBuilder {
	return &TemplateBuilder{
		pkg: opc.NewPackage(),
	}
}

// Package returns the underlying OPC package.
func (tb *TemplateBuilder) Package() *opc.Package {
	return tb.pkg
}

// Build returns the assembled OPC package.
func (tb *TemplateBuilder) Build() *opc.Package {
	return tb.pkg
}

// BuildAndRegister assembles the template and registers it with the global
// manager under name.
func (tb *TemplateBuilder) BuildAndRegister(name TemplateType) error {
	globalTemplateManager.templates.Store(name, tb.pkg)
	return nil
}
