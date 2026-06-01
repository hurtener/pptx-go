package scene

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/frames"
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

// FrameRecipe draws a device frame's bezel into a region and returns the
// interior Box the renderer inserts an image into, plus the number of bezel
// shapes emitted. It composes the public pptx builder only (P1). Register one
// under a name with WithFrameExtension (RFC §14.4, D-038).
type FrameRecipe = frames.Recipe

// frameExtension is a caller frame registered for one render.
type frameExtension struct {
	name   string
	recipe FrameRecipe
}

// renderConfig accumulates RenderOptions.
type renderConfig struct {
	resolver  AssetResolver
	logger    *slog.Logger
	workers   int
	theme     *pptx.Theme
	layoutMap LayoutMap
	frameExt  []frameExtension
	frames    *frames.Registry // built in Render: curated ∪ frameExt
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

// WithTheme applies t as the active theme for the render — the brand-kit flow
// (RFC §13.1, §13.3): a scene authored against token roles re-renders in the
// brand's palette and fonts (P2). It takes precedence over the Scene's Theme
// field. A nil theme is ignored.
func WithTheme(t *pptx.Theme) RenderOption {
	return func(c *renderConfig) {
		if t != nil {
			c.theme = t
		}
	}
}

// WithFrameExtension registers a caller frame recipe under name for this render
// (RFC §14.4, D-038). The name joins the closed curated set {browser, phone,
// desktop, laptop}; registering a curated name overrides that frame for this
// render only. Extensions are per-render, not global state — concurrent renders
// with different extensions do not interfere. An Image whose resolved frame
// name is neither curated nor registered fails Stage-1 validation. A blank name
// or nil recipe is ignored.
func WithFrameExtension(name string, recipe FrameRecipe) RenderOption {
	return func(c *renderConfig) {
		if name != "" && recipe != nil {
			c.frameExt = append(c.frameExt, frameExtension{name: name, recipe: recipe})
		}
	}
}

// WithLayoutMap maps each slide's LayoutKind to a named layout in the active
// template's master (RFC §13.2). A slide whose mapped layout the template
// defines is related to it; an unmapped kind, or a name the template lacks,
// falls back to the blank layout (the latter records a LayoutWarning).
func WithLayoutMap(m LayoutMap) RenderOption {
	return func(c *renderConfig) { c.layoutMap = m }
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
	// Build the per-render frame registry (curated ∪ extensions) and run the
	// registry-aware half of Stage-1 validation: an Image's resolved frame name
	// must be curated or registered (RFC §14.4, D-038). The registry is built
	// here, before composition, and is read-only during the parallel compose.
	cfg.frames = frames.Curated()
	for _, ext := range cfg.frameExt {
		cfg.frames = cfg.frames.With(ext.name, ext.recipe)
	}
	if err := validateFrameRefs(s, cfg.frames); err != nil {
		return Stats{}, err
	}
	// Theme precedence: WithTheme option > Scene.Theme field > the presentation's
	// existing theme (RFC §13.1/§13.3).
	switch {
	case cfg.theme != nil:
		pres.SetTheme(cfg.theme)
	case s.Theme != nil:
		pres.SetTheme(s.Theme)
	}

	workers := cfg.workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	base := &renderer{pres: pres, cfg: cfg, theme: pres.Theme(), ctx: context.Background()}

	// Phase 1: create every slide in scene order. AddSlide serializes on the
	// presentation lock and appends in call order, so this fixes slide ordering
	// and the stable slideN.xml file numbers before any concurrency. The slide's
	// LayoutKind resolves through the LayoutMap to a template layout (RFC §13.2);
	// a mapped name the template lacks falls back to blank and warns.
	slides := make([]*pptx.Slide, len(s.Slides))
	layoutWarn := make([]*LayoutWarning, len(s.Slides))
	for i := range s.Slides {
		name := cfg.layoutMap.nameFor(s.Slides[i].Layout)
		if name != "" && !pres.HasLayout(name) {
			layoutWarn[i] = &LayoutWarning{
				SlideID: s.Slides[i].ID,
				Message: fmt.Sprintf("layout %q not in template; using blank layout", name),
			}
			name = ""
		}
		slides[i] = pres.AddSlide(name)
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
		if layoutWarn[i] != nil {
			stats.Warnings = append(stats.Warnings, *layoutWarn[i])
		}
		stats.Warnings = append(stats.Warnings, results[i].warnings...)
		stats.Timings = append(stats.Timings, SlideTiming{SlideID: s.Slides[i].ID, Duration: results[i].dur})
	}
	return stats, nil
}
