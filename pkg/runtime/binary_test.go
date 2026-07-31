package runtime

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	stdruntime "runtime"
	"slices"
	"strings"
	"testing"
	"time"

	ferretsource "github.com/MontFerret/ferret/v2/pkg/source"
)

func TestBinaryRunUsesFerretCLIv2Contract(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	dir := t.TempDir()
	script := filepath.Join(dir, "echo-cli.sh")
	content := "#!/bin/sh\nprintf 'arg:%s\\n' \"$@\"\nprintf 'script:'\ncat \"$2\"\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	queryPath := filepath.Join(dir, "query with spaces.fql")
	queryContent := "RETURN 1"
	if err := os.WriteFile(queryPath, []byte(queryContent), 0o644); err != nil {
		t.Fatalf("failed to write query: %v", err)
	}

	rt, err := NewBinary(BinaryOptions{
		Path:     script,
		Protocol: ProtocolCLI,
		Flags:    []string{"--log-output=none"},
		Params:   map[string]any{"limit": 3},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out, err := rt.Run(context.Background(), ferretsource.New(queryPath, queryContent), map[string]any{
		"foo": "bar",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := strings.Join([]string{
		"arg:run",
		"arg:" + queryPath,
		"arg:--log-output=none",
		"arg:--param=limit=3",
		`arg:--param=foo="bar"`,
		"script:RETURN 1",
	}, "\n")
	if string(out) != want {
		t.Fatalf("unexpected CLI input:\nwant:\n%s\ngot:\n%s", want, out)
	}
}

func TestBinaryArgumentsIncludePoliciesAndAreDeterministic(t *testing.T) {
	duration := 2 * time.Second
	maxRequestSize := int64(32)
	maxResponseSize := int64(64)
	maxHeaderSize := int64(128)
	maxRedirects := 3

	rt, err := NewBinary(BinaryOptions{
		Path:     "/tmp/ferret",
		Protocol: ProtocolCLI,
		Flags:    []string{"--log-output=none"},
		Params: map[string]any{
			"zeta":  2,
			"alpha": 1,
		},
		FSPolicy: &FileSystemPolicy{
			Root:     "./fixtures",
			ReadOnly: pointerTo(false),
		},
		HTTPPolicy: &HTTPPolicy{
			AllowedSchemes:        []string{"https"},
			AllowedMethods:        []string{"GET", "POST"},
			AllowedHosts:          []string{"example.test", "api.example.test:8443"},
			BlockedHosts:          []string{"blocked.example.test"},
			AllowLocalhost:        pointerTo(true),
			AllowPrivateNetworks:  pointerTo(false),
			AllowLinkLocal:        pointerTo(false),
			DefaultHeaders:        map[string]string{"X-Zeta": "z", "X-Alpha": "a"},
			BlockedRequestHeaders: []string{"X-Secret", "X-Internal"},
			Timeout:               &duration,
			NoTimeout:             pointerTo(false),
			MaxRequestSize:        &maxRequestSize,
			UnlimitedRequestSize:  pointerTo(false),
			MaxResponseSize:       &maxResponseSize,
			UnlimitedResponseSize: pointerTo(false),
			MaxResponseHeaderSize: &maxHeaderSize,
			FollowRedirects:       pointerTo(false),
			MaxRedirects:          &maxRedirects,
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	args, err := rt.runArgs("test.fql", map[string]any{"queryZeta": 4, "queryAlpha": 3})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{
		"run",
		"test.fql",
		"--log-output=none",
		"--policy-fs-root=./fixtures",
		"--policy-fs-read-only=false",
		"--policy-http-allowed-schemes=https",
		"--policy-http-allowed-methods=GET,POST",
		"--policy-http-allowed-hosts=example.test,api.example.test:8443",
		"--policy-http-blocked-hosts=blocked.example.test",
		"--policy-http-allow-localhost=true",
		"--policy-http-allow-private-networks=false",
		"--policy-http-allow-link-local=false",
		`--policy-http-default-headers={"X-Alpha":"a","X-Zeta":"z"}`,
		"--policy-http-blocked-request-headers=X-Secret,X-Internal",
		"--policy-http-timeout=2s",
		"--policy-http-no-timeout=false",
		"--policy-http-max-request-size=32",
		"--policy-http-unlimited-request-size=false",
		"--policy-http-max-response-size=64",
		"--policy-http-unlimited-response-size=false",
		"--policy-http-max-response-header-size=128",
		"--policy-http-follow-redirects=false",
		"--policy-http-max-redirects=3",
		"--param=alpha=1",
		"--param=zeta=2",
		"--param=queryAlpha=3",
		"--param=queryZeta=4",
	}
	if !slices.Equal(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\ngot:  %#v", want, args)
	}
}

func TestNewBinaryDefaultsToCLIDirectProtocol(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{Path: "/tmp/ferret"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rt.protocol != ProtocolCLIDirect {
		t.Fatalf("expected %q protocol, got %q", ProtocolCLIDirect, rt.protocol)
	}

	args, err := rt.runArgs("test.fql", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !slices.Equal(args, []string{"test.fql"}) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBinaryCLIDirectArgumentsContainOnlyScriptPath(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{
		Path:     "/tmp/runtime",
		Protocol: ProtocolCLIDirect,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	params := map[string]any{
		"lab": map[string]any{
			"static": map[string]any{},
			"mock":   map[string]any{},
		},
	}
	args, err := rt.runArgs("path with spaces/test.fql", params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !slices.Equal(args, []string{"path with spaces/test.fql"}) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestNewBinaryRejectsCLIDirectPolicies(t *testing.T) {
	tests := []struct {
		name string
		opts BinaryOptions
		want string
	}{
		{
			name: "filesystem policy",
			opts: BinaryOptions{
				FSPolicy: &FileSystemPolicy{Root: "./fixtures"},
			},
			want: `runtime protocol "cli-direct" does not support filesystem policy options`,
		},
		{
			name: "HTTP policy",
			opts: BinaryOptions{
				HTTPPolicy: &HTTPPolicy{AllowLocalhost: pointerTo(true)},
			},
			want: `runtime protocol "cli-direct" does not support HTTP policy options`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Path = "/tmp/runtime"
			tt.opts.Protocol = ProtocolCLIDirect

			_, err := NewBinary(tt.opts)
			if err == nil || err.Error() != tt.want {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
		})
	}
}

func TestBinaryCLIDirectArgumentsIncludeFlagsAndParams(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{
		Path:  "/tmp/runtime",
		Flags: []string{"--log-level=debug", "--dry-run"},
		Params: map[string]any{
			"zeta":  2,
			"alpha": "shared",
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	args, err := rt.runArgs("path with spaces/test.fql", map[string]any{
		"lab":   map[string]any{"static": map[string]any{"app": "http://localhost"}},
		"alpha": "run",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{
		"--log-level=debug",
		"--dry-run",
		`--param=alpha:"shared"`,
		"--param=zeta:2",
		`--param=alpha:"run"`,
		`--param=lab:{"static":{"app":"http://localhost"}}`,
		"path with spaces/test.fql",
	}
	if !slices.Equal(args, want) {
		t.Fatalf("unexpected args:\nwant: %#v\ngot:  %#v", want, args)
	}
}

func TestBinaryCLIDirectReturnsParamSerializationErrorBeforeStartingProcess(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	marker := filepath.Join(t.TempDir(), "executed")
	t.Setenv("LAB_BINARY_EXECUTION_MARKER", marker)
	binary := writeExecutable(t, "#!/bin/sh\ntouch \"$LAB_BINARY_EXECUTION_MARKER\"\n")

	rt, err := NewBinary(BinaryOptions{Path: binary})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = rt.Run(
		context.Background(),
		ferretsource.New("generated", "RETURN 1"),
		map[string]any{"invalid": make(chan struct{})},
	)
	if err == nil || !strings.Contains(err.Error(), "failed to serialize parameter: invalid") {
		t.Fatalf("expected parameter serialization error, got %v", err)
	}

	if _, statErr := os.Stat(marker); !os.IsNotExist(statErr) {
		t.Fatalf("expected runtime not to execute, got %v", statErr)
	}
}

func TestBinaryRunMaterializesSyntheticSourceAndCleansUp(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "direct-runtime.sh")
	capturedPath := filepath.Join(dir, "script-path.txt")
	content := "#!/bin/sh\nprintf '%s' \"$1\" > \"$LAB_CAPTURED_SCRIPT\"\ncat \"$1\"\n"
	if err := os.WriteFile(binary, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}
	t.Setenv("LAB_CAPTURED_SCRIPT", capturedPath)

	rt, err := NewBinary(BinaryOptions{
		Path:     binary,
		Protocol: ProtocolCLIDirect,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out, err := rt.Run(
		context.Background(),
		ferretsource.New("https://example.test/query.fql", "RETURN 42"),
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "RETURN 42" {
		t.Fatalf("unexpected output: %q", out)
	}

	path, err := os.ReadFile(capturedPath)
	if err != nil {
		t.Fatalf("failed to read captured script path: %v", err)
	}
	if _, err := os.Stat(string(path)); !os.IsNotExist(err) {
		t.Fatalf("expected temporary script to be removed, got %v", err)
	}
}

func TestBinaryRunDoesNotFallbackAfterCLIFailure(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "failing-runtime.sh")
	callsPath := filepath.Join(dir, "calls.txt")
	argsPath := filepath.Join(dir, "args.txt")
	content := "#!/bin/sh\nprintf 'call\\n' >> \"$LAB_CALLS_PATH\"\nprintf '%s\\n' \"$@\" > \"$LAB_ARGS_PATH\"\nprintf 'failed'\nexit 7\n"
	if err := os.WriteFile(binary, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}
	t.Setenv("LAB_CALLS_PATH", callsPath)
	t.Setenv("LAB_ARGS_PATH", argsPath)

	queryPath := filepath.Join(dir, "test.fql")
	if err := os.WriteFile(queryPath, []byte("RETURN 1"), 0o644); err != nil {
		t.Fatalf("failed to write query: %v", err)
	}

	rt, err := NewBinary(BinaryOptions{Path: binary, Protocol: ProtocolCLI})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = rt.Run(context.Background(), ferretsource.New(queryPath, "RETURN 1"), nil)
	if err == nil || err.Error() != "failed" {
		t.Fatalf("expected process failure, got %v", err)
	}

	calls, err := os.ReadFile(callsPath)
	if err != nil {
		t.Fatalf("failed to read calls: %v", err)
	}
	if string(calls) != "call\n" {
		t.Fatalf("expected exactly one process invocation, got %q", calls)
	}

	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("failed to read args: %v", err)
	}
	if string(args) != "run\n"+queryPath+"\n" {
		t.Fatalf("unexpected invocation args: %q", args)
	}
}

func TestBinaryRunPreservesProcessEnvironmentAndWorkingDirectory(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	dir := t.TempDir()
	binary := filepath.Join(dir, "environment-runtime.sh")
	content := "#!/bin/sh\nprintf '%s\\n%s' \"$LAB_BINARY_ENV_TEST\" \"$PWD\"\n"
	if err := os.WriteFile(binary, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}
	t.Setenv("LAB_BINARY_ENV_TEST", "inherited")

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	rt, err := NewBinary(BinaryOptions{
		Path:     binary,
		Protocol: ProtocolCLIDirect,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out, err := rt.Run(context.Background(), ferretsource.New("generated", "RETURN 1"), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "inherited\n"+workingDirectory {
		t.Fatalf("unexpected process environment: %q", out)
	}
}

func TestBinaryRunPreservesCombinedOutputAndExitCodeHandling(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	t.Run("combined output", func(t *testing.T) {
		capturedPath := filepath.Join(t.TempDir(), "script-path.txt")
		t.Setenv("LAB_CAPTURED_SCRIPT", capturedPath)
		binary := writeExecutable(t, "#!/bin/sh\nprintf '%s' \"$1\" > \"$LAB_CAPTURED_SCRIPT\"\nprintf 'stdout'\nprintf 'stderr' >&2\nexit 7\n")
		rt, err := NewBinary(BinaryOptions{
			Path:     binary,
			Protocol: ProtocolCLIDirect,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = rt.Run(context.Background(), ferretsource.New("generated", "RETURN 1"), nil)
		if err == nil || err.Error() != "stdoutstderr" {
			t.Fatalf("expected combined output error, got %v", err)
		}

		scriptPath, err := os.ReadFile(capturedPath)
		if err != nil {
			t.Fatalf("failed to read captured script path: %v", err)
		}
		if _, err := os.Stat(string(scriptPath)); !os.IsNotExist(err) {
			t.Fatalf("expected failed invocation snapshot to be removed, got %v", err)
		}
	})

	t.Run("exit code without output", func(t *testing.T) {
		binary := writeExecutable(t, "#!/bin/sh\nexit 7\n")
		rt, err := NewBinary(BinaryOptions{
			Path:     binary,
			Protocol: ProtocolCLIDirect,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = rt.Run(context.Background(), ferretsource.New("generated", "RETURN 1"), nil)
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
			t.Fatalf("expected exit code 7, got %v", err)
		}
	})
}

func TestBinaryRunHonorsContextCancellation(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	capturedPath := filepath.Join(t.TempDir(), "script-path.txt")
	t.Setenv("LAB_CAPTURED_SCRIPT", capturedPath)
	binary := writeExecutable(t, "#!/bin/sh\nprintf '%s' \"$1\" > \"$LAB_CAPTURED_SCRIPT\"\nexec sleep 10\n")
	rt, err := NewBinary(BinaryOptions{
		Path:     binary,
		Protocol: ProtocolCLIDirect,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	done := make(chan error, 1)
	go func() {
		_, runErr := rt.Run(ctx, ferretsource.New("generated", "RETURN 1"), nil)
		done <- runErr
	}()

	var scriptPath []byte
	deadline := time.After(2 * time.Second)
	for len(scriptPath) == 0 {
		scriptPath, _ = os.ReadFile(capturedPath)
		if len(scriptPath) > 0 {
			break
		}

		select {
		case <-deadline:
			t.Fatal("timed out waiting for runtime process to start")
		case <-time.After(10 * time.Millisecond):
		}
	}

	cancel()
	err = <-done
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("expected prompt cancellation, took %s", elapsed)
	}

	if _, err := os.Stat(string(scriptPath)); !os.IsNotExist(err) {
		t.Fatalf("expected cancelled invocation snapshot to be removed, got %v", err)
	}
}

func TestBinaryArgumentsIncludeUnlimitedPolicyFlags(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{
		Path:     "/tmp/ferret",
		Protocol: ProtocolCLI,
		HTTPPolicy: &HTTPPolicy{
			NoTimeout:             pointerTo(true),
			UnlimitedRequestSize:  pointerTo(true),
			UnlimitedResponseSize: pointerTo(true),
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, expected := range []string{
		"--policy-http-no-timeout=true",
		"--policy-http-unlimited-request-size=true",
		"--policy-http-unlimited-response-size=true",
	} {
		if !slices.Contains(rt.baseArgs, expected) {
			t.Fatalf("expected args to contain %q, got %#v", expected, rt.baseArgs)
		}
	}
}

func TestBinaryArgumentsPreserveExplicitEmptyPolicyCollections(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{
		Path:     "/tmp/ferret",
		Protocol: ProtocolCLI,
		HTTPPolicy: &HTTPPolicy{
			AllowedHosts:   []string{},
			DefaultHeaders: map[string]string{},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, expected := range []string{
		"--policy-http-allowed-hosts=",
		"--policy-http-default-headers={}",
	} {
		if !slices.Contains(rt.baseArgs, expected) {
			t.Fatalf("expected args to contain %q, got %#v", expected, rt.baseArgs)
		}
	}
}

func TestNewBinaryRejectsRawManagedPolicyConflicts(t *testing.T) {
	tests := []struct {
		name   string
		flags  []string
		fs     *FileSystemPolicy
		http   *HTTPPolicy
		wanted string
	}{
		{
			name:   "filesystem exact flag",
			flags:  []string{"--policy-fs-root=/raw"},
			fs:     &FileSystemPolicy{Root: "/managed"},
			wanted: "--policy-fs-root",
		},
		{
			name:   "timeout pair",
			flags:  []string{"--policy-http-timeout=1s"},
			http:   &HTTPPolicy{NoTimeout: pointerTo(true)},
			wanted: "--policy-http-timeout",
		},
		{
			name:   "request limit pair",
			flags:  []string{"--policy-http-max-request-size=1"},
			http:   &HTTPPolicy{UnlimitedRequestSize: pointerTo(true)},
			wanted: "--policy-http-max-request-size",
		},
		{
			name:   "response limit pair",
			flags:  []string{"--policy-http-unlimited-response-size"},
			http:   &HTTPPolicy{MaxResponseSize: pointerTo(int64(1))},
			wanted: "--policy-http-unlimited-response-size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBinary(BinaryOptions{
				Path:       "/tmp/ferret",
				Protocol:   ProtocolCLI,
				Flags:      tt.flags,
				FSPolicy:   tt.fs,
				HTTPPolicy: tt.http,
			})
			if err == nil || !strings.Contains(err.Error(), tt.wanted) {
				t.Fatalf("expected %q conflict, got %v", tt.wanted, err)
			}
		})
	}
}

func TestNewBinaryAllowsRawPolicyFlagWhenUnmanaged(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{
		Path:     "/tmp/ferret",
		Protocol: ProtocolCLI,
		Flags:    []string{"--policy-http-allow-localhost=true"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !slices.Contains(rt.baseArgs, "--policy-http-allow-localhost=true") {
		t.Fatalf("expected raw policy flag, got %#v", rt.baseArgs)
	}
}

func TestNewBinaryRejectsInvalidConfiguration(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		_, err := NewBinary(BinaryOptions{})
		if err == nil || err.Error() != "binary runtime path cannot be empty" {
			t.Fatalf("expected empty path error, got %v", err)
		}
	})

	t.Run("HTTP policy", func(t *testing.T) {
		_, err := NewBinary(BinaryOptions{
			Path:       "/tmp/ferret",
			Protocol:   ProtocolCLI,
			HTTPPolicy: &HTTPPolicy{AllowedHosts: []string{"bad host"}},
		})
		if err == nil || !strings.Contains(err.Error(), "WithAllowedHosts") {
			t.Fatalf("expected HTTP policy error, got %v", err)
		}
	})

	t.Run("shared parameter", func(t *testing.T) {
		_, err := NewBinary(BinaryOptions{
			Path:   "/tmp/ferret",
			Params: map[string]any{"invalid": make(chan struct{})},
		})
		if err == nil || !strings.Contains(err.Error(), "failed to serialize parameter: invalid") {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})
}

func TestNewResolvesRelativeBinaryPath(t *testing.T) {
	rt, err := New(Options{Type: "bin:./ferret"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	binary, ok := rt.(*Binary)
	if !ok {
		t.Fatalf("expected binary runtime, got %T", rt)
	}

	if binary.path != "./ferret" {
		t.Fatalf("expected relative path, got %q", binary.path)
	}
}

func TestBinaryVersionUsesVersionCommand(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	script := filepath.Join(t.TempDir(), "version.sh")
	content := "#!/bin/sh\n[ \"$1\" = \"version\" ] || exit 2\nprintf 'v2-test\\n'\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	for _, protocol := range []Protocol{"", ProtocolCLI, ProtocolCLIDirect} {
		rt, err := NewBinary(BinaryOptions{
			Path:     script,
			Protocol: protocol,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		version, err := rt.Version(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if version != "v2-test" {
			t.Fatalf("expected version output, got %q", version)
		}
	}
}

func writeExecutable(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "runtime.sh")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	return path
}
