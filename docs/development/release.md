# Development, CI, and release

Repository workflow files are authoritative:

- `go.mod` defines the Go toolchain version, module path, and dependencies.
- `Makefile` defines local build, test, format, lint, and release entrypoints.
- `.github/workflows/build.yml` defines pull-request and main-branch CI.
- `.github/workflows/release.yml` and `.goreleaser.yml` define tag-driven releases.
- `.github/workflows/update-ferret-core.yml` defines Ferret dependency updates.
- Dockerfiles, `docker-entrypoint.sh`, `install.sh`, and `scripts` support packaging, versioning, installation, and release initiation.

This guide summarizes their roles. Check the files themselves before changing commands or automation.

## Local development commands

The default Make target is the broad local build workflow. The current targets provide:

- `make build`: vet, test, and compile
- `make test`: repository Go tests
- `make compile`: build the Lab binary with version metadata
- `make vet`: Go vet
- `make lint`: staticcheck and revive
- `make fmt`: Go formatting and repository-local goimports formatting
- `make install-tools`: install the lint and formatting tools used by the Makefile
- `make release <version>`: invoke the local tag release script

There is no repository-defined `make generate` target. Do not add generation to contributor instructions unless generated artifacts and their owning workflow are intentionally introduced.

Go is required. Use the version from `go.mod` rather than copying it into multiple documents. `make` is optional but is the preferred workflow entrypoint. Docker is needed only for container validation and packaging.

## Continuous integration

`.github/workflows/build.yml` runs for pushes and pull requests targeting the main branch. Its jobs establish the required validation shape:

- static analysis installs repository tools, runs vet and lint, applies formatting, and fails if formatting changes tracked files
- build validation downloads dependencies, builds the Lab binary, and runs the repository tests

Keep local targets and CI behavior aligned. A required new tool belongs in `make install-tools` or must be documented as an explicit prerequisite.

For narrow changes, run focused package checks first. Use the broader Make/CI-equivalent workflow when the scope warrants it. Never claim a command that was not actually run.

## Version metadata

Lab has two build-time version values:

- the Lab binary version injected into `main.version`
- the embedded Ferret version injected into `pkg/runtime.version`

The Makefile derives these values through `scripts/versions.sh`. GoReleaser injects the release tag and the Ferret version supplied by the release workflow. Changes must keep local builds, released binaries, and `lab version` consistent.

Version reporting should remain cheap and must not execute tests.

## Tag and release flow

The local release script validates a supplied version, requires a clean working tree, creates the Git tag, and pushes it. Pushing a matching version tag triggers `.github/workflows/release.yml`.

The release workflow:

1. Check out full tag history.
2. Derive the embedded Ferret version.
3. Authenticate to the configured container registries.
4. Invoke GoReleaser with a clean release workspace.
5. Notify the website through the organization reusable workflow after a successful v2 release.

GoReleaser creates a draft GitHub release. It builds static Lab binaries for the configured operating-system and architecture matrix, archives them, writes a checksum file, and builds the configured container images.

Version-specific container tags are always produced by the primary image definition. Stable aliases, including the unversioned latest tag, are skipped for prereleases. The release Dockerfile packages the prebuilt Lab binary with the Chromium base image and the Lab-aware entrypoint.

The entrypoint starts the browser environment for Lab command invocations while allowing an explicit non-Lab command to pass through to the container.

Release changes should preserve credential secrecy, least-privilege permissions, tag behavior, version injection, archive naming, image publication rules, and the draft-release contract.

## Ferret dependency automation

`.github/workflows/update-ferret-core.yml` receives the `ferret-core-released` repository dispatch. The workflow separates unprivileged preparation from privileged publication:

1. Validate the supplied v2 semantic version before privileged work.
2. Check out the repository without persisted credentials and record the base commit.
3. Run `go get` and `go mod tidy` for the released Ferret version.
4. Detect a no-op update or restrict staged changes to `go.mod` and `go.sum`.
5. Run tests and build validation when the dependency changed.
6. Verify the prepared diff and upload the dependency files as a short-lived artifact.
7. Delegate publication to the organization reusable workflow with the validated version and base commit.

The preparation job has read-only repository permissions. Secrets are passed only to the reusable publication job. Errors and logs must not expose application credentials or tokens.

Dependency automation changes must keep validation before authentication/publication, preserve cancellation behavior, avoid overwriting divergent work, and restrict generated changes to the intended dependency files.

## Website notification

The successful v2 release job calls the organization-owned website notification workflow with the release tag, release URL, and source commit. Lab announces the release; the website repository owns its storage, branch, content update, and pull request behavior.

Keep website implementation details out of this repository's workflow. Notification must remain conditional on successful publication and must not expose application credentials.

## Validating workflow changes

Use the narrowest checks that can validate the edited source:

- dry-run Make targets when checking command documentation
- shell syntax checks for scripts
- GoReleaser configuration checks for packaging changes when the tool is available
- Docker builds or focused entrypoint tests for container behavior
- workflow syntax and permission review for GitHub Actions
- `git diff --check` and a complete-diff review for every change

State any release, registry, reusable-workflow, or external-service behavior that could not be exercised locally.
