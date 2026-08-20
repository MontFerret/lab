# Runtime adapters

`pkg/runtime` executes FQL through one of Lab's Ferret integration adapters. It owns adapter selection, adapter-specific configuration, filesystem and outbound HTTP policy conversion, version reporting, execution, cancellation, and cleanup.

Lab does not own FQL semantics. Adapters pass source identity, query content, and parameters to Ferret without rewriting the language.

## Common contract

Every runtime implements three operations:

- report its Ferret version without executing a test
- run a Ferret source with a parameter map and context
- close resources owned by the adapter

All adapters honor context cancellation where their integration permits it. Callers close the runtime after all runs finish, including error paths.

Runtime selection is centralized in `pkg/runtime`:

- HTTP and HTTPS URLs select the remote adapter.
- `bin:` URLs select the external Ferret CLI adapter.
- An empty value or a parsed URL with a scheme other than HTTP, HTTPS, or `bin` selects the built-in adapter. Invalid URL syntax fails during selection.

Adapter-specific settings are validated before execution or external resource startup. Options that do not apply to the selected adapter are rejected rather than silently ignored where the contract defines them as unsupported.

## Built-in runtime

The built-in adapter embeds Ferret using its supported Go APIs. It compiles and runs the supplied source with the configured parameters and policies and returns Ferret's output bytes.

The adapter:

- preserves source name and FQL content
- applies Lab-configured filesystem and outbound HTTP policy through Ferret APIs
- registers the Lab build's embedded Ferret version for cheap version reporting
- releases the embedded runtime when closed

Language behavior, compiler options, VM semantics, and runtime values remain Ferret responsibilities. A requested behavior that needs a Ferret change should be implemented upstream rather than emulated in this adapter.

## Remote HTTP runtime

The remote adapter uses an explicit HTTP contract:

- the configured base URL identifies the service
- version reporting sends `GET` to the service's information endpoint
- execution sends `POST` with JSON containing the FQL text and parameter map
- a configured runtime `path` overrides the run endpoint only
- runtime parameters may configure headers, cookies, and the run path
- only successful HTTP status codes are accepted

Requests are created with the caller's context. Errors retain request/response operation context without dumping sensitive headers, cookies, credentials, or response data by default.

Filesystem and outbound HTTP policies configure Ferret execution itself and therefore are not accepted by the remote adapter. Such policy must be enforced by the remote service under its own contract.

Remote contract changes should cover request method, resolved URL, headers, cookies, JSON encoding, response handling, errors, and cancellation in `pkg/runtime` tests.

## Binary runtime

The binary adapter runs a configured Ferret CLI v2 executable.

Version reporting invokes the executable's `version` command. Test execution invokes its `run` command, sends FQL content through stdin, and captures combined process output. The process is created with the caller's context so cancellation can terminate it.

Argument construction is part of the adapter contract:

1. Start with the `run` subcommand.
2. Append validated raw runtime flags.
3. Append Lab-managed filesystem and HTTP policy flags.
4. Append shared runtime parameters as deterministic `--param` arguments.
5. Append per-test query parameters using the same deterministic serialization.

Parameter keys are sorted before JSON serialization and argument construction. Raw flags are runtime configuration, not FQL query parameters. Raw flags that conflict with Lab-managed policy flags are rejected before the process starts.

Binary tests should cover exact argument order, deterministic parameter serialization, stdin, combined output, exit failures, invalid flags, policy conversion, version reporting, and cancellation. Benchmarks cover argument and invocation preparation where performance may change.

## Function-backed runtime

The function-backed adapter wraps a Go function in the common runtime interface. It is useful for composition and isolated tests. It has no owned resource to close and reports the embedded runtime version.

It is not a separate FQL implementation; the supplied function owns whatever test behavior it provides.

## Policy boundaries

Filesystem and outbound HTTP policy types live in `pkg/runtime` because they configure runtime execution. The CLI parses user-facing flags and converts them to these types, while adapters validate and apply them.

Policy changes should preserve:

- explicit support by adapter
- validation before execution
- deterministic CLI argument construction
- no conflict between raw and managed binary flags
- cancellation and error context
- Ferret ownership of the policy's execution semantics

Policy behavior is security-sensitive. Tests should include invalid combinations, explicit false/zero values, path and URL boundaries, and unsupported adapter cases.

## Performance

Runtime changes are significant when they can affect compilation/execution setup, HTTP request latency, process startup, argument serialization, allocations, or cleanup. Run the relevant existing benchmark before and after the change, or add one when the changed hot path is not covered.
