package scene_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// bigScene builds an n-slide text-only (media-free) scene for determinism and
// parallelism checks.
func bigScene(n int) scene.Scene {
	sc := scene.Scene{}
	for i := 0; i < n; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{
				scene.Heading{Text: rt("Section"), Level: 2},
				scene.Prose{Paragraphs: []scene.RichText{rt("Body text one."), rt("Body text two.")}},
				scene.List{Items: []scene.ListItem{{Text: rt("a")}, {Text: rt("b")}}},
			},
			Notes: rt("speaker notes"),
		})
	}
	return sc
}

func renderBytes(t *testing.T, sc scene.Scene, opts ...scene.RenderOption) []byte {
	t.Helper()
	pres := pptx.New()
	if _, err := scene.Render(pres, sc, opts...); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestRenderDeterministic_ParallelMatchesSequential is the D-015 idempotency
// guard: the parallel worker pool must produce byte-identical output to a
// single-worker render, and successive parallel renders must agree (RFC §10.1).
func TestRenderDeterministic_ParallelMatchesSequential(t *testing.T) {
	sc := bigScene(40)

	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("parallel render (%d bytes) differs from sequential render (%d bytes)", len(par), len(seq))
	}

	// Default (GOMAXPROCS) workers, run twice — must be stable and match sequential.
	a := renderBytes(t, sc)
	b := renderBytes(t, sc)
	if !bytes.Equal(a, b) {
		t.Fatal("two default-worker renders are not byte-identical")
	}
	if !bytes.Equal(a, seq) {
		t.Fatal("default-worker render differs from sequential render")
	}
}

