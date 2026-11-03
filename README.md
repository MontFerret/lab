# Lab - Ferret Test Runner

<p align="center">
	<a href="https://goreportcard.com/report/github.com/MontFerret/lab">
		<img alt="Go Report Status" src="https://goreportcard.com/badge/github.com/MontFerret/lab">
	</a>
<!-- 	<a href="https://codecov.io/gh/MontFerret/lab">
		<img alt="Code coverage" src="https://codecov.io/gh/MontFerret/lab/branch/main/graph/badge.svg" />
	</a> -->
	<a href="https://discord.gg/kzet32U">
		<img alt="Discord Chat" src="https://img.shields.io/discord/501533080880676864.svg">
	</a>
	<a href="https://github.com/MontFerret/lab/releases">
		<img alt="Lab release" src="https://img.shields.io/github/release/MontFerret/lab.svg">
	</a>
	<a href="https://opensource.org/licenses/Apache-2.0">
		<img alt="Apache-2.0 License" src="http://img.shields.io/badge/license-Apache-brightgreen.svg">
	</a>
</p>

**Lab** is a flexible test runner designed specifically for [Ferret](https://www.github.com/MontFerret/ferret) scripts. It helps automate testing for web scraping, browser automation, API integration, and regression scenarios written in Ferret Query Language (FQL).

> **📢 Notice:** This branch contains the upcoming **Lab for Ferret v2**. For the stable v1 release, please visit [Lab v1](https://github.com/MontFerret/lab/tree/v1).

---

**🚀 Perfect for:**

- End-to-end web application testing
- Web scraping validation and monitoring
- API integration testing
- Browser automation testing
- Regression testing for web applications

Read the introductory blog post about Lab [here](https://www.montferret.dev/blog/say-hello-to-lab/).

<p align="center">
<img alt="lab" src="https://raw.githubusercontent.com/MontFerret/lab/main/assets/landing.png" style="margin-left: auto; margin-right: auto;" width="495px" height="501px" />
</p>

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Test Suites](#test-suites)
- [Choose a Runtime](#choose-a-runtime)
- [Advanced Usage](#advanced-usage)
- [Configuration Reference](#configuration-reference)
- [Architecture](#architecture)
- [Development](#development)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Features

### 🏃‍♂️ Performance & Scalability

- **Parallel execution** - Run multiple tests concurrently for faster feedback
- **Configurable concurrency** - Control the number of simultaneous test executions
- **Test retry mechanism** - Automatically retry failed tests with customizable attempts
- **Batch execution** - Run tests multiple times with configurable intervals

### 🌐 Flexible Runtime Support

- **Built-in Ferret runtime** - Execute tests using the embedded Ferret engine
- **Remote HTTP runtime** - Connect to remote Ferret services over HTTP/HTTPS
- **Wire runtime** - Execute tests through a hosted Universal Ferret API over gRPC
- **Multi-runtime testing** - Test against different Ferret versions or runtime configurations

### 📁 Multiple Source Types

- **Local filesystem** - Execute scripts from local directories
- **Git repositories** - Fetch and run tests directly from Git repositories over HTTP/HTTPS
- **HTTP sources** - Download and execute scripts from web URLs
- **Glob pattern matching** - Select multiple files using wildcard patterns

### 🌍 Static Content Serving

- **Built-in HTTP server** - Serve static files for testing web applications
- **Dedicated `serve` command** - Run local static and mock services without executing tests
- **Multiple static endpoints** - Host different directories at separate URLs
- **Custom aliases** - Name content endpoints for better organization
- **Dynamic port allocation** - Automatically find available ports

### 🔌 Mock API Serving

- **OpenAPI-compatible specs** - Serve lightweight REST mocks from YAML or JSON files
- **Lab mock extensions** - Define responses with operation-level `x-lab-mock`
- **Dynamic templates** - Use request method, path params, query params, headers, and JSON bodies in responses
- **Stable local API URLs** - Expose mock endpoints to FQL scripts under `@lab.mock.<alias>`

### 📊 Rich Reporting & Monitoring

- **Multiple output formats** - Console and simple reporters are available
- **Detailed test results** - Execution metrics and timing information
- **Wait conditions** - Wait for external services before running tests
- **Environment variable support** - Configure tests via environment variables

## Installation

### 📦 Binary Downloads

Choose a published **v2 release tag** from the [releases page](https://github.com/MontFerret/lab/releases), including prereleases. Set the tag before using the examples below; replace the placeholder with the exact tag you selected:

```bash
LAB_VERSION='<v2-release-tag>'
```

GitHub's `releases/latest` endpoint and Docker's `latest` tag select stable releases and may still point to v1. Select a v2 version explicitly.

The recommended local install location is `~/.local/bin`, which does not require elevated permissions.

#### Linux

For Linux x86-64:

```bash
curl -fL "https://github.com/MontFerret/lab/releases/download/${LAB_VERSION}/lab_linux_x86_64.tar.gz" -o lab.tar.gz
tar -xzf lab.tar.gz
mkdir -p ~/.local/bin
mv lab ~/.local/bin/
```

For Linux ARM64, choose `lab_linux_arm64.tar.gz` from the same release.

#### macOS Intel

```bash
curl -fL "https://github.com/MontFerret/lab/releases/download/${LAB_VERSION}/lab_darwin_x86_64.tar.gz" -o lab.tar.gz
tar -xzf lab.tar.gz
mkdir -p ~/.local/bin
mv lab ~/.local/bin/
```

#### macOS Apple Silicon

```bash
curl -fL "https://github.com/MontFerret/lab/releases/download/${LAB_VERSION}/lab_darwin_arm64.tar.gz" -o lab.tar.gz
tar -xzf lab.tar.gz
mkdir -p ~/.local/bin
mv lab ~/.local/bin/
```

Make sure `~/.local/bin` is in your `PATH`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Add that line to your shell profile to keep the change across sessions.

#### System-wide install

If you prefer installing Lab system-wide, move the binary to `/usr/local/bin` instead:

```bash
sudo mv lab /usr/local/bin/
```

This location may require administrator privileges.

#### Windows

Download `lab_windows_x86_64.zip` or `lab_windows_arm64.zip` from the selected v2 release, extract `lab.exe`, and place it in a directory included in your `PATH`.

### 🚀 Install Script

The repository's [install.sh](install.sh) is a Bash installer. It uses `VERSION` and `LOCATION`, defaults to `$HOME/.ferret`, and requires the destination directory to exist. Its version validation does not accept prerelease tags, so use the binary downloads above or a source build for v2 prereleases.

### 🐳 Docker

Use the selected v2 release's image tag, without the leading `v`:

```bash
# LAB_VERSION is the v2 release tag selected above.
LAB_IMAGE_VERSION="${LAB_VERSION#v}"
docker pull "montferret/lab:${LAB_IMAGE_VERSION}"

# Run local tests in the container
docker run --rm -v "$PWD:/workspace" "montferret/lab:${LAB_IMAGE_VERSION}" run /workspace/tests/

# With custom options
docker run --rm -v "$PWD:/workspace" "montferret/lab:${LAB_IMAGE_VERSION}" \
  run --concurrency=4 --reporter=simple /workspace/tests/
```

The image starts its browser environment for Lab commands; the selected Ferret runtime still determines which FQL modules are available. The entrypoint executes explicit non-Lab commands directly in the container.

**Docker Compose Example:**

```yaml
services:
  lab:
    image: montferret/lab:${LAB_IMAGE_VERSION}
    volumes:
      - ./tests:/workspace/tests
      - ./static:/workspace/static
    command: ["run", "--serve", "/workspace/static", "/workspace/tests/"]
```

Set `LAB_IMAGE_VERSION` in Compose's environment or `.env` file to the v2 image tag you selected.

### 🛠️ Build from Source

Use the Go toolchain required by [go.mod](go.mod), plus Git and Make:

```bash
git clone https://github.com/MontFerret/lab.git
cd lab
make compile
./bin/lab version
```

`make compile` injects both Lab and embedded Ferret version metadata. Use `make build` to run vet and tests before compiling.

### ✅ Verify Installation

```bash
lab version
lab --help
lab run --help
```

`lab version` reports the Lab binary version and the selected Ferret runtime version.

## Quick Start

### 🎯 Basic Usage

Script execution requires the `run` subcommand:

```bash
# Execute a single FQL script
lab run myscript.fql

# Run all supported tests in a directory, recursively
lab run tests/

# Run with increased concurrency
lab run --concurrency=4 tests/

# Require three successful executions of each test
lab run --times=3 myscript.fql
```

To select an HTTP or Wire host, see [Choose a Runtime](#choose-a-runtime). Root invocation such as `lab tests/` is not supported.

### 📝 Your First Test

Create `example.fql`. This example uses only Ferret core and runs with the built-in runtime:

```sql
LET products = [
  { name: "Keyboard", price: 129 },
  { name: "USB-C Dock", price: 89 }
]

T::LEN(products, 2, "expected two products")
T::NOT::EMPTY(products[0].name, "the first product needs a name")

RETURN true
```

Run it:

```bash
lab run example.fql
```

A test passes when execution completes without an error. Lab discards a unit's return value, so `RETURN false` also passes. Use `T::TRUE(condition)`, `T::EQ(actual, expected)`, or another assertion to raise an error when a condition is false.

### Using a Different Runtime

Lab runs FQL with its built-in Ferret core runtime by default. To use optional modules, including browser functionality such as `web::html`, connect to an external runtime that provides them. Lab connects to an already-running service; it does not start that host for you.

| Runtime | Use when | Target |
| --- | --- | --- |
| Built-in | Your tests use Ferret core and its standard library. | Omit `--runtime`, or use `builtin`. |
| HTTP/HTTPS | You have a Ferret service implementing Lab's HTTP protocol. | The service's execution URL. |
| Wire | You have a host exposing the Universal Ferret API over gRPC. | `wire+tcp://127.0.0.1:<port>`. |

```bash
# Built-in runtime
lab run tests/

# External HTTP runtime
lab run --runtime=https://ferret.example.com/api tests/

# External Wire runtime
lab run --runtime=wire+tcp://127.0.0.1:50051 tests/

# Check the version reported by an external runtime
lab version --runtime=wire+tcp://127.0.0.1:50051
```

Wire accepts IPv4 loopback endpoints only. The host must support Wire's chunked output protocol. For HTTP headers, Wire connection timeouts, host requirements, and execution behavior, see [Remote Ferret Runtime](#-remote-ferret-runtime) and [Runtime Options](#-runtime-options).

### Using Chrome DevTools Protocol

Browser-backed FQL requires an external runtime with the relevant browser module and driver configured. Browser endpoints are resolved from that runtime's environment. Lab has no `--cdp` flag or `LAB_CDP` setting.

After starting the host and browser, you can wait for a browser health endpoint reachable from Lab before running browser tests:

```bash
lab run \
  --runtime=wire+tcp://127.0.0.1:50051 \
  --wait=http://127.0.0.1:9222/json/version \
  browser-tests/
```

`--wait` checks readiness; it does not configure browser access. See the [CDP guide](https://www.montferret.dev/docs/tools/lab/cdp/) for runtime-specific configuration and [Remote Ferret Runtime](#-remote-ferret-runtime) for host integration.

### Example Output

Use the simple reporter for line-oriented output:

```bash
lab run --reporter=simple example.fql
```

```text
PASS file="example.fql" duration=1ms attempts=1 times=1
DONE passed=1 failed=0 duration=2ms
```

Paths and durations vary. A failed test produces a `FAIL` line and a nonzero exit status; expected-error tests that meet their expectation produce `PASS`.

## Test Suites

Lab runs `.fql` units and YAML suites (`.yaml` or `.yml`). Use a unit for ordinary execution with in-query assertions. Use a suite to separate a query from its assertions or require an execution error.

A suite contains `query` and either `assert` or `expect.error`, with an optional `timeout` in seconds. Each script defines exactly one of `text` or `ref` and may include `params`. Setup and cleanup are resource lifecycle operations; there are no YAML `setup` or `cleanup` script stages.

### 📋 Basic Test Suite Structure

Save this as `products.yaml`:

```yaml
query:
  text: |
    LET products = [
      { name: "Keyboard", price: 129 },
      { name: "USB-C Dock", price: 89 }
    ]

    RETURN products[*].name

assert:
  text: |
    LET names = @lab.data.query.result
    T::LEN(names, 2)
    RETURN T::CONTAINS(names, "Keyboard")
```

Run it with the built-in runtime:

```bash
lab run products.yaml
```

The assertion receives the completed query's JSON result in `@lab.data.query.result` and its effective parameters in `@lab.data.query.params`. If the query fails, the assertion does not run. Assertion return values are discarded; assertions must raise errors to fail the test.

### Expected Runtime Errors

Use `expect.error` when a query must fail. An empty error object accepts an execution error, including a compilation error:

```yaml
query:
  text: |
    RETURN 1 / 0

expect:
  error: {}
```

Add `contains` to require a stable part of the error message:

```yaml
query:
  text: |
    RETURN 1 / 0

expect:
  error:
    contains: "division by zero"
```

Only `contains` is supported inside `expect.error`; unknown fields fail suite construction. A suite using `expect.error` must not define `assert`. If the query succeeds, the test fails.

The legacy `.fail.fql` filename convention accepts execution errors and emits a deprecation warning. Prefer YAML `expect.error` for new negative tests. Output, session, and plan cleanup failures observed by Lab remain failures; they cannot satisfy an expected-error test or be hidden by a later successful retry. Admission errors already combined inside a runtime retain their existing expected-error handling.

### 🔗 Reference External Scripts

Save a reusable query as `scripts/products.fql`:

```sql
RETURN { category: @category, names: ["Keyboard", "USB-C Dock"] }
```

Reference it from `suite.yaml` in the parent directory:

```yaml
query:
  ref: ./scripts/products.fql
  params:
    category: accessories

assert:
  text: |
    LET result = @lab.data.query.result
    T::EQ(result.category, "accessories")
    RETURN T::LEN(result.names, 2)
```

References resolve relative to the suite's source location. Assertion scripts may also use `ref`.

### Complex Test Scenarios

Use several assertions to validate a structured result. This inventory example runs without external services:

```yaml
timeout: 10
query:
  params:
    inventory:
      - sku: keyboard
        stock: 3
        price: 129
      - sku: dock
        stock: 5
        price: 89
  text: |
    RETURN @inventory

assert:
  text: |
    LET inventory = @lab.data.query.result
    T::LEN(inventory, 2)

    FOR product IN inventory
      T::NOT::EMPTY(product.sku)
      T::GTE(product.stock, 0)
      T::GT(product.price, 0)
      RETURN true
```

The top-level timeout covers the query and assertion together on each attempt. Arrange any external fixture setup and teardown in your surrounding test environment.

### Parameterized Tests

Save this as `test-suite.yaml`:

```yaml
query:
  text: |
    RETURN { name: @productName, stock: @stock }

assert:
  text: |
    LET product = @lab.data.query.result
    T::NOT::EMPTY(product.name)
    RETURN T::GTE(product.stock, @minStock)
```

Quote each complete parameter binding so the shell preserves JSON strings:

```bash
lab run \
  --param='productName:"Keyboard"' \
  --param='stock:5' \
  --param='minStock:2' \
  test-suite.yaml
```

Query and assertion `params` override common user parameters independently; assertion parameters do not inherit query-only overrides. Lab system values remain under the reserved `@lab` namespace.

Use `--param-bind` when a script should keep a production-oriented parameter while Lab supplies its value. The target is a dot-separated user parameter path; the source starts with `@` and must already exist when Lab finishes local-service setup:

```bash
lab run tests/ \
  --param='environment:{"api":{"baseUrl":"https://example.com"}}' \
  --param-bind config.api.baseUrl=@environment.api.baseUrl
```

Bindings preserve the source value's type and resolve from one parameter snapshot. Duplicate or overlapping targets, conflicts with `--param`, malformed paths, and missing sources are configuration errors. Paths are case-sensitive; each segment starts with a letter or underscore, followed by letters, numbers, underscores, or hyphens. Bracket expressions and targets in the `lab` namespace are rejected.

### Data-Driven Testing

Validate multiple cases in one suite without treating returned booleans as assertions:

```yaml
query:
  text: |
    RETURN [
      { name: "ready status", actual: "ready", expected: "ready" },
      { name: "queued status", actual: "queued", expected: "queued" }
    ]

assert:
  text: |
    FOR item IN @lab.data.query.result
      T::EQ(item.actual, item.expected, item.name)
      RETURN true
```

## Advanced Usage

### 📁 File Resolution

Lab reads local files and directories, Git repositories over HTTP/HTTPS, and HTTP/HTTPS source URLs. Directories are traversed recursively; `.fql`, `.yaml`, and `.yml` are supported.

#### Local Files

```bash
# Single file or directory
lab run /path/to/test.fql
lab run tests/integration/

# Filter files within a directory using a quoted source URL
lab run 'file://tests?filter=**/*.fql'

# Multiple locations
lab run --files=tests/unit/ --files=tests/integration/ --files=scripts/smoke.fql
```

The `filter` query parameter uses glob matching against full filesystem paths. A quoted shell pattern such as `"tests/**/*.fql"` is not expanded by Lab; use a directory with a filter instead.

#### Git Repositories

Use the repository URL with the `git+http` or `git+https` scheme:

```bash
lab run git+https://github.com/username/test-repo.git
lab run git+http://git.example.com/tests.git
```

Lab reads supported files from the cloned repository's default HEAD. It does not provide the `repo.git//subdirectory`, `repo.git@revision`, or `git+ssh` URL forms. To use a particular revision or an authenticated checkout, prepare it with Git and run Lab against the resulting local directory.

#### HTTP Sources

```bash
# Direct script URL
lab run https://raw.githubusercontent.com/user/repo/main/test.fql

# Multiple HTTP sources
lab run https://example.com/tests/suite1.yaml https://example.com/tests/suite2.yaml
```

These URLs supply test source files. Select an external execution service separately with `--runtime`.

### 🌐 Static File Serving

Lab can serve local directories over HTTP during test execution. Bind a generated endpoint to an ordinary script parameter when the script should remain unaware of Lab.

#### Standalone Static Server

Use `lab serve --static` when you want to expose local directories over HTTP without running a test suite.

```bash
# Serve a single directory
lab serve --static ./website

# Serve multiple directories with aliases
lab serve --static ./frontend@app --static ./mockdata@api
```

#### Production-Style Fixture Parameters

```bash
lab run tests/ \
  --serve ./tests/fixtures@fixtures \
  --param-bind baseUrl=@lab.static.fixtures \
  --policy-http-allow-localhost
```

FQL script:

```sql
LET response = IO::NET::HTTP::GET(@baseUrl + "/products.json")
RETURN T::NOT::EMPTY(TO_STRING(response))
```

Lab resolves the binding after the fixture server has selected its port and before test execution starts. Scripts that deliberately depend on Lab can continue to use `@lab.static.<alias>` directly.

```sql
LET response = IO::NET::HTTP::GET(@lab.static.fixtures + "/products.json")
RETURN T::NOT::EMPTY(TO_STRING(response))
```

#### Multiple Static Endpoints

```bash
lab run --policy-http-allow-localhost --serve ./app --serve ./api-mocks tests/
```

FQL script:

```sql
LET appPage = IO::NET::HTTP::GET(@lab.static.app + "/index.html")
LET apiData = IO::NET::HTTP::GET(@lab.static["api-mocks"] + "/users.json")
RETURN [TO_STRING(appPage), TO_STRING(apiData)]
```

#### Custom Static Aliases

```bash
lab run --policy-http-allow-localhost --serve ./frontend@app --serve ./mockdata@api tests/
```

FQL script:

```sql
LET homePage = IO::NET::HTTP::GET(@lab.static.app + "/index.html")
LET userData = IO::NET::HTTP::GET(@lab.static.api + "/user/123.json")
RETURN [TO_STRING(homePage), TO_STRING(userData)]
```

#### Advanced Static Server Example

```bash
lab run \
  --policy-http-allow-localhost \
  --serve ./dist@webapp \
  --serve ./test-fixtures@fixtures \
  --serve ./mock-apis@mocks \
  --concurrency=3 \
  tests/e2e/
```

#### Reachable Static URLs for Remote Runtimes

When Lab runs against a remote runtime or inside a containerized environment, use `--serve-host` to advertise a client-reachable host instead of loopback. Use `--serve-bind` when you need to control the listener interface explicitly.

```bash
lab run \
  --runtime=https://ferret.example.com \
  --serve ./dist@app \
  --serve-host host.docker.internal \
  --serve-bind 0.0.0.0 \
  tests/
```

If `--serve-host` is set without `--serve-bind`, Lab automatically binds static servers to all interfaces: `0.0.0.0` for IPv4 and hostnames, or `::` for IPv6 literals.

### 🔌 Mock API Serving

Lab can serve local REST mock APIs from OpenAPI-compatible YAML or JSON specs during test execution. Mock endpoints are available in FQL scripts under `@lab.mock.<alias>`.

Mock specs use operation-level `x-lab-mock` blocks:

```yaml
openapi: 3.1.0
info:
  title: Users API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      x-lab-mock:
        status: 200
        body:
          id: "{{ .Path.id }}"
          name: "User {{ .Path.id }}"
```

Run tests with the mock API:

```bash
lab run --policy-http-allow-localhost --mock ./users.yaml@api tests/
```

Ferret's built-in HTTP client blocks localhost by default. Tests that access `@lab.static` or `@lab.mock` through `IO::NET::HTTP` must opt in with `--policy-http-allow-localhost` (or `LAB_POLICY_HTTP_ALLOW_LOCALHOST=true`). Starting a local service does not broaden the HTTP policy automatically.

Serve the mock API without running tests:

```bash
lab serve --mock ./users.yaml@api
```

Static and mock services can be served together:

```bash
lab serve --static ./dist@app --mock ./users.yaml@api
```

FQL script:

```sql
LET payload = ENCODING::JSON_PARSE(TO_STRING(IO::NET::HTTP::GET(@lab.mock.api + "/users/123")))
RETURN T::EQ(payload.id, "123")
```

Mock API entries use the same binding syntax as static serving: `<path>`, `<path>:<port>`, `<path>@<alias>`, and `<path>@<alias>:<port>`. `--serve-bind` and `--serve-host` also apply to mock API servers.

### 🔒 Filesystem Policy

The built-in runtime exposes FQL filesystem functions through a sandbox rooted at Lab's current working directory. Use `--policy-fs-root` to select a narrower relative or absolute root, and add `--policy-fs-read-only` to permit reads while rejecting writes, directory changes, and removals.

```bash
lab run \
  --policy-fs-root=./fixtures \
  --policy-fs-read-only \
  tests/
```

The equivalent environment variables are `LAB_POLICY_FS_ROOT` and `LAB_POLICY_FS_READ_ONLY`.

### Remote Ferret Runtime

For basic commands, see [Choose a Runtime](#choose-a-runtime). This section covers adapter protocols and host requirements; [Runtime Options](#-runtime-options) lists configuration values and precedence.

Lab's adapters implement the Universal Ferret API. The command owns the shared runtime, and each test case owns its execution resources through a case-scoped executor. Outputs settle and temporary sessions close on every invocation; cached plans close before the case result is reported. Go callers using compilation directly own and close their plans and sessions.

#### HTTP/HTTPS Runtime

```bash
# Configure an existing HTTP service
lab run \
  --runtime=https://ferret.example.com \
  --runtime-option='headers:{"Authorization":"Bearer token123"}' \
  --runtime-option='path:"/v1/execute"' \
  --param='value:"shared"' \
  tests/
```

When the runtime URL already includes a path, Lab sends `run` requests to that exact path. The optional `runtime-option=path` value overrides the run endpoint only. `lab version --runtime=...` uses the runtime URL path and appends `/info` (`/api/info` for a runtime URL ending in `/api`).

Lab rejects `--policy-fs-*` and `--policy-http-*` options for HTTP runtimes because their request protocol has no policy contract. The built-in runtime supports these options.

The HTTP runtime sends POST requests with:

```json
{
  "text": "FQL script content",
  "params": {
    "key": "value"
  }
}
```

A successful HTTP run returns a one-shot output that streams the response body. Collection preserves the response content type and any bytes received before a read or cleanup error. Suites decode completed query output as JSON. The HTTP protocol supports execution and version reporting; UAPI compilation, filesystem-root, and output-codec options are rejected without sending requests.

#### Wire Runtime

Wire exposes the Universal Ferret API over gRPC. Configure connection setup separately from FQL parameters:

```bash
lab run \
  --runtime=wire+tcp://127.0.0.1:50051 \
  --runtime-option='connect-timeout:"10s"' \
  --param='value:"shared"' \
  tests/

lab version \
  --runtime=wire+tcp://127.0.0.1:50051 \
  --runtime-option='connect-timeout:"10s"'

LAB_RUNTIME=wire+tcp://127.0.0.1:50051 \
LAB_RUNTIME_OPTION='connect-timeout:"10s"' \
lab run tests/
```

Start a Wire host separately. The target must be `wire+tcp://127.0.0.1:<port>` with a port from 1 to 65535. Hostnames, IPv6, other addresses, paths, credentials, queries, and fragments are not supported. Plaintext loopback transport is intended for trusted local development, including a host exposed through a local port mapping. The URI scheme adds no encryption or authentication.

The `connect-timeout` runtime option defaults to `"5s"` and must be a positive duration string. It bounds connection and handshake setup; `--timeout` continues to bound each test. Both `run` and `version` accept the same runtime target and options.

Lab preserves source identity/content, uses Wire's one-shot `Runtime.Run` by default, and reuses compiled plans for repetitions or retries. Compiled sessions receive the same nested YAML parameter conversion and validation as one-shot execution. The host must support chunked output framing (Wire `v1.0.0-alpha.3`). Lab consumes each output before marking an attempt complete, so a terminal failure after the first chunk still fails the attempt. `lab version` reports the hosted `api.Runtime.Version` exactly, captured during the handshake.

Lab owns the gRPC transport and closes the logical Wire runtime before that transport. Setup failures also release the transport. Transport loss requires a new Lab invocation; the adapter does not redial or switch runtimes.

The host configures Ferret and its execution policies. Wire therefore rejects Lab's `--policy-fs-*` and `--policy-http-*` options. Only `connect-timeout` is a Wire adapter option. FQL inputs, including `headers`, `cookies`, `path`, and `flags`, use `--param`, bindings, or YAML query/assertion parameters. HTTP/Worker support remains separate.

#### Runtime Comparison Testing

Test against multiple runtime versions:

```bash
# Test with built-in runtime
lab run --reporter=simple tests/ > builtin-results.txt

# Test with remote runtime
lab run --reporter=simple --runtime=https://ferret.example.com/api tests/ > remote-results.txt

# Review results (paths, durations, and other host-specific details can differ)
diff builtin-results.txt remote-results.txt
```

### ⚡ Performance Optimization

#### Parallel Execution

```bash
# Run up to 8 tests simultaneously
lab run --concurrency=8 tests/

# Balance speed and resource usage
lab run --concurrency=4 --timeout=60 large-test-suite/
```

#### Test Repetition & Retry

```bash
# Run each test 3 times for reliability testing
lab run --times=3 tests/flaky/

# Allow up to three attempts for a failing test
lab run --attempts=3 tests/

# Add delay between test cycles
lab run --times=5 --times-interval=10 stress-tests/
```

By default, Lab uses one-shot execution. With `--times > 1` or `--attempts > 1`, built-in and Wire execution lazily reuse successful plans within each case and create a fresh session with fresh parameters on every invocation. Changed referenced script names or content select a different plan. Failed compilations are not cached; HTTP remembers unsupported compilation and continues with one-shot execution. All cached plans are released before the case result is reported.

`--times` counts successful executions. The reported attempt counter includes both repetitions and retries; it is not reset between repetitions. The interval applies between invocations and can be interrupted by cancellation.

#### Conditional Execution

```bash
# Wait for services to be available before running tests
lab run \
  --wait=http://127.0.0.1:9222/json/version \
  --wait-timeout=30 \
  tests/integration/
```

Only HTTP/HTTPS readiness checks are registered. `--wait-timeout` is the interval between attempts, in seconds; `--wait-attempts` sets the attempt count.

## Configuration Reference

### 🎛️ Command Line Flags

These flags apply to `lab run`.

| Flag | Short | Environment Variable | Default | Description |
|------|-------|----------------------|---------|-------------|
| `--files` | `-f` | `LAB_FILES` | - | Location of FQL script files to run |
| `--timeout` | `-t` | `LAB_TIMEOUT` | `30` | Test timeout in seconds |
| `--reporter` | - | `LAB_REPORTER` | `console` | Output reporter: `console`, `simple` |
| `--runtime` | `-r` | `LAB_RUNTIME` | Built-in | Complete runtime target: `builtin`, HTTP/HTTPS URL, or `wire+tcp://127.0.0.1:<port>` |
| `--runtime-option` | - | `LAB_RUNTIME_OPTION` | - | Repeatable adapter configuration: `name:<JSON value>` |
| `--concurrency` | `-c` | `LAB_CONCURRENCY` | `1` | Number of parallel test executions |
| `--times` | - | `LAB_TIMES` | `1` | Number of successful executions required per test |
| `--attempts` | `-a` | `LAB_ATTEMPTS` | `1` | Maximum attempts for a failing test |
| `--times-interval` | - | `LAB_TIMES_INTERVAL` | `0` | Interval between test cycles in seconds |
| `--serve` | - | `LAB_SERVE` | - | Served directory mapping exposed over HTTP |
| `--mock` | - | `LAB_MOCK` | - | OpenAPI mock API spec exposed over HTTP |
| `--serve-bind` | - | `LAB_SERVE_BIND` | - | Host to bind local servers to, without port |
| `--serve-host` | - | `LAB_SERVE_HOST` | - | Host to advertise local server URLs, without port |
| `--param` | `-p` | `LAB_PARAM` | - | Query parameters for tests |
| `--param-bind` | - | `LAB_PARAM_BIND` | - | Bind a user parameter path to an existing parameter reference |
| `--wait` | `-w` | `LAB_WAIT` | - | Wait for resource availability |
| `--wait-timeout` | `--wt` | `LAB_WAIT_TIMEOUT` | `5` | Interval between readiness attempts, in seconds |
| `--wait-attempts` | - | `LAB_WAIT_ATTEMPTS` | `5` | Number of wait attempts |
| `--policy-fs-root` | - | `LAB_POLICY_FS_ROOT` | Current working directory | Filesystem root for the built-in runtime |
| `--policy-fs-read-only` | - | `LAB_POLICY_FS_READ_ONLY` | `false` | Make the built-in runtime filesystem read-only |
| `--policy-http-allowed-schemes` | - | `LAB_POLICY_HTTP_ALLOWED_SCHEMES` | `http,https` | Allowed outbound HTTP URL schemes |
| `--policy-http-allowed-methods` | - | `LAB_POLICY_HTTP_ALLOWED_METHODS` | `GET,HEAD,POST,PUT,PATCH,DELETE,OPTIONS` | Allowed outbound HTTP methods |
| `--policy-http-allowed-hosts` | - | `LAB_POLICY_HTTP_ALLOWED_HOSTS` | - | Allowed exact hosts or `host:port` values |
| `--policy-http-blocked-hosts` | - | `LAB_POLICY_HTTP_BLOCKED_HOSTS` | - | Blocked exact hosts or `host:port` values |
| `--policy-http-allow-localhost` | - | `LAB_POLICY_HTTP_ALLOW_LOCALHOST` | `false` | Allow localhost and loopback addresses |
| `--policy-http-allow-private-networks` | - | `LAB_POLICY_HTTP_ALLOW_PRIVATE_NETWORKS` | `false` | Allow private network addresses |
| `--policy-http-allow-link-local` | - | `LAB_POLICY_HTTP_ALLOW_LINK_LOCAL` | `false` | Allow link-local addresses |
| `--policy-http-default-headers` | - | `LAB_POLICY_HTTP_DEFAULT_HEADERS` | - | Default request headers as a JSON string map |
| `--policy-http-blocked-request-headers` | - | `LAB_POLICY_HTTP_BLOCKED_REQUEST_HEADERS` | - | Blocked request header names |
| `--policy-http-timeout` | - | `LAB_POLICY_HTTP_TIMEOUT` | `30s` | Overall HTTP timeout |
| `--policy-http-no-timeout` | - | `LAB_POLICY_HTTP_NO_TIMEOUT` | `false` | Explicitly disable the overall HTTP timeout |
| `--policy-http-max-request-size` | - | `LAB_POLICY_HTTP_MAX_REQUEST_SIZE` | `16777216` | Maximum request body size in bytes |
| `--policy-http-unlimited-request-size` | - | `LAB_POLICY_HTTP_UNLIMITED_REQUEST_SIZE` | `false` | Explicitly disable the request body limit |
| `--policy-http-max-response-size` | - | `LAB_POLICY_HTTP_MAX_RESPONSE_SIZE` | `16777216` | Maximum response body size in bytes |
| `--policy-http-unlimited-response-size` | - | `LAB_POLICY_HTTP_UNLIMITED_RESPONSE_SIZE` | `false` | Explicitly disable the response body limit |
| `--policy-http-max-response-header-size` | - | `LAB_POLICY_HTTP_MAX_RESPONSE_HEADER_SIZE` | `1048576` | Maximum response header size in bytes |
| `--policy-http-follow-redirects` | - | `LAB_POLICY_HTTP_FOLLOW_REDIRECTS` | `true` | Follow HTTP redirects |
| `--policy-http-max-redirects` | - | `LAB_POLICY_HTTP_MAX_REDIRECTS` | `10` | Maximum redirects to follow |

These flags apply to `lab serve`.

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--static` | `LAB_STATIC` | - | Served directory mapping exposed over HTTP |
| `--mock` | `LAB_MOCK` | - | OpenAPI mock API spec exposed over HTTP |
| `--serve-bind` | `LAB_SERVE_BIND` | - | Host to bind local servers to, without port |
| `--serve-host` | `LAB_SERVE_HOST` | - | Host to advertise local server URLs, without port |

### 🌍 Environment Variables

Set environment variables for consistent configuration across environments:

```bash
# Basic configuration
export LAB_TIMEOUT=60
export LAB_CONCURRENCY=4
export LAB_REPORTER=simple

# Runtime configuration
export LAB_RUNTIME=https://ferret-api.example.com
export LAB_RUNTIME_OPTION='headers:{"API-Key":"secret123"}'

# Run tests
lab run tests/
```

### 📝 Configuration Examples

#### CI/CD Configuration

```bash
#!/bin/bash
# ci-test.sh

# Set CI-friendly defaults
export LAB_TIMEOUT=120
export LAB_CONCURRENCY=2
export LAB_REPORTER=simple
export LAB_ATTEMPTS=3

# Wait for services
lab run \
  --wait=http://app:3000/health \
  --wait-timeout=60 \
  tests/integration/
```

#### Local Development

```bash
#!/bin/bash
# dev-test.sh

export LAB_TIMEOUT=30
export LAB_CONCURRENCY=1

# Serve local assets and run tests
lab run \
  --serve ./dist@app \
  --serve ./fixtures@data \
  tests/dev/
```

#### Load Testing

```bash
#!/bin/bash
# load-test.sh

lab run \
  --concurrency=20 \
  --times=100 \
  --times-interval=1 \
  --timeout=10 \
  tests/performance/
```

### ⚙️ Runtime Options

Configure remote Ferret runtime behavior:

```bash
# HTTP runtime with custom headers
lab run \
  --runtime=https://ferret.api.com \
  --runtime-option='headers:{"Authorization":"Bearer token"}' \
  --runtime-option='path:"/v2/execute"' \
  tests/
```

Runtime options configure adapters only. Unsupported names, wrong-adapter options, invalid types, and invalid values fail before external setup. An omitted or empty runtime target, or `builtin`, selects the built-in runtime. Unknown targets and schemes fail; a bare `wire` or `tcp://` target does not select Wire.

| Adapter | Option | JSON value |
| --- | --- | --- |
| Built-in | None | Adapter options are rejected. |
| HTTP/HTTPS | `headers` | Object of strings, added to request headers. |
| HTTP/HTTPS | `cookies` | Object of strings, sent as request cookies. |
| HTTP/HTTPS | `path` | String URL reference overriding execution only; empty retains the runtime URL path. |
| Wire TCP | `connect-timeout` | Positive duration string; default `"5s"`, connection and handshake only. |

Values use the existing `name:<JSON value>` syntax. Quote the entire shell argument to preserve JSON strings, spaces, colons, and punctuation. Repeat the flag for multiple options, or separate entries with commas outside JSON objects, arrays, and strings. `LAB_RUNTIME_OPTION` accepts the same comma-separated entries. Any explicit `--runtime-option` replaces the environment option list. Later duplicate keys replace the whole previous value, including objects; they do not merge.

FQL inputs use the existing parameter pipeline across all adapters:

```bash
lab run tests/
lab run --param='value:"shared"' tests/
lab version

lab version \
  --runtime=https://ferret.example.com/api \
  --runtime-option='headers:{"Authorization":"Bearer token123"}'
```

`--param` values use last-key-wins handling, and CLI flags replace values from `LAB_PARAM`. Bindings resolve after local service setup from one parameter snapshot; duplicate, overlapping, or reserved targets fail. YAML query and assertion parameters override the common user values independently. Assertion `@lab.data.query` retains the query result and effective query parameters. `@lab` remains reserved for Lab system values. Ordinary FQL parameters named `headers` or `path` do not configure HTTP, and adapter options are never injected as FQL inputs.

#### Runtime configuration migration

This intentionally replaces the earlier v2 runtime interface. Removed flags fail argument validation, and obsolete environment variables no longer configure Lab.

| Previous usage | Replacement |
| --- | --- |
| `--runtime=wire --runtime-endpoint=tcp://host:port` | `--runtime=wire+tcp://host:port` (existing loopback restriction applies) |
| `--runtime-connect-timeout=10s` | `--runtime-option='connect-timeout:"10s"'` |
| `--runtime-param='headers:{...}'` | `--runtime-option='headers:{...}'` |
| `--runtime-param='path:"/v1/execute"'` | `--runtime-option='path:"/v1/execute"'` |
| FQL input through `--runtime-param` | Existing `--param`, bindings, or suite/query parameter facilities |
| `LAB_RUNTIME_PARAM` containing adapter settings | `LAB_RUNTIME_OPTION` |
| `LAB_RUNTIME_ENDPOINT` | Include the endpoint in `LAB_RUNTIME=wire+tcp://127.0.0.1:<port>` |
| `LAB_RUNTIME_CONNECT_TIMEOUT=10s` | `LAB_RUNTIME_OPTION='connect-timeout:"10s"'` |
| `--rp` | Removed; use `--runtime-option` for adapter configuration |

The separate runtime FQL defaults layer has been removed. A configuration combining runtime defaults and per-run values may need to be expressed using existing CLI, binding, or query/assertion parameter facilities; there is no guaranteed lossless one-for-one migration of that layer. No new defaults layer is introduced. External CLI binary execution remains unsupported.

## Architecture

### 🏗️ System Overview

Lab is built with a modular architecture that separates source resolution, test orchestration, runtime execution, local static and mock serving, and reporting.

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Test Sources  │    │   Test Runner   │    │     Ferret      │
│                 │    │                 │    │     Runtime     │
│ • File System   │───▶│ • Orchestration │───▶│                 │
│ • Git Repos     │    │ • Parallelism   │    │ • Built-in      │
│ • HTTP URLs     │    │ • Retry Logic   │    │ • Remote HTTP   │
└─────────────────┘    │ • Reporting     │    │ • Wire / gRPC   │
                       └─────────────────┘    └─────────────────┘
                                │
                ┌───────────────┴───────────────┐
                ▼                               ▼
       ┌─────────────────┐             ┌─────────────────┐
       │  Static Server  │             │ API Mock Server │
       │                 │             │                 │
       │ • Served Dirs   │             │ • OpenAPI Specs │
       │ • @lab.static   │             │ • x-lab-mock    │
       │ • Auto Ports    │             │ • @lab.mock     │
       └─────────────────┘             └─────────────────┘
```

### 📦 Core Components

#### Sources (`pkg/sources/`)

Handles fetching test files from various locations:

- **FileSystem Source** - Local directory and file access with glob pattern support
- **Git Source** - Clone and fetch files from Git repositories over HTTP/HTTPS
- **HTTP Source** - Download scripts from web URLs
- **Aggregate Source** - Combine multiple source types

#### Runtime (`pkg/runtime/`)

Manages Ferret script execution:

- **Built-in Runtime** - Owns Ferret through `uapi.New`, preserving embedded version injection and execution policies
- **Remote Runtime** - Communicates with remote Ferret services over HTTP
- **Wire Runtime** - Uses the hosted Universal Ferret API over gRPC

#### Test Runner (`pkg/runner/`)

Orchestrates test execution:

- **Parallel Processing** - Manages concurrent test execution
- **Retry Mechanism** - Handles failed test retries
- **Resource Management** - Controls timeouts and resource allocation
- **Lifecycle Management** - Owns each case executor across repetitions and retries, closing it before reporting the result

#### Static Server (`pkg/staticserver/`)

Built-in HTTP server for local static assets:

- **Multiple endpoints** - Serve multiple directories simultaneously
- **Dynamic ports** - Allocate free ports automatically
- **Alias support** - Assign stable names to served directories

#### Mock Server (`pkg/mockserver/`)

OpenAPI-compatible mock REST API server:

- **Spec-driven routes** - Register endpoints from `x-lab-mock` operations
- **Template responses** - Render body values and raw templates from request context
- **Local lifecycle** - Reuses shared local server binding and endpoint management

#### Local Server (`pkg/localserver/`)

Shared local HTTP server runtime:

- **Host settings** - Normalizes bind and advertised hosts
- **Entry parsing** - Handles path, alias, and port bindings
- **Lifecycle management** - Starts, stops, and reports endpoint URLs

#### Reporters (`pkg/reporters/`)

Output formatting and result presentation:

- **Console Reporter** - Rich output for interactive use
- **Simple Reporter** - Plain text output suitable for CI/CD

#### Testing Framework (`pkg/testing/`)

Test suite definition and validation:

- **YAML Parser** - Parse test suite definitions
- **Parameter Injection** - Handle user/system parameters and setup-time parameter rebinding
- **Execution Semantics** - Run units, query/assertion suites, and expected-error tests
- **Case Executor** - Own output settlement, fresh sessions, and optional compiled-plan reuse

### 🔄 Execution Flow

1. **Input Processing** - Parse locations, parameters, bindings, and runtime options
2. **Dependency Readiness** - Wait for configured HTTP/HTTPS endpoints
3. **Runtime Setup** - Initialize the built-in, HTTP, or Wire adapter
4. **Source Setup** - Construct sources and runner settings
5. **Local Services** - Start static/mock services, publish URLs, and resolve parameter bindings
6. **Test Execution** - Discover cases and run units or suite query/assertion scripts with configured concurrency, timeouts, repetitions, and retries
7. **Reporting** - Present progress results and the final summary
8. **Cleanup** - Close case executors before their results, then release command-owned services and the runtime

### 🎯 Design Principles

- **Modularity** - Each component has a focused responsibility
- **Extensibility** - New source types, runtimes, and reporters can be added without changing the whole system
- **Performance** - Parallel execution and resource-conscious orchestration
- **Reliability** - Retry support, wait conditions, and explicit timeouts
- **Flexibility** - Works with local, remote, and containerized Ferret runtimes

## Development

### 🛠️ Building from Source

**Prerequisites:**

- The Go toolchain specified in [go.mod](go.mod)
- Git
- Make for the workflow commands below

**Build Steps:**

```bash
# Clone the repository
git clone https://github.com/MontFerret/lab.git
cd lab

# Install development tools
make install-tools

# Build the project
make build

# Compile without running vet and tests
make compile
```

**Development Workflow:**

```bash
# Run tests
make test
# Or:
go test ./...

# Format code
make fmt

# Lint code
make lint

# Run vet, tests, and compilation (lint and formatting are separate)
make build
```

### 🧪 Testing Lab Itself

```bash
# Run unit tests
go test -v ./...

# Run specific package tests
go test -v ./pkg/sources/...
go test -v ./pkg/runtime/...

# Run tests with coverage locally
go test -race -coverprofile=coverage.txt -covermode=atomic ./...
```

### 🏗️ Project Structure

```text
lab/
├── main.go              # Application entry point
├── cmd/                 # CLI commands, flags, and validation
├── pkg/
│   ├── localserver/     # Shared local HTTP server lifecycle
│   ├── mockserver/      # OpenAPI-compatible mock APIs
│   ├── staticserver/    # Static file serving
│   ├── reporters/       # Result presentation
│   ├── runner/          # Scheduling, repetitions, and retries
│   ├── runtime/         # Built-in, HTTP, and Wire adapters
│   ├── sources/         # Test source loading
│   └── testing/         # Test semantics, parameters, and case executors
├── internal/testutil/   # Controlled runtime and Wire-host fixtures
├── docs/development/    # Contributor and subsystem guides
├── assets/              # Documentation assets
├── scripts/             # Build and release support
├── Dockerfile           # Source-based container build
├── Dockerfile.release   # Release-binary container build
├── Makefile             # Development workflow
└── README.md
```

### 📚 Adding New Features

#### New Source Type

1. Implement the `Source` interface in `pkg/sources/`.
2. Add URL scheme handling in `pkg/sources/source.go`.
3. Add tests in `pkg/sources/`.

#### New Runtime

1. Implement the Universal API's `api.Runtime` in `pkg/runtime/`, including explicit unsupported-operation errors where the integration lacks a capability.
2. Add runtime type detection in `pkg/runtime/runtime.go`.
3. Add configuration handling.

#### New Reporter

1. Implement the `Reporter` interface in `pkg/reporters/`.
2. Register it in `pkg/reporters/reporter.go` and update CLI help when user-selectable.
3. Add output format tests.

## Best Practices

### 📋 Test Organization

#### Directory Structure

```text
project/
├── tests/                 # Run with lab run tests/
│   ├── unit/              # Individual components
│   │   ├── api/
│   │   └── ui/
│   ├── integration/       # Integration tests
│   │   ├── user-flows/
│   │   └── data-validation/
│   ├── e2e/               # End-to-end tests
│   │   ├── critical-path/
│   │   └── smoke/
│   └── fixtures/          # Non-test assets such as JSON and HTML
│       ├── pages/
│       └── data/
└── scripts/               # Referenced FQL helpers, outside test discovery
    ├── common/
    └── helpers/
```

#### Naming Conventions

- Use descriptive test names, such as `user-registration-flow.yaml`.
- Prefix test types when useful, such as `smoke-`, `regression-`, or `load-`.
- Use kebab-case for files, such as `checkout-process.fql`.

#### Test Suite Best Practices

Keep each suite focused, use explicit assertions, and set a timeout appropriate to the work:

```yaml
timeout: 10
query:
  text: |
    RETURN { status: "ready", items: ["keyboard", "dock"] }

assert:
  text: |
    LET result = @lab.data.query.result
    T::EQ(result.status, "ready", "inventory should be ready")
    RETURN T::LEN(result.items, 2)
```

Keep helper queries outside directories you run as test collections, or select tests with a source filter. Lab discovers supported files recursively, including referenced `.fql` files if they are under the selected directory. Manage external fixture setup and teardown outside the suite manifest.

### ⚡ Performance Optimization

#### Concurrency Guidelines

```bash
# Local development: low concurrency
lab run --concurrency=2 tests/

# CI environments: medium concurrency
lab run --concurrency=4 tests/

# Dedicated test infrastructure: higher concurrency
lab run --concurrency=8 tests/
```

#### Resource Management

- Use appropriate timeouts for different test types.
- Manage external fixture setup and teardown in the surrounding test environment.
- Monitor memory usage with large test suites.
- Use the static server for shared static assets.

#### Test Efficiency

```bash
# Run faster tests first
lab run tests/smoke/ && lab run tests/integration/ && lab run tests/e2e/

# Use separate paths for different test categories
lab run --timeout=60 tests/critical/
lab run --timeout=300 --concurrency=1 tests/extended/
```

### 🔒 Security Considerations

- Keep credentials out of committed test files.
- Supply secret parameter bindings through your CI secret store using `LAB_PARAM`.
- Use JSON-encoded values, such as `apiKey:"..."`, with the same syntax as `--param`.
- Keep runtime authentication in `LAB_RUNTIME_OPTION` or `--runtime-option`, separate from FQL inputs.
- Avoid logging secret parameters or sensitive query results.

```bash
# LAB_PARAM and LAB_RUNTIME_OPTION can be injected by your CI secret store.
lab run --reporter=simple tests/
```

## Troubleshooting

### 🐛 Common Issues

#### Chrome/CDP Connection Issues

Use an external runtime that provides the browser module your scripts require. Configure the browser endpoint in that host's Ferret/module configuration; there is no Lab CDP flag.

Check the browser endpoint from the environment that must reach it:

```bash
curl http://127.0.0.1:9222/json/version
```

For an HTTP runtime in a container, `127.0.0.1` in its browser configuration refers to that container. Make Lab fixture URLs reachable from the host with `--serve-bind` and `--serve-host` when needed. See [Using Chrome DevTools Protocol](#-using-chrome-devtools-protocol).

#### Test Timeouts

Increase the per-attempt timeout when the work requires it:

```bash
lab run --timeout=60 tests/
```

A suite may set its own top-level `timeout`. Query execution, output consumption, and assertions share that attempt's deadline. For browser-specific waits and timeouts, use the options supported by the module on your external host.

#### Git Source Issues

Use a complete repository URL with the `git+http` or `git+https` scheme. Lab does not support `git+ssh`, inline revision selectors, or a `//tests/` suffix.

For private repositories or a chosen revision, prepare a checkout with Git's normal authentication facilities, then run the relevant local directory. Set `TEST_REVISION` to the branch, tag, or commit you need before running these commands:

```bash
git clone https://github.com/your-org/test-repo.git test-repo
cd test-repo
git checkout "${TEST_REVISION}"
lab run tests/
```

Clone and checkout errors belong to Git; Lab source errors retain the relevant location or filename.

#### Static Server Port Conflicts

If binding a configured port fails with an address-in-use error, choose another port or stop the process already listening there.

**Solutions:**

1. Lab normally finds free ports automatically, but you can specify a port if needed:

   ```bash
   lab run --serve ./static@app:8081 tests/
   ```

2. Check for port conflicts:

   ```bash
   lsof -nP -iTCP:8080 -sTCP:LISTEN
   ```

### 📊 Performance Issues

#### High Memory Usage

- Reduce concurrency with `--concurrency=2`.
- Release resources owned by your runtime host and surrounding test environment.
- Use a separately hosted HTTP or Wire runtime for memory-intensive tests.

#### Slow Test Execution

- Enable parallel execution with `--concurrency=4`.
- Use the static server for local assets.
- Optimize FQL scripts.
- Profile tests to identify bottlenecks.

### 🔍 Debugging Tips

#### Verbose Output

Use the simple reporter to see file identity, duration, attempt count, and final errors without interactive console formatting:

```bash
lab run --reporter=simple tests/
```

There is no Lab `LOG_LEVEL` environment setting. Configure additional execution logging in the selected runtime host when it supports it.

#### Test Individual Scripts

```bash
# Test one file at a time
lab run specific-test.fql

# Run with retries disabled
lab run --attempts=1 problematic-test.fql
```

#### Validate Test Syntax

Run a small core-only test to check the local setup:

```bash
printf 'RETURN T::EQ(1 + 1, 2)\n' > check.fql
lab run check.fql
```

Lab compiles and executes the script; it is not a syntax-only checker. When a function is unavailable, verify that the selected runtime includes the required module.

## Contributing

### 🤝 How to Contribute

We welcome contributions to Lab. To get started:

1. Fork the repository on GitHub.
2. Create a feature branch: `git checkout -b feature/awesome-feature`.
3. Make your changes and add tests.
4. Run the test suite: `make test`.
5. Commit your changes: `git commit -am 'Add awesome feature'`.
6. Push to the branch: `git push origin feature/awesome-feature`.
7. Submit a pull request.

### 📝 Development Guidelines

- Write tests for new features.
- Follow Go conventions and formatting with `make fmt`.
- Pass linting checks with `make lint`.
- Update documentation for user-facing changes.
- Keep commits atomic and write clear commit messages.

### 🐛 Reporting Issues

When reporting bugs, please include:

- Lab version from `lab version`
- Operating system and version
- Go version, if building from source
- Complete command that failed
- Full error message and stack trace
- Minimal reproduction case

### 💡 Feature Requests

Before requesting features:

- Check existing issues and discussions.
- Describe the use case and problem being solved.
- Consider whether it fits Lab's scope.
- Share examples of the desired behavior when possible.

### 🔄 Development Process

1. **Discussion** - Major features should be discussed in issues first.
2. **Implementation** - Write code with tests and documentation.
3. **Review** - Submit a pull request for code review.
4. **Testing** - Ensure all CI checks pass.
5. **Merge** - A maintainer will merge when ready.

## License

Lab is licensed under the [Apache License 2.0](LICENSE).

---

**Happy Testing!** 🚀

For more information about Ferret and FQL, visit the [Ferret documentation](https://www.montferret.dev/docs/).

Join the community on [Discord](https://discord.gg/kzet32U) for support and discussions.
