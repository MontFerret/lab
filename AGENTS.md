# AGENTS.md

This file is the canonical operating guide for coding agents working in this repository. It is written for Ferret Lab only. If repository documentation conflicts with this file, prefer `Makefile`, `go.mod`, and `.github/workflows/build.yml` for commands, toolchain, and CI behavior.

## Repo snapshot

- Repository: Ferret Lab.
- Purpose: local development and testing tools around Ferret/FQL, including runnable examples, browser-facing experiments, mock services, and developer workflows.
- Go version: verify `go.mod` before making toolchain assumptions.
- This repository is not the Ferret core language/runtime repository. Do not assume core Ferret package layout such as parser, compiler, bytecode, VM, stdlib, or runtime unless this repository explicitly vendors or wraps those APIs.
- Lab should depend on Ferret through public or intentionally exposed integration surfaces, not by recreating core language semantics locally.

## Architectural mental model

Ferret Lab is a development companion around Ferret.

Its job is to make Ferret easier to try, test, demonstrate, debug, and integrate. Agents should reason about changes by user-facing workflow and ownership boundary rather than by Ferret compiler/runtime internals.

Common Lab flow:

```text
configuration/spec/input -> Lab service/tooling -> Ferret execution or mock/runtime adapter -> observable result/log/response/UI
```

Agents should reason about changes by feature area:

- CLI/server behavior usually begins in the command or service package that owns process startup, flags, configuration loading, logging, and shutdown.
- API or mock-server behavior usually begins in the package that owns HTTP routing, request matching, response rendering, and mock-spec loading.
- OpenAPI-based behavior should preserve OpenAPI semantics and keep Lab-specific extensions clearly namespaced.
- Dynamic mock responses should be rendered through the chosen template layer, not by scattering ad hoc string replacement across handlers.
- Ferret execution behavior should use Ferret’s public embedding/API surface. Do not duplicate FQL semantics in Lab.
- Browser/UI behavior should preserve the contract between frontend state, API calls, execution results, diagnostics, and logs.
- File loading, examples, fixtures, and workspace behavior should be explicit about paths, sandboxing, and user-controlled input.
- Diagnostics and errors should remain user-facing and actionable. Lab is often the first surface a user sees when experimenting with Ferret.

## Canonical invariants

- Lab is an integration and developer-experience repository, not the owner of Ferret language semantics.
- Lab must not silently change, reinterpret, or emulate FQL behavior when Ferret itself should decide the result.
- User-provided scripts, specs, templates, fixtures, paths, request payloads, and environment values are untrusted input.
- Mock behavior must be deterministic by default unless randomness or dynamic data generation is explicitly requested by the spec.
- Lab-specific OpenAPI extensions must use the `x-lab-*` namespace. Do not introduce new `x-ferret-*` mock extensions.
- Mock API extensions should be additive and should not make the base OpenAPI document invalid for normal OpenAPI tooling.
- Dynamic mock responses should have clear precedence rules: explicit Lab extensions first, then OpenAPI examples/defaults/fallbacks as documented by the implementation.
- Long-running services must support graceful shutdown and must not leak listeners, browser processes, temporary files, or child processes.
- Do not assume behavior from old design notes, stale branches, or Ferret core internals unless reflected in the current Lab code.

## Package map

Agents should begin with the package whose responsibility owns the requested behavior. Do not infer ownership from file names alone when a package in this map already describes the intended boundary. If the current repository uses different package names, map these responsibilities to the closest existing package before changing code.

### Commands, server, and configuration

* CLI/command package
    * Owns command names, flags, environment variables, process startup, command help, and command-level validation.
    * Prefer this area when changing how users invoke Lab.
    * Keep command handlers thin; delegate behavior to service packages.
* Server/service package
    * Owns HTTP server startup, routing composition, middleware, logging, shutdown, and runtime service wiring.
    * Long-running behavior must respect context cancellation and graceful shutdown.
* Config package
    * Owns configuration loading, defaults, validation, and normalization.
    * Do not scatter config parsing across handlers, commands, or UI-specific code.
    * Configuration errors should identify the source field, file, or flag whenever practical.

### Ferret integration

