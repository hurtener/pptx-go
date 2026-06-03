package pptx_test

import (
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// reopenShapes authors a single slide via build, serializes it, reopens it with
// pptx.NewFromBytes, and returns the reopened slide's Shapes() — the
// author → save → Open round trip every Phase 18 read accessor is asserted over.
func reopenShapes(t *testing.T, build func(s *pptx.Slide)) []*pptx.Shape {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()
	build(s)
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	slides := re.Slides()
	if len(slides) != 1 {
		t.Fatalf("reopened deck has %d slides, want 1", len(slides))
	}
	return slides[0].Shapes()
}

// TestSlideShapes_EnumeratesInOrder is PR#1 acceptance criterion 1 (enumeration):
// Slide.Shapes() returns every authored shape on a reopened deck, in document
// order, recovering each shape's geometry and box.
func TestSlideShapes_EnumeratesInOrder(t *testing.T) {
	geoms := []pptx.ShapeGeometry{pptx.ShapeRect, pptx.ShapeEllipse, pptx.ShapeRoundRect}
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		for _, g := range geoms {
			s.AddShape(g, fxBox)
		}
	})
	if len(shapes) != len(geoms) {
		t.Fatalf("Shapes() returned %d shapes, want %d", len(shapes), len(geoms))
	}
	for i, sh := range shapes {
		if got := sh.Geometry(); got != geoms[i] {
			t.Errorf("shape %d geometry = %q, want %q", i, got, geoms[i])
		}
		if got := sh.Box(); got != fxBox {
			t.Errorf("shape %d box = %+v, want %+v", i, got, fxBox)
		}
	}
}

// TestShapeRead_GeometryRotation is PR#1 acceptance criterion 1 (geometry +
// rotation): a reopened shape's Geometry/Rotation equal what was authored.
func TestShapeRead_GeometryRotation(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeChevron, fxBox, pptx.WithRotation(45))
	})
	sh := shapes[0]
	if got := sh.Geometry(); got != pptx.ShapeChevron {
		t.Errorf("Geometry() = %q, want %q", got, pptx.ShapeChevron)
	}
	if got := sh.Rotation(); got != 45 {
		t.Errorf("Rotation() = %v, want 45", got)
	}
}

// TestShapeRead_SolidFill is PR#1 acceptance criterion 1 (fill): an opaque and a
// translucent solid fill both reopen field-equal to the authored Fill.
func TestShapeRead_SolidFill(t *testing.T) {
	tests := []struct {
		name string
		want pptx.Fill
	}{
		{"opaque", pptx.SolidFill(pptx.RGB("2563EB"))},
		{"alpha", pptx.SolidFill(pptx.RGBA(pptx.RGB("2563EB"), 50000))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shapes := reopenShapes(t, func(s *pptx.Slide) {
				s.AddShape(pptx.ShapeRect, fxBox, pptx.WithFill(tc.want))
			})
			got := shapes[0].Fill()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Fill() = %#v, want %#v", got, tc.want)
			}
			if got.Kind() != pptx.FillSolid {
				t.Errorf("Kind() = %v, want FillSolid", got.Kind())
			}
			if _, ok := got.SolidColor(); !ok {
				t.Error("SolidColor() ok = false, want true")
			}
		})
	}
}

