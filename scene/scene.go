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
	"github.com/hurtener/pptx-go/scene/icons"
	"github.com/hurtener/pptx-go/scene/ornaments"
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

// Variant selects a named theme variant for a slide (RFC §13.3). VariantDark
// is implemented: it derives a per-slide dark theme and swaps the active theme
// for the duration of that slide's composition. VariantPrint is not yet
// implemented and surfaces a LayoutWarning rather than silently rendering with
// the active theme.
type Variant int

const (
	VariantLight Variant = iota
	VariantDark
	VariantPrint
)

// String returns the variant's name.
func (v Variant) String() string {
	switch v {
	case VariantDark:
		return "dark"
	case VariantPrint:
		return "print"
	default:
		return "light"
	}
}

// Metadata is deck-level core metadata.
type Metadata struct {
	Title   string
	Author  string
	Subject string
}

// Chrome configures optional, opt-in slide chrome (RFC §10.2): recurring
// per-slide furniture drawn outside the body region — a top section eyebrow and
// a bottom footer (brand slot + "N / total" page number). The zero value
// (Enabled == false) draws no chrome, so a chrome-free deck is byte-identical to
// one authored before this field existed.
//
// Chrome is a mechanism, not a judgment (D-026): the engine draws the bands it
// is handed and composes the page-number string, but invents no brand and no
// section names. Colors resolve through theme tokens (TextMuted, ColorSurfaceAlt)
// so a theme swap re-skins chrome.
type Chrome struct {
	Enabled    bool    // master switch; false (zero value) = no chrome
	Brand      string  // footer-left brand text; used when BrandAsset is empty
	BrandAsset AssetID // footer-left brand image, resolved via the AssetResolver
	Total      int     // page-number denominator ("N / Total"); 0 = len(Scene.Slides)
}

// Scene is the input to Render.
type Scene struct {
	Theme  *pptx.Theme // optional; the builder's default theme if nil
	Slides []SceneSlide
	Meta   Metadata
	Chrome Chrome // optional opt-in slide chrome; zero value = disabled
}

