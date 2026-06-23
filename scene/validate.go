package scene

import (
	"errors"
	"fmt"
)

// Stage 1 validation (RFC §10.4): structural correctness — the union is
// well-formed and field-level constraints hold. Stage 2 (token + asset
// resolution against the active theme/resolver) runs at render time in a later
// phase. Stage 1 returns a joined error so a caller sees every problem at once.

// ValidateScene runs Stage 1 structural validation over a scene.
func ValidateScene(s Scene) error {
	var errs []error
	for i := range s.Slides {
		sl := &s.Slides[i]
		where := sl.ID
		if where == "" {
			where = fmt.Sprintf("#%d", i)
		}
		for j, n := range sl.Nodes {
			if n == nil {
				errs = append(errs, fmt.Errorf("slide %s node %d: nil node", where, j))
				continue
			}
			if err := validateNode(n); err != nil {
				errs = append(errs, fmt.Errorf("slide %s node %d (%s): %w", where, j, n.NodeKind(), err))
			}
		}
	}
	return errors.Join(errs...)
}

// validateNode checks one node's structural constraints, recursing into
// container children.
func validateNode(n SlideNode) error {
	switch v := n.(type) {
	case Heading:
		if v.Level < 1 || v.Level > 6 {
			return fmt.Errorf("heading level %d out of range 1..6", v.Level)
		}
	case List:
		if v.Kind < ListBullet || v.Kind > ListChecklist {
			return fmt.Errorf("invalid list kind %d", v.Kind)
		}
		if len(v.Items) == 0 {
			return errors.New("list has no items")
		}
	case Callout:
		if v.Kind < CalloutNote || v.Kind > CalloutImportant {
			return fmt.Errorf("invalid callout kind %d", v.Kind)
		}
	case Stat:
		if v.Value == "" {
			return errors.New("stat requires a value")
		}
	case Button:
		if v.Label == "" {
			return errors.New("button requires a label")
		}
	case Checklist:
		if len(v.Items) == 0 {
			return errors.New("checklist has no items")
		}
		if v.Columns < 0 || v.Columns > 3 {
			return fmt.Errorf("checklist columns %d out of range 0..3", v.Columns)
		}
		for i, it := range v.Items {
			if it.State < CheckDone || it.State > CheckNeutral {
				return fmt.Errorf("checklist item %d has invalid state %d", i, it.State)
			}
		}
	case ChipRow:
		if len(v.Chips) == 0 {
			return errors.New("chip_row has no chips")
		}
		for i, c := range v.Chips {
			if c.Tone < ChipTint || c.Tone > ChipOutline {
				return fmt.Errorf("chip_row chip %d has invalid tone %d", i, c.Tone)
			}
		}
	case Banner:
		return validateChildren(v.Trailing)
	case Image:
		if v.AssetID == "" {
			return errors.New("image requires an asset id")
		}
		if err := validateCrop(v.Crop); err != nil {
			return err
		}
	case Chart:
		if v.AssetID == "" {
			return errors.New("chart requires an asset id")
		}
	case CodeBlock:
		if v.AssetID == "" {
			return errors.New("code_block requires an asset id")
		}
	case Decoration:
		switch v.Kind {
		case DecorationPreset:
			if v.Preset == "" {
				return errors.New("preset decoration requires a preset name")
			}
		case DecorationAsset:
			if v.AssetID == "" {
				return errors.New("asset decoration requires an asset id")
			}
		default:
			return fmt.Errorf("invalid decoration kind %d", v.Kind)
		}
		if v.Opacity < 0 || v.Opacity > 1 {
			return fmt.Errorf("decoration opacity %.3f out of range [0,1]", v.Opacity)
		}
	case Flow:
		if len(v.Steps) == 0 {
			return errors.New("flow has no steps")
		}
	case Table:
		width := len(v.Headers)
		if width == 0 {
			return errors.New("table has no header columns")
		}
		for r, row := range v.Rows {
			if len(row) != width {
				return fmt.Errorf("table row %d has %d cells, want %d (header width)", r, len(row), width)
			}
		}
	case TwoColumn:
		if len(v.Left) == 0 || len(v.Right) == 0 {
			return errors.New("two_column requires non-empty left and right")
		}
		return validateChildren(append(append([]SlideNode{}, v.Left...), v.Right...))
	case Grid:
		if v.Columns < 2 || v.Columns > 4 {
			return fmt.Errorf("grid columns %d out of range 2..4", v.Columns)
		}
		if len(v.Ratio) != 0 && len(v.Ratio) != v.Columns {
			return fmt.Errorf("grid ratio length %d does not match columns %d", len(v.Ratio), v.Columns)
		}
		if len(v.Cells) == 0 {
			return errors.New("grid has no cells")
		}
		if len(v.Cells)%v.Columns != 0 {
			return fmt.Errorf("grid cell count %d is not a multiple of columns %d (a partial last row)", len(v.Cells), v.Columns)
		}
		return validateChildren(v.Cells)
	case Card:
		return validateChildren(v.Body)
	case CardSection:
		if len(v.Body) == 0 {
			return errors.New("card_section has no body")
		}
		return validateChildren(v.Body)
	case Bento:
		if v.Columns < 1 {
			return fmt.Errorf("bento columns %d must be >= 1", v.Columns)
		}
		if len(v.Rows) == 0 {
			return errors.New("bento has no rows")
		}
		for ri, row := range v.Rows {
			if len(row.Cells) == 0 {
				return fmt.Errorf("bento row %d has no cells", ri)
			}
			sum := 0
			for ci, cell := range row.Cells {
				if cell.Span < 1 {
					return fmt.Errorf("bento row %d cell %d span %d must be >= 1", ri, ci, cell.Span)
				}
				if cell.Node == nil {
					return fmt.Errorf("bento row %d cell %d: nil node", ri, ci)
				}
				sum += cell.Span
			}
			if sum > v.Columns {
				return fmt.Errorf("bento row %d spans sum to %d, exceeds columns %d", ri, sum, v.Columns)
			}
		}
		return validateChildren(v.cellNodes())
	}
	return nil
}

// validateCrop checks a crop is well-formed: each edge fraction is in [0,1] and
// opposite edges do not over-trim (Left+Right < 1, Top+Bottom < 1), so the
// composed source rectangle is non-degenerate. An over-crop is a structural
// error, not a silently clamped image (D-026/D-039).
func validateCrop(c Crop) error {
	for _, e := range []struct {
		name string
		v    float64
	}{{"left", c.Left}, {"top", c.Top}, {"right", c.Right}, {"bottom", c.Bottom}} {
		if e.v < 0 || e.v > 1 {
			return fmt.Errorf("image crop %s %.3f out of range [0,1]", e.name, e.v)
		}
	}
	if c.Left+c.Right >= 1 {
		return fmt.Errorf("image crop left+right %.3f >= 1 (no image remains)", c.Left+c.Right)
	}
	if c.Top+c.Bottom >= 1 {
		return fmt.Errorf("image crop top+bottom %.3f >= 1 (no image remains)", c.Top+c.Bottom)
	}
	return nil
}

// validateChildren validates each child node (and reports nil children).
func validateChildren(children []SlideNode) error {
	var errs []error
	for i, c := range children {
		if c == nil {
			errs = append(errs, fmt.Errorf("child %d: nil node", i))
			continue
		}
		if err := validateNode(c); err != nil {
			errs = append(errs, fmt.Errorf("child %d (%s): %w", i, c.NodeKind(), err))
		}
	}
	return errors.Join(errs...)
}
