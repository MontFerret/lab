package runtime

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MontFerret/api"
)

func BenchmarkBuiltinRun(b *testing.B) {
	rt, err := NewBuiltin(map[string]any{"shared": "value"})
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() { _ = rt.Close() })
	src := api.NewSource("benchmark.fql", "RETURN [@shared, @value]")
	params := map[string]any{"value": 42}
	b.ReportAllocs()

	for b.Loop() {
		if _, err := rt.Run(b.Context(), src, api.WithParams(params)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTPRun(b *testing.B) {
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[42]"))
	}))
	b.Cleanup(host.Close)
	rt, err := NewRemote(host.URL, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() { _ = rt.Close() })
	src := api.NewSource("benchmark.fql", "RETURN @value")
	params := map[string]any{"value": 42}
	b.ReportAllocs()

	for b.Loop() {
		if _, err := rt.Run(b.Context(), src, api.WithParams(params)); err != nil {
			b.Fatal(err)
		}
	}
}
