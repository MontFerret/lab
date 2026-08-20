# AGENTS.md

This file is the canonical operating guide for coding agents working in this repository. It applies to **Lab for Ferret v2** only. Do not infer current behavior from the separate v1 branch, historical discussions, stale comments, or old design notes.

For implementation behavior, prefer the current code and tests. For toolchain and workflow details, use these repository sources of truth:

- `go.mod` for the Go version, module path, and dependencies
- `Makefile` for local build, test, formatting, lint, and release entrypoints
- `.github/workflows/build.yml` for CI validation
- `.github/workflows/release.yml`, `.github/workflows/update-ferret-core.yml`, `.goreleaser.yml`, Dockerfiles, and `scripts` for release, packaging, and dependency automation

Repository documentation explains the system but does not override these files.

## Purpose and universal boundaries

Lab is a Ferret-oriented test runner and local test-environment companion. Its primary execution flow is:

```text
CLI -> sources -> testing -> runner -> runtime adapter -> Ferret -> reporter
```

Keep these boundaries intact:

- Lab coordinates execution; Ferret owns FQL syntax, compilation, VM execution, and runtime value semantics.
- The built-in runtime uses Ferret's embedding API. Remote and binary runtimes remain explicit integration adapters.
- User parameters and Lab system parameters stay isolated until materialized for Ferret. Lab system values remain under `@lab`, including `@lab.static` and `@lab.mock`.
- Owned resources must be released on normal return, error, timeout, and cancellation. Cleanup must remain safe after partial startup.
- Sources preserve useful identity and error context. Reporters observe runner output and do not control execution semantics.
- Behavior that depends on ordering must be deterministic; do not rely on Go map iteration.

If a requested change requires Ferret core work, identify that dependency instead of reproducing Ferret behavior in Lab.

## Development documentation

Detailed subsystem documentation lives under `docs/development/`. Before making a substantial change, read only the guides relevant to the owning subsystem:

- [Architecture](docs/development/architecture.md): execution flows, dependency direction, package responsibilities, and stability boundaries
- [CLI](docs/development/cli.md): command wiring, flags and environment values, validation, startup, and cleanup
- [Test execution](docs/development/test-execution.md): sources, suites, parameters, runner orchestration, streams, and reporters
- [Runtime](docs/development/runtime.md): built-in, remote, binary, and function-backed adapter contracts and policies
- [Local services](docs/development/local-services.md): shared lifecycle, static serving, OpenAPI mocks, routing, and template safety
- [Release](docs/development/release.md): development commands, CI, versioning, packaging, releases, and dependency automation

Do not require every guide for every task. Use progressive disclosure based on ownership.

## Concise ownership map

- `main.go` wires the application, version, signals, and top-level commands.
- `cmd` owns `run`, `serve`, `version`, default-command behavior, flags, `LAB_*` environment bindings, command validation, and conversion into package-owned options.
- `pkg/sources` owns filesystem, Git, HTTP, aggregate, glob, and no-op source loading.
- `pkg/testing` owns direct FQL test cases, YAML query/assertion suites, parameters, and test-level validation.
- `pkg/runner` owns worker pools, concurrency, retries, repeated runs, timeouts, cancellation, result streams, and summaries.
- `pkg/runtime` owns Ferret execution adapters and their policy/configuration contracts.
- `pkg/reporters` owns console, simple, and silent result presentation.
- `pkg/localserver` owns shared entry parsing, host settings, endpoint formatting, and manager/node lifecycle.
- `pkg/staticserver` owns static directory serving.
- `pkg/mockserver` owns OpenAPI-compatible mock APIs, `x-lab-mock`, route matching, response rendering, and mock templates.
- Root workflows, scripts, assets, Dockerfiles, and `.goreleaser.yml` own build and release infrastructure.

Start with the package that owns the behavior. Do not move behavior across packages merely to simplify a test or avoid using an existing contract.

## Public API and compatibility rules

- Treat packages under `pkg` as internal-to-Lab boundaries even when their symbols are exported.
- Do not export a symbol unless another package or a test genuinely needs the contract. Prefer unexported helpers in the owning package.
- Add a doc comment when a new exported symbol is necessary and explain its cross-package contract.
- Do not expose Ferret internals through Lab APIs unless explicitly requested.
- Preserve existing behavior unless the task intentionally changes it.
- Treat CLI commands, flags, environment bindings, runtime HTTP/process contracts, parameter shapes, local-service entry syntax, endpoint aliases, and reporter output as compatibility-sensitive.
- Machine-readable output, if introduced, requires an explicit and stricter compatibility contract than human console output.
- Preserve cancellation, error context, source identity, deterministic serialization, and cleanup behavior at integration boundaries.

