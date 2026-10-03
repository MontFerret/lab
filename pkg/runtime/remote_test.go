package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/MontFerret/api"
)

func TestHTTPOptionsAndProtocol(t *testing.T) {
	rt, err := NewRemote("http://worker.test/base", map[string]any{
		"path": "/execute", "headers": map[string]any{"X-Lab": "value"}, "cookies": map[string]any{"token": "fixture"},
	})
	if err != nil {
		t.Fatal(err)
	}

	shared := map[string]any{"overlap": "early", "retained": 42}
	later := map[string]any{"overlap": "late"}
	original := maps.Clone(shared)
	var requests int
	rt.client = &http.Client{Transport: responseTransport{fn: func(req *http.Request) (*http.Response, error) {
		requests++
		if req.Header.Get("X-Lab") != "value" || req.Header.Get("Content-Type") != "application/json" || len(req.Cookies()) != 1 {
			t.Fatalf("request headers/cookies: %v", req.Header)
		}

		data := " opaque \n"
		if req.Method == http.MethodGet {
			if req.URL.String() != "http://worker.test/base/info" {
				t.Fatalf("info URL: %s", req.URL)
			}

			encoded, _ := json.Marshal(remoteInfo{Version: remoteVersion{Ferret: data}})
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(encoded)), Header: http.Header{}}, nil
		}

		var query remoteQuery
		if err := json.NewDecoder(req.Body).Decode(&query); err != nil {
			t.Fatal(err)
		}

		if req.Method != http.MethodPost || req.URL.String() != "http://worker.test/execute" || query.Text != "\r\nRETURN @overlap\n" ||
			!reflect.DeepEqual(query.Params, map[string]any{"overlap": "late", "retained": float64(42), "single": "value"}) {
			t.Fatalf("execution request: %s %s %#v", req.Method, req.URL, query)
		}

		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(data)), Header: http.Header{"Content-Type": {"custom/opaque"}}}, nil
	}}}
	var order []int
	for range 2 {
		out, err := rt.Run(t.Context(), api.Source{Name: "日本.fql", Content: "\r\nRETURN @overlap\n"},
			api.WithParams(shared), nil,
			func(opts api.SessionOptions) error { order = append(order, 1); return opts.SetParam("single", "value") },
			func(opts api.SessionOptions) error { order = append(order, 2); return opts.SetParams(later) })
		if err != nil || out == nil || string(out.Content) != " opaque \n" || out.ContentType != "custom/opaque" {
			t.Fatalf("output: %#v, %v", out, err)
		}
	}

	if !reflect.DeepEqual(order, []int{1, 2, 1, 2}) || !reflect.DeepEqual(shared, original) || len(later) != 1 {
		t.Fatalf("callback/input preservation: %v %#v %#v", order, shared, later)
	}

	if err := rt.Close(); err != nil {
		t.Fatal(err)
	}

	value, err := rt.Version(t.Context())
	if err != nil || value != " opaque \n" || requests != 3 {
		t.Fatalf("version/borrowed client: %q, %v, %d requests", value, err, requests)
	}

	if rt.params.Headers.Get("Cookie") != "" {
		t.Fatal("requests mutated shared headers")
	}
}

func TestHTTPUnsupportedAndCanceledOperationsMakeNoRequests(t *testing.T) {
	rt, err := NewRemote("http://worker.test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rt.client = &http.Client{Transport: responseTransport{fn: func(*http.Request) (*http.Response, error) {
		t.Error("unsupported or canceled operation made an HTTP request")

		return nil, errors.New("unexpected request")
	}}}
	first := errors.New("first")
	second := errors.New("second")
	var order []int
	out, err := rt.Run(t.Context(), api.Source{},
		func(api.SessionOptions) error { order = append(order, 1); return first },
		api.WithFSRoot(""), nil, api.WithOutputContentType(""),
		func(api.SessionOptions) error { order = append(order, 2); return second })
	if out != nil || !errors.Is(err, first) || !errors.Is(err, second) || !errors.Is(err, errors.ErrUnsupported) ||
		!reflect.DeepEqual(order, []int{1, 2}) {
		t.Fatalf("joined option failures: %#v %v %v", out, err, order)
	}

	for _, setter := range []api.SessionOption{api.WithFSRoot(""), api.WithOutputContentType("")} {
		out, err := rt.Run(t.Context(), api.Source{}, setter)
		if out != nil || !errors.Is(err, errors.ErrUnsupported) {
			t.Fatalf("unsupported session option: %#v %v", out, err)
		}
	}

	for _, compile := range []func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error){rt.Compile, rt.CompileDebug} {
		plan, err := compile(t.Context(), api.Source{}, func(api.PlanOptions) error {
			t.Error("unsupported compilation invoked options")

			return first
		})
		if plan != nil || !errors.Is(err, errors.ErrUnsupported) {
			t.Fatalf("unsupported compilation: %v, %v", plan, err)
		}
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
		_, err := rt.Run(ctx, api.Source{}, func(api.SessionOptions) error {
			t.Error("canceled Run invoked options")

			return first
		})
		if !errors.Is(err, want) {
			t.Fatalf("Run cancellation: %v", err)
		}

		_, err = rt.Version(ctx)
		if !errors.Is(err, want) {
			t.Fatalf("Version cancellation: %v", err)
		}

		for _, compile := range []func(context.Context, api.Source, ...api.PlanOption) (api.Plan, error){rt.Compile, rt.CompileDebug} {
			_, err := compile(ctx, api.Source{})
			if !errors.Is(err, want) || errors.Is(err, errors.ErrUnsupported) {
				t.Fatalf("compilation cancellation: %v", err)
			}
		}
	}
}

