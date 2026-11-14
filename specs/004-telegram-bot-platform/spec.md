# Feature Specification: Telegram Bot Platform

**Feature Directory**: `specs/004-telegram-bot-platform`

**Feature Branch**: N/A (migrated from `master`; implemented before Spec Kit adoption)

**Created**: 2026-11-14

**Status**: migrated

**Input**: Reverse-engineered from `cmd/bot/` (excluding the flow-specific
parts covered by 002 and 003), `build/docker/bot/Dockerfile`,
`deployments/docker-compose/bot-docker-compose.yml`, and the README "Telegram
Bot" section. History: `0a76d1d` (Telegram bot, #221).

## Project Scope and Compatibility

- **Affected interfaces**: Telegram bot only. This feature is the bot's
  runtime platform; the user capabilities it hosts (conversion and tax
  calculation) exist in both interfaces and are specified in
  [002-currency-conversion](../002-currency-conversion/spec.md) and
  [003-tax-calculation](../003-tax-calculation/spec.md).
- **Affected boundaries**: `cmd/bot/` (`main.go`, `bot.go`, `session.go`,
  `users.go`, `send_message.go`, command routing in `handlers.go`, keyboard
  helpers in `keyboards.go`), `build/docker/bot/`,
  `deployments/docker-compose/`, `Makefile` (`bot-up`, `bot-down`,
  `bot-logs`), `.github/workflows/dockerfiles.yml`, `.goreleaser.yml` (`bot`
  build).
- **Tax categories and rates**: N/A.
- **Currencies and rate dates**: N/A.
- **User-facing contract**: Bot commands `/start`, `/help`, `/cancel` (plus
  `/calculate` and `/convert` from 002 and 003); broadcast messages; environment
  variables `TELEGRAM_BOT_TOKEN` (required), `TELEGRAM_BOT_LOG_LEVEL`
  (default `info`), `USER_STORE_PATH` (default `users.json`).
- **Release impact**: goreleaser `bot` build (binary `ge-tax`, linux/darwin,
  amd64/arm64) and the `ge-bot` Docker image built in CI (not pushed).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run the bot with environment configuration (Priority: P1)

An operator starts the bot with a token and optional log level and user store
path; the bot connects to Telegram and processes updates until stopped.

**Why this priority**: Nothing else works without a running bot.

**Independent Test**: Manual: run `ge-bot` with a test token and check the
"Bot started" log with the bot username.

**Acceptance Scenarios**:

1. **Given** `TELEGRAM_BOT_TOKEN` is not set, **When** the bot starts,
   **Then** it exits with "TELEGRAM_BOT_TOKEN environment variable is not
   set".
2. **Given** a valid token, **When** the bot starts, **Then** it calls
   `getMe`, logs "Bot started" with the username, and receives updates through
   long polling.
3. **Given** `TELEGRAM_BOT_LOG_LEVEL=debug`, **When** users interact, **Then**
   incoming input and outgoing responses are logged at debug level.
4. **Given** SIGINT or SIGTERM, **When** received, **Then** polling stops, the
   maintenance broadcast runs (up to 30 seconds), and the process exits.

---

### User Story 2 - Use service commands (Priority: P1)

A user discovers the bot, reads help, and cancels a flow.

**Why this priority**: Entry point and recovery for every conversation.

**Independent Test**: Manual in a test bot; no automated tests.

**Acceptance Scenarios**:

1. **Given** any state, **When** the user sends `/start`, **Then** the session
   is reset and a welcome message lists `/calculate`, `/convert`, `/cancel`,
   `/help`.
2. **Given** any state, **When** the user sends `/help`, **Then** the session
   is reset and command descriptions are shown.
3. **Given** a flow in progress, **When** the user sends `/cancel`, **Then**
   the session is reset and the bot replies "✅ Operation cancelled. Use
   /calculate or /convert to start again."
4. **Given** no flow, **When** the user sends text or presses a button,
   **Then** the bot replies "Please use /calculate or /convert to start."
5. **Given** a flow step that expects a button, **When** the user sends text,
   **Then** the bot replies "⚠️ Unexpected input. Please follow the flow or
   use /cancel to start over."
6. **Given** a button from an earlier step, another flow, or an older
   keyboard of the same step whose value the current step does not offer
   (e.g., day 31 while April is selected), **When** it is pressed, **Then**
   the bot replies "⚠️ Unexpected input …" and the step does not change
   (FR-015).
7. **Given** a handler panics, **When** the update is processed, **Then** the
   bot logs the panic, replies "❌ Something went wrong. Please use /cancel to
   start over.", and keeps running (FR-016).

