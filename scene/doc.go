// Package scene is Layer 2 of pptx-go: a typed scene IR and a Render entrypoint
// that composes the pptx builder (Layer 1). A caller builds a Scene — an
// ordered list of SceneSlides, each a list of typed SlideNodes — and Render
// turns it into a *pptx.Presentation.
//
// scene composes pptx; it never reaches under the builder (P1). Its token
// enums are aliases of pptx's, so callers use one vocabulary.
//
// This package currently provides the IR catalog, Stage 1 structural
// validation, the AssetResolver seam, and the per-node rendering-policy table.
// Render is a no-op stub here; rendering lands in later phases.
package scene
