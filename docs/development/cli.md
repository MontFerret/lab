# CLI

The CLI layer turns user input into package-owned options and coordinates the application lifecycle. It should remain thin: commands validate and assemble components, while execution semantics stay in their owning packages.

For current flags, defaults, and environment bindings, use the declarations in `cmd` as the source of truth.

## Application entrypoint

`main.go` constructs the `lab` application and registers `run`, `serve`, and `version`. It also:

- injects the Lab version at build time
- connects stdout and stderr to the command framework
- converts interrupt and termination signals into cancellation of the root context
- maps returned errors to process output and exit status

The root command shows application help. Script execution requires the `run` subcommand.

## Command ownership

### `run`

`lab run` coordinates a complete test run:

1. Resolve test locations from positional arguments or `--files`.
2. Wait for requested dependencies when `--wait` is configured.
3. Parse runtime adapter parameters and select a runtime.
4. Construct runner settings for concurrency, attempts, timeout, repetitions, and intervals.
5. Construct the aggregate source for the requested locations.
6. Parse user query parameters while keeping system parameters separate.
7. Parse and start optional static and mock services.
8. Publish service endpoints under `@lab.static` and `@lab.mock`.
9. Run the source through the runner and selected reporter.
10. Close the runtime and stop owned local services on every return path.

Runtime and local-service options are validated before execution proceeds. Cleanup uses bounded contexts for local servers. A failure returned by runtime cleanup is surfaced when no earlier run error already owns the result.

### `serve`

`lab serve` runs static and mock services without executing tests. Entries must be provided through `--static` or `--mock`; positional entries are rejected.

The command starts the configured managers, prints their advertised endpoints, waits for context cancellation, and then stops all started services with a bounded shutdown context. If mock startup fails after static startup, the static manager is stopped before the error is returned.

### `version`

`lab version` reports both the Lab version and the selected Ferret runtime version. Runtime selection uses the same adapter construction path as `run`, but version reporting must not execute a test. The adapter is closed before the command returns.

## Flags and environment values

Each CLI option that supports environment configuration declares its `LAB_*` binding beside the flag. When adding or changing a flag:

- keep the flag and environment binding in the same command declaration
- validate values before starting a runtime, process, network request, or local service
- convert CLI values into the owning package's options rather than passing the command object downstream
- update command help and user-visible error behavior together
- add top-level coverage when package-local tests do not prove the full command behavior

Positional `run` locations take precedence when present; otherwise `--files` supplies the locations. Missing locations produce command help and a failing exit status.

Runtime parameters and query parameters use separate flags and maps. The binary runtime's `flags` runtime parameter is extracted as adapter configuration rather than forwarded as an FQL query parameter.

Filesystem and outbound HTTP policy flags are parsed in `cmd`, converted to `pkg/runtime` policy types, and validated by the adapter layer. Their support differs by adapter; see [Runtime](runtime.md).

## Lifecycle and error handling

The command context is the common cancellation signal for waiting, source loading, execution, reporters, processes, HTTP requests, and server startup. Components that own resources expose cleanup through their package contracts.

Command errors should retain the context that helps a user identify the invalid flag, parameter, runtime, source, or service. Credentials, authorization values, and sensitive response bodies should not be included in diagnostics.

Help and usage errors are command behavior. Changes to their output or exit status require command-level tests.

## Where command changes belong

- application construction, signal wiring, and top-level registration: `main.go`
- command definitions, flags, environment sources, help, and CLI validation: `cmd`
- concurrency, retries, repetition, timeout, and streams: `pkg/runner`
- runtime selection and adapter-specific validation: `pkg/runtime`
- source parsing and loading: `pkg/sources`
- service entry parsing and lifecycle: `pkg/localserver`, `pkg/staticserver`, or `pkg/mockserver`
- output formatting and summary interpretation: `pkg/reporters`

Do not duplicate downstream behavior in the command layer to make a flag easier to implement.

## Testing CLI changes

Use package tests for conversion helpers and narrow flag validation. Add top-level application tests for behavior observable through command invocation, including:

- help and usage failures
- flag/environment precedence
- runtime and policy selection
- process exit behavior
- local-service startup and cleanup
- cancellation
- stdout and stderr routing

Tests should isolate external processes and networks where practical. Changes that cross into another subsystem should also use that subsystem's tests rather than relying only on command mocks.
