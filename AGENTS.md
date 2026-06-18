# AGENTS.md

This file is the canonical operating guide for coding agents working in this repository. It is written for **Lab for Ferret v2**. If repository documentation conflicts with this file, prefer `Makefile`, `go.mod`, and `.github/workflows/build.yml` for commands, toolchain, and CI behavior.

## Repo snapshot

- Module path: `github.com/MontFerret/lab/v2`
- Go version: `1.25+`
- `go.mod` currently declares `go 1.25.6`.
- CI uses Go `>=1.25`.
- This repository root is Lab for Ferret v2. Do not mix assumptions from the separate v1 branch.
- Lab is a Ferret-oriented test runner and local test environment companion. It is not the owner of FQL syntax, compiler behavior, VM behavior, or Ferret runtime semantics.

High-level flow:

```text
CLI command -> sources -> testing case/suite/unit -> runner -> runtime adapter -> Ferret execution -> reporter
                                      |
                                      +-> local static/mock servers -> @lab.static / @lab.mock params
```

## Architectural mental model

Lab coordinates Ferret test execution. It discovers test files, prepares Lab-managed parameters and local services, runs tests through a selected Ferret runtime, and reports results.

Primary execution pipeline:

```text
locations -> pkg/sources -> pkg/testing -> pkg/runner -> pkg/runtime -> Ferret -> pkg/reporters
```

Local service pipeline:

```text
CLI flags/env -> pkg/localserver entries/settings -> staticserver/mockserver managers -> endpoints -> @lab params
```

Agents should reason about changes by ownership boundary:

- Command flags, environment variable bindings, command help, and command-level validation usually begin in `cmd`.
- Test suite/unit structure, YAML/FQL test interpretation, setup/query/assert/cleanup behavior, and Lab parameter materialization usually belong in `pkg/testing`.
- Execution orchestration, concurrency, retries, repeated runs, timeout handling, streams, and result aggregation usually belong in `pkg/runner`.
- Ferret execution adapters belong in `pkg/runtime`. Lab should call Ferret through public APIs, a remote HTTP runtime, or an external binary adapter; it should not duplicate Ferret language semantics.
- Source discovery and loading belong in `pkg/sources`.
- User-facing output belongs in `pkg/reporters`.
- Shared local server entry parsing, endpoint address handling, bind/advertise host behavior, lifecycle coordination, and alias behavior belong in `pkg/localserver`.
- Static file serving belongs in `pkg/staticserver`.
- OpenAPI-compatible mock REST APIs and `x-lab-mock` behavior belong in `pkg/mockserver`.
- Release/package assets belong at the repository root, `.goreleaser.yml`, Dockerfiles, scripts, and `assets`.

## Canonical invariants

- Lab does not own FQL syntax, semantic analysis, bytecode, VM execution, or runtime value semantics.
- Lab should preserve Ferret-facing behavior by delegating execution to Ferret runtime adapters.
- The built-in runtime should use the Ferret embedding API rather than reimplementing Ferret behavior.
- Remote and binary runtimes are integration adapters and should keep their wire/CLI contracts explicit.
- Test parameters are split into user parameters and Lab system parameters. System parameters are exposed under `lab`, for example `@lab.static` and `@lab.mock`.
- Local static and mock server endpoints must be deterministic from the user's perspective once started and must be cleaned up on normal return, error, timeout, and cancellation paths.
- Mock API route matching must be deterministic and independent of map iteration order where route specificity matters.
- Mock API templates must not expose unsafe host capabilities such as environment lookup, filesystem access, DNS lookup, network access, or process execution.
- Reporters should be observational: changing output formatting must not change test execution semantics.
- Do not treat historical v1 behavior as authoritative unless the task explicitly targets v1 compatibility.

## Package map

Agents should begin with the package whose responsibility owns the requested behavior. Do not infer ownership from file names alone when a package in this map already describes the intended boundary.

### Commands and application entrypoints

* `main.go`
    * Owns top-level CLI application wiring, version injection, and command registration.
    * Keep this file focused on application construction and startup.

