package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

type safeBuffer struct {
	mu  sync.RWMutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.buf.String()
}

func TestRunCommandExecutesScript(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", script)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRunCommandExecutesFilesFlag(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", "-f", script)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
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
	assertContains(t, stdout, "--serve string")
	assertNotContains(t, stdout, "--cdn")
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
	assertEqual(t, stderr, "")
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
	assertEqual(t, stderr, "")
}

func TestVersionCommandUsesRuntimeBasePath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/info" {
			t.Fatalf("expected /api/info request, got %s", r.URL.Path)
		}

		_, _ = w.Write([]byte(`{"version":{"ferret":"remote-version"}}`))
	}))
	defer srv.Close()

	stdout, stderr, err := runCLI(t, "version", "--runtime", srv.URL+"/api")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Runtime: remote-version")
	assertEqual(t, stderr, "")
}

func TestRootWithoutArgsShowsHelp(t *testing.T) {
	stdout, stderr, err := runCLI(t)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "serve")
	assertContains(t, stdout, "version")
	assertEqual(t, stderr, "")
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
	assertContains(t, stdout, "serve")
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
	assertContains(t, stdout, "serve")
	assertContains(t, stdout, "version")
	assertContains(t, stdout, "--help, -h")
	assertContains(t, stdout, "OPTIONS:")
	assertNotContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRootHelpShowsCommandsOnly(t *testing.T) {
	stdout, stderr, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab [command] [command options]")
	assertContains(t, stdout, "run")
	assertContains(t, stdout, "serve")
	assertContains(t, stdout, "version")
	assertNotContains(t, stdout, "--files string")
	assertNotContains(t, stdout, "--timeout uint")
	assertEqual(t, stderr, "")
}

func TestRunHelpShowsExecutionFlags(t *testing.T) {
	stdout, stderr, err := runCLI(t, "run", "--help")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "lab run [options] [files...]")
	assertContains(t, stdout, "--files string")
	assertContains(t, stdout, "--serve string")
	assertContains(t, stdout, "--serve-bind string")
	assertContains(t, stdout, "--serve-host string")
	assertContains(t, stdout, "--timeout uint")
	assertContains(t, stdout, "--runtime string")
	assertNotContains(t, stdout, "--cdn")
	assertNotContains(t, stdout, "CDN")
	assertEqual(t, stderr, "")
}

func TestServeCommandWithoutEntriesShowsHelp(t *testing.T) {
	stdout, stderr, err := runCLI(t, "serve")
	helpStdout, helpStderr, helpErr := runCLI(t, "serve", "--help")

	assertExitCode(t, err, 1)

	if helpErr != nil {
		t.Fatalf("expected no error from help, got %v", helpErr)
	}

	assertContains(t, stdout, "lab serve [entries...]")
	assertContains(t, stdout, "--serve string")
	assertNotContains(t, stdout, "--cdn")
	assertNotContains(t, stderr, "No help topic for 'serve'")

	if stdout != helpStdout {
		t.Fatalf("expected bare serve help to match --help output, got %q and %q", stdout, helpStdout)
	}

	if stderr != helpStderr {
		t.Fatalf("expected bare serve stderr to match --help stderr, got %q and %q", stderr, helpStderr)
	}
}

func TestServeHelpUsesStaticServerTerminology(t *testing.T) {
	stdout, stderr, err := runCLI(t, "serve", "--help")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Serve one or more local directories over HTTP")
	assertContains(t, stdout, "lab serve [entries...]")
	assertContains(t, stdout, "--serve string")
	assertContains(t, stdout, "--serve-bind string")
	assertContains(t, stdout, "--serve-host string")
	assertNotContains(t, stdout, "CDN")
	assertNotContains(t, stdout, "--cdn")
	assertEqual(t, stderr, "")
}

