// Package pptx provides a high-level API for authoring PPTX files.
package pptx

import (
	"archive/zip"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// ============================================================================
// MasterCache - read-only cache for slide masters and layouts
// ============================================================================
//
// Design principles:
// 1. Write once, read everywhere — sync.Once ensures initialization runs exactly once.
// 2. Frozen after init: subsequent reads need no lock, giving optimal performance.
// 3. Designed for concurrent safety; suitable for high-concurrency / streaming scenarios.
// ============================================================================

// MasterCache is a read-only cache of slide masters and layouts.
// All fields are frozen after initialization and support lock-free concurrent reads.
type MasterCache struct {
	// initialization guard
	once sync.Once

	// read-only data (frozen after init)
	masters map[string]*slide.SlideMasterData // key: masterID
	layouts map[string]*slide.SlideLayoutData // key: layoutID

	// auxiliary indexes (built during init)
	layoutByName map[string]string // layoutName -> layoutID
	masterByName map[string]string // masterName -> masterID

	// placeholder index (built during init)
	// key format: "layoutID:phType" or "masterID:phType"
	placeholderIndex map[string]*slide.Placeholder
}

// NewMasterCache creates a new MasterCache instance.
func NewMasterCache() *MasterCache {
	return &MasterCache{
		masters:          make(map[string]*slide.SlideMasterData),
		layouts:          make(map[string]*slide.SlideLayoutData),
		layoutByName:     make(map[string]string),
		masterByName:     make(map[string]string),
		placeholderIndex: make(map[string]*slide.Placeholder),
	}
}

// ============================================================================
// Initialization (runs at most once)
// ============================================================================

// Init populates the cache with the provided data. Only the first call has
// any effect; subsequent calls are silently ignored.
func (c *MasterCache) Init(masters []*slide.SlideMasterData, layouts []*slide.SlideLayoutData) {
	c.once.Do(func() {
		c.buildIndex(masters, layouts)
	})
}

// InitFunc lazily initializes the cache using the provided factory function.
// The function is called only on the first access.
func (c *MasterCache) InitFunc(initFn func() ([]*slide.SlideMasterData, []*slide.SlideLayoutData)) {
	c.once.Do(func() {
		masters, layouts := initFn()
		c.buildIndex(masters, layouts)
	})
}

// buildIndex constructs all internal indexes. Called only during initialization.
func (c *MasterCache) buildIndex(masters []*slide.SlideMasterData, layouts []*slide.SlideLayoutData) {
	// index slide masters
	for _, master := range masters {
		if master == nil {
			continue
		}
		c.masters[master.ID()] = master
		if master.Name() != "" {
			c.masterByName[master.Name()] = master.ID()
		}

		// index master-level placeholders
		for phID, ph := range master.Placeholders() {
			key := master.ID() + ":" + ph.Type().String()
			c.placeholderIndex[key] = ph
			// also index by ID
			c.placeholderIndex[master.ID()+":"+phID] = ph
		}
	}

	// index slide layouts
	for _, layout := range layouts {
		if layout == nil {
			continue
		}
		c.layouts[layout.ID()] = layout
		if layout.Name() != "" {
			c.layoutByName[layout.Name()] = layout.ID()
		}

		// index layout-level placeholders
		for phID, ph := range layout.Placeholders() {
			key := layout.ID() + ":" + ph.Type().String()
			c.placeholderIndex[key] = ph
			// also index by ID
			c.placeholderIndex[layout.ID()+":"+phID] = ph
		}
	}
}

// ============================================================================
// Read accessors — lock-free, concurrency-safe after init
// ============================================================================

// GetMaster returns the slide master with the given ID.
func (c *MasterCache) GetMaster(masterID string) (*slide.SlideMasterData, bool) {
	m, ok := c.masters[masterID]
	return m, ok
}

// GetMasterByName returns the slide master with the given name.
func (c *MasterCache) GetMasterByName(name string) (*slide.SlideMasterData, bool) {
	if id, ok := c.masterByName[name]; ok {
		return c.GetMaster(id)
	}
	return nil, false
}

// GetLayout returns the slide layout with the given ID.
func (c *MasterCache) GetLayout(layoutID string) (*slide.SlideLayoutData, bool) {
	l, ok := c.layouts[layoutID]
	return l, ok
}

// GetLayoutByName returns the slide layout with the given name.
func (c *MasterCache) GetLayoutByName(name string) (*slide.SlideLayoutData, bool) {
	if id, ok := c.layoutByName[name]; ok {
		return c.GetLayout(id)
	}
	return nil, false
}

// GetPlaceholder returns the placeholder for the given layout ID and placeholder
// type. phType should be a value returned by PlaceholderType.String(), e.g.
// "title" or "body".
func (c *MasterCache) GetPlaceholder(layoutID, phType string) (*slide.Placeholder, bool) {
	key := layoutID + ":" + phType
	ph, ok := c.placeholderIndex[key]
	return ph, ok
}

// GetPlaceholderByID returns the placeholder for the given layout ID and
// placeholder ID.
func (c *MasterCache) GetPlaceholderByID(layoutID, placeholderID string) (*slide.Placeholder, bool) {
	key := layoutID + ":" + placeholderID
	ph, ok := c.placeholderIndex[key]
	return ph, ok
}

// GetMasterPlaceholder returns the placeholder for the given master ID and
// placeholder type.
func (c *MasterCache) GetMasterPlaceholder(masterID, phType string) (*slide.Placeholder, bool) {
	key := masterID + ":" + phType
	ph, ok := c.placeholderIndex[key]
	return ph, ok
}

// ============================================================================
// Bulk read accessors
// ============================================================================

// AllMasters returns all slide masters (read-only).
func (c *MasterCache) AllMasters() map[string]*slide.SlideMasterData {
	return c.masters
}

// AllLayouts returns all slide layouts (read-only).
func (c *MasterCache) AllLayouts() map[string]*slide.SlideLayoutData {
	return c.layouts
}

// MasterCount returns the number of slide masters.
func (c *MasterCache) MasterCount() int {
	return len(c.masters)
}

// LayoutCount returns the number of slide layouts.
func (c *MasterCache) LayoutCount() int {
	return len(c.layouts)
}

// ============================================================================
// Helper methods
// ============================================================================

// LayoutExists reports whether a layout with the given ID exists.
func (c *MasterCache) LayoutExists(layoutID string) bool {
	_, ok := c.layouts[layoutID]
	return ok
}

// MasterExists reports whether a master with the given ID exists.
func (c *MasterCache) MasterExists(masterID string) bool {
	_, ok := c.masters[masterID]
	return ok
}

// ListLayoutIDs returns the IDs of all slide layouts.
func (c *MasterCache) ListLayoutIDs() []string {
	ids := make([]string, 0, len(c.layouts))
	for id := range c.layouts {
		ids = append(ids, id)
	}
	return ids
}

// ListMasterIDs returns the IDs of all slide masters.
func (c *MasterCache) ListMasterIDs() []string {
	ids := make([]string, 0, len(c.masters))
	for id := range c.masters {
		ids = append(ids, id)
	}
	return ids
}

// ListLayoutNames returns the names of all slide layouts.
func (c *MasterCache) ListLayoutNames() []string {
	names := make([]string, 0, len(c.layoutByName))
	for name := range c.layoutByName {
		names = append(names, name)
	}
	return names
}

// ============================================================================
// MasterManager - facade for slide master / layout management
// ============================================================================
//
// Entry point for external API callers. Responsibilities:
// 1. Load slide masters and layouts from a ZIP file.
// 2. Parse XML and convert it to read-only data structures.
// 3. Populate MasterCache for high-concurrency reads.
// ============================================================================

// MasterManager manages slide masters and layouts.
type MasterManager struct {
	cache *MasterCache
}

// NewMasterManager creates a new MasterManager.
func NewMasterManager() *MasterManager {
	return &MasterManager{
		cache: NewMasterCache(),
	}
}

// NewMasterManagerWithCache creates a MasterManager backed by the given cache.
func NewMasterManagerWithCache(cache *MasterCache) *MasterManager {
	return &MasterManager{
		cache: cache,
	}
}

// Cache returns the internal read-only cache.
func (m *MasterManager) Cache() *MasterCache {
	return m.cache
}

// ============================================================================
// Loading from ZIP
// ============================================================================

// LoadFromZip loads slide masters and layouts from the given ZIP reader.
// It scans /ppt/slideMasters/ and /ppt/slideLayouts/ inside the ZIP.
func (m *MasterManager) LoadFromZip(zipReader *zip.Reader) error {
	var masters []*slide.SlideMasterData
	var layouts []*slide.SlideLayoutData

	// collect master and layout files
	masterFiles := m.collectFiles(zipReader, "ppt/slideMasters/", "slideMaster")
	layoutFiles := m.collectFiles(zipReader, "ppt/slideLayouts/", "slideLayout")

	// sort by filename to ensure deterministic ordering
	sort.Slice(masterFiles, func(i, j int) bool {
		return masterFiles[i].name < masterFiles[j].name
	})
	sort.Slice(layoutFiles, func(i, j int) bool {
		return layoutFiles[i].name < layoutFiles[j].name
	})

	// parse slide masters
	for _, f := range masterFiles {
		data, err := m.readFile(f.file)
		if err != nil {
			return fmt.Errorf("reading master file %s: %w", f.name, err)
		}

		master, err := slide.ParseMaster(data)
		if err != nil {
			return fmt.Errorf("parsing master %s: %w", f.name, err)
		}

		masters = append(masters, master)
	}

	// parse slide layouts
	for _, f := range layoutFiles {
		data, err := m.readFile(f.file)
		if err != nil {
			return fmt.Errorf("reading layout file %s: %w", f.name, err)
		}

		layout, err := slide.ParseLayout(data)
		if err != nil {
			return fmt.Errorf("parsing layout %s: %w", f.name, err)
		}

		layouts = append(layouts, layout)
	}

	// initialize cache (runs at most once)
	m.cache.Init(masters, layouts)

	return nil
}

// LoadFromZipFile loads slide masters and layouts from the ZIP file at filePath.
func (m *MasterManager) LoadFromZipFile(filePath string) error {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return fmt.Errorf("opening ZIP file: %w", err)
	}
	defer reader.Close()

	return m.LoadFromZip(&reader.Reader)
}

