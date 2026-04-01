package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
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

func TestVersionCommandUsesExplicitRuntimeOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/info" {
			t.Fatalf("expected /info request, got %s", r.URL.Path)
		}

		_, _ = w.Write([]byte(`{"version":{"ferret":"remote-version"}}`))
	}))
	defer srv.Close()

	stdout, stderr, err := runCLI(t, "version", "--runtime", srv.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Runtime: remote-version")
	assertNotContains(t, stderr, "deprecated")
}

func TestRootWithoutArgsShowsHelp(t *testing.T) {
	stdout, stderr, err := runCLI(t)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "version")
	assertNotContains(t, stderr, "Implicit script execution")
}

func TestRootPositionalScriptShowsMigrationError(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, script)

	assertExitCode(t, err, 1)

	assertContains(t, stderr, "Implicit script execution is no longer supported")
	assertContains(t, stderr, "lab run")
	assertContains(t, stdout, "lab [command] [command options]")
	assertNotContains(t, stdout, "Done")
}

func TestRootFilesFlagShowsMigrationError(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "-f", script)

	assertExitCode(t, err, 1)

	assertContains(t, stderr, "Implicit script execution is no longer supported")
	assertContains(t, stderr, "lab run")
	assertContains(t, stdout, "lab [command] [command options]")
	assertNotContains(t, stdout, "Done")
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

	assertContains(t, stdout, "lab version [options]")
	assertContains(t, stdout, "--runtime value")
	assertNotContains(t, stdout, "--timeout value")
	assertNotContains(t, stdout, "--files value")
	assertNotContains(t, stderr, "deprecated")
}

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := newApp("test-version", &stdout, &stderr)
	app.ExitErrHandler = func(_ *cli.Context, _ error) {}
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

func assertExitCode(t *testing.T, err error, expected int) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected exit code %d, got nil error", expected)
	}

	var exitErr cli.ExitCoder

	if !errors.As(err, &exitErr) {
		t.Fatalf("expected cli.ExitCoder, got %T (%v)", err, err)
	}

	if exitErr.ExitCode() != expected {
		t.Fatalf("expected exit code %d, got %d", expected, exitErr.ExitCode())
	}
}