func TestServeCommandServesPositionalEntries(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	mustMkdir(t, appDir)
	mustWriteFile(t, filepath.Join(appDir, "hello.txt"), "hello")

	stdout, stderr, done, cancel := startCLI(t, "serve", appDir)
	defer cancel()

	url := waitForServeURL(t, stdout, "app")
	assertHTTPBody(t, url+"/hello.txt", "hello")
	assertEqual(t, stderr.String(), "")

	cancel()

	if err := <-done; err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestServeCommandMergesPositionalAndFlaggedEntries(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "frontend")
	apiDir := filepath.Join(root, "mockdata")
	mustMkdir(t, appDir)
	mustMkdir(t, apiDir)
	mustWriteFile(t, filepath.Join(appDir, "app.txt"), "app")
	mustWriteFile(t, filepath.Join(apiDir, "api.txt"), "api")

	stdout, stderr, done, cancel := startCLI(t, "serve", appDir+"@app", "--serve", apiDir+"@api")
	defer cancel()

	appURL := waitForServeURL(t, stdout, "app")
	apiURL := waitForServeURL(t, stdout, "api")
	assertHTTPBody(t, appURL+"/app.txt", "app")
	assertHTTPBody(t, apiURL+"/api.txt", "api")
	assertEqual(t, stderr.String(), "")

	cancel()

	if err := <-done; err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestServeCommandSupportsAdvertisedHost(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	mustMkdir(t, appDir)

	stdout, stderr, done, cancel := startCLI(t, "serve", "--serve-bind", "0.0.0.0", "--serve-host", "example.test", appDir+"@app")
	defer cancel()

	waitForServeURLWithHost(t, stdout, "app", `example\.test`)
	assertEqual(t, stderr.String(), "")

	cancel()

	if err := <-done; err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunCommandWithServeFetchesStaticContent(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	mustMkdir(t, appDir)
	mustWriteFile(t, filepath.Join(appDir, "hello.txt"), "hello")

	script := writeNamedScript(t, "static.fql", `
LET content = TO_STRING(IO::NET::HTTP::GET(@lab.static.app + "/hello.txt"))
RETURN T::EQ(content, "hello")
`)

	stdout, stderr, err := runCLI(t, "run", "--serve", appDir+"@app", script)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertContains(t, stdout, "Passed")
	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRunCommandUsesExplicitConsoleReporter(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", "--reporter=console", script)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertContains(t, stdout, "Passed")
	assertContains(t, stdout, "Done")
	assertNotContains(t, stdout, "PASS file=")
	assertEqual(t, stderr, "")
}

func TestRunCommandUsesSimpleReporter(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", "--reporter=simple", script)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertContains(t, stdout, "PASS file=")
	assertContains(t, stdout, "attempts=1")
	assertContains(t, stdout, "times=1")
	assertContains(t, stdout, "DONE passed=1 failed=0 duration=")
	assertNotContains(t, stdout, "Passed")
	assertNotContains(t, stdout, "Done")
	assertNotContains(t, stdout, "INF")
	assertEqual(t, stderr, "")
}

func TestRunCommandRejectsUnknownReporter(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", "--reporter=bogus", script)

	assertExitCode(t, err, 1)
	assertErrorMessage(t, err, "unknown reporter: bogus")
	assertEqual(t, stdout, "")
	assertEqual(t, stderr, "")
}

func TestRunCommandAdvertisesConfiguredStaticHostToRemoteRuntime(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	mustMkdir(t, appDir)

	type remoteRunRequest struct {
		Params struct {
			Lab struct {
				Static map[string]string `json:"static"`
			} `json:"lab"`
		} `json:"params"`
	}

	requests := make(chan remoteRunRequest, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/" {
			t.Fatalf("expected / request, got %s", r.URL.Path)
		}

		var req remoteRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		requests <- req
		_, _ = w.Write([]byte("1"))
	}))
	defer srv.Close()

	script := writeScript(t)

	stdout, stderr, err := runCLI(
		t,
		"run",
		"--runtime", srv.URL,
		"--serve", appDir+"@app",
		"--serve-host", "example.test",
		script,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	select {
	case req := <-requests:
		addr := req.Params.Lab.Static["app"]
		assertMatches(t, addr, `^http://example\.test:\d+$`)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for remote runtime request")
	}

	assertContains(t, stdout, "Passed")
	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRunCommandUsesEmbeddedRuntimePath(t *testing.T) {
	type remoteRunRequest struct {
		Text string `json:"text"`
	}

	script := writeScript(t)
	requests := make(chan remoteRunRequest, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/api" {
			t.Fatalf("expected /api request, got %s", r.URL.Path)
		}

		var req remoteRunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		requests <- req
		_, _ = w.Write([]byte("1"))
	}))
	defer srv.Close()

	stdout, stderr, err := runCLI(t, "run", "--runtime", srv.URL+"/api", script)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	select {
	case req := <-requests:
		assertEqual(t, req.Text, "RETURN 1\n")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for remote runtime request")
	}

	assertContains(t, stdout, "Passed")
	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRunCommandUsesRuntimePathOverride(t *testing.T) {
	script := writeScript(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/v1/execute" {
			t.Fatalf("expected /v1/execute request, got %s", r.URL.Path)
		}

		_, _ = w.Write([]byte("1"))
	}))
	defer srv.Close()

	stdout, stderr, err := runCLI(t, "run", "--runtime", srv.URL, "--runtime-param=path:\"/v1/execute\"", script)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertContains(t, stdout, "Passed")
	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRunCommandSupportsServeEntriesFromEnv(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "app")
	mustMkdir(t, appDir)
	mustWriteFile(t, filepath.Join(appDir, "env.txt"), "env")

	script := writeNamedScript(t, "env.fql", `
LET content = TO_STRING(IO::NET::HTTP::GET(@lab.static.app + "/env.txt"))
RETURN T::EQ(content, "env")
`)

	stdout, stderr, err := runCLIWithEnv(t, map[string]string{
		"LAB_SERVE": appDir + "@app",
	}, "run", script)
	if err != nil {
		t.Fatalf("expected no error, got %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertContains(t, stdout, "Passed")
	assertContains(t, stdout, "Done")
	assertEqual(t, stderr, "")
}

func TestRunCommandWithoutServeDoesNotInitializeStaticServer(t *testing.T) {
	script := writeScript(t)

	stdout, stderr, err := runCLI(t, "run", script)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertContains(t, stdout, "Done")
	assertNotContains(t, stdout, "Serving ")
	assertEqual(t, stderr, "")
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
	assertEqual(t, stderr, "")
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

func runCLIWithEnv(t *testing.T, env map[string]string, args ...string) (string, string, error) {
	t.Helper()

	for key, value := range env {
		original, existed := os.LookupEnv(key)
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("failed to set %s: %v", key, err)
		}

		defer func(k, v string, ok bool) {
			if !ok {
				_ = os.Unsetenv(k)
				return
			}

			_ = os.Setenv(k, v)
		}(key, original, existed)
	}

	return runCLI(t, args...)
}

func startCLI(t *testing.T, args ...string) (*safeBuffer, *safeBuffer, <-chan error, context.CancelFunc) {
	t.Helper()

	stdout := &safeBuffer{}
	stderr := &safeBuffer{}

	ctx, cancel := context.WithCancel(context.Background())
	app := newApp("test-version", stdout, stderr)
	app.ExitErrHandler = func(_ context.Context, _ *cli.Command, _ error) {}

	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, append([]string{"lab"}, args...))
	}()

	return stdout, stderr, done, cancel
}

