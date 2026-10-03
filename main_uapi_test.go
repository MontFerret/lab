package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/internal/testutil/apiruntime"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
)

func TestHTTPCommandOneShotFQLAndYAML(t *testing.T) {
	var requests atomic.Int32
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		if req.Method != http.MethodPost || req.URL.Path != "/run" {
			t.Errorf("unexpected operation: %s %s", req.Method, req.URL.Path)
		}

		var query struct {
			Text   string
			Params map[string]any
		}
		if err := json.NewDecoder(req.Body).Decode(&query); err != nil {
			t.Error(err)
		}

		if query.Params["value"] != "per-run" {
			t.Errorf("parameters: %#v", query.Params)
		}

		// Suite decoding continues to use JSON bytes regardless of this header.
		w.Header().Set("Content-Type", "custom/opaque")
		_, _ = w.Write([]byte("true"))
	}))
	t.Cleanup(host.Close)
	for _, script := range []struct{ name, body string }{
		{"test.fql", "RETURN @value"},
		{"suite.yaml", "query:\n  text: RETURN @value\nassert:\n  text: RETURN @lab.data.query.result\n"},
	} {
		path := writeNamedScript(t, script.name, script.body)
		stdout, stderr, err := runCLI(t, "run", "--runtime="+host.URL+"/run", "--param=value:\"per-run\"", path)
		if err != nil || stderr != "" || !strings.Contains(stdout, "Passed") {
			t.Fatalf("HTTP command: %q %q %v", stdout, stderr, err)
		}
	}

	if requests.Load() != 3 {
		t.Fatalf("expected exactly three one-shot requests, got %d", requests.Load())
	}
}

func TestWireVersionCommandPreservesOpaqueValues(t *testing.T) {
	for _, value := range []api.Version{"", " opaque \n"} {
		host := wirehost.New(t, &apiruntime.Runtime{VersionValue: value})
		stdout, stderr, err := runCLI(t, "version", "--runtime=wire", "--runtime-endpoint="+host.Endpoint)
		if err != nil || stderr != "" || !strings.Contains(stdout, "  Runtime: "+string(value)+"\n") {
			t.Fatalf("opaque version: %q %q %v", stdout, stderr, err)
		}

		assertWireCommandCleanup(t, host)
	}
}
