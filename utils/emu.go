package utils

// ============================================================================
// PowerPoint unit conversion utilities
// ============================================================================
//
// PowerPoint uses EMU (English Metric Units) as its internal unit of length.
//   1 inch      = 914 400 EMU
//   1 centimeter = 360 000 EMU
//
// Reference: ECMA-376 Office Open XML File Formats
// ============================================================================

import (
	"math"
	"strconv"
)

// EMU (English Metric Unit) constants.
const (
	// EMUsPerInch is the number of EMU in one inch.
	EMUsPerInch int64 = 914400

	// EMUsPerCentimeter is the number of EMU in one centimeter.
	EMUsPerCentimeter int64 = 360000

	// EMUsPerMillimeter is the number of EMU in one millimeter.
	EMUsPerMillimeter int64 = 36000

	// EMUsPerPoint is the number of EMU in one point (1 pt = 1/72 inch).
	EMUsPerPoint int64 = EMUsPerInch / 72

	// EMUsPerPixel is the number of EMU in one pixel at the default 96 DPI.
	EMUsPerPixel int64 = EMUsPerInch / 96
)

// Floating-point EMU constants (for multiplication).
var (
	// EMUsPerInchF is EMUsPerInch as a float64.
	EMUsPerInchF float64 = 914400.0

	// EMUsPerCentimeterF is EMUsPerCentimeter as a float64.
	EMUsPerCentimeterF float64 = 360000.0

	// EMUsPerMillimeterF is EMUsPerMillimeter as a float64.
	EMUsPerMillimeterF float64 = 36000.0

	// EMUsPerPointF is EMUsPerPoint as a float64.
	EMUsPerPointF float64 = 914400.0 / 72.0

	// EMUsPerPixelF is EMUsPerPixel as a float64 (96 DPI).
	EMUsPerPixelF float64 = 914400.0 / 96.0
)

// ============================================================================
// Types
// ============================================================================

// EMU is a typed int64 that carries English Metric Unit values.
type EMU int64

// Unit represents a measurement value together with its unit.
type Unit struct {
	Value float64
	Unit  UnitType
}

// UnitType enumerates the supported measurement units.
type UnitType int

const (
	UnitTypeEMU        UnitType = iota // EMU
	UnitTypeInch                       // inches
	UnitTypeCentimeter                 // centimeters
	UnitTypeMillimeter                 // millimeters
	UnitTypePoint                      // points
	UnitTypePixel                      // pixels
)

// ============================================================================
// EMU conversion functions
// ============================================================================

// InchesToEMU converts inches to EMU.
func InchesToEMU(inches float64) int64 {
	return int64(math.Round(inches * EMUsPerInchF))
}

// EMUToInches converts EMU to inches.
func EMUToInches(emu int64) float64 {
	return float64(emu) / EMUsPerInchF
}

// CentimetersToEMU converts centimeters to EMU.
func CentimetersToEMU(cm float64) int64 {
	return int64(math.Round(cm * EMUsPerCentimeterF))
}

// EMUToCentimeters converts EMU to centimeters.
func EMUToCentimeters(emu int64) float64 {
	return float64(emu) / EMUsPerCentimeterF
}

// MillimetersToEMU converts millimeters to EMU.
func MillimetersToEMU(mm float64) int64 {
	return int64(math.Round(mm * EMUsPerMillimeterF))
}

// EMUToMillimeters converts EMU to millimeters.
func EMUToMillimeters(emu int64) float64 {
	return float64(emu) / EMUsPerMillimeterF
}

// PointsToEMU converts points to EMU.
func PointsToEMU(points float64) int64 {
	return int64(math.Round(points * EMUsPerPointF))
}

// EMUToPoints converts EMU to points.
func EMUToPoints(emu int64) float64 {
	return float64(emu) / EMUsPerPointF
}

// PixelsToEMU converts pixels to EMU at 96 DPI.
func PixelsToEMU(pixels float64) int64 {
	return int64(math.Round(pixels * EMUsPerPixelF))
}

// EMUToPixels converts EMU to pixels at 96 DPI.
func EMUToPixels(emu int64) float64 {
	return float64(emu) / EMUsPerPixelF
}

// ============================================================================
// EMU methods
// ============================================================================

// NewEMU creates an EMU value from a raw int64.
func NewEMU(value int64) EMU {
	return EMU(value)
}

// Inches converts the EMU value to inches.
func (e EMU) Inches() float64 {
	return EMUToInches(int64(e))
}

// Centimeters converts the EMU value to centimeters.
func (e EMU) Centimeters() float64 {
	return EMUToCentimeters(int64(e))
}

