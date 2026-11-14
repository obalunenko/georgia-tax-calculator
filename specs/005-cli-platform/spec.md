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
`--help`, with `run`, `convert`, and an unknown command. The unknown command
and the command list are covered by `cmd/ge-tax-calc/main_test.go`
(BUG-001).

**Acceptance Scenarios**:

1. **Given** no arguments, **When** `ge-tax-calc` runs, **Then** it prints
   ~~the banner,~~ the help (name, usage, description, commands `run` and
   `convert`, global options) ~~, and "Exit..."~~. Since 2026-11-15 the
   banner and "Exit..." belong to `run` and `convert` only (T019).
2. **Given** `run` or `convert`, **When** the command runs, **Then** the
   banner is printed and the flow from 003 or 002 starts.
3. **Given** an unknown command `foo`, **When** it runs, **Then** the CLI
   prints ~~"Command [foo] not supported.\nTry --help flag to see how to use
   it" and "Exit..."~~ ~~the banner and,~~ on stderr, "No help topic for 'foo'",
   and exits with ~~status 0~~ status 3; "Exit..." is not printed (BUG-001:
   status 0 made the error invisible to scripts; FR-013).
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
3. **Given** a binary built from a clean tree (`vcs.modified=false`),
   **When** `--version` runs, **Then** the commit fields have no `+CHANGES`
   suffix (BUG-002; fixed by `github.com/obalunenko/version` v1.4.0).

---

### User Story 3 - See progress during a request (Priority: P3)

A user sees that the tool is waiting for NBG rates.

**Why this priority**: Rates come over the network; without feedback the CLI
looks frozen.

**Independent Test**: Manual: run `convert` and watch stderr. No automated
tests.

**Acceptance Scenarios**:

1. **Given** a conversion ~~(also each income in `run`)~~, **When** the
   ~~service requests rates~~ CLI calls the service, **Then** a yellow spinner
   with "in progress...Converting <amount><from> to <to>" is shown on stderr
   with the cursor hidden, and "Converted <amount><from> to <to>" is printed
   when it stops.
2. **Given** `run`, **When** the CLI calls the service, **Then** one spinner
   "Calculating taxes for <N> incomes" runs for the whole calculation and
   "Calculated taxes for <N> incomes" is printed when it stops. (Since
   2026-11-15, T018.)

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

- ~~An unknown command exits with status 0, so scripts cannot detect it.~~
  Superseded by BUG-001: it exits with status 3 (FR-013, fixed by T012).
- ~~The banner is printed before every command, including unknown commands
  and the help; "Exit..." is printed to stdout with `fmt.Println`, not to
  `cmd.Writer`, after every run except an unknown command, where urfave/cli
  exits before the `After` hook (BUG-001).~~ Superseded by FR-003 (T019).
- A binary built without VCS information (e.g., `go install` from the module
  proxy) prints `unset` for commit and build date.
- The `+CHANGES` suffix is added whenever build info has a `vcs.modified`
  key, including `vcs.modified=false`, so clean builds also show it. This is
  a defect of `github.com/obalunenko/version` v1.3.1 (BUG-002), not intended
  behavior; fixed by upgrading to v1.4.0 (T014).
- ~~The spinner runs inside `internal/service`, so it also runs in the bot
  process (constitution known deviation).~~ Since 2026-11-15 the CLI starts
  it around service calls; the bot has no spinner (constitution 1.0.4).
- ~~The CLI context has no signal handling; Ctrl+C inside a form is handled by
  Bubble Tea (002 US2.4), Ctrl+C during a rate request ends the process.~~
  Since 2026-11-15 SIGINT and SIGTERM cancel the CLI context: a rate request
  in progress returns, the spinner stops and shows the cursor again, and the
  CLI exits with "Run failed". Ctrl+C inside a form is still a key press
  handled by Bubble Tea (002 US2.4). (T020)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide the commands `run` (tax calculation, 003)
  and `convert` (currency conversion, 002).
- **FR-002**: The CLI MUST show name, usage, description, author, commands,
  and global options in its help, and support `--help`/`-h` and
  `--version`/`-v`.
- **FR-003**: The CLI MUST print the banner before running ~~a command~~ `run`
  or `convert` and "Exit..." after it, both to the command's writer; help,
  version, and unknown commands print neither. (Changed 2026-11-15, T019.)
- **FR-004**: An unknown command MUST print ~~"Command [<name>] not
  supported." and a hint to use `--help`~~ urfave/cli's "No help topic for
  '<name>'" on stderr (BUG-001: the custom handler could not return an
  error).
- **FR-005**: A command error MUST be logged and end the process with a
  non-zero status.
- **FR-006**: `--version` MUST report app name, version, short and full
  commit, build date, and Go version from the binary's build info
  (`github.com/obalunenko/version`), marking ~~builds from a modified tree~~
  only builds whose build info has `vcs.modified=true` (BUG-002).
- **FR-007**: Long operations MUST show a spinner on stderr with a start and
  a final message (`internal/spinner`), started by the CLI around service
  calls, not by the service.
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
  goreleaser. [Was not met until 2026-11-15: `bin/` stayed empty with two
  goreleaser builds; fixed by T017.]
- **FR-013**: An unknown command MUST end the process with a non-zero exit
  status, in addition to the message in FR-004. Implemented on 2026-11-15:
  status 3 from urfave/cli (T012).

**Bugfix**: 2026-11-15 — BUG-001 Added FR-013 (non-zero exit for an unknown
command), updated acceptance scenario US1.3, Edge Cases, and SC-004.

**Bugfix**: 2026-11-15 — BUG-002 Updated FR-006, added acceptance scenario
US2.3, updated Edge Cases, Assumptions, and SC-005.

### Key Entities

- **Root command**: `cli.Command` named `ge-tax-calc` with commands, version
  string, `Before` (banner), and `After` (exit line), built by `newApp`;
  ~~`CommandNotFound`~~ left unset on purpose (BUG-001).
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
- **SC-004**: `ge-tax-calc foo; echo $?` prints a non-zero value (BUG-001).
- **SC-005**: `--version` of a binary built from a clean worktree shows no
  `+CHANGES` (BUG-002).

## Assumptions

- Users run the CLI in an interactive terminal; output is not meant to be
  parsed.
- `github.com/obalunenko/version` (same author) is the source of version
  data; goreleaser does not inject version through `-ldflags`. v1.3.1
  treated `vcs.modified=false` as modified (BUG-002); v1.4.0 (requires Go
  1.27) fixes it and is used since 2026-11-15.
- Releases are cut from `master` as described in the constitution.