* `cmd`
    * Owns CLI commands, flags, environment variable bindings, command help, command-level validation, and command-level option conversion.
    * Current commands include `run`, `serve`, default command behavior, and `version`.
    * Keep command actions thin. Delegate execution to `pkg/runner`, `pkg/runtime`, `pkg/sources`, `pkg/localserver`, `pkg/staticserver`, and `pkg/mockserver`.
    * Do not put test execution logic, runtime adapter behavior, or server internals directly in command files.

### Test execution model

* `pkg/testing`
    * Owns Lab test case modeling and interpretation.
    * Includes FQL unit tests, YAML suites, test params, setup/query/assert/cleanup structure, and test-level validation.
    * Owns how user parameters and Lab system parameters are converted into the map passed to Ferret.
    * Changes here should preserve the shape and meaning of `@lab` system parameters unless the task explicitly changes that contract.
    * Test behavior here at the Lab test-language level whenever practical.

* `pkg/runner`
    * Owns orchestration of test execution.
    * Includes runner options, runner context, worker pool behavior, concurrency, retries, repeated runs, timeout handling, result streams, and summary/result calculation.
    * Runner should coordinate sources, testing cases, runtimes, and reporters rather than own their internals.
    * Changes here may be performance-sensitive when they affect concurrency, scheduling, allocations, streams, or timeout/cancellation behavior.

* `pkg/runtime`
    * Owns Ferret execution adapters.
    * Supports built-in Ferret execution, remote HTTP runtime execution, external binary runtime execution, and function-backed runtime test adapters.
    * Built-in runtime should use the Ferret embedding API.
    * Remote runtime should preserve HTTP request/response contracts, headers, cookies, path handling, and error quality.
    * Binary runtime should preserve CLI argument serialization, stdin query passing, raw flag passthrough, and process cancellation behavior.
    * Do not duplicate Ferret FQL semantics here.

### Test sources

* `pkg/sources`
    * Owns discovering and loading test files from local filesystem paths, Git repositories, HTTP URLs, glob patterns, aggregate source lists, and no-op sources.
    * Source changes must preserve file names, path handling, cleanup, and useful error messages.
    * Git and HTTP source behavior may involve external resources; tests should isolate network/filesystem behavior where possible.
    * Do not mix source discovery with test execution semantics.

### Reporting

* `pkg/reporters`
    * Owns user-facing output formats for test progress, failures, and summaries.
    * Current reporters include console, simple, and silent styles.
    * Reporters should consume runner streams/results; they should not control execution, retries, cancellation, or runtime behavior.
    * Formatting changes should be validated with focused tests or snapshots when practical.

### Local server foundation

* `pkg/localserver`
    * Owns shared local server concepts used by static and mock servers.
    * Includes entry parsing, alias and port binding, bind host, advertised host, endpoint URL formatting, manager lifecycle, and node lifecycle.
    * Use this package when behavior is shared across local services.
    * Do not duplicate entry parsing, host/port validation, alias rules, or endpoint formatting in `pkg/staticserver` or `pkg/mockserver`.

* `pkg/staticserver`
    * Owns serving local static directories for the `serve` command and during `run --serve` execution.
    * Should reuse `pkg/localserver` for entry/settings/lifecycle behavior.
    * Changes must preserve directory validation, endpoint reporting, startup, shutdown, and cancellation behavior.

* `pkg/mockserver`
    * Owns OpenAPI-compatible mock REST API serving for `run --mock-api`.
    * Owns parsing mock specs, extracting `x-lab-mock`, route construction, route matching, status codes, headers, static bodies, templated bodies, JSON body handling, and mock server tests.
    * Should reuse `pkg/localserver` for entry/settings/lifecycle behavior where possible.
    * Template behavior belongs here unless it becomes shared by another Lab-owned package.

### Release, packaging, and assets

* `assets`
    * Owns repository image assets used by README or packaging.
    * Do not treat this as a frontend application unless an explicit UI is added.

* `scripts`
    * Owns release/version helper scripts.
    * Keep scripts small, deterministic, and consistent with `Makefile` and CI.

* `.github/workflows`
    * Owns CI and release workflow definitions.
    * Treat `.github/workflows/build.yml` as the source of truth for CI validation behavior.

