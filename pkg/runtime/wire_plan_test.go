package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/MontFerret/api"
	"github.com/MontFerret/api/debugger"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
)

func TestWirePlanPreservesSessionOptionBehavior(t *testing.T) {
	first, second := errors.New("first callback"), errors.New("second callback")

	for _, mode := range []string{"run", "compile/session", "compile/debug", "compileDebug/session", "compileDebug/debug"} {
		t.Run(mode, func(t *testing.T) {
			input := map[string]any{"object": map[any]any{"nested": []any{map[any]any{"value": "YAML"}}}, "value": "before"}
			var order []int
			var captured apiruntime.Options
			capture := func(opts []api.SessionOption) error {
				captured = apiruntime.Options{Params: apiruntime.Params{}}
				var failures []error

				for _, option := range opts {
					if option != nil {
						failures = append(failures, option(&captured))
					}
				}

				return errors.Join(failures...)
			}
			closes := 0
			underlying := &apiruntime.Plan{ParamsFunc: func(context.Context) ([]string, error) { return []string{"object", "value"}, nil }, NewSessionFunc: func(_ context.Context, opts ...api.SessionOption) (api.Session, error) {
				return nil, capture(opts)
			}, NewDebugSessionFunc: func(_ context.Context, opts ...api.SessionOption) (debugger.Session, error) {
				return nil, capture(opts)
			}, CloseFunc: func() error { closes++; return nil }}
			rt := &wireRuntime{remote: &apiruntime.Runtime{RunFunc: func(_ context.Context, _ api.Source, opts ...api.SessionOption) (api.Output, error) {
				return nil, capture(opts)
			}, CompileFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) { return underlying, nil }, CompileDebugFunc: func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error) { return underlying, nil }}}
			opts := []api.SessionOption{func(target api.SessionOptions) error {
				order = append(order, 1)

				return errors.Join(first, target.SetParams(input))
			}, nil, func(target api.SessionOptions) error {
				order = append(order, 2)

				return errors.Join(second, target.SetParam("value", "after"))
			}, func(target api.SessionOptions) error {
				order = append(order, 3)

				return errors.Join(target.SetFSRoot("/fixtures"), target.SetOutputContentType("application/json"))
			}}
			var err error

			if mode == "run" {
				_, err = rt.Run(t.Context(), api.Source{}, opts...)
			} else {
				compile := rt.Compile
				if mode == "compileDebug/session" || mode == "compileDebug/debug" {
					compile = rt.CompileDebug
				}

				plan, compileErr := compile(t.Context(), api.Source{})
				if compileErr != nil {
					t.Fatal(compileErr)
				}

				params, paramsErr := plan.Params(t.Context())
				if paramsErr != nil || !reflect.DeepEqual(params, []string{"object", "value"}) {
					t.Fatalf("plan metadata: %v %v", params, paramsErr)
				}

				if mode == "compile/debug" || mode == "compileDebug/debug" {
					_, err = plan.NewDebugSession(t.Context(), opts...)
				} else {
					_, err = plan.NewSession(t.Context(), opts...)
				}

				if closeErr := plan.Close(); closeErr != nil || closes != 1 {
					t.Fatalf("close forwarding: %v, calls=%d", closeErr, closes)
				}
			}

			want := apiruntime.Params{"object": map[string]any{"nested": []any{map[string]any{"value": "YAML"}}}, "value": "after"}
			if !errors.Is(err, first) || !errors.Is(err, second) || !reflect.DeepEqual(order, []int{1, 2, 3}) || !reflect.DeepEqual(captured.Params, want) || captured.FSRoot != "/fixtures" || captured.ContentType != "application/json" {
				t.Fatalf("options=%+v order=%v err=%v", captured, order, err)
			}

			captured.Params["object"].(map[string]any)["nested"].([]any)[0].(map[string]any)["value"] = "changed"
			if input["value"] != "before" || input["object"].(map[any]any)["nested"].([]any)[0].(map[any]any)["value"] != "YAML" {
				t.Fatalf("caller-owned input changed: %v", input)
			}
		})
	}
}

func TestWireCompiledParametersRoundTrip(t *testing.T) {
	hosted, err := NewBuiltin()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = hosted.Close() })
	host := wirehost.New(t, hosted)
	rt, err := New(t.Context(), Options{Target: "wire+" + host.Endpoint})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	src := api.NewSource("nested.yaml -> query", "RETURN [@object.nested.value, @value]")
	for _, mode := range []string{"run", "compile", "compileDebug"} {
		t.Run(mode, func(t *testing.T) {
			var plan api.Plan

			if mode != "run" {
				compile := rt.Compile
				if mode == "compileDebug" {
					compile = rt.CompileDebug
				}

				var err error
				plan, err = compile(t.Context(), src)
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = plan.Close() })
			}

			for index := range 2 {
				params := map[string]any{"object": map[any]any{"nested": map[any]any{"value": "YAML"}}, "value": "before"}
				opts := []api.SessionOption{api.WithParams(params), api.WithParam("value", index)}
				var out api.Output
				if plan == nil {
					out, err = rt.Run(t.Context(), src, opts...)
				} else {
					session, sessionErr := plan.NewSession(t.Context(), opts...)
					if sessionErr != nil {
						t.Fatal(sessionErr)
					}

					t.Cleanup(func() { _ = session.Close() })
					out, err = session.Run(t.Context())
				}

				if err != nil {
					t.Fatal(err)
				}

				content, err := out.Collect(t.Context())
				if err != nil || content == nil || string(content.Data) != fmt.Sprintf(`["YAML",%d]`, index) {
					t.Fatalf("result=%v err=%v", content, err)
				}
			}

			cycle := map[string]any{}
			cycle["self"] = cycle
			for _, invalid := range []any{map[any]any{1: "numeric key"}, cycle, make(chan int), func() {}} {
				var err error
				if plan == nil {
					_, err = rt.Run(t.Context(), src, api.WithParam("object", invalid))
				} else {
					_, err = plan.NewSession(t.Context(), api.WithParam("object", invalid))
				}

				if err == nil {
					t.Fatalf("accepted nonportable parameter %T", invalid)
				}
			}
		})
	}
}
