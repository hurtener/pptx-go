// Package scene is Layer 2 of pptx-go: a typed scene IR and a Render entrypoint
// that composes the pptx builder (Layer 1). A caller builds a Scene — an
// ordered list of SceneSlides, each a list of typed SlideNodes — and Render
// turns it into a *pptx.Presentation.
//
// scene composes pptx; it never reaches under the builder (P1). Its token
// enums are aliases of pptx's, so callers use one vocabulary.
//
// This package provides the IR catalog, two-stage validation (structural plus
// token/asset/registry resolution), the AssetResolver seam, the per-node
// rendering-policy table, the curated icon/ornament/frame registries with
// per-render caller extension, and a fully implemented, internally parallel,
// deterministic Render that composes every node kind onto the builder and
// returns Stats.
package scene
