package render

// SVG single-path → OOXML custom geometry translator (RFC §14.1, brief 03).
//
// Translate accepts a curated/caller icon SVG and emits an <a:custGeom> path,
// or an error if the SVG falls outside the documented subset:
//
//   - exactly one <path> element (no circle/rect/line/polyline/polygon/ellipse —
//     author those as a path),
//   - a solid fill (not fill="none", not a url(#…) gradient/pattern reference),
//   - path commands M L H V C S Q T Z (absolute or relative); S/T expand to
//     cubic/quadratic by reflecting the previous control point,
//   - NO elliptical arcs (A/a) — author curves with Béziers.
//
// The SVG fill *color* is discarded: an icon renders filled with the active
// accent token at AddIcon time (P2). Coordinates map from the viewBox into the
// path's own w×h space (×coordScale, rounded to integers); SVG and DrawingML
// share a top-left, y-down origin, so there is no axis flip. The translator is
// pure and deterministic.

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// coordScale multiplies viewBox user units to give integer path coordinates
// with sub-unit precision (a 24-unit viewBox becomes a 2400-unit path grid).
const coordScale = 100

// Translate converts an icon SVG to a custom geometry, or returns an error if it
// violates the documented subset.
func Translate(svg []byte) (*slide.XCustomGeometry, error) {
	vb, d, err := parseSVG(svg)
	if err != nil {
		return nil, err
	}
	raw, err := tokenizePath(d)
	if err != nil {
		return nil, err
	}
	return buildGeometry(raw, vb)
}

// viewBox is the SVG coordinate window (minX minY width height).
type viewBox struct {
	minX, minY, w, h float64
}

// parseSVG walks the SVG, validating the single-solid-path constraint, and
// returns the viewBox and the path's d data.
func parseSVG(svg []byte) (viewBox, string, error) {
	dec := xml.NewDecoder(strings.NewReader(string(svg)))
	var (
		vb        viewBox
		haveVB    bool
		widthAtt  string
		heightAtt string
		pathD     string
		pathFill  string
		nPaths    int
	)
	for {
		tok, err := dec.Token()
		if err != nil {
			break // io.EOF (or malformed; checked below by nPaths)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		switch se.Name.Local {
		case "svg":
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "viewBox":
					if box, ok := parseViewBox(a.Value); ok {
						vb, haveVB = box, true
					}
				case "width":
					widthAtt = a.Value
				case "height":
					heightAtt = a.Value
				}
			}
		case "path":
			nPaths++
			for _, a := range se.Attr {
				switch a.Name.Local {
				case "d":
					pathD = a.Value
				case "fill":
					pathFill = a.Value
				}
			}
		case "circle", "rect", "line", "polyline", "polygon", "ellipse":
			return viewBox{}, "", fmt.Errorf("svg: unsupported element <%s> (author the icon as a single <path>)", se.Name.Local)
		case "linearGradient", "radialGradient", "pattern":
			return viewBox{}, "", fmt.Errorf("svg: gradient/pattern fills are not supported (solid fill only)")
		}
	}
	if nPaths == 0 {
		return viewBox{}, "", fmt.Errorf("svg: no <path> element found")
	}
	if nPaths > 1 {
		return viewBox{}, "", fmt.Errorf("svg: %d paths found, exactly one is supported", nPaths)
	}
	if strings.TrimSpace(pathD) == "" {
		return viewBox{}, "", fmt.Errorf("svg: <path> has no d data")
	}
	if f := strings.TrimSpace(strings.ToLower(pathFill)); f == "none" {
		return viewBox{}, "", fmt.Errorf("svg: fill=\"none\" is not supported (the path must be filled)")
	} else if strings.HasPrefix(f, "url(") {
		return viewBox{}, "", fmt.Errorf("svg: gradient/pattern fill reference is not supported (solid fill only)")
	}
	if !haveVB {
		w, werr := strconv.ParseFloat(strings.TrimSpace(widthAtt), 64)
		h, herr := strconv.ParseFloat(strings.TrimSpace(heightAtt), 64)
		if werr != nil || herr != nil || w <= 0 || h <= 0 {
			return viewBox{}, "", fmt.Errorf("svg: missing or invalid viewBox (and no usable width/height)")
		}
		vb = viewBox{0, 0, w, h}
	}
	if vb.w <= 0 || vb.h <= 0 {
		return viewBox{}, "", fmt.Errorf("svg: viewBox has non-positive size")
	}
	return vb, pathD, nil
}

func parseViewBox(s string) (viewBox, bool) {
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == '\t' || r == '\n' })
	if len(fields) != 4 {
		return viewBox{}, false
	}
	var v [4]float64
	for i, f := range fields {
		n, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return viewBox{}, false
		}
		v[i] = n
	}
	if v[2] <= 0 || v[3] <= 0 {
		return viewBox{}, false
	}
	return viewBox{minX: v[0], minY: v[1], w: v[2], h: v[3]}, true
}

// rawCmd is one parsed SVG path command: its operator letter and numeric args.
type rawCmd struct {
	op   byte
	args []float64
}

