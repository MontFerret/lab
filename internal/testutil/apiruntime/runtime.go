// Package apiruntime provides controlled, transport-independent UAPI test fixtures.
package apiruntime

import (
	"context"
	"errors"

	"github.com/MontFerret/api"
)

// Runtime controls metadata and one-shot execution without implementing FQL.
type Runtime struct {
	VersionValue     api.Version
	VersionFunc      func(context.Context) (api.Version, error)
	RunFunc          func(context.Context, api.Source, ...api.SessionOption) (*api.Output, error)
	CloseFunc        func() error
	CompileFunc      func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error)
	CompileDebugFunc func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error)
}

// Version supplies the configured version.
func (r *Runtime) Version(ctx context.Context) (api.Version, error) {
	if r.VersionFunc != nil {
		return r.VersionFunc(ctx)
	}

	return r.VersionValue, ctx.Err()
}

// Run delegates only to the controlled one-shot implementation.
func (r *Runtime) Run(ctx context.Context, src api.Source, opts ...api.SessionOption) (*api.Output, error) {
	if r.RunFunc == nil {
		return nil, errors.New("unexpected Run")
	}

	return r.RunFunc(ctx, src, opts...)
}

// Compile fails so integration tests cannot accidentally use plan APIs.
func (r *Runtime) Compile(ctx context.Context, src api.Source, opts ...api.PlanOption) (api.Plan, error) {
	if r.CompileFunc != nil {
		return r.CompileFunc(ctx, src, opts...)
	}

	return nil, errors.New("unexpected Compile")
}

// CompileDebug fails so integration tests cannot accidentally use debugger APIs.
func (r *Runtime) CompileDebug(ctx context.Context, src api.Source, opts ...api.PlanOption) (api.Plan, error) {
	if r.CompileDebugFunc != nil {
		return r.CompileDebugFunc(ctx, src, opts...)
	}

	return nil, errors.New("unexpected CompileDebug")
}

// Close observes resource ownership tests.
func (r *Runtime) Close() error {
	if r.CloseFunc != nil {
		return r.CloseFunc()
	}

	return nil
}

var _ api.Runtime = (*Runtime)(nil)