* Ferret execution/integration package
    * Owns calls into Ferret public APIs, execution setup, parameters, module wiring, result handling, and diagnostic conversion.
    * Do not duplicate Ferret parser/compiler/runtime behavior in Lab.
    * Keep Lab-specific execution concerns, such as examples, UI sessions, or request-scoped options, outside Ferret core assumptions.
* Diagnostics/result package
    * Owns Lab-facing diagnostic formatting, result envelopes, logs, and API response shapes for execution results.
    * Preserve source spans, labels, notes, and hints when Ferret provides them.
    * Do not replace specific Ferret diagnostics with generic Lab errors.

### Mock API support

* Mock server package
    * Owns fake REST API server behavior: route registration, request matching, response selection, response rendering, latency/error simulation, and request logs.
    * Mock behavior must be safe with untrusted requests and specs.
    * Keep matching rules deterministic and easy to explain.
* Mock spec/OpenAPI package
    * Owns parsing, validating, and normalizing OpenAPI documents plus Lab-specific extensions.
    * Lab-specific extensions must use `x-lab-*` names.
    * OpenAPI compatibility should be preserved whenever possible.
    * Extension parsing should be strict enough to catch mistakes, but errors should be helpful.
* Template/rendering package
    * Owns dynamic response rendering, template function registration, escaping policy, and render context shape.
    * Prefer Go templates for dynamic mock behavior when that is the repository direction.
    * Keep template function registration centralized.
    * Do not allow templates to access arbitrary host resources unless explicitly designed and reviewed.
* Random/fake-data helpers
    * Own deterministic random data generation support for mocks, examples, and tests.
    * Support seeding when reproducibility matters.
    * Do not make tests depend on non-deterministic random output.

### Browser, UI, and assets

* Frontend/UI package or directory
    * Owns user-facing Lab screens, editor state, request/response views, logs, and browser interactions.
    * Preserve API contracts between UI and backend.
    * Avoid embedding backend-only assumptions in UI state names or response parsing.
* Static asset/build package or directory
    * Owns asset generation, bundling, embedding, and cache/version behavior.
    * Generated assets should be treated as derived output and regenerated from their source inputs.

### Examples, fixtures, and file access

* Examples/fixtures package or directory
    * Owns runnable examples, sample specs, mock data, test fixtures, and documentation-adjacent demos.
    * Examples should be small, realistic, and easy to run.
    * Keep fixture behavior deterministic.
* File/workspace package
    * Owns controlled file access, workspace paths, temporary directories, and sandbox-like behavior.
    * Do not let request handlers, templates, or examples read arbitrary paths without going through the owning abstraction.

### Internals and support packages

* Logging package
    * Owns structured logging and log formatting.
    * Logging should remain observational.
    * Do not make behavior depend on log output.
* Internal support packages
    * Own implementation-only helpers that should not become public contracts.
    * Do not move code into an internal helper only to avoid thinking about ownership.

## Where to start by task

- Add or change a CLI command or flag:
    - inspect the command package and command help tests
    - inspect config defaults and validation
    - keep command handlers thin
    - add command-level tests where practical

- Change server startup, routing, or shutdown:
    - inspect server/service wiring
    - verify context cancellation and graceful shutdown
    - add tests for routing/middleware behavior when practical

- Add or change Ferret execution behavior:
    - inspect the Ferret integration package
    - use public Ferret APIs
    - preserve diagnostics and result shape
    - add integration coverage using real Ferret execution where practical

- Add or change mock REST API support:
    - inspect mock server routing and request matching
    - inspect mock spec/OpenAPI normalization
    - define precedence between `x-lab-*` extensions, OpenAPI examples, defaults, and fallbacks
    - add tests for route matching, response selection, rendering, and errors

- Add or change OpenAPI extensions:
    - use the `x-lab-*` namespace
    - update extension parsing and validation
    - add valid and invalid spec fixtures
    - document extension shape in examples or docs-adjacent fixtures

- Add or change dynamic mock responses:
    - inspect template/rendering ownership
    - centralize template functions
    - define render context explicitly
    - add deterministic tests for rendered output