// tokenizePath splits a d string into operator+args commands. It rejects an
// elliptical-arc command and any unrecognized operator.
func tokenizePath(d string) ([]rawCmd, error) {
	var cmds []rawCmd
	i, n := 0, len(d)
	for i < n {
		c := d[i]
		switch {
		case c == ' ' || c == ',' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == 'A' || c == 'a':
			return nil, fmt.Errorf("svg: elliptical-arc command %q is not supported (use Béziers)", string(c))
		case strings.IndexByte("MmLlHhVvCcSsQqTtZz", c) >= 0:
			args, next, err := scanNumbers(d, i+1)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, rawCmd{op: c, args: args})
			i = next
		default:
			return nil, fmt.Errorf("svg: unexpected character %q in path data", string(c))
		}
	}
	return cmds, nil
}

// scanNumbers reads whitespace/comma-separated numbers starting at i until the
// next command letter or end of string.
func scanNumbers(d string, i int) ([]float64, int, error) {
	var nums []float64
	n := len(d)
	for i < n {
		c := d[i]
		if c == ' ' || c == ',' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			break // a command letter (or an unknown one — the tokenizer reports it)
		}
		// number: optional sign, digits, decimal, exponent.
		start := i
		if c == '+' || c == '-' {
			i++
		}
		seenDigit := false
		for i < n && d[i] >= '0' && d[i] <= '9' {
			i++
			seenDigit = true
		}
		if i < n && d[i] == '.' {
			i++
			for i < n && d[i] >= '0' && d[i] <= '9' {
				i++
				seenDigit = true
			}
		}
		if i < n && (d[i] == 'e' || d[i] == 'E') {
			i++
			if i < n && (d[i] == '+' || d[i] == '-') {
				i++
			}
			for i < n && d[i] >= '0' && d[i] <= '9' {
				i++
			}
		}
		if !seenDigit {
			return nil, 0, fmt.Errorf("svg: malformed number in path data at offset %d", start)
		}
		f, err := strconv.ParseFloat(d[start:i], 64)
		if err != nil {
			return nil, 0, fmt.Errorf("svg: malformed number %q: %w", d[start:i], err)
		}
		nums = append(nums, f)
	}
	return nums, i, nil
}

// builder threads coordinate state through command processing.
type builder struct {
	vb           viewBox
	cx, cy       float64 // current point (user units)
	sx, sy       float64 // current subpath start (for Z)
	lastCtrlX    float64 // previous control point (for S/T reflection)
	lastCtrlY    float64
	lastWasCubic bool
	lastWasQuad  bool
	out          []slide.XPathCommand
}

// buildGeometry converts raw commands to an absolute-coordinate custGeom path.
func buildGeometry(raw []rawCmd, vb viewBox) (*slide.XCustomGeometry, error) {
	b := &builder{vb: vb}
	for _, c := range raw {
		if err := b.apply(c); err != nil {
			return nil, err
		}
	}
	if len(b.out) == 0 {
		return nil, fmt.Errorf("svg: path produced no drawing commands")
	}
	return &slide.XCustomGeometry{
		AvLst: &slide.XAvLst{},
		GdLst: &slide.XGdLst{},
		PathList: slide.XPathList{Paths: []slide.XPath{{
			W:        int64(math.Round(vb.w * coordScale)),
			H:        int64(math.Round(vb.h * coordScale)),
			Commands: b.out,
		}}},
	}, nil
}

// pt maps a user-space point to a scaled integer path point.
func (b *builder) pt(x, y float64) slide.XPoint {
	return slide.XPoint{
		X: int64(math.Round((x - b.vb.minX) * coordScale)),
		Y: int64(math.Round((y - b.vb.minY) * coordScale)),
	}
}

// emit appends a command and records control-point/current state.
func (b *builder) emit(cmd string, ptsUser ...[2]float64) {
	pts := make([]slide.XPoint, len(ptsUser))
	for i, p := range ptsUser {
		pts[i] = b.pt(p[0], p[1])
	}
	b.out = append(b.out, slide.XPathCommand{Cmd: cmd, Pts: pts})
}

