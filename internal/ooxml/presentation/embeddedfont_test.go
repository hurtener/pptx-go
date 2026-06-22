package presentation

import "testing"

func TestHasEmbeddedFace(t *testing.T) {
	p := NewPresentationPart()
	if p.HasEmbeddedFace("Inter", "regular") {
		t.Fatal("empty part reports a face")
	}
	p.AddEmbeddedFont("Inter", "regular", "rId10")
	p.AddEmbeddedFont("Inter", "bold", "rId11")

	if !p.HasEmbeddedFace("Inter", "regular") {
		t.Error("recorded (Inter, regular) not found")
	}
	if !p.HasEmbeddedFace("Inter", "bold") {
		t.Error("recorded (Inter, bold) not found")
	}
	if p.HasEmbeddedFace("Inter", "italic") {
		t.Error("unrecorded (Inter, italic) reported present")
	}
	if p.HasEmbeddedFace("Playfair Display", "regular") {
		t.Error("unrecorded typeface reported present")
	}
}