* Dockerfiles and `.goreleaser.yml`
    * Own container and binary release packaging.
    * Changes here should be validated with the narrowest available build/release dry-run path when practical.

## Where to start by task

### Add or change a CLI flag or command

- Inspect `cmd` first.
- Check how the flag is bound to `LAB_*` environment variables.
- Convert flag values into package-owned options in command helpers.
- Keep command actions thin.
- Add command-level tests in `main_test.go` or focused package tests if command behavior is observable there.

### Change test suite, unit, params, or assertion behavior

- Inspect `pkg/testing` first.
- Identify whether the behavior belongs to FQL unit execution, YAML suite interpretation, params, setup/query/assert/cleanup, or helper conversion.
- Preserve the `@lab` parameter shape unless intentionally changing it.
- Add tests in `pkg/testing` and integration-style command tests when behavior crosses CLI/runtime boundaries.

### Change execution orchestration

- Inspect `pkg/runner` first.
- Verify interactions with context cancellation, timeouts, concurrency, retries, repeated runs, and result streams.
- Add tests for ordering guarantees only when the code promises ordering.
- Treat changes affecting worker pools, streams, or result aggregation as potentially significant.

### Change Ferret runtime execution

- Inspect `pkg/runtime` first.
- For built-in runtime changes, verify Ferret embedding API usage.
- For remote runtime changes, verify HTTP request/response behavior, endpoint path handling, headers, cookies, and error wrapping.
- For binary runtime changes, verify process args, stdin, output, exit errors, raw flags, and cancellation.
- Do not change Ferret language behavior from Lab.

### Change source loading

- Inspect `pkg/sources` first.
- Identify whether the source is filesystem, Git, HTTP, glob, aggregate, or no-op.
- Preserve cleanup behavior for temporary directories and fetched sources.
- Keep path handling and user-facing errors clear.
- Add tests for new source forms and failure modes.

### Change reporter output

- Inspect `pkg/reporters` first.
- Preserve runner/reporting separation.
- Validate output formatting and exit behavior.
- Do not make execution semantics depend on reporter output.

### Change static server behavior

- Inspect `pkg/staticserver` and shared behavior in `pkg/localserver`.
- Keep shared entry parsing and endpoint formatting in `pkg/localserver`.
- Validate server startup, shutdown, directory validation, and endpoint params.

### Change mock API behavior

- Inspect `pkg/mockserver` and shared behavior in `pkg/localserver`.
- Keep `x-lab-mock` as the Lab-owned OpenAPI extension namespace unless the task explicitly changes it.
- Validate route matching, path params, query/header/body template context, status codes, response headers, body rendering, and malformed spec errors.
- Add tests in `pkg/mockserver` for both static and dynamic responses.

### Change local server entry parsing or endpoint settings

- Inspect `pkg/localserver` first.
- Update `pkg/staticserver` and `pkg/mockserver` only where their service-specific behavior differs.
- Validate alias, port, bind host, advertised host, default port, dynamic port, and endpoint URL behavior.

### Change build, CI, release, or packaging

- Inspect `Makefile`, `.github/workflows/build.yml`, `.github/workflows/release.yml`, `.goreleaser.yml`, Dockerfiles, and `scripts`.
- Keep local commands and CI aligned.
- Do not add a new required developer tool without updating `make install-tools` or documenting the prerequisite.

## Stability guide

Treat these as relatively stable unless the task explicitly targets them:

- CLI command names and broad purpose: `run`, `serve`, `version`
- test execution flow through sources, testing, runner, runtime, and reporters
- Lab system parameter namespace under `@lab`
- local server entry syntax for static and mock services
- runtime adapter split: built-in, remote HTTP, binary

Treat these as implementation-sensitive and verify current code before proposing changes:

- runner concurrency, retries, repeated runs, stream behavior, and timeout handling
- parameter cloning/materialization
- remote runtime HTTP contracts
- binary runtime argument serialization
- source cleanup and temporary directory behavior
- local server lifecycle and cancellation
- mock API route matching and template rendering
- reporter exit behavior and summary calculation

Do not treat historical discussion, stale comments, or old branches as authoritative.

## Public API and package boundary rules

