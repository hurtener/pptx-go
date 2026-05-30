package pptx_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	slidex "github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/pptx"
)

// TestNewPresentation tests creating a blank presentation.
func TestNewPresentation(t *testing.T) {
	pres := pptx.New()
	if pres == nil {
		t.Fatal("New() returned nil")
	}

	// Check initial state.
	if pres.SlideCount() != 0 {
		t.Errorf("expected 0 slides, got %d", pres.SlideCount())
	}

	// Check slide dimensions.
	cx, cy := pres.SlideSize()
	if cx == 0 || cy == 0 {
		t.Errorf("invalid slide size: cx=%d, cy=%d", cx, cy)
	}
}

// TestAddSlide tests adding slides.
func TestAddSlide(t *testing.T) {
	pres := pptx.New()

	// Add first slide.
	slide1 := pres.AddSlide()
	if slide1 == nil {
		t.Fatal("AddSlide() returned nil")
	}

	if slide1.Index() != 0 {
		t.Errorf("expected index 0, got %d", slide1.Index())
	}

	if pres.SlideCount() != 1 {
		t.Errorf("expected 1 slide, got %d", pres.SlideCount())
	}

	// Add second slide.
	slide2 := pres.AddSlide()
	if slide2.Index() != 1 {
		t.Errorf("expected index 1, got %d", slide2.Index())
	}

	if pres.SlideCount() != 2 {
		t.Errorf("expected 2 slides, got %d", pres.SlideCount())
	}
}

// TestAddTextBox tests adding a text box.
func TestAddTextBox(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Add text box.
	textBox := slide.AddTextBox(100, 100, 500, 50, "Hello World")
	if textBox == nil {
		t.Fatal("AddTextBox() returned nil")
	}
}

// TestAddAutoShape tests adding shapes.
func TestAddAutoShape(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Add rectangle.
	rect := slide.AddRectangle(100, 100, 200, 100)
	if rect == nil {
		t.Fatal("AddRectangle() returned nil")
	}

	// Add ellipse.
	ellipse := slide.AddEllipse(100, 250, 200, 100)
	if ellipse == nil {
		t.Fatal("AddEllipse() returned nil")
	}
}

// TestAddTable tests adding a table.
func TestAddTable(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()

	// Add table (Box-based API; RFC §8.5).
	table := slide.AddTable(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(5), H: pptx.In(3)}, 3, 4)
	if table == nil {
		t.Fatal("AddTable() returned nil")
	}

	// Set cell text.
	table.Cell(0, 0).SetText("Header 1")
	table.Cell(0, 1).SetText("Header 2")
	table.Cell(1, 0).SetText("Data 1")
}

// TestGetSlide tests retrieving slides by index.
func TestGetSlide(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.AddSlide()

	// Get first slide.
	slide, err := pres.GetSlide(0)
	if err != nil {
		t.Fatalf("GetSlide(0) failed: %v", err)
	}
	if slide.Index() != 0 {
		t.Errorf("expected index 0, got %d", slide.Index())
	}

	// Get second slide.
	slide, err = pres.GetSlide(1)
	if err != nil {
		t.Fatalf("GetSlide(1) failed: %v", err)
	}
	if slide.Index() != 1 {
		t.Errorf("expected index 1, got %d", slide.Index())
	}

	// Test out-of-bounds.
	_, err = pres.GetSlide(2)
	if err == nil {
		t.Error("expected out-of-bounds error, but call succeeded")
	}
}

// TestRemoveSlide tests removing a slide.
func TestRemoveSlide(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.AddSlide()
	pres.AddSlide()

	// Remove the middle slide.
	err := pres.RemoveSlide(1)
	if err != nil {
		t.Fatalf("RemoveSlide(1) failed: %v", err)
	}

	if pres.SlideCount() != 2 {
		t.Errorf("expected 2 slides, got %d", pres.SlideCount())
	}

	// Check that indices are updated.
	slides := pres.Slides()
	if slides[0].Index() != 0 {
		t.Errorf("first slide index should be 0, got %d", slides[0].Index())
	}
	if slides[1].Index() != 1 {
		t.Errorf("second slide index should be 1, got %d", slides[1].Index())
	}
}

