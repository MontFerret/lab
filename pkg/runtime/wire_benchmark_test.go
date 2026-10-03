package runtime

import (
	"context"
	"testing"

	"github.com/MontFerret/api"
	"github.com/MontFerret/ferret/v2"
	"github.com/MontFerret/lab/v2/internal/testutil/wirehost"
)

func BenchmarkWire(b *testing.B) {
	host := wirehost.New(b, &wirehost.Runtime{VersionValue: "benchmark-host", RunFunc: func(context.Context, api.Source, ...api.SessionOption) (*api.Output, error) {
		return &api.Output{Content: []byte("[true]")}, nil
	}})
	opts := Options{Type: "wire", Endpoint: host.Endpoint, Params: map[string]any{"shared": "value"}}
	b.Run("ConstructionCleanup", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			rt, err := New(b.Context(), opts)
			if err != nil {
				b.Fatal(err)
			}

			if err := rt.Close(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("RepeatedRun", func(b *testing.B) {
		rt, err := New(b.Context(), opts)
		if err != nil {
			b.Fatal(err)
		}

		defer func() {
			if err := rt.Close(); err != nil {
				b.Fatal(err)
			}
		}()
		src := ferret.NewSource("benchmark.fql", "RETURN @shared")
		params := map[string]any{"perRun": "value"}
		b.ReportAllocs()
		for b.Loop() {
			if _, err := rt.Run(b.Context(), src, params); err != nil {
				b.Fatal(err)
			}
		}
	})
}
