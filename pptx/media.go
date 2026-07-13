package pptx

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// ============================================================================
// Media — image insertion (RFC §8.6)
// ============================================================================
//
// Images enter a slide through an ImageSource (file, bytes, or reader; §4.4's
// interface + factory + driver seam) and AddImage. Identical bytes are written
// to the package once (dedup, preserving the upstream MediaManager); each slide
// still gets its own relationship id.
//
// Security (§7): pptx-go verifies the byte signature matches a known image type
// and rejects malformed or mismatched data, but it never parses pixel data — a
// malicious-but-well-formed image is the caller's problem at display time.

// ErrUnknownImageFormat is returned when image bytes carry no recognizable image
// signature (PNG, JPEG, GIF, BMP, WebP).
var ErrUnknownImageFormat = errors.New("pptx: unrecognized or malformed image data")

// ErrImageMIMEMismatch is returned when the declared MIME type does not match
// the type sniffed from the bytes.
var ErrImageMIMEMismatch = errors.New("pptx: declared image MIME does not match content")

// ImageSource is image input for AddImage. Construct one with ImageFile,
// ImageBytes, or ImageReader. The interface is sealed (resolveImage is
// unexported) so callers cannot inject a source the builder can't encode; new
// backends are added here behind the same seam.
type ImageSource interface {
	resolveImage() (imageData, error)
}

// imageData is the resolved, verified result of an ImageSource: the bytes, the
// canonical MIME type sniffed from them, and the matching file extension.
type imageData struct {
	bytes       []byte
	contentType string
	ext         string // e.g. ".png"
}

// fileImageSource reads an image from a filesystem path at resolve time.
type fileImageSource struct{ path string }

// ImageFile returns an ImageSource that reads the image at path. The format is
// taken from the bytes, not the file extension.
func ImageFile(path string) ImageSource { return fileImageSource{path: path} }

func (s fileImageSource) resolveImage() (imageData, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return imageData{}, fmt.Errorf("read image file %q: %w", s.path, err)
	}
	return verifyImage(data, "")
}

// bytesImageSource carries raw image bytes plus a declared MIME type.
type bytesImageSource struct {
	data []byte
	mime string
}

// ImageBytes returns an ImageSource from raw bytes with a declared MIME type
// (e.g. "image/png"). The declared type is verified against the bytes.
func ImageBytes(data []byte, mime string) ImageSource {
	return bytesImageSource{data: data, mime: mime}
}

func (s bytesImageSource) resolveImage() (imageData, error) {
	return verifyImage(s.data, s.mime)
}

// readerImageSource carries an io.Reader drained at resolve time.
type readerImageSource struct {
	r    io.Reader
	mime string
}

// ImageReader returns an ImageSource that reads all bytes from r with a declared
// MIME type. r is drained when the image is added.
func ImageReader(r io.Reader, mime string) ImageSource {
	return readerImageSource{r: r, mime: mime}
}

func (s readerImageSource) resolveImage() (imageData, error) {
	if s.r == nil {
		return imageData{}, fmt.Errorf("pptx: ImageReader: nil reader")
	}
	data, err := io.ReadAll(s.r)
	if err != nil {
		return imageData{}, fmt.Errorf("read image: %w", err)
	}
	return verifyImage(data, s.mime)
}

// verifyImage sniffs the canonical image type from data's magic bytes and, when
// a MIME type is declared, confirms it matches. It never inspects pixels (§7).
func verifyImage(data []byte, declaredMIME string) (imageData, error) {
	ct, ext := sniffImage(data)
	if ct == "" {
		return imageData{}, ErrUnknownImageFormat
	}
	if declaredMIME != "" && declaredMIME != ct {
		return imageData{}, fmt.Errorf("%w: declared %q, content is %q", ErrImageMIMEMismatch, declaredMIME, ct)
	}
	return imageData{bytes: data, contentType: ct, ext: ext}, nil
}

// sniffImage returns the canonical MIME type and file extension for a supported
// image format, or ("", "") if the signature is unrecognized. Recognition is by
// magic bytes only — no pixel parsing.
func sniffImage(b []byte) (mime, ext string) {
	switch {
	case len(b) >= 8 && string(b[:8]) == "\x89PNG\r\n\x1a\n":
		return "image/png", ".png"
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return "image/jpeg", ".jpg"
	case len(b) >= 6 && (string(b[:6]) == "GIF87a" || string(b[:6]) == "GIF89a"):
		return "image/gif", ".gif"
	case len(b) >= 2 && b[0] == 'B' && b[1] == 'M':
		return "image/bmp", ".bmp"
	case len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP":
		return "image/webp", ".webp"
	default:
		return "", ""
	}
}

