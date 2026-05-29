package pptx_test

import (
	"testing"

	slidex "github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// Test component definitions
// ============================================================================

// TextComponent is a simple text component used in tests.
type TextComponent struct {
	x, y, cx, cy int
	text         string
}

func NewTextComponent(x, y, cx, cy int, text string) *TextComponent {
	return &TextComponent{
		x:    x,
		y:    y,
		cx:   cx,
		cy:   cy,
		text: text,
	}
}

func (t *TextComponent) Render(ctx *pptx.SlideContext) error {
	// Allocate a shape ID.
	id := ctx.NextShapeID()

	// Build shape XML.
	sp := &slidex.XSp{
		NonVisual: slidex.XNonVisualDrawingShape{
			CNvPr: &slidex.XNvCxnSpPr{
				ID:   int(id),
				Name: "TextBox",
			},
			CNvSpPr: &slidex.XNvSpPr{},
		},
		ShapeProperties: &slidex.XShapeProperties{
			Transform2D: &slidex.XTransform2D{
				Offset: &slidex.XOv2DrOffset{X: t.x, Y: t.y},
				Extent: &slidex.XOv2DrExtent{Cx: t.cx, Cy: t.cy},
			},
		},
		TextBody: &slidex.XTextBody{
			BodyPr:   &slidex.XBodyPr{},
			LstStyle: &slidex.XTextParagraphList{},
			Paragraphs: []slidex.XTextParagraph{
				{Content: []any{&slidex.XTextRun{Text: t.text}}},
			},
		},
	}

	// Mount onto the slide.
	ctx.AppendShape(sp)
	return nil
}

// RectangleComponent is a simple rectangle component used in tests.
type RectangleComponent struct {
	x, y, cx, cy int
}

func NewRectangleComponent(x, y, cx, cy int) *RectangleComponent {
	return &RectangleComponent{x: x, y: y, cx: cx, cy: cy}
}

func (r *RectangleComponent) Render(ctx *pptx.SlideContext) error {
	id := ctx.NextShapeID()

	sp := &slidex.XSp{
		NonVisual: slidex.XNonVisualDrawingShape{
			CNvPr: &slidex.XNvCxnSpPr{
				ID:   int(id),
				Name: "Rectangle",
			},
			CNvSpPr: &slidex.XNvSpPr{},
		},
		ShapeProperties: &slidex.XShapeProperties{
			Transform2D: &slidex.XTransform2D{
				Offset: &slidex.XOv2DrOffset{X: r.x, Y: r.y},
				Extent: &slidex.XOv2DrExtent{Cx: r.cx, Cy: r.cy},
			},
			PresetGeom: &slidex.XPresetGeometry{Prst: "rect", AvLst: &slidex.XAvLst{}},
		},
	}

	ctx.AppendShape(sp)
	return nil
}

// CompositeTestComponent composes multiple child components.
type CompositeTestComponent struct {
	children []pptx.Component
}

func NewCompositeTestComponent(children ...pptx.Component) *CompositeTestComponent {
	return &CompositeTestComponent{children: children}
}

func (c *CompositeTestComponent) Render(ctx *pptx.SlideContext) error {
	for i, child := range c.children {
		if err := child.Render(ctx); err != nil {
			return err
		}
		_ = i // suppress unused-variable warning
	}
	return nil
}

// ============================================================================
// Test cases
// ============================================================================

// TestAddComponent tests adding a single component to a slide.
func TestAddComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Add text component.
	textComp := NewTextComponent(100, 100, 500, 50, "Hello Component")
	err := slide.AddComponent(textComp)
	if err != nil {
		t.Fatalf("AddComponent failed: %v", err)
	}

	// Verify shape ID was allocated.
	ctx := slide.NewContext()
	if ctx.CurrentShapeID() < 1 {
		t.Error("shape ID was not allocated correctly")
	}
}

// TestAddComponents tests adding multiple components at once.
func TestAddComponents(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Add components in bulk.
	err := slide.AddComponents(
		NewTextComponent(100, 100, 500, 50, "Title"),
		NewRectangleComponent(100, 200, 300, 150),
		NewTextComponent(100, 400, 500, 30, "Footer"),
	)
	if err != nil {
		t.Fatalf("AddComponents failed: %v", err)
	}
}

// TestCompositeComponent tests a composite component.
func TestCompositeComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Create composite component.
	composite := NewCompositeTestComponent(
		NewTextComponent(100, 100, 500, 50, "Header"),
		NewRectangleComponent(100, 200, 300, 150),
	)

	err := slide.AddComponent(composite)
	if err != nil {
		t.Fatalf("AddComponent(composite) failed: %v", err)
	}
}

// TestFuncComponent tests the functional component adapter.
func TestFuncComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Use a functional component.
	renderCount := 0
	funcComp := pptx.FuncComponent(func(ctx *pptx.SlideContext) error {
		renderCount++
		id := ctx.NextShapeID()

		sp := &slidex.XSp{
			NonVisual: slidex.XNonVisualDrawingShape{
				CNvPr: &slidex.XNvCxnSpPr{
					ID:   int(id),
					Name: "FuncComponent",
				},
				CNvSpPr: &slidex.XNvSpPr{},
			},
			ShapeProperties: &slidex.XShapeProperties{
				Transform2D: &slidex.XTransform2D{
					Offset: &slidex.XOv2DrOffset{X: 100, Y: 100},
					Extent: &slidex.XOv2DrExtent{Cx: 200, Cy: 100},
				},
			},
		}

		ctx.AppendShape(sp)
		return nil
	})

	err := slide.AddComponent(funcComp)
	if err != nil {
		t.Fatalf("AddComponent(functional component) failed: %v", err)
	}

	if renderCount != 1 {
		t.Errorf("expected renderCount=1, got=%d", renderCount)
	}
}

