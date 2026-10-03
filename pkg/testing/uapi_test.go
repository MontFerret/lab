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

func TestUnitUsesCanonicalOneShotSourceAndIsolatedParameters(t *testing.T) {
	file := sources.File{Name: "日本/unit.fql", Content: []byte("\r\nRETURN @value\n")}
	unit, err := labtesting.NewUnit(labtesting.Options{File: file, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}

	params := labtesting.NewParams()
	params.SetUserValue("value", "user")
	params.SetUserValue("lab", "user must not override system")
	params.SetSystemValue("static", map[string]any{"app": "http://fixtures"})
	original := params.ToMap()
	calls := 0
	rt := &apiruntime.Runtime{
		RunFunc: func(ctx context.Context, src api.Source, opts ...api.SessionOption) (*api.Output, error) {
			calls++
			captured, err := apiruntime.NewParams(opts...)
			if err != nil {
				return nil, err
			}

			if src.Name != file.Name || src.Content != string(file.Content) || !reflect.DeepEqual(map[string]any(captured), original) {
				t.Fatalf("unit source/parameters: %#v %#v", src, captured)
			}

			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("unit did not set its timeout")
			}

			captured["value"] = "changed"
			captured["lab"].(map[string]any)["static"].(map[string]any)["app"] = "changed"

			return nil, nil
		},
		CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
			t.Fatal("Lab must use one-shot Run")

			return nil, nil
		},
		CompileDebugFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
			t.Fatal("Lab must not compile debug plans")

			return nil, nil
		},
	}
	if err := unit.Run(t.Context(), rt, params); err != nil || calls != 1 || !reflect.DeepEqual(params.ToMap(), original) {
		t.Fatalf("unit result/parameter isolation: %v calls=%d %#v", err, calls, params.ToMap())
	}
}

func TestSuiteCanonicalOutputContract(t *testing.T) {
	failure := errors.New("query failure")
	for _, tc := range []struct {
		name       string
		output     *api.Output
		err        error
		wantResult any
		wantCalls  int
	}{
		{"absent", nil, nil, nil, 2},
		{"empty", &api.Output{}, nil, nil, 2},
		{"opaque type JSON", &api.Output{Content: []byte("42"), ContentType: "custom/opaque"}, nil, float64(42), 2},
		{"output with error", &api.Output{Content: []byte("invalid JSON"), ContentType: "application/json"}, failure, nil, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			suite, err := labtesting.NewSuite(labtesting.Options{
				File:    sources.File{Name: "日本/suite.yaml", Content: []byte("query:\n  text: 'RETURN @value'\n  params:\n    value: query\nassert:\n  text: 'RETURN @lab.data.query.result'\n  params:\n    value: assert\n")},
				Timeout: time.Second,
			})
			if err != nil {
				t.Fatal(err)
			}

			params := labtesting.NewParams()
			params.SetUserValue("value", "user")
			calls := 0
			rt := &apiruntime.Runtime{
				RunFunc: func(_ context.Context, src api.Source, opts ...api.SessionOption) (*api.Output, error) {
					calls++
					captured, err := apiruntime.NewParams(opts...)
					if err != nil {
						return nil, err
					}

					if calls == 1 {
						if src != (api.Source{Name: "日本/suite.yaml -> query", Content: "RETURN @value"}) || captured["value"] != "query" {
							t.Fatalf("query identity/parameters: %#v %#v", src, captured)
						}

						return tc.output, tc.err
					}

					if src != (api.Source{Name: "日本/suite.yaml -> assert", Content: "RETURN @lab.data.query.result"}) || captured["value"] != "assert" {
						t.Fatalf("assertion identity/parameters: %#v %#v", src, captured)
					}

					lab := captured["lab"].(map[string]any)
					query := lab["data"].(map[string]any)["query"].(map[string]any)
					if !reflect.DeepEqual(query["result"], tc.wantResult) || query["params"].(map[string]any)["value"] != "query" {
						t.Fatalf("query result/parameter context: %#v", query)
					}

					return nil, nil
				},
				CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
					t.Fatal("suite must use Run")

					return nil, nil
				},
				CompileDebugFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) {
					t.Fatal("suite must not compile debug plans")

					return nil, nil
				},
			}
			err = suite.Run(t.Context(), rt, params.Clone())
			if (tc.err == nil && err != nil) || (tc.err != nil && !errors.Is(err, tc.err)) || calls != tc.wantCalls || params.ToMap()["value"] != "user" {
				t.Fatalf("suite result: %v calls=%d %#v", err, calls, params.ToMap())
			}
		})
	}
}

func TestTestCasesPreserveCancellation(t *testing.T) {
	for _, content := range []struct{ name, body string }{
		{"cancel.fql", "RETURN 1"},
		{"cancel.yaml", "query:\n  text: RETURN 1\nassert:\n  text: RETURN true\n"},
	} {
		t.Run(content.name, func(t *testing.T) {
			testCase, err := labtesting.New(labtesting.Options{File: sources.File{Name: content.name, Content: []byte(content.body)}, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}

			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			rt := &apiruntime.Runtime{RunFunc: func(ctx context.Context, _ api.Source, _ ...api.SessionOption) (*api.Output, error) {
				return nil, ctx.Err()
			}}
			if err := testCase.Run(ctx, rt, labtesting.NewParams()); !errors.Is(err, context.Canceled) {
				t.Fatalf("lost cancellation: %v", err)
			}
		})
	}
}
