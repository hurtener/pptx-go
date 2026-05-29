package pptx

import (
	"fmt"
	"strings"

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
)

// SpeakerNotes sets the slide's speaker-notes text (D-022). Passing an empty
// string still marks the slide as having notes (an empty notes page). The
// scene IR's SceneSlide.Notes maps directly to this.
func (s *Slide) SpeakerNotes(text string) {
	s.notesText = text
	s.hasNotes = true
}

// HasSpeakerNotes reports whether the slide carries speaker notes.
func (s *Slide) HasSpeakerNotes() bool { return s.hasNotes }

// syncNotes materializes the notes parts for every slide carrying notes: a
// shared notes master (created once, wired from presentation.xml) and a
// per-slide notesSlide part wired from the slide. The slide→notesSlide
// relationship is added to the slide's own relationship set so syncSlides
// mirrors it onto the package part; syncNotes therefore runs before syncSlides.
// Callers hold p.mu.
func (p *Presentation) syncNotes() error {
	hasAny := false
	for _, s := range p.slides {
		if s.hasNotes {
			hasAny = true
			break
		}
	}
	if !hasAny {
		return nil
	}

	p.ensureNotesMaster()

	for _, s := range p.slides {
		if !s.hasNotes {
			continue
		}
		p.syncNotesSlide(s)
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

	part := opc.NewPart(uri, opc.ContentTypeNotesMaster, []byte(scaffoldNotesMasterXML))
	// The notes master references the deck theme (reusing theme1).
	_, _ = part.AddRelationship(opc.RelTypeTheme, "../theme/theme1.xml", false)
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
func (p *Presentation) syncNotesSlide(s *Slide) {
	n := s.num
	uri := opc.NewPackURI(fmt.Sprintf("/ppt/notesSlides/notesSlide%d.xml", n))
	blob := []byte(buildNotesSlideXML(s.notesText))

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
}

// buildNotesSlideXML renders a minimal, valid notesSlide carrying text in its
// body placeholder. It is hand-authored namespaced OOXML (the A2 scaffold
// pattern); the text is XML-escaped.
func buildNotesSlideXML(text string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:notes xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
<p:cSld>
<p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr/>
<p:sp>
<p:nvSpPr><p:cNvPr id="2" name="Notes Placeholder 1"/><p:cNvSpPr><a:spLocks noGrp="1"/></p:cNvSpPr><p:nvPr><p:ph type="body" idx="1"/></p:nvPr></p:nvSpPr>
<p:spPr/>
<p:txBody><a:bodyPr/><a:lstStyle/><a:p><a:r><a:t>` + escapeXMLText(text) + `</a:t></a:r></a:p></p:txBody>
</p:sp>
</p:spTree>
</p:cSld>
<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:notes>`
}

// escapeXMLText escapes the XML metacharacters that matter inside element text.
func escapeXMLText(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return r.Replace(s)
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