---

### User Story 3 - Notify known chats about availability (Priority: P2)

The bot remembers chats it has talked to and tells them when it goes offline
and comes back.

**Why this priority**: Users in the middle of a flow learn that their session
was lost because of a restart.

**Independent Test**: Manual: interact, restart the bot, and check both
messages; inspect the `USER_STORE_PATH` file.

**Acceptance Scenarios**:

1. **Given** a message or button press from a new chat, **When** it is
   handled, **Then** the chat ID is added to the store and the file is
   rewritten; "bot: new user registered" is logged.
2. **Given** known chats in the store file, **When** the bot starts, **Then**
   each receives "👋 Bot is back online! Use /start to continue."
3. **Given** the bot is stopping, **When** shutdown begins, **Then** each known
   chat receives "🔧 Bot is going offline for maintenance. We'll be back
   soon!"
4. **Given** a send fails for one chat, **When** broadcasting, **Then** a
   warning is logged and the broadcast continues.

---

### User Story 4 - Deploy with Docker Compose (Priority: P3)

An operator runs the bot locally in Docker.

**Why this priority**: Alternative to running the binary (README option 2).

**Independent Test**: `make bot-up`, `make bot-logs`, `make bot-down`; CI
builds the image in `.github/workflows/dockerfiles.yml`.

**Acceptance Scenarios**:

1. **Given** `deployments/docker-compose/.env` with a token, **When**
   `make bot-up` runs, **Then** the image is built (static binary, Alpine
   runtime) and started with `restart: unless-stopped`.
2. **Given** no token in `.env`, **When** compose starts, **Then** it fails
   with "TELEGRAM_BOT_TOKEN is required".
3. **Given** a running container, **When** `make bot-logs` or `make bot-down`
   runs, **Then** logs are followed or the container is removed.
4. **Given** known chats stored by a running container, **When** `make
   bot-up` recreates the container (e.g., after an image change), **Then**
   the chats are still known and receive the welcome-back message, because
   the store lives in the `/data` volume (FR-019).
5. **Given** the running container, **When** its user is checked, **Then**
   the bot runs as a non-root user that can write to `/data` (FR-019).

---

### Edge Cases

- Unknown commands are handled as text input (`AnyMessageWithText`).
- ~~A missing or unreadable user store file is ignored at startup; write
  failures are ignored.~~ Superseded by BUG-003: a missing file is ignored;
  unreadable or invalid files and write failures are logged and the bot keeps
  running (FR-020).
- Sessions are kept in memory only; a restart loses all in-progress flows.
- Each update is handled in its own goroutine; one user can trigger concurrent
  handlers by pressing buttons quickly. Since BUG-002, updates of one user
  wait for each other (FR-017); updates of different users still run in
  parallel.
- Buttons sent before the `cb:<step>:<value>` format (BUG-001) are rejected
  as unexpected input after the upgrade.
- ~~In Docker, the user store file lives inside the container (no volume) and
  is lost when the container is recreated.~~ Superseded by BUG-003: it lives
  in the `/data` volume (FR-019); `make bot-down` keeps the volume, `docker
  compose down -v` removes it.
- ~~Chats that blocked the bot stay in the store and fail on every
  broadcast.~~ Since 2026-11-15 a chat that answers 403 is removed (FR-022).
- ~~On shutdown, `bh.Start` returns when the updates channel closes without
  waiting for running handlers; a calculation in progress is cut off when the
  process exits after the maintenance broadcast.~~ Since 2026-11-15 the bot
  waits up to 30 seconds for running handlers before the broadcast (FR-023).
- With one message per 50 ms, the 30-second shutdown broadcast reaches about
  600 chats; the rest do not get the maintenance message.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The bot MUST read the token from `TELEGRAM_BOT_TOKEN` and exit
  with a fatal log if it is empty.
- **FR-002**: The log level MUST come from `TELEGRAM_BOT_LOG_LEVEL`
  (default `info`).
- **FR-003**: The bot MUST receive updates through long polling and stop on
  SIGINT or SIGTERM.
- **FR-004**: The bot MUST register `/start`, `/help`, `/cancel`,
  `/calculate`, `/convert`, a text handler, and a callback handler.
- **FR-005**: `/start`, `/help`, and `/cancel` MUST reset the user's session.
- **FR-006**: Every callback query MUST be answered to clear the loading
  indicator, even if handling fails.
- **FR-007**: Text and callbacks MUST be routed to the active flow
  (calculate or convert); without a flow the bot MUST prompt to start one.