- Treat `pkg/testing`, `pkg/runner`, `pkg/runtime`, `pkg/sources`, `pkg/localserver`, `pkg/staticserver`, `pkg/mockserver`, and `pkg/reporters` as internal-to-Lab package boundaries even when exported symbols exist.
- Do not export new symbols unless another package or test genuinely needs the contract.
- Prefer unexported helpers inside the owning package before adding exported APIs.
- If a new exported symbol is necessary, add a doc comment that explains the external contract.
- Do not move behavior across packages only to make tests easier.
- Do not expose Ferret internals through Lab APIs unless explicitly requested.

## Ferret integration rules

- Lab depends on Ferret; Ferret does not depend on Lab.
- Do not change Ferret language semantics, compiler behavior, VM behavior, or runtime value semantics in this repository.
- Built-in runtime changes should use stable Ferret embedding APIs.
- Remote runtime changes should keep Lab's HTTP contract explicit and should not assume a specific Ferret server implementation beyond the documented endpoints.
- Binary runtime changes should preserve compatibility with Ferret CLI-style parameter passing unless the task explicitly changes that contract.
- Any change that requires a Ferret core change should be called out explicitly rather than hidden inside Lab code.

## Test behavior rules

- Preserve existing test semantics unless the task explicitly changes them.
- User parameters and Lab system parameters must remain isolated until materialized for Ferret execution.
- `Params.Clone` and related conversion helpers must avoid accidental shared mutable state between tests, attempts, or repeated runs.
- Setup and cleanup behavior must remain predictable under success, failure, timeout, retry, and cancellation.
- Errors should preserve enough context for users to identify the failing file, suite, unit, phase, or parameter.
- Do not make test behavior depend on reporter formatting.

## Runtime adapter correctness rules

- All runtime adapters must honor context cancellation where possible.
- Built-in runtime should pass source identity and content to Ferret without rewriting FQL.
- Remote runtime should preserve request body encoding, headers, configured path, cookies, and useful error wrapping.
- Binary runtime should serialize params predictably and pass query text through stdin.
- Raw binary flags are runtime configuration, not Ferret query params.
- Runtime adapter version reporting should remain cheap and should not run tests.

## Local server lifecycle rules

- Local servers started by `lab run` must be stopped before command completion.
- Stop paths must run on normal completion, runtime error, reporter error, timeout, and cancellation.
- Server shutdown should use bounded contexts.
- Dynamic port allocation must avoid collisions as much as practical and must report the actual endpoint used.
- Bind host and advertised host are separate concerns and should not be conflated.
- Endpoint aliases should be stable and suitable for use in `@lab.static.<alias>` and `@lab.mock.<alias>`.
- Shared entry parsing and endpoint formatting belong in `pkg/localserver`.

## Mock API rules

- Lab-owned OpenAPI extensions should use the `x-lab-*` namespace.
- The current mock response extension is `x-lab-mock`.
- Do not introduce `x-ferret-*` mock extensions.
- Mock specs should remain OpenAPI-compatible YAML or JSON where practical.
- Operation-level mock behavior should be explicit and local to the operation unless a shared/default mechanism is intentionally added.
- Route matching must be deterministic.
- Prefer exact/static routes over parameterized routes when both could match, if route specificity is implemented or changed.
- Unsupported methods should produce useful HTTP errors.
- Malformed specs or malformed mock extensions should fail early with useful errors.
- Response status, headers, body, and templates should be tested together.
- Keep request-derived template context explicit: method, path params, query params, headers, and parsed body.

## Mock template rules

- Mock response templates use Go `text/template`.
- Template function registration belongs in `pkg/mockserver` unless shared elsewhere intentionally.
- Sprig functions may be used, but unsafe functions must remain removed.
- At minimum, do not expose functions equivalent to `env`, `expandenv`, DNS lookup, filesystem access, network access, or process execution.
- Template evaluation must not mutate server state unless a stateful mock feature is explicitly designed.
- Validate templates early when loading specs when practical.
- Template errors should identify the response or operation that failed when practical.
- Random/fake data helpers, if added, should be seedable or controllable for deterministic tests.

## Source loading rules

