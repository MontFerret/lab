package testing_test

import (
	"testing"
	"time"

	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func BenchmarkSuiteRun(b *testing.B) {
	rt, err := runtime.NewBuiltin(nil)
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() { _ = rt.Close() })
	suite, err := labtesting.NewSuite(labtesting.Options{
		File:    sources.File{Name: "benchmark.yaml", Content: []byte("query:\n  text: RETURN @value\nassert:\n  text: RETURN @lab.data.query.result\n")},
		Timeout: time.Second,
	})
	if err != nil {
		b.Fatal(err)
	}

	params := labtesting.NewParams()
	params.SetUserValue("value", 42)
	b.ReportAllocs()

	for b.Loop() {
		if err := suite.Run(b.Context(), rt, params.Clone()); err != nil {
			b.Fatal(err)
		}
	}
}
