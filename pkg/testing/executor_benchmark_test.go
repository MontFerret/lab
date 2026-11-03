package testing_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/MontFerret/api"
	"github.com/MontFerret/lab/v2/pkg/runtime"
	"github.com/MontFerret/lab/v2/pkg/sources"
	labtesting "github.com/MontFerret/lab/v2/pkg/testing"
)

func BenchmarkExecutor(b *testing.B) {
	for _, collect := range []bool{false, true} {
		for _, count := range []int{3, 10} {
			for _, reuse := range []bool{false, true} {
				name := fmt.Sprintf("%s/times=%d/%s", map[bool]string{false: "Consume", true: "Collect"}[collect], count, map[bool]string{false: "OneShot", true: "Compiled"}[reuse])
				b.Run(name, func(b *testing.B) {
					rt, err := runtime.NewBuiltin()
					if err != nil {
						b.Fatal(err)
					}

					b.Cleanup(func() { _ = rt.Close() })
					src := api.NewSource("benchmark.fql", "RETURN [@value, @shared]")
					params := map[string]any{"value": 42, "shared": "value"}
					b.ReportAllocs()

					for b.Loop() {
						executor := labtesting.NewExecutor(rt, reuse)
						for range count {
							if collect {
								_, err = executor.Collect(b.Context(), src, api.WithParams(params))
							} else {
								err = executor.Consume(b.Context(), src, api.WithParams(params))
							}

							if err != nil {
								break
							}
						}

						if err := errors.Join(err, executor.Close()); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		}
	}
}

func BenchmarkSuiteRepeated(b *testing.B) {
	for _, count := range []int{3, 10} {
		for _, reuse := range []bool{false, true} {
			b.Run(fmt.Sprintf("times=%d/%s", count, map[bool]string{false: "OneShot", true: "Compiled"}[reuse]), func(b *testing.B) {
				rt, err := runtime.NewBuiltin()
				if err != nil {
					b.Fatal(err)
				}

				b.Cleanup(func() { _ = rt.Close() })
				suite, err := labtesting.NewSuite(labtesting.Options{File: sources.File{Name: "benchmark.yaml", Content: []byte("query:\n  text: RETURN @value\nassert:\n  text: RETURN @lab.data.query.result\n")}, Timeout: time.Second})
				if err != nil {
					b.Fatal(err)
				}

				params := labtesting.NewParams()
				params.SetUserValue("value", 42)
				b.ReportAllocs()

				for b.Loop() {
					executor := labtesting.NewExecutor(rt, reuse)
					for range count {
						err = suite.Run(b.Context(), executor, params.Clone())
						if err != nil {
							break
						}
					}

					if err := errors.Join(err, executor.Close()); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
