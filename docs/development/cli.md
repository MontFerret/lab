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
2. Parse user query parameters and parameter bindings while keeping system parameters separate.
3. Parse the runtime target and adapter options and validate them without I/O.
4. Wait for requested dependencies when `--wait` is configured, then construct the runtime.
5. Construct runner settings for concurrency, attempts, timeout, repetitions, and intervals.
6. Construct the aggregate source for the requested locations.
7. Parse and start optional static and mock services.
8. Publish service endpoints under `@lab.static` and `@lab.mock`.
9. Resolve parameter bindings from one materialized snapshot into user parameter paths.
10. Run the source through the runner and selected reporter.
11. Close the runtime and stop owned local services on every return path.

Runtime and local-service options are validated before execution proceeds. Cleanup uses bounded contexts for local servers. A failure returned by runtime cleanup is surfaced when no earlier run error already owns the result.

### `serve`

`lab serve` runs static and mock services without executing tests. Entries must be provided through `--static` or `--mock`; positional entries are rejected.

The command starts the configured managers, prints their advertised endpoints, waits for context cancellation, and then stops all started services with a bounded shutdown context. If mock startup fails after static startup, the static manager is stopped before the error is returned.

### `version`

`lab version` reports both the Lab version and the selected Ferret runtime version. Runtime selection uses the same context-aware adapter construction path as `run`, but version reporting must not execute a test. The helper returns `api.Runtime`; the opaque `api.Version` is printed unchanged. The adapter is closed before the command returns.

## Flags and environment values

Each CLI option that supports environment configuration declares its `LAB_*` binding beside the flag. When adding or changing a flag:

- keep the flag and environment binding in the same command declaration
- validate values before starting a runtime, process, network request, or local service
- convert CLI values into the owning package's options rather than passing the command object downstream
- update command help and user-visible error behavior together
- add top-level coverage when package-local tests do not prove the full command behavior

Positional `run` locations take precedence when present; otherwise `--files` supplies the locations. Missing locations produce command help and a failing exit status.

Both `run` and `version` use shared runtime flag declarations and `runtimeOptionsFromCommand`. `--runtime` (`-r`, `LAB_RUNTIME`) is the complete target: omitted/empty or `builtin`, an HTTP/HTTPS URL, or `wire+tcp://127.0.0.1:<port>`. Unknown targets and schemes are errors. Wire retains its loopback-only endpoint restrictions and normalizes the public scheme into the internal TCP endpoint.

`--runtime-option` (`LAB_RUNTIME_OPTION`) contains adapter settings only. HTTP accepts `headers` and `cookies` objects of strings and a string `path`; Wire accepts a positive duration string `connect-timeout`, default `"5s"`. Built-in rejects adapter options. The name/value decoder is shared with query parameters, but a flag-local collector preserves JSON punctuation. Commas outside JSON separate entries. CLI occurrences replace the environment list; later duplicate keys replace whole earlier values. Diagnostics must not echo raw option values.

`runtime.Options.Validate` performs the same pure resolution as runtime construction. Run validates before dependency waiting, connects afterward, and preserves the existing per-test execution timeout. Version uses the same target and option handling without executing FQL. HTTP `path` affects execution only, never version URL derivation.

Commands construct `api.Runtime`; test cases supply FQL inputs through existing parameter facilities and execute through their case-scoped executor. It uses one-shot `Run` by default and reuses compiled plans for repetitions or retries when supported. Runtime options never supply shared FQL defaults. Names such as `headers`, `path`, and `flags` remain ordinary query inputs when supplied through `--param`. See the README migration table for the intentionally removed interface and defaults layer.

`--param-bind <target>=@<source>` adapts an existing parameter to an ordinary user parameter path. Binding declarations are validated before external setup where possible. Sources are resolved only after dynamic Lab values are available, and all sources use the same pre-binding snapshot so declaration order cannot create binding chains. Targets may be nested but cannot use the reserved `lab` namespace or overlap another binding or `--param` target.

Filesystem and outbound HTTP policy flags are parsed in `cmd`, converted to `pkg/runtime` policy types, and validated by the adapter layer. Their support differs by adapter; see [Runtime](runtime.md).

## Lifecycle and error handling

The command context is the common cancellation signal for waiting, source loading, execution, reporters, HTTP/Wire requests, and server startup. Components that own resources expose cleanup through their package contracts.

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
