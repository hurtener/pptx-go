package pptx

import (
	"github.com/hurtener/pptx-go/internal/ooxml/core"
	"github.com/hurtener/pptx-go/internal/opc"
)

// Metadata is the deck's core document properties (OPC core properties,
// docProps/core.xml). Any field may be empty.
type Metadata struct {
	Title   string
	Author  string // dc:creator
	Subject string
}

// SetMetadata writes the deck's core properties (title / author / subject) into
// docProps/core.xml. Values are XML-escaped; no created/modified timestamps are
// emitted, so output stays byte-identical for the same input (D-035). It is the
// builder API behind scene.Scene.Meta (D-042).
func (p *Presentation) SetMetadata(m Metadata) {
	part := p.pkg.GetPart(opc.NewPackURI("/docProps/core.xml"))
	if part == nil {
		return
	}
	part.SetBlob(core.BuildCorePropsXML(m.Title, m.Author, m.Subject))
}
