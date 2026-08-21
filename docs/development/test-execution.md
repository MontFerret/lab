# Test execution

Lab separates discovery, test semantics, scheduling, runtime execution, and reporting. This keeps source I/O independent from FQL execution and keeps presentation independent from result semantics.

```text
pkg/sources -> pkg/testing -> pkg/runner -> pkg/runtime -> pkg/reporters
```

Runtime adapter details are covered in [Runtime](runtime.md).

## Sources

`pkg/sources` turns configured locations into streams of `File` values and contextual source errors. A source discovers and returns test content; it does not execute tests.

Supported source forms include:

- local files, directories, and globs
- Git repositories and revisions
- HTTP and HTTPS URLs
- aggregates of multiple sources
- a no-op source used where an empty implementation is useful

Source identity travels with each file so later errors and results can identify the original path or URL. Aggregate sources preserve the identity of their children.

Filesystem traversal observes context cancellation and reports useful file context. Git and HTTP sources preserve repository or URL context without exposing credentials or unnecessary response data. Any temporary resource owned by a source must be released across success, error, timeout, and cancellation paths.

New source types implement the existing source contract and are registered in the location/scheme selection path. Source tests belong in `pkg/sources` and should cover selection, identity, errors, cancellation, and owned-resource cleanup.

## Test cases and suites

`pkg/testing` converts a source file into an executable Lab test case. Direct FQL files execute as units. YAML suite files define a query followed by either an assertion or a structured `expect.error` runtime-error expectation. Query and assertion scripts may be inline FQL or referenced scripts.

An empty `expect.error` object accepts any error returned by the runtime. Its optional `contains` field performs a substring match against the error message. Unknown fields inside `expect.error` fail during suite construction rather than degrading to an unqualified error expectation. Expected-error suites do not deserialize query output or resolve and run an assertion, and combining `assert` with `expect.error` is invalid.

The `.fail.fql` expected-failure convention remains supported for compatibility but is deprecated. Its execution semantics stay unchanged, and the test case exposes a deprecation warning that the runner carries once per file result for reporters to present.

Lab owns the test-language lifecycle around FQL; Ferret owns the meaning of the FQL itself. Changes to syntax, compilation, runtime values, or VM behavior belong in Ferret rather than `pkg/testing`.

Errors should identify the relevant file, suite, unit, query/assertion phase, or parameter whenever that context is available. Query and assertion behavior must remain predictable on success, assertion failure, runtime failure, timeout, retry, and cancellation.

Suite, unit, assertion, parameter, and helper behavior is tested in `pkg/testing`. Add command-level coverage only when the observable contract crosses the CLI/runtime boundary.

## Parameters

`pkg/testing.Params` keeps user values separate from Lab system values. At execution time, Lab materializes system values under the `lab` key, which exposes service endpoints such as `@lab.static.<alias>` and `@lab.mock.<alias>`.

Parameter isolation is a correctness boundary:

- user input must not overwrite the Lab namespace accidentally
- service managers populate system endpoint maps before test execution
- per-test and per-attempt state must not leak through shared mutable maps
- `Params.Clone` and related helpers must protect parallel tests, retries, and repeated runs from shared mutations

Parameter serialization for a particular runtime remains owned by `pkg/runtime`.

## Runner orchestration

`pkg/runner` coordinates source streams, test cases, runtimes, and result streams. It owns:

- bounded concurrency through the worker pool
- retry attempts for failed tests
- repeated successful executions
- intervals between retries or repeated runs
- per-test timeout configuration
- cancellation and worker shutdown
- progress results and final summary calculation

Runner settings normalize zero values to the established defaults. A source file receives a cloned parameter set before work is scheduled.

Each scheduled file becomes one progress result containing its identity, attempts, successful run count, duration, and final error. Source errors also become progress results so reporters can present them consistently. The summary counts passed and failed results and records total wall-clock duration.

Cancellation must stop new scheduling, release worker-pool capacity, interrupt supported runtime work, and allow output channels to close. Intervals use cancellable timers rather than uninterruptible sleeps.

Ordering is promised only where the implementation explicitly guarantees it. Parallel result order should not be stabilized accidentally by tests or presentation code.

## Reporters

`pkg/reporters` consumes the runner's progress and summary streams. Registered command reporters include interactive console output and simple plain-text output. The package also contains a silent reporter for internal composition.

Reporters may:

- format progress and errors
- present test-case deprecation warnings
- format the final summary
- translate a failed summary into a command error
- stop waiting when the context is canceled

Reporters do not control scheduling, retries, runtime behavior, cleanup, or result calculation. Important information cannot rely only on color or TTY styling.

Console output is human-facing. Any future machine-readable reporter needs a separately defined compatibility contract for field names, ordering, encoding, stdout purity, and failure behavior.

Reporter tests belong in `pkg/reporters` and should validate formatting, cancellation, summary interpretation, and errors without reproducing runner internals.

## Correctness and performance checks

Changes to sources or runner orchestration are both correctness- and performance-sensitive. Relevant tests should cover:

- success and contextual failures
- cancellation before and during work
- retry and repeated-run boundaries
- timeout behavior
- parameter isolation under concurrency
- stream closure and summary counts
- cleanup after partial progress

Benchmark source discovery/loading, scheduling, streams, aggregation, or reporter hot paths when a change could affect throughput or allocations. Capture comparable before/after measurements with `ns/op`, `B/op`, and `allocs/op` where available.
