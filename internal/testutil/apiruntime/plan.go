package apiruntime

import (
	"context"
	"errors"

	"github.com/MontFerret/api"
	"github.com/MontFerret/api/debugger"
)

type (
	// Plan controls session construction and observes plan resource ownership.
	Plan struct {
		ParamsFunc          func(context.Context) ([]string, error)
		NewSessionFunc      func(context.Context, ...api.SessionOption) (api.Session, error)
		NewDebugSessionFunc func(context.Context, ...api.SessionOption) (debugger.Session, error)
		CloseFunc           func() error
	}

	// Session controls one execution and its independently owned cleanup.
	Session struct {
		RunFunc   func(context.Context) (api.Output, error)
		CloseFunc func() error
	}
)

// Params returns controlled metadata without inferring it from FQL.
func (p *Plan) Params(ctx context.Context) ([]string, error) {
	if p.ParamsFunc != nil {
		return p.ParamsFunc(ctx)
	}

	return nil, ctx.Err()
}

// NewSession creates a caller-owned session using the invocation's options.
func (p *Plan) NewSession(ctx context.Context, opts ...api.SessionOption) (api.Session, error) {
	if p.NewSessionFunc == nil {
		return nil, errors.New("unexpected NewSession")
	}

	return p.NewSessionFunc(ctx, opts...)
}

// NewDebugSession rejects debugger work unless explicitly controlled.
func (p *Plan) NewDebugSession(ctx context.Context, opts ...api.SessionOption) (debugger.Session, error) {
	if p.NewDebugSessionFunc == nil {
		return nil, errors.New("unexpected NewDebugSession")
	}

	return p.NewDebugSessionFunc(ctx, opts...)
}

// Close observes plan release without closing its independently owned sessions.
func (p *Plan) Close() error {
	if p.CloseFunc != nil {
		return p.CloseFunc()
	}

	return nil
}

// Run returns the controlled output or admission failure.
func (s *Session) Run(ctx context.Context) (api.Output, error) {
	if s.RunFunc == nil {
		return nil, errors.New("unexpected Session.Run")
	}

	return s.RunFunc(ctx)
}

// Close observes session release.
func (s *Session) Close() error {
	if s.CloseFunc != nil {
		return s.CloseFunc()
	}

	return nil
}
