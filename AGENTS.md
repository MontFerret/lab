# AGENTS.md

This file is the canonical operating guide for coding agents working in this repository. It applies to **Lab for Ferret v2** only. If repository documentation conflicts with this file, prefer `Makefile`, `go.mod`, and `.github/workflows/build.yml` for commands, toolchain, and CI behavior.

## Repo snapshot

- Module path: `github.com/MontFerret/lab/v2`
- Go version declared by `go.mod`: `1.25.6`
- CI uses Go `>=1.25`.
- This repository root is Lab for Ferret v2. Do not mix assumptions from the separate v1 branch.
- Lab is a Ferret-oriented test runner and local test-environment companion. It does not own FQL syntax, compiler behavior, VM behavior, or Ferret runtime semantics.
- High-level flow: `CLI` -> `sources` -> `testing` -> `runner` -> `runtime adapter` -> `Ferret` -> `reporter`

## Architectural mental model

Lab coordinates Ferret test execution. It discovers tests, prepares user and Lab-managed parameters, starts optional local services, runs tests through a selected Ferret runtime, and reports results.

Primary execution pipeline:

```text
locations -> pkg/sources -> pkg/testing -> pkg/runner -> pkg/runtime -> Ferret -> pkg/reporters
```

Local service pipeline:

```text
CLI flags/env -> pkg/localserver entries/settings -> staticserver/mockserver managers -> endpoints -> @lab params
```

Agents should reason about changes by pipeline stage and ownership boundary:

- Command flags, environment bindings, help, and command-level validation usually begin in `cmd`.
- Test suites, units, parameters, setup/query/assert/cleanup behavior, and Lab parameter materialization usually belong in `pkg/testing`.
- Concurrency, retries, repeated runs, timeouts, cancellation, result streams, and aggregation usually belong in `pkg/runner`.
- Ferret execution adapters belong in `pkg/runtime`; they should delegate Ferret semantics rather than reproduce them.
- Test discovery and loading belong in `pkg/sources`.
- User-facing progress, failures, and summaries belong in `pkg/reporters`.
- Shared local-server entry parsing, bind/advertise settings, endpoint formatting, aliases, and lifecycle coordination belong in `pkg/localserver`.
- Static file serving belongs in `pkg/staticserver`.
- OpenAPI-compatible mock APIs and `x-lab-mock` behavior belong in `pkg/mockserver`.
- Release and package assets belong at the repository root, in `.github/workflows`, `scripts`, `assets`, Dockerfiles, and `.goreleaser.yml`.

## Canonical invariants

- Lab coordinates execution; Ferret owns FQL syntax, compilation, VM execution, and runtime value semantics.
- The built-in runtime must use Ferret's embedding API rather than reimplement Ferret behavior.
- Remote and binary runtimes are integration adapters with explicit HTTP or CLI contracts.
- User parameters and Lab system parameters remain isolated until materialized for Ferret execution.
- Lab system parameters are exposed under `lab`, including `@lab.static` and `@lab.mock`.
- Local static and mock servers must expose deterministic endpoints once started and must be stopped on normal return, error, timeout, and cancellation paths.
- Mock route selection must be deterministic and independent of map iteration order wherever route specificity matters.
- Mock templates must not expose unsafe host capabilities such as environment lookup, filesystem access, DNS lookup, network access, or process execution.
- Reporters are observational; formatting must not change execution semantics.
- Do not infer current behavior from historical v1 code, stale comments, or old design notes.

## Package map

Agents should begin with the package that owns the requested behavior. Do not infer ownership from file names when an established package boundary already describes the responsibility.

### Commands and application entrypoints

* `main.go`
    * Owns top-level CLI construction, version injection, and command registration.
    * Keep it focused on application wiring and startup.
* `cmd`
    * Owns `run`, `serve`, `version`, default-command behavior, flags, `LAB_*` environment bindings, help, command validation, and conversion into package-owned options.
    * Keep command actions thin and delegate execution, source loading, runtime behavior, and local services to their owning packages.

### Test execution model

* `pkg/testing`
    * Owns Lab test cases, FQL units, YAML suites, parameters, setup/query/assert/cleanup behavior, and test-level validation.
    * Owns conversion of user and Lab system parameters into the map passed to Ferret.
    * Test behavior at the Lab test-language level whenever practical.
