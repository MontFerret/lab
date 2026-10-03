// Package wirehost provides a test-owned Wire server and controlled UAPI runtime.
package wirehost

import (
	"context"
	"errors"

	"github.com/MontFerret/api"
)

// Runtime controls hosted metadata and one-shot execution without implementing FQL.
type Runtime struct {
	VersionValue api.Version
	VersionFunc  func(context.Context) (api.Version, error)
	RunFunc      func(context.Context, api.Source, ...api.SessionOption) (*api.Output, error)
	CloseFunc    func() error
}

// Version supplies the hosted value used by the Connect handshake.
func (r *Runtime) Version(ctx context.Context) (api.Version, error) {
	if r.VersionFunc != nil {
		return r.VersionFunc(ctx)
	}

	return r.VersionValue, ctx.Err()
}

// Run delegates only to the controlled one-shot implementation.
func (r *Runtime) Run(ctx context.Context, src api.Source, opts ...api.SessionOption) (*api.Output, error) {
	if r.RunFunc == nil {
		return nil, errors.New("unexpected hosted Run")
	}

	return r.RunFunc(ctx, src, opts...)
}

// Compile fails so integration tests cannot accidentally use plan APIs.
func (r *Runtime) Compile(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
	return nil, errors.New("unexpected hosted Compile")
}

// CompileDebug fails so integration tests cannot accidentally use debugger APIs.
func (r *Runtime) CompileDebug(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
	return nil, errors.New("unexpected hosted CompileDebug")
}

// Close observes ownership tests; a Wire server must leave its borrowed host open.
func (r *Runtime) Close() error {
	if r.CloseFunc != nil {
		return r.CloseFunc()
	}

	return nil
}
