package runner

import (
	"context"
	"errors"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

type singleFileSource struct {
	file sources.File
}

func (s singleFileSource) Read(_ context.Context) (<-chan sources.File, <-chan sources.Error) {
	onNext := make(chan sources.File, 1)
	onError := make(chan sources.Error)

	onNext <- s.file
	close(onNext)
	close(onError)

	return onNext, onError
}

func (s singleFileSource) Resolve(_ context.Context, _ *url.URL) (<-chan sources.File, <-chan sources.Error) {
	onNext := make(chan sources.File)
	onError := make(chan sources.Error)

	close(onNext)
	close(onError)

	return onNext, onError
}

func TestRunnerStopsDuringTimesIntervalWhenContextCanceled(t *testing.T) {
	var calls atomic.Int32
	firstCall := make(chan struct{}, 1)

	rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		return nil, errors.ErrUnsupported
	}, RunFunc: func(_ context.Context, _ api.Source, _ ...api.SessionOption) (api.Output, error) {
		calls.Add(1)

		return &apiruntime.Output{ConsumeFunc: func(context.Context, api.Consumer) error {
			firstCall <- struct{}{}

			return nil
		}}, nil
	}}

	r, err := New(Options{
		Runtime:       rt,
		Times:         2,
		TimesInterval: 10,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	stream := r.Run(NewContext(ctx, labtesting.NewParams()), singleFileSource{
		file: sources.File{
			Name:    "test.fql",
			Content: []byte("RETURN 1"),
		},
	})

	select {
	case <-firstCall:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first runtime invocation")
	}

	start := time.Now()
	cancel()

	select {
	case res, ok := <-stream.Progress:
		if !ok {
			t.Fatal("progress channel closed before delivering result")
		}

		if res.Error != nil {
			t.Fatalf("expected no error, got %v", res.Error)
		}

		if res.Times != 1 {
			t.Fatalf("expected a single completed run, got %d", res.Times)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for canceled run to finish")
	}

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("expected cancellation to stop interval wait promptly, took %s", elapsed)
	}

	select {
	case sum, ok := <-stream.Summary:
		if !ok {
			t.Fatal("summary channel closed before delivering summary")
		}

		if sum.Passed != 1 || sum.Failed != 0 {
			t.Fatalf("unexpected summary: %+v", sum)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for summary")
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly one runtime invocation, got %d", got)
	}
}

func TestRunnerReportsLegacyExpectedFailureDeprecationOnce(t *testing.T) {
	var calls atomic.Int32

	rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		return nil, errors.ErrUnsupported
	}, RunFunc: func(_ context.Context, _ api.Source, _ ...api.SessionOption) (api.Output, error) {
		calls.Add(1)

		return nil, errors.New("expected runtime failure")
	}}

	r, err := New(Options{
		Runtime: rt,
		Times:   2,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	stream := r.Run(NewContext(context.Background(), labtesting.NewParams()), singleFileSource{
		file: sources.File{
			Name:    "test.fail.fql",
			Content: []byte("RETURN 1"),
		},
	})

	result := <-stream.Progress
	const warning = "'.fail.fql' expected-failure tests are deprecated; use a YAML test with 'expect.error' instead"

	if result.Error != nil {
		t.Fatalf("expected no error, got %v", result.Error)
	}

	if result.Warning != warning {
		t.Fatalf("expected warning %q, got %q", warning, result.Warning)
	}

	if result.Times != 2 {
		t.Fatalf("expected two successful runs, got %d", result.Times)
	}

	summary := <-stream.Summary
	if summary.Passed != 1 || summary.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected exactly two runtime invocations, got %d", got)
	}
}

func TestRunnerRetriesAfterAssertionOutputFails(t *testing.T) {
	calls, compiles, sessions, closes := 0, 0, 0, 0
	rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		compiles++

		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			sessions++

			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) {
				calls++
				if calls == 2 {
					return &apiruntime.Output{Data: []byte("true"), Err: errors.New("late assertion failure")}, nil
				}

				return &apiruntime.Output{Data: []byte("true")}, nil
			}}, nil
		}, CloseFunc: func() error { closes++; return nil }}, nil
	}}
	r, err := New(Options{Runtime: rt, Attempts: 2})
	if err != nil {
		t.Fatal(err)
	}

	stream := r.Run(NewContext(t.Context(), labtesting.NewParams()), singleFileSource{file: sources.File{
		Name: "retry.yaml", Content: []byte("query:\n  text: RETURN 1\nassert:\n  text: RETURN true"),
	}})
	result := <-stream.Progress
	summary := <-stream.Summary
	if result.Error != nil || result.Attempts != 2 || result.Times != 1 || calls != 4 || compiles != 2 || sessions != 4 || closes != 2 || summary.Passed != 1 || summary.Failed != 0 {
		t.Fatalf("result=%+v summary=%+v calls=%d compiles=%d sessions=%d plan closes=%d", result, summary, calls, compiles, sessions, closes)
	}
}
