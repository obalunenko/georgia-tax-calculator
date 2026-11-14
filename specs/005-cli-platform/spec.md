# Feature Specification: CLI Platform

**Feature Directory**: `specs/005-cli-platform`

**Feature Branch**: N/A (migrated from `master`; implemented before Spec Kit adoption)

**Created**: 2026-11-15

**Status**: migrated

**Input**: Reverse-engineered from `cmd/ge-tax-calc/` (excluding the flows
covered by 002 and 003), `internal/spinner/`, `.goreleaser.yml` (`cli` build),
`.github/workflows/release.yml`, `scripts/build/app.sh`, `scripts/release/`,
and the README "Usage" section. Related history: `79b82c2` (Bubble Tea menu,
#196), `ddc2785` (urfave/cli v3 and Bubble Tea v2, #278).

## Project Scope and Compatibility

- **Affected interfaces**: CLI only. This feature is the CLI's runtime and
  release platform; the capabilities it hosts (tax calculation and currency
  conversion) are specified in
  [002-currency-conversion](../002-currency-conversion/spec.md) and
  [003-tax-calculation](../003-tax-calculation/spec.md). The bot platform is
  [004-telegram-bot-platform](../004-telegram-bot-platform/spec.md).
- **Affected boundaries**: `cmd/ge-tax-calc/` (`main.go`, `commands.go`,
  `handlers.go`, `version.go`), `internal/spinner/`, `.goreleaser.yml`
  (`cli` build, universal binary, `cli` archive, checksum, changelog,
  milestones), `.github/workflows/release.yml`, `scripts/build/app.sh`,
  `scripts/release/`, `Makefile` (`build`, `compile-app`, `release`,
  `release-local-snapshot`, `new-version`).
- **Tax categories and rates**: N/A.
- **Currencies and rate dates**: N/A.
- **User-facing contract**: Commands `run` and `convert`; global flags
  `--help`/`-h` and `--version`/`-v`; the banner, the "Exit..." line, and the
  unknown-command message; release archives
  `ge-tax-calc-<version>-<os>-<arch>` with `LICENSE` and `README.md`.
- **Release impact**: This feature is the CLI release itself.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start the CLI and choose a command (Priority: P1)

A user runs `ge-tax-calc`, sees the available commands, and starts tax
calculation or currency conversion.

**Why this priority**: Entry point to every CLI capability.

**Independent Test**: Manual: run the binary with no arguments, with
`--help`, with `run`, `convert`, and an unknown command. No automated tests.

**Acceptance Scenarios**:

1. **Given** no arguments, **When** `ge-tax-calc` runs, **Then** it prints
   the banner, the help (name, usage, description, commands `run` and
   `convert`, global options), and "Exit...".
2. **Given** `run` or `convert`, **When** the command runs, **Then** the
   banner is printed and the flow from 003 or 002 starts.
3. **Given** an unknown command `foo`, **When** it runs, **Then** the CLI
   prints "Command [foo] not supported.\nTry --help flag to see how to use
   it" and "Exit..." and exits with status 0.
4. **Given** a command returns an error, **When** the CLI finishes, **Then**
   the error is logged with "Run failed" and the process exits with a
   non-zero status (`log.Fatal`).

---

### User Story 2 - Check the version (Priority: P2)

A user or maintainer checks which build is installed.

**Why this priority**: Needed for bug reports and release checks.

**Independent Test**: Manual: `ge-tax-calc --version` on a release binary and
on a local build.

**Acceptance Scenarios**:

1. **Given** a binary built from a git checkout, **When** `--version` runs,
   **Then** it prints a table with `app_name` (module path of the main
   package), `version` (module version from build info), `short_commit`,
   `commit`, `build_date` (commit time), and `goversion`, followed by an
   ASCII cow.
2. **Given** a binary built from a tree with uncommitted changes, **When**
   `--version` runs, **Then** both commit fields end with `+CHANGES`.

---

### User Story 3 - See progress during a request (Priority: P3)

A user sees that the tool is waiting for NBG rates.

**Why this priority**: Rates come over the network; without feedback the CLI
looks frozen.

**Independent Test**: Manual: run `convert` and watch stderr. No automated
tests.

**Acceptance Scenarios**:

1. **Given** a conversion (also each income in `run`), **When** the service
   requests rates, **Then** a yellow spinner with "in progress...Converting
   <amount><from> to <to>" is shown on stderr with the cursor hidden, and
   "Converted <amount><from> to <to>" is printed when it stops.

---

### User Story 4 - Release CLI binaries (Priority: P3)

A maintainer publishes a new version.

**Why this priority**: Users install the CLI from GitHub releases (README
"Usage").

**Independent Test**: `make release-local-snapshot`; the release workflow on
a `v*` tag.

**Acceptance Scenarios**:

1. **Given** a pushed tag `vX.Y.Z`, **When** the release workflow runs,
   **Then** build, regression tests, and lint run first, and `make release`
   (goreleaser) publishes the GitHub release.
2. **Given** the `cli` build, **When** goreleaser runs, **Then** it builds
   `ge-tax-calc` with `CGO_ENABLED=0` and `-trimpath` for linux, darwin, and
   windows on amd64, arm, and arm64 (except darwin/arm and windows/arm), plus
   a macOS universal binary.
3. **Given** the archives, **When** they are created, **Then** each is
   `ge-tax-calc-<version>-<os>-<arch>[v<arm>]` (tar.gz, zip on windows),
   wrapped in a directory with `LICENSE` and `README.md`, and a checksum file
   is published.
4. **Given** the changelog, **When** the release is created, **Then** commits
   are grouped into "New Features" (`feat`), "Bug fixes" (`fix`), and "Other
   work", excluding `docs:`, `test:`, merge commits, and similar, and the
   matching milestone is closed.
5. **Given** `make build`, **When** it runs locally, **Then** goreleaser
   builds a single-target binary into `bin/${APP_NAME}` (default
   `ge-tax-calc`) after `sync-vendor` and `generate`.

---

### Edge Cases

- An unknown command exits with status 0, so scripts cannot detect it.
- The banner is printed before every command, including unknown commands
  and the help; "Exit..." is printed to stdout with `fmt.Println`, not to
  `cmd.Writer`, after every run.
- A binary built without VCS information (e.g., `go install` from the module
  proxy) prints `unset` for commit and build date.
- The `+CHANGES` suffix is added whenever build info has a `vcs.modified`
  key, including `vcs.modified=false`, so clean builds also show it (see
  tasks.md gaps).
- The spinner runs inside `internal/service`, so it also runs in the bot
  process (constitution known deviation).
- The CLI context has no signal handling; Ctrl+C inside a form is handled by
  Bubble Tea (002 US2.4), Ctrl+C during a rate request ends the process.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide the commands `run` (tax calculation, 003)
  and `convert` (currency conversion, 002).
- **FR-002**: The CLI MUST show name, usage, description, author, commands,
  and global options in its help, and support `--help`/`-h` and
  `--version`/`-v`.
- **FR-003**: The CLI MUST print the banner before running a command and
  "Exit..." after it.
- **FR-004**: An unknown command MUST print "Command [<name>] not supported."
  and a hint to use `--help`.
- **FR-005**: A command error MUST be logged and end the process with a
  non-zero status.
- **FR-006**: `--version` MUST report app name, version, short and full
  commit, build date, and Go version from the binary's build info
  (`github.com/obalunenko/version`), marking builds from a modified tree.
- **FR-007**: Long operations MUST show a spinner on stderr with a start and
  a final message (`internal/spinner`).
- **FR-008**: Releases MUST be built by goreleaser from a `v*` tag after the
  build, regression test, and lint jobs pass.
- **FR-009**: The `cli` build MUST be static (`CGO_ENABLED=0`), use
  `-trimpath` and the commit timestamp, and cover the OS/arch matrix in US4.2
  plus a macOS universal binary.
- **FR-010**: Release archives MUST follow the naming in US4.3, include
  `LICENSE` and `README.md`, and be listed in a checksum file.
- **FR-011**: The release changelog MUST group commits by Conventional
  Commit type as in US4.4.
- **FR-012**: `make build` MUST produce a local binary in `bin/` through
  goreleaser.

### Key Entities

- **Root command**: `cli.Command` named `ge-tax-calc` with commands, version
  string, `Before` (banner), `After` (exit line), and `CommandNotFound`.
- **Version info**: App name, version, commit, short commit, build date, Go
  version, read once at startup from `runtime/debug.ReadBuildInfo`.
- **Release artifact**: Archive per OS/arch, universal macOS binary, checksum
  file, changelog.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `ge-tax-calc --help` lists `run` and `convert`.
- **SC-002**: A release binary's `--version` shows the tag as `version` and
  the tagged commit.
- **SC-003**: Every `v*` tag produces a GitHub release with archives for all
  targets in US4.2 and a checksum file.

## Assumptions

- Users run the CLI in an interactive terminal; output is not meant to be
  parsed.
- `github.com/obalunenko/version` (same author) is the source of version
  data; goreleaser does not inject version through `-ldflags`.
- Releases are cut from `master` as described in the constitution.