// apply processes one raw command, expanding implicit repeats and S/T smoothing.
func (b *builder) apply(c rawCmd) error {
	rel := c.op >= 'a' && c.op <= 'z'
	switch up := upper(c.op); up {
	case 'M':
		if len(c.args) < 2 || len(c.args)%2 != 0 {
			return fmt.Errorf("svg: M expects pairs of coordinates, got %d", len(c.args))
		}
		for i := 0; i < len(c.args); i += 2 {
			x, y := b.abs(rel, c.args[i], c.args[i+1])
			if i == 0 {
				b.cx, b.cy = x, y
				b.sx, b.sy = x, y
				b.emit(slide.PathMoveTo, [2]float64{x, y})
			} else { // implicit lineto
				b.cx, b.cy = x, y
				b.emit(slide.PathLnTo, [2]float64{x, y})
			}
			b.clearCtrl()
		}
	case 'L':
		if len(c.args) < 2 || len(c.args)%2 != 0 {
			return fmt.Errorf("svg: L expects pairs of coordinates, got %d", len(c.args))
		}
		for i := 0; i < len(c.args); i += 2 {
			x, y := b.abs(rel, c.args[i], c.args[i+1])
			b.cx, b.cy = x, y
			b.emit(slide.PathLnTo, [2]float64{x, y})
			b.clearCtrl()
		}
	case 'H':
		for _, a := range c.args {
			x := a
			if rel {
				x += b.cx
			}
			b.cx = x
			b.emit(slide.PathLnTo, [2]float64{x, b.cy})
			b.clearCtrl()
		}
	case 'V':
		for _, a := range c.args {
			y := a
			if rel {
				y += b.cy
			}
			b.cy = y
			b.emit(slide.PathLnTo, [2]float64{b.cx, y})
			b.clearCtrl()
		}
	case 'C':
		if len(c.args) < 6 || len(c.args)%6 != 0 {
			return fmt.Errorf("svg: C expects sextuples, got %d", len(c.args))
		}
		for i := 0; i < len(c.args); i += 6 {
			x1, y1 := b.abs(rel, c.args[i], c.args[i+1])
			x2, y2 := b.abs(rel, c.args[i+2], c.args[i+3])
			x, y := b.abs(rel, c.args[i+4], c.args[i+5])
			b.emit(slide.PathCubicTo, [2]float64{x1, y1}, [2]float64{x2, y2}, [2]float64{x, y})
			b.setCubicCtrl(x2, y2, x, y)
		}
	case 'S':
		if len(c.args) < 4 || len(c.args)%4 != 0 {
			return fmt.Errorf("svg: S expects quadruples, got %d", len(c.args))
		}
		for i := 0; i < len(c.args); i += 4 {
			x1, y1 := b.reflectCubic()
			x2, y2 := b.abs(rel, c.args[i], c.args[i+1])
			x, y := b.abs(rel, c.args[i+2], c.args[i+3])
			b.emit(slide.PathCubicTo, [2]float64{x1, y1}, [2]float64{x2, y2}, [2]float64{x, y})
			b.setCubicCtrl(x2, y2, x, y)
		}
	case 'Q':
		if len(c.args) < 4 || len(c.args)%4 != 0 {
			return fmt.Errorf("svg: Q expects quadruples, got %d", len(c.args))
		}
		for i := 0; i < len(c.args); i += 4 {
			x1, y1 := b.abs(rel, c.args[i], c.args[i+1])
			x, y := b.abs(rel, c.args[i+2], c.args[i+3])
			b.emit(slide.PathQuadTo, [2]float64{x1, y1}, [2]float64{x, y})
			b.setQuadCtrl(x1, y1, x, y)
		}
	case 'T':
		if len(c.args) < 2 || len(c.args)%2 != 0 {
			return fmt.Errorf("svg: T expects pairs of coordinates, got %d", len(c.args))
		}
		for i := 0; i < len(c.args); i += 2 {
			x1, y1 := b.reflectQuad()
			x, y := b.abs(rel, c.args[i], c.args[i+1])
			b.emit(slide.PathQuadTo, [2]float64{x1, y1}, [2]float64{x, y})
			b.setQuadCtrl(x1, y1, x, y)
		}
	case 'Z':
		b.cx, b.cy = b.sx, b.sy
		b.emit(slide.PathClose)
		b.clearCtrl()
	default:
		return fmt.Errorf("svg: unsupported path command %q", string(c.op))
	}
	return nil
}

func (b *builder) abs(rel bool, x, y float64) (float64, float64) {
	if rel {
		return b.cx + x, b.cy + y
	}
	return x, y
}

func (b *builder) clearCtrl() {
	b.lastWasCubic, b.lastWasQuad = false, false
}

func (b *builder) setCubicCtrl(x2, y2, x, y float64) {
	b.cx, b.cy = x, y
	b.lastCtrlX, b.lastCtrlY = x2, y2
	b.lastWasCubic, b.lastWasQuad = true, false
}

func (b *builder) setQuadCtrl(x1, y1, x, y float64) {
	b.cx, b.cy = x, y
	b.lastCtrlX, b.lastCtrlY = x1, y1
	b.lastWasCubic, b.lastWasQuad = false, true
}

// reflectCubic returns the first control point for a smooth cubic (S): the
// reflection of the previous cubic's second control about the current point, or
// the current point if the previous command was not a cubic.
func (b *builder) reflectCubic() (float64, float64) {
	if b.lastWasCubic {
		return 2*b.cx - b.lastCtrlX, 2*b.cy - b.lastCtrlY
	}
	return b.cx, b.cy
}

// reflectQuad returns the control point for a smooth quad (T).
func (b *builder) reflectQuad() (float64, float64) {
	if b.lastWasQuad {
		return 2*b.cx - b.lastCtrlX, 2*b.cy - b.lastCtrlY
	}
	return b.cx, b.cy
}

func upper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}