Verify implementation-sensitive behavior in current code and tests before changing it. This especially includes runner scheduling, parameter cloning, runtime adapters, source cleanup, local-server lifecycle, mock routing/templates, and reporter summaries.

## Go type and file structure rules

These rules are mandatory unless the task explicitly requires otherwise.

- Prefer grouped `type ( ... )` declarations for package-level types.
- Types declared in the same file should normally be placed in a single grouped
  `type` declaration rather than written as independent `type` declarations.
- This applies equally to structs, interfaces, aliases, named primitive types,
  and method-bearing types.
- Do not split types into independent declarations merely because one or more of
  them have methods.
- Keep related types together when they belong to the same narrow responsibility
  and their proximity makes ownership or lifecycle easier to understand.
- A file may contain multiple related behavioral types when they form one
  cohesive concern.
- Split types into separate files based on responsibility and ownership, not
  simply because multiple types have methods.
- When a file contains only one package-level type, a standalone declaration is
  acceptable; do not create an artificial group containing a single type.
- When adding a package-level type to a file that already contains type
  declarations, incorporate it into the existing type group when it belongs to
  the same concern.
- Avoid scattering a cohesive family of small types across multiple files.
- Do not create `helpers.go`, `utils.go`, or similarly generic files as dumping
  grounds. Organize files around predictable responsibilities.

Preferred:

```go
type (
	Options struct {
		Runtime  runtime.Runtime
		PoolSize uint64
	}

	Result struct {
		Name   string
		Status Status
	}

	Stream interface {
		Next() (Result, bool)
	}
)
```

Avoid independent declarations when the types belong to the same concern:

```go
type Options struct {
	Runtime  runtime.Runtime
	PoolSize uint64
}

type Result struct {
	Name   string
	Status Status
}

type Stream interface {
	Next() (Result, bool)
}
```

The grouped declaration expresses that these types form one related family.

## Function and method ownership rules

These rules are mandatory unless the task explicitly requires otherwise.

- Organize files around cohesive responsibilities rather than individual types.
- A file may contain multiple related types and their methods when they
  participate in the same narrow concern.
- Keep methods close to the types they belong to.
- A file containing methods must not also contain regular package-level
  functions unless those functions are constructors for types owned by that
  file.
- Constructors include conventional `New...` functions and other explicit
  construction functions whose primary responsibility is creating or
  initializing one of the file's types.
- Do not keep a regular helper function beside methods merely because those
  methods are its only callers.
- If behavior belongs to a type's state, invariants, lifecycle, synchronization,
  or owned resources, implement it as a method.
- If package-level behavior genuinely has no natural receiver, place it in a
  separate responsibility-focused file.
- Prefer package-level functions only when no natural owning type exists or the
  behavior is genuinely package-level.
- Split files when responsibilities diverge, not merely because several types
  have methods.
- Do not split cohesive behavior across files merely to enforce one type or one
  method-bearing type per file.

Preferred:

```go
type (
	Runner struct {
		poolSize uint64
	}

	Result struct {
		Status Status
	}
)

func NewRunner(poolSize uint64) *Runner {
	return &Runner{
		poolSize: poolSize,
	}
}

func (r *Runner) Run(ctx context.Context) (*Result, error) {
	// ...
}
```

Avoid mixing regular package-level functions with methods:

```go
func (r *Runner) Run(ctx context.Context) (*Result, error) {
	// ...
}

func normalizeResult(result *Result) {
	// ...
}

func (r *Runner) Close() error {
	// ...
}
```

If `normalizeResult` belongs to runner state or behavior, make it a method. If it
is genuinely package-level behavior, move it to an appropriately named
responsibility-focused file.

## Comment rules

- Do not add comments to every function or method by default.
- Exported functions and methods should usually have doc comments, especially when they form cross-package contracts.
- Comment unexported functions and methods only when they carry non-obvious semantics, invariants, side effects, ownership, cleanup, security, protocol, or lifecycle constraints.
- Explain intent, contracts, invariants, side effects, or lifecycle behavior rather than restating names and signatures.
- Prefer fewer semantic comments over implementation narration or comment wallpaper.

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

These rules are mandatory for handwritten Go code. Blank lines should separate logical units and make control-flow boundaries visually obvious.

### Immediate producer and check

A declaration, assignment, call, assertion, lookup, parse operation, or similar statement may remain directly adjacent to a following `if` when the `if` immediately checks or consumes the produced value. This includes error checks, boolean/result checks, type assertions, nil checks, and bounds checks.

Preferred:

