package main

import "testing"

func TestCommandsRejectRemovedBinaryRuntime(t *testing.T) {
	script := writeScript(t)

	for _, command := range []string{"run", "version"} {
		for _, selector := range []string{"bin:./ferret", "bin://ferret"} {
			t.Run(command+"/"+selector, func(t *testing.T) {
				args := []string{command, "--runtime=" + selector}
				if command == "run" {
					args = append(args, script)
				}

				stdout, stderr, err := runCLI(t, args...)
				assertErrorMessage(t, err, "binary runtimes are no longer supported; use the built-in, HTTP, or Wire runtime")
				assertEqual(t, stdout, "")
				assertEqual(t, stderr, "")
			})
		}

		t.Run(command+"/environment", func(t *testing.T) {
			args := []string{command}
			if command == "run" {
				args = append(args, script)
			}

			stdout, stderr, err := runCLIWithEnv(t, map[string]string{"LAB_RUNTIME": "bin:/missing/ferret"}, args...)
			assertErrorMessage(t, err, "binary runtimes are no longer supported; use the built-in, HTTP, or Wire runtime")
			assertEqual(t, stdout, "")
			assertEqual(t, stderr, "")

			args = append([]string{command, "--runtime="}, args[1:]...)

			stdout, stderr, err = runCLIWithEnv(t, map[string]string{"LAB_RUNTIME": "bin:/missing/ferret"}, args...)
			if err != nil {
				t.Fatalf("explicit built-in runtime must override the removed environment selector: %v", err)
			}

			assertEqual(t, stderr, "")
			if command == "run" {
				assertContains(t, stdout, "Passed")
			} else {
				assertContains(t, stdout, "Runtime:")
			}
		})
	}
}

func TestRunCommandKeepsFlagsAsFQLParameter(t *testing.T) {
	script := writeNamedScript(t, "flags.fql", `RETURN T::EQ(@flags, ["--browser-headless"])`)

	stdout, stderr, err := runCLI(t, "run", `--param=flags:["--browser-headless"]`, script)
	if err != nil {
		t.Fatalf("flags parameter: %v\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
	}

	assertContains(t, stdout, "Passed")
	assertEqual(t, stderr, "")
}

func TestRuntimeHelpListsSupportedAdapters(t *testing.T) {
	for _, command := range []string{"run", "version"} {
		t.Run(command, func(t *testing.T) {
			stdout, stderr, err := runCLI(t, command, "--help")
			if err != nil {
				t.Fatal(err)
			}

			assertContains(t, stdout, "builtin, HTTP/HTTPS URL, or wire+tcp://")
			assertNotContains(t, stdout, "bin:")
			assertNotContains(t, stdout, "log-output")
			assertNotContains(t, stdout, "binary runtime")
			assertEqual(t, stderr, "")
		})
	}
}