* `pkg/runner`
    * Owns orchestration, worker pools, concurrency, retries, repeated runs, timeouts, cancellation, result streams, and result/summary calculation.
    * Coordinates sources, test cases, runtimes, and reporters without taking ownership of their internals.
    * Treat scheduling, streams, allocation behavior, and timeout/cancellation changes as performance- and correctness-sensitive.
* `pkg/runtime`
    * Owns built-in, remote HTTP, external binary, and function-backed Ferret execution adapters.
    * The built-in adapter uses Ferret's embedding API.
    * Remote adapters preserve their HTTP request/response contract, configured path, headers, cookies, and error context.
    * Binary adapters preserve CLI argument ordering, parameter serialization, query input, raw flag boundaries, and process cancellation.

### Sources and reporting

* `pkg/sources`
    * Owns local filesystem, Git, HTTP, glob, aggregate, and no-op test sources.
    * Preserve source identity, path behavior, temporary-resource cleanup, and useful error context.
    * Source loading discovers and returns tests; it does not execute them.
* `pkg/reporters`
    * Owns console, simple, and silent output for progress, failures, and summaries.
    * Consumes runner streams/results without controlling retries, cancellation, runtimes, or result semantics.

### Local services

* `pkg/localserver`
    * Owns shared entry parsing, alias and port validation, bind host, advertised host, endpoint URLs, and manager/node lifecycle.
    * Shared local-service behavior belongs here rather than being duplicated by service-specific packages.
* `pkg/staticserver`
    * Owns serving local directories for `serve` and `run --serve` flows.
    * Reuses `pkg/localserver` for shared settings and lifecycle behavior.
* `pkg/mockserver`
    * Owns OpenAPI-compatible mock REST APIs, `x-lab-mock`, route construction and matching, response status/headers/body, templates, and request-derived template context.
    * Reuses `pkg/localserver` for shared service settings and lifecycle behavior.
    * Template behavior remains here unless another Lab-owned package genuinely needs the same contract.

### Release, packaging, and assets

* `assets`
    * Owns repository image assets used by documentation or packaging.
* `scripts`
    * Owns small, deterministic release and version helpers.
* `.github/workflows`
    * Owns CI and release workflows; `.github/workflows/build.yml` is the source of truth for CI validation.
* Dockerfiles and `.goreleaser.yml`
    * Own container and binary release packaging.

## Where to start by task

- Add or change a CLI flag or command:
    - inspect `cmd` first
    - check the corresponding `LAB_*` environment binding
    - convert command values into package-owned options
    - add top-level command coverage when the behavior is observable through the CLI

- Change a test suite, unit, parameter, or assertion:
    - inspect `pkg/testing`
    - identify the setup/query/assert/cleanup or parameter-materialization boundary
    - preserve the `@lab` parameter shape unless the task intentionally changes it
    - add package tests and command-level coverage when the behavior crosses the CLI/runtime boundary

- Change execution orchestration:
    - inspect `pkg/runner`
    - verify concurrency, retry, repeated-run, timeout, cancellation, stream, and summary interactions
    - add ordering tests only for ordering that the implementation promises

- Change Ferret runtime execution:
    - inspect `pkg/runtime`
    - verify embedding API usage for built-in changes
    - verify HTTP request/response behavior for remote changes
    - verify process args, query input, output, exit errors, raw flags, and cancellation for binary changes
    - do not change Ferret language behavior from Lab

- Change source loading:
    - inspect `pkg/sources`
    - identify whether the source is filesystem, Git, HTTP, glob, aggregate, or no-op
    - preserve source identity and cleanup behavior
    - isolate external network and filesystem behavior in tests where practical

- Change reporter output:
    - inspect `pkg/reporters`
    - preserve the runner/reporting separation
    - validate formatting, summary interpretation, and exit behavior without changing execution semantics

- Change static or mock service behavior:
    - inspect the service-specific package and `pkg/localserver`
    - keep shared entry, host, port, endpoint, and lifecycle behavior in `pkg/localserver`
    - validate startup, shutdown, cancellation, and advertised endpoint parameters
    - for mock APIs, also validate route matching, request context, status, headers, bodies, templates, and malformed specifications

