package opc

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"strings"
)

// DefaultMaxPartBytes is the default per-part decompressed size ceiling applied
// when opening a package. A part larger than this is rejected with
// ErrPartTooLarge rather than allocated, bounding memory use on untrusted input
// (CLAUDE.md §7). Callers override via WithMaxPartBytes; a value <= 0 disables
// the bound.
const DefaultMaxPartBytes int64 = 100 << 20 // 100 MB

// ErrPartTooLarge is returned when a package part's decompressed size exceeds the
// configured per-part limit. It wraps via %w so callers can errors.Is it.
var ErrPartTooLarge = errors.New("opc: part exceeds maximum size")

// ErrUnsafePartPath is returned when a ZIP entry's path escapes the package root
// — an absolute path or one containing a ".." segment (a zip-slip attempt). Such
// entries are rejected at parse time (CLAUDE.md §7).
var ErrUnsafePartPath = errors.New("opc: unsafe part path")

// openConfig holds the tunables for opening a package.
type openConfig struct {
	maxPartBytes int64
}

// OpenOption customizes how a package is opened. Both the eager (Open/OpenFile)
// and streaming (OpenStream/OpenStreamFromReader) entry points accept these.
type OpenOption func(*openConfig)

// WithMaxPartBytes sets the per-part decompressed size ceiling. A value <= 0
// disables the bound (unlimited).
func WithMaxPartBytes(n int64) OpenOption {
	return func(c *openConfig) { c.maxPartBytes = n }
}

// resolveOpenConfig applies opts over the defaults.
func resolveOpenConfig(opts []OpenOption) openConfig {
	c := openConfig{maxPartBytes: DefaultMaxPartBytes}
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	return c
}

// safePartPath validates a NormalizeZipPath result against zip-slip: the path
// must stay within the package root. NormalizeZipPath already strips the leading
// slash, so any remaining absolute or parent-escaping component is unsafe.
func safePartPath(normalized string) error {
	if normalized == "" {
		return nil
	}
	if strings.HasPrefix(normalized, "/") {
		return fmt.Errorf("%w: %q is absolute", ErrUnsafePartPath, normalized)
	}
	for _, seg := range strings.Split(normalized, "/") {
		if seg == ".." {
			return fmt.Errorf("%w: %q escapes the package root", ErrUnsafePartPath, normalized)
		}
	}
	return nil
}

// readZipEntry reads a ZIP entry's full decompressed content, enforcing the
// per-part size bound. It checks the declared uncompressed size first (cheap),
// then guards the actual read with a hard ceiling so a lying header (a zip bomb
// whose declared size is small) cannot over-allocate. limit <= 0 is unlimited.
func readZipEntry(f *zip.File, limit int64) ([]byte, error) {
	if limit > 0 && f.UncompressedSize64 > uint64(limit) {
		return nil, fmt.Errorf("%w: %q declares %d bytes (limit %d)", ErrPartTooLarge, f.Name, f.UncompressedSize64, limit)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	if limit <= 0 {
		return io.ReadAll(rc)
	}
	// Read up to limit+1 so an entry that exceeds the bound despite a small or
	// zero declared size is still caught.
	data, err := io.ReadAll(io.LimitReader(rc, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w: %q exceeds %d bytes", ErrPartTooLarge, f.Name, limit)
	}
	return data, nil
}