func writeScript(t *testing.T) string {
	t.Helper()
	return writeNamedScript(t, "test.fql", "RETURN 1")
}

func writeNamedScript(t *testing.T, name string, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o644); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	return path
}

func waitForServeURL(t *testing.T, stdout *safeBuffer, alias string) string {
	t.Helper()

	return waitForServeURLWithHost(t, stdout, alias, `127\.0\.0\.1`)
}

func waitForServeURLWithHost(t *testing.T, stdout *safeBuffer, alias string, hostPattern string) string {
	t.Helper()

	pattern := regexp.MustCompile(regexp.QuoteMeta(fmt.Sprintf("Serving %q at ", alias)) + `(http://` + hostPattern + `:\d+)`)
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		if matches := pattern.FindStringSubmatch(stdout.String()); len(matches) == 2 {
			return matches[1]
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for serve output for alias %q; stdout=%q", alias, stdout.String())
	return ""
}

func assertHTTPBody(t *testing.T, target string, expected string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.Get(target)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if readErr == nil && string(body) == expected {
				return
			}
		}

		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s to return %q", target, expected)
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

func assertContains(t *testing.T, output string, expected string) {
	t.Helper()

	if !strings.Contains(output, expected) {
		t.Fatalf("expected output to contain %q, got %q", expected, output)
	}
}

func assertMatches(t *testing.T, value string, pattern string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(value) {
		t.Fatalf("expected %q to match %q", value, pattern)
	}
}

func assertNotContains(t *testing.T, output string, unexpected string) {
	t.Helper()

	if strings.Contains(output, unexpected) {
		t.Fatalf("expected output not to contain %q, got %q", unexpected, output)
	}
}

func assertEqual(t *testing.T, actual string, expected string) {
	t.Helper()

	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
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