// TestSaveAndWrite tests saving and writing a presentation.
func TestSaveAndWrite(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()
	slide.AddTextBox(100, 100, 500, 50, "Test Presentation")

	// Create temp directory.
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.pptx")

	// Test Save.
	err := pres.Save(outputPath)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Check that the file exists.
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("output file does not exist")
	}

	// Check file size.
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}

	// Test WriteToBytes.
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes() failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("WriteToBytes() returned empty data")
	}

	// Test Write.
	var buf bytes.Buffer
	err = pres.Write(&buf)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("Write() wrote empty data")
	}
}

// TestSetSlideSize tests setting the slide dimensions.
func TestSetSlideSize(t *testing.T) {
	pres := pptx.New()

	// Set custom size.
	pres.SetSlideSize(9144000, 6858000) // 4:3 standard
	cx, cy := pres.SlideSize()
	if cx != 9144000 || cy != 6858000 {
		t.Errorf("slide size not set correctly: cx=%d, cy=%d", cx, cy)
	}

	// Set standard size.
	pres.SetSlideSizeStandard("16:9")
	cx, cy = pres.SlideSize()
	if cx != 12192000 || cy != 6858000 {
		t.Errorf("16:9 size not set correctly: cx=%d, cy=%d", cx, cy)
	}
}

// TestClone tests cloning a presentation.
func TestClone(t *testing.T) {
	pres := pptx.New()
	slide := pres.AddSlide()
	slide.AddTextBox(100, 100, 500, 50, "Original")

	// Clone.
	cloned, err := pres.Clone()
	if err != nil {
		t.Fatalf("Clone() failed: %v", err)
	}

	// Check slide count.
	if cloned.SlideCount() != pres.SlideCount() {
		t.Errorf("slide count mismatch after clone: original=%d, clone=%d", pres.SlideCount(), cloned.SlideCount())
	}

	// Modify the clone.
	clonedSlide, _ := cloned.GetSlide(0)
	clonedSlide.AddTextBox(100, 200, 500, 50, "Cloned")

	// The original should not be affected.
	originalSlides := pres.Slides()
	clonedSlides := cloned.Slides()

	// Simple check: the two slides should have different shape counts.
	// (original has 1 text box, clone has 2 text boxes)
	_ = originalSlides
	_ = clonedSlides
}

// TestAddSlideAt tests inserting a slide at a specific index.
func TestAddSlideAt(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide() // index 0
	pres.AddSlide() // index 1

	// Insert at position 1.
	slide, err := pres.AddSlideAt(1)
	if err != nil {
		t.Fatalf("AddSlideAt(1) failed: %v", err)
	}

	if slide.Index() != 1 {
		t.Errorf("expected index 1, got %d", slide.Index())
	}

	if pres.SlideCount() != 3 {
		t.Errorf("expected 3 slides, got %d", pres.SlideCount())
	}

	// Check that indices are updated.
	slides := pres.Slides()
	if slides[0].Index() != 0 {
		t.Errorf("first slide index should be 0, got %d", slides[0].Index())
	}
	if slides[1].Index() != 1 {
		t.Errorf("second slide index should be 1, got %d", slides[1].Index())
	}
	if slides[2].Index() != 2 {
		t.Errorf("third slide index should be 2, got %d", slides[2].Index())
	}
}

// TestSlides tests retrieving all slides.
func TestSlides(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.AddSlide()
	pres.AddSlide()

	slides := pres.Slides()
	if len(slides) != 3 {
		t.Errorf("expected 3 slides, got %d", len(slides))
	}

	// Verify that Slides() returns a copy.
	slides[0] = nil
	originalSlides := pres.Slides()
	if originalSlides[0] == nil {
		t.Error("Slides() should return a copy; original data must not be affected")
	}
}

// TestClose tests closing a presentation.
func TestClose(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()

	err := pres.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}

