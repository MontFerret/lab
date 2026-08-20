# Architecture

Lab is a test runner and local test-environment companion for Ferret v2. It discovers FQL tests, prepares parameters and optional local services, schedules execution through a Ferret runtime, and reports the results.

The current implementation and tests are authoritative for behavior. This guide describes the durable package boundaries that contributors should preserve.

## Execution model

The primary execution pipeline is:

```text
locations -> pkg/sources -> pkg/testing -> pkg/runner
          -> pkg/runtime -> Ferret -> pkg/reporters
```

The local-service pipeline is:

```text
CLI flags and environment
    -> pkg/localserver entries and settings
    -> pkg/staticserver or pkg/mockserver
    -> endpoint URLs
    -> @lab.static and @lab.mock
```

The command layer assembles these components. It does not own source discovery, test semantics, scheduling, runtime integration, server internals, or result formatting.

## Package responsibilities

### Application and commands

- `main.go` constructs the application, injects the Lab version, registers commands, and turns operating-system signals into context cancellation.
- `cmd` defines `run`, `serve`, and `version`, binds `LAB_*` environment values, validates command input, and converts CLI values into package-owned options.

Command behavior and lifecycle are described in [CLI](cli.md).

### Test execution

- `pkg/sources` discovers content from local files, directories, globs, Git repositories, HTTP URLs, and aggregates.
- `pkg/testing` represents direct FQL tests and YAML query/assertion suites, manages user and Lab parameters, and executes test-level validation.
- `pkg/runner` schedules source files, clones per-test parameters, applies concurrency, retries, repeated runs, intervals, timeouts, and cancellation, and produces progress and summary streams.
- `pkg/reporters` consumes runner streams and presents results without controlling execution.

These stages are described in [Test execution](test-execution.md).

### Runtime adapters

`pkg/runtime` is the only Lab package that executes FQL. It selects and configures the embedded Ferret runtime, the remote HTTP adapter, the external Ferret CLI adapter, or a function-backed adapter. It passes source content and parameters through those contracts without taking ownership of FQL semantics.

Adapter behavior is described in [Runtime](runtime.md).

### Local services

- `pkg/localserver` provides shared entry parsing, bind/advertise settings, endpoint formatting, and manager/node lifecycle.
- `pkg/staticserver` supplies handlers for local directories.
- `pkg/mockserver` loads OpenAPI-compatible specifications and provides deterministic mock routing and response templates.

Local-service behavior is described in [Local services](local-services.md).

### Build and release infrastructure

The `Makefile`, workflows under `.github/workflows`, scripts, Dockerfiles, assets, and `.goreleaser.yml` own developer workflows, CI, dependency updates, binary archives, containers, and releases. See [Release](release.md).

## Dependency direction

Lab depends on Ferret; Ferret does not depend on Lab. Within Lab, dependencies follow the execution flow rather than reversing it for convenience:

- commands depend on package-owned contracts
- the runner coordinates sources, tests, and runtimes without taking ownership of their internals
- reporters depend on runner results, but the runner does not depend on presentation details
- static and mock servers reuse `pkg/localserver` instead of duplicating shared lifecycle logic
- runtime adapters use Ferret APIs or explicit external protocols; other Lab packages do not reach through adapters into Ferret internals

Packages under `pkg` are internal-to-Lab boundaries despite exported Go names. Exports support cross-package composition within Lab and are not a promise of a general-purpose external library API.

## Architectural invariants

- Ferret owns FQL syntax, compilation, bytecode, VM behavior, and runtime values.
- User parameters and Lab system parameters remain isolated until execution materializes them. Lab values use the `lab` namespace.
- Context cancellation must propagate through source loading, scheduling, runtimes, reporters, and local services wherever the integration supports it.
- Owned resources must be released on success, failure, timeout, cancellation, and partial startup.
- Source identity and useful error context must survive the pipeline.
- Reporters are observational; formatting cannot alter retries, cancellation, runtime behavior, or result semantics.
- Ordering that affects behavior must be deterministic and independent of Go map iteration.

## Stability boundaries

The following are relatively stable unless a change explicitly targets them:

- the `run`, `serve`, and `version` command roles
- the execution flow through sources, testing, runner, runtime, and reporters
- the `@lab` system parameter namespace
- local static/mock entry and alias behavior
- the built-in, remote HTTP, and external binary runtime split

The following are implementation-sensitive and should be rechecked in current code and tests before modification:

- concurrency, retries, repeated runs, result streams, timeouts, and cancellation
- parameter cloning and materialization
- remote HTTP requests and binary process arguments
- source fetching, identity, temporary resources, and cleanup
- local-server startup, shutdown, dynamic ports, and advertised endpoints
- mock route specificity, request context, and template rendering
- reporter summary and exit behavior

Historical v1 behavior and old design discussions are not authoritative for this branch.