// ============================================================================
// File collection helpers
// ============================================================================

type zipFileEntry struct {
	name string
	file *zip.File
}

// collectFiles returns all XML files under dir whose basename begins with prefix.
func (m *MasterManager) collectFiles(zipReader *zip.Reader, dir, prefix string) []zipFileEntry {
	var files []zipFileEntry

	for _, f := range zipReader.File {
		// check that the file lives in the target directory
		fileDir := path.Dir(f.Name)
		if fileDir != dir && !strings.HasPrefix(fileDir, dir) {
			continue
		}

		// check filename prefix and extension
		fileName := path.Base(f.Name)
		if !strings.HasPrefix(fileName, prefix) {
			continue
		}
		if path.Ext(fileName) != ".xml" {
			continue
		}

		files = append(files, zipFileEntry{
			name: f.Name,
			file: f,
		})
	}

	return files
}

// readFile reads and returns the full contents of a ZIP entry.
func (m *MasterManager) readFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// ============================================================================
// Convenience accessors (delegated to Cache)
// ============================================================================

// GetLayout returns the layout with the given ID.
func (m *MasterManager) GetLayout(layoutID string) (*slide.SlideLayoutData, bool) {
	return m.cache.GetLayout(layoutID)
}

// GetLayoutByName returns the layout with the given name.
func (m *MasterManager) GetLayoutByName(name string) (*slide.SlideLayoutData, bool) {
	return m.cache.GetLayoutByName(name)
}

