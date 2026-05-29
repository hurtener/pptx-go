// Package layout is a placeholder for the scene layout engine (RFC §10.2): a
// deterministic, priority-ordered placement algorithm that assigns each
// top-level node a slot and resolves it to a pptx.Box. The engine lands in
// later phases (two_column, grid, card, flow, decoration); this package exists
// now so the scene package layout reaches the §3 tree shape incrementally.
package layout
