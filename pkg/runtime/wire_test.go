package runtime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
	wireclient "github.com/MontFerret/wire/client"
)

func TestRuntimeSelection(t *testing.T) {
	for _, value := range []string{"", "builtin", "http://example.test", "https://example.test"} {
		t.Run(value, func(t *testing.T) {
			rt, err := New(t.Context(), Options{Target: value})
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = rt.Close() })
			var correct bool
			switch {
			case strings.HasPrefix(value, "http"):
				_, correct = rt.(*Remote)
			default:
				_, correct = rt.(*Builtin)
			}

			if !correct {
				t.Fatalf("runtime %q selected %T", value, rt)
			}
		})
	}

	if _, err := New(t.Context(), Options{Target: "%"}); err == nil {
		t.Fatal("invalid URL must fail selection")
	}
}

func TestWireEndpointValidation(t *testing.T) {
	for _, endpoint := range []string{"", "tcp://localhost:1", "tcp://127.0.0.2:1", "tcp://[::1]:1", "grpc://127.0.0.1:1", "tcp://127.0.0.1:", "tcp://127.0.0.1:0", "tcp://127.0.0.1:65536", "tcp://127.0.0.1:-1", "tcp://127.0.0.1:1/", "tcp://127.0.0.1:1?x=1", "tcp://127.0.0.1:1#x", "tcp://user@127.0.0.1:1", " tcp://127.0.0.1:1", "tcp://127.0.0.1:1 "} {
		t.Run(endpoint, func(t *testing.T) {
			rt, err := New(t.Context(), Options{Target: "wire+" + endpoint})
			if err == nil || rt != nil {
				t.Fatalf("invalid endpoint returned %T, %v", rt, err)
			}
		})
	}

	for endpoint, address := range map[string]string{"tcp://127.0.0.1:1": "127.0.0.1:1", "tcp://127.0.0.1:65535": "127.0.0.1:65535", "tcp://127.0.0.1:00042": "127.0.0.1:42"} {
		got, err := wireAddress(endpoint)
		if err != nil || got != address {
			t.Fatalf("address %q = %q, %v", endpoint, got, err)
		}
	}
}

func TestWireVersionAndIndependentConstructionContext(t *testing.T) {
	for _, value := range []api.Version{"v2.0.0-alpha.test", "", " opaque version \n"} {
		t.Run(string(value), func(t *testing.T) {
			var calls atomic.Int32
			host := wirehost.New(t, &apiruntime.Runtime{VersionFunc: func(context.Context) (api.Version, error) {
				calls.Add(1)

				return value, nil
			}, RunFunc: func(context.Context, api.Source, ...api.SessionOption) (api.Output, error) {
				return &apiruntime.Output{Data: []byte("[true]")}, nil
			}})
			ctx, cancel := context.WithCancel(t.Context())
			rt, err := New(ctx, Options{Target: "wire+" + host.Endpoint})
			cancel()
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = rt.Close() })
			for range 2 {
				got, err := rt.Version(t.Context())
				if err != nil || got != value {
					t.Fatalf("version = %q, %v", got, err)
				}
			}

			out, err := runAndCollect(t.Context(), rt, api.NewSource("independent.fql", "RETURN true"), api.WithParams(nil))
			if err != nil || out == nil || string(out.Data) != "[true]" {
				t.Fatalf("construction context retained: %#v, %v", out, err)
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
					t.Fatalf("version context: %v", err)
				}
			}

			if calls.Load() != 1 {
				t.Fatalf("version performed another RPC: %d calls", calls.Load())
			}
		})
	}
}

