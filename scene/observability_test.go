package scene_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestWithLogger_EmitsEvents checks the logger is actually wired (RFC §18): a
// render emits a completion summary, and an unresolved asset surfaces both in
// Stats.Warnings and as a Warn-level log event.
func TestWithLogger_EmitsEvents(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s1",
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://missing"}}, // no resolver → warning
	}}}

	stats, err := scene.Render(pptx.New(), sc, scene.WithLogger(logger))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(stats.Warnings) == 0 {
		t.Fatal("expected a LayoutWarning for the unresolved asset")
	}

	out := buf.String()
	for _, want := range []string{"render complete", "slides=1", "layout warning", "s1"} {
		if !strings.Contains(out, want) {
			t.Errorf("log output missing %q in:\n%s", want, out)
		}
	}
}

// TestWithLogger_Nil confirms the no-logger path is silent and panic-free (the
// default for every other test).
func TestWithLogger_Nil(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}}}}
	if _, err := scene.Render(pptx.New(), sc); err != nil {
		t.Fatalf("Render without a logger: %v", err)
	}
}

// TestVariant_SurfacedNotDropped checks that VariantDark is now implemented
// (no variant warning) and that VariantPrint (still unimplemented) still warns,
// while VariantLight (default) is always silent.
func TestVariant_SurfacedNotDropped(t *testing.T) {
	body := []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "light", Nodes: body},                              // default → no warning
		{ID: "dark", Variant: scene.VariantDark, Nodes: body},   // implemented → no warning
		{ID: "print", Variant: scene.VariantPrint, Nodes: body}, // unimplemented → warning
	}}
	stats, err := scene.Render(pptx.New(), sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	var printWarns int
	for _, w := range stats.Warnings {
		if w.SlideID == "print" && strings.Contains(w.Message, "variant") {
			printWarns++
		}
		if w.SlideID == "light" {
			t.Errorf("default-variant slide unexpectedly warned: %s", w.Message)
		}
		if w.SlideID == "dark" && strings.Contains(w.Message, "variant") {
			t.Errorf("VariantDark is now implemented but still warns: %s", w.Message)
		}
	}
	if printWarns != 1 {
		t.Errorf("VariantPrint warnings = %d, want 1; all warnings: %+v", printWarns, stats.Warnings)
	}
}

// TestRender_MetaReachesCoreProps is the carried-forward fix (D-042): a Scene's
// Meta (title/author/subject) is written into docProps/core.xml instead of being
// silently dropped.
func TestRender_MetaReachesCoreProps(t *testing.T) {
	sc := scene.Scene{
		Meta:   scene.Metadata{Title: "Annual Deck", Author: "Acme", Subject: "Q3"},
		Slides: []scene.SceneSlide{{ID: "s", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}}},
	}
	data, _ := render(t, sc)
	core := zipPart(t, data, "docProps/core.xml")
	for _, want := range []string{"<dc:title>Annual Deck</dc:title>", "<dc:creator>Acme</dc:creator>", "<dc:subject>Q3</dc:subject>"} {
		if !strings.Contains(core, want) {
			t.Errorf("core.xml missing %q in:\n%s", want, core)
		}
	}
}

// ctxKey is a context tag used to prove the caller's context reaches the resolver.
type ctxKey struct{}

type recordingResolver struct{ gotValue any }

func (r *recordingResolver) Resolve(ctx context.Context, _ scene.AssetID) ([]byte, string, error) {
	r.gotValue = ctx.Value(ctxKey{})
	return append([]byte("\x89PNG\r\n\x1a\n"), []byte("x")...), "image/png", nil
}

// TestWithContext_ReachesResolver checks the caller's context is threaded to the
// AssetResolver (previously hardcoded to context.Background()).
func TestWithContext_ReachesResolver(t *testing.T) {
	res := &recordingResolver{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "tagged")
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s",
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x"}},
	}}}
	if _, err := scene.Render(pptx.New(), sc, scene.WithContext(ctx), scene.WithAssetResolver(res)); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if res.gotValue != "tagged" {
		t.Errorf("resolver received context value %v, want \"tagged\" (caller ctx not threaded)", res.gotValue)
	}
}

// TestWithContext_HonorsCancellation checks a canceled context makes Render
// return the context error instead of composing.
func TestWithContext_HonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "a", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}}},
		{ID: "b", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("y")}}}},
	}}
	_, err := scene.Render(pptx.New(), sc, scene.WithContext(ctx))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Render error = %v, want context.Canceled", err)
	}
}
