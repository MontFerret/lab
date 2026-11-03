package testing_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func newOneShotExecutor(t testing.TB, rt api.Runtime) *labtesting.CaseExecutor {
	t.Helper()
	executor := labtesting.NewExecutor(rt, false)
	t.Cleanup(func() { _ = executor.Close() })

	return executor
}

func TestExecutorSourceIdentityAndFreshSessions(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		t.Run(fmt.Sprint(reuse), func(t *testing.T) {
			var compiled, closed []api.Source
			var runs, sessions, sessionCloses, outputCloses int
			rt := &apiruntime.Runtime{CloseFunc: func() error {
				t.Error("executor closed borrowed runtime")

				return nil
			}}
			run := func(_ context.Context, _ api.Source, opts ...api.SessionOption) (api.Output, error) {
				params, err := apiruntime.NewParams(opts...)
				if err != nil {
					return nil, err
				}

				if params["value"] != runs {
					t.Errorf("stale session parameters: %v, invocation %d", params, runs)
				}

				runs++

				return &apiruntime.Output{Data: []byte("42"), CloseFunc: func() error { outputCloses++; return nil }}, nil
			}
			rt.RunFunc = run
			rt.CompileFunc = func(_ context.Context, src api.Source, _ ...api.PlanOption) (api.Plan, error) {
				compiled = append(compiled, src)

				return &apiruntime.Plan{NewSessionFunc: func(ctx context.Context, opts ...api.SessionOption) (api.Session, error) {
					sessions++

					return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) {
						return run(ctx, src, opts...)
					}, CloseFunc: func() error { sessionCloses++; return nil }}, nil
				}, CloseFunc: func() error { closed = append(closed, src); return nil }}, nil
			}
			executor := labtesting.NewExecutor(rt, reuse)
			a := api.NewSource("one.fql", "RETURN 42")
			b := api.NewSource("two.fql", a.Content)
			c := api.NewSource(a.Name, "RETURN 43")
			for index, src := range []api.Source{a, a, b, c, a} {
				opts := api.WithParam("value", index)
				if index%2 == 0 {
					content, err := executor.Collect(t.Context(), src, opts)
					if err != nil || content == nil || string(content.Data) != "42" {
						t.Fatalf("collect: %#v %v", content, err)
					}
				} else if err := executor.Consume(t.Context(), src, opts); err != nil {
					t.Fatal(err)
				}
			}

			for range 2 {
				if err := executor.Close(); err != nil {
					t.Fatal(err)
				}
			}

			if runs != 5 || outputCloses != 5 || sessions != sessionCloses {
				t.Fatalf("runs=%d outputs=%d sessions=%d closed=%d", runs, outputCloses, sessions, sessionCloses)
			}

			if reuse {
				if sessions != 5 || !reflect.DeepEqual(compiled, []api.Source{a, b, c}) || !reflect.DeepEqual(closed, []api.Source{c, b, a}) {
					t.Fatalf("sessions=%d compiled=%v closed=%v", sessions, compiled, closed)
				}
			} else if len(compiled) != 0 || sessions != 0 {
				t.Fatalf("one-shot compiled=%v sessions=%d", compiled, sessions)
			}

			if err := executor.Consume(t.Context(), a); err == nil || runs != 5 {
				t.Fatalf("closed executor admitted work: %v runs=%d", err, runs)
			}
		})
	}
}

