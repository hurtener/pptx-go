package pptx

import "sort"

// Read warnings (RFC §16 best-effort clause; D-048). Opening a deck pptx-go did
// not author is best-effort: unrecognized content is ignored at parse time
// rather than preserved, but every degradation is reported here so the caller
// knows what was lost. A self-authored deck round-trips losslessly (D-047) and
// reports no warnings.

// ReadWarningKind classifies a non-fatal issue encountered while reading a deck.
type ReadWarningKind int

const (
	// WarnDroppedElement reports that an unrecognized element was ignored at
	// parse time (e.g. a group shape or mc:AlternateContent in a slide's shape
	// tree). The element is named in ReadWarning.Element.
	WarnDroppedElement ReadWarningKind = iota
	// WarnUnreadablePart reports that a referenced part was missing or could not
	// be parsed, and was skipped rather than failing the open.
	WarnUnreadablePart
)

// String renders the kind for logs.
func (k ReadWarningKind) String() string {
	switch k {
	case WarnDroppedElement:
		return "dropped-element"
	case WarnUnreadablePart:
		return "unreadable-part"
	default:
		return "unknown"
	}
}

// ReadWarning is one non-fatal degradation noted while reading a (third-party)
// deck. It carries enough to locate the issue without surfacing any raw OOXML
// (P3): the part URI, the element local-name (for WarnDroppedElement), and a
// human-readable detail.
type ReadWarning struct {
	Kind    ReadWarningKind
	Part    string // the part URI the warning relates to, e.g. "/ppt/slides/slide2.xml"
	Element string // element local-name (WarnDroppedElement); empty otherwise
	Detail  string // human-readable context
}

// ReadWarnings returns the warnings collected when the deck was opened, in a
// stable order (by part, then element, then kind). It is nil for a deck pptx-go
// authored, which round-trips losslessly (D-047).
func (p *Presentation) ReadWarnings() []ReadWarning {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if len(p.readWarnings) == 0 {
		return nil
	}
	out := make([]ReadWarning, len(p.readWarnings))
	copy(out, p.readWarnings)
	return out
}

// addReadWarning appends w, de-duplicated per (Kind, Part, Element) so a deck
// with many instances of the same unrecognized element on a part yields one
// warning (R1). When a logger is injected (WithLogger), each distinct
// degradation is also emitted as a Warn event so it is visible to logs, not just
// ReadWarnings (CLAUDE.md §8). Called single-threaded during open, before the
// presentation is handed to the caller, so it does not take p.mu.
func (p *Presentation) addReadWarning(w ReadWarning) {
	for _, e := range p.readWarnings {
		if e.Kind == w.Kind && e.Part == w.Part && e.Element == w.Element {
			return
		}
	}
	p.readWarnings = append(p.readWarnings, w)
	if p.logger != nil {
		p.logger.Warn("pptx: read degradation",
			"kind", w.Kind.String(),
			"part", w.Part,
			"element", w.Element,
			"detail", w.Detail,
		)
	}
}

// sortReadWarnings orders the collected warnings deterministically (by part,
// then element, then kind), matching ReadWarnings' documented order.
func (p *Presentation) sortReadWarnings() {
	sort.SliceStable(p.readWarnings, func(i, j int) bool {
		a, b := p.readWarnings[i], p.readWarnings[j]
		if a.Part != b.Part {
			return a.Part < b.Part
		}
		if a.Element != b.Element {
			return a.Element < b.Element
		}
		return a.Kind < b.Kind
	})
}
