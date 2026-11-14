---

description: "Migrated task list for the Telegram bot platform"
---

# Tasks: Telegram Bot Platform

**Input**: Design documents from `/specs/004-telegram-bot-platform/`

**Prerequisites**: plan.md, spec.md; hosts the flows from `specs/002-currency-conversion` and `specs/003-tax-calculation`

**Status**: migrated — all tasks reflect code that already exists on `master`.
Gaps found during migration are listed at the end without checkboxes; they are
candidates for follow-up features or bug reports, not pending work for this
feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Feature Preparation

- [x] T001 Add `github.com/mymmrac/telego` to `go.mod` and run `make sync-vendor`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T002 [P] Implement `session` and `sessionStore` (`get`, `reset`) in `cmd/bot/session.go`
- [x] T003 [P] Implement `sendMessage` with debug logging in `cmd/bot/send_message.go`
- [x] T004 [P] Implement `buildInlineKeyboard`, `buildConfirmKeyboard`, and the `cb:` prefix in `cmd/bot/keyboards.go`

---

## Phase 3: User Story 1 - Run the bot with environment configuration (Priority: P1) 🎯 MVP

- [x] T005 [US1] Read `TELEGRAM_BOT_TOKEN` and `TELEGRAM_BOT_LOG_LEVEL`, initialize the logger, and install the signal context in `cmd/bot/main.go`
- [x] T006 [US1] Implement `run`: create the bot, `getMe`, long polling, handler registration, and `bh.Start` in `cmd/bot/bot.go`

**Checkpoint**: The bot starts and stops cleanly.

---

## Phase 4: User Story 2 - Use service commands (Priority: P1)

- [x] T007 [US2] Implement `/start`, `/help`, and `/cancel` with session reset in `cmd/bot/handlers.go`
- [x] T008 [US2] Implement text and callback routing by active flow, callback answering, and `sendUnexpectedInput` in `cmd/bot/handlers.go`

**Checkpoint**: Users can discover commands and recover from any state.

---

## Phase 5: User Story 3 - Notify known chats about availability (Priority: P2)

- [x] T009 [US3] Implement `userStore` (load, save, `Track`, `All`) with `USER_STORE_PATH` in `cmd/bot/users.go` and `userStorePath` in `cmd/bot/bot.go`
- [x] T010 [US3] Wrap handlers with `trackUserMsg` and `trackUserCallback` in `cmd/bot/bot.go`
- [x] T011 [US3] Implement `broadcast` and the welcome-back and maintenance messages, with the 30-second shutdown broadcast, in `cmd/bot/users.go` and `cmd/bot/bot.go`

**Checkpoint**: Known chats are notified on restart.

---

## Phase 6: User Story 4 - Deploy with Docker Compose (Priority: P3)

- [x] T012 [P] [US4] Add the multi-stage `build/docker/bot/Dockerfile` (vendored build, static binary, Alpine runtime)
- [x] T013 [P] [US4] Add `deployments/docker-compose/bot-docker-compose.yml`, `.env.example`, and Makefile targets `bot-up`, `bot-down`, `bot-logs`
- [x] T014 [P] [US4] Add hadolint and the image build to `.github/workflows/dockerfiles.yml` and the `bot` build to `.goreleaser.yml`

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T015 [P] Document running the bot (binary and Docker Compose) and its commands in `README.md`

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **Possible data race on sessions.** telegohandler runs each update in its
   own goroutine (`vendor/github.com/mymmrac/telego/telegohandler/bot_handler.go`),
   but `sessionStore` locks only the map; handlers mutate `*session` fields
   without a lock. Fast repeated presses by one user can race. Not verified
   with `-race` because there are no tests.
2. **No panic recovery.** No `PanicRecovery` middleware is registered, so a
   panic in any handler terminates the process. Handlers dereference
   `msg.From` without a nil check (`cmd/bot/handlers.go:25,47,70,84,109,129`),
   unlike `trackUserMsg`.
3. **User store lost in Docker.** Compose defines no volume and the runtime
   image sets no `WORKDIR`, so `users.json` is written to `/users.json` inside
   the container and is lost when the container is recreated (`make bot-up`
   uses `--build`). Load and save errors are ignored silently.
4. **Broadcast behavior.** Every start and stop messages all known chats,
   sequentially and without throttling for Telegram rate limits; chats that
   blocked the bot are never removed; there is no opt-out.
5. **No tests in `cmd/bot`.** Routing, stores, keyboards, and formatting are
   untested.
6. **Unbounded sessions.** Sessions are never removed; one per user stays in
   memory until restart. Sessions are keyed by user ID while the store tracks
   chat IDs.
7. **Sensitive data in debug logs.** At debug level, user input and outgoing
   messages (including amounts) are logged.
8. **Container runs as root** (no `USER` in `build/docker/bot/Dockerfile`).
9. **Known deviation**: the binary is `ge-tax` in `.goreleaser.yml` but
   `ge-bot` in the Dockerfile and README.