```go
res, err := doSome()
if err != nil {
	return err
}
```

```go
named, ok := typeOf.(*types.Named)
if !ok || named.Obj().Pkg() == nil || !w.localPackage(named.Obj().Pkg().Path()) {
	return ErrUnsupported
}
```

```go
value := lookup(name)
if value == nil {
	return ErrNotFound
}
```

The producer and its immediate check form one logical unit and should not be separated by a blank line.

### Separation from preceding logic

If an immediate producer/check unit follows another statement or logical unit, separate it from the preceding code with a blank line.

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

No leading blank line is required when the producer begins the enclosing block.

### Independent control flow

Separate independent `if` statements with a blank line, even when both are short.

Avoid:

```go
if foo != nil {
	useFoo(foo)
}
if bar != nil {
	useBar(bar)
}
```

Preferred:

```go
if foo != nil {
	useFoo(foo)
}

if bar != nil {
	useBar(bar)
}
```

Add a blank line after a completed control-flow block before continuing with a separate statement or logical unit.

## Engineering workflow

For every non-trivial change:

1. Identify the owning subsystem and read its development guide.
2. Identify the contract, invariant, or behavior being preserved or changed.
3. Choose the smallest implementation that fits the existing design.
4. Determine whether the change is performance-significant.
5. Add or update correctness tests for every behavior change.
6. Capture a benchmark baseline and add or update benchmarks when the change is significant.
7. Run the narrowest relevant validation first, then broaden as appropriate.
8. Evaluate documentation impact and update affected repository and public documentation.
9. Review the complete resulting diff using the mandatory final self-review below.
10. Correct substantive findings and rerun affected validation and benchmarks.
11. Report implementation, documentation impact, evidence, review findings, and limitations accurately.

Apply these change-discipline rules throughout:

- Preserve correctness, subsystem boundaries, and established behavior unless the task explicitly changes them.
- Adapt an existing local pattern before introducing a new abstraction, dependency, wrapper, interface, or helper layer.
- Keep diffs focused. Do not perform unrelated opportunistic refactors or stylistic churn.
- Introduce cleanup or refactoring only when required for correctness, maintainability, or the requested design, and keep it tightly scoped.
- Do not optimize by intuition. Measure performance-sensitive work.
- Do not perform unrelated documentation rewrites; documentation updates required to keep affected contracts, behavior, examples, and guidance accurate are part of the task.
- Report only tests, benchmarks, and validation that actually ran.

## Testing expectations

Place tests with the subsystem that owns the behavior:

- top-level application tests for user-visible command behavior
- `pkg/testing` for suites, units, parameters, assertions, and helpers
- `pkg/runner` for retries, repeated runs, concurrency, timeouts, cancellation, streams, and summaries
- `pkg/runtime` for adapter request/process/embedding contracts and policy behavior
- `pkg/sources` for filesystem, Git, HTTP, glob, aggregate, identity, cleanup, and errors
- `pkg/reporters` for formatting and result interpretation
- `pkg/localserver`, `pkg/staticserver`, and `pkg/mockserver` for their shared lifecycle, static serving, mock parsing, routing, rendering, and error contracts

Add cross-package command coverage when package-local tests do not prove the complete user-visible flow. Add ordering tests only for ordering the implementation promises. For bug fixes, prefer a regression test that fails without the fix.

Tests should exercise contracts rather than mirror implementation details. Include meaningful negative cases, boundaries, timeout/cancellation paths, cleanup, and partial-start failure where relevant.

## Performance and benchmarks

A change is significant when it could reasonably affect:

- test execution throughput or worker-pool scheduling
- retries, repeated runs, streams, aggregation, timeout, or cancellation cost
- runtime adapter latency
- source discovery, fetching, or loading cost
- local static/mock request latency, route matching, or template rendering
- reporter behavior on large result sets
- allocations, memory reuse, or cleanup behavior

This commonly includes changes in `pkg/runner`, `pkg/runtime`, `pkg/sources`, `pkg/localserver`, `pkg/staticserver`, and `pkg/mockserver`. Documentation-only, formatting-only, test-only, and pure rename changes are normally not significant. When uncertain, treat the change as significant.

For significant changes:

- Run the relevant benchmark before implementation and save the baseline.
- Run the same benchmark afterward.
- Compare `ns/op`, `B/op`, and `allocs/op` where available.
- Add a benchmark when no existing benchmark covers the changed hot path.
- Report the exact command and before/after result.
- If the environment cannot run benchmarks, state the limitation and do not claim benchmark validation.

## Validation and evidence

