package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

func codePNG() []byte { return append([]byte("\x89PNG\r\n\x1a\n"), []byte("code-shot")...) }

func codeResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "code1" {
			return codePNG(), "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

// TestCodeBlockBadge is Phase 16 criteria 1 & 3: a code_block with a Language
// renders an overlay badge pill containing the language; an empty Language emits
// no badge (and fewer shapes).
func TestCodeBlockBadge(t *testing.T) {
	mk := func(lang string) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "code",
			Nodes: []scene.SlideNode{scene.CodeBlock{AssetID: "asset://code1", Language: lang, Caption: "main.go"}},
		}}}
	}

	withLang, sWith := render(t, mk("go"), scene.WithAssetResolver(codeResolver()))
	_, sNone := render(t, mk(""), scene.WithAssetResolver(codeResolver()))

	xml := zipPart(t, withLang, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<p:pic>") {
		t.Errorf("code_block missing the raster pic:\n%s", xml)
	}
	if !strings.Contains(xml, `prst="roundRect"`) || !strings.Contains(xml, "<a:t>go</a:t>") {
		t.Errorf("code_block with Language missing the badge pill/text:\n%s", xml)
	}
	if !strings.Contains(xml, "<a:t>main.go</a:t>") {
		t.Errorf("code_block missing caption")
	}
	if sWith.Shapes != sNone.Shapes+2 { // badge = pill + text
		t.Errorf("badge shape delta = %d, want +2 (pill + text)", sWith.Shapes-sNone.Shapes)
	}
}

// TestCodeBlockParallel is criterion 5: byte-identical at workers=1 vs N.
func TestCodeBlockParallel(t *testing.T) {
	mk := func() scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "c",
			Nodes: []scene.SlideNode{scene.CodeBlock{AssetID: "asset://code1", Language: "rust", Caption: "lib.rs"}},
		}}}
	}
	seq, _ := render(t, mk(), scene.WithAssetResolver(codeResolver()), scene.WithWorkers(1))
	par, _ := render(t, mk(), scene.WithAssetResolver(codeResolver()), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Error("code_block render differs between workers=1 and workers=4")
	}
}
