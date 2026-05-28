package pptx

// Position is a point on the slide canvas, in EMU from the top-left origin.
type Position struct {
	X, Y EMU
}

// Size is a width/height extent in EMU.
type Size struct {
	W, H EMU
}

// Box is a positioned rectangle: the offset (X, Y) and extent (W, H) of a
// shape on the slide canvas, in EMU. It is the geometry every builder
// primitive accepts.
type Box struct {
	X, Y, W, H EMU
}

// Position returns the box's top-left offset.
func (b Box) Position() Position { return Position{X: b.X, Y: b.Y} }

// Size returns the box's extent.
func (b Box) Size() Size { return Size{W: b.W, H: b.H} }

// Right returns the X coordinate of the box's right edge.
func (b Box) Right() EMU { return b.X + b.W }

// Bottom returns the Y coordinate of the box's bottom edge.
func (b Box) Bottom() EMU { return b.Y + b.H }

// Inset shrinks the box inward by the given inset and returns the result.
func (b Box) Inset(in Inset) Box {
	return Box{
		X: b.X + in.Left,
		Y: b.Y + in.Top,
		W: b.W - in.Left - in.Right,
		H: b.H - in.Top - in.Bottom,
	}
}

// Inset is per-edge padding in EMU.
type Inset struct {
	Top, Right, Bottom, Left EMU
}

// UniformInset returns an Inset with the same value on all four edges.
func UniformInset(v EMU) Inset {
	return Inset{Top: v, Right: v, Bottom: v, Left: v}
}

// Anchor names a reference point on a shape — used to position a Decoration
// or attach a Connector endpoint. The scene renderer translates an Anchor
// (plus offset) into EMU coordinates at render time.
type Anchor int

const (
	AnchorTopLeft Anchor = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorCenterLeft
	AnchorCenter
	AnchorCenterRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
)

// Point returns the EMU coordinate of the anchor on the given box.
func (a Anchor) Point(b Box) Position {
	var x, y EMU
	switch a {
	case AnchorTopLeft, AnchorCenterLeft, AnchorBottomLeft:
		x = b.X
	case AnchorTopCenter, AnchorCenter, AnchorBottomCenter:
		x = b.X + b.W/2
	case AnchorTopRight, AnchorCenterRight, AnchorBottomRight:
		x = b.X + b.W
	}
	switch a {
	case AnchorTopLeft, AnchorTopCenter, AnchorTopRight:
		y = b.Y
	case AnchorCenterLeft, AnchorCenter, AnchorCenterRight:
		y = b.Y + b.H/2
	case AnchorBottomLeft, AnchorBottomCenter, AnchorBottomRight:
		y = b.Y + b.H
	}
	return Position{X: x, Y: y}
}
