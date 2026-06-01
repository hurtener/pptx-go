// Package icons embeds the curated icon set (RFC §14.1, D-005): lucide-style
// glyphs authored as single-path, solid-fill SVGs. They are rendered as native
// PPTX custom-geometry shapes by the SVG→OOXML translator; every embedded icon
// is validated by icons_test.go to satisfy the translator's constraints.
//
// This is the starter set; it grows toward the ≈60-icon V1.0.0 target as a
// content follow-up (D-040). Each addition is one validated .svg file — no code
// change.
package icons

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed *.svg
var fsys embed.FS

// Read returns the embedded SVG bytes for the named icon (no ".svg" suffix), or
// (nil, false) if no such icon is curated.
func Read(name string) ([]byte, bool) {
	b, err := fsys.ReadFile(name + ".svg")
	if err != nil {
		return nil, false
	}
	return b, true
}

// Names returns the curated icon names (without the ".svg" suffix), sorted.
func Names() []string {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, strings.TrimSuffix(e.Name(), ".svg"))
	}
	sort.Strings(out)
	return out
}
