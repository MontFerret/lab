package runtime

import (
	"context"

	"github.com/MontFerret/api"
	"github.com/MontFerret/api/debugger"
)

// wirePlan applies the same parameter normalization as one-shot Run while
// preserving the underlying plan's metadata, ownership, and close contract.
type wirePlan struct {
	api.Plan
}

func (p *wirePlan) NewSession(ctx context.Context, opts ...api.SessionOption) (api.Session, error) {
	return p.Plan.NewSession(ctx, wireSessionSetters(opts)...)
}

func (p *wirePlan) NewDebugSession(ctx context.Context, opts ...api.SessionOption) (debugger.Session, error) {
	return p.Plan.NewDebugSession(ctx, wireSessionSetters(opts)...)
}
