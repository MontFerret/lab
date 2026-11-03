package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
	labruntime "github.com/MontFerret/lab/v2/pkg/runtime"
)

func TestRuntimeOptionsHTTPCommands(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(map[bool]string{false: "HTTP", true: "HTTPS"}[secure], func(t *testing.T) {
			requests := make(chan *http.Request, 4)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r

				if r.Method == http.MethodGet {
					_, _ = io.WriteString(w, `{"version":{"ferret":"configured-version"}}`)
					return
				}

				var query struct {
					Params map[string]any `json:"params"`
				}

				if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
					t.Errorf("decode query: %v", err)
				}

				if query.Params["headers"] != "FQL headers" || query.Params["path"] != "FQL path" {
					t.Errorf("adapter/query values mixed: %#v", query.Params)
				}

				for _, name := range []string{"cookies", "connect-timeout"} {
					if _, exists := query.Params[name]; exists {
						t.Errorf("injected adapter option %q", name)
					}
				}

				_, _ = io.WriteString(w, "1")
			})
			var host *httptest.Server

			if secure {
				host = httptest.NewTLSServer(handler)
				original := http.DefaultClient
				http.DefaultClient = host.Client()
				t.Cleanup(func() { http.DefaultClient = original })
			} else {
				host = httptest.NewServer(handler)
			}

			t.Cleanup(host.Close)
			common := []string{"--runtime=" + host.URL + "/api", `--runtime-option=headers:{"Authorization":"Bearer a:b,c {\"x\":1}","X-Other":"other"}`, `--runtime-option=cookies:{"session":"fixture"}`}
			for _, override := range []bool{false, true} {
				args := append([]string{"run"}, common...)
				path := "/api"

				if override {
					args = append(args, `--runtime-option=path:"/v1/execute"`)
					path = "/v1/execute"
				}

				args = append(args, `--param=headers:"FQL headers"`, `--param=path:"FQL path"`, writeScript(t))
				stdout, stderr, err := runCLI(t, args...)
				if err != nil || stderr != "" || !strings.Contains(stdout, "Passed") {
					t.Fatalf("run: %q, %q, %v", stdout, stderr, err)
				}

				req := receiveRuntimeCall(t, requests)
				if req.Method != http.MethodPost || req.URL.Path != path || req.Header.Get("Authorization") != `Bearer a:b,c {"x":1}` || req.Header.Get("X-Other") != "other" {
					t.Fatalf("execution method, path, or headers changed: %s %s", req.Method, req.URL.Path)
				}

				cookie, err := req.Cookie("session")
				if err != nil || cookie.Value != "fixture" {
					t.Fatalf("cookie configuration: %v", err)
				}
			}

			args := append([]string{"version"}, common...)
			args = append(args, `--runtime-option=path:"/ignored"`)
			stdout, stderr, err := runCLI(t, args...)
			if err != nil || stderr != "" || stdout != "Version\n  Self: test-version\n  Runtime: configured-version\n" {
				t.Fatalf("version contract: %q, %q, %v", stdout, stderr, err)
			}

			req := receiveRuntimeCall(t, requests)
			if req.Method != http.MethodGet || req.URL.Path != "/api/info" || req.Header.Get("Authorization") == "" {
				t.Fatalf("version endpoint or headers: %s %s", req.Method, req.URL.Path)
			}
		})
	}
}

