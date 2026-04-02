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

	"github.com/urfave/cli/v3"
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

func TestRunCommandWithoutFilesShowsHelp(t *testing.T) {
	stdout, stderr, err := runCLI(t, "run")
	helpStdout, helpStderr, helpErr := runCLI(t, "run", "--help")

	assertExitCode(t, err, 1)

	if helpErr != nil {
		t.Fatalf("expected no error from help, got %v", helpErr)
	}

	assertContains(t, stdout, "lab run [options] [files...]")
	assertContains(t, stdout, "--files string")
	assertNotContains(t, stderr, "No help topic for 'run'")

	if stdout != helpStdout {
		t.Fatalf("expected bare run help to match --help output, got %q and %q", stdout, helpStdout)
	}

	if stderr != helpStderr {
		t.Fatalf("expected bare run stderr to match --help stderr, got %q and %q", stderr, helpStderr)
	}
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

func TestRootPositionalScriptShowsRootHelp(t *testing.T) {
	script := writeScript(t)

	helpStdout, helpStderr, helpErr := runCLI(t)
	stdout, stderr, err := runCLI(t, script)

	if helpErr != nil {
		t.Fatalf("expected no error from root help, got %v", helpErr)
	}

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "version")
	assertNotContains(t, stdout, "Done")

	if stdout != helpStdout {
		t.Fatalf("expected positional root invocation to match root help output, got %q and %q", stdout, helpStdout)
	}

	if stderr != helpStderr {
		t.Fatalf("expected positional root invocation stderr to match root help stderr, got %q and %q", stderr, helpStderr)
	}
}

func TestRootFilesFlagShowsUsageError(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "-f", script)

	assertErrorMessage(t, err, "flag provided but not defined: -f")

	assertContains(t, stdout, "Incorrect Usage: flag provided but not defined: -f")
	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "version")
	assertNotContains(t, stdout, "Done")
	assertContains(t, stdout, "--help, -h")
	assertContains(t, stdout, "OPTIONS:")
	if stderr != "" {
		t.Fatalf("expected stderr to be empty, got %q", stderr)
	}
}

func TestRootHelpShowsCommandsOnly(t *testing.T) {
	stdout, stderr, err := runCLI(t, "--help")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "version")
	assertNotContains(t, stdout, "--files string")
	assertNotContains(t, stdout, "--timeout uint")
	assertNotContains(t, stderr, "deprecated")
}

func TestRunHelpShowsExecutionFlags(t *testing.T) {
	stdout, stderr, err := runCLI(t, "run", "--help")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab run [options] [files...]")
	assertContains(t, stdout, "--files string")
	assertContains(t, stdout, "--timeout uint")
	assertContains(t, stdout, "--runtime string")
	assertNotContains(t, stderr, "deprecated")
}

func TestVersionHelpRemainsMinimal(t *testing.T) {
	stdout, stderr, err := runCLI(t, "version", "--help")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab version [options]")
	assertContains(t, stdout, "--runtime string")
	assertNotContains(t, stdout, "--timeout uint")
	assertNotContains(t, stdout, "--files string")
	assertNotContains(t, stderr, "deprecated")
}

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := newApp("test-version", &stdout, &stderr)
	app.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}
	err := app.Run(context.Background(), append([]string{"lab"}, args...))

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

func assertErrorMessage(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error %q, got nil", expected)
	}

	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}