- **FR-008**: Sessions MUST be kept per Telegram user ID in memory.
- **FR-009**: Chat IDs MUST be tracked for every message and callback and
  persisted as a JSON array at `USER_STORE_PATH` (default `users.json`,
  file mode 0600).
- **FR-010**: On startup, known chats MUST receive the welcome-back message;
  on shutdown, the maintenance message, bounded by a 30-second timeout.
- **FR-011**: Broadcast failures for individual chats MUST be logged and MUST
  NOT stop the broadcast.
- **FR-012**: Inline keyboards MUST ~~prefix callback data with `cb:`~~ set
  callback data to `cb:<step>:<value>`, where `<step>` identifies the flow
  and step, within Telegram's 64-byte limit (BUG-001), and lay out options in
  rows of a fixed width.
- **FR-013**: The Docker image MUST build a static (`CGO_ENABLED=0`) binary
  from vendored dependencies and run it as the entrypoint; compose MUST pass
  the three environment variables. See FR-019 for the runtime user and
  storage (BUG-003).
- **FR-014**: Outgoing messages and user input MUST be logged at debug level
  with the chat ID.
- **FR-015**: A callback MUST be accepted only when its step equals the
  session's current step and its value is one of the options that step
  offers at that moment; otherwise the bot MUST reply with the unexpected
  input message and leave the session unchanged.
- **FR-016**: A panic in a handler MUST NOT stop the process; it MUST be
  logged with the stack, and the user MUST get a generic error reply.
- **FR-017**: Updates of one user MUST be handled one at a time; callback
  queries MUST still be answered before waiting (FR-006).
- **FR-018**: Messages without a sender (`From` empty) MUST be ignored by the
  handlers.
- **FR-019**: The Docker image MUST run the bot as a non-root user with
  working directory `/data`, writable by that user; compose MUST mount a
  named volume at `/data`, so the default `USER_STORE_PATH` (`users.json`)
  survives container recreation.
- **FR-020**: Errors reading the user store (other than a missing file) and
  errors writing it MUST be logged with the path; the bot MUST keep running
  with the chats it has in memory.
- **FR-021**: Broadcasts MUST send at most one message per 50 ms, below
  Telegram's limit of about 30 messages per second, and stop when their
  context is done. (Added 2026-11-15, T027.)
- **FR-022**: A chat that answers a broadcast with Telegram error 403 (the
  user blocked the bot, or the bot was removed) MUST be removed from the
  store; other errors MUST be logged and keep the chat. (Added 2026-11-15,
  T027.)
- **FR-023**: On shutdown, the bot MUST wait up to 30 seconds for running
  handlers to finish, without canceling them, and only then send the
  maintenance broadcast. (Added 2026-11-15, T028.)

**Bugfix**: 2026-11-15 — BUG-001 Changed FR-012 (step-tagged callback data),
added FR-015 and acceptance scenario US2.6, updated Edge Cases.

**Bugfix**: 2026-11-15 — BUG-002 Added FR-016, FR-017, FR-018, acceptance
scenario US2.7, SC-005, and updated Edge Cases.

**Bugfix**: 2026-11-15 — BUG-003 Added FR-019, FR-020, acceptance scenarios
US4.4 and US4.5, SC-006, and updated FR-013 and Edge Cases.

### Key Entities

- **Session**: Active flow (none, calculate, convert) and the flow-specific
  state from 002 and 003; one per user ID.
- **User store**: Set of chat IDs persisted to a JSON file.
- **Broadcast message**: Welcome-back or maintenance text sent to all known
  chats.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: With a valid token, the bot starts and answers `/start` in a
  test chat.
- **SC-002**: After a restart of the binary with the same `USER_STORE_PATH`,
  every previously seen chat receives the welcome-back message.
- **SC-003**: `/cancel` at any step returns the user to the idle state.
- **SC-004**: The CI image build (`build-bot-image`) succeeds for every PR.
- **SC-005**: `go test -race ./cmd/bot/...` passes and covers callback
  validation (stale step, same-step invalid value, other flow) and concurrent
  updates of one user.
- **SC-006**: In the bot image, `id` reports a non-root user, and a chat ID
  written by one container is loaded by a new container on the same volume
  (BUG-003).

## Assumptions

- One bot instance runs per token; sessions are not shared between
  instances.
- The number of known chats is small enough for sequential broadcasts
  (about 600 fit into the 30-second shutdown window at one message per
  50 ms).
- Users accept being notified about restarts; there is no opt-out.
