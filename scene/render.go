package scene

import (
	"context"
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// The render core: a deterministic top-level body layout, per-node dispatch,
// and the scene-RichText → builder-Paragraph mapping. Per-leaf composers live
// in render_leaves.go. The renderer calls only the public pptx builder (P1)
// and adds no product behavior (D-026): no modes, no legibility heuristics, no
// text-size boosting.

// renderer carries per-Render state.
type renderer struct {
	pres  *pptx.Presentation
	cfg   renderConfig
	theme *pptx.Theme
	ctx   context.Context
	stats Stats
}

// placement is a node assigned to a slot.
type placement struct {
	node SlideNode
	box  pptx.Box
}

// bodyMargin is the uniform inset from the slide edge to the body region.
var bodyMargin = pptx.In(0.5)

// renderSlide adds a slide, lays out its top-level nodes, and composes them.
func (r *renderer) renderSlide(sl *SceneSlide) {
	ps := r.pres.AddSlide()

	if len(sl.Notes) > 0 {
		nf := ps.SpeakerNotes()
		p := nf.AddParagraph(pptx.ParagraphOpts{})
		r.addRichText(ps, p, sl.Notes, pptx.TypeBody)
	}

	for _, pl := range r.layout(sl.Nodes, sl.ID) {
		r.renderNode(ps, pl.box, pl.node, sl.ID)
	}
}

// bodyRegion returns the margin-inset content region of the slide.
func (r *renderer) bodyRegion() pptx.Box {
	cx, cy := r.pres.SlideSize() // EMU
	return pptx.Box{
		X: bodyMargin,
		Y: bodyMargin,
		W: pptx.EMU(cx) - 2*bodyMargin,
		H: pptx.EMU(cy) - 2*bodyMargin,
	}
}

// layout assigns each top-level node a slot. section_divider takes the full
// slide; every other node stacks vertically in the body region (in IR order).
func (r *renderer) layout(nodes []SlideNode, slideID string) []placement {
	cx, cy := r.pres.SlideSize()
	var out []placement
	var stacked []SlideNode
	for _, n := range nodes {
		if _, ok := n.(SectionDivider); ok {
			out = append(out, placement{n, pptx.Box{X: 0, Y: 0, W: pptx.EMU(cx), H: pptx.EMU(cy)}})
			continue
		}
		stacked = append(stacked, n)
	}
	return append(out, r.stackIn(r.bodyRegion(), stacked, slideID)...)
}

// stackIn places nodes top-to-bottom within box (full box width, per-node
// preferred height, SpaceMD gap). Shared by the body layout and container
// columns. Overflow past the box records a LayoutWarning.
func (r *renderer) stackIn(box pptx.Box, nodes []SlideNode, slideID string) []placement {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	out := make([]placement, 0, len(nodes))
	y := box.Y
	for _, n := range nodes {
		h := preferredHeight(n)
		out = append(out, placement{n, pptx.Box{X: box.X, Y: y, W: box.W, H: h}})
		y += h + gap
	}
	if len(nodes) > 0 && y-gap > box.Bottom() {
		r.warn(slideID, "content overflows its region")
	}
	return out
}

// renderNode dispatches a node to its composer per the §12 policy.
func (r *renderer) renderNode(ps *pptx.Slide, box pptx.Box, n SlideNode, slideID string) {
	switch v := n.(type) {
	case Hero:
		r.renderHero(ps, box, v)
	case Prose:
		r.renderProse(ps, box, v)
	case Heading:
		r.renderHeading(ps, box, v)
	case List:
		r.renderList(ps, box, v)
	case Divider:
		r.renderDivider(ps, box, v)
	case Quote:
		r.renderQuote(ps, box, v)
	case Callout:
		r.renderCallout(ps, box, v)
	case Chip:
		r.renderChip(ps, box, v)
	case Arrow:
		r.renderArrow(ps, box, v)
	case CodeBlock:
		r.renderCodeBlock(ps, box, v, slideID)
	case SectionDivider:
		r.renderSectionDivider(ps, box, v)
	case TwoColumn:
		r.renderTwoColumn(ps, box, v, slideID)
	case Grid:
		r.renderGrid(ps, box, v, slideID)
	case Table:
		r.renderTable(ps, box, v, slideID)
	default:
		// image/chart/decoration/table/flow + card/card_section are later phases.
		r.warn(slideID, fmt.Sprintf("%s rendering is not yet implemented; node skipped", n.NodeKind()))
	}
}

// addRichText maps a scene RichText onto a builder paragraph: one run per
// TextRun, carrying inline style + color, at the node's base type role. A link
// run becomes a hyperlink. Returns the number of resolver-dependent extras (0
// for text).
func (r *renderer) addRichText(_ *pptx.Slide, p *pptx.Paragraph, rt RichText, base pptx.TypeRole) {
	for _, run := range rt {
		style := pptx.RunStyle{
			TypeRole: base,
			Bold:     run.Style.Bold,
			Italic:   run.Style.Italic,
			Code:     run.Style.Code,
		}
		if run.Style.Underline {
			style.Underline = pptx.UnderlineSingle
		}
		if run.Style.Strike {
			style.Strike = pptx.StrikeSingle
		}
		style.Color = colorFor(run.Color)

		if run.Style.Link && run.Style.Href != "" {
			p.AddHyperlink(run.Text, run.Style.Href, style)
		} else {
			p.AddRun(run.Text, style)
		}
	}
}

// colorFor maps a scene TextColor to a builder Color (token or literal).
func colorFor(c TextColor) pptx.Color {
	if c.IsLiteral() {
		return c.Literal()
	}
	return pptx.TokenTextColor(c.Role())
}

// warn records a non-fatal layout warning.
func (r *renderer) warn(slideID, msg string) {
	r.stats.Warnings = append(r.stats.Warnings, LayoutWarning{SlideID: slideID, Message: msg})
}

// preferredHeight returns a node's preferred slot height (EMU). These are
// deterministic placement sizes, not content opinions (D-026).
func preferredHeight(n SlideNode) pptx.EMU {
	switch v := n.(type) {
	case Hero:
		return pptx.In(2.2)
	case Heading:
		return pptx.In(0.6)
	case Prose:
		lines := len(v.Paragraphs)
		if lines < 1 {
			lines = 1
		}
		return pptx.In(0.4) * pptx.EMU(lines)
	case List:
		items := len(v.Items)
		if items < 1 {
			items = 1
		}
		return pptx.In(0.32) * pptx.EMU(items)
	case Divider:
		return pptx.In(0.2)
	case Quote:
		return pptx.In(1.1)
	case Callout:
		return pptx.In(1.0)
	case Chip:
		return pptx.In(0.4)
	case Arrow:
		return pptx.In(0.6)
	case CodeBlock:
		return pptx.In(2.6)
	case Table:
		rows := len(v.Rows)
		if len(v.Headers) > 0 {
			rows++
		}
		h := pptx.In(0.4) * pptx.EMU(rows)
		if v.Caption != "" {
			h += pptx.In(0.4)
		}
		return h
	case TwoColumn:
		return maxEMU(nodesHeight(v.Left), nodesHeight(v.Right))
	case Grid:
		cols := v.Columns
		if cols < 1 {
			cols = 1
		}
		rows := (len(v.Cells) + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
		var maxCell pptx.EMU
		for _, c := range v.Cells {
			if h := preferredHeight(c); h > maxCell {
				maxCell = h
			}
		}
		return pptx.EMU(rows)*maxCell + estGap*pptx.EMU(rows-1)
	default:
		return pptx.In(1.0)
	}
}

// estGap is a fixed gap estimate used only for sizing nested containers (the
// actual gap at render time comes from the theme).
const estGap = pptx.EMU(137160) // ~0.15"

// nodesHeight estimates the stacked height of a node list (for container slot
// sizing).
func nodesHeight(nodes []SlideNode) pptx.EMU {
	var sum pptx.EMU
	for i, n := range nodes {
		if i > 0 {
			sum += estGap
		}
		sum += preferredHeight(n)
	}
	return sum
}

func maxEMU(a, b pptx.EMU) pptx.EMU {
	if a > b {
		return a
	}
	return b
}
