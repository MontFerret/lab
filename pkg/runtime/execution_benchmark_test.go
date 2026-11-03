package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MontFerret/api"
)

func BenchmarkBuiltinRun(b *testing.B) {
	rt, err := NewBuiltin()
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() { _ = rt.Close() })
	src := api.NewSource("benchmark.fql", "RETURN [@shared, @value]")
	params := map[string]any{"value": 42, "shared": "value"}
	b.ReportAllocs()

	for b.Loop() {
		if _, err := runAndCollect(b.Context(), rt, src, api.WithParams(params)); err != nil {
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
		if _, err := runAndCollect(b.Context(), rt, src, api.WithParams(params)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTPConsumeLargeBody(b *testing.B) {
	data := make([]byte, 1<<20)
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(data)
	}))
	b.Cleanup(host.Close)
	rt, err := NewRemote(host.URL, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() { _ = rt.Close() })
	src := api.NewSource("large.fql", "RETURN 1")
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()

	for b.Loop() {
		out, err := rt.Run(b.Context(), src)
		if err != nil {
			b.Fatal(err)
		}

		var size int
		err = out.Consume(b.Context(), func(_ context.Context, chunk []byte) error {
			size += len(chunk)

			return nil
		})
		if err != nil || size != len(data) {
			b.Fatalf("delivered %d bytes: %v", size, err)
		}
	}
}
