# Implementation Plan: Telegram Bot Platform

**Feature Directory**: `specs/004-telegram-bot-platform` | **Branch**: N/A (migrated)

**Date**: 2026-11-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/004-telegram-bot-platform/spec.md`

**Note**: Reverse-engineered from existing code (`status: migrated`). It
describes what is implemented, not a proposal.

## Summary

`cmd/bot` is a telego-based process. `main` reads environment configuration,
installs a signal-aware context, and calls `run`, which connects to Telegram,
starts long polling, and registers command, text, and callback handlers
wrapped with chat tracking. Conversation state lives in an in-memory session
store keyed by user ID; known chat IDs persist in a JSON file and receive
broadcasts at startup and shutdown. The bot ships as a goreleaser binary and
a Docker image run through Docker Compose.

## Technical Context

**Language/Version**: Go 1.27.3, single module with committed `vendor/`.

**Primary Dependencies**: `github.com/mymmrac/telego` v1.12 and
`telegohandler`; `github.com/obalunenko/logger`; `internal/service` for flows.

**Storage**: In-memory sessions; JSON file of chat IDs at
`USER_STORE_PATH`.

**Testing**: None in `cmd/bot`. Flow logic below the bot is tested in
`internal/service`, `internal/converter`, and `internal/taxes`.

**Target Platform**: Linux and macOS binaries (goreleaser `bot` build:
amd64/arm64); Docker image on Alpine 3.24 built from `golang:1.27.3-alpine`.

**Project Type**: Long-running Telegram bot process.

**Performance Goals**: Not specified.

**Constraints**: Calls only `internal/service` for domain work
(constitution I); capabilities must match the CLI (constitution II).

**Scale/Scope**: One process per token; sequential broadcasts; ~1270 lines in
`cmd/bot` including flows.

## Constitution Check

- [x] I. `cmd/bot` uses `service.Service`; imports from `internal/taxes`,
      `pkg/moneyutils`, `pkg/dateutils`, and `pkg/nbggovge/currencies` are
      for keyboards and display only.
- [x] II. Hosts the same capabilities as the CLI; differences are recorded in
      002 and 003.
- [x] III. N/A.
- [x] IV. N/A (no direct NBG access).
- [x] V. N/A.
- [ ] VI. No tests in `cmd/bot` (allowed by VI, recorded as a gap).
- [x] VII. N/A.
- [ ] Known deviations: binary named `ge-tax` in `.goreleaser.yml` but
      `ge-bot` in Docker and README; the service spinner runs in the bot
      process (existing deviations, not extended).

## Project Structure

### Documentation (this feature)

```text
specs/004-telegram-bot-platform/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/bot/
├── main.go              # env config, logger, signal context, run
├── bot.go               # telego bot, long polling, handler registration, chat tracking, broadcasts, shutdown
├── handlers.go          # /start, /help, /cancel, text and callback routing (flows: 002, 003)
├── session.go           # session, sessionStore (per user ID)
├── users.go             # userStore (JSON file), broadcast, messages
├── keyboards.go         # buildInlineKeyboard, buildConfirmKeyboard, callback prefix
└── send_message.go      # SendMessage with debug logging
build/docker/bot/Dockerfile
deployments/docker-compose/
├── bot-docker-compose.yml
└── .env.example
Makefile                 # bot-up, bot-down, bot-logs
.github/workflows/dockerfiles.yml   # hadolint and image build
.goreleaser.yml          # `bot` build
```

**Structure Decision**: All bot code is in `package main` under `cmd/bot`;
platform concerns (config, polling, state, broadcasts) and flows share the
package.

**Validation**: `go vet ./cmd/bot/...`; `make build`; `make bot-up` with a
test token; CI `build-bot-image` and hadolint.

## Technical Decisions

1. **Long polling** instead of webhooks: no public endpoint needed.
2. **Handler wrappers** (`trackUserMsg`, `trackUserCallback`) record chat IDs
   and log input before delegating.
3. **In-memory sessions** keyed by user ID, guarded by a mutex on the map
   only; flows mutate the returned `*session` directly.
4. **File-backed chat store**: whole file rewritten on each new chat; errors
   are ignored.
5. **Graceful shutdown**: the signal context stops polling; a goroutine
   broadcasts the maintenance message with a fresh 30-second context; `run`
   waits for it before returning.
6. **Multi-stage Docker build** with vendored dependencies and a static binary.

## Complexity Tracking

No justified violations. Gaps are recorded in [tasks.md](tasks.md).

| Factor | Assessment |
|--------|------------|
| Files | 7 Go files, Dockerfile, compose, env example |
| Concurrency | One goroutine per update (telegohandler); session fields not locked |
| State | In-memory sessions; JSON file of chat IDs |
| External coupling | Telegram Bot API; NBG through the service |