// addImagePart registers image bytes with the presentation's media manager
// (global content-dedup; identical bytes share one package part) and adds the
// slide-local image relationship, returning its relationship id. The media
// bytes are written to the package by syncMedia at save time.
func (s *Slide) addImagePart(data []byte, ext string) string {
	if ext == "" {
		ext = ".png"
	}
	_, res := s.mediaManager.AddMediaForSlide(s.index, data, "image"+ext)
	// Absolute pack URI so the relationship resolves to the media part
	// regardless of the slide's directory.
	return s.part.AddImageRel("/" + res.Target())
}

// Image is an opaque handle to an image added to a slide. It exposes typed
// mutators (alt text, crop, fit) without surfacing the OOXML wire type (P3), and
// read accessors over a reopened deck (RFC §16).
type Image struct {
	s   *Slide // owning slide (read side: resolves the embedded bytes)
	pic *slide.XPicture
}

// Crop is a per-edge crop expressed as a fraction (0..1) of the corresponding
// image dimension trimmed from that edge.
type Crop struct {
	Left, Top, Right, Bottom float64
}

// Fit selects how an image fills its frame. PowerPoint stores no single "fit"
// value, so V1 ships FitFill / FitNone and lets caller-side Box sizing drive
// aspect (engine, not product — D-026). Aspect-aware cover/contain is a V1.x
// candidate: it needs the image's dimensions, which can be read from the format
// header via image.DecodeConfig — the dimension header is not pixel data, so it
// is permitted (§7/D-046); the chart composer already reads it for aspect-fit.
type Fit int

const (
	// FitFill stretches the image to fill the frame (the default).
	FitFill Fit = iota
	// FitNone places the image without a stretch fill mode.
	FitNone
)

// AddImage places an image on the slide, positioned by box (EMU), and returns a
// handle for optional alt text / crop / fit. Identical bytes across the deck are
// written once (dedup). It errors if the source can't be read or the bytes are
// not a recognized image (§7). (RFC §8.6.)
func (s *Slide) AddImage(src ImageSource, box Box) (*Image, error) {
	if src == nil {
		return nil, fmt.Errorf("pptx: AddImage: nil image source")
	}
	img, err := src.resolveImage()
	if err != nil {
		return nil, fmt.Errorf("pptx: AddImage: %w", err)
	}

	rID := s.addImagePart(img.bytes, img.ext)
	pic := s.builder.AddPicture(int(box.X), int(box.Y), int(box.W), int(box.H), rID)
	return &Image{s: s, pic: pic}, nil
}

// SetAltText sets the image's alternative text (the cNvPr/@descr attribute).
func (im *Image) SetAltText(text string) *Image {
	if im != nil && im.pic != nil && im.pic.NonVisual.CNvPr != nil {
		im.pic.NonVisual.CNvPr.Descr = text
	}
	return im
}

// SetCrop sets a source-rectangle crop (fractions 0..1 trimmed per edge).
func (im *Image) SetCrop(c Crop) *Image {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil {
		return im
	}
	im.pic.BlipFill.SrcRect = &slide.XSrcRect{
		L: cropPermille(c.Left),
		T: cropPermille(c.Top),
		R: cropPermille(c.Right),
		B: cropPermille(c.Bottom),
	}
	return im
}

// SetFit sets the image fill mode (FitFill stretches; FitNone omits the stretch
// fill).
func (im *Image) SetFit(f Fit) *Image {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil {
		return im
	}
	switch f {
	case FitNone:
		im.pic.BlipFill.Stretch = nil
	default:
		if im.pic.BlipFill.Stretch == nil {
			im.pic.BlipFill.Stretch = &slide.XStretchProperties{FillRect: &slide.XFillRectProperties{}}
		}
	}
	return im
}

type GeometryOptions struct {
	Shape  ShapeGeometry
	Radius RadiusRole
}

