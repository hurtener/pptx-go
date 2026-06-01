package pptx

import (
	"sort"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// Master / Layout — read-only template views (RFC §13.2)
// ============================================================================
//
// When a deck is opened (or a presentation is seeded from a brand kit via
// FromTemplate), pptx-go builds a read-only registry of its slide masters and
// the layouts they own. Master/Layout are the public, OOXML-free (P3) view of
// that registry: they expose names and the master→layout grouping so a caller
// (or the scene renderer's LayoutMap) can pick a layout by name. The XML wire
// types stay in internal/ooxml.

// Layout is a read-only view of one slide layout in a template.
type Layout struct {
	name string                // cSld@name (the PowerPoint layout-picker name)
	uri  string                // package part URI, e.g. /ppt/slideLayouts/slideLayout1.xml
	typ  slide.SlideLayoutType // sldLayout@type, mapped to the internal enum
}

// Name returns the layout's display name (may be empty if the template omits it).
func (l *Layout) Name() string { return l.name }

// Master is a read-only view of one slide master and the layouts it owns.
type Master struct {
	name    string
	uri     string
	layouts []*Layout
}

// Name returns the master's name (may be empty).
func (m *Master) Name() string { return m.name }

// Layouts returns the layouts owned by this master, in document order.
func (m *Master) Layouts() []*Layout {
	out := make([]*Layout, len(m.layouts))
	copy(out, m.layouts)
	return out
}

// Layout returns the master's layout with the given name, if present.
func (m *Master) Layout(name string) (*Layout, bool) {
	for _, l := range m.layouts {
		if l.name == name {
			return l, true
		}
	}
	return nil, false
}

// Masters returns the presentation's slide masters and their layouts. It is
// non-nil only for a deck opened from a file/stream or seeded with
// FromTemplate; a blank New() deck returns nil (its scaffold master is not
// surfaced). The returned slice is a copy; the masters themselves are read-only.
func (p *Presentation) Masters() []*Master {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]*Master, len(p.masters))
	copy(out, p.masters)
	return out
}

// layoutURIByName resolves a layout name to its package part URI across every
// master, for targeting a slide→layout relationship. Caller holds p.mu (or is a
// constructor before the value is shared).
func (p *Presentation) layoutURIByName(name string) (string, bool) {
	if name == "" {
		return "", false
	}
	for _, m := range p.masters {
		if l, ok := m.Layout(name); ok && l.uri != "" {
			return l.uri, true
		}
	}
	return "", false
}

// HasLayout reports whether a layout with the given name exists in the
// presentation's master/layout registry (e.g. a layout an ingested template
// defines). Callers — including the scene renderer's LayoutMap — use it to
// decide whether a layout selection will resolve before adding a slide.
func (p *Presentation) HasLayout(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, ok := p.layoutURIByName(name)
	return ok
}

// resolveLayoutURI maps an optional layout name to the absolute pack URI a slide
// should relate to: a named layout from the registry when it resolves, else the
// default blank layout. Caller holds p.mu.
func (p *Presentation) resolveLayoutURI(layout ...string) string {
	if len(layout) > 0 && layout[0] != "" {
		if uri, ok := p.layoutURIByName(layout[0]); ok {
			return uri
		}
	}
	return defaultSlideLayoutURI
}

// buildMasterRegistry reads the package's slide masters and layouts into the
// read-only Master/Layout registry. Layouts are grouped under the master that
// references them (via the master's slideLayout relationships); any layout not
// referenced by a master — and every layout when the deck has no master rels —
// falls under the first master so it is still reachable by name. The parse is
// permissive (brief 01 F6): a master/layout that fails to parse contributes a
// nameless entry rather than failing the open. Caller holds p.mu (or is a
// constructor before the value is shared).
func buildMasterRegistry(pkg *opc.Package) []*Master {
	// Index every layout by its part URI.
	layoutByURI := make(map[string]*Layout)
	var layoutURIs []string
	for _, lp := range pkg.GetPartsByType(opc.ContentTypeSlideLayout) {
		uri := lp.PartURI().URI()
		l := &Layout{uri: uri}
		if ld, err := slide.ParseLayout(lp.Blob()); err == nil {
			l.name = ld.Name()
			l.typ = ld.LayoutType()
		}
		layoutByURI[uri] = l
		layoutURIs = append(layoutURIs, uri)
	}
	sort.Strings(layoutURIs) // deterministic order for the fallback grouping

	masterParts := pkg.GetPartsByType(opc.ContentTypeSlideMaster)
	sort.Slice(masterParts, func(i, j int) bool {
		return masterParts[i].PartURI().URI() < masterParts[j].PartURI().URI()
	})

	var masters []*Master
	claimed := make(map[string]bool)
	for _, mp := range masterParts {
		m := &Master{uri: mp.PartURI().URI()}
		if md, err := slide.ParseMaster(mp.Blob()); err == nil {
			m.name = md.Name()
		}
		for _, rel := range mp.Relationships().GetByType(opc.RelTypeSlideLayout) {
			if rel.TargetURI() == nil {
				continue
			}
			turi := rel.TargetURI().URI()
			if l, ok := layoutByURI[turi]; ok && !claimed[turi] {
				m.layouts = append(m.layouts, l)
				claimed[turi] = true
			}
		}
		masters = append(masters, m)
	}

	if len(masters) == 0 && len(layoutURIs) > 0 {
		masters = append(masters, &Master{})
	}
	// Attach any unclaimed layouts (no master rels, or a layout the master
	// doesn't reference) to the first master so it stays reachable by name.
	if len(masters) > 0 {
		for _, uri := range layoutURIs {
			if !claimed[uri] {
				masters[0].layouts = append(masters[0].layouts, layoutByURI[uri])
			}
		}
	}
	return masters
}
