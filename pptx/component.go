// Package pptx provides a high-level API for authoring PPTX files.
package pptx

import "github.com/hurtener/pptx-go/internal/ooxml/slide"

// ============================================================================
// Component interface — the building-block abstraction
// ============================================================================
//
// Component is the unified interface for anything that can be rendered onto a
// slide. Any type that implements this interface can be used as a "building
// block" added to a slide.
//
// Design principles:
// 1. Single responsibility — a component only generates its own XML structure.
// 2. Dependency injection — resources (IDs, media, etc.) are obtained via
//    SlideContext.
// 3. Composability — components can be combined to form more complex components.
//
// Usage example:
//
//	type TitleComponent struct {
//		text string
//	}
//
//	func (t *TitleComponent) Render(ctx *SlideContext) error {
//		id := ctx.NextShapeID()
//		sp := &slide.XSp{...}
//		ctx.AppendShape(sp)
//		return nil
//	}
//
//	slide.AddComponent(&TitleComponent{text: "Hello"})
//
// ============================================================================

// Component is the interface that all renderable slide building blocks must
// implement.
type Component interface {
	// Render writes the component's shapes into the slide via ctx.
	// An error indicates a rendering failure.
	Render(ctx *SlideContext) error
}

// ============================================================================
// Extended component interfaces
// ============================================================================

// ComponentWithSize is a component that exposes its bounding box.
type ComponentWithSize interface {
	Component
	// Bounds returns the bounding box (x, y, cx, cy in EMU).
	Bounds() (x, y, cx, cy int)
}

// ComponentWithName is a component that has a human-readable name (useful for
// debugging and logging).
type ComponentWithName interface {
	Component
	// Name returns the component's name.
	Name() string
}

// ComponentWithPosition is a component whose position can be read and set.
type ComponentWithPosition interface {
	Component
	// SetPosition sets the component's position in EMU.
	SetPosition(x, y int)
	// Position returns the component's current position in EMU.
	Position() (x, y int)
}

// ComponentWithSizeSetter is a component whose size can be read and set.
type ComponentWithSizeSetter interface {
	Component
	// SetSize sets the component's size in EMU.
	SetSize(cx, cy int)
	// Size returns the component's current size in EMU.
	Size() (cx, cy int)
}

// ============================================================================
// Component utilities
// ============================================================================

// ComponentList is an ordered collection of components.
type ComponentList []Component

// Add appends a component to the list.
func (cl *ComponentList) Add(c Component) {
	*cl = append(*cl, c)
}

// RenderAll renders every component in order, stopping on the first error.
func (cl ComponentList) RenderAll(ctx *SlideContext) error {
	for i, c := range cl {
		if err := c.Render(ctx); err != nil {
			return &ComponentRenderError{
				Index:      i,
				Component:  c,
				Underlying: err,
			}
		}
	}
	return nil
}

// Count returns the number of components in the list.
func (cl ComponentList) Count() int {
	return len(cl)
}

// ============================================================================
// Error types
// ============================================================================

// ComponentRenderError is returned when a component fails to render.
type ComponentRenderError struct {
	Index      int
	Component  Component
	Underlying error
}

// Error implements the error interface.
func (e *ComponentRenderError) Error() string {
	name := "<unknown>"
	if n, ok := e.Component.(ComponentWithName); ok {
		name = n.Name()
	}
	return "component render error at index " + string(rune(e.Index)) + " (" + name + "): " + e.Underlying.Error()
}

// Unwrap returns the underlying error.
func (e *ComponentRenderError) Unwrap() error {
	return e.Underlying
}

// ============================================================================
// Built-in component implementations
// ============================================================================

// FuncComponent wraps a plain function as a Component.
type FuncComponent func(ctx *SlideContext) error

// Render implements Component.
func (fc FuncComponent) Render(ctx *SlideContext) error {
	return fc(ctx)
}

// ============================================================================
// Composite component
// ============================================================================

// CompositeComponent groups multiple components and renders them in order.
type CompositeComponent struct {
	components []Component
	name       string
}

// NewCompositeComponent creates a CompositeComponent with the given name and
// initial children.
func NewCompositeComponent(name string, components ...Component) *CompositeComponent {
	return &CompositeComponent{
		components: components,
		name:       name,
	}
}

// Add appends a child component.
func (cc *CompositeComponent) Add(c Component) {
	cc.components = append(cc.components, c)
}

