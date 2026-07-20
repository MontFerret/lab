package runtime

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ferrethttp "github.com/MontFerret/ferret/v2/pkg/net/http"
	"github.com/MontFerret/ferret/v2/pkg/source"
)

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

func TestNewRejectsHTTPPolicyForExternalRuntimes(t *testing.T) {
	for _, runtimeURL := range []string{"http://example.test", "bin:/usr/local/bin/ferret"} {
		t.Run(runtimeURL, func(t *testing.T) {
			_, err := New(Options{
				Type:       runtimeURL,
				HTTPPolicy: []ferrethttp.PolicyOption{ferrethttp.WithAllowLocalhost(true)},
			})
			if err == nil || !strings.Contains(err.Error(), "only supported by the built-in runtime") {
				t.Fatalf("expected unsupported policy error, got %v", err)
			}
		})
	}
}
