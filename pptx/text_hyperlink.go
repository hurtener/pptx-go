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
	return &Run{tf: p.tf, run: run}
}

// Hyperlink returns the run's link target (the external URL) and true, or "" and
// false when the run is not a hyperlink. It is the read inverse of
// AddHyperlink: the run's <a:hlinkClick r:id> is resolved through the slide's
// relationships to the verbatim URL the caller supplied (§7 — pptx-go does not
// validate it).
func (r *Run) Hyperlink() (string, bool) {
	pr := r.run.TextProperties
	if pr == nil || pr.HlinkClick == nil || pr.HlinkClick.RID == "" {
		return "", false
	}
	if r.tf == nil || r.tf.s == nil {
		return "", false
	}
	rel := r.tf.s.part.Relationships().Get(pr.HlinkClick.RID)
	if rel == nil {
		return "", false
	}
	return rel.TargetRef(), true
}