- Change build, CI, release, or packaging:
    - inspect `Makefile`, the relevant workflow, `.goreleaser.yml`, Dockerfiles, and `scripts`
    - keep local commands and CI aligned
    - do not introduce a required developer tool without updating `make install-tools` or documenting the prerequisite

## Stability guide

Treat these as relatively stable unless a task explicitly targets them:

- CLI command names and broad purpose: `run`, `serve`, and `version`
- the execution flow through sources, testing, runner, runtime, and reporters
- the `@lab` system parameter namespace
- local static/mock entry syntax and endpoint alias behavior
- the built-in, remote HTTP, and binary runtime adapter split

Treat these as implementation-sensitive and verify current code before proposing changes:

- runner concurrency, retries, repeated runs, streams, timeouts, and cancellation
- parameter cloning and materialization
- remote runtime HTTP contracts
- binary runtime argument and parameter serialization
- source fetching, temporary files, and cleanup
- local-server startup, shutdown, and endpoint advertisement
- mock route specificity and template rendering
- reporter summary and exit behavior

Do not treat historical discussions, stale comments, or old branches as authoritative.

## Public API and package boundary rules

- Treat `pkg/testing`, `pkg/runner`, `pkg/runtime`, `pkg/sources`, `pkg/localserver`, `pkg/staticserver`, `pkg/mockserver`, and `pkg/reporters` as internal-to-Lab package boundaries even when symbols are exported.
- Do not export new symbols unless another package or test genuinely needs the contract.
- Prefer unexported helpers inside the owning package.
- If a new exported symbol is necessary, add a doc comment explaining its cross-package contract.
- Do not move behavior across packages only to simplify tests.
- Do not expose Ferret internals through Lab APIs unless explicitly requested.

## Ferret integration rules

- Lab depends on Ferret; Ferret does not depend on Lab.
- Do not implement or modify FQL syntax, semantic analysis, bytecode, VM behavior, or runtime value semantics in this repository.
- Built-in runtime changes should use stable Ferret embedding APIs.
- Remote runtime changes must keep Lab's HTTP contract explicit and must not assume an undocumented Ferret server implementation.
- Binary runtime changes must keep CLI arguments, parameters, query input, raw flags, and cancellation behavior explicit.
- If a change requires Ferret core work, call that dependency out rather than hiding Ferret behavior inside Lab.

## Test behavior rules

- Preserve existing test semantics unless the task explicitly changes them.
- Keep user and Lab system parameters isolated until materialization.
- `Params.Clone` and related helpers must avoid shared mutable state between tests, attempts, or repeated runs.
- Setup and cleanup behavior must remain predictable on success, failure, timeout, retry, and cancellation.
- Errors should identify the relevant file, suite, unit, phase, or parameter whenever that context exists.
- Reporter formatting must not influence test execution.

## Runtime adapter correctness rules

- All adapters must honor context cancellation where their integration permits it.
- The built-in adapter passes source identity and FQL content to Ferret without rewriting the query.
- The remote adapter preserves request encoding, headers, configured paths, cookies, response handling, and useful error context.
- The binary adapter serializes parameters deterministically and passes query content through the configured CLI contract.
- Raw binary flags are runtime configuration, not Ferret query parameters.
- Policy and adapter-specific options must be validated before starting execution or external resources.
- Runtime version reporting should remain cheap and must not execute tests.
- Adapters that own resources must expose and honor deterministic cleanup.

## Local service lifecycle rules

- Local services started by `lab run` or `lab serve` must stop before command completion.
- Cleanup paths must cover normal completion, startup failure, execution error, reporter error, timeout, and cancellation.
- Shutdown should use bounded contexts.
- A manager must remain safe to stop after partial startup.
- Dynamic ports must report the actual endpoint selected and avoid collisions as much as practical.
- Bind host and advertised host are separate concerns and must not be conflated.
- Endpoint aliases must be stable and suitable for `@lab.static.<alias>` and `@lab.mock.<alias>`.
- Shared entry parsing, validation, endpoint formatting, and lifecycle behavior belong in `pkg/localserver`.

## Mock API and template rules

