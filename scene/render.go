package scene

import (
	"context"
	"fmt"
	"time"

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

	// Chrome state (Phase 24): deck-level chrome config copied onto each per-slide
	// renderer, the resolved page-number total, and this slide's scene index (for
	// the default 1-based page number). Inert when chrome.Enabled is false.
	chrome      Chrome
	chromeTotal int
	slideIndex  int
}

// placement is a node assigned to a slot.
type placement struct {
	node   SlideNode
	box    pptx.Box
	hAlign HAlign // effective horizontal alignment for text leaf nodes (zero = HAlignLeft)
}

// slideResult is one slide's composition outcome, merged into Stats in scene
// order so the aggregate stays deterministic regardless of worker scheduling.
type slideResult struct {
	shapes   int
	assets   int
	warnings []LayoutWarning
	dur      time.Duration
	colors   SlideColors
}

// bodyMargin is the uniform inset from the slide edge to the body region.
var bodyMargin = pptx.In(0.5)

// composeOne composes one already-created slide using a fresh per-slide renderer
// (its own Stats), so concurrent slides never share mutable render state. The
// returned slideResult is merged by the caller in scene order. ps and sl must
// belong to the same scene index.
func (base *renderer) composeOne(ps *pptx.Slide, sl *SceneSlide, idx int) slideResult {
	sr := &renderer{pres: base.pres, cfg: base.cfg, theme: base.theme, ctx: base.ctx,
		chrome: base.chrome, chromeTotal: base.chromeTotal, slideIndex: idx}
	start := time.Now()
	sr.composeSlide(ps, sl)
	// Capture the colors the slide actually rendered with. composeSlide leaves
	// sr.theme as the derived dark theme for VariantDark (and the active theme
	// otherwise), so these RGBs are exactly what the codec emitted with — the
	// dark palette included (D-058). No contrast logic here (D-026).
	return slideResult{
		shapes:   sr.stats.Shapes,
		assets:   sr.stats.Assets,
		warnings: sr.stats.Warnings,
		dur:      time.Since(start),
		colors: SlideColors{
			SlideID:     sl.ID,
			Canvas:      sr.theme.ResolveColor(pptx.ColorCanvas),
			Surface:     sr.theme.ResolveColor(pptx.ColorSurface),
			PrimaryText: sr.theme.ResolveTextColor(pptx.TextPrimary),
		},
	}
}

// composeSlide lays out a slide's top-level nodes and composes them onto ps
// (notes first). It mutates only ps and the renderer's own Stats; the only
// presentation-shared touch is the global media manager, which is concurrency-
// safe — and slides that reach it are scheduled sequentially (see Render).
//
// VariantDark derives a per-slide dark theme (darkThemeFrom), temporarily
// replaces the presentation's active theme so that all token colors resolve to
// dark-palette values, draws a dark canvas fill when no explicit Background is
// set, then restores the original theme on return. Dark slides always render in
// the sequential pass (slideNeedsSerial) so the theme swap is never concurrent
// with another slide's composition.
func (r *renderer) composeSlide(ps *pptx.Slide, sl *SceneSlide) {
	// VariantDark: swap in the dark theme for this slide's composition.
	// The presentation theme is restored via defer so it is reset even if a
	// future code path adds an early return. VariantPrint is not yet implemented
	// and surfaces a LayoutWarning instead of silently using the light theme.
	switch sl.Variant {
	case VariantDark:
		orig := r.pres.Theme()
		dark := darkThemeFrom(orig)
		r.pres.SetTheme(dark) // token resolution in AddShape/AddRun now uses dark
		r.theme = dark        // spacing/radius lookups on the per-slide renderer
		defer r.pres.SetTheme(orig)
	case VariantLight:
		// default — no change
	default:
		r.warn(sl.ID, fmt.Sprintf("theme variant %q requested but variant selection is not yet implemented; rendered with the active theme", sl.Variant))
	}

	// Background fill — drawn first so it sits behind all decorations and body
	// content (the z-order requirement from RFC §10.2 and the Phase-13 spec).
	r.renderBackground(ps, sl)

	if len(sl.Notes) > 0 {
		nf := ps.SpeakerNotes()
		p := nf.AddParagraph(pptx.ParagraphOpts{})
		r.addRichText(ps, p, sl.Notes, pptx.TypeBody)
	}

	for _, pl := range r.layout(sl.Nodes, sl.ID, sl.Content) {
		r.renderNode(ps, pl.box, pl.node, sl.ID, pl.hAlign)
	}

	// Chrome is drawn last so the eyebrow and footer page number stay visible
	// even over a full-bleed background or section divider. It occupies the
	// margin the body region vacated (see bodyRegion), so it never overlaps the
	// body content. Inert when chrome is disabled.
	r.renderChrome(ps, sl)
}

// renderBackground draws a full-slide fill as the lowest layer of the slide
// when sl.Background.Kind != BackgroundNone. For VariantDark + BackgroundNone,
// it draws a dark canvas rect so a bare dark slide is visually dark rather than
// inheriting the presentation's white default.
//
// The presentation's active theme is already dark (swapped by composeSlide)
// when this is called for a VariantDark slide, so TokenColor roles resolve to
// the dark palette automatically.
func (r *renderer) renderBackground(ps *pptx.Slide, sl *SceneSlide) {
	cx, cy := r.pres.SlideSize()
	full := pptx.Box{X: 0, Y: 0, W: pptx.EMU(cx), H: pptx.EMU(cy)}
	bg := sl.Background

	// Draw the base fill; only overlay a scrim when a fill was actually drawn so a
	// scrim never hangs over an empty (BackgroundNone, light) slide.
	if r.drawBackgroundFill(ps, sl, full, bg) && bg.Scrim != nil {
		r.renderScrim(ps, full, *bg.Scrim)
	}
}

