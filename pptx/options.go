package pptx

import (
	"log/slog"

	"github.com/hurtener/pptx-go/internal/ooxml/presentation"
)

// Format is a standard slide canvas aspect ratio (RFC §8.1). Pass one to
// pptx.New via WithFormat.
type Format int

const (
	// Slides16x9 is the 16:9 widescreen canvas (12192000 x 6858000 EMU,
	// 13.333" x 7.5"). This is the default.
	Slides16x9 Format = iota

	// Slides4x3 is the 4:3 standard canvas (9144000 x 6858000 EMU, 10" x 7.5").
	Slides4x3
)

// size returns the EMU canvas dimensions for the format.
func (f Format) size() presentation.SlideSize {
	switch f {
	case Slides4x3:
		return presentation.SlideSize{Cx: int(Slide4x3Width), Cy: int(Slide4x3Height)}
	default:
		return presentation.SlideSize{Cx: int(Slide16x9Width), Cy: int(Slide16x9Height)}
	}
}

// Option configures a Presentation at construction time (pptx.New). Options
// apply in order, before the scaffold is seeded.
type Option func(*Presentation)

// WithFormat sets the slide canvas aspect ratio (default Slides16x9).
func WithFormat(f Format) Option {
	return func(p *Presentation) {
		p.presentationPart.SetSlideSize(f.size())
	}
}

// WithFontSource registers the FontSource used by EmbedFont (D-019). It is the
// option-form of the SetFontSource setter and the documented registration path.
func WithFontSource(src FontSource) Option {
	return func(p *Presentation) {
		p.fontSource = src
	}
}

// WithLogger injects a structured logger (RFC §18, D-042). When set, the
// builder emits a Debug write-boundary event on each write/save; no logger =
// no logs (zero-cost). The handler's performance is the caller's concern.
func WithLogger(l *slog.Logger) Option {
	return func(p *Presentation) { p.logger = l }
}

// WithTheme sets the active theme (default DefaultTheme). The theme drives
// token resolution; theme-token emission into theme1.xml lands with the Color
// interface work.
func WithTheme(t *Theme) Option {
	return func(p *Presentation) {
		if t != nil {
			p.theme = t
		}
	}
}
