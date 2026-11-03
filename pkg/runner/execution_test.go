package runner

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestRunnerRepetitionsReusePlansAndIsolateParameters(t *testing.T) {
	for _, file := range []sources.File{
		{Name: "unit.fql", Content: []byte("RETURN @value")},
		{Name: "suite.yaml", Content: []byte("query:\n  text: RETURN @value\n  params:\n    queryOnly: query\nassert:\n  text: RETURN @lab.data.query.result\n  params:\n    value: assertion")},
	} {
		t.Run(file.Name, func(t *testing.T) {
			compiles := map[api.Source]int{}
			sessions, closes := 0, 0
			rt := &apiruntime.Runtime{CompileFunc: func(_ context.Context, src api.Source, _ ...api.PlanOption) (api.Plan, error) {
				compiles[src]++

				return &apiruntime.Plan{NewSessionFunc: func(_ context.Context, opts ...api.SessionOption) (api.Session, error) {
					sessions++

					params, err := apiruntime.NewParams(opts...)
					if err != nil {
						return nil, err
					}

					lab := params["lab"].(map[string]any)
					if lab["static"].(map[string]any)["app"] != "http://fixtures" {
						t.Errorf("system parameter leaked between invocations: %v", lab)
					}

					if strings.HasSuffix(src.Name, " -> assert") {
						query := lab["data"].(map[string]any)["query"].(map[string]any)
						if params["value"] != "assertion" || params["queryOnly"] != nil || query["result"] != float64(42) || query["params"].(map[string]any)["value"] != "base" {
							t.Errorf("assertion overrides/query context: %v", params)
						}
					} else if params["value"] != "base" || lab["data"] != nil {
						t.Errorf("parameters leaked between attempts: %v", params)
					}

					lab["static"].(map[string]any)["app"] = "changed"
					params["value"] = "changed"

					return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) {
						return &apiruntime.Output{Data: []byte("42")}, nil
					}}, nil
				}, CloseFunc: func() error { closes++; return nil }}, nil
			}}
			r, err := New(Options{Runtime: rt, Times: 3})
			if err != nil {
				t.Fatal(err)
			}

			params := labtesting.NewParams()
			params.SetUserValue("value", "base")
			params.SetSystemValue("static", map[string]any{"app": "http://fixtures"})
			stream := r.Run(NewContext(t.Context(), params), singleFileSource{file: file})
			result := <-stream.Progress
			summary := <-stream.Summary
			wantPlans := 1
			if file.Name == "suite.yaml" {
				wantPlans = 2
			}

			if result.Error != nil || result.Times != 3 || result.Attempts != 3 || summary.Passed != 1 || len(compiles) != wantPlans || closes != wantPlans || sessions != 3*wantPlans {
				t.Fatalf("result=%+v summary=%+v compiles=%v sessions=%d closes=%d", result, summary, compiles, sessions, closes)
			}

			for _, count := range compiles {
				if count != 1 {
					t.Fatalf("plan compiled %d times", count)
				}
			}
		})
	}
}

func TestRunnerRetainsCleanupFailureAfterSuccessfulRetry(t *testing.T) {
	failure := errors.New("first session cleanup")
	sessions, planCloses := 0, 0
	rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			sessions++
			first := sessions == 1

			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) { return &apiruntime.Output{}, nil }, CloseFunc: func() error {
				if first {
					return failure
				}

				return nil
			}}, nil
		}, CloseFunc: func() error { planCloses++; return nil }}, nil
	}}
	r, err := New(Options{Runtime: rt, Attempts: 2})
	if err != nil {
		t.Fatal(err)
	}

	stream := r.Run(NewContext(t.Context(), labtesting.NewParams()), singleFileSource{file: sources.File{Name: "retry.fql", Content: []byte("RETURN 1")}})
	result := <-stream.Progress
	summary := <-stream.Summary
	if !errors.Is(result.Error, failure) || result.Attempts != 2 || sessions != 2 || planCloses != 1 || summary.Failed != 1 {
		t.Fatalf("result=%+v summary=%+v sessions=%d plan closes=%d", result, summary, sessions, planCloses)
	}
}

func TestRunnerClosesPlanBeforeReportingResult(t *testing.T) {
	closing, release := make(chan struct{}), make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(release) })
	failure := errors.New("plan cleanup")
	rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) { return &apiruntime.Output{}, nil }}, nil
		}, CloseFunc: func() error {
			close(closing)
			<-release

			return failure
		}}, nil
	}}
	r, err := New(Options{Runtime: rt, Times: 2})
	if err != nil {
		t.Fatal(err)
	}

	stream := r.Run(NewContext(t.Context(), labtesting.NewParams()), singleFileSource{file: sources.File{Name: "close.fql", Content: []byte("RETURN 1")}})
	select {
	case <-closing:
	case <-time.After(5 * time.Second):
		t.Fatal("plan cleanup did not start")
	}

	select {
	case result := <-stream.Progress:
		t.Fatalf("result published before cleanup: %+v", result)
	default:
	}

	unblock.Do(func() { close(release) })
	result := <-stream.Progress
	summary := <-stream.Summary
	if !errors.Is(result.Error, failure) || summary.Failed != 1 {
		t.Fatalf("result=%+v summary=%+v", result, summary)
	}
}

func TestConcurrentCasesOwnSeparatePlans(t *testing.T) {
	var compiles, closes, sessions atomic.Int32
	ready, release := make(chan struct{}, 2), make(chan struct{})
	var unblock sync.Once
	defer unblock.Do(func() { close(release) })
	rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		compiles.Add(1)
		var uses int

		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			uses++
			sessions.Add(1)
			if uses == 1 {
				ready <- struct{}{}
				<-release
			}

			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) { return &apiruntime.Output{}, nil }}, nil
		}, CloseFunc: func() error { closes.Add(1); return nil }}, nil
	}}
	r, err := New(Options{Runtime: rt, Times: 3, PoolSize: 2})
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan Result, 2)
	for range 2 {
		go func() {
			results <- r.runCase(t.Context(), sources.File{Name: "same.fql", Content: []byte("RETURN 1")}, labtesting.NewParams())
		}()
	}

	for range 2 {
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent cases did not create independent sessions")
		}
	}

	unblock.Do(func() { close(release) })
	for range 2 {
		select {
		case result := <-results:
			if result.Error != nil || result.Times != 3 {
				t.Fatalf("concurrent result=%+v", result)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("concurrent case did not finish")
		}
	}

	if compiles.Load() != 2 || closes.Load() != 2 || sessions.Load() != 6 {
		t.Fatalf("compiles=%d closes=%d sessions=%d", compiles.Load(), closes.Load(), sessions.Load())
	}
}
