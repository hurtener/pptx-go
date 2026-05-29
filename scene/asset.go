package scene

import (
	"context"
	"errors"
	"strings"
)

// Assets enter the IR by reference (an AssetID), not by value. An AssetResolver
// maps the id to bytes at render time. pptx-go never rasterizes — the caller
// pre-rasterizes and supplies bytes (§10.6, §12.3, D-024).

// ErrAssetNotFound is returned by a resolver when an id has no bytes.
var ErrAssetNotFound = errors.New("scene: asset not found")

// AssetID is a free-form asset reference. pptx-go imposes no scheme; callers
// choose (pengui-slides uses asset://<uuid> — see URIAssetResolver). (D-024.)
type AssetID string

// AssetResolver maps an AssetID to bytes and a content-type hint
// (image/png, image/jpeg, image/svg+xml, …). A missing asset returns
// (nil, "", ErrAssetNotFound).
type AssetResolver interface {
	Resolve(ctx context.Context, id AssetID) ([]byte, string, error)
}

// URIAssetResolver returns an AssetResolver that accepts asset://<uuid> ids and
// delegates to fn with the bare uuid (D-024). A non-asset:// id is passed to fn
// unchanged.
func URIAssetResolver(fn func(uuid string) ([]byte, string, error)) AssetResolver {
	return uriAssetResolver{fn: fn}
}

type uriAssetResolver struct {
	fn func(uuid string) ([]byte, string, error)
}

func (r uriAssetResolver) Resolve(_ context.Context, id AssetID) ([]byte, string, error) {
	if r.fn == nil {
		return nil, "", ErrAssetNotFound
	}
	uuid := strings.TrimPrefix(string(id), "asset://")
	return r.fn(uuid)
}
