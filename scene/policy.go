package scene

// Per-node rendering policy (RFC §12, D-011/D-018). Every node has a fixed
// policy intrinsic to its type: Native PPTX shapes, or an Image (pic shape)
// built from caller-supplied bytes. The flavor is not configurable. A node
// renders as an image iff its IR struct carries an AssetID field — policy_test
// asserts that invariant against the structs so the table can't drift.

// Policy is a node type's intrinsic rendering policy.
type Policy struct {
	// Image reports whether the node renders as a pic shape (vs native shapes).
	Image bool
	// HasAsset reports whether the node's IR carries an AssetID field.
	HasAsset bool
}

// policyTable is the §12.1 per-node policy. Containers render nothing
// themselves (their children render per their own policy).
var policyTable = map[NodeKind]Policy{
	KindHero:           {},
	KindProse:          {},
	KindHeading:        {},
	KindList:           {},
	KindDivider:        {},
	KindQuote:          {},
	KindCallout:        {},
	KindImage:          {Image: true, HasAsset: true},
	KindChip:           {},
	KindArrow:          {},
	KindCodeBlock:      {Image: true, HasAsset: true},
	KindChart:          {Image: true, HasAsset: true}, // V1 image-shape (D-004)
	KindTable:          {},
	KindFlow:           {},
	KindDecoration:     {HasAsset: true}, // asset-kind renders as an image at render time; preset is native
	KindSectionDivider: {},
	KindTwoColumn:      {},
	KindGrid:           {},
	KindCard:           {},
	KindCardSection:    {},
	KindBento:          {},
	KindStat:           {},
	KindButton:         {}, // native pill + custGeom icons (no asset)
	KindChecklist:      {}, // native text + custGeom glyphs (no asset)
	KindChipRow:        {}, // native pills + text + optional custGeom icons (no asset)
	KindBanner:         {}, // native filled strip + text + icon; children per their own policy
	KindIconRows:       {}, // native icon + text rows + optional pill frame (no asset)
}

// PolicyFor returns the rendering policy for a node kind.
func PolicyFor(k NodeKind) Policy { return policyTable[k] }