- Lab-owned OpenAPI extensions use the `x-lab-*` namespace; the current mock response extension is `x-lab-mock`.
- Do not introduce `x-ferret-*` mock extensions.
- Mock specifications should remain OpenAPI-compatible YAML or JSON where practical.
- Malformed specifications and malformed mock extensions should fail early with useful operation/response context.
- Route matching must be deterministic; exact/static routes take precedence over parameterized routes when both match.
- Unsupported methods should return useful HTTP errors and accurate allowed-method information.
- Test response status, headers, and body together.
- Request-derived template context must remain explicit: method, path parameters, query parameters, headers, and parsed body.
- Mock templates use Go `text/template`; function registration belongs in `pkg/mockserver` unless intentionally shared.
- Sprig functions may be used only after removing unsafe functions, including environment, filesystem, DNS, network, and process access.
- Template evaluation must not mutate server state unless a stateful mock feature is explicitly designed.
- Validate templates while loading specifications when practical, and identify the operation/response in template errors.
- Random or fake-data helpers must be seedable or otherwise controllable for deterministic tests.

## Source loading rules

- Source loading discovers and returns test content; it must not execute tests.
- Sources must preserve useful file names, paths, URLs, and downstream identity.
- Filesystem sources must handle files, directories, and globs predictably.
- Git sources must clean up temporary clones and avoid leaking credentials in errors.
- HTTP sources must report URL and HTTP failure context without dumping sensitive response data by default.
- Aggregate sources must preserve the identity of their underlying source entries.
- Cleanup must run on success, failure, timeout, and cancellation wherever temporary resources are owned.

## Reporting rules

- Reporters consume streams and results; they do not drive execution.
- Reporters must not swallow cancellation or execution errors.
- Output should remain stable and useful for humans and CI logs.
- Machine-readable output, if introduced, requires a stricter compatibility contract than console output.
- Color or TTY-only styling must not be the sole carrier of important information.

## Go type and file structure rules

These rules are mandatory unless the task explicitly requires otherwise.

- Do not define multiple method-bearing structs in the same `.go` file.
- Declare a method-bearing struct as a standalone `type Name struct { ... }`.
- A method-bearing struct should live in a file named after its primary type or responsibility whenever practical, for example:
    - `runner.go` for `Runner`
    - `manager.go` for `Manager`
    - `server.go` for `Server`
- Grouped `type ( ... )` declarations are allowed for interfaces, passive data-only structs, and small related helper/value types from the same narrow concern.
- A grouped declaration may contain exactly one method-bearing struct only when it is the sole behavioral type and all other declarations are passive helpers from the same concern.
- Do not use grouped declarations to hide multiple behavioral types.
- If a helper struct gains methods and creates a second method-bearing type, extract it into its own file immediately.
- Keep a struct's methods in the same file as the struct unless there is a strong, explicit reason to split them.
- Do not add a method-bearing struct to an existing file merely because it compiles.

Allowed:

```go
type (
	Options struct {
		Runtime  runtime.Runtime
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

- one method-bearing type per file keeps behavioral ownership obvious
- standalone behavioral types make diffs and reviews clearer
- grouped declarations remain useful for passive, closely related types without obscuring ownership

## Function and method ownership rules

These rules are mandatory unless the task explicitly requires otherwise.

- A file centered on a method-bearing type contains that type, its methods, and constructors only.
- Do not mix non-constructor package-level helper functions into a type-centered file.
- If logic belongs to the primary type, implement it as a method.
- If logic does not belong to the type and must remain a package-level function, place it in a helper-focused file.
- Prefer package-level functions only when no natural owning type exists or the behavior is genuinely package-level.
- A file containing both methods and non-constructor package-level functions is normally a structure violation and should be refactored.

## Comment rules for functions and methods

- Do not add comments to every function or method by default.
- Exported functions and methods should usually have doc comments, especially when they form cross-package contracts.
- Comment unexported functions and methods only when they carry non-obvious semantics, invariants, side effects, ownership, cleanup, security, protocol, or lifecycle constraints.
- Comments explain intent, contracts, invariants, side effects, or lifecycle behavior rather than restating names and signatures.
- For runner, runtime, source, local-service, mock, and reporter internals, prefer semantic and ownership comments over implementation narration.
- Avoid comment wallpaper; use fewer, meaningful comments.

Preferred:

```go
// Stop shuts down every managed server using the supplied context.
// It remains safe after a partial Start failure.
func (m *Manager) Stop(ctx context.Context) error
```

Avoid:

```go
// Stop stops the manager.
func (m *Manager) Stop(ctx context.Context) error
```

## Go control-flow spacing rules

These rules are mandatory for handwritten Go code.

Blank lines should separate logical units and make control-flow boundaries visually obvious.

### Immediate producer + check

A declaration, assignment, function call, type assertion, lookup, parse operation, or similar statement may remain directly adjacent to a following `if` when the `if` immediately checks or consumes the value produced by that statement.

This includes error checks, boolean/result checks, type assertions, nil checks, bounds checks, and other immediate validation.

Preferred:

```go
res, err := doSome()
if err != nil {
	return err
}
```

Preferred:

```go
named, ok := typeOf.(*types.Named)
if !ok || named.Obj().Pkg() == nil || !w.localPackage(named.Obj().Pkg().Path()) {
	return w.source.errorAt(
		ErrorUnsupportedRegistration,
		expression.Pos(),
		"New selects a module root dynamically",
	)
}
```

Preferred:

```go
value := lookup(name)
if value == nil {
	return ErrNotFound
}
```

Preferred:

```go
count := len(items)
if count == 0 {
	return nil
}
```

The producer and its immediate check form one logical unit and should not be separated by a blank line.

### Separation from preceding logic

If an immediate producer + check unit follows another statement or logical unit, separate it from the preceding code with a blank line.

Preferred:

```go
prepareState()