func TestHTTPResponseOutputAndCleanup(t *testing.T) {
	readFailure := errors.New("read failed")
	closeFailure := errors.New("close failed")
	for _, tc := range []struct {
		name       string
		content    string
		readErr    error
		closeErr   error
		status     int
		wantOutput bool
	}{
		{"empty", "", nil, nil, 200, true},
		{"bytes", "\x00\xff\n", nil, nil, 201, true},
		{"partial", "partial", readFailure, nil, 200, true},
		{"partial and close", "partial", readFailure, closeFailure, 200, true},
		{"read without output", "", readFailure, closeFailure, 200, false},
		{"close with output", "available", nil, closeFailure, 200, true},
		{"status and close", "ignored", nil, closeFailure, 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt, err := NewRemote("http://worker.test", nil)
			if err != nil {
				t.Fatal(err)
			}

			body := &responseBody{reader: strings.NewReader(tc.content), readErr: tc.readErr, closeErr: tc.closeErr}
			rt.client = &http.Client{Transport: responseTransport{fn: func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Status: "500 failure", Body: body, Header: http.Header{"Content-Type": {"opaque/type"}}}, nil
			}}}
			out, err := rt.Run(t.Context(), api.Source{})
			if (out != nil) != tc.wantOutput || body.closes != 1 || (tc.readErr != nil && !errors.Is(err, tc.readErr)) ||
				(tc.closeErr != nil && !errors.Is(err, tc.closeErr)) || (tc.status == 500 && (err == nil || !strings.Contains(err.Error(), "500 failure"))) {
				t.Fatalf("response output/errors/cleanup: %#v %v closes=%d", out, err, body.closes)
			}

			if out != nil && (string(out.Content) != tc.content || out.ContentType != "opaque/type") {
				t.Fatalf("changed output: %#v", out)
			}
		})
	}
}

func TestHTTPVersionRetainsCloseFailure(t *testing.T) {
	rt, err := NewRemote("http://worker.test", nil)
	if err != nil {
		t.Fatal(err)
	}

	failure := errors.New("close failed")
	body := &responseBody{reader: strings.NewReader("{\"version\":{\"ferret\":\" exact \"}}"), closeErr: failure}
	rt.client = &http.Client{Transport: responseTransport{fn: func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body}, nil
	}}}
	value, err := rt.Version(t.Context())
	if value != " exact " || !errors.Is(err, failure) || body.closes != 1 {
		t.Fatalf("version cleanup: %q %v closes=%d", value, err, body.closes)
	}
}

func TestHTTPParametersNilEmptyAndRequestIsolation(t *testing.T) {
	rt, err := NewRemote("http://worker.test", nil)
	if err != nil {
		t.Fatal(err)
	}

	var query remoteQuery
	rt.client = &http.Client{Transport: responseTransport{fn: func(req *http.Request) (*http.Response, error) {
		if err := json.NewDecoder(req.Body).Decode(&query); err != nil {
			t.Fatal(err)
		}

		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil
	}}}
	for _, params := range []map[string]any{{"previous": true}, nil, {}} {
		_, err := rt.Run(t.Context(), api.Source{}, api.WithParams(params))
		if err != nil || (query.Params == nil) != (params == nil) || len(query.Params) != len(params) {
			t.Fatalf("request parameters: %#v %v", query.Params, err)
		}
	}
}
