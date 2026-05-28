package opc

import (
	"sync"
)

// ResourcePool is a global pool that manages shareable static resources.
// It uses a zero-copy strategy so multiple Packages can share the same binary data.
type ResourcePool struct {
	mu      sync.RWMutex
	themes  map[string][]byte // theme URI -> blob
	masters map[string][]byte // master URI -> blob
	layouts map[string][]byte // layout URI -> blob
	media   map[string][]byte // media URI -> blob (images, audio, video, etc.)
	fonts   map[string][]byte // font URI -> blob

	// Reference counts for tracking resource usage.
	refCount map[string]int
}

// globalPool is the singleton global resource pool.
var globalPool = &ResourcePool{
	themes:   make(map[string][]byte),
	masters:  make(map[string][]byte),
	layouts:  make(map[string][]byte),
	media:    make(map[string][]byte),
	fonts:    make(map[string][]byte),
	refCount: make(map[string]int),
}

// GetGlobalPool returns the global resource pool.
func GetGlobalPool() *ResourcePool {
	return globalPool
}

// GetOrLoad returns the cached resource for the given URI, loading it via loader if absent.
// The loader is called at most once per unique URI.
func (p *ResourcePool) GetOrLoad(uri string, contentType string, loader func() ([]byte, error)) ([]byte, error) {
	// Fast path: try the read lock first.
	p.mu.RLock()
	if data, ok := p.getResource(uri, contentType); ok {
		p.mu.RUnlock()
		p.incrementRef(uri)
		return data, nil
	}
	p.mu.RUnlock()

	// Slow path: acquire write lock.
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring the write lock.
	if data, ok := p.getResource(uri, contentType); ok {
		p.incrementRefLocked(uri)
		return data, nil
	}

	// Load the resource.
	data, err := loader()
	if err != nil {
		return nil, err
	}

	// Store in the appropriate sub-pool.
	p.storeResource(uri, contentType, data)
	p.incrementRefLocked(uri)

	return data, nil
}

// getResource retrieves a resource from the pool (caller must hold the lock).
func (p *ResourcePool) getResource(uri string, contentType string) ([]byte, bool) {
	var data []byte
	var ok bool

	switch {
	case contentType == ContentTypeTheme || contentType == ContentTypeThemeOverride:
		data, ok = p.themes[uri]
	case contentType == ContentTypeSlideMaster:
		data, ok = p.masters[uri]
	case contentType == ContentTypeSlideLayout:
		data, ok = p.layouts[uri]
	case contentType == ContentTypeFont:
		data, ok = p.fonts[uri]
	case IsLargeBinaryContentType(contentType):
		data, ok = p.media[uri]
	default:
		return nil, false
	}

	return data, ok
}

// storeResource stores a resource in the pool (caller must hold the write lock).
func (p *ResourcePool) storeResource(uri string, contentType string, data []byte) {
	switch {
	case contentType == ContentTypeTheme || contentType == ContentTypeThemeOverride:
		p.themes[uri] = data
	case contentType == ContentTypeSlideMaster:
		p.masters[uri] = data
	case contentType == ContentTypeSlideLayout:
		p.layouts[uri] = data
	case contentType == ContentTypeFont:
		p.fonts[uri] = data
	case IsLargeBinaryContentType(contentType):
		p.media[uri] = data
	}
}

// incrementRef increments the reference count (caller must hold the read lock).
func (p *ResourcePool) incrementRef(uri string) {
	p.mu.Lock()
	p.refCount[uri]++
	p.mu.Unlock()
}

// incrementRefLocked increments the reference count (caller must already hold the write lock).
func (p *ResourcePool) incrementRefLocked(uri string) {
	p.refCount[uri]++
}

// Release decrements the reference count for the given URI.
// When the count reaches zero the resource is removed from the pool.
func (p *ResourcePool) Release(uri string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if count, ok := p.refCount[uri]; ok {
		if count <= 1 {
			// Reference count is zero — remove the resource.
			delete(p.refCount, uri)
			delete(p.themes, uri)
			delete(p.masters, uri)
			delete(p.layouts, uri)
			delete(p.media, uri)
			delete(p.fonts, uri)
		} else {
			p.refCount[uri] = count - 1
		}
	}
}

// ReleaseAll removes all resources from the pool. Use with care.
func (p *ResourcePool) ReleaseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.themes = make(map[string][]byte)
	p.masters = make(map[string][]byte)
	p.layouts = make(map[string][]byte)
	p.media = make(map[string][]byte)
	p.fonts = make(map[string][]byte)
	p.refCount = make(map[string]int)
}

// Stats returns resource pool statistics.
func (p *ResourcePool) Stats() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]int{
		"themes":  len(p.themes),
		"masters": len(p.masters),
		"layouts": len(p.layouts),
		"media":   len(p.media),
		"fonts":   len(p.fonts),
		"total":   len(p.themes) + len(p.masters) + len(p.layouts) + len(p.media) + len(p.fonts),
	}
}

// Prefetch pre-loads a set of resources into the pool to avoid load latency at render time.
func (p *ResourcePool) Prefetch(resources map[string]func() ([]byte, error)) error {
	for uri, loader := range resources {
		if _, err := p.GetOrLoad(uri, "", loader); err != nil {
			return err
		}
	}
	return nil
}

// CreateSharedPart creates a shared Part backed by the pool.
// If the resource is not already in the pool it is loaded via loader.
func (p *ResourcePool) CreateSharedPart(uri *PackURI, contentType string, loader func() ([]byte, error)) (*Part, error) {
	data, err := p.GetOrLoad(uri.URI(), contentType, loader)
	if err != nil {
		return nil, err
	}
	return NewSharedPart(uri, contentType, data), nil
}
