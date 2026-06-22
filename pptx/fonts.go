package pptx

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/hurtener/pptx-go/internal/ooxml/embeddings"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// Font embedding (RFC §7.6, D-019, D-065). pptx-go provides the *mechanism* to
// embed fonts; whether to embed (and which) is the caller's distribution
// decision. A caller embeds one face at a time with EmbedFont, or opts into the
// automatic save-time pass (WithFontEmbedding → autoEmbedFonts, D-065), which
// embeds every face a deck actually uses via the registered FontSource. With no
// FontSource registered nothing is embedded.

// ErrNoFontSource is returned by EmbedFont when no FontSource is registered.
var ErrNoFontSource = errors.New("pptx: no font source registered (use SetFontSource)")

// ErrFontNotFound is returned when the FontSource cannot resolve a font.
var ErrFontNotFound = errors.New("pptx: font not found")

// FontSource resolves a font name + style + weight to its raw bytes. A missing
// font returns (nil, ErrFontNotFound). Callers inject one via SetFontSource or
// pptx.WithFontSource.
//
// The returned bytes are embedded **verbatim** — pptx-go applies no size cap and
// no signature/format validation (the caller's responsibility, parallel to image
// and SVG bytes under CLAUDE.md §7). A FontSource may be invoked from the save
// path (the automatic embedding/fallback passes) and, when shared across several
// presentations saved concurrently, must be safe for concurrent use.
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

	// Atomic, matching slideCounter/relCounter/chartCounter — embedFontLocked is
	// reached both from EmbedFont (under p.mu.Lock) and from autoEmbedFonts in the
	// save path, so the increment must not be a bare read-modify-write.
	n := int(atomic.AddInt32(&p.fontCounter, 1))
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

// embedBucket identifies one OOXML embeddedFont slot: a typeface and whether the
// cut is bold and/or italic (the four slots regular/bold/italic/boldItalic).
type embedBucket struct {
	family string
	bold   bool
	italic bool
}

// nominalWeight returns the canonical weight for a bucket — 700 for a bold
// bucket, 400 otherwise — used to pick the nearest used weight when several map
// to the same slot.
func (b embedBucket) nominalWeight() int {
	if b.bold {
		return 700
	}
	return 400
}

// autoEmbedFonts is the opt-in save-time pass (R9.1, D-065) with weight-aware
// file selection (R9.8, D-068): it walks every slide's runs, collects the
// distinct used faces (family, weight, italic), and embeds — per OOXML
// embeddedFont bucket (the four regular/bold/italic/boldItalic slots) — the
// actual weighted file nearest the bucket's nominal weight, so a soul's medium
// (500) regular role ships the medium file rather than a synthetic 400. It is a
// no-op unless WithFontEmbedding is set and a FontSource is registered, so the
// default output is byte-identical. A face the source cannot resolve is warned
// (not fatal); a bucket already recorded (e.g. by a manual EmbedFont) is skipped,
// so the pass is idempotent. When several used weights collide on one bucket the
// extra ones are coalesced (PowerPoint exposes only four cuts per family) and
// logged. The caller holds p.mu (it runs inside prepareForWrite).
func (p *Presentation) autoEmbedFonts() {
	if !p.fontEmbedding || p.fontSource == nil {
		return
	}

	// Collect the distinct (family, weight, italic) faces across all slides.
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

	// Group into OOXML buckets; per bucket pick the weight nearest the nominal
	// (ties → the lower weight) and count how many distinct weights collided.
	type pick struct {
		weight int
		count  int
	}
	buckets := map[embedBucket]*pick{}
	for _, f := range faces {
		b := embedBucket{family: f.Typeface, bold: f.Weight >= 600, italic: f.Italic}
		cur, ok := buckets[b]
		if !ok {
			buckets[b] = &pick{weight: f.Weight, count: 1}
			continue
		}
		cur.count++
		nominal := b.nominalWeight()
		better := abs(f.Weight-nominal) < abs(cur.weight-nominal) ||
			(abs(f.Weight-nominal) == abs(cur.weight-nominal) && f.Weight < cur.weight)
		if better {
			cur.weight = f.Weight
		}
	}

	// Embed in a deterministic order so font parts and rel ids are byte-identical
	// regardless of slide render order or worker count.
	keys := make([]embedBucket, 0, len(buckets))
	for b := range buckets {
		keys = append(keys, b)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.family != b.family {
			return a.family < b.family
		}
		if a.bold != b.bold {
			return !a.bold // regular before bold
		}
		return !a.italic && b.italic // upright before italic
	})

	for _, b := range keys {
		pk := buckets[b]
		style := ""
		if b.italic {
			style = "italic"
		}
		// Skip a bucket already recorded (manual EmbedFont) — keep the pass
		// idempotent. The slot key matches what AddEmbeddedFont records.
		if p.presentationPart.HasEmbeddedFace(b.family, embeddings.StyleFor(pk.weight, b.italic)) {
			continue
		}
		if err := p.embedFontLocked(b.family, style, pk.weight); err != nil {
			if p.logger != nil {
				p.logger.Warn("pptx: font embedding skipped face",
					"family", b.family, "bold", b.bold, "italic", b.italic, "weight", pk.weight, "err", err)
			}
			continue
		}
		if pk.count > 1 && p.logger != nil {
			p.logger.Debug("pptx: font weights coalesced into one embedded cut",
				"family", b.family, "bold", b.bold, "italic", b.italic,
				"embedded", pk.weight, "weights", pk.count)
		}
	}
}

