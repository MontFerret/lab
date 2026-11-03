# Runtime adapters

`pkg/runtime` executes FQL through one of Lab's Ferret integration adapters. It owns adapter selection, adapter-specific configuration, filesystem and outbound HTTP policy conversion, version reporting, execution, cancellation, and cleanup.

Lab does not own FQL semantics. Adapters pass source identity, query content, and parameters to Ferret without rewriting the language.

## Common contract

Lab uses the Universal API's `api.Runtime` directly in construction, commands, and runner options. Test cases use the execution-only `testing.Executor` contract. There is no Lab-owned runtime interface. Adapters implement `Version`, one-shot `Run`, `Compile`, `CompileDebug`, and `Close` with canonical sources, session options, and outputs.

The case executor calls one-shot `Run` with the source and materialized parameters by default. For repetitions or retries it lazily compiles each distinct source name/content pair and opens a fresh session per invocation. Unsupported compilation selects remembered one-shot fallback for that executor; syntax, transport, and cancellation errors propagate. An admission error returns no handle. Once admitted, the caller owns the `api.Output` and keeps the invocation context alive through consumption. Units, assertions, and expected-error queries discard chunks using `Consume`. Suite queries use `Collect`, check its terminal error, and decode completed `Content.Data` as JSON regardless of content type. Absent and empty content both map to a nil suite result. `Close` abandons unread output and releases a handle when cancellation prevents consumption from starting.

The runner owns one executor per case. The executor settles each output, closes its session, and retains successful plans until case completion. It closes plans in reverse acquisition order before the runner reports the result. Cleanup errors cannot satisfy expected failures and remain observable even after a later successful retry. The shared runtime is owned and closed by the command after cases finish. Direct Go callers likewise own their plans and sessions and close them before the runtime.

`New(ctx, opts)` uses the caller context to bound Wire transport establishment and handshake. It does not retain that context for execution.

All adapters honor context cancellation where their integration permits it. Callers close the runtime after all runs finish, including error paths.

Runtime selection is centralized in `pkg/runtime`:

- Omitted/empty targets and the existing `builtin` identifier select the built-in adapter.
- HTTP and HTTPS URLs select the remote adapter and preserve explicit URL paths.
- `wire+tcp://127.0.0.1:<port>` selects Wire, translated into the validated internal `tcp://` endpoint.
- Malformed targets, unknown schemes/selectors, bare Wire/TCP selectors, and `bin:` targets are errors.

`Options` contains `Target`, `AdapterOptions`, and the existing execution policies. `Options.Validate()` resolves configuration without I/O; `New` uses the same resolver before construction. Adapter option names, decoded types, and values are validated before external startup. Unsupported keys fail with option and adapter context without echoing sensitive values. The built-in runtime accepts no adapter options; HTTP accepts `headers`, `cookies`, and `path`; Wire accepts only `connect-timeout`. FQL parameters arrive exclusively through session options.

## Built-in runtime

The built-in adapter owns a Ferret UAPI runtime constructed with `uapi.New(api.Version(version), ...)`. It delegates all UAPI operations and transfers native output handles directly to callers, including responsibility for observing terminal and cleanup errors.

The adapter:

- preserves source name and FQL content
- applies Lab-configured filesystem and outbound HTTP policy through Ferret APIs
- receives FQL parameters through invocation/session options without a Lab runtime defaults layer
- registers the Lab build's embedded Ferret version for cheap version reporting
- releases the native runtime before its owned HTTP network, retaining the cleanup result for repeated and concurrent closes

Language behavior, compiler options, VM semantics, and runtime values remain Ferret responsibilities. A requested behavior that needs a Ferret change should be implemented upstream rather than emulated in this adapter.

## Remote HTTP runtime

The remote adapter uses an explicit HTTP contract:

- the configured base URL identifies the service
- version reporting sends `GET` to the service's information endpoint
- execution sends `POST` with JSON containing the FQL text and parameter map
- a configured runtime `path` overrides the run endpoint only
- adapter options configure headers and cookies (objects of strings), and the run path (a string URL reference)
- only successful HTTP status codes are accepted

