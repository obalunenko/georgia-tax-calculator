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

---

### Edge Cases

- Unknown commands are handled as text input (`AnyMessageWithText`).
- A missing or unreadable user store file is ignored at startup; write
  failures are ignored.
- Sessions are kept in memory only; a restart loses all in-progress flows.
- Each update is handled in its own goroutine; one user can trigger concurrent
  handlers by pressing buttons quickly.
- In Docker, the user store file lives inside the container (no volume) and
  is lost when the container is recreated.
- Chats that blocked the bot stay in the store and fail on every broadcast.

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
- **FR-012**: Inline keyboards MUST prefix callback data with `cb:` and lay
  out options in rows of a fixed width.
- **FR-013**: The Docker image MUST build a static (`CGO_ENABLED=0`) binary
  from vendored dependencies and run it as the entrypoint; compose MUST pass
  the three environment variables.
- **FR-014**: Outgoing messages and user input MUST be logged at debug level
  with the chat ID.

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

## Assumptions

- One bot instance runs per token; sessions are not shared between
  instances.
- The number of known chats is small enough for sequential broadcasts.
- Users accept being notified about restarts; there is no opt-out.
