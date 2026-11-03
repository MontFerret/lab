package runtime

import (
	"context"

	"github.com/MontFerret/api"
)

// New validates and constructs a runtime. The context bounds Wire construction;
// callers supply independent execution contexts and close the runtime after runs settle.
func New(ctx context.Context, opts Options) (api.Runtime, error) {
	config, err := opts.resolve()
	if err != nil {
		return nil, err
	}

	switch config.kind {
	case "http", "https":
		return newRemote(config.url, config.http), nil
	case "wire+tcp":
		rt, err := newWire(ctx, config.endpoint, config.connectTimeout)
		if err != nil {
			return nil, err
		}

		return rt, nil
	default:
		rt, err := newBuiltin(opts.FSPolicy, config.httpPolicyOptions...)
		if err != nil {
			return nil, err
		}

		return rt, nil
	}
}