- Use `Makefile` and `.github/workflows/build.yml` as the source of truth for commands. There is no repository-defined `make generate` target; do not invent a generation step.
- Run package-local validation first. Broaden to repository tests or the default build workflow when the change warrants it and the toolchain is available.
- Format Go changes with the repository workflow and run applicable vet/lint checks.
- Validate CI, release, Docker, installation, and documentation changes with the narrowest applicable checks, and state any unvalidated portions.
- Do not edit vendored or generated dependency content.
- Do not claim validation that did not run.

When finishing a non-trivial change, report:

- owning subsystem
- files changed
- tests added or updated
- benchmarks added or updated
- validation commands actually run
- benchmark commands and before/after results when applicable
- documentation updated, or documentation impact explicitly evaluated as none
- notable invariants preserved or intentionally changed

## Documentation synchronization

Documentation is part of the change, not a follow-up activity. Before completing
every non-trivial task, evaluate whether the implementation changes any
documented architecture, ownership boundary, invariant, lifecycle, workflow,
API, behavior, example, setup instruction, or contributor guidance.

Update the relevant documentation in the same task:

- Update `docs/development/*` when repository architecture, subsystem
  responsibilities, internal contracts, lifecycle behavior, development
  workflows, tooling, testing, benchmarking, local-service design, or release
  behavior changes.
- Update `README.md` or other repository-facing documentation when documented
  commands, configuration, setup, behavior, status, or examples change.
- Update the corresponding website documentation when changes affect publicly
  documented Lab behavior, Ferret integration, CLI usage, configuration,
  testing workflows, local services, runtime integration, or other user-facing
  functionality.
- Update both repository and website documentation when both internal
  contributor guidance and public behavior are affected.

Do not update documentation mechanically when its contract, behavior, examples,
or guidance are unaffected. Documentation-only churn is not a substitute for
evaluating documentation impact.

When the website repository or another required documentation source is
unavailable, identify the exact required follow-up explicitly in the final
report rather than silently leaving known documentation stale.

## Mandatory final self-review

After implementation and initial validation for any non-trivial task, review the complete resulting diff before considering the task finished. This is a second pass over the actual implementation, not a generic reminder.

Review the final change for:

- **Correctness:** confirm the requirements are complete; inspect boundaries, failure paths, error handling, cancellation, cleanup, state transitions, and lifecycle behavior.
- **Code clarity:** remove unnecessary complexity, duplication, excessive nesting, awkward control flow, misleading names, and obsolete implementation artifacts.
- **Repository and Go practices:** check the mandatory structure and spacing rules, error handling, concurrency, ownership, and API design.
- **Architecture:** verify responsibilities remain in the correct package and dependency direction; reject misplaced semantics, duplicated ownership, unwanted coupling, and leaked implementation details.
- **Organization:** ensure files, types, methods, and helpers have clear responsibilities without creating unnecessary fragmentation. Verify related package-level types use grouped declarations where appropriate and files containing methods do not mix in non-constructor package-level functions.
- **Tests:** look for missing negative and boundary cases, brittle or redundant tests, implementation-coupled assertions, and tests too weak to catch plausible regressions.
- **Performance:** for significant changes, inspect allocations and repeated work and compare the required benchmark evidence without sacrificing clarity or correctness for speculative optimization.
- **Documentation:** verify affected repository and public documentation has been updated, or any unavailable external documentation dependency is explicitly reported.

When the review finds an actual problem:

1. Fix correctness issues and regressions.
2. Fix meaningful architectural, ownership, lifecycle, or maintainability problems.
3. Simplify unnecessarily complicated code when that clearly improves the implementation.
4. Add or improve tests when the review exposes a behavioral coverage gap.
5. Update documentation when the review exposes stale contracts, examples, or guidance.
6. Rerun validation affected by the correction.
7. Rerun relevant benchmarks if the correction affects benchmarked code.

Do not use self-review to justify unrelated cleanup, speculative refactoring, API redesign, or stylistic churn. Distinguish actual problems from optional preferences. The first working implementation is not automatically the final implementation.

Immediately before finishing, inspect the complete final diff and verify that:

- every changed line belongs to the task or a necessary supporting change;
- unrelated user work remains untouched;
- no temporary or abandoned code remains;
- no accidental behavior, API, dependency, generated-file, or documentation change slipped in;
- tests and comments express current contracts;
- grouped type declarations and function/method organization follow the repository rules above;
- affected documentation is current;
- the final result is the smallest complete and coherent change.

If self-review causes an edit, rerun every validation or benchmark whose result may have been invalidated.

## Communication

Keep responses concise, practical, and engineering-focused. Explain why a change is needed before implementation detail, summarize each changed file's responsibility, and report evidence accurately without repeating context.