- Add random or fake-data generation:
    - inspect random/fake-data helper ownership
    - support deterministic seeding for tests
    - avoid non-deterministic assertions

- Change UI/API contracts:
    - inspect both frontend usage and backend response definitions
    - update tests or fixtures on both sides
    - preserve backward compatibility unless the task explicitly changes it

- Change examples or fixtures:
    - keep examples focused and runnable
    - avoid large fixtures unless they are testing a specific edge case
    - update any docs or README snippets that reference the changed example

- Change file or workspace behavior:
    - inspect the owning file/workspace abstraction
    - validate path handling, sandbox behavior, and cleanup
    - add tests for unsafe or edge-case paths

## Stability guide

Treat these as relatively stable unless the task explicitly targets them:

- Lab’s role as a companion/developer-experience tool around Ferret
- use of Ferret public APIs for execution behavior
- command and server entrypoints used by users or scripts
- documented mock-spec extension names
- public API response shapes consumed by the UI

Treat these as implementation-sensitive and verify current code before proposing changes:

- server lifecycle and shutdown paths
- request routing and mock matching precedence
- template rendering and escaping behavior
- file/workspace path handling
- browser process or external process lifecycle
- Ferret execution result and diagnostic conversion
- generated frontend/static assets

Do not treat historical discussion, stale comments, or old branches as authoritative.

## Public API and package boundary rules

- Treat command names, flags, config fields, HTTP API routes, mock extension names, and UI-consumed response shapes as API-sensitive.
- Do not rename or remove public-facing fields unless the task explicitly requires a breaking change.
- Prefer unexported helpers inside the owning package before adding exported APIs.
- If a new exported symbol is necessary, add a doc comment that explains the external contract and stability expectation.
- Do not create broad helper packages only to share a few lines of code.
- Do not expose internal mock/template/file-system machinery through public APIs unless explicitly requested.

## Ferret integration rules

- Lab must not duplicate Ferret language semantics.
- Use Ferret public APIs for parsing, compiling, running, diagnostics, modules, and result handling.
- Any intentional difference between Lab behavior and Ferret CLI/core behavior must be called out explicitly in the final summary.
- When behavior differs from old docs or memory, prefer current Lab code and current Ferret public APIs.
- Preserve existing Lab behavior unless the task explicitly requires changing it.

## Mock API rules

- Lab-specific OpenAPI extensions must be named with `x-lab-*`.
- Do not introduce or preserve `x-ferret-mock` as the mock-extension namespace unless the task is explicitly about backward compatibility or migration.
- If compatibility with an old extension is required, support it through a documented migration path and prefer `x-lab-mock` for new specs.
- Keep request matching deterministic.
- Make response precedence explicit in code and tests.
- Dynamic response templates must have a documented render context.
- Template functions must be registered centrally.
- Random data generation must be seedable or otherwise testable.
- Do not let mock templates access the local filesystem, network, environment, or process state unless explicitly designed and reviewed.
- Do not let mock responses accidentally expose secrets from the host environment.

## Resource and lifecycle rules

- Long-running services must honor context cancellation.
- HTTP servers must shut down gracefully.
- Browser processes, child processes, listeners, temporary files, and open response bodies must be cleaned up.
- Values or results that own resources must document ownership and cleanup behavior.
- Cleanup must be deterministic where the API exposes `Close`, `Shutdown`, `Stop`, or equivalent lifecycle methods.
- Tests that start servers or create temporary files must clean them up.

## Diagnostic and error quality rules

- User-facing errors should identify the failing command, config field, spec path, route, template, request, or file whenever practical.
- Prefer actionable hints when an error is likely caused by a common misuse.
- Do not replace specific Ferret diagnostics with generic Lab errors.
- OpenAPI/spec validation errors should include useful path information.
- Template errors should identify the template field or response being rendered.
- Tests for diagnostics should verify message quality when behavior changes.

## Security and untrusted input rules

