package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// Speaker notes (RFC §8.8, D-022)
// ============================================================================
//
// Every slide may carry speaker notes, emitted into a notesSlideN.xml part
// related from the slide, with one notesMaster1.xml related from
// presentation.xml. V1 stores plain text; the rich-text TextFrame accessor the
// RFC sketches arrives with the rich-text model (a later phase).

const (
	notesMasterURI = "/ppt/notesMasters/notesMaster1.xml"
	// notesThemeURI is the notes master's own theme part. PowerPoint requires
	// each master to reference a distinct theme part — sharing the slide
	// master's theme1.xml repairs the deck (it splits off a theme2.xml on
	// open), so we seed one ourselves.
	notesThemeURI = "/ppt/theme/theme2.xml"
)

// SpeakerNotes returns the slide's speaker-notes text frame, creating it on
// first use (D-022, RFC §8.8). Author notes through the returned TextFrame just
// like any other text. The scene IR's SceneSlide.Notes maps onto it.
func (s *Slide) SpeakerNotes() *TextFrame {
	if s.notes == nil {
		s.notes = &TextFrame{
			s: s,
			body: &slide.XTextBody{
				BodyPr:   &slide.XBodyPr{},
				LstStyle: &slide.XTextParagraphList{},
			},
		}
	}
	return s.notes
}

// SetSpeakerNotes is a convenience that replaces the notes with a single
// plain-text paragraph (themed body text).
func (s *Slide) SetSpeakerNotes(text string) {
	tf := s.SpeakerNotes()
	tf.body.Paragraphs = nil
	tf.AddParagraph(ParagraphOpts{}).AddRun(text, RunStyle{TypeRole: TypeBody})
}

// HasSpeakerNotes reports whether the slide carries speaker notes.
func (s *Slide) HasSpeakerNotes() bool { return s.notes != nil }

// syncNotes materializes the notes parts for every slide carrying notes: a
// shared notes master (created once, wired from presentation.xml) and a
// per-slide notesSlide part wired from the slide. The slide→notesSlide
// relationship is added to the slide's own relationship set so syncSlides
// mirrors it onto the package part; syncNotes therefore runs before syncSlides.
// Callers hold p.mu.
func (p *Presentation) syncNotes() error {
	hasAny := false
	for _, s := range p.slides {
		if s.notes != nil {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return nil
	}

	p.ensureNotesMaster()

	for _, s := range p.slides {
		if s.notes == nil {
			continue
		}
		if err := p.syncNotesSlide(s); err != nil {
			return err
		}
	}
	return nil
}

// ensureNotesMaster adds notesMaster1.xml (once), wires it from
// presentation.xml, and records it in the notes-master id list.
func (p *Presentation) ensureNotesMaster() {
	uri := opc.NewPackURI(notesMasterURI)
	if p.pkg.GetPart(uri) != nil {
		return
	}

	// Seed the notes master's own theme part (theme2.xml). PowerPoint repairs a
	// deck whose notes master shares the slide master's theme1.xml, so the notes
	// master gets a distinct theme part of its own (same visual theme content).
	if p.pkg.GetPart(opc.NewPackURI(notesThemeURI)) == nil {
		_ = p.pkg.AddPart(opc.NewPart(opc.NewPackURI(notesThemeURI), opc.ContentTypeTheme, []byte(scaffoldThemeXML)))
	}

	part := opc.NewPart(uri, opc.ContentTypeNotesMaster, []byte(scaffoldNotesMasterXML))
	_, _ = part.AddRelationship(opc.RelTypeTheme, "../theme/theme2.xml", false)
	_ = p.pkg.AddPart(part)

	presPart := p.pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))
	if presPart == nil {
		return
	}
	rel, _ := presPart.AddRelationship(opc.RelTypeNotesMaster, "notesMasters/notesMaster1.xml", false)
	p.presentationPart.SetNotesMaster(rel.RID())
}

// syncNotesSlide creates (or refreshes) the notesSlide part for a slide and
// wires the slide→notesSlide relationship. The notesSlide is named after the
// slide's stable file number so it survives slide reordering.
func (p *Presentation) syncNotesSlide(s *Slide) error {
	n := s.num
	uri := opc.NewPackURI(fmt.Sprintf("/ppt/notesSlides/notesSlide%d.xml", n))

	txBody, err := slide.MarshalTextBody(s.notes.body)
	if err != nil {
		return fmt.Errorf("serialize notes for slide %d: %w", n, err)
	}
	blob := []byte(buildNotesSlideXML(string(txBody)))

	part := p.pkg.GetPart(uri)
	if part == nil {
		part = opc.NewPart(uri, opc.ContentTypeNotesSlide, blob)
		_, _ = part.AddRelationship(opc.RelTypeNotesMaster, "../notesMasters/notesMaster1.xml", false)
		_, _ = part.AddRelationship(opc.RelTypeSlide, fmt.Sprintf("../slides/slide%d.xml", n), false)
		_ = p.pkg.AddPart(part)
	} else {
		part.SetBlob(blob)
	}

	// slide → notesSlide, allocated in the slide's rId namespace (mirrored to
	// the package part by syncSlides). Dedup by target so repeated saves don't
	// add it twice.
	target := opc.NewPackURI(uri.URI())
	if s.part.Relationships().GetByTarget(target) == nil {
		_, _ = s.part.Relationships().AddNew(opc.RelTypeNotesSlide, fmt.Sprintf("../notesSlides/notesSlide%d.xml", n), false)
	}
	return nil
}

// buildNotesSlideXML renders a minimal, valid notesSlide whose body placeholder
// carries the given serialized <p:txBody> (from slide.MarshalTextBody). It is
// hand-authored namespaced OOXML (the A2 scaffold pattern).
func buildNotesSlideXML(txBody string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:notes xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr/>
<p:sp>
<p:nvSpPr><p:cNvPr id="2" name="Notes Placeholder 1"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="body" idx="1"/></p:nvPr></p:nvSpPr>
<p:spPr/>
` + txBody + `
</p:sp>
</p:spTree>
</p:cSld>
<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:notes>`
}

// scaffoldNotesMasterXML is /ppt/notesMasters/notesMaster1.xml — a minimal,
// valid notes master (color map + empty shape tree + a notes text style). It
// follows the A2 hand-authored scaffold pattern.
const scaffoldNotesMasterXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:notesMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:bg><p:bgRef idx="1001"><a:schemeClr val="bg1"/></p:bgRef></p:bg>
<p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>
</p:spTree>
</p:cSld>
<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
<p:notesStyle>
<a:lvl1pPr><a:defRPr sz="1200"><a:solidFill><a:schemeClr val="tx1"/></a:solidFill><a:latin typeface="+mn-lt"/></a:defRPr></a:lvl1pPr>
</p:notesStyle>
</p:notesMaster>`
