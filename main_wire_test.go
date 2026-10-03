package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/MontFerret/api"
	"github.com/MontFerret/ferret/v2"
	native "github.com/MontFerret/ferret/v2/pkg/runtime"
	"github.com/MontFerret/ferret/v2/uapi"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
)

func TestWireVersionCommand(t *testing.T) {
	host := wirehost.New(t, &wirehost.Runtime{VersionValue: "v2.0.0-alpha.test"})
	stdout, stderr, err := runCLI(t, "version", "--runtime=wire", "--runtime-endpoint="+host.Endpoint)
	if err != nil || stderr != "" || !strings.Contains(stdout, "  Self: test-version\n  Runtime: v2.0.0-alpha.test\n") {
		t.Fatalf("Wire version: %q, %q, %v", stdout, stderr, err)
	}

	assertWireCommandCleanup(t, host)
}

func TestWireRunCommandFQLAndYAML(t *testing.T) {
	engine, err := ferret.New(ferret.WithFSRoot(t.TempDir()), ferret.WithFunctionsRegistrar(func(ns native.Namespace) {
		ns.Namespace("HOST").Function().A1().Add("IDENTITY", func(_ context.Context, value native.Value) (native.Value, error) {
			return value, nil
		})
	}))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = engine.Close() })
	host := wirehost.New(t, uapi.Wrap(engine, "native-test-host"))
	for name, content := range map[string]string{
		"test.fql": `RETURN T::EQ(HOST::IDENTITY(@value), "per-run")`,
		"suite.yaml": `query:
  text: RETURN HOST::IDENTITY(@value)
assert:
  text: RETURN T::EQ(@lab.data.query.result, "per-run")
`,
		"nested.yaml": `query:
  text: RETURN @nested.value
  params:
    nested:
      value: nested
assert:
  text: RETURN T::EQ(@lab.data.query.result, "nested")
`,
	} {
		t.Run(name, func(t *testing.T) {
			script := writeNamedScript(t, name, content)
			if name == "nested.yaml" {
				_, _, err := runCLI(t, "run", script)
				if err != nil {
					t.Fatalf("existing built-in nested YAML behavior: %v", err)
				}
			}

			stdout, stderr, err := runCLI(t, "run", "--runtime=wire", "--runtime-endpoint="+host.Endpoint,
				`--runtime-param=value:"shared"`, `--param=value:"per-run"`, script)
			if err != nil || stderr != "" || !strings.Contains(stdout, "Passed") || !strings.Contains(stdout, "Done") {
				t.Fatalf("Wire execution: %q, %q, %v", stdout, stderr, err)
			}

			assertWireCommandCleanup(t, host)
		})
	}
}

func TestWireCommandConfigurationAndHelp(t *testing.T) {
	host := wirehost.New(t, &wirehost.Runtime{VersionValue: "configured"})
	stdout, stderr, err := runCLIWithEnv(t, map[string]string{
		"LAB_RUNTIME": "wire", "LAB_RUNTIME_ENDPOINT": "tcp://127.0.0.1:1", "LAB_RUNTIME_CONNECT_TIMEOUT": "0s",
	}, "version", "--runtime-endpoint="+host.Endpoint, "--runtime-connect-timeout=1s")
	if err != nil || stderr != "" || !strings.Contains(stdout, "Runtime: configured") {
		t.Fatalf("flag precedence: %q, %q, %v", stdout, stderr, err)
	}

	stdout, stderr, err = runCLIWithEnv(t, map[string]string{
		"LAB_RUNTIME": "wire", "LAB_RUNTIME_ENDPOINT": host.Endpoint, "LAB_RUNTIME_CONNECT_TIMEOUT": "1s",
	}, "version")
	if err != nil || stderr != "" || !strings.Contains(stdout, "Runtime: configured") {
		t.Fatalf("environment configuration: %q, %q, %v", stdout, stderr, err)
	}

	assertWireCommandCleanup(t, host)
	for _, command := range []string{"run", "version"} {
		stdout, _, err := runCLI(t, command, "--help")
		if err != nil || !strings.Contains(stdout, "--runtime-endpoint") || !strings.Contains(stdout, "--runtime-connect-timeout") || !strings.Contains(stdout, "wire with --runtime-endpoint") || !strings.Contains(stdout, "LAB_RUNTIME_ENDPOINT") || !strings.Contains(stdout, "LAB_RUNTIME_CONNECT_TIMEOUT") {
			t.Fatalf("Wire help %s: %q, %v", command, stdout, err)
		}
	}
}