func TestRuntimeOptionEnvironmentAndRepetition(t *testing.T) {
	requests := make(chan *http.Request, 8)
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r

		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"version":{"ferret":"environment-version"}}`)
		} else {
			_, _ = io.WriteString(w, "1")
		}
	}))
	t.Cleanup(host.Close)
	env := map[string]string{
		"LAB_RUNTIME":        host.URL + "/api",
		"LAB_RUNTIME_OPTION": `headers:{"Authorization":"Bearer a:b,c","X-Env":"present"},path:"/env:execute"`,
	}
	for _, command := range []string{"run", "version"} {
		args := []string{command}

		if command == "run" {
			args = append(args, writeScript(t))
		}

		_, stderr, err := runCLIWithEnv(t, env, args...)
		if err != nil || stderr != "" {
			t.Fatalf("environment: %q, %v", stderr, err)
		}

		req := receiveRuntimeCall(t, requests)
		expected := "/api/info"

		if command == "run" {
			expected = "/env:execute"
		}

		if req.URL.Path != expected || req.Header.Get("Authorization") != "Bearer a:b,c" || req.Header.Get("X-Env") != "present" {
			t.Fatalf("environment values changed: %s", req.URL.Path)
		}

		args = []string{command, "-r", host.URL + "/override", `--runtime-option=headers:{"X-Early":"discarded"}`, `--runtime-option=headers:{"X-Late":"last"}`}

		if command == "run" {
			args = append(args, writeScript(t))
		}

		_, stderr, err = runCLIWithEnv(t, env, args...)
		if err != nil || stderr != "" {
			t.Fatalf("CLI precedence: %q, %v", stderr, err)
		}

		req = receiveRuntimeCall(t, requests)
		expected = "/override/info"

		if command == "run" {
			expected = "/override"
		}

		if req.URL.Path != expected || req.Header.Get("Authorization") != "" || req.Header.Get("X-Env") != "" || req.Header.Get("X-Early") != "" || req.Header.Get("X-Late") != "last" {
			t.Fatalf("CLI/environment or repeated-key precedence changed: %s", req.URL.Path)
		}
	}

	wire := wirehost.New(t, mustBuiltinRuntime(t))
	for _, command := range []string{"run", "version"} {
		args := []string{command}

		if command == "run" {
			args = append(args, writeScript(t))
		}

		stdout, stderr, err := runCLIWithEnv(t, map[string]string{"LAB_RUNTIME": "wire+" + wire.Endpoint, "LAB_RUNTIME_OPTION": `connect-timeout:"10s"`}, args...)
		if err != nil || stderr != "" || !strings.Contains(stdout, map[string]string{"run": "Passed", "version": "Runtime:"}[command]) {
			t.Fatalf("Wire environment: %q, %q, %v", stdout, stderr, err)
		}

		assertWireCommandCleanup(t, wire)
	}
}

func TestRuntimeOptionsRejectBeforeExternalSetup(t *testing.T) {
	var requests atomic.Int32
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = io.WriteString(w, "1")
	}))
	t.Cleanup(host.Close)
	wire := wirehost.New(t, mustBuiltinRuntime(t))
	for _, command := range []string{"run", "version"} {
		for _, tc := range []struct{ target, option string }{
			{"", `headers:{}`}, {"builtin", `connect-timeout:"5s"`},
			{"wire+" + wire.Endpoint, `headers:{"Authorization":"private-token"}`},
			{"wire+" + wire.Endpoint, `path:"/run"`}, {"wire+" + wire.Endpoint, `value:"shared"`},
			{"wire+" + wire.Endpoint, `connect-timeout:5`}, {"wire+" + wire.Endpoint, `connect-timeout:null`},
			{"wire+" + wire.Endpoint, `connect-timeout:"0s"`}, {"wire+" + wire.Endpoint, `connect-timeout:"-1s"`},
			{"wire+" + wire.Endpoint, `connect-timeout:"wrong"`},
			{host.URL, `connect-timeout:"5s"`}, {host.URL, `headers:[]`}, {host.URL, `headers:{"Authorization":42}`},
			{host.URL, `headers:{"Authorization":"private-token",broken}`}, {host.URL, `private-token`},
			{host.URL, `cookies:{"session":true}`}, {host.URL, `path:false`}, {host.URL, `path:"%"`},
		} {
			args := []string{command, "--runtime=" + tc.target, "--runtime-option=" + tc.option}

			if command == "run" {
				args = append(args, "--wait="+host.URL, "--wait-attempts=1", writeScript(t))
			}

			stdout, stderr, err := runCLI(t, args...)
			if err == nil || !strings.Contains(err.Error(), "runtime-option") && !strings.Contains(err.Error(), "runtime option") || strings.Contains(err.Error(), "private-token") || stdout != "" || stderr != "" {
				t.Fatalf("invalid option %q: %q, %q, %v", tc.option, stdout, stderr, err)
			}
		}
	}

	for _, flag := range []string{"--policy-fs-read-only=false", "--policy-http-no-timeout=false"} {
		_, _, err := runCLI(t, "run", "--runtime=wire+"+wire.Endpoint, flag, writeScript(t))
		if err == nil {
			t.Fatalf("unsupported Wire policy accepted: %s", flag)
		}
	}

	if requests.Load() != 0 || wire.Connections() != 0 {
		t.Fatalf("invalid configuration performed I/O: %d HTTP requests, %d Wire connections", requests.Load(), wire.Connections())
	}
}

func TestRuntimeRemovedInterface(t *testing.T) {
	for _, command := range []string{"run", "version"} {
		for _, flag := range []string{`--runtime-param=value:"old"`, `--rp=value:"old"`, `--runtime-endpoint=tcp://127.0.0.1:1`, `--runtime-connect-timeout=5s`} {
			args := []string{command, flag}

			if command == "run" {
				args = append(args, writeScript(t))
			}

			_, _, err := runCLI(t, args...)
			if err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
				t.Fatalf("removed flag accepted: %s, %v", flag, err)
			}
		}

		args := []string{command}

		if command == "run" {
			args = append(args, writeScript(t))
		}

		_, stderr, err := runCLIWithEnv(t, map[string]string{"LAB_RUNTIME_PARAM": "not valid JSON", "LAB_RUNTIME_ENDPOINT": "invalid", "LAB_RUNTIME_CONNECT_TIMEOUT": "invalid"}, args...)
		if err != nil || stderr != "" {
			t.Fatalf("obsolete environment configured runtime: %q, %v", stderr, err)
		}

		stdout, _, err := runCLI(t, command, "--help")
		if err != nil {
			t.Fatal(err)
		}

		for _, entry := range []string{"--runtime-option", "LAB_RUNTIME_OPTION", "LAB_RUNTIME", "wire+tcp://", "connect-timeout"} {
			assertContains(t, stdout, entry)
		}
		for _, entry := range []string{"--runtime-param", "--rp ", "--runtime-endpoint", "--runtime-connect-timeout", "LAB_RUNTIME_PARAM", "LAB_RUNTIME_ENDPOINT", "LAB_RUNTIME_CONNECT_TIMEOUT"} {
			assertNotContains(t, stdout, entry)
		}

		args = []string{command, "--runtime=wire"}

		if command == "run" {
			args = append(args, writeScript(t))
		}

		_, _, err = runCLI(t, args...)
		if err == nil || !strings.Contains(err.Error(), "unsupported --runtime target") {
			t.Fatalf("bare Wire selection accepted: %v", err)
		}
	}
}