func (im *Image) SetGeometry(opt GeometryOptions) *Image {
	if im == nil || im.pic == nil || im.pic.ShapeProperties == nil {
		return im
	}

	spPr := im.pic.ShapeProperties
	spPr.PresetGeom = &slide.XPresetGeometry{
		Prst:  string(opt.Shape),
		AvLst: &slide.XAvLst{},
	}

	if opt.Shape == ShapeRoundRect && im.s != nil {
		r := im.s.activeTheme().ResolveRadius(opt.Radius)
		if r > 0 {
			applyCornerRadius(spPr, r, picBox(spPr))
		}
	}

	return im
}

// SetCornerRadius rounds the picture's corners using a theme radius token (P2,
// D-114): it sets the picture geometry to roundRect and converts the absolute
// token radius to the OOXML adjust against the picture box. RadiusNone (the zero
// value) resolves to 0 and leaves the picture rectangular — byte-identical. The
// rounded picture matches the card/surface radius finish.
func (im *Image) SetCornerRadius(role RadiusRole) *Image {
	return im.SetGeometry(GeometryOptions{
		Shape:  ShapeRoundRect,
		Radius: role,
	})
}

// SetElevation casts a soft drop shadow on the picture from a theme elevation
// token (P2, D-114), matching the card/surface elevation finish. ElevationFlat
// (the zero value) resolves to a flat elevation and emits no shadow —
// byte-identical.
func (im *Image) SetElevation(role ElevationRole) *Image {
	if im == nil || im.pic == nil || im.pic.ShapeProperties == nil || im.s == nil {
		return im
	}
	applyShadow(im.pic.ShapeProperties, im.s.activeTheme().ResolveElevation(role))
	return im
}

// SetDuotone recolors the picture as a two-tone (duotone) image: the picture's
// shadows map to shadow and its highlights to highlight, producing an on-brand
// tint of a photo (R14.1). Both colors accept theme tokens or literals (P2), so
// a theme swap re-tints the photo. A nil color on either side leaves the picture
// un-recolored (byte-identical). The colors are resolved against the active
// theme at call time and emitted as literal <a:srgbClr> values inside an
// <a:duotone> blip effect.
func (im *Image) SetDuotone(shadow, highlight Color) *Image {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil || im.pic.BlipFill.Blip == nil || im.s == nil {
		return im
	}
	if shadow == nil || highlight == nil {
		return im // need both tones; leave the picture un-recolored
	}
	t := im.s.activeTheme()
	im.pic.BlipFill.Blip.Duotone = &slide.XDuotone{
		Colors: []slide.XSrgbClr{
			{Val: string(shadow.resolve(t).rgb)},
			{Val: string(highlight.resolve(t).rgb)},
		},
	}
	return im
}

// Duotone returns the picture's two-tone (shadow, highlight) recolor as resolved
// hex values and ok=true when a duotone effect is set (the read inverse of
// SetDuotone); ok=false when the picture is not recolored.
func (im *Image) Duotone() (shadow, highlight RGB, ok bool) {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil || im.pic.BlipFill.Blip == nil {
		return "", "", false
	}
	d := im.pic.BlipFill.Blip.Duotone
	if d == nil || len(d.Colors) != 2 {
		return "", "", false
	}
	return RGB(d.Colors[0].Val), RGB(d.Colors[1].Val), true
}

// picBox reconstructs the picture's box from its transform (offset + extent).
func picBox(spPr *slide.XShapeProperties) Box {
	t := spPr.Transform2D
	if t == nil || t.Offset == nil || t.Extent == nil {
		return Box{}
	}
	return Box{X: EMU(t.Offset.X), Y: EMU(t.Offset.Y), W: EMU(t.Extent.Cx), H: EMU(t.Extent.Cy)}
}

// SetRotation rotates the image clockwise by deg degrees about its centre
// (the picture's <a:xfrm rot>, normalized to [0,360°)).
func (im *Image) SetRotation(deg float64) *Image {
	if im == nil || im.pic == nil || im.pic.ShapeProperties == nil || im.pic.ShapeProperties.Transform2D == nil {
		return im
	}
	im.pic.ShapeProperties.Transform2D.Rotation = normalizeAngle60k(deg)
	return im
}