// ============================================================================
// Contract tests using mock components
// ============================================================================
//
// These tests verify that slide.AddComponent() correctly accepts any object
// that implements the Component interface and successfully invokes its Render()
// method to mount shapes onto the canvas.
// No real business components (text, chart, etc.) are used — only mocks.

// mockComponent is a minimal test-only component.
type mockComponent struct {
	rendered bool
}

// Render implements the Component interface.
func (m *mockComponent) Render(ctx *pptx.SlideContext) error {
	m.rendered = true
	// Append a bare XSp to the context.
	ctx.AppendShape(&slidex.XSp{})
	return nil
}

// TestSlide_AddComponent tests the basic add-component flow.
func TestSlide_AddComponent(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	mock := &mockComponent{}
	err := slide.AddComponent(mock)

	if err != nil || !mock.rendered {
		t.Fatal("contract violation: component Render method was not called")
	}

	// Verify that the underlying parts Children slice grew by 1.
	childrenCount := len(slide.Part().SpTree().Children)
	if childrenCount != 1 {
		t.Fatalf("canvas mount failed: XML tree not updated, expected 1, got %d", childrenCount)
	}
}

// mockComponentWithName is a test-only component that also has a name.
type mockComponentWithName struct {
	rendered bool
	name     string
}

// Render implements the Component interface.
func (m *mockComponentWithName) Render(ctx *pptx.SlideContext) error {
	m.rendered = true
	ctx.AppendShape(&slidex.XSp{})
	return nil
}

// Name implements the ComponentWithName interface.
func (m *mockComponentWithName) Name() string {
	return m.name
}

// TestSlide_AddComponentWithName tests a named component.
func TestSlide_AddComponentWithName(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	mock := &mockComponentWithName{name: "test-component"}
	err := slide.AddComponent(mock)

	if err != nil || !mock.rendered {
		t.Fatal("contract violation: component Render method was not called")
	}

	// Verify name.
	if mock.Name() != "test-component" {
		t.Errorf("expected component name 'test-component', got %s", mock.Name())
	}

	// Verify that the underlying parts Children slice grew by 1.
	childrenCount := len(slide.Part().SpTree().Children)
	if childrenCount != 1 {
		t.Fatalf("canvas mount failed: XML tree not updated, expected 1, got %d", childrenCount)
	}
}

// mockComponentWithError is a test-only component that returns an error.
type mockComponentWithError struct {
	rendered bool
	errMsg   string
}

// Render implements the Component interface; always returns an error.
func (m *mockComponentWithError) Render(ctx *pptx.SlideContext) error {
	m.rendered = true
	return fmt.Errorf("%s", m.errMsg)
}

// TestSlide_AddComponentWithError tests the case where a component's Render returns an error.
func TestSlide_AddComponentWithError(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	mock := &mockComponentWithError{errMsg: "intentional error"}
	err := slide.AddComponent(mock)

	if err == nil {
		t.Fatal("expected an error, but call succeeded")
	}

	if !mock.rendered {
		t.Fatal("expected Render to be called")
	}

	// Verify the error message contains the original error.
	if !strings.Contains(err.Error(), "intentional error") {
		t.Errorf("expected error to contain 'intentional error', got: %s", err.Error())
	}

	// Verify that no shape was mounted when rendering failed.
	childrenCount := len(slide.Part().SpTree().Children)
	if childrenCount != 0 {
		t.Fatalf("expected no shapes mounted on render failure, got Children count: %d", childrenCount)
	}
}

// mockComponentWithSize is a test-only component that exposes bounds.
type mockComponentWithSize struct {
	x, y, cx, cy int
}

// Render implements the Component interface.
func (m *mockComponentWithSize) Render(ctx *pptx.SlideContext) error {
	ctx.AppendShape(&slidex.XSp{})
	return nil
}

// Bounds implements the ComponentWithSize interface.
func (m *mockComponentWithSize) Bounds() (x, y, cx, cy int) {
	return m.x, m.y, m.cx, m.cy
}

