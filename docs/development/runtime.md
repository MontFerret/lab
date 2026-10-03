# Runtime adapters

`pkg/runtime` executes FQL through one of Lab's Ferret integration adapters. It owns adapter selection, adapter-specific configuration, filesystem and outbound HTTP policy conversion, version reporting, execution, cancellation, and cleanup.

Lab does not own FQL semantics. Adapters pass source identity, query content, and parameters to Ferret without rewriting the language.

## Common contract

Lab uses the Universal API's `api.Runtime` directly in construction, commands, runner options, and test cases. There is no Lab-owned runtime interface. Adapters implement `Version`, one-shot `Run`, `Compile`, `CompileDebug`, and `Close` with canonical sources, session options, and outputs.

The runner and test cases only call `Run`. Units pass an `api.Source` and `api.WithParams` for their materialized parameters. Suites execute the query and assertion separately, decode `Output.Content` as JSON, and treat nil or empty output as an absent result. They do not interpret `ContentType`. Query errors retain their existing handling even when output is available.

Callers using compilation directly own their plans and sessions and must close them before the runtime. Lab does not manage reusable plans or sessions.

`New(ctx, opts)` uses the caller context to bound Wire transport establishment and handshake. It does not retain that context for execution.

All adapters honor context cancellation where their integration permits it. Callers close the runtime after all runs finish, including error paths.

Runtime selection is centralized in `pkg/runtime`:

- HTTP and HTTPS URLs select the remote adapter.
- `bin:` URLs are rejected because external CLI binary execution is no longer supported.
- Explicit `wire` mode selects the Wire adapter with `--runtime-endpoint tcp://127.0.0.1:<port>`, using Ferret CLI's type normalization and endpoint grammar.
- After checking explicit Wire mode, an empty value or an unrelated unknown URL scheme selects the built-in adapter. Invalid URL syntax fails during selection.

Adapter-specific settings are validated before execution or external resource startup. Options that do not apply to the selected adapter are rejected rather than silently ignored where the contract defines them as unsupported.

## Built-in runtime

The built-in adapter owns a Ferret UAPI runtime constructed with `uapi.New(api.Version(version), ...)`. It delegates all UAPI operations and preserves output pointers and errors independently, including native cleanup failures.

The adapter:

- preserves source name and FQL content
- applies Lab-configured filesystem and outbound HTTP policy through Ferret APIs
- applies shared FQL parameters, including `flags`, before per-run overrides
- registers the Lab build's embedded Ferret version for cheap version reporting
- releases the native runtime before its owned HTTP network, retaining the cleanup result for repeated and concurrent closes

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

HTTP session callbacks run once in order, merge parameters into a request-local map, and join their returned errors. `SetFSRoot`, `SetOutputContentType`, `Compile`, and `CompileDebug` return errors wrapping `errors.ErrUnsupported` without HTTP I/O. Already canceled contexts take precedence.

Successful responses become `*api.Output` with exact bytes and the response's content type. Read failures retain any partial output, and body-close failures are joined with other errors. Version reporting preserves the opaque `api.Version` and response cleanup errors. Each response is closed within its operation; closing the adapter leaves the borrowed shared HTTP client usable.

Remote contract changes should cover request method, resolved URL, headers, cookies, JSON encoding, response handling, errors, and cancellation in `pkg/runtime` tests.

## Wire runtime

Wire exposes the Universal Ferret API over gRPC. Lab selects it with `--runtime=wire` and `--runtime-endpoint=tcp://127.0.0.1:<port>`, matching Ferret CLI. Only IPv4 loopback TCP ports from 1 to 65535 are accepted; hostnames, other addresses, URL paths, credentials, queries, and fragments are rejected. This plaintext transport is intended for trusted local development, including hosts exposed through a local port mapping.

`--runtime-connect-timeout` defaults to five seconds and must be positive. The caller context can impose an earlier cancellation or deadline. Lab closes the transport to unblock stalled stream creation as well as failed handshakes. Canceling the construction context after success does not cancel later executions.

Lab execution uses UAPI's one-shot `Runtime.Run`: source name and exact content are passed unchanged, shared runtime parameters are applied first, and per-run parameters override matching keys. String-keyed nested YAML objects are converted to Wire's portable map shape; binary and numeric values retain their types. Inputs are not mutated. Shared `headers`, `cookies`, `path`, and `flags` runtime parameters are rejected; those names remain available as ordinary per-test FQL parameters.

The adapter returns the remote output pointer and error unchanged. Lab consumes `Output.Content` without interpreting `ContentType`, retaining available output even when execution also returns an error. Shared parameters precede caller session callbacks; a parameter-setter adapter preserves nested YAML conversion while forwarding filesystem and output settings. Non-nil callbacks run once in order and their errors are joined.

`Compile` and `CompileDebug` forward directly to the hosted runtime. Shared Lab parameters are defaults for one-shot `Run`; compiled plans and their sessions retain the host's UAPI contract without Lab wrappers. Version reporting returns the hosted `api.Runtime.Version(ctx)` exactly; Wire captures that value during Connect, so version reads do not require another RPC.

Lab owns the gRPC client transport it creates. Wire borrows it and owns the logical connection. After all runs settle, the adapter closes the logical runtime before the transport, retains both cleanup errors, and supports repeated/concurrent close calls. A lost physical connection is terminal for that adapter; it does not redial or select another runtime. The server owns its hosted runtime and configuration.

Filesystem and outbound HTTP execution policies and reserved adapter configuration are rejected before connection. Configure Ferret execution policy at the remote host. HTTP/Worker support remains a separate adapter.

## Test support

Transport-independent controlled runtimes and option-capture fixtures live in `internal/testutil/apiruntime`. They implement UAPI directly and reject unexpected compilation by default. `internal/testutil/wirehost` owns only the real loopback server and connection observations. Production adapters do not include a function-backed test shim.

## Policy boundaries

Filesystem and outbound HTTP policy types live in `pkg/runtime` because they configure runtime execution. The CLI parses user-facing flags and converts them to these types, while adapters validate and apply them.

Policy changes should preserve:

- explicit support by adapter
- validation before execution
- cancellation and error context
- Ferret ownership of the policy's execution semantics

Policy behavior is security-sensitive. Tests should include invalid combinations, explicit false/zero values, path and URL boundaries, and unsupported adapter cases.

## Performance

Runtime changes are significant when they can affect compilation/execution setup, HTTP/gRPC request latency, Wire handshake/cleanup, parameter conversion, allocations, or cleanup. Run the relevant existing benchmark before and after the change, or add one when the changed hot path is not covered.
