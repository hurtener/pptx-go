// Package frames is the scene-side frame registry: it wires the curated device
// frames (assets/frames) to their reserved names and provides the per-render
// caller-extension overlay (RFC §14.4, D-038).
//
// The registry is the project's first curated-asset extension seam — the
// interface + factory + driver shape (CLAUDE.md §4.4) that icons (Phase 12)
// and ornaments (Phase 13) repeat: a closed curated name set plus caller
// extension. Curated() is the factory; a Recipe is a driver; With overlays a
// caller driver for one render. The registry is built once per Render and is
// read-only during the parallel compose, so it is concurrency-safe and keeps
// output byte-identical (D-035).
package frames

import (
	"sort"

	assetframes "github.com/hurtener/pptx-go/assets/frames"
	"github.com/hurtener/pptx-go/pptx"
)

// Recipe draws a device frame's bezel into region as native shapes and returns
// the interior box an image is placed into, plus the number of bezel shapes
// emitted. It composes the public pptx builder only (P1). The signature matches
// the curated recipes in assets/frames exactly.
type Recipe func(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int)

// The reserved curated frame names (RFC §14.3). The Image IR's FrameKind enum
// maps onto these (D-038).
const (
	NameBrowser = "browser"
	NamePhone   = "phone"
	NameDesktop = "desktop"
	NameLaptop  = "laptop"
)

// Registry is an immutable, name-keyed set of frame recipes. Build the curated
// set with Curated and overlay caller frames with With; both Lookup and Names
// are safe on a nil *Registry (treated as empty).
type Registry struct {
	m map[string]Recipe
}

// Curated returns a registry seeded with the four curated frames (RFC §14.3).
func Curated() *Registry {
	return &Registry{m: map[string]Recipe{
		NameBrowser: assetframes.Browser,
		NamePhone:   assetframes.Phone,
		NameDesktop: assetframes.Desktop,
		NameLaptop:  assetframes.Laptop,
	}}
}

// With returns a copy of the registry with name bound to rec (overriding any
// existing entry, curated or not). The receiver is not mutated — extensions are
// per-render, never global (RFC §14.4). A blank name or nil recipe is ignored.
func (r *Registry) With(name string, rec Recipe) *Registry {
	size := 0
	if r != nil {
		size = len(r.m)
	}
	cp := &Registry{m: make(map[string]Recipe, size+1)}
	if r != nil {
		for k, v := range r.m {
			cp.m[k] = v
		}
	}
	if name != "" && rec != nil {
		cp.m[name] = rec
	}
	return cp
}

// Lookup returns the recipe registered under name, or (nil, false).
func (r *Registry) Lookup(name string) (Recipe, bool) {
	if r == nil {
		return nil, false
	}
	rec, ok := r.m[name]
	return rec, ok
}

// Names returns the registered frame names in sorted order (used in validation
// messages).
func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}
	out := make([]string, 0, len(r.m))
	for k := range r.m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