// SetOpacity scales the image opacity via the blip's <a:alphaModFix> (alpha
// 0..100000; AlphaOpaque clears the effect). It is the picture analogue of a
// fill's alpha — used by a Decoration's opacity.
func (im *Image) SetOpacity(alpha int) *Image {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil || im.pic.BlipFill.Blip == nil {
		return im
	}
	a := clampAlpha(alpha)
	if a >= AlphaOpaque {
		im.pic.BlipFill.Blip.AlphaModFix = nil
		return im
	}
	im.pic.BlipFill.Blip.AlphaModFix = &slide.XAlphaModFix{Amt: a}
	return im
}

// ============================================================================
// Read accessors (RFC §16) — the read inverse of the image authoring API.
// ============================================================================

// ErrImagePartMissing is returned by Image.Bytes when the picture's embedded
// relationship or its media part cannot be resolved in the reopened package.
var ErrImagePartMissing = errors.New("pptx: image media part not found")

// AltText returns the image's alternative text — the read inverse of SetAltText
// (empty when unset).
func (im *Image) AltText() string {
	if im != nil && im.pic != nil && im.pic.NonVisual.CNvPr != nil {
		return im.pic.NonVisual.CNvPr.Descr
	}
	return ""
}

// Crop returns the image's per-edge crop as fractions (0..1) — the read inverse
// of SetCrop (a zero Crop when uncropped).
func (im *Image) Crop() Crop {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil || im.pic.BlipFill.SrcRect == nil {
		return Crop{}
	}
	r := im.pic.BlipFill.SrcRect
	return Crop{
		Left:   float64(r.L) / 100000.0,
		Top:    float64(r.T) / 100000.0,
		Right:  float64(r.R) / 100000.0,
		Bottom: float64(r.B) / 100000.0,
	}
}

// Fit returns the image's fill mode — the read inverse of SetFit (FitFill when a
// stretch fill is present, FitNone otherwise).
func (im *Image) Fit() Fit {
	if im != nil && im.pic != nil && im.pic.BlipFill != nil && im.pic.BlipFill.Stretch != nil {
		return FitFill
	}
	return FitNone
}

// Rotation returns the image's clockwise rotation in degrees within [0, 360°) —
// the read inverse of SetRotation (0 when unset).
func (im *Image) Rotation() float64 {
	if im == nil || im.pic == nil || im.pic.ShapeProperties == nil || im.pic.ShapeProperties.Transform2D == nil {
		return 0
	}
	return float64(im.pic.ShapeProperties.Transform2D.Rotation) / 60000.0
}

// Opacity returns the image's opacity (OOXML alpha 0..100000) — the read inverse
// of SetOpacity (AlphaOpaque when no alpha-modulation effect is set).
func (im *Image) Opacity() int {
	if im != nil && im.pic != nil && im.pic.BlipFill != nil && im.pic.BlipFill.Blip != nil &&
		im.pic.BlipFill.Blip.AlphaModFix != nil {
		return im.pic.BlipFill.Blip.AlphaModFix.Amt
	}
	return AlphaOpaque
}

// Bytes resolves the image's embedded bytes by following the picture's
// <a:blip r:embed> relationship to its media part in the reopened package (R4).
// It returns ErrImagePartMissing when the relationship or part is absent. The
// bytes are returned verbatim — pptx-go does not decode pixel data (§7).
func (im *Image) Bytes() ([]byte, error) {
	if im == nil || im.pic == nil || im.pic.BlipFill == nil || im.pic.BlipFill.Blip == nil {
		return nil, ErrImagePartMissing
	}
	rid := im.pic.BlipFill.Blip.Embed
	if rid == "" || im.s == nil || im.s.part == nil || im.s.presentation == nil || im.s.presentation.pkg == nil {
		return nil, ErrImagePartMissing
	}
	rel := im.s.part.Relationships().Get(rid)
	if rel == nil {
		return nil, fmt.Errorf("%w: relationship %q", ErrImagePartMissing, rid)
	}
	part := im.s.presentation.pkg.GetPart(rel.TargetURI())
	if part == nil {
		return nil, fmt.Errorf("%w: %s", ErrImagePartMissing, rel.TargetURI().URI())
	}
	return part.Blob(), nil
}

// cropPermille converts a 0..1 crop fraction to OOXML's thousandths-of-a-percent
// (0..100000), clamped to range.
func cropPermille(frac float64) int {
	v := int(math.Round(frac * 100000))
	switch {
	case v < 0:
		return 0
	case v > 100000:
		return 100000
	default:
		return v
	}
}
