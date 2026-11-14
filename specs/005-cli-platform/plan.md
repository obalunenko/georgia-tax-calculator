# Implementation Plan: CLI Platform

**Feature Directory**: `specs/005-cli-platform` | **Branch**: N/A (migrated)

**Date**: 2026-11-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-cli-platform/spec.md`

**Note**: Reverse-engineered from existing code (`status: migrated`). It
describes what is implemented, not a proposal.

## Summary

`cmd/ge-tax-calc` is a urfave/cli v3 application. `main` builds the root
command with two commands (`run`, `convert`), a `Before` hook that prints the
banner, an `After` hook that prints "Exit...", a `CommandNotFound` handler,
and a version string rendered at startup from build info. The flows behind
the commands are Bubble Tea forms (002, 003). `internal/spinner` wraps
`briandowns/spinner` and is started by `internal/service`. Releases are
built by goreleaser from `v*` tags in GitHub Actions; local builds use the
same goreleaser config through Docker.

## Technical Context

**Language/Version**: Go 1.27.3, single module with committed `vendor/`.

**Primary Dependencies**: `github.com/urfave/cli/v3` v3.14.0;
`github.com/obalunenko/version` v1.3.1 (build info through `runtime/debug`;
known defect: treats `vcs.modified=false` as modified, BUG-002);
`github.com/obalunenko/logger`; `github.com/briandowns/spinner`;
`github.com/savioxavier/termlink` (declarations link, 003); goreleaser v2.

**Storage**: None.

**Testing**: No tests for the root command, hooks, or version output;
`cmd/ge-tax-calc` tests cover prompts and forms only (002, 003).
`internal/spinner` is untested (constitution VI allows it).

**Target Platform**: linux, darwin, windows; amd64, arm, arm64 (no
darwin/arm, windows/arm); macOS universal binary.

**Project Type**: Interactive command-line application.

**Performance Goals**: Not specified.

**Constraints**: Calls only `internal/service` for domain work
(constitution I); capabilities match the bot (constitution II); static
binaries with vendored dependencies.

**Scale/Scope**: ~200 lines of platform code, one goreleaser build, one
release workflow.

## Constitution Check

- [x] I. `cmd/ge-tax-calc` imports `internal/service`, and `internal/taxes`,
      `pkg/moneyutils`, `pkg/dateutils`, `pkg/nbggovge/currencies` only for
      options and display.
- [x] II. Hosts the same capabilities as the bot (002, 003, 004).
- [x] III. N/A.
- [x] IV. N/A (no direct NBG access).
- [x] V. N/A.
- [ ] VI. No tests for the root command, hooks, version output, or
      `internal/spinner` (allowed by VI, recorded as gaps).
- [x] VII. No generated code involved.
- [ ] Known deviations: the spinner is started in `internal/service`, so it
      also runs for bot requests; the bot binary name differs between
      goreleaser and Docker (004). Existing deviations, not extended.

## Project Structure

### Documentation (this feature)

```text
specs/005-cli-platform/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/ge-tax-calc/
├── main.go          # root command: name, description, authors, commands, version, hooks; log.Fatal on error
├── commands.go      # `run` → menuCalcTaxes, `convert` → menuConvert (flows: 002, 003)
├── handlers.go      # printHeader (Before), notFound (CommandNotFound), onExit (After)
└── version.go       # printVersion: version table from github.com/obalunenko/version
internal/spinner/
└── spinner.go       # Start(name, finishMsg) StopFunc on stderr
.goreleaser.yml      # `cli` build, universal_binaries, `cli` archive, checksum, changelog, milestones
.github/workflows/release.yml   # build, regression_test, linting → release (make release)
scripts/build/app.sh            # goreleaser build --single-target --output bin/${APP_NAME}
scripts/release/                # release.sh, local-snapshot-release.sh, check.sh, new-version.sh
Makefile             # build, compile-app, release, release-local-snapshot, new-version
```

**Structure Decision**: `package main` holds the platform and the flow
forms; platform concerns are small and split by role (`main`, `commands`,
`handlers`, `version`).

**Validation**: `go build ./cmd/ge-tax-calc`; `make build`;
`make release-local-snapshot`; manual `--help`, `--version`, unknown command.

## Technical Decisions

1. **urfave/cli v3 root command** with hooks instead of a custom dispatcher:
   help, version flag, and command routing come from the library.
   BUG-001: when `CommandNotFound` is set, urfave/cli v3 calls it and returns
   `nil` (exit 0); without it, it returns `Exit("No help topic …", 3)`.
   FR-013 needs either removing `notFound` (and the custom message) or
   returning a `cli.Exit` error from the root command after `notFound` ran;
   the choice is made in T012.
2. **Version from build info** (`runtime/debug.ReadBuildInfo`) instead of
   `-ldflags -X`: goreleaser needs no ldflags, and `go install` builds also
   report a module version. BUG-002: the modified flag must come from the
   value of `vcs.modified`, fixed in the library (T013) and picked up by a
   dependency bump (T014).

**Bugfix**: 2026-11-15 — BUG-001 Updated from bugfix patch

**Bugfix**: 2026-11-15 — BUG-002 Updated from bugfix patch
3. **Spinner on stderr** so stdout keeps only results; the service starts it
   around each conversion.
4. **goreleaser for local and CI builds** so `make build` and releases use
   one configuration; tools run in Docker (`go-tools` image).
5. **Release only from tags** (`on.push.tags: v*`, branches excluded), gated
   by build, regression tests, and lint jobs.

## Complexity Tracking

No justified violations. Gaps are recorded in [tasks.md](tasks.md).

| Factor | Assessment |
|--------|------------|
| Files | 4 Go files in `cmd/ge-tax-calc` (platform part), 1 in `internal/spinner`, goreleaser config, 1 workflow, 5 scripts |
| Dependency depth | cmd → service (flows); version and logger libraries |
| State | None |
| External coupling | GitHub Releases, goreleaser, the `go-tools` Docker image |