Requests are created with the caller's context. Errors retain request/response operation context without dumping sensitive headers, cookies, credentials, or response data by default.

Filesystem and outbound HTTP policies configure Ferret execution itself and therefore are not accepted by the remote adapter. Such policy must be enforced by the remote service under its own contract.

HTTP session callbacks run once in order, merge parameters into a request-local map, and join their returned errors. `SetFSRoot`, `SetOutputContentType`, `Compile`, and `CompileDebug` return errors wrapping `errors.ErrUnsupported` without HTTP I/O. Already canceled contexts take precedence.

A successful run transfers its response body to an `api.Output` as soon as the headers arrive. Metadata carries the content type and the length when supplied. Consumption streams the bytes and closes the body. Collection returns detached bytes, including a prefix received before a failure. Cancellation and explicit close interrupt blocked reads. Status failures close the body before returning. Version reporting reads and closes its response eagerly, preserving the opaque `api.Version` and cleanup errors. The adapter borrows its HTTP client.

Remote contract changes should cover request method, resolved URL, headers, cookies, JSON encoding, response handling, errors, and cancellation in `pkg/runtime` tests.

## Wire runtime

Wire exposes the Universal Ferret API over gRPC. Lab selects it with `--runtime=wire+tcp://127.0.0.1:<port>`. Only IPv4 loopback TCP ports from 1 to 65535 are accepted; hostnames, other addresses, URL paths, credentials, queries, and fragments are rejected. This plaintext transport is intended for trusted local development, including hosts exposed through a local port mapping. The public URI scheme adds no encryption or authentication.

The `connect-timeout` adapter option is a positive duration string and defaults to `"5s"`. It bounds connection and handshake only; the existing test timeout remains separate. The caller context can impose an earlier cancellation or deadline. Lab closes the transport to unblock stalled stream creation as well as failed handshakes. Canceling the construction context after success does not cancel later executions.

Lab uses UAPI's one-shot `Runtime.Run` by default and compiled plans with fresh sessions for repetitions or retries. Source name, exact content, and invocation parameters are preserved. String-keyed nested YAML objects are converted to Wire's portable map shape; binary and numeric values retain their types. Inputs are not mutated. `headers`, `cookies`, `path`, and `flags` remain ordinary per-test FQL parameter names; adapter options are never injected into them.

The host must support chunked output framing as provided by Wire `v1.0.0-alpha.3`. Wire checks framing compatibility during Connect. The adapter forwards the remote one-shot output handle to the caller. Consumption observes terminal execution, delivery, and cleanup errors. A parameter-setter adapter preserves nested YAML conversion while forwarding filesystem and output settings. Non-nil callbacks run once in order and their errors are joined.

`Compile` and `CompileDebug` compile through the hosted runtime. A thin plan wrapper applies the same session-option normalization to `NewSession` and `NewDebugSession` as one-shot `Run`, preserving callbacks, overrides, validation, and input isolation. Plan metadata and closure delegate to the underlying plan. Version reporting returns the hosted `api.Runtime.Version(ctx)` exactly; Wire captures that value during Connect, so version reads do not require another RPC.

Lab owns the gRPC client transport it creates. Wire borrows it and owns the logical connection. After all runs settle, the adapter closes the logical runtime before the transport, retains both cleanup errors, and supports repeated/concurrent close calls. A lost physical connection is terminal for that adapter; it does not redial or select another runtime. The server owns its hosted runtime and configuration.

Filesystem and outbound HTTP execution policies and unsupported adapter options are rejected before connection. Configure Ferret execution policy at the remote host. HTTP/Worker support remains a separate adapter.

## Test support

Transport-independent controlled runtimes and option-capture fixtures live in `internal/testutil/apiruntime`. They implement UAPI directly, provide controlled plans and sessions, and reject unexpected compilation by default. `internal/testutil/wirehost` owns only the real loopback server and connection observations. Production adapters do not include a function-backed test shim.

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