// renderScrim draws a darkening/tinting overlay over the slide's background fill
// for text legibility (R14.1). It is a pure mechanism (D-026): the caller chooses
// the color/opacity. A gradient scrim runs transparent → Color at Opacity along
// GradientAngle (zero defaults to 90°, top transparent → bottom dense).
func (r *renderer) renderScrim(ps *pptx.Slide, full pptx.Box, s Scrim) {
	var fill pptx.Fill
	if s.Gradient {
		angle := s.GradientAngle
		if angle == 0 {
			angle = 90
		}
		fill = pptx.LinearGradient(float64(angle),
			pptx.GradientStop{Pos: 0, Color: pptx.TokenColorAlpha(s.Color, 0)},
			pptx.GradientStop{Pos: 1, Color: pptx.TokenColorAlpha(s.Color, s.Opacity)},
		)
	} else {
		fill = pptx.SolidFill(pptx.TokenColorAlpha(s.Color, s.Opacity))
	}
	ps.AddShape(pptx.ShapeRect, full, pptx.WithFill(fill))
	r.stats.Shapes++
}

// drawBackgroundFill draws the slide's base background fill and reports whether
// anything was drawn (so the caller knows whether a scrim overlay applies). It is
// the body of renderBackground; splitting it lets the scrim overlay run after a
// successful fill regardless of kind.
func (r *renderer) drawBackgroundFill(ps *pptx.Slide, sl *SceneSlide, full pptx.Box, bg Background) bool {
	switch bg.Kind {
	case BackgroundNone:
		if sl.Variant != VariantDark {
			return false // no background fill; byte-identical to pre-Phase-13 output
		}
		// Dark variant with no explicit background: fill the canvas with the
		// dark canvas color so the slide is not transparently white.
		ps.AddShape(pptx.ShapeRect, full,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorCanvas))))
		r.stats.Shapes++
		return true

	case BackgroundColor:
		ps.AddShape(pptx.ShapeRect, full,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(bg.Color))))
		r.stats.Shapes++
		return true

	case BackgroundGradient:
		stops, ok := backgroundGradientStopsFor(bg)
		if !ok {
			r.warn(sl.ID, fmt.Sprintf("background gradient stops invalid (need 2..8 ascending in [0,1]): %v", bg.Stops))
			return false
		}
		ps.AddShape(pptx.ShapeRect, full, pptx.WithFill(pptx.LinearGradient(float64(bg.Angle), stops...)))
		r.stats.Shapes++
		return true

	case BackgroundRadial:
		stops, ok := backgroundGradientStopsFor(bg)
		if !ok {
			r.warn(sl.ID, fmt.Sprintf("background radial stops invalid (need 2..8 ascending in [0,1]): %v", bg.Stops))
			return false
		}
		ps.AddShape(pptx.ShapeRect, full, pptx.WithFill(pptx.RadialGradient(stops...)))
		r.stats.Shapes++
		return true

	case BackgroundMesh:
		if len(bg.Mesh) == 0 {
			// No glows configured: behave like BackgroundNone (absent config draws
			// nothing on a light slide; a dark variant still gets its dark canvas).
			if sl.Variant != VariantDark {
				return false
			}
			ps.AddShape(pptx.ShapeRect, full,
				pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorCanvas))))
			r.stats.Shapes++
			return true
		}
		// Base canvas fill (the paper/dark canvas the glows pool over), then the
		// glows in slice order — deterministic. bg.Color zero = ColorCanvas.
		ps.AddShape(pptx.ShapeRect, full,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(bg.Color))))
		r.stats.Shapes++
		for _, g := range bg.Mesh {
			if g.Radius <= 0 {
				continue
			}
			center := g.Anchor.Point(full)
			gbox := pptx.Box{X: center.X - g.Radius, Y: center.Y - g.Radius, W: 2 * g.Radius, H: 2 * g.Radius}
			ps.AddShape(pptx.ShapeEllipse, gbox, pptx.WithFill(pptx.RadialGradient(
				pptx.GradientStop{Pos: 0, Color: pptx.TokenColorAlpha(g.Color, g.Alpha)},
				pptx.GradientStop{Pos: 1, Color: pptx.TokenColorAlpha(g.Color, 0)},
			)))
			r.stats.Shapes++
		}
		return true

	case BackgroundAsset:
		data, ct, err := r.resolve(bg.AssetID)
		if err != nil {
			r.warn(sl.ID, fmt.Sprintf("background asset %q unresolved: %v", bg.AssetID, err))
			return false
		}
		img, aerr := ps.AddImage(pptx.ImageBytes(data, ct), full)
		if aerr != nil {
			r.warn(sl.ID, fmt.Sprintf("background image %q: %v", bg.AssetID, aerr))
			return false
		}
		if bg.Duotone != nil {
			img.SetDuotone(pptx.TokenColor(bg.Duotone.Shadow), pptx.TokenColor(bg.Duotone.Highlight))
		}
		r.stats.Shapes++
		r.stats.Assets++
		return true
	}
	return false
}

// backgroundGradientStopsFor resolves a background's gradient stops, shared by
// the linear (BackgroundGradient) and radial (BackgroundRadial) paths (D-106). A
// non-empty Stops list is validated via backgroundGradientStops (2..8 ascending
// in [0,1]); otherwise the legacy two-role Gradient pair maps to stops at Pos 0
// and 1. The legacy mapping is identical to the pre-D-106 linear path, so a
// two-role gradient background stays byte-identical.
func backgroundGradientStopsFor(bg Background) ([]pptx.GradientStop, bool) {
	if len(bg.Stops) > 0 {
		return backgroundGradientStops(bg.Stops)
	}
	return []pptx.GradientStop{
		{Pos: 0, Color: pptx.TokenColor(bg.Gradient[0])},
		{Pos: 1, Color: pptx.TokenColor(bg.Gradient[1])},
	}, true
}

// backgroundGradientStops validates scene gradient stops and maps them to the
// builder's gradient stops (D-105). It accepts 2..8 stops, each Pos in [0,1] and
// strictly ascending, and returns (stops, true); any violation returns (nil,
// false) so renderBackground can warn and skip (RFC §10.2, D-026 — no panic).
// The mapping is pure (TokenColor + Pos), so output is deterministic regardless
// of worker count.
func backgroundGradientStops(in []GradientStop) ([]pptx.GradientStop, bool) {
	if len(in) < 2 || len(in) > 8 {
		return nil, false
	}
	out := make([]pptx.GradientStop, len(in))
	prev := -1.0
	for i, s := range in {
		if s.Pos < 0 || s.Pos > 1 || s.Pos <= prev {
			return nil, false
		}
		prev = s.Pos
		out[i] = pptx.GradientStop{Pos: s.Pos, Color: pptx.TokenColor(s.Color)}
	}
	return out, true
}