func TestWireRunRoundTrip(t *testing.T) {
	type invocation struct {
		src    api.Source
		params apiruntime.Params
	}
	calls := make(chan invocation, 4)
	output := []byte{0, 255, 'x', '\n'}
	var hostedCloses atomic.Int32
	host := wirehost.New(t, &apiruntime.Runtime{VersionValue: "host", CloseFunc: func() error {
		hostedCloses.Add(1)

		return nil
	}, RunFunc: func(_ context.Context, src api.Source, options ...api.SessionOption) (api.Output, error) {
		params := apiruntime.Params{}
		for _, option := range options {
			if err := option(params); err != nil {
				return nil, err
			}
		}

		calls <- invocation{src, params}

		return &apiruntime.Output{Data: output, ContentType: "application/octet-stream"}, nil
	}})
	shared := map[string]any{"shared": "value", "overlap": "shared", "nested": map[any]any{"items": []any{map[any]any{"binary": []byte{0, 255}}}}}
	perRun := map[string]any{"overlap": "run", "headers": "FQL", "lab": map[string]any{"static": map[string]any{"app": "http://fixtures"}}}
	rt, err := New(t.Context(), Options{Target: "wire+" + host.Endpoint})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	src := api.NewSource("fixtures/日本.fql", "\r\nRETURN @shared // exact text\n")
	for range 2 {
		out, err := runAndCollect(t.Context(), rt, src, api.WithParams(shared), api.WithParams(perRun))
		if err != nil || out == nil || !bytes.Equal(out.Data, output) {
			t.Fatalf("run = %v, %v", out, err)
		}

		call := <-calls
		wantNested := map[string]any{"items": []any{map[string]any{"binary": []byte{0, 255}}}}
		if !reflect.DeepEqual(call.params["nested"], wantNested) {
			t.Fatalf("nested YAML/binary parameters changed: %#v", call.params["nested"])
		}

		if call.src.Name != src.Name || call.src.Content != src.Content || call.params["shared"] != "value" || call.params["overlap"] != "run" || !reflect.DeepEqual(call.params["lab"], perRun["lab"]) || call.params["headers"] != "FQL" {
			t.Fatalf("round trip changed invocation: %#v", call)
		}
	}

	if shared["overlap"] != "shared" || len(shared) != 3 || len(perRun) != 3 {
		t.Fatal("input parameters were mutated")
	}

	if _, ok := shared["nested"].(map[any]any); !ok {
		t.Fatal("shared YAML object was mutated")
	}

	wire := rt.(*wireRuntime)
	conn := wire.transport.connection.(*grpc.ClientConn)
	if err := wire.remote.Close(); err != nil {
		t.Fatal(err)
	}

	if conn.GetState() != connectivity.Ready {
		t.Fatal("logical close also closed transport")
	}

	if _, err := wire.remote.Run(t.Context(), api.Source{}); !errors.Is(err, wireclient.ErrClosed) {
		t.Fatalf("logical runtime remained open: %v", err)
	}

	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}

	if conn.GetState() != connectivity.Shutdown {
		t.Fatal("adapter did not close transport")
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := host.WaitForClientsClosed(ctx); err != nil || host.Connections() != 1 || hostedCloses.Load() != 0 {
		t.Fatalf("ownership/cleanup: %v, %d connections, %d host closes", err, host.Connections(), hostedCloses.Load())
	}
}

func TestWirePreservesLargeOutputWithTerminalError(t *testing.T) {
	// Exceeds the default 4 MiB gRPC message limit.
	prefix := bytes.Repeat([]byte("available"), 600_000)
	host := wirehost.New(t, &apiruntime.Runtime{VersionValue: "host", RunFunc: func(context.Context, api.Source, ...api.SessionOption) (api.Output, error) {
		return &apiruntime.Output{Data: prefix, Err: errors.New("host failure")}, nil
	}})
	rt, err := New(t.Context(), Options{Target: "wire+" + host.Endpoint})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	out, err := runAndCollect(t.Context(), rt, api.NewSource("failure.fql", "RETURN 1"), api.WithParams(nil))
	if err == nil || out == nil || !bytes.Equal(out.Data, prefix) {
		t.Fatalf("lost streamed prefix or terminal failure: %v", err)
	}
}

