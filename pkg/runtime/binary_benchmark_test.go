package runtime

import "testing"

var benchmarkBinaryArgs []string

func BenchmarkBinaryArgumentAssembly(b *testing.B) {
	rt := &Binary{}
	params := map[string]any{
		"active": true,
		"count":  42,
		"labels": []string{"alpha", "beta", "gamma"},
		"name":   "ferret",
		"nested": map[string]any{"enabled": true, "limit": 10},
	}

	b.ReportAllocs()

	for b.Loop() {
		args, err := rt.paramsToArg(params)
		if err != nil {
			b.Fatal(err)
		}

		benchmarkBinaryArgs = args
	}
}