// abs returns the absolute value of x.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// fallbackRoles is the fixed iteration order for building the fallback
// resolution map, so the map (and any substitution) is deterministic.
var fallbackRoles = []TypeRole{
	TypeDisplay, TypeH1, TypeH2, TypeH3, TypeH4, TypeH5,
	TypeBody, TypeBodySmall, TypeCaption, TypeMono, TypeCode,
}

// fallbackKey identifies a resolved face by its declared family and whether the
// run is italic. Fallback is realized per (family, italic) so an italic run can
// degrade to a different fallback than an upright one when only the italic cut is
// missing (R9.7).
type fallbackKey struct {
	family string
	italic bool
}

// resolveFontFallbacks realizes the declared per-role fallback chains (R9.6,
// D-066) with italic-cut awareness (R9.7, D-067). When a FontSource is
// registered and a role's primary cut cannot be resolved by it, the run's
// single-valued a:latin typeface is rewritten to the first family in [Family] +
// Fallback whose matching cut (italic for an italic run, regular otherwise) the
// source can resolve — so output degrades to a controlled near-match (and an
// italic emphasis run keeps a real italic cut) instead of an arbitrary host
// default or a faux-italic. It is a no-op (and zero FontSource calls,
// byte-identical) when no FontSource is registered or no role declares a
// Fallback. It runs before syncSlides so the serialized runs carry the resolved
// face, and before autoEmbedFonts so the embedded bytes match. The caller holds
// p.mu (the save path holds p.mu.Lock, so the run rewrite is exclusive).
//
// The substitution mutates the in-memory runs in place: after a save that
// substitutes a face, Run.Font reports the resolved (fallback) family, not the
// original. This is intentional — the run carries its realized face — and keeps
// the pass idempotent (a second save finds the fallback family, not a primary
// key, and rewrites nothing).
func (p *Presentation) resolveFontFallbacks() {
	if p.fontSource == nil {
		return
	}
	theme := p.theme
	if theme == nil {
		theme = DefaultTheme()
	}

	// Cheap early-out: nothing to do unless some role declares a fallback chain.
	declared := false
	for _, role := range fallbackRoles {
		if len(theme.ResolveType(role).Fallback) > 0 {
			declared = true
			break
		}
	}
	if !declared {
		return
	}

	// Probe the source for cut availability, memoized. A family is "available"
	// for a cut when that cut resolves: the italic cut for an italic run, the
	// regular cut otherwise.
	avail := map[fallbackKey]bool{}
	resolvable := func(family string, italic bool) bool {
		if family == "" {
			return false
		}
		k := fallbackKey{family, italic}
		if v, ok := avail[k]; ok {
			return v
		}
		style := ""
		if italic {
			style = "italic"
		}
		data, err := p.fontSource.Resolve(family, style, 400)
		ok := err == nil && len(data) > 0
		avail[k] = ok
		return ok
	}

	// Build (family, italic) -> resolved face, first-seen (lowest role) wins.
	mapping := map[fallbackKey]string{}
	for _, role := range fallbackRoles {
		spec := theme.ResolveType(role)
		if len(spec.Fallback) == 0 || spec.Family == "" {
			continue
		}
		for _, italic := range []bool{false, true} {
			key := fallbackKey{spec.Family, italic}
			if _, done := mapping[key]; done {
				continue
			}
			if resolvable(spec.Family, italic) {
				continue // primary cut available — it wins, no substitution
			}
			for _, fb := range spec.Fallback {
				if fb != spec.Family && resolvable(fb, italic) {
					mapping[key] = fb
					break
				}
			}
		}
	}
	if len(mapping) == 0 {
		return
	}

	resolve := func(typeface string, _ /*bold*/, italic bool) string {
		return mapping[fallbackKey{typeface, italic}]
	}
	rewritten := 0
	for _, s := range p.slides {
		if s == nil || s.part == nil {
			continue
		}
		rewritten += s.part.RewriteFontFaces(resolve)
	}
	if p.logger != nil && rewritten > 0 {
		p.logger.Debug("pptx: font fallback substituted faces",
			"faces", len(mapping), "runs", rewritten)
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
