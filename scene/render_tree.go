package scene

import "github.com/hurtener/pptx-go/pptx"

// Tree composer (R14.10, D-127). A balanced tidy tree: leaves are spread evenly
// across the cross-axis, internal nodes centered over their leaf descendants, and
// parent→child edges drawn as elbow connectors (all horizontal/vertical, so no
// diagonal-line flip). Top-down (default) puts depth on the Y axis; left-right
// transposes it to the X axis. Pure integer-EMU → byte-identical / deterministic.

const (
	treeNodeMaxW = pptx.EMU(1828800) // In(2.0); a node card's max width
	treeNodeH    = pptx.EMU(548640)  // In(0.60); a node card height
	treeGap      = pptx.EMU(91440)   // In(0.10); inter-node gap within a level
	treeIconSz   = pptx.EMU(228600)  // In(0.25); a node icon
)

type treePlacement struct {
	box  pptx.Box
	node TreeNode
}

func (r *renderer) renderTree(ps *pptx.Slide, box pptx.Box, v Tree, slideID string) {
	depth := treeDepth(v.Root)   // number of levels (>=1)
	leaves := treeLeaves(v.Root) // number of leaf columns (>=1)
	if depth < 1 || leaves < 1 || box.W <= 0 || box.H <= 0 {
		return
	}
	leftRight := v.Orientation == FlowHorizontal

	crossSpan, mainSpan := box.W, box.H
	if leftRight {
		crossSpan, mainSpan = box.H, box.W
	}
	leafSlot := crossSpan / pptx.EMU(leaves)
	levelSlot := mainSpan / pptx.EMU(depth)

	nodeW := leafSlot - treeGap
	if !leftRight && nodeW > treeNodeMaxW {
		nodeW = treeNodeMaxW
	}
	if leftRight {
		nodeW = levelSlot - treeGap
		if nodeW > treeNodeMaxW {
			nodeW = treeNodeMaxW
		}
	}
	if nodeW < treeGap {
		nodeW = treeGap
		r.warn(slideID, "tree: breadth exceeds the region; nodes clamped")
	}

	var placements []treePlacement
	var edges [][2]pptx.Position
	leafIdx := 0

	// walk assigns each node a box (depth on the main axis, leaf-centroid on the
	// cross axis) and records parent→child edge anchors; returns the node's box.
	var walk func(n TreeNode, level int) pptx.Box
	walk = func(n TreeNode, level int) pptx.Box {
		var cross pptx.EMU
		childBoxes := make([]pptx.Box, 0, len(n.Children))
		if len(n.Children) == 0 {
			cross = pptx.EMU(leafIdx)*leafSlot + leafSlot/2
			leafIdx++
		} else {
			var first, last pptx.EMU
			for i := range n.Children {
				cb := walk(n.Children[i], level+1)
				childBoxes = append(childBoxes, cb)
				cc := crossCenterOf(cb, leftRight, box)
				if i == 0 {
					first = cc
				}
				last = cc
			}
			cross = (first + last) / 2
		}
		var nb pptx.Box
		if leftRight {
			nb = pptx.Box{X: box.X + pptx.EMU(level)*levelSlot + (levelSlot-nodeW)/2, Y: box.Y + cross - treeNodeH/2, W: nodeW, H: treeNodeH}
		} else {
			nb = pptx.Box{X: box.X + cross - nodeW/2, Y: box.Y + pptx.EMU(level)*levelSlot + (levelSlot-treeNodeH)/2, W: nodeW, H: treeNodeH}
		}
		placements = append(placements, treePlacement{box: nb, node: n})
		for _, cb := range childBoxes {
			edges = append(edges, parentChildAnchors(nb, cb, leftRight))
		}
		return nb
	}
	walk(v.Root, 0)

	line := pptx.Line{Width: pptx.Pt(1.25), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)}
	for _, e := range edges {
		r.drawElbow(ps, e[0], e[1], leftRight, line)
	}
	for _, pl := range placements {
		r.drawTreeNode(ps, pl.box, pl.node, slideID)
	}
}

