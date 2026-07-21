package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

func TestBuiltinFilesystemPolicyUsesConfiguredRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "fixture.txt"), []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}

	rt, err := New(Options{
		FSPolicy: &FileSystemPolicy{Root: root},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() { _ = rt.Close() })

	_, err = rt.Run(
		context.Background(),
		source.New("fs_root.fql", `RETURN TO_STRING(IO::FS::READ("fixture.txt"))`),
		nil,
	)
	if err != nil {
		t.Fatalf("expected configured root read to succeed, got %v", err)
	}
}

func TestBuiltinFilesystemPolicyEnforcesReadOnly(t *testing.T) {
	root := t.TempDir()
	rt, err := New(Options{
		FSPolicy: &FileSystemPolicy{Root: root, ReadOnly: pointerTo(true)},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() { _ = rt.Close() })

	_, err = rt.Run(
		context.Background(),
		source.New("fs_read_only.fql", `
IO::FS::WRITE("output.txt", TO_BINARY("blocked"))
RETURN true
`),
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "filesystem is read-only") {
		t.Fatalf("expected read-only error, got %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(root, "output.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("expected output file not to exist, got %v", statErr)
	}
}

func TestBuiltinFilesystemPolicyRejectsMissingRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	_, err := New(Options{
		FSPolicy: &FileSystemPolicy{Root: root},
	})
	if err == nil || !strings.Contains(err.Error(), "filesystem policy") || !strings.Contains(err.Error(), root) {
		t.Fatalf("expected filesystem root error, got %v", err)
	}
}

func TestBuiltinHTTPPolicyBlocksLocalhostByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("unexpected"))
	}))
	defer srv.Close()

	rt, err := NewBuiltin(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() { _ = rt.Close() })

	_, err = rt.Run(
		context.Background(),
		source.New("blocked_localhost.fql", fmt.Sprintf("RETURN IO::NET::HTTP::GET(%q)", srv.URL)),
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "localhost is not allowed") {
		t.Fatalf("expected localhost policy error, got %v", err)
	}
}

func TestBuiltinHTTPPolicyAllowsConfiguredLocalhost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	rt, err := NewBuiltin(nil, ferrethttp.WithAllowLocalhost(true))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() { _ = rt.Close() })

	_, err = rt.Run(
		context.Background(),
		source.New("allowed_localhost.fql", fmt.Sprintf("RETURN IO::NET::HTTP::GET(%q)", srv.URL)),
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestBuiltinHTTPPolicyRejectsInvalidConfiguration(t *testing.T) {
	_, err := NewBuiltin(nil, ferrethttp.WithAllowedHosts("bad host"))
	if err == nil || !strings.Contains(err.Error(), "WithAllowedHosts") {
		t.Fatalf("expected allowed-host policy error, got %v", err)
	}
}

func TestBuiltinHTTPPolicyEnforcesResponseLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("too large"))
	}))
	defer srv.Close()

	rt, err := NewBuiltin(
		nil,
		ferrethttp.WithAllowLocalhost(true),
		ferrethttp.WithMaxResponseSize(1),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	t.Cleanup(func() { _ = rt.Close() })

	_, err = rt.Run(
		context.Background(),
		source.New("response_limit.fql", fmt.Sprintf("RETURN IO::NET::HTTP::GET(%q)", srv.URL)),
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "response body exceeds") {
		t.Fatalf("expected response-size error, got %v", err)
	}
}

func TestBuiltinCloseSucceedsWithConfiguredNetwork(t *testing.T) {
	rt, err := NewBuiltin(nil, ferrethttp.WithAllowLocalhost(true))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := rt.Close(); err != nil {
		t.Fatalf("expected close to succeed, got %v", err)
	}
}

func TestNewRejectsHTTPPolicyForHTTPRuntime(t *testing.T) {
	_, err := New(Options{
		Type:       "http://example.test",
		HTTPPolicy: &HTTPPolicy{AllowLocalhost: pointerTo(true)},
	})
	if err == nil || err.Error() != "HTTP policy options are not supported by HTTP runtimes" {
		t.Fatalf("expected unsupported policy error, got %v", err)
	}
}

func TestNewRejectsFilesystemPolicyForHTTPRuntime(t *testing.T) {
	_, err := New(Options{
		Type:     "http://example.test",
		FSPolicy: &FileSystemPolicy{ReadOnly: pointerTo(true)},
	})
	if err == nil || err.Error() != "filesystem policy options are not supported by HTTP runtimes" {
		t.Fatalf("expected unsupported policy error, got %v", err)
	}
}

func TestNewRejectsBinaryFlagsForBuiltinRuntime(t *testing.T) {
	_, err := New(Options{
		BinaryFlags: []string{"--log-output=none"},
	})
	if err == nil || err.Error() != "binary flags are only supported by binary runtimes" {
		t.Fatalf("expected unsupported binary flags error, got %v", err)
	}
}