// Render implements Component.
func (cc *CompositeComponent) Render(ctx *SlideContext) error {
	for i, c := range cc.components {
		if err := c.Render(ctx); err != nil {
			return &ComponentRenderError{
				Index:      i,
				Component:  c,
				Underlying: err,
			}
		}
	}
	return nil
}

// Name implements ComponentWithName.
func (cc *CompositeComponent) Name() string {
	return cc.name
}

// Components returns all child components.
func (cc *CompositeComponent) Components() []Component {
	return cc.components
}

// ============================================================================
// Conditional component
// ============================================================================

// ConditionalComponent renders one of two components depending on a runtime
// condition.
type ConditionalComponent struct {
	condition     func() bool
	component     Component
	elseComponent Component
}

// NewConditionalComponent creates a ConditionalComponent. ifComponent is
// rendered when condition returns true; elseComponent is rendered otherwise.
// Either may be nil.
func NewConditionalComponent(condition func() bool, ifComponent, elseComponent Component) *ConditionalComponent {
	return &ConditionalComponent{
		condition:     condition,
		component:     ifComponent,
		elseComponent: elseComponent,
	}
}

// Render implements Component.
func (cc *ConditionalComponent) Render(ctx *SlideContext) error {
	if cc.condition() {
		if cc.component != nil {
			return cc.component.Render(ctx)
		}
	} else {
		if cc.elseComponent != nil {
			return cc.elseComponent.Render(ctx)
		}
	}
	return nil
}

// ============================================================================
// Repeated component
// ============================================================================

// RepeatedComponent renders a template component once for each item in a
// data slice.
type RepeatedComponent struct {
	template func(index int) Component
	count    int
}

// NewRepeatedComponent creates a RepeatedComponent that calls template(i) for
// i in [0, count) and renders the result.
func NewRepeatedComponent(count int, template func(index int) Component) *RepeatedComponent {
	return &RepeatedComponent{
		template: template,
		count:    count,
	}
}

// Render implements Component.
func (rc *RepeatedComponent) Render(ctx *SlideContext) error {
	for i := 0; i < rc.count; i++ {
		c := rc.template(i)
		if c == nil {
			continue
		}
		if err := c.Render(ctx); err != nil {
			return &ComponentRenderError{
				Index:      i,
				Component:  c,
				Underlying: err,
			}
		}
	}
	return nil
}

// ============================================================================
// Shape component helper
// ============================================================================

// ShapeComponent is the simplest component type: it wraps a single XSp shape.
type ShapeComponent struct {
	sp   *slide.XSp
	x, y int
	name string
}

// NewShapeComponent creates a ShapeComponent at the given position (EMU).
func NewShapeComponent(sp *slide.XSp, x, y int) *ShapeComponent {
	return &ShapeComponent{
		sp: sp,
		x:  x,
		y:  y,
	}
}

// Render implements Component.
func (sc *ShapeComponent) Render(ctx *SlideContext) error {
	if sc.sp == nil {
		return nil
	}
	ctx.AppendShape(sc.sp)
	return nil
}

// Name implements ComponentWithName.
func (sc *ShapeComponent) Name() string {
	return sc.name
}

// SetName sets the component's name.
func (sc *ShapeComponent) SetName(name string) {
	sc.name = name
}

// Bounds implements ComponentWithSize.
func (sc *ShapeComponent) Bounds() (x, y, cx, cy int) {
	if sc.sp != nil && sc.sp.ShapeProperties != nil && sc.sp.ShapeProperties.Transform2D != nil {
		x = sc.sp.ShapeProperties.Transform2D.Offset.X
		y = sc.sp.ShapeProperties.Transform2D.Offset.Y
		cx = sc.sp.ShapeProperties.Transform2D.Extent.Cx
		cy = sc.sp.ShapeProperties.Transform2D.Extent.Cy
	}
	return
}

// SetPosition implements ComponentWithPosition.
func (sc *ShapeComponent) SetPosition(x, y int) {
	sc.x = x
	sc.y = y
	if sc.sp != nil && sc.sp.ShapeProperties != nil && sc.sp.ShapeProperties.Transform2D != nil {
		sc.sp.ShapeProperties.Transform2D.Offset.X = x
		sc.sp.ShapeProperties.Transform2D.Offset.Y = y
	}
}

// Position implements ComponentWithPosition.
func (sc *ShapeComponent) Position() (x, y int) {
	return sc.x, sc.y
}
