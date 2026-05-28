package parts

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// XMLDeclaration is the standard XML declaration header used in all OPC package files.
const XMLDeclaration = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`

// StripNamespacePrefixes pre-processes XML data by removing namespace prefixes so
// that Go's xml.Unmarshal can handle it. Go's decoder cannot handle prefixed
// elements such as <p:presentation>; this function converts <p:xxx> to <xxx>
// and transforms prefixed attributes to their prefix-concatenated form
// (e.g. r:id becomes rid).
func StripNamespacePrefixes(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(data))

	// global namespace URI -> prefix map (accumulated across all elements)
	nsToPrefix := make(map[string]string)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("XML token error: %w", err)
		}

		switch v := token.(type) {
		case xml.StartElement:
			// collect xmlns declarations to build the URI -> prefix map
			for _, attr := range v.Attr {
				if attr.Name.Space == "xmlns" {
					// xmlns:p="..." -> prefix="p", URI=attr.Value
					nsToPrefix[attr.Value] = attr.Name.Local
				} else if attr.Name.Local == "xmlns" {
					// xmlns="..." -> default namespace, empty prefix
					nsToPrefix[attr.Value] = ""
				}
			}

			// strip the element name prefix (e.g. p:presentation -> presentation)
			buf.WriteString("<")
			buf.WriteString(v.Name.Local)

			// process attributes: convert r:id to rid (drop the colon), drop xmlns declarations
			for _, attr := range v.Attr {
				// skip xmlns declarations
				if attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns" {
					continue
				}
				buf.WriteString(" ")

				// if the attribute has a namespace, look up its prefix and prepend it (no colon)
				if attr.Name.Space != "" {
					if prefix, ok := nsToPrefix[attr.Name.Space]; ok && prefix != "" {
						buf.WriteString(prefix)
						// intentionally no colon: r:id -> rid
					}
				}
				buf.WriteString(attr.Name.Local)
				buf.WriteString("=\"")
				buf.WriteString(attr.Value)
				buf.WriteString("\"")
			}
			buf.WriteString(">")
		case xml.EndElement:
			buf.WriteString("</")
			buf.WriteString(v.Name.Local)
			buf.WriteString(">")
		case xml.CharData:
			buf.Write(v)
		case xml.Comment:
			buf.WriteString("<!--")
			buf.Write(v)
			buf.WriteString("-->")
		case xml.ProcInst:
			// skip the XML declaration
			if v.Target == "xml" {
				continue
			}
		}
	}

	return buf.Bytes(), nil
}