// slideUsesAssets reports whether composing sl may register global media (images,
// charts, code-block rasters). Such slides render sequentially in scene order so
// media part numbering stays deterministic (RFC §10.1 byte-identical
// requirement); see Render and slideNeedsSerial.
func slideUsesAssets(sl *SceneSlide) bool {
	return nodesUseAssets(sl.Nodes) || sl.Background.Kind == BackgroundAsset
}

// slideNeedsSerial reports whether sl must compose in the sequential pass rather
// than the parallel free pool. It extends slideUsesAssets with VariantDark: dark
// slides temporarily swap the presentation's active theme (see composeSlide),
// which is not concurrent-safe; they must render sequentially so no other slide
// composition reads the presentation theme while it is dark.
func slideNeedsSerial(sl *SceneSlide) bool {
	return slideUsesAssets(sl) || sl.Variant == VariantDark
}

func nodesUseAssets(nodes []SlideNode) bool {
	for _, n := range nodes {
		if nodeUsesAssets(n) {
			return true
		}
	}
	return false
}

// nodeUsesAssets classifies a node as asset-bearing (it or a descendant may
// register media). It is conservative: an unrecognized node is assumed to use
// assets so it renders sequentially — a new node type can never silently break
// idempotency, only forgo parallelism until it is classified here.
func nodeUsesAssets(n SlideNode) bool {
	switch v := n.(type) {
	case CodeBlock, Image, Chart:
		return true
	case Decoration:
		return v.Kind == DecorationAsset // preset ornaments are native shapes
	case TwoColumn:
		return nodesUseAssets(v.Left) || nodesUseAssets(v.Right)
	case Grid:
		return nodesUseAssets(v.Cells)
	case Card:
		// The card chrome (incl. a native custGeom icon) registers no media; an
		// asset-bearing body node does (D-015, D-043), as does a cover-fit photo
		// surface fill (ImageFill registers a media part — R14.1, D-117).
		return v.ImageFill != "" || nodesUseAssets(v.Body)
	case CardSection:
		return nodesUseAssets(v.Body)
	case Bento:
		return nodesUseAssets(v.cellNodes())
	case Banner:
		// The strip + icon register no media; only an asset-bearing Trailing child does.
		return nodesUseAssets(v.Trailing)
	case Lockup:
		// The asset variant registers a pic; the icon variant is media-free (R12.9).
		return v.AssetID != ""
	case Flow:
		// Native pills + connectors + custGeom step icons register no media.
		return false
	case Hero, Prose, Heading, List, Divider, Quote, Callout, Chip, Arrow, Stat, Button, Checklist, ChipRow, IconRows, SectionDivider, Table:
		// Button / Checklist / ChipRow / IconRows are native (pills / glyphs + custGeom icons) — no media.
		return false
	default:
		return true
	}
}

// bodyRegion returns the margin-inset content region of the slide. When chrome
// is enabled (Phase 24) the region shrinks by the eyebrow band at the top and
// the footer band at the bottom, so the body never extends into the chrome and
// overlap is structurally impossible.
func (r *renderer) bodyRegion() pptx.Box {
	cx, cy := r.pres.SlideSize() // EMU
	top := bodyMargin
	bottom := bodyMargin
	if r.chrome.Enabled {
		top += chromeEyebrowH + chromeBandGap
		bottom += chromeFooterH + chromeBandGap
	}
	return pptx.Box{
		X: bodyMargin,
		Y: top,
		W: pptx.EMU(cx) - 2*bodyMargin,
		H: pptx.EMU(cy) - top - bottom,
	}
}

// safeArea is the slide's printable region — the slide box minus the content
// margins minus the reserved chrome bands (R11.3). It is the body region: nothing
// a container draws may extend below its bottom. A named alias of bodyRegion so the
// clamp's intent is legible and the constant has a single source of truth.
func (r *renderer) safeArea() pptx.Box { return r.bodyRegion() }

// clampToSafeArea caps box so its bottom never exceeds the slide safe area (R11.3,
// D-083): a container handed a box that an over-full stack pushed below the
// printable region would otherwise divide that box and draw cells off the slide
// bottom (overlapping the chrome footer). When box already fits (Bottom() <=
// safeArea bottom) it is returned unchanged — byte-identical; only an overflowing
// box is shrunk, and a single warning is logged. Pure integer clamp → deterministic
// at any worker count. Defense-in-depth complementary to the opt-in VAlignFit
// (which reflows an over-full stack before placement); this caps the drawn height
// so the off-canvas invariant holds even for a default top-anchored stack.
func (r *renderer) clampToSafeArea(box pptx.Box, slideID string) pptx.Box {
	if sb := r.safeArea().Bottom(); box.Bottom() > sb {
		h := sb - box.Y
		if h < 0 {
			h = 0
		}
		if h < box.H {
			box.H = h
			r.warn(slideID, "container overflow: content exceeds the slide safe area, clamped")
		}
	}
	return box
}

// layout assigns each top-level node a slot, imposing the RFC §10.2 z-order:
// background decorations first (behind), then section dividers and the stacked
// body, then foreground decorations (on top). Decorations are overlays placed
// against the full slide (anchor-relative) and do not consume body-stack height;
// every other node stacks vertically in the body region in IR order.
//
// align carries the slide's Content alignment; it is applied only to the body
// stack (not to decorations or section-dividers, which are full-slide overlays).
func (r *renderer) layout(nodes []SlideNode, slideID string, align Alignment) []placement {
	cx, cy := r.pres.SlideSize()
	fullSlide := pptx.Box{X: 0, Y: 0, W: pptx.EMU(cx), H: pptx.EMU(cy)}
	var bg, fg, sections, stacked []SlideNode
	for _, n := range nodes {
		switch d := n.(type) {
		case Decoration:
			if d.Layer == LayerForeground {
				fg = append(fg, n)
			} else {
				bg = append(bg, n)
			}
		case SectionDivider:
			sections = append(sections, n)
		default:
			stacked = append(stacked, n)
		}
	}
	var out []placement
	for _, n := range bg {
		out = append(out, placement{node: n, box: fullSlide})
	}
	for _, n := range sections {
		out = append(out, placement{node: n, box: fullSlide})
	}
	out = append(out, r.alignedStackIn(r.bodyRegion(), stacked, slideID, align)...)
	for _, n := range fg {
		out = append(out, placement{node: n, box: fullSlide})
	}
	return out
}