- Treat scripts, OpenAPI specs, templates, request bodies, headers, paths, query strings, examples, fixtures, and environment-derived values as untrusted input.
- Do not evaluate templates with access to arbitrary host capabilities.
- Do not read arbitrary files from request-controlled paths.
- Do not write outside controlled workspaces or temporary directories.
- Avoid leaking environment variables, filesystem paths, tokens, cookies, or host metadata in responses or logs.
- Keep CORS, host binding, and local-network exposure intentional and explicit.
- Prefer localhost-only defaults for development servers unless the task explicitly requires broader exposure.

## Go type and file structure rules

These rules are mandatory unless the task explicitly requires otherwise.

- Do not define multiple method-bearing structs in the same `.go` file.
- Prefer declaring a method-bearing struct as a standalone `type Name struct { ... }`.
- A method-bearing struct should usually live in its own file, named after the primary type or responsibility whenever practical, for example:
    - `server.go` for `Server`
    - `router.go` for `Router`
    - `spec.go` for `Spec`
    - `renderer.go` for `Renderer`
- Grouped `type ( ... )` declarations are allowed for interfaces, passive data-only structs, and other small related helper/value types that belong to the same narrow concern.
- A grouped `type ( ... )` block may also contain exactly one method-bearing struct when:
    - it is the only behavioral type in the file, and
    - the other grouped types are passive helper/value types from the same narrow concern.
- Do not use grouped `type ( ... )` declarations to hide multiple substantial behavioral types.
- If a helper struct later gains methods and would create more than one method-bearing struct in the file, extract it into its own file immediately.
- Methods for a struct should live in the same file as the struct unless there is a strong, explicit reason to split by concern.
- Do not place a new method-bearing struct into an existing file just because the code compiles.

Allowed:

```go
type (
	RouteMatch struct {
		PathParams map[string]string
		Query      map[string][]string
	}

	ResponseChoice struct {
		Status int
		Body   []byte
	}

	Matcher interface {
		Match(*http.Request) (*RouteMatch, bool)
	}
)
```

Avoid:

```go
type (
	Server struct {
		// ...
	}

	renderer struct {
		// ...
	}
)
```

Rationale:

- one method-bearing type per file keeps ownership of behavior obvious
- standalone method-bearing types make diffs and reviews clearer
- grouped type blocks are fine for passive, closely related types, but should not hide substantial behavioral types

## Function and method ownership rules

These rules are mandatory unless the task explicitly requires otherwise.

- A file centered on a method-bearing type should contain the type, its methods, and its constructors only.
- Do not mix package-level helper functions into a file that already contains methods for a primary type.
- In type-centered files, constructor functions are the only normally allowed package-level functions.
- If logic conceptually belongs to the primary type, implement it as a method.
- If logic does not belong to the type and must remain a package-level function, place it in a separate helper-focused file.
- Package-level functions are preferred only when there is no natural owning type or when the behavior is genuinely package-level.
- If a file contains both methods and non-constructor package-level functions, that is usually a structure violation and should be refactored.

## Comment rules for functions and methods

- Do not add comments to every function or method by default.
- Exported functions and methods should usually have doc comments, especially for command, service, mock-spec, or integration-facing packages.
- Unexported functions and methods should be commented only when they carry non-obvious behavior, invariants, side effects, ownership rules, cleanup expectations, or protocol/lifecycle constraints.
- Comments must explain intent, contract, invariants, side effects, or lifecycle behavior.
- Prefer comments that explain why the code exists, what must remain true, or how the method is meant to be used.
- Do not write comments that merely restate the method name or signature.
- For server lifecycle, mock matching, template rendering, path handling, and Ferret integration, prefer comments on semantics and invariants over implementation narration.
- Avoid comment wallpaper. Dense, meaningful comments are preferred over mechanically documenting obvious code.

Preferred:

```go
// Shutdown stops accepting new requests and waits for in-flight requests to finish
// until the context expires. It is safe to call multiple times.
func (s *Server) Shutdown(ctx context.Context) error
```

Preferred for internal code:

```go
// selectResponse applies Lab extension precedence before falling back to OpenAPI
// examples so dynamic mocks remain explicit and predictable.
func (m *Mock) selectResponse(...)
```

Avoid:

```go
// Shutdown shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error
```

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