// GetMaster returns the master with the given ID.
func (m *MasterManager) GetMaster(masterID string) (*slide.SlideMasterData, bool) {
	return m.cache.GetMaster(masterID)
}

// GetMasterByName returns the master with the given name.
func (m *MasterManager) GetMasterByName(name string) (*slide.SlideMasterData, bool) {
	return m.cache.GetMasterByName(name)
}

// GetPlaceholder returns the placeholder for the given layout ID and type.
func (m *MasterManager) GetPlaceholder(layoutID, phType string) (*slide.Placeholder, bool) {
	return m.cache.GetPlaceholder(layoutID, phType)
}

// AllLayouts returns all slide layouts.
func (m *MasterManager) AllLayouts() map[string]*slide.SlideLayoutData {
	return m.cache.AllLayouts()
}

// AllMasters returns all slide masters.
func (m *MasterManager) AllMasters() map[string]*slide.SlideMasterData {
	return m.cache.AllMasters()
}

// LayoutCount returns the number of slide layouts.
func (m *MasterManager) LayoutCount() int {
	return m.cache.LayoutCount()
}

// MasterCount returns the number of slide masters.
func (m *MasterManager) MasterCount() int {
	return m.cache.MasterCount()
}

// ListLayoutIDs returns the IDs of all slide layouts.
func (m *MasterManager) ListLayoutIDs() []string {
	return m.cache.ListLayoutIDs()
}

// ListLayoutNames returns the names of all slide layouts.
func (m *MasterManager) ListLayoutNames() []string {
	return m.cache.ListLayoutNames()
}