- Source loading should not execute tests.
- Source packages should return file content and identity clearly.
- Filesystem source behavior must handle globs and directories predictably.
- Git source behavior must clean up temporary clones and should avoid leaking credentials in errors.
- HTTP source behavior must report URL and HTTP failure context without dumping sensitive response data by default.
- Aggregate sources should preserve useful source identity for downstream errors.

## Reporting rules

- Reporters should consume streams and results, not drive test execution.
- Reporters should not swallow cancellation or execution errors.
- Output should be stable enough for humans and CI logs.
- Machine-readable output, if added later, should have stricter compatibility expectations than console output.
- Do not add color-only or TTY-only information as the sole carrier of important status.

## Go type and file structure rules

These rules are mandatory unless the task explicitly requires otherwise.

- Do not define multiple substantial method-bearing structs in the same `.go` file.
- Prefer declaring a method-bearing struct as a standalone `type Name struct { ... }`.
- A method-bearing struct should usually live in its own file, named after the primary type or responsibility whenever practical, for example:
    - `runner.go` for `Runner`
    - `result.go` for `Result`
    - `manager.go` for `Manager`
    - `server.go` for `Server`
- Grouped `type ( ... )` declarations are allowed for interfaces, passive data-only structs, and small related helper/value types that belong to the same narrow concern.
- A grouped `type ( ... )` block may also contain exactly one method-bearing struct when:
    - it is the only behavioral type in the file, and
    - the other grouped types are passive helper/value types from the same narrow concern.
- Do not use grouped `type ( ... )` declarations to hide multiple substantial behavioral types.
- If a helper struct later gains methods and would create more than one substantial method-bearing struct in the file, extract it into its own file immediately.
- Methods for a struct should live in the same file as the struct unless there is a strong, explicit reason to split by concern.
- Do not place a new method-bearing struct into an existing file just because the code compiles.

Allowed:

```go
type (
	Options struct {
		Runtime runtime.Runtime
		PoolSize uint64
	}

	Stream interface {
		Next() (Result, bool)
	}
)
```

Avoid:

```go
type (
	Runner struct {
		// ...
	}

	Manager struct {
		// ...
	}
)
```

Rationale:

- one substantial method-bearing type per file keeps ownership of behavior obvious
- standalone method-bearing types make diffs and reviews clearer
- grouped type blocks are fine for passive, closely related types, but should not hide substantial behavioral types

## Function and method ownership rules

These rules are mandatory unless the task explicitly requires otherwise.

- A file centered on a method-bearing type should contain the type, its methods, and its constructors only.
- Do not mix unrelated package-level helper functions into a file that already contains methods for a primary type.
- In type-centered files, constructor functions are the only normally allowed package-level functions.
- If logic conceptually belongs to the primary type, implement it as a method.
- If logic does not belong to the type and must remain a package-level function, place it in a separate helper-focused file.
- Package-level functions are preferred only when there is no natural owning type or when the behavior is genuinely package-level.
- If a file contains both methods and non-constructor package-level functions, verify that the helper is tightly coupled to the file's primary responsibility; otherwise extract it.

## Comment rules for functions and methods

- Do not add comments to every function or method by default.
- Exported functions and methods should usually have doc comments, especially when they form a cross-package contract.
- Unexported functions and methods should be commented only when they carry non-obvious behavior, invariants, side effects, ownership rules, cleanup expectations, security constraints, or lifecycle constraints.
- Comments must explain intent, contract, invariants, side effects, or lifecycle behavior.
- Prefer comments that explain why the code exists, what must remain true, or how the method is meant to be used.
- Do not write comments that merely restate the method name or signature.
- For runner, runtime, source loading, local server, mock server, and reporter internals, prefer comments on semantics and invariants over implementation narration.
- Avoid comment wallpaper. Dense, meaningful comments are preferred over mechanically documenting obvious code.

Preferred:

```go
// Stop shuts down all managed servers using the supplied context.
// It is safe to call after a partial Start failure.
func (m *Manager) Stop(ctx context.Context) error
```

Avoid:

```go
// Stop stops the manager.
func (m *Manager) Stop(ctx context.Context) error
```

## Development practice expectations

