package runtime

import "testing"

func BenchmarkBuiltinLifecycle(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		rt, err := NewBuiltin(nil)
		if err != nil {
			b.Fatal(err)
		}

		if err := rt.engine.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