// TestShapeRead_Gradient is PR#1 acceptance criterion 1 (fill): a linear and a
// radial gradient reopen field-equal to the authored Fill, with the stops,
// angle, and radial flag recovered via Gradient().
func TestShapeRead_Gradient(t *testing.T) {
	linear := pptx.LinearGradient(90,
		pptx.GradientStop{Pos: 0, Color: pptx.RGB("FF0000")},
		pptx.GradientStop{Pos: 1, Color: pptx.RGB("0000FF")},
	)
	radial := pptx.RadialGradient(
		pptx.GradientStop{Pos: 0, Color: pptx.RGB("FFFFFF")},
		pptx.GradientStop{Pos: 1, Color: pptx.RGBA(pptx.RGB("000000"), 0)},
	)
	tests := []struct {
		name   string
		want   pptx.Fill
		radial bool
		angle  float64
	}{
		{"linear", linear, false, 90},
		{"radial", radial, true, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shapes := reopenShapes(t, func(s *pptx.Slide) {
				s.AddShape(pptx.ShapeRect, fxBox, pptx.WithFill(tc.want))
			})
			got := shapes[0].Fill()
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Fill() = %#v, want %#v", got, tc.want)
			}
			if got.Kind() != pptx.FillGradient {
				t.Fatalf("Kind() = %v, want FillGradient", got.Kind())
			}
			g, ok := got.Gradient()
			if !ok {
				t.Fatal("Gradient() ok = false, want true")
			}
			if g.Radial != tc.radial {
				t.Errorf("Gradient().Radial = %v, want %v", g.Radial, tc.radial)
			}
			if g.Angle != tc.angle {
				t.Errorf("Gradient().Angle = %v, want %v", g.Angle, tc.angle)
			}
			if len(g.Stops) != 2 {
				t.Errorf("Gradient().Stops len = %d, want 2", len(g.Stops))
			}
		})
	}
}

// TestShapeRead_NoFill is PR#1 acceptance criterion 1 (fill): an explicit NoFill
// reopens as a FillNone fill.
func TestShapeRead_NoFill(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox, pptx.WithFill(pptx.NoFill()))
	})
	got := shapes[0].Fill()
	if got == nil {
		t.Fatal("Fill() = nil, want NoFill")
	}
	if got.Kind() != pptx.FillNone {
		t.Errorf("Kind() = %v, want FillNone", got.Kind())
	}
	if !reflect.DeepEqual(got, pptx.NoFill()) {
		t.Errorf("Fill() = %#v, want NoFill()", got)
	}
}

// TestShapeRead_FillUnset is PR#1 acceptance criterion 1 (fill): a shape with no
// fill option reopens with a nil Fill (it inherits its style fill).
func TestShapeRead_FillUnset(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox)
	})
	if got := shapes[0].Fill(); got != nil {
		t.Errorf("Fill() = %#v, want nil", got)
	}
}

// TestShapeRead_Line is PR#1 acceptance criterion 1 (line): a reopened shape's
// Line equals the authored width / color / dash.
func TestShapeRead_Line(t *testing.T) {
	want := pptx.Line{Width: pptx.Pt(2), Color: pptx.RGB("FF0000"), Dash: "dash"}
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox, pptx.WithLine(want))
	})
	if got := shapes[0].Line(); !reflect.DeepEqual(got, want) {
		t.Errorf("Line() = %#v, want %#v", got, want)
	}
}

// TestShapeRead_LineUnset is PR#1 acceptance criterion 1 (line): a shape with no
// outline reopens with a zero Line.
func TestShapeRead_LineUnset(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox)
	})
	if got := shapes[0].Line(); !reflect.DeepEqual(got, pptx.Line{}) {
		t.Errorf("Line() = %#v, want zero Line", got)
	}
}

// TestShapeRead_Shadow is PR#1 acceptance criterion 1 (shadow): a reopened
// shape's Shadow equals the authored Elevation. A vertical offset round-trips
// exactly through the polar dist/dir storage.
func TestShapeRead_Shadow(t *testing.T) {
	want := pptx.Elevation{Blur: pptx.Pt(12), OffsetY: pptx.Pt(4), Color: "000000", Alpha: 35000}
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRoundRect, fxBox, pptx.WithShadow(want))
	})
	got, ok := shapes[0].Shadow()
	if !ok {
		t.Fatal("Shadow() ok = false, want true")
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Shadow() = %#v, want %#v", got, want)
	}
}

// TestShapeRead_ShadowUnset is PR#1 acceptance criterion 1 (shadow): a shape
// with no shadow reports ok=false.
func TestShapeRead_ShadowUnset(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox)
	})
	if _, ok := shapes[0].Shadow(); ok {
		t.Error("Shadow() ok = true, want false for a shadowless shape")
	}
}