Agents must follow repository-specific engineering discipline rather than generic style preferences.

### Core principles

- Preserve correctness first.
- Preserve subsystem boundaries and invariants.
- Prefer the smallest local change that fully solves the task.
- Avoid introducing abstractions, indirection, or refactors unless they are necessary for correctness, maintainability, or an explicitly requested design change.
- Do not optimize by intuition alone; use measurements for performance-sensitive work.
- Keep behavioral ownership obvious in code structure, naming, and file layout.

### Mandatory expectations

- Identify the owning subsystem before making a non-trivial change.
- Preserve existing behavior unless the task explicitly requires changing it.
- Add or update tests for any behavior change.
- Add or update benchmarks for any significant change.
- Run the narrowest relevant validation first, then broaden as appropriate.
- Do not claim tests, benchmarks, or validation were completed unless they were actually run.
- Do not treat historical discussions, abandoned directions, or old branches as authoritative over current code and repository guidance.
- Do not perform opportunistic refactors unrelated to the requested task unless they are required for correctness.

### Required workflow for non-trivial changes

Before making a non-trivial change, agents must:

1. Identify the owning subsystem.
2. Identify the contract, invariant, or behavior being preserved or changed.
3. Choose the smallest reasonable implementation that fits the existing design.
4. Determine whether the change is significant.
5. Add or update correctness tests.
6. Add or update benchmarks if the change is significant.
7. Run relevant validation and summarize the results accurately.

### Significant changes

A change is significant when it could reasonably affect:

- test execution throughput
- runtime adapter latency
- local server request latency or throughput
- mock API route matching performance
- source discovery or loading cost
- allocation patterns
- cancellation/timeout behavior
- worker pool scheduling
- result stream behavior
- reporter performance on large result sets

This includes, but is not limited to, changes in:

- `pkg/runner`
- `pkg/runtime`
- `pkg/sources`
- `pkg/localserver`
- `pkg/staticserver`
- `pkg/mockserver`
- worker pools, retries, repeated runs, streams, source fetching, server lifecycle, route matching, template rendering, or result aggregation

This usually does not include:

- comment-only, docs-only, or formatting-only edits
- pure renames with no behavior change
- test-only changes
- narrowly scoped refactors that do not affect behavior or hot paths

When in doubt, treat the change as significant and benchmark it.

### Benchmark workflow for significant changes

For significant changes, agents must:

- run relevant benchmarks before making the change and save the results as a baseline
- implement the change
- run the same benchmarks again after the change
- compare before/after results, preferably including `ns/op`, `B/op`, and `allocs/op`
- report the benchmark command used and summarize the performance delta

If no relevant benchmark exists for the changed hot path, add one.

If benchmark tooling or environment is unavailable, state that explicitly and do not claim benchmark validation was completed.

## Test placement rules

- CLI command behavior should have top-level command/application tests when behavior is observable through the CLI.
- `pkg/testing` behavior should have focused tests for suite/unit/params/helper behavior.
- Runner behavior should have `pkg/runner` tests for retries, repeated runs, concurrency, timeouts, cancellation, streams, and result summaries.
- Runtime adapter behavior should have `pkg/runtime` tests, with external process/network behavior isolated where practical.
- Source behavior should have `pkg/sources` tests for filesystem, Git, HTTP, glob, aggregate, cleanup, and error cases.
- Reporter behavior should have `pkg/reporters` tests when output or result interpretation changes.
- Local server entry/settings/lifecycle behavior should have `pkg/localserver` tests.
- Static server behavior should have `pkg/staticserver` tests.
- Mock API behavior should have `pkg/mockserver` tests for spec parsing, route matching, response rendering, template context, status, headers, and error cases.
- Cross-package CLI flows should be covered at the command/application level when package-local tests are not enough.

## Validation and evidence

When finishing a non-trivial change, agents should report:

- owning subsystem
- files changed
- tests added or updated
- benchmarks added or updated
- validation commands run
- benchmark commands run, if applicable
- notable invariants preserved or intentionally changed

For significant changes:

- tests alone are not sufficient
- both correctness tests and benchmarks are required
- benchmark results must be compared against a baseline when the environment allows it

