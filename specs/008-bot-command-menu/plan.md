# Implementation Plan: Bot Command Menu

**Feature Directory**: `specs/008-bot-command-menu` | **Branch**: `develop` (directory numbering does not create a branch)

**Date**: 2026-11-15 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/008-bot-command-menu/spec.md`

## Summary

At every start the Telegram bot publishes its command menu: `/calculate`,
`/convert`, `/cancel`, and `/help` with one-line descriptions, without
`/start`. Telegram then suggests them when a user types `/`. A single ordered
table in a new `cmd/bot/commands.go` defines the menu entries and their
handlers. `registerHandlers` registers them from that table, and `/start` is
registered on its own (research R2). `run` calls `setMyCommands` once after
`getMe` and before long polling, with a 10 s timeout and the default scope
and language. It logs the result, and a failure does not stop the bot
(research R1, R3, R4). Tests run against the existing fake Telegram API and a
recording fake logger (research R6). The README "Bot Commands" section gets
a note on the menu (research R7).

## Technical Context

**Language/Version**: Go 1.27.3, single module
`github.com/obalunenko/georgia-tax-calculator` with committed `vendor/`.

**Primary Dependencies**: telego (bot), obalunenko/logger. No dependency
changes: `Bot.SetMyCommands`, `SetMyCommandsParams`, and `BotCommand` are
already in the vendored telego.

**Storage**: N/A. Telegram stores the published menu; the bot stores
nothing new.

**Testing**: Go `testing` with testify, table-driven `t.Run` cases, race
detector on. Tests use the fake Telegram API in
`cmd/bot/fake_telegram_test.go`, which gains a `methodErrors` field, and the
bot harness. A new test-only recording `log.Logger` is placed in the context.
Files: new `cmd/bot/commands_test.go`; `cmd/bot/run_test.go` and
`cmd/bot/fake_telegram_test.go` extended. No network.

**Target Platform**: The bot on linux/darwin and as a Docker image.

**Project Type**: Interactive CLI and Telegram bot over a shared service layer,
plus reusable libraries in `pkg/`. This feature changes only the bot.

**Performance Goals**: One extra Bot API call per start, bounded by a 10 s
timeout (research R3). No effect on handling updates.

**Constraints**: Bot-only scope, justified in the spec (constitution II). No
domain, money, or NBG code is involved. New code stays within the
`.golangci.yml` limits (funlen, gocyclo, lll 140, wsl_v5, mnd: the timeout
and table are named).

**Scale/Scope**: Four menu entries; the five existing commands keep their
behavior. One call to Telegram per start, the same for all chats.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] I. Interfaces (`cmd/ge-tax-calc`, `cmd/bot`) reach domain logic only
      through `internal/service`; no direct `internal/converter` or
      `pkg/nbggovge` imports. *No new imports; the change uses telego and
      the logger only.*
- [x] II. CLI and bot stay in parity, or the spec justifies a one-interface
      scope. *The spec scopes the feature to the bot: the command menu is a
      Telegram feature, the CLI lists its commands in `--help`, and no
      capability changes.*
- [x] III. Money arithmetic uses `pkg/moneyutils`; rates use the NBG rate for
      the income date; `TaxType` changes update enum, rates, generated code,
      and tests together. *N/A: no money, rates, or tax types involved.*
- [x] IV. NBG access goes through `nbggovge.Client`; default tests use mocks
      or fixtures; real-API tests are gated by `INTEGRATION_TESTS`. *N/A for
      NBG. Bot tests use the fake Telegram API and need no network.*
- [x] V. `pkg/` does not import `internal/`. *`pkg/` is not changed.*
- [x] VI. Behavior changes in `internal/` and `pkg/` include table-driven
      testify tests that pass with `-race`. *No `internal/` or `pkg/`
      changes. `cmd/bot` tests, which the constitution encourages, are added
      anyway (research R6).*
- [x] VII. Generated code is regenerated, not hand-edited (`make codegen`
      for `TaxType`, the currencies generator for currency codes); `vendor/`
      is synced after `go.mod` changes. *No generated code, no `go.mod`
      change.*
- [x] Known deviations from the constitution are not extended.

**Post-design re-check (after Phase 1)**: All gates still pass. The design
adds one file to `cmd/bot/`, no package, and no cross-package dependency.

## Project Structure

### Documentation (this feature)

```text
specs/008-bot-command-menu/
├── plan.md              # This file
├── research.md          # Phase 0: decisions R1–R7
├── data-model.md        # Phase 1: menu table and lifecycle
├── quickstart.md        # Phase 1: automated and manual validation
├── contracts/
│   └── command-menu.md  # Phase 1: Bot API request, log lines, user view, README note
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 (/speckit-tasks; not created here)
```

### Source Code (repository root)

```text
cmd/
└── bot/
    ├── commands.go             # NEW: menuCommand, menuCommands table, botCommands(),
    │                           #      setCommandMenu(ctx, bot), setCommandsTimeout
    ├── bot.go                  # run: call setCommandMenu after GetMe;
    │                           # registerHandlers: /start plus a loop over menuCommands
    ├── handlers.go             # unchanged (cmd* constants and handlers reused)
    ├── commands_test.go        # NEW: table limits and order, harness replies,
    │                           #      /start and /help texts vs. menu (FR-010)
    ├── run_test.go             # setMyCommands call and params; failure keeps running; log lines
    └── fake_telegram_test.go   # methodErrors field; recording fake logger
README.md                       # "Bot Commands": menu note (FR-009)
```

**Structure Decision**: All code changes stay in `cmd/bot` (package `main`).
`commands.go` is new because the menu table is shared by `registerHandlers`
and `run` and is the single place to add a command. The `cmd*` constants and
the handler constructors stay in `handlers.go`. No new package, no new
import, and no change outside `cmd/bot/` and `README.md`.

**Validation**: `make vet`, `make test`, `make lint-pipeline`, `make build`
(Docker); locally `go vet ./...` and `go test -race ./...`. No `make codegen`,
currencies generator, `make sync-vendor`, or `make test-integration` is
needed. Manual check with a test bot per [quickstart.md](quickstart.md).

## Complexity Tracking

No constitution violations to justify.