// Millimeters converts the EMU value to millimeters.
func (e EMU) Millimeters() float64 {
	return EMUToMillimeters(int64(e))
}

// Points converts the EMU value to points.
func (e EMU) Points() float64 {
	return EMUToPoints(int64(e))
}

// Pixels converts the EMU value to pixels at 96 DPI.
func (e EMU) Pixels() float64 {
	return EMUToPixels(int64(e))
}

// ============================================================================
// Unit converter
// ============================================================================

// Converter performs pixel↔EMU conversions at a configurable DPI.
type Converter struct {
	dpi           int
	emusPerPixelF float64
}

// NewConverter creates a Converter at the default 96 DPI.
func NewConverter() *Converter {
	return &Converter{
		dpi:           96,
		emusPerPixelF: EMUsPerInchF / 96.0,
	}
}

// NewConverterWithDPI creates a Converter at the given DPI.
func NewConverterWithDPI(dpi int) *Converter {
	return &Converter{
		dpi:           dpi,
		emusPerPixelF: EMUsPerInchF / float64(dpi),
	}
}

// EMUsPerPixelForDPI returns the number of EMU per pixel at the current DPI.
func (c *Converter) EMUsPerPixelForDPI() int64 {
	return int64(math.Round(c.emusPerPixelF))
}

// PixelsToEMU converts pixels to EMU at the current DPI.
func (c *Converter) PixelsToEMU(pixels float64) int64 {
	return int64(math.Round(pixels * c.emusPerPixelF))
}

// EMUToPixels converts EMU to pixels at the current DPI.
func (c *Converter) EMUToPixels(emu int64) float64 {
	return float64(emu) / c.emusPerPixelF
}

// SetDPI updates the DPI used for pixel conversions.
func (c *Converter) SetDPI(dpi int) {
	c.dpi = dpi
	c.emusPerPixelF = EMUsPerInchF / float64(dpi)
}

// DPI returns the current DPI setting.
func (c *Converter) DPI() int {
	return c.dpi
}

// ============================================================================
// Common size constants
// ============================================================================

// SlideWidth is the standard widescreen (16:9) slide width.
const SlideWidth EMU = EMU(12192000) // 13.333 inches

// SlideHeight is the standard widescreen (16:9) slide height.
const SlideHeight EMU = EMU(6858000) // 7.5 inches

// StandardSlideWidth is the standard 4:3 slide width.
const StandardSlideWidth EMU = EMU(9144000) // 10 inches

// StandardSlideHeight is the standard 4:3 slide height.
const StandardSlideHeight EMU = EMU(6858000) // 7.5 inches

// ============================================================================
// Convenience functions
// ============================================================================

// MakeEMUPair returns an (x, y) coordinate pair in EMU.
func MakeEMUPair(x, y int64) (int64, int64) {
	return x, y
}

// MakeSizeEMU returns a (cx, cy) size pair in EMU.
func MakeSizeEMU(cx, cy int64) (int64, int64) {
	return cx, cy
}

// RectEMU is an axis-aligned rectangle with all coordinates in EMU.
type RectEMU struct {
	X  int64
	Y  int64
	Cx int64
	Cy int64
}

// NewRectEMU creates a RectEMU from raw EMU coordinates.
func NewRectEMU(x, y, cx, cy int64) RectEMU {
	return RectEMU{X: x, Y: y, Cx: cx, Cy: cy}
}

// FromInches returns a RectEMU constructed from inch measurements.
func (r RectEMU) FromInches(x, y, width, height float64) RectEMU {
	return RectEMU{
		X:  InchesToEMU(x),
		Y:  InchesToEMU(y),
		Cx: InchesToEMU(width),
		Cy: InchesToEMU(height),
	}
}

// FromCentimeters returns a RectEMU constructed from centimeter measurements.
func (r RectEMU) FromCentimeters(x, y, width, height float64) RectEMU {
	return RectEMU{
		X:  CentimetersToEMU(x),
		Y:  CentimetersToEMU(y),
		Cx: CentimetersToEMU(width),
		Cy: CentimetersToEMU(height),
	}
}

// ============================================================================
// XML attribute helpers
// ============================================================================

// WriteEMUAttr formats an EMU value as a decimal string suitable for an XML
// attribute.
func WriteEMUAttr(value int64) string {
	return strconv.FormatInt(value, 10)
}

// WriteEMUAttrs formats multiple EMU values as decimal strings.
func WriteEMUAttrs(values ...int64) []string {
	attrs := make([]string, len(values))
	for i, v := range values {
		attrs[i] = strconv.FormatInt(v, 10)
	}
	return attrs
}