### Change discipline

- Prefer adapting an existing local pattern over introducing a new architectural pattern.
- Do not add new helper layers, wrappers, interfaces, or abstractions only for aesthetic reasons.
- Do not move code across packages unless the ownership boundary is genuinely wrong.
- Keep diffs focused on the requested task.
- If cleanup is necessary to make the requested change safe, keep it tightly scoped and explain why it was needed.

### Comment and documentation discipline

- Add comments where semantics, invariants, side effects, ownership, lifecycle, security constraints, or recovery behavior are non-obvious.
- Do not add comment wallpaper.
- Prefer comments that explain why, contract, or invariants rather than implementation narration.
- Cross-package behavior should be documented more carefully than local obvious helpers.

### Decision bias when uncertain

When uncertain:

- preserve existing behavior
- prefer the smaller local change
- add a focused test
- treat the change as significant if performance might be affected
- verify ownership before introducing a new abstraction or package-level dependency

## Response and code style

When assisting with this repository, avoid large unstructured blocks of prose or code.

Prefer responses that are easy to scan:

- Use short sections with clear headings.
- Use bullet points for decisions, trade-offs, and follow-up work.
- Use code blocks only for actual code, commands, or configuration.
- Prefer focused snippets or diffs over full-file dumps.
- Explain why a change is needed before showing how to implement it.
- Keep comments in code useful and minimal.
- Avoid repeating the same context in multiple places.
- When the change touches multiple files, summarize the role of each file first.

The expected tone is practical, concise, and engineering-focused.

## Tooling prerequisites

- Go must be installed.
- `make` is optional but is the preferred entrypoint for repo-defined workflows.
- `staticcheck`, `goimports`, and `revive` are needed for lint/format flows; install them with `make install-tools`.
- Docker is needed only for Docker image validation or container-based manual checks.
- Release tooling is needed only for release tasks governed by `.goreleaser.yml` and release scripts.

## Command matrix

- Full default validation: `make build`
- Broad tests: `go test ./...`
- Repo test target: `make test`
- Build binary: `make compile`
- Vet: `make vet`
- Lint: `make lint`
- Format: `make fmt`
- Install lint/format tools: `make install-tools`

There is currently no repo-defined `make generate` target. Do not invent or run generation steps unless a task explicitly introduces generated artifacts and updates the repository workflow accordingly.

## Editing rules

- Treat `Makefile` and `.github/workflows/build.yml` as the source of truth for validation commands.
- Prefer narrow validation first, then broaden:
    - Package-local changes: run the affected `go test` package or packages.
    - Command-level changes: run relevant top-level tests and command-focused package tests.
    - Runner/runtime/source/server changes: run affected package tests, then `go test ./...` when practical.
    - Cross-cutting changes: finish with `make build` when the toolchain is available.
- If you changed formatting-sensitive Go files, run `make fmt` or the equivalent `go fmt` plus `goimports` command.
- If you changed lint-sensitive code paths or exported behavior, run `make lint` when the toolchain is available.
- If you changed CI, release, Docker, or install scripts, validate the narrowest applicable workflow and state any unvalidated parts clearly.
- Do not edit vendored or generated dependency content.

### Validation expectations

- After code changes, run the narrowest tests that prove the behavior you touched.
- Before finishing broader changes, run the relevant repo-level command from the matrix above.
- If validation cannot be run because tools are unavailable, state that explicitly.
- Do not claim validation was completed unless it actually was.

### Expectations for non-trivial changes

When proposing or implementing non-trivial changes:

- identify the owning subsystem first
- preserve invariants unless the task explicitly changes them
- prefer local, comprehensible changes before introducing new abstractions
- distinguish correctness work from performance work
- do not perform opportunistic refactors unrelated to the requested task unless they are necessary for correctness

## Secondary references

- `README.md` for product context, CLI examples, configuration reference, and user-facing behavior.
- `Makefile` for local workflow entrypoints.
- `.github/workflows/build.yml` for the current CI validation path.
- `.github/workflows/release.yml`, `.goreleaser.yml`, `Dockerfile`, `Dockerfile.release`, and `scripts` for release and packaging behavior.
