package pptx

import (
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// Hyperlink runs (RFC §8.4). A hyperlinked run carries an <a:hlinkClick r:id>
// pointing at an external relationship whose target is the URL. The URL is
// emitted verbatim — pptx-go does not fetch or validate it (§7); callers
// sanitize.

// AddHyperlink appends a run whose text links to target (an external URL) and
// returns it. The run is styled by style like any other run.
func (p *Paragraph) AddHyperlink(text, target string, style RunStyle) *Run {
	props := style.toProps(p.tf.s.activeTheme())
	if props == nil {
		props = &slide.XTextProperties{}
	}

	// External relationship on the slide; mirrored to the package part by
	// syncSlides (the Phase 03 C relationship seam). External rels are exempt
	// from the dangling-target conformance check.
	rid := ""
	if rel, err := p.tf.s.part.Relationships().AddNew(opc.RelTypeHyperlink, target, true); err == nil {
		rid = rel.RID()
	}
	props.HlinkClick = &slide.XHlinkClick{RID: rid}

	run := &slide.XTextRun{Text: text, TextProperties: props}
	xp := p.x()
	xp.Content = append(xp.Content, run)
	return &Run{run: run}
}
