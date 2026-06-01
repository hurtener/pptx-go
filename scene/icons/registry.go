// Package icons is the scene-side icon registry: it wires the curated icon set
// (assets/icons) to their names and provides the per-render caller-extension
// overlay (RFC §14.1/§14.4, D-005, D-040).
//
// It mirrors scene/frames (the Phase-10 curated-asset seam): a closed curated
// name set plus caller extension, an immutable per-render overlay, read-only
// during compose. The registry value is the icon's SVG bytes; the bytes are
// translated to native custom geometry by the builder (pptx.AddIcon) when an
// icon is placed — icon placement itself is wired by the consuming nodes (card,
// flow) in later phases.
package icons

import (
	"sort"

	asseticons "github.com/hurtener/pptx-go/assets/icons"
)

// Registry is an immutable, name-keyed set of icon SVGs. Lookup and Names are
// safe on a nil *Registry (treated as empty).
type Registry struct {
	m map[string][]byte
}

// Curated returns a registry seeded with the embedded curated icon set.
func Curated() *Registry {
	names := asseticons.Names()
	m := make(map[string][]byte, len(names))
	for _, n := range names {
		if b, ok := asseticons.Read(n); ok {
			m[n] = b
		}
	}
	return &Registry{m: m}
}

// With returns a copy of the registry with name bound to svg (overriding any
// existing entry). The receiver is not mutated — extensions are per-render, not
// global. A blank name or nil svg is ignored.
func (r *Registry) With(name string, svg []byte) *Registry {
	size := 0
	if r != nil {
		size = len(r.m)
	}
	cp := &Registry{m: make(map[string][]byte, size+1)}
	if r != nil {
		for k, v := range r.m {
			cp.m[k] = v
		}
	}
	if name != "" && svg != nil {
		cp.m[name] = svg
	}
	return cp
}

// Lookup returns the SVG bytes registered under name, or (nil, false).
func (r *Registry) Lookup(name string) ([]byte, bool) {
	if r == nil {
		return nil, false
	}
	b, ok := r.m[name]
	return b, ok
}

// Names returns the registered icon names in sorted order.
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
