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

## Phase B1: Bugfix BUG-001 - Stale callbacks

- [x] T016 [US2] Add step IDs and `session.stepID` in `cmd/bot/session.go`; build keyboards with callback data `cb:<step>:<value>` from `stepOptions` (`stepKeyboard`, `keyboardOption`) in `cmd/bot/keyboards.go` (FR-012)
- [x] T017 [US2] Accept a callback only through `callbackValue` (current step and one of its current options) in `handleCallback`, `cmd/bot/handlers.go`; tax type buttons send the type name, `taxTypeFromItem` removed (FR-015)
- [x] T018 [P] [US2] Tests for callback parsing and validation, step keyboards (every button accepted, 64-byte limit), tax type and day options in `cmd/bot/keyboards_test.go` (SC-005)

**Bugfix**: 2026-11-15 — BUG-001 Updated from bugfix patch

---

## Phase B2: Bugfix BUG-002 - Panics and concurrent updates

- [x] T019 [US1] Register the `recoverPanics` middleware first in `registerHandlers` (`cmd/bot/bot.go`): log the panic with the stack and reply with `msgInternalError` (FR-016)
- [x] T020 [US2] Add `sessionStore.lock` (one mutex per user) in `cmd/bot/session.go` and the `serializeMsg`/`serializeCallback` wrappers in `cmd/bot/bot.go`; `serializeMsg` drops messages without `From` (FR-017, FR-018)
- [x] T021 [US2] Answer callback queries in the `answerCallback` wrapper before the per-user lock (`cmd/bot/handlers.go`) (FR-006, FR-017)
- [x] T022 [P] [US2] Race test for one user's concurrent updates, per-user lock, `From == nil`, `updateChatID`, and step IDs in `cmd/bot/session_test.go`; removing the lock makes the test fail with `DATA RACE` (SC-005)

**Bugfix**: 2026-11-15 — BUG-002 Updated from bugfix patch

---

## Phase B3: Bugfix BUG-003 - Store persistence in Docker and store errors

- [x] T023 [US4] Create the `bot` system user (UID/GID 10001) and `/data` owned by it, set `WORKDIR /data`, `USER bot`, and `VOLUME /data` in `build/docker/bot/Dockerfile` (FR-019)
- [x] T024 [US4] Mount the named volume `bot-data` at `/data` in `deployments/docker-compose/bot-docker-compose.yml`; document it in `.env.example` and `README.md` (FR-019)
- [x] T025 [US3] Log store load errors (except a missing file) in `newUserStore` and return save errors from `Track`; log them in `trackUserMsg`/`trackUserCallback` (`cmd/bot/users.go`, `cmd/bot/bot.go`) (FR-020)
- [x] T026 [P] [US3] Table-driven tests for the store (round trip, missing file, invalid JSON, unwritable path) in `cmd/bot/users_test.go`; check the image with `id` and a two-container round trip on one volume (SC-006). Checked on 2026-11-15: `uid=10001(bot)`, `/data` owned by `bot`, a file written by one container is read by the next; hadolint passes

**Bugfix**: 2026-11-15 — BUG-003 Updated from bugfix patch

---

## Phase F: Follow-up for migration gaps 4 and 10 (2026-11-15)