// TestRenderDeterministic_MultiLineWrap is the Phase-22 determinism guard: a
// deck with paragraphs that wrap to several lines (content-aware preferredHeight)
// must render byte-identically across worker counts — the wrapped-line estimate
// is pure integer math, so layout never depends on scheduling (RFC §10.1).
func TestRenderDeterministic_MultiLineWrap(t *testing.T) {
	long := strings.TrimSpace(strings.Repeat("lorem ipsum dolor sit amet ", 18))
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{
				scene.Heading{Text: rt(long), Level: 1},
				scene.Prose{Paragraphs: []scene.RichText{rt(long), rt(long)}},
				scene.List{Items: []scene.ListItem{{Text: rt(long)}, {Text: rt(long)}}},
				scene.Quote{Text: rt(long)},
				scene.Callout{Body: rt(long)},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("multi-line wrap: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_VAlignFill is the Phase-23 determinism guard: a deck
// that grows flexible nodes to fill the frame (VAlignFill) must render
// byte-identically across worker counts — the slack distribution is pure integer
// math with a fixed remainder rule, so layout never depends on scheduling.
func TestRenderDeterministic_VAlignFill(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:      string(rune('A' + (i % 26))),
			Content: scene.Alignment{Vertical: scene.VAlignFill},
			Nodes: []scene.SlideNode{
				scene.Heading{Text: rt("Section"), Level: 1},
				scene.Grid{Columns: 2, Cells: []scene.SlideNode{
					scene.Card{Header: "a"}, scene.Card{Header: "b"},
					scene.Card{Header: "c"}, scene.Card{Header: "d"},
				}},
				scene.TwoColumn{Left: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("l")}}}, Right: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("r")}}}},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("VAlignFill: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_VAlignFit guards the R10.2 fit-to-region compression:
// a deck of deliberately over-full VAlignFit slides must render byte-identically
// at 1 and 8 workers (the compression is integer/basis-point math, worker-count
// independent).
func TestRenderDeterministic_VAlignFit(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:      string(rune('A' + (i % 26))),
			Content: scene.Alignment{Vertical: scene.VAlignFit},
			Nodes: []scene.SlideNode{
				scene.Heading{Text: rt("Section"), Level: 1},
				scene.Prose{Paragraphs: []scene.RichText{rt("p1"), rt("p2"), rt("p3")}},
				scene.List{Items: []scene.ListItem{{Text: rt("a")}, {Text: rt("b")}, {Text: rt("c")}}},
				scene.Grid{Columns: 2, Cells: []scene.SlideNode{
					scene.Card{Header: "a"}, scene.Card{Header: "b"},
					scene.Card{Header: "c"}, scene.Card{Header: "d"},
				}},
				scene.Callout{Body: rt("note")},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("VAlignFit: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_AutoContrast guards the R11.2 auto-contrast path: a deck
// mixing dark-variant slides, dark card fills, accent header bands, and join badges
// (every onCardSurface / accentLegible branch) must render byte-identically across
// worker counts — the luminance decision is pure integer math (the sRGB table is
// built once at init), so the chosen color never depends on scheduling.
func TestRenderDeterministic_AutoContrast(t *testing.T) {
	accent := scene.ColorAccent
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		variant := scene.VariantLight
		if i%2 == 0 {
			variant = scene.VariantDark
		}
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:      string(rune('A' + (i % 26))),
			Variant: variant,
			Nodes: []scene.SlideNode{
				scene.Card{Eyebrow: "VISION", Header: "Replies. Then waits.", HeaderFill: &accent,
					Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}},
				scene.Card{Header: "On dark", Fill: scene.ColorAccent, HeaderPill: "NEW",
					Body: []scene.SlideNode{scene.Stat{Value: "$4,000", Label: "per month"}}},
				scene.TwoColumn{Join: scene.JoinBadge, JoinLabel: "vs",
					Left:  []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("l")}}},
					Right: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("r")}}}},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("auto-contrast: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_BoundsClamp guards the R11.3 safe-area clamp: a deck of
// over-tall Bentos / Grids (whose slots overflow the safe area) must render
// byte-identically across worker counts — the clamp is a pure integer cap, so the
// drawn geometry never depends on scheduling.
func TestRenderDeterministic_BoundsClamp(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		var rows []scene.BentoRow
		for j := 0; j < 8; j++ {
			rows = append(rows, scene.BentoRow{Cells: []scene.BentoCell{
				{Span: 1, Node: scene.Card{Header: "l", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}}},
				{Span: 1, Node: scene.Card{Header: "r", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}}},
			}})
		}
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:    string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{scene.Bento{Columns: 2, Rows: rows}},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("bounds clamp: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_PillFit guards R11.5: cards with varied-length header
// pills (some long enough to clamp and shrink via FontScale) in narrow grid cells
// must render byte-identically across worker counts — cardPillWidthOf and fitScale
// are pure integer math, so the pill width and font scale never depend on
// scheduling.
func TestRenderDeterministic_PillFit(t *testing.T) {
	pills := []string{"NEW", "CUSTOMIZABLE", "FULLY CUSTOMIZABLE", "BETA"}
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		var cells []scene.SlideNode
		for j := 0; j < 4; j++ {
			cells = append(cells, scene.Card{Header: "Plan", HeaderPill: pills[(i+j)%len(pills)]})
		}
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:    string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{scene.Grid{Columns: 4, Cells: cells}},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("pill fit: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_JoinBadgeFit guards R11.7: TwoColumns with varied-length
// join labels (some grown, some capped + shrunk) must render byte-identically across
// worker counts — the badge diameter and label scale are pure integer math.
func TestRenderDeterministic_JoinBadgeFit(t *testing.T) {
	labels := []string{"vs", "One agent", "an extremely long join connector label", "or"}
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{scene.TwoColumn{
				Join: scene.JoinBadge, JoinLabel: labels[i%len(labels)],
				Left:  []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("l")}}},
				Right: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("r")}}},
			}},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("join badge fit: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_VAlignFillCapped guards the R10.6 capped fill: a deck of
// sparse capped-fill slides must render byte-identically across worker counts (the
// growth cap and the balanced-spacing residual are integer / basis-point math).
func TestRenderDeterministic_VAlignFillCapped(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:      string(rune('A' + (i % 26))),
			Content: scene.Alignment{Vertical: scene.VAlignFillCapped},
			Nodes: []scene.SlideNode{
				scene.Heading{Text: rt("Section"), Level: 1},
				scene.Grid{Columns: 2, Cells: []scene.SlideNode{
					scene.Card{Header: "a"}, scene.Card{Header: "b"},
					scene.Card{Header: "c"}, scene.Card{Header: "d"},
				}},
				scene.Card{Header: "sparse"},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("VAlignFillCapped: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_VAlignBalanced guards the R10.8 balanced rhythm: a deck
// of sparse balanced slides must render byte-identically across worker counts (the
// even-rhythm + optical-bias math is integer / basis point).
func TestRenderDeterministic_VAlignBalanced(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 24; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:      string(rune('A' + (i % 26))),
			Content: scene.Alignment{Vertical: scene.VAlignBalanced},
			Nodes: []scene.SlideNode{
				scene.Hero{Eyebrow: "FY26", Title: "Cover"},
				scene.Heading{Text: rt("Subtitle"), Level: 2},
				scene.Prose{Paragraphs: []scene.RichText{rt("A short description line.")}},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("VAlignBalanced: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_AutoFit (checkpoint NH8) guards the only float64 path in
// Wave 10: an AutoFit deck (FontScale → @sz) must render byte-identically across
// worker counts.
func TestRenderDeterministic_AutoFit(t *testing.T) {
	long := strings.Repeat("8", 60)
	sc := scene.Scene{}
	for i := 0; i < 16; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{
				scene.Hero{Title: long, AutoFit: true},
				scene.Stat{Value: long, AutoFit: true},
				scene.Heading{Text: rt(long), Level: 1, AutoFit: true},
			},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("AutoFit: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_WeightedBento (checkpoint NH8): a content-weighted bento
// deck renders byte-identically across worker counts.
func TestRenderDeterministic_WeightedBento(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 16; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + (i % 26))),
			Nodes: []scene.SlideNode{scene.Bento{Columns: 3, WeightedRows: true, Rows: []scene.BentoRow{
				{Label: "R1", Cells: []scene.BentoCell{{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}, {Span: 1, Node: scene.Card{Header: "c"}}}},
				{Label: "R2", Cells: []scene.BentoCell{{Span: 1, Node: scene.List{Items: []scene.ListItem{{Text: rt("a")}, {Text: rt("b")}, {Text: rt("c")}, {Text: rt("d")}}}}}},
			}}},
		})
	}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("WeightedBento: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}

// TestRenderDeterministic_WithAssets guards determinism when a media-bearing
// node (code_block) is mixed into a multi-slide deck: those slides render
// sequentially, so distinct image parts are numbered in scene order every run.
func TestRenderDeterministic_WithAssets(t *testing.T) {
	resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		// One unique PNG per asset id, so part numbering order is observable.
		png := append([]byte("\x89PNG\r\n\x1a\n"), []byte(uuid)...)
		return png, "image/png", nil
	})

	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "a", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("intro")}}}},
		{ID: "b", Nodes: []scene.SlideNode{scene.CodeBlock{AssetID: "asset://one", Language: "go"}}},
		{ID: "c", Nodes: []scene.SlideNode{scene.Heading{Text: rt("middle"), Level: 1}}},
		{ID: "d", Nodes: []scene.SlideNode{scene.CodeBlock{AssetID: "asset://two", Language: "go"}}},
	}}

	a := renderBytes(t, sc, scene.WithAssetResolver(resolver))
	b := renderBytes(t, sc, scene.WithAssetResolver(resolver))
	if !bytes.Equal(a, b) {
		t.Fatal("renders with mixed media-bearing slides are not byte-identical")
	}
	seq := renderBytes(t, sc, scene.WithAssetResolver(resolver), scene.WithWorkers(1))
	if !bytes.Equal(a, seq) {
		t.Fatal("parallel render differs from sequential render for a media-bearing deck")
	}
}