// crossCenterOf returns a box's cross-axis center relative to the tree box origin.
func crossCenterOf(b pptx.Box, leftRight bool, base pptx.Box) pptx.EMU {
	if leftRight {
		return b.Y + b.H/2 - base.Y
	}
	return b.X + b.W/2 - base.X
}

// parentChildAnchors returns (parent-out, child-in) anchors: bottom→top for
// top-down, right→left for left-right.
func parentChildAnchors(parent, child pptx.Box, leftRight bool) [2]pptx.Position {
	if leftRight {
		return [2]pptx.Position{{X: parent.Right(), Y: parent.Y + parent.H/2}, {X: child.X, Y: child.Y + child.H/2}}
	}
	return [2]pptx.Position{{X: parent.X + parent.W/2, Y: parent.Bottom()}, {X: child.X + child.W/2, Y: child.Y}}
}

// drawElbow draws a parent→child elbow (out · across the mid bus · in), all
// segments horizontal/vertical.
func (r *renderer) drawElbow(ps *pptx.Slide, from, to pptx.Position, leftRight bool, line pptx.Line) {
	if leftRight {
		midX := (from.X + to.X) / 2
		r.hvLine(ps, from.X, from.Y, midX-from.X, 0, line)
		r.hvLine(ps, midX, minE(from.Y, to.Y), 0, absE(to.Y-from.Y), line)
		r.hvLine(ps, midX, to.Y, to.X-midX, 0, line)
		return
	}
	midY := (from.Y + to.Y) / 2
	r.hvLine(ps, from.X, from.Y, 0, midY-from.Y, line)
	r.hvLine(ps, minE(from.X, to.X), midY, absE(to.X-from.X), 0, line)
	r.hvLine(ps, to.X, midY, 0, to.Y-midY, line)
}

func (r *renderer) hvLine(ps *pptx.Slide, x, y, w, h pptx.EMU, line pptx.Line) {
	if w == 0 && h == 0 {
		return
	}
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: x, Y: y, W: w, H: h}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	r.stats.Shapes++
}

func minE(a, b pptx.EMU) pptx.EMU {
	if a < b {
		return a
	}
	return b
}

func absE(a pptx.EMU) pptx.EMU {
	if a < 0 {
		return -a
	}
	return a
}

// drawTreeNode draws one node as a rounded rect (accent border) with its label.
func (r *renderer) drawTreeNode(ps *pptx.Slide, box pptx.Box, n TreeNode, slideID string) {
	accent := r.accentColorAt(n.AccentIndex)
	ps.AddShape(pptx.ShapeRoundRect, box,
		pptx.WithRadius(pptx.RadiusMD),
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurface))),
		pptx.WithLine(pptx.Line{Width: pptx.Pt(1.25), Color: accent}))
	r.stats.Shapes++

	textX, textW := box.X, box.W
	if n.Icon != "" {
		ib := pptx.Box{X: box.X + treeGap, Y: box.Y + (box.H-treeIconSz)/2, W: treeIconSz, H: treeIconSz}
		if svg, ok := r.cfg.icons.Lookup(n.Icon); ok {
			if _, err := ps.AddIcon(svg, ib, pptx.WithFill(pptx.SolidFill(accent))); err == nil {
				r.stats.Shapes++
			}
		} else {
			r.warn(slideID, "tree node icon "+n.Icon+" not found at compose")
		}
		textX += treeIconSz + treeGap
		textW -= treeIconSz + treeGap
	}
	tf := ps.AddTextFrame(pptx.Box{X: textX, Y: box.Y, W: textW, H: box.H}).Anchor(pptx.AnchorMiddle)
	lp := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	lp.AddRun(n.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
	if n.Detail != "" {
		dp := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		dp.AddRun(n.Detail, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
	}
	r.stats.Shapes++
}

// treeDepth returns the number of levels (a single node = 1).
func treeDepth(n TreeNode) int {
	d := 0
	for i := range n.Children {
		if cd := treeDepth(n.Children[i]); cd > d {
			d = cd
		}
	}
	return d + 1
}

// treeLeaves returns the number of leaf nodes (a single node = 1).
func treeLeaves(n TreeNode) int {
	if len(n.Children) == 0 {
		return 1
	}
	t := 0
	for i := range n.Children {
		t += treeLeaves(n.Children[i])
	}
	return t
}
