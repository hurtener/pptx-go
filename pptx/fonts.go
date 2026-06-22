package pptx

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hurtener/pptx-go/internal/ooxml/embeddings"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// Font embedding (RFC §7.6, D-019). pptx-go provides the *mechanism* to embed
// fonts; whether to embed (and which) is the caller's distribution decision.
// There is no auto-embedding: the caller registers a FontSource and calls
// EmbedFont for each face it wants shipped.

// ErrNoFontSource is returned by EmbedFont when no FontSource is registered.
var ErrNoFontSource = errors.New("pptx: no font source registered (use SetFontSource)")

// ErrFontNotFound is returned when the FontSource cannot resolve a font.
var ErrFontNotFound = errors.New("pptx: font not found")

// FontSource resolves a font name + style + weight to its raw bytes. A missing
// font returns (nil, ErrFontNotFound). Callers inject one via SetFontSource.
type FontSource interface {
	Resolve(name, style string, weight int) ([]byte, error)
}

// SetFontSource registers the FontSource used by EmbedFont.
//
// Deprecated: pass pptx.WithFontSource(src) to pptx.New instead — that is the
// documented registration path. This setter remains for post-construction use.
func (p *Presentation) SetFontSource(src FontSource) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fontSource = src
}

// EmbedFont embeds the named font face: it resolves the bytes via the
// registered FontSource, writes them as a font-data part, relates the part to
// presentation.xml, and records it in the embedded-font list so PowerPoint
// renders with it. Returns ErrNoFontSource if none is registered, or
// ErrFontNotFound if the source has no such font.
func (p *Presentation) EmbedFont(name, style string, weight int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.embedFontLocked(name, style, weight)
}

// embedFontLocked is the body of EmbedFont without the lock; the caller holds
// p.mu. It is shared by the public EmbedFont wrapper and the automatic
// embedding pass (autoEmbedFonts), which runs inside the already-locked
// prepareForWrite path.
func (p *Presentation) embedFontLocked(name, style string, weight int) error {
	if p.fontSource == nil {
		return ErrNoFontSource
	}
	data, err := p.fontSource.Resolve(name, style, weight)
	if err != nil {
		return fmt.Errorf("resolve font %q: %w", name, err)
	}
	if len(data) == 0 {
		return fmt.Errorf("font %q: %w", name, ErrFontNotFound)
	}

	p.fontCounter++
	n := int(p.fontCounter)
	uri := opc.NewPackURI(embeddings.FontPartURI(n))
	if _, err := p.pkg.CreatePart(uri, embeddings.ContentTypeFontData, data); err != nil {
		return fmt.Errorf("create font part: %w", err)
	}

	presPart := p.ensurePresentationOPCPart()
	rel, err := presPart.AddRelationship(embeddings.RelTypeFont, embeddings.FontRelTarget(n), false)
	if err != nil {
		return fmt.Errorf("relate font part: %w", err)
	}

	italic := strings.EqualFold(style, "italic") || strings.EqualFold(style, "oblique")
	p.presentationPart.AddEmbeddedFont(name, embeddings.StyleFor(weight, italic), rel.RID())
	return nil
}

// autoEmbedFonts is the opt-in save-time pass (R9.1, D-065): it walks every
// slide's runs, collects the distinct used faces in a stable sorted order, and
// embeds each via the registered FontSource. It is a no-op unless
// WithFontEmbedding is set and a FontSource is registered, so the default
// output is byte-identical. A face the source cannot resolve is warned (not
// fatal); a face already embedded (e.g. by a manual EmbedFont) is skipped, so
// the pass is idempotent. The caller holds p.mu (it runs inside
// prepareForWrite).
func (p *Presentation) autoEmbedFonts() {
	if !p.fontEmbedding || p.fontSource == nil {
		return
	}

	// Collect the distinct faces across all slides into a set, then sort so the
	// font parts and relationship ids are byte-identical regardless of slide
	// render order or worker count.
	seen := map[slide.FontFace]bool{}
	var faces []slide.FontFace
	for _, s := range p.slides {
		if s == nil || s.part == nil {
			continue
		}
		for _, f := range s.part.UsedFontFaces() {
			if seen[f] {
				continue
			}
			seen[f] = true
			faces = append(faces, f)
		}
	}
	sort.Slice(faces, func(i, j int) bool {
		a, b := faces[i], faces[j]
		if a.Typeface != b.Typeface {
			return a.Typeface < b.Typeface
		}
		if a.Bold != b.Bold {
			return !a.Bold // regular before bold
		}
		return !a.Italic && b.Italic // upright before italic
	})

	for _, f := range faces {
		weight := 400
		if f.Bold {
			weight = 700
		}
		style := ""
		if f.Italic {
			style = "italic"
		}
		// Skip a face already recorded (manual EmbedFont) — keep the pass
		// idempotent. The slot key matches what AddEmbeddedFont records.
		if p.presentationPart.HasEmbeddedFace(f.Typeface, embeddings.StyleFor(weight, f.Italic)) {
			continue
		}
		if err := p.embedFontLocked(f.Typeface, style, weight); err != nil {
			if p.logger != nil {
				p.logger.Warn("pptx: font embedding skipped face",
					"family", f.Typeface, "bold", f.Bold, "italic", f.Italic, "err", err)
			}
			continue
		}
	}
}

// ensurePresentationOPCPart returns the /ppt/presentation.xml OPC part,
// creating it (from the current presentation model) if it is not yet in the
// package, so relationships can be attached before Save. Caller holds p.mu.
func (p *Presentation) ensurePresentationOPCPart() *opc.Part {
	uri := opc.NewPackURI("/ppt/presentation.xml")
	if part := p.pkg.GetPart(uri); part != nil {
		return part
	}
	blob, _ := p.presentationPart.ToXML()
	part := opc.NewPart(uri, opc.ContentTypePresentation, blob)
	_ = p.pkg.AddPart(part)
	return part
}
