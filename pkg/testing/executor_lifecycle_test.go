package testing_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestExecutorCancellationSettlesOwnedResources(t *testing.T) {
	for _, collect := range []bool{false, true} {
		for _, phase := range []string{"compile", "session", "run", "before consumption", "consume"} {
			t.Run(phase+map[bool]string{false: "/consume", true: "/collect"}[collect], func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var order []string
				rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
					if phase == "compile" {
						cancel()

						return nil, errors.Join(errors.ErrUnsupported, ctx.Err())
					}

					return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
						if phase == "session" {
							cancel()

							return nil, ctx.Err()
						}

						return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) {
							if phase == "run" {
								cancel()

								return nil, ctx.Err()
							}

							if phase == "before consumption" {
								cancel()
							}

							return &apiruntime.Output{ConsumeFunc: func(context.Context, api.Consumer) error {
								cancel()

								return ctx.Err()
							}, CloseFunc: func() error { order = append(order, "output"); return nil }}, nil
						}, CloseFunc: func() error { order = append(order, "session"); return nil }}, nil
					}, CloseFunc: func() error { order = append(order, "plan"); return nil }}, nil
				}}
				executor := labtesting.NewExecutor(rt, true)
				var err error

				if collect {
					_, err = executor.Collect(ctx, api.NewSource("cancel", "query"))
				} else {
					err = executor.Consume(ctx, api.NewSource("cancel", "query"))
				}

				if !errors.Is(err, context.Canceled) {
					t.Fatalf("lost cancellation: %v", err)
				}

				if err := executor.Close(); err != nil {
					t.Fatal(err)
				}

				var want []string
				switch phase {
				case "session":
					want = []string{"plan"}
				case "run":
					want = []string{"session", "plan"}
				case "before consumption", "consume":
					want = []string{"output", "session", "plan"}
				}

				if !reflect.DeepEqual(order, want) {
					t.Fatalf("cleanup order=%v want=%v", order, want)
				}
			})
		}
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	executor := labtesting.NewExecutor(&apiruntime.Runtime{}, true)
	if err := executor.Consume(ctx, api.Source{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("already canceled invocation: %v", err)
	}

	if err := executor.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupErrorsDoNotSatisfyExpectedFailures(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		for _, executionErr := range []error{nil, errors.New("execution failure")} {
			for _, file := range []sources.File{
				{Name: "expected.fail.fql", Content: []byte("RETURN 1")},
				{Name: "expected.yaml", Content: []byte("query:\n  text: RETURN 1\nexpect:\n  error: {}")},
			} {
				t.Run(file.Name+map[bool]string{false: "/direct", true: "/compiled"}[reuse]+"/"+map[bool]string{false: "cleanup only", true: "execution and cleanup"}[executionErr != nil], func(t *testing.T) {
					outputErr := errors.New("output cleanup")
					sessionErr := errors.New("session cleanup")
					planErr := errors.New("plan cleanup")
					run := func(context.Context) (api.Output, error) {
						return &apiruntime.Output{Err: executionErr, CloseFunc: func() error { return outputErr }}, nil
					}
					rt := &apiruntime.Runtime{RunFunc: func(ctx context.Context, _ api.Source, _ ...api.SessionOption) (api.Output, error) { return run(ctx) }, CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
						return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
							return &apiruntime.Session{RunFunc: run, CloseFunc: func() error { return sessionErr }}, nil
						}, CloseFunc: func() error { return planErr }}, nil
					}}
					testCase, err := labtesting.New(labtesting.Options{File: file, Timeout: time.Second})
					if err != nil {
						t.Fatal(err)
					}

					executor := labtesting.NewExecutor(rt, reuse)
					err = testCase.Run(t.Context(), executor, labtesting.NewParams())
					if !errors.Is(err, outputErr) || executionErr != nil && !errors.Is(err, executionErr) || reuse && !errors.Is(err, sessionErr) {
						t.Fatalf("cleanup was accepted or lost: %v", err)
					}

					for range 2 {
						closeErr := executor.Close()
						if !errors.Is(closeErr, outputErr) || reuse && (!errors.Is(closeErr, sessionErr) || !errors.Is(closeErr, planErr)) {
							t.Fatalf("lost retained cleanup: %v", closeErr)
						}
					}
				})
			}
		}
	}
}

func TestExecutorClosesAllPlansDespiteFailures(t *testing.T) {
	failures := []error{errors.New("first close"), errors.New("second close"), errors.New("third close")}
	var closes []int
	index := 0
	executor := labtesting.NewExecutor(&apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
		current := index
		index++

		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) { return &apiruntime.Output{}, nil }}, nil
		}, CloseFunc: func() error { closes = append(closes, current); return failures[current] }}, nil
	}}, true)
	for _, name := range []string{"first", "second", "third"} {
		if err := executor.Consume(t.Context(), api.NewSource(name, "query")); err != nil {
			t.Fatal(err)
		}
	}

	for range 2 {
		err := executor.Close()
		for _, failure := range failures {
			if !errors.Is(err, failure) {
				t.Fatalf("lost plan cleanup failure: %v", err)
			}
		}
	}

	if !reflect.DeepEqual(closes, []int{2, 1, 0}) {
		t.Fatalf("plan closes=%v", closes)
	}
}

func TestSessionAdmissionFailurePreservesExecutionAndCleanup(t *testing.T) {
	executionErr, cleanupErr := errors.New("admission failure"), errors.New("session cleanup")
	for _, file := range []sources.File{
		{Name: "admission.fail.fql", Content: []byte("RETURN 1")},
		{Name: "admission.yaml", Content: []byte("query:\n  text: RETURN 1\nexpect:\n  error: {}")},
	} {
		t.Run(file.Name, func(t *testing.T) {
			closes := 0
			executor := labtesting.NewExecutor(&apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
				return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
					return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) { return nil, executionErr }, CloseFunc: func() error { closes++; return cleanupErr }}, nil
				}}, nil
			}}, true)
			testCase, err := labtesting.New(labtesting.Options{File: file, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}

			err = testCase.Run(t.Context(), executor, labtesting.NewParams())
			if !errors.Is(err, executionErr) || !errors.Is(err, cleanupErr) || closes != 1 {
				t.Fatalf("session admission: %v closes=%d", err, closes)
			}

			if err := executor.Close(); !errors.Is(err, cleanupErr) || closes != 1 {
				t.Fatalf("retained cleanup: %v closes=%d", err, closes)
			}
		})
	}
}
