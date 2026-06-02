package core

import (
	"encoding/xml"
	"strings"
)

// BuildCorePropsXML returns a deterministic /docProps/core.xml carrying the
// given title / creator / subject (any may be empty). It emits no created or
// modified timestamps, so output is byte-identical for the same input
// (RFC §10.1 idempotency, D-035). Values are XML-escaped. The namespace prefixes
// match the scaffold's empty core.xml, so a metadata-bearing deck and a blank
// one differ only by the populated Dublin Core elements.
func BuildCorePropsXML(title, creator, subject string) []byte {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	b.WriteString(`<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">`)
	writeCoreEl(&b, "dc:title", title)
	writeCoreEl(&b, "dc:creator", creator)
	writeCoreEl(&b, "dc:subject", subject)
	b.WriteString(`</cp:coreProperties>`)
	return []byte(b.String())
}

func writeCoreEl(b *strings.Builder, tag, val string) {
	if val == "" {
		return
	}
	b.WriteString("<" + tag + ">")
	_ = xml.EscapeText(b, []byte(val))
	b.WriteString("</" + tag + ">")
}
