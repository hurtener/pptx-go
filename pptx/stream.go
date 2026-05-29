package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/opc"
)

// ============================================================================
// Streaming I/O (RFC §9, §17.2)
// ============================================================================
//
// OpenStream / SaveStream are the streaming I/O mode: large decks open with
// lazy per-part loading and save through the OPC streaming writer, without
// buffering the whole package as one byte slice. Streaming is an I/O mode, not
// a persistence layer (§9) — the in-memory model is identical to the eager
// path, and the always-on repair-prompt hygiene pass (D-020) runs on the
// streaming write path too.
//
// The RFC §9 signatures are path-based (OpenStream(path), SaveStream(path)); we
// match them. (CLAUDE.md §5's context-first convention yields to the explicit
// RFC signature here — see the phase plan's deviations log.)

// OpenStream opens a .pptx file using the OPC streaming reader (lazy per-part
// loading) and returns a Presentation. (RFC §9, §17.2.)
func OpenStream(path string) (*Presentation, error) {
	sp, err := opc.OpenStream(path)
	if err != nil {
		return nil, fmt.Errorf("open stream %q: %w", path, err)
	}
	defer func() { _ = sp.Close() }()

	pkg, err := packageFromStream(sp)
	if err != nil {
		return nil, err
	}

	pres := &Presentation{
		pkg:           pkg,
		slides:        make([]*Slide, 0),
		mediaManager:  NewMediaManager(),
		masterManager: NewMasterManager(),
		theme:         DefaultTheme(),
	}
	if err := pres.loadPresentationPart(); err != nil {
		return nil, fmt.Errorf("parse presentation part: %w", err)
	}
	return pres, nil
}

// SaveStream serializes the presentation to a file through the OPC streaming
// writer. It applies the same syncing and repair-prompt hygiene (D-020) as
// Save. (RFC §9, §17.2.)
func (p *Presentation) SaveStream(path string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if err := p.prepareForWrite(); err != nil {
		return err
	}

	sp, err := streamPackageFromPackage(p.pkg)
	if err != nil {
		return err
	}
	return sp.StreamSaveFile(path)
}

// packageFromStream materializes an in-memory opc.Package from a streaming
// package, loading each part's content lazily through the stream source. This
// keeps OpenStream's read path bounded (one part at a time) while presenting
// the standard package model the builder works against.
func packageFromStream(sp *opc.StreamPackage) (*opc.Package, error) {
	pkg := opc.NewPackage()

	// Carry content types across (defaults are pre-seeded; overrides matter).
	for ext, ct := range sp.ContentTypes().Defaults() {
		pkg.ContentTypes().AddDefault(ext, ct)
	}
	for uri, ct := range sp.ContentTypes().Overrides() {
		pkg.ContentTypes().AddOverride(opc.NewPackURI(uri), ct)
	}

	for _, part := range sp.AllParts() {
		blob, err := part.Blob() // lazy-loads this part's content
		if err != nil {
			return nil, fmt.Errorf("load part %s: %w", part.PartURI().URI(), err)
		}
		np := opc.NewPart(part.PartURI(), part.ContentType(), blob)
		if part.HasRelationships() {
			if rb, err := part.RelationshipsBlob(); err == nil && rb != nil {
				_ = np.LoadRelationships(rb)
			}
		}
		if err := pkg.AddPart(np); err != nil {
			return nil, fmt.Errorf("add part %s: %w", part.PartURI().URI(), err)
		}
	}

	// Package-level relationships (the rId is not referenced from any part's
	// XML, so re-allocation is safe).
	for _, rel := range sp.Relationships().All() {
		_, _ = pkg.AddRelationship(rel.Type(), rel.TargetRef(), rel.IsExternal())
	}
	return pkg, nil
}

// streamPackageFromPackage mirrors an in-memory package into a streaming
// package so SaveStream can serialize through the OPC streaming writer.
func streamPackageFromPackage(pkg *opc.Package) (*opc.StreamPackage, error) {
	sp := opc.NewStreamPackage()

	for ext, ct := range pkg.ContentTypes().Defaults() {
		sp.ContentTypes().AddDefault(ext, ct)
	}
	for uri, ct := range pkg.ContentTypes().Overrides() {
		sp.ContentTypes().AddOverride(opc.NewPackURI(uri), ct)
	}

	for _, part := range pkg.AllParts() {
		np, err := sp.CreatePartFromBytes(part.PartURI(), part.ContentType(), part.Blob())
		if err != nil {
			return nil, fmt.Errorf("stream part %s: %w", part.PartURI().URI(), err)
		}
		if part.HasRelationships() {
			if rb, err := part.RelationshipsBlob(); err == nil && rb != nil {
				_ = np.LoadRelationships(rb)
			}
		}
	}

	for _, rel := range pkg.Relationships().All() {
		_, _ = sp.AddRelationship(rel.Type(), rel.TargetRef(), rel.IsExternal())
	}
	return sp, nil
}
