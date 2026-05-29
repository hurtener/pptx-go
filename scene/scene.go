package scene

import (
	"log/slog"

	"github.com/hurtener/pptx-go/pptx"
)

// The scene entrypoint (RFC §10.1). A Scene is a theme + ordered slides + deck
// metadata; Render composes it onto a *pptx.Presentation. Render is
// idempotent given the same scene + theme (a hard requirement — RFC §10.1).

// LayoutKind names a slide's structural intent; it maps to a master layout at
// render time (RFC §10.1).
type LayoutKind int

const (
	LayoutCover LayoutKind = iota
	LayoutTitleContent
	LayoutTwoColumn
	LayoutCardGrid
	LayoutFullBleed
	LayoutBlank
)

// Variant selects a theme variant for a slide (RFC §10.1). Per-slide overrides
// are V2; a per-scene variant is V1.
type Variant int

const (
	VariantLight Variant = iota
	VariantDark
	VariantPrint
)

// Metadata is deck-level core metadata.
type Metadata struct {
	Title   string
	Author  string
	Subject string
}

// Scene is the input to Render.
type Scene struct {
	Theme  *pptx.Theme // optional; the builder's default theme if nil
	Slides []SceneSlide
	Meta   Metadata
}

// SceneSlide is one slide in a Scene: a layout intent, the top-level node list,
// optional speaker notes, and a theme variant.
type SceneSlide struct {
	ID      string
	Layout  LayoutKind
	Nodes   []SlideNode
	Notes   RichText
	Variant Variant
}

// LayoutWarning is a non-fatal layout issue surfaced in Stats.Warnings (e.g.
// content overflow). A caller that wants warnings to be fatal inspects
// Stats.Warnings itself — pptx-go has no strict mode (RFC §10.2).
type LayoutWarning struct {
	SlideID string
	Node    string
	Message string
}

// Stats is the result of Render: per-render counts and non-fatal warnings (the
// library's observability surface — no event protocol, D-016).
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []LayoutWarning
}

// renderConfig accumulates RenderOptions.
type renderConfig struct {
	resolver AssetResolver
	logger   *slog.Logger
}

// RenderOption configures a Render call.
type RenderOption func(*renderConfig)

// WithAssetResolver registers the AssetResolver used to fetch asset bytes
// (§10.6).
func WithAssetResolver(r AssetResolver) RenderOption {
	return func(c *renderConfig) { c.resolver = r }
}

// WithLogger injects a structured logger for render diagnostics (no logger =
// no logs; D-016).
func WithLogger(l *slog.Logger) RenderOption {
	return func(c *renderConfig) { c.logger = l }
}

// Render composes a Scene onto pres and returns render Stats (RFC §10.1).
//
// It currently validates the scene (Stage 1) and returns a zero Stats without
// emitting shapes — rendering lands in later phases. The signature and
// validation contract are stable.
func Render(pres *pptx.Presentation, s Scene, opts ...RenderOption) (Stats, error) {
	var cfg renderConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if err := ValidateScene(s); err != nil {
		return Stats{}, err
	}
	// No-op stub: emission lands in Phase 06+.
	_ = pres
	return Stats{}, nil
}
