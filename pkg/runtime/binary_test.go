package runtime

import (
	"context"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"testing"

	ferretsource "github.com/MontFerret/ferret/v2/pkg/source"
)

func TestBinaryRunPassesThroughRawFlags(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("shell script test is Unix-only")
	}

	script := filepath.Join(t.TempDir(), "echo-args.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\"\n"), 0o755); err != nil {
		t.Fatalf("failed to write helper script: %v", err)
	}

	rt, err := NewBinary(script, "http://127.0.0.1:9222", map[string]any{
		"flags": []any{"--timeout=60", "--verbose"},
		"limit": 3,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	out, err := rt.Run(context.Background(), ferretsource.New("test.fql", "RETURN 1"), map[string]any{
		"foo": "bar",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	seen := make(map[string]bool, len(lines))

	for _, line := range lines {
		seen[line] = true
	}

	for _, expected := range []string{
		"--browser-address=http://127.0.0.1:9222",
		"--timeout=60",
		"--verbose",
		"--param=limit:3",
		"--param=foo:\"bar\"",
	} {
		if !seen[expected] {
			t.Fatalf("expected output to contain %q, got %q", expected, string(out))
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "--param=flags:") {
			t.Fatalf("expected flags not to be serialized as --param, got %q", line)
		}
	}
}

func TestNewBinaryRejectsInvalidFlagTypes(t *testing.T) {
	_, err := NewBinary("/tmp/ferret", "http://127.0.0.1:9222", map[string]any{
		"flags": []any{"--ok", 1},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "invalid type of flags (expected array of strings)") {
		t.Fatalf("unexpected error: %v", err)
	}
}
