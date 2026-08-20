# Local services

Lab can serve local directories and OpenAPI-compatible mock APIs during a test run or through the standalone `serve` command. Shared host, entry, endpoint, and lifecycle behavior belongs in `pkg/localserver`; service-specific behavior belongs in `pkg/staticserver` or `pkg/mockserver`.

The CLI integration is described in [CLI](cli.md).

## Entries and endpoint settings

Static and mock entries use a shared binding model:

```text
<path>
<path>:<port>
<path>@<alias>
<path>@<alias>:<port>
```

`pkg/localserver` parses the shared path, alias, and port structure. Service packages provide their own labels, default aliases, and path validation.

Aliases are stable keys used for endpoint maps and Lab parameters. Duplicate or invalid aliases and invalid port values fail before server startup.

Binding and advertising are separate concerns:

- the bind host controls the listener address
- the advertised host appears in endpoint URLs given to users and tests
- dynamic ports must report the port actually selected by the listener
- endpoint formatting must handle IPv4, IPv6, and host validation consistently

The `@lab.static.<alias>` and `@lab.mock.<alias>` values use advertised endpoint URLs, not raw listener addresses.

## Shared lifecycle

`pkg/localserver.Manager` owns a set of HTTP nodes created through a handler factory. The manager binds entries, starts nodes, reports endpoint URLs, and stops nodes.

Lifecycle guarantees include:

- startup honors context cancellation
- already-running nodes are not started twice
- partial startup failure attempts to stop nodes that did start
- stopping remains safe after partial startup
- normal return, command error, timeout, and cancellation all reach bounded shutdown paths
- serve-loop failures remain observable on the node
- dynamic port selection updates the advertised endpoint to the listener's actual port

Shared parsing, host normalization, endpoint formatting, and lifecycle fixes belong in `pkg/localserver`, not duplicated service packages. Service-specific start/stop error labels may be supplied by the wrapper.

## Static server

`pkg/staticserver` validates directory entries and builds static HTTP handlers. It delegates node and manager lifecycle to `pkg/localserver`.

The handler serves the configured directory, supports an optional prefix for direct node construction, and enables the established CORS behavior for local test assets. Static-server changes should preserve directory validation, aliases, endpoint settings, and cleanup.

Tests for shared entry/settings behavior belong in `pkg/localserver`; directory and handler behavior belongs in `pkg/staticserver`.

## Mock server

`pkg/mockserver` loads YAML or JSON OpenAPI-compatible documents and registers supported operations from the Lab-owned `x-lab-mock` extension. Lab extensions use the `x-lab-*` namespace, not `x-ferret-*`.

Specifications and extensions are validated during loading where practical. Errors should identify the path, operation, response field, or template that caused the failure.

Mock responses can define:

- an HTTP status
- response headers
- a structured body whose string values may contain templates
- a raw body template

Structured bodies and raw body templates are mutually exclusive. Response status, headers, content type, and body must be tested together.

## Route matching

Route selection is deterministic:

- exact static paths are considered before parameterized paths
- parameterized routes are ordered by specificity rather than map iteration
- supported methods map to the operation registered for the selected route
- a matched path with an unsupported method returns method-not-allowed information and accurate allowed methods
- unmatched paths return not found

Changes to route construction or matching should include overlapping static/parameterized routes, competing parameterized shapes, methods, not-found behavior, and allowed-method output. Performance changes to matching require comparable benchmarks.

## Template context and safety

Mock templates use Go `text/template` with a restricted Sprig function map. Request-derived context is explicit:

- request method
- path parameters
- query parameters
- headers
- parsed JSON body

Templates are compiled while loading the specification where practical. Execution errors return a server error without leaking host details.

Host-access helpers such as environment expansion and DNS lookup are removed. New template functions must not add environment, filesystem, DNS, network, process, or other unsafe host capabilities. Template evaluation remains stateless unless a stateful mock feature is explicitly designed.

Random or fake-data helpers must be seedable or otherwise controllable for deterministic tests.

Template and mock response behavior belongs in `pkg/mockserver`. Move it to a shared package only when another Lab-owned subsystem genuinely needs the same contract.

## Testing and performance

Local-service tests should cover:

- entry, alias, port, bind host, and advertised host validation
- IPv4/IPv6 endpoint formatting
- dynamic ports and collisions where controllable
- startup, partial failure, shutdown, timeout, and cancellation
- static directory serving
- malformed mock specifications and extensions
- route specificity, methods, status, headers, bodies, and request context
- template compilation, rendering, and unsafe-function rejection

Changes to server lifecycle, request handling, route matching, or template rendering are performance-significant when they can affect latency, throughput, allocations, or cleanup. Capture before/after benchmark evidence for those paths.