// TestConcurrentThemeReuse shares one *pptx.Theme across many simultaneous
// parallel renders. Under -race this proves the theme is safe for concurrent
// reuse (§14) — the worker pool reads it from every slide goroutine — and that
// each render is internally consistent.
func TestConcurrentThemeReuse(t *testing.T) {
	theme := pptx.DefaultTheme()
	sc := bigScene(16)
	sc.Theme = theme

	const renders = 8
	var wg sync.WaitGroup
	out := make([][]byte, renders)
	for i := 0; i < renders; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pres := pptx.New(pptx.WithTheme(theme))
			if _, err := scene.Render(pres, sc); err != nil {
				t.Errorf("render %d: %v", i, err)
				return
			}
			b, err := pres.WriteToBytes()
			if err != nil {
				t.Errorf("write %d: %v", i, err)
				return
			}
			out[i] = b
		}(i)
	}
	wg.Wait()

	for i := 1; i < renders; i++ {
		if !bytes.Equal(out[0], out[i]) {
			t.Fatalf("concurrent render %d differs from render 0 under a shared theme", i)
		}
	}
}

// TestWithWorkers_StatsStable checks the aggregate Stats are independent of the
// worker count and that per-slide timings are reported in scene order.
func TestWithWorkers_StatsStable(t *testing.T) {
	sc := bigScene(12)
	pres1 := pptx.New()
	s1, err := scene.Render(pres1, sc, scene.WithWorkers(1))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	pres8 := pptx.New()
	s8, err := scene.Render(pres8, sc, scene.WithWorkers(8))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if s1.Slides != s8.Slides || s1.Shapes != s8.Shapes || s1.Assets != s8.Assets {
		t.Fatalf("stats differ by worker count: seq=%+v par=%+v", s1, s8)
	}
	if len(s8.Timings) != len(sc.Slides) {
		t.Fatalf("Timings len = %d, want %d", len(s8.Timings), len(sc.Slides))
	}
	for i, ti := range s8.Timings {
		if ti.SlideID != sc.Slides[i].ID {
			t.Fatalf("Timings[%d].SlideID = %q, want %q (must be scene order)", i, ti.SlideID, sc.Slides[i].ID)
		}
	}
}