// stackIn places nodes top-to-bottom within box (full box width, per-node
// preferred height, SpaceMD gap). Shared by the body layout and container
// columns. Overflow past the box records a LayoutWarning.
func (r *renderer) stackIn(box pptx.Box, nodes []SlideNode, slideID string) []placement {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	out := make([]placement, 0, len(nodes))
	y := box.Y
	for _, n := range nodes {
		h := preferredHeight(n, box.W, r.theme)
		out = append(out, placement{node: n, box: pptx.Box{X: box.X, Y: y, W: box.W, H: h}})
		y += h + gap
	}
	if len(nodes) > 0 && y-gap > box.Bottom() {
		r.warn(slideID, "content overflows its region")
	}
	return out
}

// renderNode dispatches a node to its composer per the §12 policy.
// hAlign is the effective horizontal alignment computed by alignedStackIn for
// this node; text leaf renderers (Hero/Heading/Prose/Quote) use it to set
// ParagraphOpts.Align. Container and visual nodes ignore it.
func (r *renderer) renderNode(ps *pptx.Slide, box pptx.Box, n SlideNode, slideID string, hAlign HAlign) {
	// Safe-area clamp (R11.3, generalized in R11.12/D-092): cap every content node's
	// box so its bottom never exceeds the slide safe area — not just top-level
	// containers (Bento/Grid/Card) but also a leaf that an over-full card body or
	// stack would push off-slide. Full-slide overlays — Decoration (which may bleed
	// off-canvas by design) and SectionDivider — are exempt; they are not body
	// content. Byte-identical when the box already fits (the clamp is a no-op).
	switch n.(type) {
	case Decoration, SectionDivider:
		// full-slide overlay / intentional bleed — not clamped
	default:
		box = r.clampToSafeArea(box, slideID)
	}
	switch v := n.(type) {
	case Hero:
		r.renderHero(ps, box, v, hAlign)
	case Prose:
		r.renderProse(ps, box, v, hAlign)
	case Heading:
		r.renderHeading(ps, box, v, hAlign)
	case List:
		r.renderList(ps, box, v)
	case Divider:
		r.renderDivider(ps, box, v)
	case Quote:
		r.renderQuote(ps, box, v, hAlign)
	case Callout:
		r.renderCallout(ps, box, v)
	case Chip:
		r.renderChip(ps, box, v)
	case Arrow:
		r.renderArrow(ps, box, v)
	case CodeBlock:
		r.renderCodeBlock(ps, box, v, slideID)
	case Image:
		r.renderImage(ps, box, v, slideID)
	case Decoration:
		r.renderDecoration(ps, box, v, slideID)
	case SectionDivider:
		r.renderSectionDivider(ps, box, v)
	case TwoColumn:
		r.renderTwoColumn(ps, box, v, slideID)
	case Grid:
		r.renderGrid(ps, box, v, slideID)
	case Table:
		r.renderTable(ps, box, v, slideID)
	case Card:
		r.renderCard(ps, box, v, slideID)
	case CardSection:
		r.renderCardSection(ps, box, v, slideID)
	case Bento:
		r.renderBento(ps, box, v, slideID)
	case Flow:
		r.renderFlow(ps, box, v, slideID)
	case Chart:
		r.renderChart(ps, box, v, slideID)
	case Stat:
		r.renderStat(ps, box, v)
	case Button:
		r.renderButton(ps, box, v, hAlign)
	case Checklist:
		r.renderChecklist(ps, box, v)
	case ChipRow:
		r.renderChipRow(ps, box, v, hAlign)
	case Banner:
		r.renderBanner(ps, box, v, slideID)
	case IconRows:
		r.renderIconRows(ps, box, v)
	case Lockup:
		r.renderLockup(ps, box, v, slideID, hAlign)
	default:
		r.warn(slideID, fmt.Sprintf("%s rendering is not yet implemented; node skipped", n.NodeKind()))
	}
}

// addRichText maps a scene RichText onto a builder paragraph: one run per
// TextRun, carrying inline style + color, at the node's base type role. A link
// run becomes a hyperlink. Returns the number of resolver-dependent extras (0
// for text).
func (r *renderer) addRichText(ps *pptx.Slide, p *pptx.Paragraph, rt RichText, base pptx.TypeRole) {
	r.addRichTextScaled(ps, p, rt, base, 0)
}