func TestWireHandshakeFailureClosesTransport(t *testing.T) {
	host := wirehost.New(t, &apiruntime.Runtime{VersionFunc: func(context.Context) (api.Version, error) {
		return "", errors.New("handshake failed")
	}})
	rt, err := New(t.Context(), Options{Target: "wire+" + host.Endpoint})
	if err == nil || rt != nil {
		t.Fatalf("handshake failure = %T, %v", rt, err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := host.WaitForClientsClosed(ctx); err != nil || host.Connections() != 1 {
		t.Fatalf("failed handshake leaked transport: %v", err)
	}
}

func TestWireConstructionCancellation(t *testing.T) {
	for _, kind := range []string{"canceled", "expired", "handshake canceled", "handshake deadline", "HTTP2 stalled"} {
		t.Run(kind, func(t *testing.T) {
			ctx := t.Context()
			want := error(context.DeadlineExceeded)
			opts := Options{AdapterOptions: map[string]any{"connect-timeout": "100ms"}}
			started := make(chan struct{})
			var cancel context.CancelFunc
			if kind == "canceled" || kind == "handshake canceled" {
				ctx, cancel = context.WithCancel(ctx)
				want = context.Canceled
				defer cancel()
			}

			if kind == "canceled" {
				cancel()
			} else if kind == "expired" {
				ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
				defer cancel()
			}

			if kind == "HTTP2 stalled" {
				ln, err := net.Listen("tcp4", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}

				t.Cleanup(func() { _ = ln.Close() })
				opts.Target = "wire+" + "tcp://" + ln.Addr().String()
				gone := make(chan error, 1)
				go func() {
					conn, err := ln.Accept()
					if err == nil {
						_, err = io.Copy(io.Discard, conn)
						_ = conn.Close()
					}

					gone <- err
				}()
				_, err = New(ctx, opts)
				if !errors.Is(err, want) {
					t.Fatalf("stalled setup = %v", err)
				}

				select {
				case <-gone:
				case <-time.After(time.Second):
					t.Fatal("stalled transport was not closed")
				}

				return
			}

			host := wirehost.New(t, &apiruntime.Runtime{VersionFunc: func(ctx context.Context) (api.Version, error) {
				close(started)
				<-ctx.Done()

				return "", ctx.Err()
			}})
			opts.Target = "wire+" + host.Endpoint
			if kind == "handshake canceled" {
				go func() {
					select {
					case <-started:
						cancel()
					case <-ctx.Done():
					}
				}()
			}

			rt, err := New(ctx, opts)
			if rt != nil || !errors.Is(err, want) {
				t.Fatalf("construction = %T, %v", rt, err)
			}

			waitCtx, waitCancel := context.WithTimeout(t.Context(), time.Second)
			defer waitCancel()
			if err := host.WaitForClientsClosed(waitCtx); err != nil {
				t.Fatalf("canceled construction leaked transport: %v", err)
			}
		})
	}
}

func TestWireRunCancellation(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancel", true: "deadline"}[deadline], func(t *testing.T) {
			started := make(chan struct{})
			finished := make(chan struct{})
			host := wirehost.New(t, &apiruntime.Runtime{VersionValue: "host", RunFunc: func(ctx context.Context, _ api.Source, _ ...api.SessionOption) (api.Output, error) {
				close(started)
				defer close(finished)
				<-ctx.Done()

				return nil, ctx.Err()
			}})
			rt, err := New(t.Context(), Options{Target: "wire+" + host.Endpoint})
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(func() { _ = rt.Close() })
			ctx, cancel := context.WithCancel(t.Context())
			want := error(context.Canceled)
			if deadline {
				cancel()
				ctx, cancel = context.WithTimeout(t.Context(), 100*time.Millisecond)
				want = context.DeadlineExceeded
			}

			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := rt.Run(ctx, api.NewSource("cancel.fql", "RETURN true"), api.WithParams(nil))
				done <- err
			}()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("hosted Run did not start")
			}

			if !deadline {
				cancel()
			}

			select {
			case err := <-done:
				if !errors.Is(err, want) {
					t.Fatalf("run context = %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("Run did not return after cancellation")
			}

			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("hosted Run leaked after cancellation")
			}
		})
	}
}

func TestWireCloseRetainsBothFailuresAndOrder(t *testing.T) {
	logicalErr := errors.New("logical cleanup")
	transportErr := errors.New("physical cleanup")
	var mu sync.Mutex
	var order []string
	remote := &apiruntime.Runtime{CloseFunc: func() error {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "logical")

		return logicalErr
	}}
	rt := &wireRuntime{remote: remote, transport: &wireTransport{connection: &closeCallback{fn: func() error {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, "physical")

		return transportErr
	}}}}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			err := rt.Close()
			if !errors.Is(err, logicalErr) || !errors.Is(err, transportErr) {
				t.Errorf("cleanup lost errors: %v", err)
			}
		})
	}

	group.Wait()
	if !reflect.DeepEqual(order, []string{"logical", "physical"}) {
		t.Fatalf("close order/repetition = %v", order)
	}
}

func TestAdapterOutputIdentity(t *testing.T) {
	failure := errors.New("failure")
	for _, tc := range []struct {
		output api.Output
		err    error
	}{
		{nil, failure},
		{&apiruntime.Output{Data: []byte("bytes"), ContentType: "opaque/type"}, nil},
	} {
		remote := &apiruntime.Runtime{RunFunc: func(context.Context, api.Source, ...api.SessionOption) (api.Output, error) {
			return tc.output, tc.err
		}}
		for _, rt := range []api.Runtime{&wireRuntime{remote: remote}, &Builtin{Runtime: remote}} {
			out, err := rt.Run(t.Context(), api.NewSource("name", "content"), api.WithParams(nil))
			if err != tc.err || out != tc.output {
				t.Fatalf("%T changed output or admission error: %v, %v", rt, out, err)
			}

			if out != nil {
				_ = out.Close()
			}
		}
	}
}

func TestWireRejectsNonportableParameters(t *testing.T) {
	cyclic := map[string]any{}
	cyclic["self"] = cyclic
	for _, params := range []map[string]any{
		{"object": map[any]any{1: "numeric key"}},
		{"cycle": cyclic},
	} {
		rt := &wireRuntime{remote: &apiruntime.Runtime{RunFunc: func(_ context.Context, _ api.Source, opts ...api.SessionOption) (api.Output, error) {
			_, err := apiruntime.NewParams(opts...)

			return nil, err
		}}}
		out, err := rt.Run(t.Context(), api.NewSource("params.fql", "RETURN true"), api.WithParams(params))
		if out != nil || err == nil {
			t.Fatalf("nonportable parameters accepted: %v, %v", out, err)
		}
	}
}