func TestWireCommandsRejectConfigurationBeforeConnecting(t *testing.T) {
	host := wirehost.New(t, &wirehost.Runtime{VersionValue: "unused"})
	script := writeScript(t)
	for _, extra := range [][]string{
		{"--runtime-connect-timeout=0s"},
		{"--runtime-endpoint=tcp://localhost:1"},
		{"--policy-fs-read-only=false"},
		{"--policy-http-no-timeout=false"},
		{`--runtime-param=headers:{}`},
		{`--runtime-param=cookies:{}`},
		{`--runtime-param=path:"/execute"`},
		{`--runtime-param=flags:[]`},
	} {
		args := append([]string{"run", "--runtime=wire", "--runtime-endpoint=" + host.Endpoint}, extra...)
		args = append(args, script)
		_, _, err := runCLI(t, args...)
		if err == nil {
			t.Fatalf("unsupported Wire configuration accepted: %v", extra)
		}
	}

	for _, args := range [][]string{
		{"version", "--runtime=wire"},
		{"version", "--runtime=wire", "--runtime-endpoint=" + host.Endpoint, "--runtime-connect-timeout=0s"},
		{"version", "--runtime-endpoint=" + host.Endpoint},
		{"run", "--runtime-endpoint=" + host.Endpoint, script},
		{"version", "--runtime-connect-timeout=5s"},
	} {
		_, _, err := runCLI(t, args...)
		if err == nil {
			t.Fatalf("invalid runtime configuration accepted: %v", args)
		}
	}

	if host.Connections() != 0 {
		t.Fatalf("invalid configuration connected %d times", host.Connections())
	}
}

func TestWireCommandsHonorConstructionCancellation(t *testing.T) {
	for _, command := range []string{"run", "version"} {
		t.Run(command, func(t *testing.T) {
			started := make(chan struct{})
			host := wirehost.New(t, &wirehost.Runtime{VersionFunc: func(ctx context.Context) (api.Version, error) {
				close(started)
				<-ctx.Done()

				return "", ctx.Err()
			}})
			args := []string{command, "--runtime=wire", "--runtime-endpoint=" + host.Endpoint}
			if command == "run" {
				args = append(args, writeScript(t))
			}

			_, _, done, cancel := startCLI(t, args...)
			defer cancel()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("Wire handshake did not start")
			}

			cancel()
			select {
			case err := <-done:
				// run wraps construction errors in the command framework's ExitCoder.
				if err == nil || (command == "version" && !errors.Is(err, context.Canceled)) || !strings.Contains(err.Error(), context.Canceled.Error()) {
					t.Fatalf("command ignored construction cancellation: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("command did not return after cancellation")
			}

			assertWireCommandCleanup(t, host)
		})
	}
}

func TestWireRunCommandCancellationAndFailureCleanup(t *testing.T) {
	for _, cancellation := range []bool{false, true} {
		t.Run(map[bool]string{false: "failure", true: "cancellation"}[cancellation], func(t *testing.T) {
			started := make(chan struct{})
			finished := make(chan struct{})
			host := wirehost.New(t, &wirehost.Runtime{VersionValue: "host", RunFunc: func(ctx context.Context, _ api.Source, _ ...api.SessionOption) (*api.Output, error) {
				close(started)
				defer close(finished)
				if cancellation {
					<-ctx.Done()

					return nil, ctx.Err()
				}

				return nil, errors.New("host execution failure")
			}})
			_, _, done, cancel := startCLI(t, "run", "--runtime=wire", "--runtime-endpoint="+host.Endpoint, writeScript(t))
			defer cancel()
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("hosted execution did not start")
			}

			if cancellation {
				cancel()
			}

			select {
			case err := <-done:
				if err == nil {
					t.Fatal("command accepted failed/canceled execution")
				}
			case <-time.After(time.Second):
				t.Fatal("command failed to settle")
			}

			select {
			case <-finished:
			case <-time.After(time.Second):
				t.Fatal("hosted execution leaked")
			}

			assertWireCommandCleanup(t, host)
		})
	}
}

func TestWireVersionCommandDeadline(t *testing.T) {
	host := wirehost.New(t, &wirehost.Runtime{VersionFunc: func(ctx context.Context) (api.Version, error) {
		<-ctx.Done()

		return "", ctx.Err()
	}})
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	var stdout, stderr bytes.Buffer
	app := newApp("test-version", &stdout, &stderr)
	app.ExitErrHandler = func(context.Context, *cli.Command, error) {}
	err := app.Run(ctx, []string{"lab", "version", "--runtime=wire", "--runtime-endpoint=" + host.Endpoint})
	if !errors.Is(err, context.DeadlineExceeded) || stdout.Len() != 0 {
		t.Fatalf("command deadline: %q, %v", stdout.String(), err)
	}

	assertWireCommandCleanup(t, host)
}

func assertWireCommandCleanup(t *testing.T, host *wirehost.Host) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := host.WaitForClientsClosed(ctx); err != nil {
		t.Fatalf("command leaked Wire transport: %v", err)
	}
}
