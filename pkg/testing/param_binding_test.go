package testing_test

import (
	"reflect"
	"strings"
	"testing"

	testing2 "github.com/MontFerret/lab/v2/pkg/testing"
)

func TestParseParamPath(t *testing.T) {
	valid := []string{
		"baseUrl",
		"config.api.baseUrl",
		"lab.static.api-fixtures",
		"_internal.value_2",
	}

	for _, value := range valid {
		t.Run(value, func(t *testing.T) {
			path, err := testing2.ParseParamPath(value)
			if err != nil {
				t.Fatalf("expected valid path, got %v", err)
			}

			if path.String() != value {
				t.Fatalf("expected %q, got %q", value, path.String())
			}
		})
	}
}

func TestParseParamPathRejectsMalformedPaths(t *testing.T) {
	invalid := []string{"", ".config", "config.", "config..api", "config api", "@config", "config[api]", "1config"}

	for _, value := range invalid {
		t.Run(value, func(t *testing.T) {
			if _, err := testing2.ParseParamPath(value); err == nil {
				t.Fatalf("expected %q to be rejected", value)
			}
		})
	}
}

func TestParamPathOverlaps(t *testing.T) {
	var empty testing2.ParamPath
	config := mustParamPath(t, "config")
	api := mustParamPath(t, "config.api")
	baseURL := mustParamPath(t, "config.api.baseUrl")
	database := mustParamPath(t, "config.database")

	if !config.Overlaps(config) || !config.Overlaps(api) || !api.Overlaps(config) || !api.Overlaps(baseURL) {
		t.Fatal("expected identical and ancestor paths to overlap")
	}

	if api.Overlaps(database) || database.Overlaps(api) {
		t.Fatal("expected sibling paths not to overlap")
	}

	if empty.Overlaps(config) || config.Overlaps(empty) {
		t.Fatal("expected an empty path not to overlap a valid path")
	}
}

func TestParamsApplyBindings(t *testing.T) {
	params := testing2.NewParams()
	params.SetUserValues(map[string]any{
		"endpoint": "https://example.test",
		"payload": map[string]any{
			"active": true,
			"items":  []any{float64(1), "two"},
		},
		"nothing": nil,
	})
	params.SetSystemValue("static", map[string]any{
		"fixtures": "http://127.0.0.1:43123",
	})

	err := params.ApplyBindings([]testing2.ParamBinding{
		newParamBinding(t, "baseUrl", "lab.static.fixtures"),
		newParamBinding(t, "config.api.url", "endpoint"),
		newParamBinding(t, "config.api.payload", "payload"),
		newParamBinding(t, "config.database.value", "nothing"),
	})
	if err != nil {
		t.Fatalf("expected bindings to apply, got %v", err)
	}

	actual := params.ToMap()
	expected := map[string]any{
		"baseUrl": "http://127.0.0.1:43123",
		"config": map[string]any{
			"api": map[string]any{
				"url": "https://example.test",
				"payload": map[string]any{
					"active": true,
					"items":  []any{float64(1), "two"},
				},
			},
			"database": map[string]any{
				"value": nil,
			},
		},
		"endpoint": "https://example.test",
		"payload": map[string]any{
			"active": true,
			"items":  []any{float64(1), "two"},
		},
		"nothing": nil,
		"lab": map[string]any{
			"static": map[string]any{
				"fixtures": "http://127.0.0.1:43123",
			},
		},
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected materialized parameters:\nexpected: %#v\nactual:   %#v", expected, actual)
	}
}

func TestParamsApplyBindingsUsesOneSnapshotAndIsAtomic(t *testing.T) {
	params := testing2.NewParams()
	params.SetUserValue("source", true)

	err := params.ApplyBindings([]testing2.ParamBinding{
		newParamBinding(t, "first", "source"),
		newParamBinding(t, "second", "first"),
	})
	if err == nil || !strings.Contains(err.Error(), `source "@first" does not exist`) {
		t.Fatalf("expected missing snapshot source error, got %v", err)
	}

	actual := params.ToMap()
	if _, exists := actual["first"]; exists {
		t.Fatalf("expected failed application not to write first target: %#v", actual)
	}

	if _, exists := actual["second"]; exists {
		t.Fatalf("expected failed application not to write second target: %#v", actual)
	}
}

func TestParamsApplyBindingsRejectsMissingSource(t *testing.T) {
	params := testing2.NewParams()

	err := params.ApplyBindings([]testing2.ParamBinding{
		newParamBinding(t, "baseUrl", "lab.static.missing"),
	})
	if err == nil || err.Error() != `parameter binding target "baseUrl": source "@lab.static.missing" does not exist` {
		t.Fatalf("expected contextual missing source error, got %v", err)
	}
}

func TestParamsApplyBindingsRejectsExistingTarget(t *testing.T) {
	params := testing2.NewParams()
	params.SetUserValues(map[string]any{
		"source": true,
		"config": map[string]any{
			"enabled": false,
		},
	})

	err := params.ApplyBindings([]testing2.ParamBinding{
		newParamBinding(t, "config.enabled", "source"),
	})
	if err == nil || err.Error() != `parameter target "config.enabled" already exists` {
		t.Fatalf("expected existing target error, got %v", err)
	}
}

func TestParamsApplyBindingsRejectsReservedLabTarget(t *testing.T) {
	params := testing2.NewParams()
	params.SetUserValue("source", true)

	err := params.ApplyBindings([]testing2.ParamBinding{
		newParamBinding(t, "lab.value", "source"),
	})
	if err == nil || err.Error() != `parameter binding target "lab.value" uses reserved namespace "lab"` {
		t.Fatalf("expected reserved namespace error, got %v", err)
	}
}

func newParamBinding(t *testing.T, targetValue string, sourceValue string) testing2.ParamBinding {
	t.Helper()

	return testing2.ParamBinding{
		Target: mustParamPath(t, targetValue),
		Source: mustParamPath(t, sourceValue),
	}
}

func mustParamPath(t *testing.T, value string) testing2.ParamPath {
	t.Helper()

	path, err := testing2.ParseParamPath(value)
	if err != nil {
		t.Fatalf("failed to parse path %q: %v", value, err)
	}

	return path
}