// SceneSlide is one slide in a Scene: a layout intent, the top-level node list,
// optional speaker notes, a theme variant, body-stack alignment, and an optional
// full-bleed slide background. The zero value for Content ({VAlignTop,
// HAlignLeft}) reproduces the pre-Phase-13 layout unchanged — fully
// backward-compatible. A zero Background (BackgroundNone) draws nothing, so
// adding this field to existing call sites requires no change.
type SceneSlide struct {
	ID         string
	Layout     LayoutKind
	Nodes      []SlideNode
	Notes      RichText
	Variant    Variant
	Content    Alignment  // body-stack alignment; zero value = top-left (default)
	Background Background // full-bleed slide background; zero value = no background (BackgroundNone)
	Section    string     // chrome: top eyebrow label; empty = no eyebrow on this slide
	PageNumber int        // chrome: the N in "N / total"; 0 = scene position (1-based)
	// Footnotes are source/citation/disclaimer lines pinned to a reserved band at
	// the bottom of the slide (above the chrome footer), in the muted text role
	// (R14.12, D-126). The body region shrinks to reserve the band, so footnotes
	// never overlap the body or the page-number footer. Empty = no band
	// (byte-identical). Lines past a region cap are dropped with a warning.
	Footnotes []RichText
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

// SlideColors are the colors the engine actually resolved for one slide (D-058,
// extended R8.10): the canvas (base background), surface, alternate surface,
// accent + accent-alt, primary-text, and accent-text RGBs it rendered with —
// including a VariantDark slide's derived dark palette (so a soul's per-variant
// overrides are reflected). A caller uses them to verify soul→engine fidelity
// (resolved == the soul's intended token per role/variant) or to compute its own
// contrast; the engine performs no contrast logic (D-026), it only reports what
// it resolved. All fields are scalar RGB, so SlideColors stays comparable (==).
type SlideColors struct {
	SlideID     string
	Canvas      pptx.RGB // resolved ColorCanvas (the slide's base background)
	Surface     pptx.RGB // resolved ColorSurface
	SurfaceAlt  pptx.RGB // resolved ColorSurfaceAlt (R8.10)
	Accent      pptx.RGB // resolved ColorAccent (R8.10)
	AccentAlt   pptx.RGB // resolved ColorAccentAlt (R8.10)
	PrimaryText pptx.RGB // resolved TextPrimary
	TextAccent  pptx.RGB // resolved TextAccent (R8.10)
}

// Stats is the result of Render: per-render counts, per-slide timings, per-slide
// resolved colors, and non-fatal warnings (the library's observability surface —
// no event protocol, D-016).
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []LayoutWarning
	Timings  []SlideTiming
	Colors   []SlideColors // per-slide resolved colors, in scene order
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

// iconExtension is a caller icon SVG registered for one render.
type iconExtension struct {
	name string
	svg  []byte
}

// OrnamentRecipe draws an ornament into a box at a caller opacity (OOXML alpha)
// and rotation, returning the shape count. It composes the public pptx builder
// only (P1). Register one under a name with WithOrnamentExtension.
type OrnamentRecipe = ornaments.Recipe

// ornamentExtension is a caller ornament registered for one render.
type ornamentExtension struct {
	name   string
	recipe OrnamentRecipe
}

// renderConfig accumulates RenderOptions.
type renderConfig struct {
	resolver    AssetResolver
	logger      *slog.Logger
	workers     int
	theme       *pptx.Theme
	layoutMap   LayoutMap
	frameExt    []frameExtension
	frames      *frames.Registry // built in Render: curated ∪ frameExt
	iconExt     []iconExtension
	icons       *icons.Registry // built in Render: curated ∪ iconExt
	ornamentExt []ornamentExtension
	ornaments   *ornaments.Registry // built in Render: curated ∪ ornamentExt
	ctx         context.Context
}

// RenderOption configures a Render call.
type RenderOption func(*renderConfig)

// WithAssetResolver registers the AssetResolver used to fetch asset bytes
// (§10.6).
func WithAssetResolver(r AssetResolver) RenderOption {
	return func(c *renderConfig) { c.resolver = r }
}

// WithLogger injects a structured logger for render diagnostics (no logger =
// no logs; D-016). When set, Render emits a render-boundary summary and a Warn
// event for every LayoutWarning (RFC §18); the handler's performance is the
// caller's concern (slog calls are synchronous).
func WithLogger(l *slog.Logger) RenderOption {
	return func(c *renderConfig) { c.logger = l }
}

// WithContext sets the context Render uses: the AssetResolver receives it, and
// Render honors cancellation between slides (returning ctx.Err()). The default
// is context.Background(). (CLAUDE.md §5 — honor cancellation on I/O.)
func WithContext(ctx context.Context) RenderOption {
	return func(c *renderConfig) { c.ctx = ctx }
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

// WithIconExtension registers a caller icon under name for this render (RFC
// §14.1/§14.4, D-005). The SVG is validated when the option is applied; an SVG
// that violates the icon translator constraints (single path, solid fill, no
// gradients, no elliptical arcs) fails the render with a Stage-1 error — at
// registration, not at compose. Registering a curated name overrides it for this
// render only. A blank name or nil SVG is ignored. The icon is placed by the
// nodes that accept one (card, flow) in later phases.
func WithIconExtension(name string, svg []byte) RenderOption {
	return func(c *renderConfig) {
		if name != "" && svg != nil {
			c.iconExt = append(c.iconExt, iconExtension{name: name, svg: svg})
		}
	}
}

// ValidateIcon reports whether svg satisfies the icon translator constraints,
// so a caller can validate at its own registration point. It re-exports
// pptx.ValidateIcon (scene never reaches under pptx — P1).
func ValidateIcon(svg []byte) error { return pptx.ValidateIcon(svg) }

// WithOrnamentExtension registers a caller ornament recipe under name for this
// render (RFC §14.2/§14.4, D-038). The name joins the closed curated set;
// registering a curated name overrides it for this render only. Extensions are
// per-render, not global. A Decoration whose preset name is neither curated nor
// registered fails Stage-1 validation. A blank name or nil recipe is ignored.
func WithOrnamentExtension(name string, recipe OrnamentRecipe) RenderOption {
	return func(c *renderConfig) {
		if name != "" && recipe != nil {
			c.ornamentExt = append(c.ornamentExt, ornamentExtension{name: name, recipe: recipe})
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
	// Build the per-render icon registry (curated ∪ extensions) and run the
	// registry-aware half of Stage-1 validation: a caller SVG outside the
	// translator subset fails at registration (RFC §14.1, D-005), and a card's
	// Icon name must resolve to a curated or registered icon before any slide
	// composes (D-043). Curated icons are pre-validated by their build-time test.
	// The registry is read-only during the parallel compose.
	cfg.icons = icons.Curated()
	for _, ext := range cfg.iconExt {
		if err := pptx.ValidateIcon(ext.svg); err != nil {
			return Stats{}, fmt.Errorf("scene: icon extension %q: %w", ext.name, err)
		}
		cfg.icons = cfg.icons.With(ext.name, ext.svg)
	}
	if err := validateIconRefs(s, cfg.icons); err != nil {
		return Stats{}, err
	}
	// Build the per-render ornament registry (curated ∪ extensions) and validate
	// that every preset Decoration's name resolves (RFC §14.2/§14.4, D-038). The
	// registry is read-only during the parallel compose.
	cfg.ornaments = ornaments.Curated()
	for _, ext := range cfg.ornamentExt {
		cfg.ornaments = cfg.ornaments.With(ext.name, ext.recipe)
	}
	if err := validateOrnamentRefs(s, cfg.ornaments); err != nil {
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
	// Deck core metadata → docProps/core.xml (D-042). Skipped when empty so a
	// metadata-free scene leaves the scaffold's blank core.xml untouched.
	if s.Meta != (Metadata{}) {
		pres.SetMetadata(pptx.Metadata{Title: s.Meta.Title, Author: s.Meta.Author, Subject: s.Meta.Subject})
	}

	workers := cfg.workers
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	ctx := cfg.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg.logger != nil {
		cfg.logger.Debug("scene: render start", "slides", len(s.Slides), "workers", workers)
	}

	base := &renderer{pres: pres, cfg: cfg, theme: pres.Theme(), ctx: ctx,
		chrome: s.Chrome, chromeTotal: chromeTotalFor(s)}
	// A brand-image chrome is the only chrome path that registers global media;
	// like other media-bearing slides it must compose sequentially so the brand
	// part is numbered deterministically. Brand-text chrome stays parallel.
	chromeSerial := chromeNeedsSerial(s)

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
		if err := ctx.Err(); err != nil { // honor cancellation between slides
			return Stats{}, err
		}
		if workers > 1 && !chromeSerial && !slideNeedsSerial(&s.Slides[i]) {
			free = append(free, i)
			continue
		}
		results[i] = base.composeOne(slides[i], &s.Slides[i], i)
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
				if ctx.Err() != nil { // skip remaining work once canceled
					return
				}
				results[i] = base.composeOne(slides[i], &s.Slides[i], i)
			}(idx)
		}
		wg.Wait()
		if err := ctx.Err(); err != nil {
			return Stats{}, err
		}
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
		stats.Colors = append(stats.Colors, results[i].colors)
	}
	if cfg.logger != nil {
		// Warnings are logged here, post-aggregation, so order is deterministic
		// (scene order) regardless of the parallel compose (RFC §18 — layout
		// overflows, unresolved assets).
		for _, w := range stats.Warnings {
			cfg.logger.Warn("scene: layout warning", "slide", w.SlideID, "detail", w.Message)
		}
		cfg.logger.Info("scene: render complete",
			"slides", stats.Slides,
			"shapes", stats.Shapes,
			"assets", stats.Assets,
			"warnings", len(stats.Warnings),
		)
	}
	return stats, nil
}
