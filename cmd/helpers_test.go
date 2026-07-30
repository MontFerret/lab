package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestExtractBinaryFlags(t *testing.T) {
	params := map[string]interface{}{
		"flags": []any{"--log-output=none", "--browser-headless"},
		"value": 1,
	}

	flags, err := extractBinaryFlags(params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(flags) != 2 || flags[0] != "--log-output=none" || flags[1] != "--browser-headless" {
		t.Fatalf("unexpected flags: %#v", flags)
	}
	if _, exists := params["flags"]; exists {
		t.Fatalf("expected flags to be removed from params: %#v", params)
	}
}

func TestExtractBinaryFlagsRejectsInvalidValues(t *testing.T) {
	invalidFlags := []any{"--ok", 1}
	params := map[string]interface{}{
		"flags": invalidFlags,
	}

	_, err := extractBinaryFlags(params)
	if err == nil || !strings.Contains(err.Error(), "invalid type of flags (expected array of strings)") {
		t.Fatalf("expected invalid flags error, got %v", err)
	}

	if value, exists := params["flags"]; !exists || !reflect.DeepEqual(value, invalidFlags) {
		t.Fatalf("expected invalid flags to remain unchanged, got %#v", params)
	}
}
