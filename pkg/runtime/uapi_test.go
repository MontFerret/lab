package runtime

import (
	"context"
	"errors"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/api/diagnostics"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
)

func TestBuiltinVersionAndCancellation(t *testing.T) {
	old := version
	version = " opaque \n"
	t.Cleanup(func() { version = old })
	rt, err := NewBuiltin(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	got, err := rt.Version(t.Context())
	if err != nil || got != api.Version(version) {
		t.Fatalf("version: %q %v", got, err)
	}

	for _, expired := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		want := error(context.Canceled)
		if expired {
			cancel()
			ctx, cancel = context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
			want = context.DeadlineExceeded
		}

		cancel()
		_, err := rt.Version(ctx)
		if !errors.Is(err, want) {
			t.Fatalf("Version cancellation: %v", err)
		}

		_, err = rt.Run(ctx, api.Source{Name: "canceled.fql", Content: "RETURN 1"})
		if !errors.Is(err, want) {
			t.Fatalf("Run cancellation: %v", err)
		}

		for _, compile := range []func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error){rt.Compile, rt.CompileDebug} {
			plan, err := compile(ctx, api.Source{Name: "canceled.fql", Content: "RETURN 1"})
			if plan != nil || !errors.Is(err, want) {
				t.Fatalf("Compile cancellation: %v %v", plan, err)
			}
		}
	}
}

func TestRuntimeCallerOwnedCompilation(t *testing.T) {
	for _, mode := range []string{"builtin", "wire"} {
		t.Run(mode, func(t *testing.T) {
			opts := Options{Params: map[string]any{"shared": "default"}}
			if mode == "wire" {
				hosted, err := NewBuiltin(nil)
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = hosted.Close() })
				host := wirehost.New(t, hosted)
				opts.Type, opts.Endpoint = "wire", host.Endpoint
			}

			rt, err := New(t.Context(), opts)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = rt.Close() })
			src := api.Source{Name: "caller-owned.fql", Content: "RETURN @value"}
			for _, compile := range []func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error){rt.Compile, rt.CompileDebug} {
				var callbacks int
				plan, err := compile(t.Context(), src, func(options api.PlanOptions) error {
					callbacks++

					return options.SetOptimizationLevel(api.OptimizationNone)
				})
				if err != nil {
					t.Fatal(err)
				}

				// Register fallback cleanup even when an assertion fails.
				t.Cleanup(func() { _ = plan.Close() })
				params, err := plan.Params(t.Context())
				if err != nil || !reflect.DeepEqual(params, []string{"value"}) || callbacks != 1 {
					t.Fatalf("compiled parameters/callbacks: %v %v %d", params, err, callbacks)
				}

				session, err := plan.NewSession(t.Context(), api.WithParam("value", "caller"))
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = session.Close() })
				out, err := session.Run(t.Context())
				if err != nil || out == nil || string(out.Content) != "\"caller\"" {
					t.Fatalf("session output: %#v %v", out, err)
				}

				if err := session.Close(); err != nil {
					t.Fatal(err)
				}

				if err := plan.Close(); err != nil {
					t.Fatal(err)
				}
			}

			if mode == "wire" {
				plan, err := rt.Compile(t.Context(), api.Source{Name: "defaults.fql", Content: "RETURN @shared"})
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = plan.Close() })
				session, err := plan.NewSession(t.Context(), api.WithParam("shared", "session"))
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = session.Close() })
				out, err := session.Run(t.Context())
				if err != nil || out == nil || string(out.Content) != "\"session\"" {
					t.Fatalf("caller-owned session defaults: %#v %v", out, err)
				}

				if err := session.Close(); err != nil {
					t.Fatal(err)
				}

				if err := plan.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRuntimeOptionCallbacksAndParameterPreservation(t *testing.T) {
	for _, mode := range []string{"builtin", "wire"} {
		t.Run(mode, func(t *testing.T) {
			shared := map[string]any{"shared": "default", "value": "early"}
			perRun := map[string]any{"value": "late", "object": map[string]any{"nested": "value"}}
			original := maps.Clone(perRun)
			opts := Options{Params: shared}
			if mode == "wire" {
				hosted, err := NewBuiltin(nil)
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = hosted.Close() })
				host := wirehost.New(t, hosted)
				opts.Type, opts.Endpoint = "wire", host.Endpoint
			}

			rt, err := New(t.Context(), opts)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = rt.Close() })
			var order []int
			out, err := rt.Run(t.Context(), api.Source{Name: "params.fql", Content: "RETURN [@shared, @value, @object.nested]"},
				func(o api.SessionOptions) error { order = append(order, 1); return o.SetParams(perRun) }, nil,
				func(o api.SessionOptions) error { order = append(order, 2); return o.SetParam("value", "last") })
			if err != nil || out == nil || string(out.Content) != "[\"default\",\"last\",\"value\"]" ||
				!reflect.DeepEqual(order, []int{1, 2}) || !reflect.DeepEqual(original, perRun) || shared["value"] != "early" {
				t.Fatalf("options/parameters: %#v %v %v %#v", out, err, order, perRun)
			}

			first, last := errors.New("first"), errors.New("last")
			order = nil
			out, err = rt.Run(t.Context(), api.Source{Name: "failure.fql", Content: "RETURN 1"},
				func(api.SessionOptions) error { order = append(order, 1); return first },
				nil,
				func(api.SessionOptions) error { order = append(order, 2); return last })
			if out != nil || !errors.Is(err, first) || !errors.Is(err, last) || !reflect.DeepEqual(order, []int{1, 2}) {
				t.Fatalf("callback errors: %#v %v %v", out, err, order)
			}
		})
	}
}

