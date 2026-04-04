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

**Lab** is a powerful, flexible test runner designed specifically for [Ferret](https://www.github.com/MontFerret/ferret) scripts. It enables automated testing of web scraping, browser automation, and API testing scenarios using Ferret Query Language (FQL).

> **📢 Notice:** This branch contains the upcoming **Lab for Ferret v2**. For the stable v1 release, please visit [CLI v1](https://github.com/MontFerret/lab/tree/v1).

---

**🚀 Perfect for:**
- End-to-end web application testing
- Web scraping validation and monitoring  
- API integration testing
- Browser automation testing
- Regression testing for web applications

Read the introductory blog post about Lab [here!](https://www.montferret.dev/blog/say-hello-to-lab/)

<p align="center">
<img alt="lab" src="https://raw.githubusercontent.com/MontFerret/lab/main/assets/landing.png" style="margin-left: auto; margin-right: auto;" width="495px" height="501px" />
</p>

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Test Suites](#test-suites)
- [Advanced Usage](#advanced-usage)
- [Configuration Reference](#configuration-reference)
- [Architecture](#architecture)
- [Development](#development)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Features

### 🏃‍♂️ **Performance & Scalability**
- **Parallel execution** - Run multiple tests concurrently for faster feedback
- **Configurable concurrency** - Control the number of simultaneous test executions
- **Test retry mechanism** - Automatic retry of failed tests with customizable attempts
- **Batch execution** - Run tests multiple times with configurable intervals

### 🌐 **Flexible Runtime Support**
- **Built-in Ferret runtime** - Execute tests using embedded Ferret engine
- **Remote HTTP runtime** - Connect to remote Ferret services via HTTP/HTTPS
- **External binary runtime** - Use custom Ferret CLI installations
- **Multi-runtime testing** - Test against different Ferret versions or configurations

### 📁 **Multiple Source Types**
- **Local filesystem** - Execute scripts from local directories
- **Git repositories** - Fetch and run tests directly from Git repos (HTTP/HTTPS)
- **HTTP sources** - Download and execute scripts from web URLs
- **Glob pattern matching** - Select multiple files using wildcard patterns

### 🌍 **Static Content Serving**
- **Built-in HTTP server** - Serve static files for testing web applications
- **Dedicated `serve` command** - Run the static file server without executing tests
- **Multiple static endpoints** - Host different directories at separate URLs
- **Custom aliases** - Name your content endpoints for better organization
- **Dynamic port allocation** - Automatically find available ports

### 📊 **Rich Reporting & Monitoring**
- **Multiple output formats** - Console and simple reporters available
- **Detailed test results** - Comprehensive execution metrics and timing
- **Wait conditions** - Test and wait for external services to be available
- **Environment variable support** - Configure tests via environment variables

## Installation

### 📦 **Binary Downloads**
Download the latest pre-built binaries from our [releases page](https://github.com/MontFerret/lab/releases).

**Linux:**
```bash
curl -L https://github.com/MontFerret/lab/releases/latest/download/lab-linux-amd64.tar.gz | tar xz
sudo mv lab /usr/local/bin/
```

**macOS:**
```bash
curl -L https://github.com/MontFerret/lab/releases/latest/download/lab-darwin-amd64.tar.gz | tar xz
sudo mv lab /usr/local/bin/
```

**Windows:**
Download the `.zip` file from releases and extract `lab.exe` to your PATH.

### 🚀 **One-line Install Script**
The easiest way to install Lab on Unix-like systems:

```bash
curl -fsSL https://raw.githubusercontent.com/MontFerret/lab/main/install.sh | sh
```

This script automatically:
- Detects your operating system and architecture
- Downloads the appropriate binary
- Installs it to `/usr/local/bin/`
- Makes it executable

### 🐳 **Docker**
Run Lab in a container without installing it locally:

```bash
# Pull the latest image
docker pull montferret/lab:latest

# Run a simple test
docker run --rm -v $(pwd):/workspace montferret/lab:latest run /workspace/tests/

# With custom options
docker run --rm -v $(pwd):/workspace montferret/lab:latest \
    run --concurrency=4 --reporter=simple /workspace/tests/
```

The container entrypoint is Lab-aware, so `docker run <image> run ...` and `docker run <image> version` invoke the Lab binary directly, while raw commands such as `docker run <image> /bin/sh -c 'echo ok'` still pass through to the container shell.

**Docker Compose Example:**
```yaml
version: '3.8'
services:
  lab:
    image: montferret/lab:latest
    volumes:
      - ./tests:/workspace/tests
      - ./static:/workspace/static
    command: ["run", "--serve", "/workspace/static", "/workspace/tests/"]
```

### 🛠️ **Build from Source**
For development or custom builds:

```bash
# Prerequisites: Go 1.23+ required
git clone https://github.com/MontFerret/lab.git
cd lab
go build -o lab .

# Or use the Makefile
make build
```

### ✅ **Verify Installation**
```bash
lab version
lab --help
```

## Quick Start

### 🎯 **Basic Usage**

The simplest way to run Ferret scripts with Lab:

`lab run` is required for script execution. Root invocation such as `lab tests/` or `lab -f tests/` is not supported.

```bash
# Execute a single FQL script
lab run myscript.fql

# Run all FQL scripts in a directory
lab run myscripts/

# Run with increased concurrency
lab run --concurrency=4 myscripts/

# Run tests multiple times
lab run --times=3 myscript.fql
```

Use `lab version --runtime=...` when you want to inspect the version reported by a specific remote or binary runtime.

### 📝 **Your First Test**

Create a simple test file `example.fql`:

```sql
LET doc = DOCUMENT("https://www.github.com", { 
    driver: "cdp",
    userAgent: "Lab Test Runner" 
})

// Wait for page to load
WAIT_ELEMENT(doc, "header")

// Extract page title
LET title = doc.title

// Return result
RETURN {
    url: doc.url,
    title: title,
    hasGitHubLogo: ELEMENT_EXISTS(doc, "[aria-label*='GitHub']")
}
```

Run it:
```bash
lab run example.fql
```

### 🎨 **Using Chrome DevTools Protocol**

For browser automation, you'll need a Chrome/Chromium instance running in headless mode:

```bash
# Start Chrome in headless mode (separate terminal)
google-chrome --headless --remote-debugging-port=9222

# Run your tests (default CDP address)
lab run --cdp=http://127.0.0.1:9222 browser-tests/

# Or use a custom CDP address
lab run --cdp=http://localhost:9223 browser-tests/
```

### 📊 **Example Output**

```
$ lab run example.fql
✓ example.fql (1.23s)
  └─ Assertions: 1 passed, 0 failed

Tests: 1 passed, 0 failed
Time:  1.23s
```

## Test Suites

Lab supports sophisticated test suites defined in YAML format, enabling you to create complex testing scenarios with assertions, parameters, and reusable components.

### 📋 **Basic Test Suite Structure**

```yaml
query:
  text: |
    LET doc = DOCUMENT("https://github.com/", { driver: "cdp" })
    
    HOVER(doc, ".HeaderMenu-details")
    CLICK(doc, ".HeaderMenu a")
    
    WAIT_NAVIGATION(doc)
    WAIT_ELEMENT(doc, '.IconNav')
    
    FOR el IN ELEMENTS(doc, '.IconNav a')
        RETURN TRIM(el.innerText)

assert:
  text: RETURN T::NOT::EMPTY(@lab.data.query.result)
```

Save as `github-test.yaml` and run:
```bash
lab run github-test.yaml
```

### 🔗 **Reference External Scripts**

Keep your FQL scripts separate and reference them in test suites:

**navigation.fql:**
```sql
LET doc = DOCUMENT(@url, { driver: "cdp" })
WAIT_ELEMENT(doc, "body")
RETURN doc.title
```

**suite.yaml:**
```yaml
query:
  ref: ./scripts/navigation.fql
  params:
    url: "https://example.com"

assert:
  text: |
    RETURN T::NOT::EMPTY(@lab.data.query.result) 
           AND T::CONTAINS(@lab.data.query.result, "Example")
```

### 🧪 **Complex Test Scenarios**

```yaml
name: "E-commerce User Journey"
description: "Test complete user purchase flow"

setup:
  text: |
    LET baseUrl = "https://demo-shop.example.com"
    RETURN { baseUrl }

query:
  text: |
    LET doc = DOCUMENT(@lab.data.setup.result.baseUrl, { driver: "cdp" })
    
    // Navigate to product
    CLICK(doc, ".product-item:first-child a")
    WAIT_NAVIGATION(doc)
    
    // Add to cart
    CLICK(doc, ".add-to-cart")
    WAIT_ELEMENT(doc, ".cart-confirmation")
    
    // Go to checkout
    CLICK(doc, ".checkout-btn")
    WAIT_NAVIGATION(doc)
    
    RETURN {
      currentUrl: doc.url,
      cartItems: LENGTH(ELEMENTS(doc, ".cart-item")),
      totalPrice: INNER_TEXT(doc, ".total-price")
    }

assert:
  text: |
    LET result = @lab.data.query.result
    RETURN T::CONTAINS(result.currentUrl, "checkout") 
           AND result.cartItems > 0
           AND T::NOT::EMPTY(result.totalPrice)

cleanup:
  text: |
    // Clear cart or perform cleanup
    RETURN "Cleanup completed"
```

### 🎯 **Parameterized Tests**

Create reusable test suites with parameters:

```yaml
query:
  text: |
    LET doc = DOCUMENT(@testUrl, { 
      driver: "cdp",
      timeout: @pageTimeout 
    })
    
    WAIT_ELEMENT(doc, @selector)
    
    RETURN {
      title: doc.title,
      elementExists: ELEMENT_EXISTS(doc, @selector)
    }

assert:
  text: |
    LET result = @lab.data.query.result
    RETURN result.elementExists == true
```

Run with parameters:
```bash
lab run --param=testUrl:"https://example.com" \
    --param=pageTimeout:5000 \
    --param=selector:"h1" \
    test-suite.yaml
```

### 📊 **Data-Driven Testing**

Use external data sources for comprehensive testing:

```yaml
query:
  text: |
    LET testData = [
      { url: "https://site1.com", expectedTitle: "Site 1" },
      { url: "https://site2.com", expectedTitle: "Site 2" }
    ]
    
    FOR test IN testData
      LET doc = DOCUMENT(test.url, { driver: "cdp" })
      WAIT_ELEMENT(doc, "title")
      
      RETURN {
        url: test.url,
        expectedTitle: test.expectedTitle,
        actualTitle: doc.title,
        matches: doc.title == test.expectedTitle
      }

assert:
  text: |
    FOR result IN @lab.data.query.result
      FILTER result.matches != true
      RETURN false
    
    RETURN true
```

## Advanced Usage

### 📁 **File Resolution**

Lab supports multiple source locations for maximum flexibility:

#### **Local Files**
```bash
# Single file
lab run /path/to/test.fql

# Directory with glob patterns
lab run "tests/**/*.fql"
lab run tests/integration/

# Multiple paths
lab run --files=tests/unit/ --files=tests/integration/ --files=scripts/smoke.fql
```

#### **Git Repositories**
Fetch and execute tests directly from Git repositories:

```bash
# HTTPS Git repository
lab run git+https://github.com/username/test-repo.git//tests/

# HTTP Git repository  
lab run git+http://git.example.com/tests.git//integration/

# Specific branch or tag
lab run git+https://github.com/username/tests.git@v1.2.0//suite.yaml

# Private repositories (requires authentication)
lab run git+https://username:token@github.com/private/repo.git//tests/
```

#### **HTTP Sources**
Download scripts from web URLs:

```bash
# Direct script URL
lab run https://raw.githubusercontent.com/user/repo/main/test.fql

# Multiple HTTP sources
lab run https://example.com/tests/suite1.yaml https://example.com/tests/suite2.yaml
```

### 🌐 **Static File Serving**

Lab can serve local directories over HTTP during test execution. Served endpoints are available in FQL scripts under `@lab.static.<alias>`.

#### **Standalone Static Server**
Use `lab serve` when you want to expose local directories over HTTP without running a test suite. Positional entries are the primary syntax, and repeated `--serve` flags are also supported for symmetry with `lab run`.

```bash
# Serve a single directory
lab serve ./website

# Serve multiple directories with aliases
lab serve ./frontend@app ./mockdata@api
```

#### **Basic Static Serving**
```bash
# Serve files from ./website directory
lab run --serve ./website tests/

# Access in your FQL scripts
LET doc = DOCUMENT(@lab.static.website, { driver: "cdp" })
```

#### **Multiple Static Endpoints**
```bash
# Serve multiple directories
lab run --serve ./app --serve ./api-mocks tests/
```

FQL Script:
```sql
// Access different endpoints
LET appPage = DOCUMENT(@lab.static.app, { driver: "cdp" })
LET apiData = IO::NET::HTTP::GET(@lab.static["api-mocks"] + "/users.json")
```

#### **Custom Static Aliases**
```bash
# Give custom names to your content
lab run --serve ./frontend@app --serve ./mockdata@api tests/
```

FQL Script:
```sql
// Use custom aliases
LET homePage = DOCUMENT(@lab.static.app + "/index.html", { driver: "cdp" })
LET userData = IO::NET::HTTP::GET(@lab.static.api + "/user/123.json")
```

#### **Advanced Static Server Example**
```bash
# Complex setup with multiple content sources
lab run \
  --serve ./dist@webapp \
  --serve ./test-fixtures@fixtures \
  --serve ./mock-apis@mocks \
  --concurrency=3 \
  tests/e2e/
```

#### **Reachable Static URLs for Remote Runtimes**
When Lab runs against a remote runtime or inside a containerized environment, use `--serve-host` to advertise a client-reachable host instead of loopback. Use `--serve-bind` when you need to control the listener interface explicitly.

```bash
lab run \
  --runtime=https://ferret.example.com \
  --serve ./dist@app \
  --serve-host host.docker.internal \
  --serve-bind 0.0.0.0 \
  tests/
```

If `--serve-host` is set without `--serve-bind`, Lab automatically binds static servers to all interfaces (`0.0.0.0` for IPv4/hostnames, `::` for IPv6 literals).

### 🔄 **Remote Ferret Runtime**

Lab can execute tests against remote Ferret instances instead of using the built-in runtime:

#### **HTTP/HTTPS Runtime**
```bash
# Connect to remote Ferret service
lab run --runtime=https://ferret.example.com/api tests/

# With custom headers and path
lab run \
  --runtime=https://ferret.example.com \
  --runtime-param=headers:'{"Authorization": "Bearer token123"}' \
  --runtime-param=path:"/v1/execute" \
  tests/
```

The HTTP runtime sends POST requests with:
```json
{
  "query": "FQL script content",
  "params": {
    "key": "value"
  }
}
```

#### **External Binary Runtime**
Use custom Ferret CLI installations:

```bash
# Use specific Ferret binary
lab run --runtime=bin:./custom-ferret tests/

# With additional parameters
lab run \
  --runtime=bin:/usr/local/bin/ferret-v0.18 \
  --runtime-param=timeout:30 \
  tests/
```

#### **Runtime Comparison Testing**
Test against multiple runtime versions:

```bash
# Test with built-in runtime
lab run tests/ > builtin-results.txt

# Test with remote runtime
lab run --runtime=https://ferret-v0.17.example.com tests/ > remote-v0.17-results.txt

# Compare results
diff builtin-results.txt remote-v0.17-results.txt
```

### ⚡ **Performance Optimization**

#### **Parallel Execution**
```bash
# Run up to 8 tests simultaneously
lab run --concurrency=8 tests/

# Balance between speed and resource usage
lab run --concurrency=4 --timeout=60 large-test-suite/
```

#### **Test Repetition & Retry**
```bash
# Run each test 3 times for reliability testing
lab run --times=3 tests/flaky/

# Retry failed tests up to 2 additional times
lab run --attempts=3 tests/

# Add delay between test cycles
lab run --times=5 --times-interval=10 stress-tests/
```

#### **Conditional Execution**
```bash
# Wait for services to be available before running tests
lab run \
  --wait=http://127.0.0.1:9222/json/version \
  --wait=postgres://localhost:5432/testdb \
  --wait-timeout=30 \
  tests/integration/
``` 

## Configuration Reference

### 🎛️ **Command Line Flags**

These flags apply to `lab run`.

| Flag | Short | Environment Variable | Default | Description |
|------|-------|---------------------|---------|-------------|
| `--files` | `-f` | `LAB_FILES` | - | Location of FQL script files to run |
| `--timeout` | `-t` | `LAB_TIMEOUT` | `30` | Test timeout in seconds |
| `--cdp` | - | `LAB_CDP` | `http://127.0.0.1:9222` | Chrome DevTools Protocol address |
| `--reporter` | - | `LAB_REPORTER` | `console` | Output reporter (`console`, `simple`) |
| `--runtime` | `-r` | `LAB_RUNTIME` | - | URL to remote Ferret runtime |
| `--runtime-param` | `--rp` | `LAB_RUNTIME_PARAM` | - | Parameters for remote runtime |
| `--concurrency` | `-c` | `LAB_CONCURRENCY` | `1` | Number of parallel test executions |
| `--times` | - | `LAB_TIMES` | `1` | Number of times to run each test |
| `--attempts` | `-a` | `LAB_ATTEMPTS` | `1` | Number of retry attempts for failed tests |
| `--times-interval` | - | `LAB_TIMES_INTERVAL` | `0` | Interval between test cycles (seconds) |
| `--serve` | - | `LAB_SERVE` | - | Served directory mapping exposed over HTTP |
| `--serve-bind` | - | `LAB_SERVE_BIND` | - | Host to bind static servers to (host only, no port) |
| `--serve-host` | - | `LAB_SERVE_HOST` | - | Host to advertise for static server URLs (host only, no port) |
| `--param` | `-p` | `LAB_PARAM` | - | Query parameters for tests |
| `--wait` | `-w` | `LAB_WAIT` | - | Wait for resource availability |
| `--wait-timeout` | `--wt` | `LAB_WAIT_TIMEOUT` | `5` | Wait timeout in seconds |
| `--wait-attempts` | - | `LAB_WAIT_ATTEMPTS` | `5` | Number of wait attempts |

### 🌍 **Environment Variables**

Set environment variables for consistent configuration across environments:

```bash
# Basic configuration
export LAB_TIMEOUT=60
export LAB_CONCURRENCY=4
export LAB_REPORTER=simple

# CDP configuration
export LAB_CDP=http://chrome-headless:9222

# Runtime configuration  
export LAB_RUNTIME=https://ferret-api.example.com
export LAB_RUNTIME_PARAM='headers:{"API-Key":"secret123"}'

# Run tests
lab run tests/
```

### 📝 **Configuration Examples**

#### **CI/CD Configuration**
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
  --wait=postgres://db:5432/testdb \
  --wait-timeout=60 \
  tests/integration/
```

#### **Local Development**
```bash
#!/bin/bash
# dev-test.sh

export LAB_CDP=http://localhost:9222
export LAB_TIMEOUT=30
export LAB_CONCURRENCY=1

# Serve local assets and run tests
lab run \
  --serve ./dist@app \
  --serve ./fixtures@data \
  tests/dev/
```

#### **Load Testing**
```bash
#!/bin/bash
# load-test.sh

# High concurrency for performance testing
lab run \
  --concurrency=20 \
  --times=100 \
  --times-interval=1 \
  --timeout=10 \
  tests/performance/
```

### ⚙️ **Runtime Parameters**

Configure remote Ferret runtime behavior:

```bash
# HTTP runtime with custom headers
lab run \
  --runtime=https://ferret.api.com \
  --runtime-param='headers:{"Authorization":"Bearer token"}' \
  --runtime-param='path:"/v2/execute"' \
  --runtime-param='timeout:30' \
  tests/

# Binary runtime with custom flags
lab run \
  --runtime=bin:/usr/local/bin/ferret \
  --runtime-param='flags:["--timeout=60", "--verbose"]' \
  tests/
```

## Architecture

### 🏗️ **System Overview**

Lab is built with a modular architecture that separates concerns and enables flexible testing scenarios:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Test Sources  │    │   Test Runner   │    │   Ferret        │
│                 │    │                 │    │   Runtime       │
│ • File System   │───▶│ • Orchestration │───▶│                 │
│ • Git Repos     │    │ • Parallelization│    │ • Built-in      │
│ • HTTP URLs     │    │ • Retry Logic   │    │ • Remote HTTP   │
└─────────────────┘    │ • Reporting     │    │ • External Bin  │
                       └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Static Server  │
                       │                 │
                       │ • Served Dirs   │
                       │ • Static URLs   │
                       │ • Auto Ports    │
                       └─────────────────┘
```

### 📦 **Core Components**

#### **Sources (`sources/`)**
Handles fetching test files from various locations:
- **FileSystem Source**: Local directory and file access with glob pattern support
- **Git Source**: Clone and fetch files from Git repositories (HTTP/HTTPS)
- **HTTP Source**: Download scripts from web URLs
- **Aggregate Source**: Combines multiple source types

#### **Runtime (`runtime/`)**
Manages Ferret script execution:
- **Built-in Runtime**: Uses embedded Ferret engine (default)
- **Remote Runtime**: HTTP-based communication with remote Ferret services
- **Binary Runtime**: Executes external Ferret CLI binaries

#### **Test Runner (`runner/`)**
Orchestrates test execution:
- **Parallel Processing**: Manages concurrent test execution
- **Retry Mechanism**: Handles failed test retries
- **Resource Management**: Controls timeouts and resource allocation
- **Lifecycle Management**: Handles setup, execution, and cleanup phases

#### **Static Server (`staticserver/`)**
Built-in HTTP server for local static assets:
- **Multiple endpoints**: Serve multiple directories simultaneously
- **Dynamic ports**: Automatic port allocation to avoid conflicts
- **Alias support**: Custom names for served directories

#### **Reporters (`reporters/`)**
Output formatting and result presentation:
- **Console Reporter**: Rich, colored output for interactive use
- **Simple Reporter**: Plain text output suitable for CI/CD

#### **Testing Framework (`testing/`)**
Test suite definition and validation:
- **YAML Parser**: Parse test suite definitions
- **Parameter Injection**: Handle runtime parameters and data binding
- **Assertion Engine**: Validate test results

### 🔄 **Execution Flow**

1. **Input Processing**: Parse command-line arguments and environment variables
2. **Source Resolution**: Fetch test files from specified sources
3. **Static Server Initialization**: Start HTTP servers for served directories (if needed)  
4. **Runtime Setup**: Initialize Ferret runtime (built-in or remote)
5. **Test Discovery**: Find and parse test files and suites
6. **Parallel Execution**: Run tests according to concurrency settings
7. **Result Collection**: Gather execution results and timing data
8. **Reporting**: Format and output results via selected reporter
9. **Cleanup**: Stop static file servers and clean up resources

### 🎯 **Design Principles**

- **Modularity**: Each component has a single responsibility
- **Extensibility**: Easy to add new source types, runtimes, or reporters
- **Performance**: Optimized for parallel execution and resource efficiency
- **Reliability**: Built-in retry mechanisms and error handling
- **Flexibility**: Support for various deployment scenarios and configurations

## Development

### 🛠️ **Building from Source**

**Prerequisites:**
- Go 1.23 or later
- Git

**Build Steps:**
```bash
# Clone the repository
git clone https://github.com/MontFerret/lab.git
cd lab

# Install development tools
make install-tools

# Build the project
make build
# Or manually:
go build -o bin/lab -ldflags "-X main.version=dev" ./main.go
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

# Run all checks (vet, test, compile)
make build
```

### 🧪 **Testing Lab Itself**

```bash
# Run unit tests
go test -v ./...

# Run specific test suites
go test -v ./sources/...
go test -v ./runtime/...

# Run tests with coverage
make cover
```

### 🏗️ **Project Structure**

```
lab/
├── main.go              # Application entry point
├── cmd/                 # CLI command implementations
├── staticserver/        # Static file server
├── reporters/           # Output formatters
├── runner/              # Test execution orchestration  
├── runtime/             # Ferret runtime implementations
├── sources/             # Test file source handlers
├── testing/             # Test suite definitions
├── assets/              # Documentation assets
├── Dockerfile          # Container build definition
├── Makefile            # Build automation
└── README.md           # This file
```

### 📚 **Adding New Features**

#### **New Source Type**
1. Implement the `Source` interface in `sources/`
2. Add URL scheme handling in `sources/source.go`
3. Add tests in `sources/`

#### **New Runtime**
1. Implement the `Runtime` interface in `runtime/`
2. Add runtime type detection in `runtime/runtime.go`
3. Add configuration handling

#### **New Reporter**
1. Implement the `Reporter` interface in `reporters/`
2. Register the reporter in CLI flags
3. Add output format tests

## Best Practices

### 📋 **Test Organization**

#### **Directory Structure**
```
tests/
├── unit/              # Unit tests for individual components
│   ├── api/
│   └── ui/
├── integration/       # Integration tests
│   ├── user-flows/
│   └── data-validation/
├── e2e/              # End-to-end tests
│   ├── critical-path/
│   └── smoke/
├── fixtures/         # Test data and assets
│   ├── pages/
│   └── data/
└── scripts/          # Reusable FQL scripts
    ├── common/
    └── helpers/
```

#### **Naming Conventions**
- Use descriptive test names: `user-registration-flow.yaml`
- Prefix test types: `smoke-`, `regression-`, `load-`
- Use kebab-case for files: `checkout-process.fql`

#### **Test Suite Best Practices**
```yaml
# Good: Descriptive names and clear structure
name: "User Authentication Flow"
description: "Verify user login, logout, and session management"

setup:
  text: |
    // Clear any existing sessions
    // Set up test data
    
query:
  text: |
    // Main test logic with clear comments
    
assert:
  text: |
    // Specific, meaningful assertions
    
cleanup:
  text: |
    // Clean up test data
```

### ⚡ **Performance Optimization**

#### **Concurrency Guidelines**
```bash
# Local development: Low concurrency
lab run --concurrency=2 tests/

# CI environments: Medium concurrency  
lab run --concurrency=4 tests/

# Dedicated test infrastructure: High concurrency
lab run --concurrency=8 tests/
```

#### **Resource Management**
- Use appropriate timeouts for different test types
- Implement proper cleanup in test suites
- Monitor memory usage with large test suites
- Use the static server for shared static assets

#### **Test Efficiency**
```bash
# Run faster tests first
lab run tests/smoke/ && lab run tests/integration/ && lab run tests/e2e/

# Use tags for test categorization
lab run --timeout=60 tests/critical/
lab run --timeout=300 --concurrency=1 tests/extended/
```

### 🔒 **Security Considerations**

- Never commit sensitive data in test files
- Use environment variables for credentials
- Sanitize test outputs that might contain secrets
- Use separate test environments for security testing

```bash
# Good: Use environment variables
export TEST_API_KEY="your-key-here"
lab run --param=apiKey:$TEST_API_KEY tests/

# Bad: Hardcode in scripts
# Don't do this: LET apiKey = "secret-key-123"
```

## Troubleshooting

### 🐛 **Common Issues**

#### **Chrome/CDP Connection Issues**
```
Error: Failed to connect to CDP at http://127.0.0.1:9222
```

**Solutions:**
1. **Start Chrome in headless mode:**
   ```bash
   google-chrome --headless --remote-debugging-port=9222 --no-sandbox
   ```

2. **Check if Chrome is running:**
   ```bash
   curl http://127.0.0.1:9222/json/version
   ```

3. **Use custom CDP address:**
   ```bash
   lab run --cdp=http://localhost:9223 tests/
   ```

#### **Test Timeouts**
```
Error: Test timed out after 30 seconds
```

**Solutions:**
1. **Increase timeout:**
   ```bash
   lab run --timeout=60 tests/
   ```

2. **Optimize test scripts:**
   ```sql
   -- Add explicit waits
   WAIT_ELEMENT(doc, ".loading", { displayed: false })
   
   -- Use shorter timeouts for quick checks
   WAIT_ELEMENT(doc, ".button", { timeout: 5000 })
   ```

#### **Git Source Issues**
```
Error: Failed to clone repository
```

**Solutions:**
1. **Check repository URL:**
   ```bash
   git clone https://github.com/user/repo.git  # Test manually
   ```

2. **Authentication for private repos:**
   ```bash
   lab run git+https://username:token@github.com/private/repo.git//tests/
   ```

3. **Use SSH for private repos:**
   ```bash
   # Set up SSH keys, then:
   lab run git+ssh://git@github.com/private/repo.git//tests/
   ```

#### **Static Server Port Conflicts**
```
Error: failed to start static file server on port 8080
```

**Solutions:**
1. **Lab automatically finds free ports**, but you can specify:**
   ```bash
   lab run --serve ./static@app:8081 tests/
   ```

2. **Check for port conflicts:**
   ```bash
   netstat -tlnp | grep :8080
   ```

### 📊 **Performance Issues**

#### **High Memory Usage**
- Reduce concurrency: `--concurrency=2`
- Implement proper cleanup in tests
- Use external binary runtime for memory-intensive tests

#### **Slow Test Execution**
- Enable parallel execution: `--concurrency=4`
- Use the static server for local assets
- Optimize FQL scripts for better performance
- Profile tests to identify bottlenecks

### 🔍 **Debugging Tips**

#### **Verbose Output**
```bash
# Enable detailed logging (if available)
export LOG_LEVEL=debug
lab run tests/

# Use simple reporter for cleaner output
lab run --reporter=simple tests/
```

#### **Test Individual Scripts**
```bash
# Test one file at a time
lab run specific-test.fql

# Run with retries disabled
lab run --attempts=1 problematic-test.fql
```

#### **Validate Test Syntax**
```bash
# Test FQL syntax with Ferret CLI
ferret -q "RETURN 1"  # Should return [1]
```

## Contributing

### 🤝 **How to Contribute**

We welcome contributions to Lab! Here's how to get started:

1. **Fork the repository** on GitHub
2. **Create a feature branch**: `git checkout -b feature/awesome-feature`
3. **Make your changes** and add tests
4. **Run the test suite**: `make test`
5. **Commit your changes**: `git commit -am 'Add awesome feature'`
6. **Push to the branch**: `git push origin feature/awesome-feature`
7. **Submit a pull request**

### 📝 **Development Guidelines**

- **Write tests** for new features
- **Follow Go conventions** and formatting (`make fmt`)
- **Pass all linting checks** (`make lint`)
- **Update documentation** for user-facing changes
- **Keep commits atomic** and write clear commit messages

### 🐛 **Reporting Issues**

When reporting bugs, please include:
- Lab version (`lab version`)
- Operating system and version
- Go version (if building from source)
- Complete command that failed
- Full error message and stack trace
- Minimal reproduction case

### 💡 **Feature Requests**

Before requesting features:
- Check existing issues and discussions
- Describe the use case and problem you're solving
- Consider if it fits Lab's scope and philosophy
- Be prepared to help with implementation

### 🔄 **Development Process**

1. **Discussion**: Major features should be discussed in issues first
2. **Implementation**: Write code with tests and documentation
3. **Review**: Submit PR for code review
4. **Testing**: Ensure all CI checks pass
5. **Merge**: Maintainer will merge when ready

## License

Lab is licensed under the [Apache License 2.0](LICENSE).

---

**Happy Testing!** 🚀

For more information about Ferret and FQL, visit the [Ferret documentation](https://www.montferret.dev/docs/).

Join our community on [Discord](https://discord.gg/kzet32U) for support and discussions.