func TestRuntimeParameterPipelineAcrossAdapters(t *testing.T) {
	engine := mustBuiltinRuntime(t)
	wireParams := make(chan map[string]any, 4)
	wire := wirehost.New(t, &apiruntime.Runtime{VersionValue: "test-host", RunFunc: func(ctx context.Context, src api.Source, options ...api.SessionOption) (api.Output, error) {
		params, err := apiruntime.NewParams(options...)
		if err != nil {
			return nil, err
		}

		wireParams <- map[string]any(params)

		return engine.Run(ctx, src, api.WithParams(map[string]any(params)))
	}})
	httpParams := make(chan map[string]any, 4)
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query struct {
			Text   string         `json:"text"`
			Params map[string]any `json:"params"`
		}

		if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
			t.Errorf("decode query: %v", err)
			w.WriteHeader(500)
			return
		}

		httpParams <- query.Params
		output, err := engine.Run(r.Context(), api.NewSource("HTTP test host", query.Text), api.WithParams(query.Params))
		if err != nil {
			t.Errorf("HTTP execution: %v", err)
			w.WriteHeader(500)
			return
		}

		content, err := output.Collect(r.Context())
		if err != nil {
			t.Errorf("HTTP collection: %v", err)
			w.WriteHeader(500)
			return
		}

		_, _ = w.Write(content.Data)
	}))
	t.Cleanup(host.Close)
	suite := `query:
  text: RETURN [@value, @headers, @path, @alias]
  params:
    value: query
    alias: query-bound
assert:
  text: RETURN [T::EQ(@value, "assert"), T::EQ(@headers, "FQL headers"), T::EQ(@path, "FQL path"), T::EQ(@alias, "bound"), T::EQ(@lab.data.query.result, ["query", "FQL headers", "FQL path", "query-bound"]), T::EQ(@lab.data.query.params.value, "query"), T::EQ(@lab.data.query.params.alias, "query-bound")]
  params:
    value: assert
`
	for _, tc := range []struct {
		name, target string
		options      []string
	}{
		{"builtin", "builtin", nil},
		{"HTTP", host.URL, []string{`--runtime-option=headers:{"X-Adapter":"configured"}`, `--runtime-option=path:"/execute"`}},
		{"Wire", "wire+" + wire.Endpoint, []string{`--runtime-option=connect-timeout:"1s"`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, file := range []struct{ name, content string }{
				{"unit.fql", `RETURN [T::EQ(@value,"cli"),T::EQ(@headers,"FQL headers"),T::EQ(@path,"FQL path"),T::EQ(@alias,"bound")]`},
				{"suite.yaml", suite},
			} {
				args := []string{"run", "--runtime=" + tc.target}
				args = append(args, tc.options...)
				args = append(args, `--param=value:"early"`, `--param=value:"cli"`, `--param=headers:"FQL headers"`, `--param=path:"FQL path"`, `--param=source:"bound"`, `--param-bind=alias=@source`, writeNamedScript(t, file.name, file.content))
				stdout, stderr, err := runCLI(t, args...)
				if err != nil || stderr != "" || !strings.Contains(stdout, "Passed") {
					t.Fatalf("parameter pipeline: %q, %q, %v", stdout, stderr, err)
				}
			}
		})
	}

	for _, received := range []<-chan map[string]any{httpParams, wireParams} {
		for range 3 {
			params := receiveRuntimeCall(t, received)
			expectedAlias := "bound"

			if params["value"] == "query" {
				expectedAlias = "query-bound"
			}

			if params["headers"] != "FQL headers" || params["path"] != "FQL path" || params["alias"] != expectedAlias {
				t.Fatalf("adapter changed query parameters: %#v", params)
			}

			for _, name := range []string{"cookies", "connect-timeout"} {
				if _, exists := params[name]; exists {
					t.Fatalf("adapter option %q became an FQL input", name)
				}
			}

			lab := params["lab"].(map[string]any)
			if !reflect.DeepEqual(lab["static"], map[string]any{}) {
				t.Fatalf("system namespace changed: %#v", lab)
			}
		}
	}

	assertWireCommandCleanup(t, wire)
}

func mustBuiltinRuntime(t *testing.T) api.Runtime {
	t.Helper()
	rt, err := labruntime.NewBuiltin()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })

	return rt
}

// Services record calls before responding, so a completed command must already
// have published its invocation. Fail immediately if adapter selection regresses.
func receiveRuntimeCall[T any](t *testing.T, calls <-chan T) T {
	t.Helper()

	select {
	case call := <-calls:
		return call
	default:
		t.Fatal("runtime returned without the expected service invocation")

		var zero T

		return zero
	}
}