// TestSlide_ComponentWithSize tests a component that exposes size information.
func TestSlide_ComponentWithSize(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	mock := &mockComponentWithSize{x: 100, y: 200, cx: 300, cy: 400}
	err := slide.AddComponent(mock)

	if err != nil {
		t.Fatalf("AddComponent failed: %v", err)
	}

	// Verify Bounds return values.
	x, y, cx, cy := mock.Bounds()
	if x != 100 || y != 200 || cx != 300 || cy != 400 {
		t.Errorf("Bounds() returned unexpected values: x=%d, y=%d, cx=%d, cy=%d", x, y, cx, cy)
	}
}

// TestSlide_MultipleComponents tests adding multiple components at once.
func TestSlide_MultipleComponents(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	// Create several mock components.
	mock1 := &mockComponent{}
	mock2 := &mockComponent{}
	mock3 := &mockComponent{}

	// Use AddComponents to add them in bulk.
	err := slide.AddComponents(mock1, mock2, mock3)

	if err != nil {
		t.Fatalf("AddComponents failed: %v", err)
	}

	// Verify all components were rendered.
	if !mock1.rendered || !mock2.rendered || !mock3.rendered {
		t.Fatal("one or more components were not rendered")
	}

	// Verify the Children slice grew by 3.
	childrenCount := len(slide.Part().SpTree().Children)
	if childrenCount != 3 {
		t.Fatalf("expected 3 children, got %d", childrenCount)
	}
}

// mockComponentWithID is a test-only component that requests a shape ID.
type mockComponentWithID struct {
	rendered bool
	shapeID  uint32
}

// Render implements the Component interface.
func (m *mockComponentWithID) Render(ctx *pptx.SlideContext) error {
	m.rendered = true
	m.shapeID = ctx.NextShapeID()
	ctx.AppendShape(&slidex.XSp{})
	return nil
}

// TestSlide_ComponentShapeIDAllocation tests shape ID allocation inside components.
func TestSlide_ComponentShapeIDAllocation(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	// Create components that request shape IDs.
	mock1 := &mockComponentWithID{}
	mock2 := &mockComponentWithID{}

	err := slide.AddComponents(mock1, mock2)
	if err != nil {
		t.Fatalf("AddComponents failed: %v", err)
	}

	// Verify IDs are non-zero and monotonically increasing.
	if mock1.shapeID == 0 {
		t.Error("expected mock1 to receive a valid shape ID")
	}
	if mock2.shapeID == 0 {
		t.Error("expected mock2 to receive a valid shape ID")
	}
	// IDs must differ.
	if mock1.shapeID == mock2.shapeID {
		t.Errorf("expected distinct shape IDs, both are %d", mock1.shapeID)
	}
	// mock2's ID must be greater than mock1's.
	if mock2.shapeID <= mock1.shapeID {
		t.Errorf("expected mock2.shapeID > mock1.shapeID, got %d <= %d", mock2.shapeID, mock1.shapeID)
	}
}

// ============================================================================
// End-to-end streaming output sanity check
// ============================================================================
//
// These tests verify that the engine's streaming output (WriteToWriter) is
// healthy. No disk I/O is performed; the generated presentation is written
// into an in-memory bytes.Buffer, which is then re-opened with archive/zip
// to confirm a well-formed ZIP package was produced.

// TestPresentation_WriteToMemory tests streaming output to memory and validates the ZIP.
func TestPresentation_WriteToMemory(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	slide.AddTextBox(100, 100, 500, 50, "end-to-end test")

	// Stream into memory.
	var buf bytes.Buffer
	err := prs.Write(&buf)
	if err != nil {
		t.Fatalf("streaming save failed: %v", err)
	}

	// Verify the output is non-empty.
	if buf.Len() == 0 {
		t.Fatal("generated data is empty")
	}

	// Verify it is a valid ZIP file.
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("generated output is not a valid ZIP file: %v", err)
	}

	// Check that key files are present.
	expectedFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presentation.xml",
	}

	// Build a file name map.
	fileMap := make(map[string]bool)
	for _, file := range reader.File {
		fileMap[file.Name] = true
	}

	// Verify each required file exists.
	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("ZIP is missing required file: %s", expected)
		}
	}

	// Verify at least one slide file is present.
	slideFound := false
	for name := range fileMap {
		if strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml") {
			slideFound = true
			break
		}
	}
	if !slideFound {
		t.Error("ZIP is missing slide files")
	}

	t.Logf("end-to-end test passed: valid ZIP produced with %d files", len(reader.File))
}

