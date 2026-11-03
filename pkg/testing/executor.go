package testing

import (
	"context"
	"errors"
	"fmt"

	"github.com/MontFerret/api"
)

type (
	// Executor settles each script's output before returning to test semantics.
	Executor interface {
		Collect(context.Context, api.Source, ...api.SessionOption) (*api.Content, error)
		Consume(context.Context, api.Source, ...api.SessionOption) error
	}

	// CaseExecutor borrows a runtime and owns plans and temporary sessions for one
	// case. Calls, including Close, must be serialized by its owning runner.
	CaseExecutor struct {
		runtime       api.Runtime
		reusePlans    bool
		plans         map[api.Source]api.Plan
		sources       []api.Source
		cleanupErrors []error
		closed        bool
		closeErr      error
	}

	// cleanupFailure prevents resource failures from satisfying expected errors.
	cleanupFailure struct {
		cause error
	}
)

// NewExecutor borrows rt without closing it. Reuse compiles scripts lazily and
// retains successful plans until Close; each invocation still gets a new session.
func NewExecutor(rt api.Runtime, reusePlans bool) *CaseExecutor {
	return &CaseExecutor{runtime: rt, reusePlans: reusePlans}
}

// Collect returns detached content after settling the output and session.
func (e *CaseExecutor) Collect(ctx context.Context, src api.Source, opts ...api.SessionOption) (content *api.Content, err error) {
	out, session, err := e.run(ctx, src, opts...)
	if err != nil {
		return nil, err
	}

	defer func() { err = e.settle(src, out, session, err) }()

	return out.Collect(ctx)
}

// Consume discards bytes while observing terminal execution and cleanup errors.
func (e *CaseExecutor) Consume(ctx context.Context, src api.Source, opts ...api.SessionOption) (err error) {
	out, session, err := e.run(ctx, src, opts...)
	if err != nil {
		return err
	}

	defer func() { err = e.settle(src, out, session, err) }()

	return out.Consume(ctx, discardOutput)
}

func (e *CaseExecutor) run(ctx context.Context, src api.Source, opts ...api.SessionOption) (api.Output, api.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	if e.closed {
		return nil, nil, errors.New("executor is closed")
	}

	if !e.reusePlans {
		out, err := e.runtime.Run(ctx, src, opts...)

		return out, nil, err
	}

	plan := e.plans[src]
	if plan == nil {
		var err error
		plan, err = e.runtime.Compile(ctx, src)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, nil, errors.Join(err, ctxErr)
			}

			if !unsupportedCompilation(err) {
				return nil, nil, err
			}

			e.reusePlans = false

			out, err := e.runtime.Run(ctx, src, opts...)

			return out, nil, err
		}

		if e.plans == nil {
			e.plans = make(map[api.Source]api.Plan)
		}

		e.plans[src] = plan
		e.sources = append(e.sources, src)
	}

	session, err := plan.NewSession(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	out, err := session.Run(ctx)
	if err != nil {
		return nil, nil, e.settle(src, nil, session, err)
	}

	return out, session, nil
}

func (e *CaseExecutor) settle(src api.Source, out api.Output, session api.Session, err error) error {
	var cleanupErr error

	if out != nil {
		cleanupErr = e.recordCleanup(src, "output", out.Close())
	}

	if session != nil {
		cleanupErr = errors.Join(cleanupErr, e.recordCleanup(src, "session", session.Close()))
	}

	if cleanupErr != nil {
		return &cleanupFailure{cause: errors.Join(err, cleanupErr)}
	}

	return err
}

func (e *CaseExecutor) recordCleanup(src api.Source, resource string, err error) error {
	if err == nil {
		return nil
	}

	err = fmt.Errorf("close %s for %q: %w", resource, src.Name, err)
	e.cleanupErrors = append(e.cleanupErrors, err)

	return err
}

// Close releases all cached plans in reverse acquisition order and retains all
// observed cleanup failures, including those from earlier retried invocations.
// It is idempotent and never closes the borrowed runtime.
func (e *CaseExecutor) Close() error {
	if e.closed {
		return e.closeErr
	}

	e.closed = true

	for index := len(e.sources) - 1; index >= 0; index-- {
		src := e.sources[index]
		e.recordCleanup(src, "plan", e.plans[src].Close())
	}

	e.closeErr = errors.Join(e.cleanupErrors...)
	e.plans = nil
	e.sources = nil
	e.cleanupErrors = nil

	return e.closeErr
}

func (failure *cleanupFailure) Error() string {
	return failure.cause.Error()
}

func (failure *cleanupFailure) Unwrap() error {
	return failure.cause
}