// addRichTextScaled is addRichText with a uniform per-run FontScale applied to
// every run (the AutoFit shrink-to-fit path, R10.5). A scale of 0 leaves the
// runs at their role size — byte-identical to addRichText.
func (r *renderer) addRichTextScaled(_ *pptx.Slide, p *pptx.Paragraph, rt RichText, base pptx.TypeRole, scale float64) {
	for _, run := range rt {
		style := pptx.RunStyle{
			TypeRole:  base,
			Bold:      run.Style.Bold,
			Italic:    run.Style.Italic,
			Code:      run.Style.Code,
			FontScale: scale,
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
//
// Text-bearing nodes are content-aware (Phase 22, RFC §10.2): their height grows
// with the number of lines the text wraps to in the available width avail,
// estimated by wrappedLines against theme. A single line reproduces the node's
// pre-Phase-22 fixed height exactly, so single-line content stays byte-identical;
// avail <= 0 or a nil theme also falls back to the fixed (single-line) height.
// Visual and atom nodes (Hero, Divider, Chip, Arrow, Image, Chart, CodeBlock,
// Flow) do not wrap and keep fixed slot heights.
func preferredHeight(n SlideNode, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	switch v := n.(type) {
	case Hero:
		return pptx.In(2.2)
	case Heading:
		lines := wrappedLines(v.Text, headingRole(v.Level), avail, theme)
		return pptx.In(0.6) * pptx.EMU(lines)
	case Prose:
		if len(v.Paragraphs) == 0 {
			return pptx.In(0.4)
		}
		var h pptx.EMU
		for _, para := range v.Paragraphs {
			lines := wrappedLines(para, pptx.TypeBody, avail, theme)
			h += pptx.In(0.4) * pptx.EMU(lines)
		}
		return h
	case List:
		if len(v.Items) == 0 {
			return pptx.In(0.32)
		}
		var h pptx.EMU
		for _, item := range v.Items {
			lines := wrappedLines(item.Text, pptx.TypeBody, avail, theme)
			h += pptx.In(0.32) * pptx.EMU(lines)
		}
		return h
	case Divider:
		return pptx.In(0.2)
	case Quote:
		// Fixed chrome (attribution + padding) is one In(1.1) slot; each extra
		// wrapped line of the quote text adds one quote line-height.
		lines := wrappedLines(v.Text, pptx.TypeH3, avail, theme)
		return pptx.In(1.1) + quoteLineEst*pptx.EMU(lines-1)
	case Callout:
		// The body wraps within the box minus the accent bar + text inset
		// (mirrors renderCallout's In(0.2) inset).
		lines := wrappedLines(v.Body, pptx.TypeBody, avail-calloutInsetEst, theme)
		return pptx.In(1.0) + calloutLineEst*pptx.EMU(lines-1)
	case Chip:
		return pptx.In(0.4)
	case Stat:
		// A fixed number block: value (display) + label + optional delta line.
		h := pptx.In(1.2)
		if v.Delta != "" {
			h += pptx.In(0.3)
		}
		return h
	case Checklist:
		return checklistPreferredHeight(v, avail, theme)
	case ChipRow:
		return chipRowPreferredHeight(v, avail, theme)
	case Banner:
		return bannerPreferredHeight(v, avail, theme)
	case IconRows:
		return iconRowsPreferredHeight(v, avail, theme)
	case Lockup:
		return lockupPreferredHeight(v)
	case Arrow:
		return pptx.In(0.6)
	case Button:
		// A fixed pill: its slot is the size-driven button height (the pill itself
		// is content-fit horizontally; the slot height does not vary with the label).
		return buttonMetrics(v.Size).height
	case CodeBlock:
		return pptx.In(2.6)
	case Image:
		return pptx.In(3.0)
	case Chart:
		return pptx.In(3.0)
	case Table:
		return tableHeight(v, avail, theme)
	case TwoColumn:
		colW := (avail - estGap) / 2
		h := maxEMU(nodesHeight(v.Left, colW, theme), nodesHeight(v.Right, colW, theme))
		// A top/bottom bridge reserves a band beyond the columns (R12.8); JoinSeam adds 0.
		if v.Join != JoinNone && v.JoinPosition != JoinSeam {
			h += bridgeBandH
		}
		return h
	case Grid:
		cols := v.Columns
		if cols < 1 {
			cols = 1
		}
		rows := (len(v.Cells) + cols - 1) / cols
		if rows < 1 {
			rows = 1
		}
		cellW := (avail - estGap*pptx.EMU(cols-1)) / pptx.EMU(cols)
		var maxCell pptx.EMU
		for _, c := range v.Cells {
			if h := preferredHeight(c, cellW, theme); h > maxCell {
				maxCell = h
			}
		}
		return pptx.EMU(rows)*maxCell + estGap*pptx.EMU(rows-1)
	case Card:
		// Wrapped-header-aware chrome estimate (R10.10): the fixed cardChromeEst
		// baseline plus the extra eyebrow/title lines a multi-line header wraps to,
		// so the slot accounts for a wrapped header. Single-line → +0 (byte-identical).
		c := cardChrome{header: v.Header, eyebrow: v.Eyebrow, icon: v.Icon, pill: v.HeaderPill, size: v.Size, layout: v.Layout, paddingScale: v.PaddingScale, ribbon: v.Ribbon}
		return cardChromeEst + cardHeaderExtraHeight(theme, avail, c) + nodesHeight(v.Body, avail-2*cardBodyInsetEst, theme) + estGap
	case CardSection:
		c := cardChrome{header: v.Header}
		return cardChromeEst + cardHeaderExtraHeight(theme, avail, c) + nodesHeight(v.Body, avail-2*cardBodyInsetEst, theme) + estGap
	case Bento:
		cols := v.Columns
		if cols < 1 {
			cols = 1
		}
		nRows := len(v.Rows)
		if nRows < 1 {
			nRows = 1
		}
		// Mirror bentoGeometry's content width: reserve the left label gutter when
		// any row is labeled and subtract the inter-column gaps, so the span-1
		// reference width is accurate and the slot is not under-sized (a too-wide
		// reference under-counts wrapped lines). Wider-span cells still over-
		// estimate height with this unit width, which is safe (a taller slot).
		contentW := avail
		if gutterW := bentoGutterWidthOf(theme, v); gutterW > 0 {
			contentW -= gutterW + estGap
		}
		unitW := (contentW - estGap*pptx.EMU(cols-1)) / pptx.EMU(cols)
		var maxCell pptx.EMU
		for _, row := range v.Rows {
			for _, cell := range row.Cells {
				// Measure each cell at its ACTUAL span width (R10.10): a span-S cell
				// renders S units wide and wraps less, so the unit-width estimate
				// over-counts its lines. A span-1 cell yields unitW (byte-identical).
				span := cell.Span
				if span < 1 {
					span = 1
				}
				spanW := pptx.EMU(span)*unitW + estGap*pptx.EMU(span-1)
				if h := preferredHeight(cell.Node, spanW, theme); h > maxCell {
					maxCell = h
				}
			}
		}
		if maxCell == 0 {
			maxCell = pptx.In(0.4)
		}
		return pptx.EMU(nRows)*maxCell + estGap*pptx.EMU(nRows-1)
	case Flow:
		if v.Orientation == FlowVertical {
			n := len(v.Steps)
			if n < 1 {
				n = 1
			}
			return pptx.In(0.9) * pptx.EMU(n)
		}
		return pptx.In(1.4)
	default:
		return pptx.In(1.0)
	}
}

// Placement estimates (deterministic; not content opinions, D-026). Pinned
// compile-time EMU literals so output is worker-count-independent (RFC §10.1).
const (
	cardChromeEst = pptx.EMU(1097280) // ~1.2"; card header row + padding above the body
	estGap        = pptx.EMU(137160)  // ~0.15"; nested-container gap estimate

	// Content-aware (Phase 22) increments and insets. quoteLineEst / calloutLineEst
	// are the per-extra-wrapped-line height added to the fixed-chrome nodes;
	// calloutInsetEst / cardBodyInsetEst approximate the horizontal space the
	// node's text loses to chrome so the wrap estimate uses the true text column.
	quoteLineEst     = pptx.EMU(411480) // ~0.45"; per extra wrapped line of a Quote
	calloutLineEst   = pptx.EMU(274320) // ~0.30"; per extra wrapped line of a Callout body
	calloutInsetEst  = pptx.EMU(182880) // ~0.20"; accent bar + text inset (renderCallout)
	cardBodyInsetEst = pptx.EMU(182880) // ~0.20"; per-side card body padding estimate
)

// tableHeight estimates a Table's slot height: each row is the tallest cell in
// it (the cell's wrapped line count × a row line-height), summed over the header
// (if any) and body rows, plus a caption line. It falls back to the count-based
// pre-Phase-22 height when the column width can't be estimated (no columns, or
// no width/theme), which keeps single-line tables byte-identical.
func tableHeight(v Table, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	cols := tableColumns(v)
	// A grouped header (R14.3) adds one row above the headers.
	groupRows := 0
	if v.Style != nil && len(v.Style.HeaderGroups) > 0 {
		groupRows = 1
	}
	if cols < 1 || avail <= 0 || theme == nil {
		rows := len(v.Rows) + groupRows
		if len(v.Headers) > 0 {
			rows++
		}
		h := pptx.In(0.4) * pptx.EMU(rows)
		if v.Caption != "" {
			h += pptx.In(0.4)
		}
		return h
	}
	colW := avail / pptx.EMU(cols)
	var h pptx.EMU
	h += pptx.In(0.4) * pptx.EMU(groupRows) // grouped header row (single line of labels)
	if len(v.Headers) > 0 {
		h += tableRowHeight(v.Headers, colW, theme)
	}
	for _, row := range v.Rows {
		h += tableRowHeight(row, colW, theme)
	}
	if v.Caption != "" {
		h += pptx.In(0.4)
	}
	return h
}

// tableRowHeight is In(0.4) times the wrapped line count of the tallest cell in
// the row (a single-line row is In(0.4), byte-identical to the prior model).
func tableRowHeight(cells []RichText, colW pptx.EMU, theme *pptx.Theme) pptx.EMU {
	maxLines := 1
	for _, cell := range cells {
		if l := wrappedLines(cell, pptx.TypeBody, colW, theme); l > maxLines {
			maxLines = l
		}
	}
	return pptx.In(0.4) * pptx.EMU(maxLines)
}

// nodesHeight estimates the stacked height of a node list laid out in a column
// of width avail (for container slot sizing).
func nodesHeight(nodes []SlideNode, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	var sum pptx.EMU
	for i, n := range nodes {
		if i > 0 {
			sum += estGap
		}
		sum += preferredHeight(n, avail, theme)
	}
	return sum
}

func maxEMU(a, b pptx.EMU) pptx.EMU {
	if a > b {
		return a
	}
	return b
}

// alignedStackIn is the body-stack layout with Phase-13 alignment applied. It
// serves the top-level slide body stack and (R10.4) a Card's vertical body when
// it opts into Card.BodyVAlign; the other container renderers (TwoColumn, Grid,
// CardSection) still call stackIn so their sub-layouts are unaffected by
// slide-level alignment.
//
// With the zero Alignment {VAlignTop, HAlignLeft} and no per-node Align
// overrides, alignedStackIn produces placements byte-identical to stackIn
// (the pre-Phase-13 behavior). No fast-path is used because per-node Align
// fields must be checked even when the slide's Alignment is zero.
func (r *renderer) alignedStackIn(box pptx.Box, nodes []SlideNode, slideID string, align Alignment) []placement {
	n := len(nodes)
	if n == 0 {
		return nil
	}

	gap := r.theme.ResolveSpace(pptx.SpaceMD)

	// Compute per-node preferred heights and the total stack height
	// (node heights + standard gap between each pair).
	heights := make([]pptx.EMU, n)
	var bodyH pptx.EMU // sum of node heights only (no gaps)
	for i, nd := range nodes {
		heights[i] = preferredHeight(nd, box.W, r.theme)
		bodyH += heights[i]
	}
	// totalH = bodyH + gap*(n-1); gap appears between nodes, not after the last.
	var gapCount pptx.EMU
	if n > 1 {
		gapCount = pptx.EMU(n - 1)
	}
	totalH := bodyH + gap*gapCount

	// Vertical: compute the Y coordinate of the first node.
	startY := box.Y
	switch align.Vertical {
	case VAlignCenter:
		slack := box.H - totalH
		if slack > 0 {
			startY = box.Y + slack/2
		}
	case VAlignBottom:
		candidate := box.Bottom() - totalH
		if candidate > box.Y {
			startY = candidate
		}
		// VAlignTop and VAlignJustify both start at box.Y; Justify adjusts the gap.
	}

	// Effective inter-node gap: VAlignJustify distributes slack into the gaps.
	effectiveGap := gap
	if align.Vertical == VAlignJustify && n > 1 {
		slack := box.H - bodyH // total vertical space available for gaps
		if slack > gap*pptx.EMU(n-1) {
			effectiveGap = slack / pptx.EMU(n-1)
		}
	}

	// VAlignFit (R10.2): deterministic fit-to-region compression. When the body
	// stack overflows, shrink the inter-node gaps toward a pinned floor, then —
	// if still overflowing — compress slot heights toward a pinned ratio floor,
	// so the last node lands inside the region instead of clipping off-slide.
	// Top-pinned (startY stays box.Y). Fitting content is left untouched, so a
	// stack that already fits is byte-identical to VAlignTop. Mutates heights.
	if align.Vertical == VAlignFit && totalH > box.H {
		effectiveGap = r.fitCompress(heights, bodyH, gap, box)
	}

	// Overflow warning: same semantics as stackIn (fires when the content
	// is taller than the box, regardless of how vertical alignment clamped it).
	// For VAlignFit the check is recomputed against the post-compression
	// geometry, so a successful fit suppresses the warning (and an overflow that
	// the pinned floors cannot fully absorb still surfaces honestly).
	overflowing := totalH > box.H
	if align.Vertical == VAlignFit && overflowing {
		var fitted pptx.EMU
		for _, h := range heights {
			fitted += h
		}
		overflowing = fitted+effectiveGap*gapCount > box.H
	}
	if overflowing {
		r.warn(slideID, "content overflows its region")
	}

	// VAlignFill: grow the flexible nodes (containers + Image/Chart) to consume
	// the leftover body height, so the last node's bottom reaches box.Bottom().
	// Top-pinned (startY stays box.Y) with the standard gap; only positive slack
	// is distributed, so fill never overlaps and never fights the overflow case.
	if align.Vertical == VAlignFill {
		if slack := box.H - totalH; slack > 0 {
			distributeFill(nodes, heights, slack)
		}
	}

	// VAlignFillCapped (R10.6): like VAlignFill but each flexible node grows by at
	// most a pinned factor of its preferred height, so a sparse node cannot
	// balloon. The slack beyond the caps (residual) is distributed as balanced
	// spacing — an even top margin plus widened inter-node gaps (residual/(n+1)) —
	// instead of inflating one node. Integer math, worker-count independent.
	if align.Vertical == VAlignFillCapped {
		if slack := box.H - totalH; slack > 0 {
			used := distributeFillCapped(nodes, heights, slack)
			if residual := slack - used; residual > 0 {
				space := residual / pptx.EMU(n+1)
				startY = box.Y + space
				effectiveGap = gap + space
			}
		}
	}

	// VAlignBalanced (R10.8): distribute a sparse stack's slack as an even rhythm —
	// a top margin plus widened inter-node gaps (slack split across the n+1 spaces)
	// — with an optical-center upward bias (the top margin is balancedOpticalBP of
	// an even unit, so the stack seats slightly above geometric center). Integer
	// math, worker-count independent.
	if align.Vertical == VAlignBalanced {
		if slack := box.H - totalH; slack > 0 {
			unit := slack / pptx.EMU(n+1)
			startY = box.Y + unit*balancedOpticalBP/10000
			effectiveGap = gap + unit
		}
	}

	out := make([]placement, 0, n)
	y := startY
	for i, nd := range nodes {
		h := heights[i]
		hAlign := nodeEffectiveHAlign(nd, align.Horizontal)

		plBox := pptx.Box{X: box.X, Y: y, W: box.W, H: h}

		// Chip is a physical pill that should move: narrow its box and offset X
		// so the pill shape itself is positioned at center/right.
		// Text leaf nodes (Hero, Heading, Prose, Quote) keep the full body-width
		// box — their paragraph alignment (set by the renderer via hAlignToParagraph)
		// handles the per-line centering/right-alignment within the full frame.
		if hAlign != HAlignLeft {
			if _, isChip := nd.(Chip); isChip {
				nw := nodeNaturalWidth(nd, r.theme)
				if nw > box.W {
					nw = box.W
				}
				if nw > 0 && nw < box.W {
					var offsetX pptx.EMU
					switch hAlign {
					case HAlignCenter:
						offsetX = (box.W - nw) / 2
					case HAlignRight:
						offsetX = box.W - nw
					}
					plBox.X = box.X + offsetX
					plBox.W = nw
				}
			}
		}

		out = append(out, placement{node: nd, box: plBox, hAlign: hAlign})
		y += h + effectiveGap
	}
	return out
}

// isFlexible reports whether a node grows under VAlignFill. The flexible set is
// intrinsic (D-026): the containers (which subdivide a taller box into taller
// cells) plus the two stretchable visuals. Text leaves and atoms stay at
// preferred height — stretching text is meaningless — and CodeBlock is excluded
// because growing a monospaced-code raster distorts the listing.
func isFlexible(n SlideNode) bool {
	switch n.(type) {
	case Grid, TwoColumn, Card, CardSection, Bento, Table, Chart, Image:
		return true
	case Checklist:
		// A checklist subdivides a taller box into evenly spread rows when Fill is set
		// (R12.2); flexible so a VAlignFill/BodyVAlign parent can grow it. A non-Fill
		// checklist in a grown box top-aligns its rows (harmless).
		return true
	case IconRows:
		// Same as Checklist (R12.7): flexible so a Fill icon-row list can span a card.
		return true
	default:
		return false
	}
}

// distributeFill grows the flexible nodes in place so their added heights sum to
// exactly slack (slack > 0). The share is proportional to each flexible node's
// preferred height (the larger node grows more, relative proportions preserved),
// with the rounding remainder assigned to the last flexible node so the total is
// exact. When the flexible heights sum to zero, the slack is split equally. With
// no flexible node nothing grows (the slide top-aligns). Pure integer EMU math,
// so the result is deterministic regardless of worker scheduling.
func distributeFill(nodes []SlideNode, heights []pptx.EMU, slack pptx.EMU) {
	var flex []int
	var flexH pptx.EMU
	for i, nd := range nodes {
		if isFlexible(nd) {
			flex = append(flex, i)
			flexH += heights[i]
		}
	}
	if len(flex) == 0 {
		return
	}
	var used pptx.EMU
	for k, idx := range flex {
		var add pptx.EMU
		switch {
		case k == len(flex)-1:
			add = slack - used // last flexible node absorbs the rounding remainder
		case flexH > 0:
			add = slack * heights[idx] / flexH
		default:
			add = slack / pptx.EMU(len(flex)) // all flexible heights zero → equal split
		}
		heights[idx] += add
		used += add
	}
}

// fillGrowthMaxBP caps how much a flexible node grows under VAlignFillCapped: at
// most this fraction (in basis points) of its preferred height is ADDED, so a
// node grows to at most (1 + fillGrowthMaxBP/10000)× its preferred height.
const fillGrowthMaxBP = 10000 // +1.0× preferred (at most double)

// balancedOpticalBP is the optical-center bias for VAlignBalanced: the top margin
// is this fraction (basis points) of an even spacing unit, so the stack seats
// slightly above geometric center (the freed space falls to the bottom margin).
const balancedOpticalBP = 8500 // top margin = 85% of an even unit

// distributeFillCapped grows the flexible nodes toward their proportional share
// of slack, but caps each node's growth at fillGrowthMaxBP × its preferred
// height, so a sparse node cannot balloon. It mutates heights in place and
// returns the total height actually consumed by growth (≤ slack); the caller
// distributes the residual (slack − used) as balanced spacing. With no flexible
// node (or all flexible heights zero — no cap room) nothing grows and it returns
// 0. Pure integer / basis-point math, so the result is deterministic regardless
// of worker scheduling.
func distributeFillCapped(nodes []SlideNode, heights []pptx.EMU, slack pptx.EMU) pptx.EMU {
	var flex []int
	var flexH pptx.EMU
	for i, nd := range nodes {
		if isFlexible(nd) {
			flex = append(flex, i)
			flexH += heights[i]
		}
	}
	if len(flex) == 0 || flexH <= 0 {
		return 0
	}
	var used pptx.EMU
	for _, idx := range flex {
		share := slack * heights[idx] / flexH            // proportional, floored
		capAdd := heights[idx] * fillGrowthMaxBP / 10000 // ceiling for this node
		if share > capAdd {
			share = capAdd
		}
		heights[idx] += share
		used += share
	}
	return used
}

// fitCompress applies the R10.2 deterministic fit-to-region compression to an
// over-full body stack (the VAlignFit mode). It runs two pinned steps:
//
//  1. Shrink the inter-node gap from gap toward gapMin (the SpaceXS floor) so
//     the reclaimed gap space absorbs the overflow.
//  2. If the stack still overflows at gapMin, scale every node's slot height by
//     a single factor toward a pinned ratio floor (sMin), so the bodies shrink
//     proportionally and the last node's bottom lands inside the region.
//
// It mutates heights in place and returns the compressed inter-node gap. All
// math is integer EMU / basis-point, so the result is a pure function of the
// heights, the gap, and box.H — deterministic regardless of worker scheduling.
// The caller invokes it only when the stack overflows (bodyH + gap*(n-1) >
// box.H); content that already fits is never passed here, so fitting layouts
// stay byte-identical to VAlignTop. The card-padding and display-type-scale
// sub-steps of the R10.2 spec are layered in by separate engine units; the gap
// and slot-height steps here satisfy the up-to-~25% overflow band on their own.
func (r *renderer) fitCompress(heights []pptx.EMU, bodyH, gap pptx.EMU, box pptx.Box) pptx.EMU {
	n := len(heights)
	var gapCount pptx.EMU
	if n > 1 {
		gapCount = pptx.EMU(n - 1)
	}

	// Step 1 — gaps toward the pinned floor (SpaceXS, the smallest spacing token).
	gapMin := r.theme.ResolveSpace(pptx.SpaceXS)
	if gapMin > gap {
		gapMin = gap // never widen the gap
	}
	effGap := gap
	if gapCount > 0 {
		// The gap that would make bodyH + g*gapCount == box.H exactly (floored,
		// so the realized total never exceeds box.H), clamped to [gapMin, gap].
		needed := (box.H - bodyH) / gapCount
		switch {
		case needed < gapMin:
			effGap = gapMin
		case needed < gap:
			effGap = needed
		}
	}

	// Step 2 — if still overflowing at the gap floor, compress slot heights
	// proportionally toward the pinned ratio floor (sMin = 0.60).
	if bodyH > 0 && bodyH+effGap*gapCount > box.H {
		const (
			full   = 10000 // basis-point unity (scale 1.0)
			sMinBP = 6000  // 0.60 — pinned minimum scale (the ratio floor)
		)
		avail := box.H - effGap*gapCount
		if avail < 0 {
			avail = 0
		}
		sBP := avail * full / bodyH
		if sBP < sMinBP {
			sBP = sMinBP
		}
		if sBP < full {
			for i := range heights {
				heights[i] = heights[i] * sBP / full
			}
		}
	}
	return effGap
}

// nodeEffectiveHAlign returns the horizontal alignment that applies to n in
// the body stack. The priority rule mirrors the spec: a non-zero per-node
// Align field overrides the slide-level slideHAlign. Container and visual
// nodes (Grid, TwoColumn, Table, Card, CardSection, Chart, CodeBlock, Image,
// Callout, Flow, Divider, Arrow) always return HAlignLeft so they keep their
// full box width — alignment within them is their own concern.
func nodeEffectiveHAlign(n SlideNode, slideHAlign HAlign) HAlign {
	var nodeAlign HAlign
	switch v := n.(type) {
	case Hero:
		nodeAlign = v.Align
	case Heading:
		nodeAlign = v.Align
	case Prose:
		nodeAlign = v.Align
	case Quote:
		nodeAlign = v.Align
	case Chip:
		nodeAlign = v.Align
	case Button:
		nodeAlign = v.Align
	case ChipRow:
		nodeAlign = v.Align
	case Lockup:
		nodeAlign = v.Align
	case SectionDivider:
		nodeAlign = v.Align
	default:
		// Containers and visuals: always full-width. Not subject to h-align.
		return HAlignLeft
	}
	if nodeAlign != 0 {
		return nodeAlign
	}
	return slideHAlign
}
