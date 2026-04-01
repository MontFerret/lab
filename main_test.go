package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCommandExecutesScript(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", script)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertNotContains(t, stderr, "deprecated")
	assertContains(t, stdout, "Done")
}

func TestRunCommandExecutesFilesFlag(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", "-f", script)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertNotContains(t, stderr, "deprecated")
	assertContains(t, stdout, "Done")
}

func TestVersionCommand(t *testing.T) {
	stdout, stderr, err := runCLI(t, "version")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Version")
	assertContains(t, stdout, "Self: test-version")
	assertNotContains(t, stderr, "deprecated")
}

func TestDeprecatedRootExecutesPositionalScript(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, script)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stderr, "deprecated")
	assertContains(t, stderr, "lab run")
	assertContains(t, stdout, "Done")
}

func TestDeprecatedRootExecutesFilesFlag(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "-f", script)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stderr, "deprecated")
	assertContains(t, stderr, "lab run")
	assertContains(t, stdout, "Done")
}

func TestRootHelpShowsCommandsOnly(t *testing.T) {
	stdout, stderr, err := runCLI(t, "--help")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "version")
	assertNotContains(t, stdout, "--files value")
	assertNotContains(t, stdout, "--timeout value")
	assertNotContains(t, stderr, "deprecated")
}

func TestRunHelpShowsExecutionFlags(t *testing.T) {
	stdout, stderr, err := runCLI(t, "run", "--help")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab run [options] [files...]")
	assertContains(t, stdout, "--files value")
	assertContains(t, stdout, "--timeout value")
	assertContains(t, stdout, "--runtime value")
	assertNotContains(t, stderr, "deprecated")
}

func TestVersionHelpRemainsMinimal(t *testing.T) {
	stdout, stderr, err := runCLI(t, "version", "--help")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab version")
	assertNotContains(t, stdout, "--timeout value")
	assertNotContains(t, stdout, "--files value")
	assertNotContains(t, stderr, "deprecated")
}

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := newApp("test-version", &stdout, &stderr)
	err := app.RunContext(context.Background(), append([]string{"lab"}, args...))

	return stdout.String(), stderr.String(), err
}

func writeScript(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.fql")

	if err := os.WriteFile(path, []byte("RETURN 1"), 0o644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	return path
}

func assertContains(t *testing.T, output string, expected string) {
	t.Helper()

	if !strings.Contains(output, expected) {
		t.Fatalf("expected output to contain %q, got %q", expected, output)
	}
}

func assertNotContains(t *testing.T, output string, unexpected string) {
	t.Helper()

	if strings.Contains(output, unexpected) {
		t.Fatalf("expected output not to contain %q, got %q", unexpected, output)
	}
}
