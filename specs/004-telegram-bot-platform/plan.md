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

**Testing**: ~~None in `cmd/bot`.~~ Since BUG-001 and BUG-002, `cmd/bot`
tests cover callback parsing and validation, step keyboards, amount
validation, and per-user serialization (run with `-race`); handlers that call
Telegram are still untested. Flow logic below the bot is tested in
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
- [ ] VI. ~~No tests in `cmd/bot`~~ Partial tests in `cmd/bot` since
      BUG-001/BUG-002; Telegram I/O is untested (allowed by VI).
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
   and log input before delegating. Since BUG-002: the `recoverPanics`
   middleware is registered first (`bh.Use`) and replies with a generic
   error; `serializeMsg` and `serializeCallback` hold a per-user lock for the
   whole update, and `serializeMsg` drops messages without `From`;
   `answerCallback` answers the query before the lock is taken. Order:
   track → answer → serialize → handler. (FR-016, FR-017, FR-018)
3. **In-memory sessions** keyed by user ID, guarded by a mutex on the map
   ~~only; flows mutate the returned `*session` directly~~. Since BUG-002,
   `sessionStore.lock` keeps one mutex per user, separate from the session
   because `reset` replaces it; flows still mutate `*session` directly, but
   only under that lock.
4. **File-backed chat store**: whole file rewritten on each new chat; errors
   ~~are ignored~~ are logged with the path since BUG-003: `newUserStore`
   logs load errors except a missing file, `Track` returns the save error and
   `trackUserMsg`/`trackUserCallback` log it. (FR-020)
5. **Graceful shutdown**: the signal context stops polling; ~~a goroutine
   broadcasts the maintenance message with a fresh 30-second context; `run`
   waits for it before returning.~~ since 2026-11-15 `bh.Start` returns when
   long polling closes the updates channel, then `shutdown` waits up to 30 s
   for running handlers and broadcasts the maintenance message within another
   30 s (FR-023). Running handlers are counted by `inflight`: `track` adds
   one per received update, the first middleware `done` marks it handled.
   `telegohandler.BotHandler.Stop` is not used because it cancels the
   handlers' contexts. Broadcasts are throttled to one message per 50 ms and
   drop chats that answer 403 (FR-021, FR-022).
6. **Multi-stage Docker build** with vendored dependencies and a static binary.
   Since BUG-003 the runtime stage creates a system user `bot` (UID/GID
   10001; numeric `USER` because hadolint DL3066 fails CI at
   `error_level: 2`) and `/data` owned by it, sets `WORKDIR /data` and
   `USER 10001:10001`, and declares
   `VOLUME /data`; compose mounts the named volume `bot-data` there. A new
   named volume copies the image's `/data` ownership, so the non-root user
   can write to it. (FR-013, FR-019)

**Bugfix**: 2026-11-15 — BUG-003 Updated from bugfix patch
7. **Step-tagged callbacks** (BUG-001): `stepOptions` returns the buttons of
   the current step; `stepKeyboard` builds the keyboard from them with
   callback data `cb:<step>:<value>`, and `callbackValue` accepts a callback
   only if its step is the current one and its value is among the same
   options. Building and checking use one source, so a button the current
   keyboard would not show is rejected. Tax type buttons send the type name
   as the value, so `taxTypeFromItem` is removed. (FR-012, FR-015)

**Bugfix**: 2026-11-15 — BUG-001 Updated from bugfix patch

**Bugfix**: 2026-11-15 — BUG-002 Updated from bugfix patch

## Complexity Tracking

No justified violations. Gaps are recorded in [tasks.md](tasks.md).

| Factor | Assessment |
|--------|------------|
| Files | 7 Go files, Dockerfile, compose, env example |
| Concurrency | One goroutine per update (telegohandler); ~~session fields not locked~~ updates of one user serialized by a per-user lock (BUG-002) |
| State | In-memory sessions; JSON file of chat IDs |
| External coupling | Telegram Bot API; NBG through the service |