func TestExecutorCompileErrorsAndUnsupportedFallback(t *testing.T) {
	for _, failure := range []error{errors.New("syntax"), errors.New("transport"), context.Canceled, context.DeadlineExceeded, fmt.Errorf("HTTP: %w", errors.ErrUnsupported), errors.Join(errors.ErrUnsupported, context.Canceled), errors.Join(errors.ErrUnsupported, errors.New("transport"))} {
		t.Run(failure.Error(), func(t *testing.T) {
			var compiles, runs int
			rt := &apiruntime.Runtime{CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
				compiles++

				return nil, failure
			}, RunFunc: func(context.Context, api.Source, ...api.SessionOption) (api.Output, error) {
				runs++

				return &apiruntime.Output{}, nil
			}}
			executor := labtesting.NewExecutor(rt, true)
			fallback := failure.Error() == "HTTP: "+errors.ErrUnsupported.Error()
			for index := range 3 {
				err := executor.Consume(t.Context(), api.NewSource(fmt.Sprintf("query%d", index), "text"))
				if fallback && err != nil || !fallback && !errors.Is(err, failure) {
					t.Fatalf("compile error: %v", err)
				}
			}

			if fallback && (compiles != 1 || runs != 3) || !fallback && (compiles != 3 || runs != 0) {
				t.Fatalf("compiles=%d runs=%d", compiles, runs)
			}

			if err := executor.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestExpectedCompilationFailureIsNotCached(t *testing.T) {
	for _, file := range []sources.File{
		{Name: "query.fail.fql", Content: []byte("invalid")},
		{Name: "suite.yaml", Content: []byte("query:\n  text: invalid\nexpect:\n  error:\n    contains: syntax")},
	} {
		t.Run(file.Name, func(t *testing.T) {
			testCase, err := labtesting.New(labtesting.Options{File: file, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}

			compiles := 0
			executor := labtesting.NewExecutor(&apiruntime.Runtime{CompileFunc: func(ctx context.Context, _ api.Source, _ ...api.PlanOption) (api.Plan, error) {
				compiles++
				if _, ok := ctx.Deadline(); !ok {
					t.Error("compilation outside test deadline")
				}

				return nil, errors.New("syntax failure")
			}}, true)
			for range 3 {
				if err := testCase.Run(t.Context(), executor, labtesting.NewParams()); err != nil {
					t.Fatal(err)
				}
			}

			if err := executor.Close(); err != nil || compiles != 3 {
				t.Fatalf("compiles=%d cleanup=%v", compiles, err)
			}
		})
	}
}

func TestSuiteCompilesAssertionOnlyAfterSuccessfulQuery(t *testing.T) {
	failure := errors.New("terminal query failure")
	compiles := 0
	executor := labtesting.NewExecutor(&apiruntime.Runtime{CompileFunc: func(_ context.Context, src api.Source, _ ...api.PlanOption) (api.Plan, error) {
		compiles++
		if src.Content != "RETURN 1" {
			t.Fatalf("compiled assertion before query succeeded: %v", src)
		}

		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) {
				return &apiruntime.Output{Data: []byte("1"), Err: failure}, nil
			}}, nil
		}}, nil
	}}, true)
	suite, err := labtesting.NewSuite(labtesting.Options{File: sources.File{Name: "suite.yaml", Content: []byte("query:\n  text: RETURN 1\nassert:\n  text: invalid assertion")}, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}

	for range 2 {
		if err := suite.Run(t.Context(), executor, labtesting.NewParams()); !errors.Is(err, failure) {
			t.Fatalf("lost terminal error: %v", err)
		}
	}

	if err := executor.Close(); err != nil || compiles != 1 {
		t.Fatalf("compiles=%d cleanup=%v", compiles, err)
	}
}

func TestReferencedScriptContentChanges(t *testing.T) {
	dir := t.TempDir()
	src, err := sources.New(dir)
	if err != nil {
		t.Fatal(err)
	}

	suite, err := labtesting.NewSuite(labtesting.Options{File: sources.File{Source: src, Name: filepath.Join(dir, "suite.yaml"), Content: []byte("query:\n  ref: query.fql\nassert:\n  text: RETURN true")}, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}

	var compiled []string
	executor := labtesting.NewExecutor(&apiruntime.Runtime{CompileFunc: func(_ context.Context, src api.Source, _ ...api.PlanOption) (api.Plan, error) {
		compiled = append(compiled, src.Content)

		return &apiruntime.Plan{NewSessionFunc: func(context.Context, ...api.SessionOption) (api.Session, error) {
			return &apiruntime.Session{RunFunc: func(context.Context) (api.Output, error) { return &apiruntime.Output{Data: []byte("1")}, nil }}, nil
		}}, nil
	}}, true)
	for _, content := range []string{"RETURN 1", "RETURN 2", "RETURN 1"} {
		if err := os.WriteFile(filepath.Join(dir, "query.fql"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}

		if err := suite.Run(t.Context(), executor, labtesting.NewParams()); err != nil {
			t.Fatal(err)
		}
	}

	if err := executor.Close(); err != nil || !reflect.DeepEqual(compiled, []string{"RETURN 1", "RETURN true", "RETURN 2"}) {
		t.Fatalf("compiled=%v cleanup=%v", compiled, err)
	}
}

func TestExecutorHTTPFallback(t *testing.T) {
	requests := 0
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte("42"))
	}))
	defer host.Close()
	rt, err := runtime.NewRemote(host.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	defer rt.Close()
	executor := labtesting.NewExecutor(rt, true)
	for range 3 {
		content, err := executor.Collect(t.Context(), api.NewSource("query", "RETURN 42"))
		if err != nil || content == nil || string(content.Data) != "42" {
			t.Fatalf("HTTP collect: %v %v", content, err)
		}
	}

	if err := executor.Close(); err != nil || requests != 3 {
		t.Fatalf("requests=%d cleanup=%v", requests, err)
	}
}
