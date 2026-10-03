package runtime

import (
	"reflect"
	"strings"
	"testing"

	"github.com/MontFerret/api"
)

func TestRuntimeRejectsRemovedBinarySelection(t *testing.T) {
	for _, selector := range []string{"bin:", "bin:./ferret", "bin:/missing/ferret", "bin://ferret", "bin:///usr/local/bin/ferret"} {
		t.Run(selector, func(t *testing.T) {
			rt, err := New(t.Context(), Options{Type: selector})
			if rt != nil || err == nil || !strings.Contains(err.Error(), "binary runtimes are no longer supported") {
				t.Fatalf("removed runtime %q returned %T, %v", selector, rt, err)
			}
		})
	}
}

func TestBuiltinFlagsAreSharedFQLParameters(t *testing.T) {
	shared := map[string]any{"flags": []any{"--browser-headless", true}}

	rt, err := New(t.Context(), Options{Params: shared})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = rt.Close() })
	source := api.NewSource("flags.fql", "RETURN @flags")

	out, err := rt.Run(t.Context(), source, api.WithParams(nil))
	if err != nil || string(out.Content) != `["--browser-headless",true]` {
		t.Fatalf("shared flags: %s, %v", out, err)
	}

	perRun := map[string]any{"flags": "per-run"}

	out, err = rt.Run(t.Context(), source, api.WithParams(perRun))
	if err != nil || string(out.Content) != `"per-run"` {
		t.Fatalf("per-run flags: %s, %v", out, err)
	}

	if !reflect.DeepEqual(shared, map[string]any{"flags": []any{"--browser-headless", true}}) || !reflect.DeepEqual(perRun, map[string]any{"flags": "per-run"}) {
		t.Fatalf("parameter inputs changed: %#v, %#v", shared, perRun)
	}
}