func TestWireForwardsSessionSettersAndConvertsYAML(t *testing.T) {
	var captured apiruntime.Options
	host := wirehost.New(t, &apiruntime.Runtime{VersionValue: "host", RunFunc: func(_ context.Context, _ api.Source, opts ...api.SessionOption) (*api.Output, error) {
		captured = apiruntime.Options{Params: apiruntime.Params{}}
		var failures []error
		for _, option := range opts {
			if option != nil {
				failures = append(failures, option(&captured))
			}
		}

		return &api.Output{}, errors.Join(failures...)
	}})
	rt, err := New(t.Context(), Options{Type: "wire", Endpoint: host.Endpoint})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	input := map[any]any{"nested": map[any]any{"value": "original"}}
	out, err := rt.Run(t.Context(), api.Source{}, api.WithParam("yaml", input), api.WithFSRoot("/caller"), api.WithOutputContentType("custom/type"))
	if err != nil || out == nil || captured.FSRoot != "/caller" || captured.ContentType != "custom/type" ||
		!reflect.DeepEqual(captured.Params["yaml"], map[string]any{"nested": map[string]any{"value": "original"}}) ||
		!reflect.DeepEqual(input, map[any]any{"nested": map[any]any{"value": "original"}}) {
		t.Fatalf("Wire setters: %#v %#v %v", captured, out, err)
	}
}

func TestBuiltinCleanupIsOrderedConcurrentAndRetained(t *testing.T) {
	failure := errors.New("native close failed")
	var order []string
	rt := &Builtin{
		Runtime: &apiruntime.Runtime{CloseFunc: func() error { order = append(order, "native"); return failure }},
		network: &networkCleanup{close: func() { order = append(order, "network") }},
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			if err := rt.Close(); !errors.Is(err, failure) {
				t.Errorf("lost native cleanup failure: %v", err)
			}
		})
	}

	group.Wait()
	if err := rt.Close(); !errors.Is(err, failure) || !reflect.DeepEqual(order, []string{"native", "network"}) {
		t.Fatalf("cleanup order/repetition: %v %v", order, err)
	}
}

func TestBuiltinClosesOwnedIdleHTTPConnection(t *testing.T) {
	gone := make(chan struct{}, 1)
	idle := make(chan struct{}, 1)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fixture"))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateIdle:
			idle <- struct{}{}
		case http.StateClosed:
			gone <- struct{}{}
		}
	}

	server.Start()
	t.Cleanup(server.Close)
	rt, err := New(t.Context(), Options{HTTPPolicy: &HTTPPolicy{AllowLocalhost: pointerTo(true)}})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	_, err = rt.Run(t.Context(), api.Source{Name: "network.fql", Content: "RETURN IO::NET::HTTP::GET(\"" + server.URL + "\")"})
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-idle:
	case <-time.After(time.Second):
		t.Fatal("owned HTTP connection did not become idle")
	}

	select {
	case <-gone:
		t.Fatal("HTTP connection closed before runtime cleanup")
	default:
	}

	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-gone:
	case <-time.After(time.Second):
		t.Fatal("owned idle HTTP connection survived runtime cleanup")
	}
}

func TestRuntimeConstructionFailureReturnsNilInterface(t *testing.T) {
	for _, opts := range []Options{
		{FSPolicy: &FileSystemPolicy{Root: t.TempDir() + "/missing"}},
		{HTTPPolicy: &HTTPPolicy{AllowedHosts: []string{"bad host"}}},
		{FSPolicy: &FileSystemPolicy{Root: t.TempDir() + "/missing"}, HTTPPolicy: &HTTPPolicy{AllowLocalhost: pointerTo(true)}},
		{Type: "http://worker.test", Params: map[string]any{"headers": "invalid"}},
	} {
		rt, err := New(t.Context(), opts)
		if rt != nil || err == nil {
			t.Fatalf("failed construction returned %T %v", rt, err)
		}
	}
}

func TestBuiltinSourceIdentityInDiagnostics(t *testing.T) {
	rt, err := NewBuiltin(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	src := api.Source{Name: "日本/identity.fql", Content: "RETURN missingFunction()"}
	_, err = rt.Run(t.Context(), src)
	var issues diagnostics.Diagnostics
	if !errors.As(err, &issues) || len(issues) != 1 || issues[0].Source != src {
		t.Fatalf("lost diagnostic source identity: %#v %v", issues, err)
	}
}

func TestBuiltinPortableSessionSettings(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "fixture.txt"), []byte("caller root"), 0o644); err != nil {
		t.Fatal(err)
	}

	rt, err := NewBuiltin(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	out, err := rt.Run(t.Context(), api.Source{Name: "root.fql", Content: `RETURN TO_STRING(IO::FS::READ("fixture.txt"))`},
		api.WithFSRoot(root), api.WithOutputContentType("application/json"))
	if err != nil || out == nil || string(out.Content) != `"caller root"` || out.ContentType != "application/json" {
		t.Fatalf("portable session settings: %#v %v", out, err)
	}
}
