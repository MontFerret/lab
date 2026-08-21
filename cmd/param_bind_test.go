package cmd

import (
	"strings"
	"testing"
)

func TestToParamBindings(t *testing.T) {
	bindings, err := toParamBindings([]string{
		"baseUrl=@lab.static.fixtures",
		"config.api.token=@credentials.token",
		"config.database.dsn=@databaseDsn",
	}, nil)
	if err != nil {
		t.Fatalf("expected bindings to parse, got %v", err)
	}

	if len(bindings) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(bindings))
	}

	if bindings[0].Target.String() != "baseUrl" || bindings[0].Source.String() != "lab.static.fixtures" {
		t.Fatalf("unexpected first binding: %#v", bindings[0])
	}
}

func TestToParamBindingsRejectsMalformedDeclarations(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "missing separator", value: "baseUrl", expected: "expected <target>=@<source>"},
		{name: "target reference", value: "@baseUrl=@lab.static.fixtures", expected: "must not start with @"},
		{name: "malformed target", value: "config..baseUrl=@lab.static.fixtures", expected: "invalid parameter binding target"},
		{name: "reserved target", value: "lab.value=@source", expected: `"lab" is reserved`},
		{name: "literal source", value: "baseUrl=https://example.test", expected: "must start with @"},
		{name: "malformed source", value: "baseUrl=@lab..static", expected: "invalid parameter binding source"},
		{name: "computed source", value: `baseUrl=@lab.static["fixtures"]`, expected: "invalid parameter binding source"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toParamBindings([]string{tt.value}, nil)
			if err == nil || !strings.Contains(err.Error(), tt.expected) {
				t.Fatalf("expected error containing %q, got %v", tt.expected, err)
			}
		})
	}
}

func TestToParamBindingsRejectsDuplicateAndOverlappingTargets(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "duplicate",
			values:   []string{"baseUrl=@one", "baseUrl=@two"},
			expected: `duplicate parameter binding target "baseUrl"`,
		},
		{
			name:     "ancestor",
			values:   []string{"config.api=@one", "config.api.baseUrl=@two"},
			expected: `parameter binding target "config.api.baseUrl" overlaps target "config.api"`,
		},
		{
			name:     "descendant",
			values:   []string{"config.api.baseUrl=@one", "config.api=@two"},
			expected: `parameter binding target "config.api" overlaps target "config.api.baseUrl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toParamBindings(tt.values, nil)
			if err == nil || err.Error() != tt.expected {
				t.Fatalf("expected error %q, got %v", tt.expected, err)
			}
		})
	}
}

func TestToParamBindingsRejectsParamConflicts(t *testing.T) {
	tests := []struct {
		name        string
		binding     string
		param       string
		expectedErr string
	}{
		{
			name:        "same target",
			binding:     "baseUrl=@source",
			param:       `baseUrl:"https://example.test"`,
			expectedErr: `parameter target "baseUrl" is assigned by both --param and --param-bind`,
		},
		{
			name:        "param ancestor",
			binding:     "config.api.baseUrl=@source",
			param:       `config:{"api":{}}`,
			expectedErr: `parameter binding target "config.api.baseUrl" overlaps --param target "config"`,
		},
		{
			name:        "binding ancestor",
			binding:     "config.api=@source",
			param:       `config.api.baseUrl:"https://example.test"`,
			expectedErr: `parameter binding target "config.api" overlaps --param target "config.api.baseUrl"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toParamBindings([]string{tt.binding}, []string{tt.param})
			if err == nil || err.Error() != tt.expectedErr {
				t.Fatalf("expected error %q, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestToParamBindingsAllowsSiblingTargets(t *testing.T) {
	bindings, err := toParamBindings([]string{
		"config.api.baseUrl=@apiUrl",
		"config.api.token=@apiToken",
		"config.database.dsn=@databaseDsn",
	}, nil)
	if err != nil {
		t.Fatalf("expected sibling targets to be accepted, got %v", err)
	}

	if len(bindings) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(bindings))
	}
}

func TestToParamBindingsPreservesLegacyInvalidParamNames(t *testing.T) {
	bindings, err := toParamBindings([]string{"baseUrl=@source"}, []string{`invalid name:"value"`})
	if err != nil {
		t.Fatalf("expected invalid legacy param name not to affect bindings, got %v", err)
	}

	if len(bindings) != 1 {
		t.Fatalf("expected one binding, got %d", len(bindings))
	}
}
