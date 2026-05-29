package pptx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hurtener/pptx-go/internal/ooxml/embeddings"
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