- request latency
- server startup or shutdown behavior
- execution throughput
- memory use or allocation patterns
- template rendering cost
- route matching performance
- mock response selection performance
- browser/process lifecycle behavior
- Ferret execution setup or result handling cost
- asset build or generated output behavior

This includes, but is not limited to, changes in:

- server routing or middleware
- mock request matching
- OpenAPI spec loading or normalization
- template rendering and function registration
- file/workspace scanning
- Ferret execution integration
- caching, pooling, buffering, or process lifecycle code

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

- CLI behavior should have command-focused tests where practical.
- Config behavior should test defaults, explicit values, invalid values, and precedence.
- Server routing behavior should test route registration, middleware, shutdown-sensitive behavior, and error responses.
- Mock API behavior should test request matching, response selection, template rendering, latency/error simulation, and request logs.
- OpenAPI extension behavior should include valid and invalid spec fixtures.
- Template behavior should test render context, function behavior, escaping expectations, and error reporting.
- Random/fake-data behavior should test deterministic seeded output.
- Ferret execution behavior should have integration tests using real Ferret execution where practical.
- UI/API contract behavior should be tested at the API boundary when frontend tests are unavailable.
- File/workspace behavior should test cleanup and unsafe path handling.

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
- If a cleanup is necessary to make the requested change safe, keep it tightly scoped and explain why it was needed.

### Comment and documentation discipline

- Add comments where semantics, invariants, side effects, ownership, lifecycle, or recovery behavior are non-obvious.
- Do not add comment wallpaper.
- Prefer comments that explain why, contract, or invariants rather than implementation narration.
- Public and integration-facing behavior should be documented more carefully than local obvious helpers.

### Decision bias when uncertain

When uncertain:

- preserve existing behavior
- prefer the smaller local change
- add a focused test
- treat the change as significant if performance might be affected
- verify ownership before introducing a new abstraction or package-level dependency

## Tooling prerequisites

- Go must be installed.
- `make` is optional but is the preferred entrypoint for repo-defined workflows when present.
- If the repository has frontend assets, verify the Node/package-manager version from the repository before running asset commands.
- Lint, format, and generation tools should be installed through repository-defined commands when available.

## Command matrix

Prefer repository-defined commands from `Makefile` and `.github/workflows/build.yml`. Verify these names against the current repository before using them.

- Broad validation: `go test ./...`
- Lint: `make lint`
- Format: `make fmt`
- Generate derived files/assets: `make generate`
- Build: `make build` or the repository-defined build target
- Run local Lab server: use the repository-defined command or documented `go run` entrypoint

Run generation commands only when generator inputs change.

## Editing rules

- Never hand-edit generated files when their source inputs are available.
- If generated frontend/static assets are committed, regenerate them through the repository-defined build/generate command.
- Treat `Makefile` and `.github/workflows/build.yml` as the source of truth for validation commands.
- Prefer narrow validation first, then broaden:
    - Package-local changes: run the affected `go test` package or packages.
    - Mock/API changes: run mock/spec/server tests.
    - Ferret integration changes: run integration tests that execute real Ferret behavior.
    - Cross-cutting changes: finish with `go test ./...` or the repo-level test target.

### Validation expectations

- After code changes, run the narrowest tests that prove the behavior you touched.
- Before finishing broader changes, run the relevant repo-level command from the matrix above.
- If you changed formatting-sensitive files, run the repository format command.
- If you changed lint-sensitive code paths or public behavior, run the repository lint command when the toolchain is available.
- If you changed generator inputs, generated output must be included and reviewed when the repository commits generated output.

### Expectations for non-trivial changes

When proposing or implementing non-trivial changes:

- identify the owning subsystem first
- preserve invariants unless the task explicitly changes them
- prefer local, comprehensible changes before introducing new abstractions
- distinguish correctness work from performance work
- do not perform opportunistic refactors unrelated to the requested task unless they are necessary for correctness

## Secondary references

- `README.md` for product context, usage examples, and links to the broader Ferret ecosystem.
- `CONTRIBUTING.md` for human contributor process when present.
- `.github/workflows/build.yml` for the current CI validation path.
- Existing examples and fixtures for expected Lab behavior.
- Ferret core documentation for FQL language semantics.
