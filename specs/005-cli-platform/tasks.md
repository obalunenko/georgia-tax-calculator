---

description: "Migrated task list for the CLI platform"
---

# Tasks: CLI Platform

**Input**: Design documents from `/specs/005-cli-platform/`

**Prerequisites**: plan.md, spec.md; hosts the flows from `specs/002-currency-conversion` and `specs/003-tax-calculation`

**Status**: migrated — all tasks reflect code that already exists on `master`.
Gaps found during migration are listed at the end without checkboxes; they are
candidates for follow-up features or bug reports, not pending work for this
feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Feature Preparation

- [x] T001 Add `github.com/urfave/cli/v3`, `github.com/obalunenko/version`, and `github.com/briandowns/spinner` to `go.mod` and run `make sync-vendor`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T002 [P] Implement `spinner.Start` with a final message, hidden cursor, and stderr output in `internal/spinner/spinner.go`

---

## Phase 3: User Story 1 - Start the CLI and choose a command (Priority: P1) 🎯 MVP

- [x] T003 [US1] Build the root command (name, description, usage, authors, hooks) and exit with `log.Fatal` on error in `cmd/ge-tax-calc/main.go`
- [x] T004 [US1] Register `run` and `convert` in `cmd/ge-tax-calc/commands.go`
- [x] T005 [US1] Implement `printHeader`, `notFound`, and `onExit` in `cmd/ge-tax-calc/handlers.go`

**Checkpoint**: The CLI starts and routes to both flows.

---

## Phase 4: User Story 2 - Check the version (Priority: P2)

- [x] T006 [US2] Render the version table from `github.com/obalunenko/version` in `cmd/ge-tax-calc/version.go`

---

## Phase 5: User Story 3 - See progress during a request (Priority: P3)

- [x] T007 [US3] Start the spinner around each conversion in `service.Convert` (`internal/service/service.go`); see the known deviation in the constitution

---

## Phase 6: User Story 4 - Release CLI binaries (Priority: P3)

- [x] T008 [P] [US4] Configure the `cli` build, universal binary, `cli` archive, checksum, changelog groups, and milestones in `.goreleaser.yml`
- [x] T009 [P] [US4] Add the tag-triggered release workflow (build, regression tests, lint, release) in `.github/workflows/release.yml`
- [x] T010 [P] [US4] Add `scripts/build/app.sh`, `scripts/release/*.sh`, and Makefile targets `build`, `compile-app`, `release`, `release-local-snapshot`, `new-version`

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T011 [P] Document installation, commands, and global options in the README "Usage" section

---

## Phase B1: Bugfix BUG-001 - Unknown command exit status

- [x] T012 [US1] End an unknown command with a non-zero exit status (see plan decision 1) in `cmd/ge-tax-calc/main.go` and `cmd/ge-tax-calc/handlers.go`; test the exit code through `cli.OsExiter` in `cmd/ge-tax-calc/main_test.go`; update README if the message changes (FR-013, SC-004). Done on 2026-11-15: `notFound` removed, root command built by `newApp`; README does not show the message

**Bugfix**: 2026-11-15 — BUG-001 Updated from bugfix patch

---

## Phase B2: Bugfix BUG-002 - Modified flag on clean builds

- [x] T013 [US2] In `github.com/obalunenko/version`: set the modified flag from `setting.Value == "true"`, add a test for `vcs.modified=false`, and release a new version (separate repository). Done on 2026-11-15: obalunenko/version#37, tag `v1.4.0` (also bumps Go to 1.27 and runs tests in CI)
- [x] T014 [US2] Depends on T013: bump `github.com/obalunenko/version` in `go.mod`, run `make sync-vendor`, and check `--version` of a build from a clean worktree (FR-006, SC-005). Done on 2026-11-15: v1.4.0; a clean build of `7dae708` prints `short_commit: 7dae708`

**Bugfix**: 2026-11-15 — BUG-002 Updated from bugfix patch

---

## Phase F: Follow-up for migration gaps 5–7

- [x] T015 [P] [US1] Fix the description typo in `cmd/ge-tax-calc/main.go` and replace the README help block with the urfave/cli v3 output (VERSION section omitted); document the unknown-command exit status in `README.md` (gap 5)
- [x] T016 [P] [US4] Remove the unused `AOC_SESSION` secret from `.github/workflows/release.yml` and the `GOPRIVATE=example.com/blah` placeholder from `.goreleaser.yml`; `goreleaser check` passes (gap 6)
- [x] T017 [US4] Select the build with `--id "${BUILD_ID:-cli}"` in `scripts/build/app.sh` and fail when the binary is missing (gap 7, FR-012)

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **Fixed by T012 (BUG-001).** **Unknown command exits with status 0.** `notFound`
   (`cmd/ge-tax-calc/handlers.go`) only prints a message; `ge-tax-calc foo`
   returns 0 (checked on 2026-11-15), so scripts cannot detect the error.
2. **Fixed by T013 and T014 (BUG-002).** **`+CHANGES` on clean builds.** `github.com/obalunenko/version` sets
   `modified = true` whenever build info has a `vcs.modified` key
   (`vendor/github.com/obalunenko/version/version.go`, `init`), including
   `vcs.modified=false`. A build from a clean worktree of `0bafe32` printed
   `short_commit: 0bafe32+CHANGES` (checked on 2026-11-15). The fix belongs
   in the library (same author), then `make sync-vendor`.
3. **Partly fixed by T012.** **No tests for the platform.** The command list
   and the unknown command are tested; hooks, version output, and
   `internal/spinner` are not.
4. **Output noise.** The banner is printed before every command, including
   the help and unknown commands, and "Exit..." is printed with
   `fmt.Println` to stdout instead of `cmd.Writer`, so it cannot be
   redirected or tested through the command's writer.
5. **Fixed by T015.** **Stale README help.** The help block in README "Usage" comes from
   urfave/cli v2 (for example, `--help, -h  show help (default: false)`); v3
   prints `--help, -h  show help` and a different layout. The description
   also has a typo ("It get income amount") in both `main.go` and README.
6. **Fixed by T016.** **Release workflow leftovers.** The `regression_test` job of
   `.github/workflows/release.yml` passes `AOC_SESSION`, a secret not used by
   this project; `.goreleaser.yml` `gomod.env` contains the placeholder
   `GOPRIVATE=example.com/blah` (inactive because `proxy: false`).
7. **`make build` with two goreleaser builds.** `scripts/build/app.sh` runs
   `goreleaser build --single-target --output bin/${APP_NAME}` without
   `--id`, while `.goreleaser.yml` has two builds (`cli`, `bot`). Which
   binary ends up in `bin/` was not verified. **Fixed by T017.** Checked on
   2026-11-15: goreleaser ignored `--output`, built both binaries into
   `dist/`, left `bin/` empty, and the script still printed "Binary compiled
   at bin/ge-tax-calc".
8. **Spinner and signals.** The CLI context has no signal handling; the
   spinner hides the cursor, and whether the cursor is restored when Ctrl+C
   ends the process during a request was not verified. The spinner also
   runs in the bot process (known deviation).
9. **SC-002 not verified.** Whether a release binary reports the tag as
   `version` (module version from build info with `-trimpath` and
   `mod: vendor`) was not checked against a published release.