- [x] T027 [US3] Throttle `broadcast` to one message per `broadcastInterval` (50 ms), stop on context cancel, and remove chats that answer 403 via `userStore.Remove` in `cmd/bot/users.go`; send through a `sendFunc` so it is testable; tests in `cmd/bot/users_test.go` (gap 4; FR-021, FR-022)
- [x] T028 [US1] Wait for running handlers before the maintenance broadcast: `inflight` (`track`, `done` middleware registered first, `wait`) and `shutdown` in `cmd/bot/bot.go`; test `TestInflightWaitsForHandledUpdates` in `cmd/bot/session_test.go` (gap 10, FR-023)
- [x] T029 [US2] Forget users idle for more than `sessionTTL` (24 h): `sessionStore.touch` (called by `get`, `reset`, and `lock`, sweeping at most once per `sessionSweepInterval`) in `cmd/bot/session.go`; test `TestSessionStoreForgetsIdleUsers` in `cmd/bot/session_test.go` (gap 6, FR-024)
- [x] T030 [P] [US1] Log user input by command or length (`inputLogField`) and outgoing messages by length in `cmd/bot/send_message.go` and `cmd/bot/bot.go`; test `TestInputLogField` in `cmd/bot/session_test.go` (gap 7, FR-014)
- [x] T031 [P] [US4] Rename the goreleaser `bot` binary from `ge-tax` to `ge-bot` in `.goreleaser.yml` (archives become `ge-bot-<version>-<os>-<arch>`); `goreleaser check` and a snapshot build of `--id bot` produce `ge-bot` (gap 9; constitution 1.0.5)

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **Fixed by T020 (BUG-002).** **Possible data race on sessions.** telegohandler runs each update in its
   own goroutine (`vendor/github.com/mymmrac/telego/telegohandler/bot_handler.go`),
   but `sessionStore` locks only the map; handlers mutate `*session` fields
   without a lock. Fast repeated presses by one user can race. Not verified
   with `-race` because there are no tests.
2. **Fixed by T019 and T020 (BUG-002).** **No panic recovery.** No `PanicRecovery` middleware is registered, so a
   panic in any handler terminates the process. Handlers dereference
   `msg.From` without a nil check (`cmd/bot/handlers.go:25,47,70,84,109,129`),
   unlike `trackUserMsg`.
3. **Fixed by T023–T026 (BUG-003).** **User store lost in Docker.** Compose defines no volume and the runtime
   image sets no `WORKDIR`, so `users.json` is written to `/users.json` inside
   the container and is lost when the container is recreated (`make bot-up`
   uses `--build`). Load and save errors are ignored silently.
4. **Mostly fixed by T027:** throttled and drops chats that answer 403; no opt-out yet. **Broadcast behavior.** Every start and stop messages all known chats,
   sequentially and without throttling for Telegram rate limits; chats that
   blocked the bot are never removed; there is no opt-out.
5. **Partly fixed by T018 and T022.** **No tests in `cmd/bot`.** Callback
   validation, keyboards, amount validation, and serialization are tested;
   handlers that call Telegram, the user store, and formatting are not.
6. **Fixed by T029.** **Unbounded sessions.** Sessions are never removed; one per user stays in
   memory until restart. Sessions are keyed by user ID while the store tracks
   chat IDs.
7. **Fixed by T030.** **Sensitive data in debug logs.** At debug level, user input and outgoing
   messages (including amounts) are logged.
8. **Fixed by T023 (BUG-003).** **Container runs as root** (no `USER` in `build/docker/bot/Dockerfile`).
   Fix together with gap 3: with a non-root `USER` and no writable
   `WORKDIR` or volume, writing `/users.json` fails, and the error is
   ignored, so chat tracking stops without any log.
9. **Fixed by T031.** **Known deviation**: the binary is `ge-tax` in `.goreleaser.yml` but
   `ge-bot` in the Dockerfile and README.
10. **Fixed by T028.** **Shutdown does not wait for handlers.** `bh.Start()` returns as soon as
    the long-polling channel closes
    (`vendor/github.com/mymmrac/telego/telegohandler/bot_handler.go:85`);
    only `Stop`/`StopWithContext` wait for running handlers. `run` waits for
    the maintenance broadcast but not for handlers, so a calculation in
    progress during shutdown is cut off without a reply.
11. **Fixed by T016–T018 (BUG-001).** **Stale callbacks.** Routing does not check which keyboard a callback
    came from; the effects on the flows are described in 002 gap 4 and 003
    gap 9. FR-012 (`cb:` prefix) will change to `cb:<step>:<value>`.
