package runtime

import "testing"

var (
	benchmarkBinaryArgs    []string
	benchmarkBinaryRuntime *Binary
)

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

func BenchmarkBinaryArgumentConcatenation(b *testing.B) {
	opts := BinaryOptions{
		Path:  "ferret",
		Flags: []string{"--log-output=none", "--browser-headless"},
		Params: map[string]any{
			"active": true,
			"count":  42,
		},
		FSPolicy: &FileSystemPolicy{ReadOnly: pointerTo(false)},
		HTTPPolicy: &HTTPPolicy{
			AllowedHosts:   []string{"example.test"},
			DefaultHeaders: map[string]string{"X-Lab": "benchmark"},
		},
	}
	queryParams := map[string]any{
		"name":   "ferret",
		"labels": []string{"alpha", "beta", "gamma"},
	}

	b.Run("BaseArgs", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			rt, err := NewBinary(opts)
			if err != nil {
				b.Fatal(err)
			}

			benchmarkBinaryRuntime = rt
		}
	})

	rt, err := NewBinary(opts)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("RunArgs", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			args, runErr := rt.runArgs(queryParams)
			if runErr != nil {
				b.Fatal(runErr)
			}

			benchmarkBinaryArgs = args
		}
	})
}