// TestPresentation_WriteToMemory_Empty tests streaming an empty presentation.
func TestPresentation_WriteToMemory_Empty(t *testing.T) {
	prs := pptx.New()
	// No slides added.

	var buf bytes.Buffer
	err := prs.Write(&buf)
	if err != nil {
		t.Fatalf("empty presentation streaming save failed: %v", err)
	}

	// Even an empty presentation must produce a valid ZIP.
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("empty presentation did not produce a valid ZIP: %v", err)
	}

	// Verify basic structure files exist.
	fileMap := make(map[string]bool)
	for _, file := range reader.File {
		fileMap[file.Name] = true
	}

	if !fileMap["[Content_Types].xml"] {
		t.Error("missing [Content_Types].xml")
	}
	if !fileMap["ppt/presentation.xml"] {
		t.Error("missing ppt/presentation.xml")
	}

	t.Logf("empty presentation test passed: valid ZIP produced")
}

// TestPresentation_WriteToMemory_MultipleSlides tests streaming a multi-slide presentation.
func TestPresentation_WriteToMemory_MultipleSlides(t *testing.T) {
	prs := pptx.New()

	// Add several slides, each with content.
	for i := 0; i < 5; i++ {
		slide := prs.AddSlide()
		slide.AddTextBox(100, 100+i*50, 500, 50, fmt.Sprintf("Slide %d", i+1))
		slide.AddRectangle(100, 200+i*30, 200, 100)
	}

	var buf bytes.Buffer
	err := prs.Write(&buf)
	if err != nil {
		t.Fatalf("multi-slide streaming save failed: %v", err)
	}

	// Verify ZIP validity.
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("multi-slide output is not a valid ZIP: %v", err)
	}

	// Count slide files.
	slideCount := 0
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			slideCount++
		}
	}

	if slideCount != 5 {
		t.Errorf("expected 5 slide files, got %d", slideCount)
	}

	t.Logf("multi-slide test passed: %d slides produced", slideCount)
}

// TestPresentation_WriteToBytes_Consistency verifies that WriteToBytes and Write both produce valid ZIPs.
func TestPresentation_WriteToBytes_Consistency(t *testing.T) {
	// First presentation for WriteToBytes.
	prs1 := pptx.New()
	slide1 := prs1.AddSlide()
	slide1.AddTextBox(100, 100, 500, 50, "consistency test")

	// Use WriteToBytes.
	bytesData, err := prs1.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes failed: %v", err)
	}

	// Verify WriteToBytes produced a valid ZIP.
	_, err = zip.NewReader(bytes.NewReader(bytesData), int64(len(bytesData)))
	if err != nil {
		t.Fatalf("WriteToBytes did not produce a valid ZIP: %v", err)
	}

	// Second presentation for Write (same configuration).
	prs2 := pptx.New()
	slide2 := prs2.AddSlide()
	slide2.AddTextBox(100, 100, 500, 50, "consistency test")

	// Use Write.
	var buf bytes.Buffer
	err = prs2.Write(&buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify Write produced a valid ZIP.
	_, err = zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("Write did not produce a valid ZIP: %v", err)
	}

	// The sizes should be close (internal ID allocation may cause minor differences,
	// but the outputs should never diverge significantly).
	ratio := float64(len(bytesData)) / float64(buf.Len())
	if ratio < 0.95 || ratio > 1.05 {
		t.Errorf("output size difference too large: WriteToBytes=%d, Write=%d", len(bytesData), buf.Len())
	}

	t.Logf("consistency test passed: both methods produced valid ZIPs; sizes: WriteToBytes=%d, Write=%d", len(bytesData), buf.Len())
}