named, ok := typeOf.(*types.Named)
if !ok {
	return ErrUnsupported
}
```

Avoid:

```go
prepareState()
named, ok := typeOf.(*types.Named)
if !ok {
	return ErrUnsupported
}
```

No leading blank line is required when the producer begins the enclosing block:

```go
func inspect(typeOf types.Type) error {
	named, ok := typeOf.(*types.Named)
	if !ok {
		return ErrUnsupported
	}

	return inspectNamed(named)
}
```

### Consecutive control-flow blocks

Separate independent `if` statements with a blank line.

Avoid:

```go
if foo != nil {
	useFoo(foo)
}
if bar != nil {
	useBar(bar)
}
```

Prefer:

```go
if foo != nil {
	useFoo(foo)
}

if bar != nil {
	useBar(bar)
}
```

This applies even when both conditions are short. Independent control-flow decisions should remain visually distinct.

### Statements after control flow

Add a blank line after a completed `if` block before continuing with a separate statement or logical unit.

Avoid:

```go
if foo == bar {
	doFoo()
}
doSomething()
```

Prefer:

```go
if foo == bar {
	doFoo()
}

doSomething()
```

## Response and code style

When assisting with this repository, keep responses practical, concise, and engineering-focused.

- Use short sections and clear headings.
- Use bullets for decisions, trade-offs, and follow-up work.
- Use code blocks only for code, commands, or configuration.
- Prefer focused snippets or diffs over full-file dumps.
- Explain why a change is needed before describing its implementation.
- Avoid repeating the same context.
- When multiple files change, summarize each file's responsibility before implementation details.

## Development practice expectations

Agents must follow repository-specific engineering discipline rather than generic style preferences.

### Core principles

- Preserve correctness first.
- Preserve subsystem boundaries and invariants.
- Prefer the smallest local change that fully solves the task.
- Introduce abstractions or refactors only when required for correctness, maintainability, or an explicitly requested design change.
- Do not optimize by intuition; measure performance-sensitive work.
- Keep behavioral ownership obvious in structure, naming, and file layout.

### Mandatory expectations

- Identify the owning subsystem before making a non-trivial change.
- Preserve existing behavior unless the task explicitly changes it.
- Add or update tests for every behavior change.
- Add or update benchmarks for significant changes.
- Run the narrowest relevant validation first, then broaden as appropriate.
- Report only tests, benchmarks, and validation that actually ran.
- Prefer current code, tests, and repository guidance over historical discussion or abandoned designs.
- Do not perform unrelated opportunistic refactors unless required for correctness.

### Required workflow for non-trivial changes

For every non-trivial change:

1. Identify the owning subsystem.
2. Identify the contract, invariant, or behavior being preserved or changed.
3. Choose the smallest implementation that fits the existing design.
4. Determine whether the change is significant.
5. Add or update correctness tests.
6. Capture a benchmark baseline and add or update benchmarks when the change is significant.
7. Run the relevant initial validation and summarize the evidence accurately.
8. Review the complete resulting diff using the mandatory final self-review requirements below.
9. Correct substantive findings and rerun all validation or benchmarks affected by those corrections.

### Significant changes

A change is significant when it could reasonably affect:

- test execution throughput
- runtime adapter latency
- local static/mock request latency or throughput
- mock route matching or template-rendering performance
- source discovery, fetching, or loading cost
- allocations, memory reuse, or cleanup behavior
- cancellation or timeout behavior
- worker-pool scheduling, result streams, or aggregation
- reporter performance on large result sets

This includes, but is not limited to, changes in:

- `pkg/runner`
- `pkg/runtime`
- `pkg/sources`
- `pkg/localserver`
- `pkg/staticserver`
- `pkg/mockserver`
- worker pools, retries, repeated runs, streams, source fetching, server lifecycle, route matching, template rendering, or result aggregation

This usually excludes comment-only, documentation-only, formatting-only, test-only, and pure rename changes without behavior or hot-path impact.

When uncertain, treat the change as significant and benchmark it.

### Benchmark workflow for significant changes

- Run the relevant benchmark before implementation and save the baseline.
- Run the same benchmark after implementation.
- Compare `ns/op`, `B/op`, and `allocs/op` where available.
- Report the exact benchmark command and summarize the delta.
- Add a benchmark when no relevant benchmark covers the changed hot path.
- If the environment cannot run benchmarks, state that limitation and do not claim benchmark validation.

### Mandatory final self-review

After completing the implementation and initial validation for any non-trivial task, agents must review the complete resulting diff before considering the task finished.

The purpose of this review is to catch problems in the implementation itself, not to generate additional work or redesign unrelated parts of the repository.

Review the final change for:

- Correctness
    - Verify that the implementation satisfies the task requirements completely.
    - Look for missing cases, incorrect assumptions, regressions, boundary conditions, and failure paths.
    - Check error handling, cancellation, cleanup, state transitions, ownership, and lifecycle behavior where applicable.
    - Verify that tests exercise the intended contract rather than merely mirroring the implementation.
- Code clarity and cleanliness
    - Look for unnecessary complexity, duplication, excessive nesting, awkward control flow, misleading naming, or code that is difficult to reason about.
    - Prefer straightforward and idiomatic Go over clever implementations.
    - Remove implementation artifacts that are no longer necessary after the final design has taken shape.
- Repository and Go best practices
    - Verify that the implementation follows the conventions and mandatory structure rules in this file.
    - Check relevant Go practices, error handling, resource ownership, concurrency behavior, and API design.
    - Do not introduce a pattern merely because it is generally fashionable; it must improve this repository specifically.
- Architecture
    - Verify that responsibilities remain in the correct package, type, and layer.
    - Check dependency direction and existing architectural boundaries.
    - Look for unwanted coupling, leaked implementation details, misplaced semantics, or abstractions at the wrong level.
    - Verify that shared semantics remain owned by the appropriate subsystem rather than being duplicated by consumers.
- Code organization and split
    - Verify that files, types, methods, functions, and packages have clear responsibilities.
    - Check compliance with the Go type/file and function/method ownership rules in this file.
    - Look for files or functions doing too much.
    - Also avoid unnecessary fragmentation where closely related logic has been split into excessive helpers or files.
    - Ensure that the primary execution path remains easy to follow.
- Tests
    - Look for meaningful behavioral gaps, especially negative cases and boundary conditions.
    - Check for brittle tests, redundant tests, tests coupled unnecessarily to implementation details, and assertions too weak to catch plausible regressions.
    - For bug fixes, verify that a test would fail without the fix whenever practical.
- Performance
    - For significant changes, inspect the final implementation for accidental allocations, repeated work, unnecessary materialization, or additional hot-path overhead.
    - Compare required benchmark results with the baseline.
    - Do not trade clear correctness for speculative micro-optimization.

When the review finds a problem:

1. Fix correctness issues and regressions.
2. Fix meaningful architectural, ownership, lifecycle, or maintainability problems.
3. Simplify unnecessarily complicated code when doing so clearly improves the implementation.
4. Add or improve tests when the review exposes a behavioral coverage gap.
5. Rerun validation affected by the review-driven changes.
6. Rerun relevant benchmarks if a review-driven change affects benchmarked code.

Do not use the self-review as justification for speculative refactoring, unrelated cleanup, API redesign, or stylistic churn.

Distinguish actual problems from optional preferences. Existing code that is already clear, correct, idiomatic, and appropriately structured should be left alone.

The first working implementation is not automatically the final implementation. The task is complete only after implementation, validation, self-review, necessary corrections, and final validation have been performed.

## Test placement rules

- CLI behavior observable through commands belongs in top-level command/application tests.
- Test model, suite, unit, parameter, and helper behavior belongs in `pkg/testing` tests.
- Retry, repeated-run, concurrency, timeout, cancellation, stream, and summary behavior belongs in `pkg/runner` tests.
- Built-in, remote, binary, and function-backed adapter behavior belongs in `pkg/runtime` tests, with external process/network behavior isolated where practical.
- Filesystem, Git, HTTP, glob, aggregate, cleanup, and error behavior belongs in `pkg/sources` tests.
- Reporter formatting and result interpretation belongs in `pkg/reporters` tests.
- Shared entry/settings/lifecycle behavior belongs in `pkg/localserver` tests.
- Static directory serving belongs in `pkg/staticserver` tests.
- Mock parsing, route matching, response rendering, request context, status, headers, and errors belong in `pkg/mockserver` tests.
- Add cross-package command coverage when package-local tests do not prove the complete user-visible flow.

## Validation and evidence

When finishing a non-trivial change, report:

- owning subsystem
- files changed
- tests added or updated
- benchmarks added or updated
- validation commands run
- benchmark commands and before/after results, when applicable
- notable invariants preserved or intentionally changed

For significant changes, tests alone are insufficient. Correctness tests and comparable before/after benchmark evidence are required when the environment permits them.

### Change discipline

- Adapt an existing local pattern before introducing a new architecture.
- Do not add wrappers, interfaces, helper layers, or indirection for aesthetics alone.
- Move behavior across packages only when the ownership boundary is genuinely wrong.
- Keep diffs focused on the requested task.
- If cleanup is necessary for a safe change, keep it tightly scoped and explain why.

### Comment and documentation discipline

- Document non-obvious semantics, invariants, side effects, ownership, lifecycle, security, or recovery behavior.
- Do not add comment wallpaper.
- Prefer why, contract, and invariant comments over implementation narration.
- Document cross-package behavior more carefully than obvious local helpers.

### Decision bias when uncertain

- preserve existing behavior
- prefer the smaller local change
- add a focused test
- treat potentially performance-sensitive changes as significant
- verify ownership before introducing a new abstraction or dependency

## Tooling prerequisites

- Go must be installed.
- `make` is optional but is the preferred entrypoint for repository workflows.
- `staticcheck`, `goimports`, and `revive` are required for lint/format flows; install them with `make install-tools`.
- Docker is required only for container validation.
- Release tooling is required only for release work governed by `.goreleaser.yml` and release scripts.

## Command matrix

- Full default validation: `make build`
- Broad tests: `go test ./...`
- Repository test target: `make test`
- Build the binary: `make compile`
- Vet: `make vet`
- Lint: `make lint`
- Format: `make fmt`
- Install lint/format tools: `make install-tools`

There is no repository-defined `make generate` target. Do not invent or run generation steps unless a task explicitly adds generated artifacts and updates the repository workflow.

## Editing rules

- Treat `Makefile` and `.github/workflows/build.yml` as the source of truth for validation commands.
- Prefer narrow validation first, then broaden:
    - package-local changes: run the affected package tests
    - command changes: run relevant top-level and command-focused tests
    - runner, runtime, source, or server changes: run affected package tests, then `go test ./...` when practical
    - cross-cutting changes: finish with `make build` when the toolchain is available
- Run `make fmt`, or equivalent `go fmt` plus `goimports`, after formatting-sensitive Go changes.
- Run `make lint` after lint-sensitive or exported-behavior changes when tooling is available.
- Validate CI, release, Docker, and install-script changes with the narrowest applicable workflow and state unvalidated portions.
- Do not edit vendored or generated dependency content.
- Do not claim validation that did not run.

## Secondary references

- `README.md` for product context, CLI examples, configuration, and user-facing behavior.
- `Makefile` for local workflow entrypoints.
- `.github/workflows/build.yml` for CI validation.
- `.github/workflows/release.yml`, `.goreleaser.yml`, Dockerfiles, and `scripts` for release and packaging behavior.
