package runtime

import (
	"context"
	"os"
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

	script := filepath.Join(t.TempDir(), "echo-cli.sh")
	content := "#!/bin/sh\nprintf 'arg:%s\\n' \"$@\"\nprintf 'stdin:'\ncat\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	rt, err := NewBinary(BinaryOptions{
		Path:   script,
		Flags:  []string{"--log-output=none"},
		Params: map[string]any{"limit": 3},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out, err := rt.Run(context.Background(), ferretsource.New("test.fql", "RETURN 1"), map[string]any{
		"foo": "bar",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := strings.Join([]string{
		"arg:run",
		"arg:--log-output=none",
		"arg:--param=limit=3",
		`arg:--param=foo="bar"`,
		"stdin:RETURN 1",
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
		Path:  "/tmp/ferret",
		Flags: []string{"--log-output=none"},
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

	args, err := rt.runArgs(map[string]any{"queryZeta": 4, "queryAlpha": 3})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{
		"run",
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

func TestBinaryArgumentsIncludeUnlimitedPolicyFlags(t *testing.T) {
	rt, err := NewBinary(BinaryOptions{
		Path: "/tmp/ferret",
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
		Path: "/tmp/ferret",
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
		Path:  "/tmp/ferret",
		Flags: []string{"--policy-http-allow-localhost=true"},
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

	rt, err := NewBinary(BinaryOptions{Path: script})
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
