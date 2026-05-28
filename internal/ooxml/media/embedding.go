package media

// ============================================================================
// EmbeddingPart - embedded object part
// ============================================================================
//
// Corresponds to /ppt/embeddings/Microsoft_Excel_WorksheetN.xlsx
// Stores embedded binary data (e.g. Excel workbooks).
//
// ============================================================================

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/hurtener/pptx-go/internal/opc"
)

// EmbeddingType identifies the kind of embedded object.
type EmbeddingType int

const (
	EmbeddingTypeUnknown EmbeddingType = iota
	EmbeddingTypeExcel                 // Excel worksheet
	EmbeddingTypeWord                  // Word document
	EmbeddingTypeOther                 // other binary
)

// EmbeddingPart holds an embedded binary file inside the package.
type EmbeddingPart struct {
	uri       *opc.PackURI
	data      []byte
	embedType EmbeddingType
	mu        sync.RWMutex
}

// NewEmbeddingPart creates a new embedding part with the given numeric ID.
func NewEmbeddingPart(id int, embedType EmbeddingType) *EmbeddingPart {
	return &EmbeddingPart{
		uri:       opc.NewPackURI(fmt.Sprintf("/ppt/embeddings/Microsoft_Excel_Worksheet%d.xlsx", id)),
		embedType: embedType,
	}
}

// NewEmbeddingPartWithURI creates an embedding part using the specified URI.
func NewEmbeddingPartWithURI(uri *opc.PackURI, embedType EmbeddingType) *EmbeddingPart {
	return &EmbeddingPart{
		uri:       uri,
		embedType: embedType,
	}
}

// PartURI returns the part URI.
func (e *EmbeddingPart) PartURI() *opc.PackURI {
	return e.uri
}

// Data returns the embedded bytes.
func (e *EmbeddingPart) Data() []byte {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.data
}

// SetData sets the embedded bytes.
func (e *EmbeddingPart) SetData(data []byte) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.data = data
}

// SetDataReader reads all bytes from r and stores them as the embedded data.
func (e *EmbeddingPart) SetDataReader(r io.Reader) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read embedding data: %w", err)
	}
	e.data = data
	return nil
}

// EmbedType returns the embedding type.
func (e *EmbeddingPart) EmbedType() EmbeddingType {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.embedType
}

// Size returns the size of the embedded data in bytes.
func (e *EmbeddingPart) Size() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.data)
}

// ============================================================================
// Data writing
// ============================================================================

// WriteTo writes the embedded data to w.
func (e *EmbeddingPart) WriteTo(w io.Writer) (int64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.data == nil {
		return 0, nil
	}

	n, err := w.Write(e.data)
	return int64(n), err
}

// Reader returns a reader over the embedded data.
func (e *EmbeddingPart) Reader() io.Reader {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return bytes.NewReader(e.data)
}

// ============================================================================
// Type detection
// ============================================================================

// DetectEmbeddingType infers the embedding type from a file name.
func DetectEmbeddingType(filename string) EmbeddingType {
	// simple extension-based detection
	if len(filename) >= 5 {
		ext := filename[len(filename)-5:]
		switch ext {
		case ".xlsx":
			return EmbeddingTypeExcel
		case ".docx":
			return EmbeddingTypeWord
		}
	}
	return EmbeddingTypeUnknown
}

// IsExcel reports whether this is an Excel embedding.
func (e *EmbeddingPart) IsExcel() bool {
	return e.embedType == EmbeddingTypeExcel
}

// IsWord reports whether this is a Word embedding.
func (e *EmbeddingPart) IsWord() bool {
	return e.embedType == EmbeddingTypeWord
}