// TestSlideContextShapeID tests shape ID allocation via SlideContext.
func TestSlideContextShapeID(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()
	ctx := slide.NewContext()

	// Allocate several shape IDs.
	id1 := ctx.NextShapeID()
	id2 := ctx.NextShapeID()
	id3 := ctx.NextShapeID()

	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Error("duplicate shape IDs allocated")
	}

	// Verify monotonic increase.
	if id2 <= id1 || id3 <= id2 {
		t.Error("shape IDs are not monotonically increasing")
	}
}

// TestSlideContextSlideSize tests retrieving slide dimensions via SlideContext.
func TestSlideContextSlideSize(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()
	ctx := slide.NewContext()

	cx, cy := ctx.SlideSize()
	if cx == 0 || cy == 0 {
		t.Error("invalid slide size")
	}

	// Log if the size differs from the expected default (16:9 widescreen).
	if cx != 12192000 || cy != 6858000 {
		t.Logf("slide size: cx=%d, cy=%d (not the default 16:9)", cx, cy)
	}
}

// TestSlideContextUnitConversion tests unit conversion helpers on SlideContext.
func TestSlideContextUnitConversion(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()
	ctx := slide.NewContext()

	// Test px -> EMU (96 DPI: 1 px = 9525 EMU).
	px := 96
	emu := ctx.PxToEMU(px)
	if emu != 914400 { // 96 * 9525 = 914400 (= 1 inch)
		t.Errorf("PxToEMU(96) = %d, expected 914400", emu)
	}

	// Test EMU -> px conversion.
	backPx := ctx.EMUToPx(914400)
	if backPx != 96 {
		t.Errorf("EMUToPx(914400) = %d, expected 96", backPx)
	}

	// Test 1 px = 9525 EMU.
	emu = ctx.PxToEMU(1)
	if emu != 9525 {
		t.Errorf("PxToEMU(1) = %d, expected 9525", emu)
	}
}

// TestComponentWithSave tests saving a presentation that contains components.
func TestComponentWithSave(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Add multiple components.
	slide.AddComponent(NewTextComponent(100, 100, 500, 50, "Title"))
	slide.AddComponent(NewRectangleComponent(100, 200, 300, 150))
	slide.AddComponent(NewTextComponent(100, 400, 500, 30, "Footer"))

	// Save to bytes.
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("output data is empty")
	}

	t.Logf("generated PPTX size: %d bytes", len(data))
}

// TestBuiltInCompositeComponent tests the built-in CompositeComponent helper.
func TestBuiltInCompositeComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Use the built-in CompositeComponent.
	composite := pptx.NewCompositeComponent("test-composite",
		NewTextComponent(100, 100, 500, 50, "Header"),
		NewRectangleComponent(100, 200, 300, 150),
	)

	err := slide.AddComponent(composite)
	if err != nil {
		t.Fatalf("AddComponent(CompositeComponent) failed: %v", err)
	}

	if composite.Name() != "test-composite" {
		t.Errorf("expected name 'test-composite', got '%s'", composite.Name())
	}
}

// TestConditionalComponent tests a conditional component.
func TestConditionalComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	condition := true

	// Create conditional component.
	condComp := pptx.NewConditionalComponent(
		func() bool { return condition },
		NewTextComponent(100, 100, 500, 50, "Condition is true"),
		NewTextComponent(100, 100, 500, 50, "Condition is false"),
	)

	err := slide.AddComponent(condComp)
	if err != nil {
		t.Fatalf("AddComponent(ConditionalComponent) failed: %v", err)
	}
}

// TestRepeatedComponent tests a repeated component.
func TestRepeatedComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Create a repeated component (generates 3 text boxes).
	repeatedComp := pptx.NewRepeatedComponent(3, func(index int) pptx.Component {
		return NewTextComponent(100, 100+index*60, 500, 50, "Item")
	})

	err := slide.AddComponent(repeatedComp)
	if err != nil {
		t.Fatalf("AddComponent(RepeatedComponent) failed: %v", err)
	}
}

// TestShapeComponent tests the ShapeComponent wrapper.
func TestShapeComponent(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Create a shape component.
	sp := &slidex.XSp{
		NonVisual: slidex.XNonVisualDrawingShape{
			CNvPr: &slidex.XNvCxnSpPr{
				ID:   1,
				Name: "TestShape",
			},
		},
		ShapeProperties: &slidex.XShapeProperties{
			Transform2D: &slidex.XTransform2D{
				Offset: &slidex.XOv2DrOffset{X: 100, Y: 100},
				Extent: &slidex.XOv2DrExtent{Cx: 200, Cy: 100},
			},
		},
	}

	shapeComp := pptx.NewShapeComponent(sp, 100, 100)
	shapeComp.SetName("TestShapeComponent")

	err := slide.AddComponent(shapeComp)
	if err != nil {
		t.Fatalf("AddComponent(ShapeComponent) failed: %v", err)
	}

	if shapeComp.Name() != "TestShapeComponent" {
		t.Errorf("expected name 'TestShapeComponent', got '%s'", shapeComp.Name())
	}
}

// TestComponentList tests the ComponentList helper.
func TestComponentList(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Build a component list.
	var list pptx.ComponentList
	list.Add(NewTextComponent(100, 100, 500, 50, "Item 1"))
	list.Add(NewTextComponent(100, 160, 500, 50, "Item 2"))
	list.Add(NewTextComponent(100, 220, 500, 50, "Item 3"))

	if list.Count() != 3 {
		t.Errorf("expected 3 components, got %d", list.Count())
	}

	// Render all components via context.
	ctx := slide.NewContext()
	err := list.RenderAll(ctx)
	if err != nil {
		t.Fatalf("RenderAll failed: %v", err)
	}
}
