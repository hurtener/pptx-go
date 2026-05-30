package scene

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

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

// SlideTiming is the wall-clock time spent composing one slide, in scene order
// (D-015, D-016). Callers use it to detect render imbalance across a deck. It is
// never serialized into the PPTX, so it does not affect render idempotency.
type SlideTiming struct {
	SlideID  string
	Duration time.Duration
}

// Stats is the result of Render: per-render counts, per-slide timings, and
// non-fatal warnings (the library's observability surface — no event protocol,
// D-016).
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []LayoutWarning
	Timings  []SlideTiming
}

// renderConfig accumulates RenderOptions.
type renderConfig struct {
	resolver AssetResolver
	logger   *slog.Logger
	workers  int
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

// WithWorkers sets the number of slides composed concurrently (D-015). The
// default (n <= 0) is runtime.GOMAXPROCS(0); n == 1 forces sequential rendering.
// Render stays idempotent (byte-identical output) regardless of n: slides are
// created in scene order before composition, and any slide that registers
// global media renders sequentially in scene order so media numbering is stable.
func WithWorkers(n int) RenderOption {
	return func(c *renderConfig) { c.workers = n }
}

// Render composes a Scene onto pres and returns render Stats (RFC §10.1). It
// applies the scene's theme (if any), validates (Stage 1), then lays out and
// composes each slide's nodes via the builder (P1). Render is deterministic
// given the same scene + theme: re-rendering produces byte-identical output.
//
// Slides are created in scene order, then composed concurrently across a worker
// pool sized to runtime.GOMAXPROCS(0) (configurable via WithWorkers; D-015).
// A slide that may register global media renders sequentially in scene order so
// media part numbering — and therefore the bytes — stay deterministic; every
// other slide is independent and composes in parallel.
//
// V1 renders the text-heavy leaf nodes (Phase 06); container and asset nodes
// not yet implemented surface a LayoutWarning and are skipped.
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
	if s.Theme != nil {
		pres.SetTheme(s.Theme)
	}

	workers := cfg.workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	base := &renderer{pres: pres, cfg: cfg, theme: pres.Theme(), ctx: context.Background()}

	// Phase 1: create every slide in scene order. AddSlide serializes on the
	// presentation lock and appends in call order, so this fixes slide ordering
	// and the stable slideN.xml file numbers before any concurrency.
	slides := make([]*pptx.Slide, len(s.Slides))
	for i := range s.Slides {
		slides[i] = pres.AddSlide()
	}

	// Phase 2: compose. Media-bearing slides render sequentially in scene order
	// (deterministic global-media numbering); media-free slides — pure functions
	// of their nodes + theme, touching only their own slide part — fan out across
	// the worker pool. Results merge in scene order so Stats is deterministic.
	results := make([]slideResult, len(s.Slides))
	free := make([]int, 0, len(s.Slides))
	for i := range s.Slides {
		if workers > 1 && !slideUsesAssets(&s.Slides[i]) {
			free = append(free, i)
			continue
		}
		results[i] = base.composeOne(slides[i], &s.Slides[i])
	}

	if len(free) > 0 {
		sem := make(chan struct{}, workers)
		var wg sync.WaitGroup
		for _, idx := range free {
			wg.Add(1)
			sem <- struct{}{}
			go func(i int) {
				defer wg.Done()
				defer func() { <-sem }()
				results[i] = base.composeOne(slides[i], &s.Slides[i])
			}(idx)
		}
		wg.Wait()
	}

	var stats Stats
	for i := range results {
		stats.Slides++
		stats.Shapes += results[i].shapes
		stats.Assets += results[i].assets
		stats.Warnings = append(stats.Warnings, results[i].warnings...)
		stats.Timings = append(stats.Timings, SlideTiming{SlideID: s.Slides[i].ID, Duration: results[i].dur})
	}
	return stats, nil
}
