package pptx

import "math"

// EMU is the English Metric Unit, OOXML's canonical integer length unit.
// 1 inch = 914400 EMU. All public geometry in pptx is expressed in EMU; the
// Pt/Cm/In/Px constructors convert from human units at the call site.
type EMU int64

// EMU-per-unit constants (ISO/IEC 29500). Exact integers where the unit
// divides evenly; Pt and Px are computed with rounding in their constructors.
const (
	emuPerInch       = 914400
	emuPerCentimeter = 360000
	emuPerMillimeter = 36000
	emuPerPoint      = emuPerInch / 72 // 12700
	emuPerPixel      = emuPerInch / 96 // 9525 (96 DPI reference)
)

// Pt converts points to EMU (1 pt = 1/72 inch).
func Pt(pt float64) EMU { return EMU(math.Round(pt * float64(emuPerPoint))) }

// Cm converts centimeters to EMU.
func Cm(cm float64) EMU { return EMU(math.Round(cm * float64(emuPerCentimeter))) }

// In converts inches to EMU.
func In(in float64) EMU { return EMU(math.Round(in * float64(emuPerInch))) }

// Px converts pixels to EMU at the 96-DPI reference PowerPoint uses.
func Px(px float64) EMU { return EMU(math.Round(px * float64(emuPerPixel))) }

// Points returns the EMU value expressed in points.
func (e EMU) Points() float64 { return float64(e) / float64(emuPerPoint) }

// Inches returns the EMU value expressed in inches.
func (e EMU) Inches() float64 { return float64(e) / float64(emuPerInch) }

// Centimeters returns the EMU value expressed in centimeters.
func (e EMU) Centimeters() float64 { return float64(e) / float64(emuPerCentimeter) }

// Pixels returns the EMU value expressed in pixels at the 96-DPI reference.
func (e EMU) Pixels() float64 { return float64(e) / float64(emuPerPixel) }

// Standard slide canvas sizes (RFC §8.1 / D-023).
const (
	// Slide16x9Width / Slide16x9Height — the 13.333" × 7.5" widescreen canvas.
	Slide16x9Width  EMU = 12192000
	Slide16x9Height EMU = 6858000
	// Slide4x3Width / Slide4x3Height — the 10" × 7.5" classic canvas.
	Slide4x3Width  EMU = 9144000
	Slide4x3Height EMU = 6858000
)
