# Feature Specification: Bot Command Menu

**Feature Directory**: `specs/008-bot-command-menu`

**Feature Branch**: `develop` (current branch; directory numbering does not create a branch)

**Created**: 2026-11-15

**Status**: Implemented (2026-11-15). A start against the test bot changed `getMyCommands` from `[]` to the four commands of the contract and logged `bot: command menu set commands=4`. A list of `start` and `old` set before a start was replaced by the four commands, and three more restarts kept exactly those four (quickstart scenarios 5–6). The developer ran quickstart scenarios 2–4 and 7 in a Telegram app and confirmed them (2026-11-15)

**Input**: User description: "добавь чтобы в telegram bot при старте команды
сами регистрировались чтобы когда пользователь начинал вводить / ему
предлагались варианты команды с описанием" (Make the Telegram bot register its
commands at start, so that when a user starts typing `/` Telegram suggests the
commands with their descriptions.)

## Project Scope and Compatibility

- **Affected interfaces**: Telegram bot only. The command menu is a Telegram
  feature: the list shown when a user types `/` in the chat, and behind the
  Menu button next to the input field. The CLI already lists its commands
  with descriptions in `ge-tax-calc --help`, and no calculation or conversion
  capability changes, so parity (constitution II) holds without CLI changes.
- **Affected boundaries**:
  - `cmd/bot/`: the bot sends its command list to Telegram at every start and
    logs the result.
  - `README.md`: the "Bot Commands" section says that the bot sets the menu
    itself (FR-009).
  - `cmd/ge-tax-calc/`, `internal/`, `pkg/`: not changed.
- **Tax categories and rates**: N/A.
- **Currencies and rate dates**: N/A.
- **User-facing contract**: The Telegram command menu lists `/calculate`,
  `/convert`, `/cancel`, and `/help` with the descriptions in
  [Command Menu](#command-menu). `/start` is not in the menu and keeps
  working when typed or sent by Telegram's Start button. Command behavior,
  the `/start` and `/help` texts, and the flows do not change. No new
  commands, flags, or environment variables.
- **Release impact**: The bot binary and the bot Docker image. No generated
  code changes.

## Command Menu

The list the bot publishes, in this order. These are the commands listed by
`/start` and `/help`. The descriptions are those of the "Bot Commands" list in
`README.md`, without its "(guided step-by-step)" notes, so that they fit on
one line in the menu:

| Command | Description |
|---------|-------------|
| `/calculate` | Start the tax calculation flow |
| `/convert` | Start the currency conversion flow |
| `/cancel` | Cancel the current operation |
| `/help` | Show available commands |

## Clarifications

### Session 2026-11-15

- Q: Which commands should the Telegram menu list, and in what order? → A:
  Four commands without `/start`: `/calculate`, `/convert`, `/cancel`,
  `/help`, the same list as in `/start` and `/help`. `/start` keeps working
  when typed and through Telegram's Start button.
- Q: How should the command lists in the `/start` and `/help` texts stay in
  line with the menu? → A: The texts stay as they are; a test fails if either
  text lists other commands than the menu.
- Q: What should the bot do if it cannot publish the command menu at start?
  → A: Log the error and keep running, with no retry until the next start
  (FR-006). Telegram keeps the menu of the last successful start. Rejected:
  retries in the background, and stopping the start.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pick a command from the suggestions (Priority: P1)

A user opens the chat with the bot and types `/`. Telegram shows the bot's
commands with a short description of each. The user taps one, and the bot
answers as if the user had typed the command in full.

**Why this priority**: The user asked for this. Users can find the commands
without first reading `/help`.

**Independent Test**: Start the bot with a test token, open its chat in a
Telegram app, type `/`, check that the four commands from
[Command Menu](#command-menu) are listed with their descriptions, and tap
`/convert`. The bot asks for the first conversion step.

**Acceptance Scenarios**:

1. **Given** the bot has started, **When** a user types `/` in a private chat
   with the bot, **Then** Telegram lists `/calculate`, `/convert`, `/cancel`,
   and `/help` in this order, each with its description, and does not list
   `/start`.
2. **Given** the list is shown, **When** the user types `/c`, **Then** only
   `/calculate`, `/convert`, and `/cancel` remain (Telegram filters the list
   as the user types).
3. **Given** the list is shown, **When** the user picks any of the four
   commands, **Then** the bot gives the same reply as when the user types the
   command by hand.
4. **Given** the user is in the middle of a calculation, **When** the user
   picks `/cancel` from the list, **Then** the flow is canceled as with a
   typed `/cancel`.

---

### User Story 2 - Deploy the bot without setting the menu by hand (Priority: P2)

A bot operator creates a bot in BotFather, sets the token, and starts the
bot. The command menu appears without the BotFather `/setcommands` step. If
an earlier list was set by hand or by an older release, the bot replaces it,
so users do not see commands the bot no longer answers.

**Why this priority**: The menu stays correct across releases with no manual
step. Without this the operator must update BotFather by hand on every
command change.

**Independent Test**: With a test bot whose menu was set in BotFather to a
different list (for example `/start` and `/old`), start the bot and check
that the menu now shows exactly the four commands.

**Acceptance Scenarios**:

1. **Given** a new bot with no commands set, **When** the operator starts
   the bot, **Then** the menu shows the four commands.
2. **Given** a bot whose menu was set in BotFather to a different list,
   **When** the bot starts, **Then** the menu shows exactly the four commands
   and none of the earlier ones.
3. **Given** the bot is restarted several times, **When** a user types `/`,
   **Then** each command appears exactly once.

---

### User Story 3 - The bot works when the menu cannot be set (Priority: P3)

At start, Telegram rejects the command list or does not answer. The bot
still starts and answers users. The operator sees the failure in the log.

**Why this priority**: The menu is a convenience. A failure to set it must
not take the calculator offline.

**Independent Test**: Start the bot against a fake Telegram API that fails
the command list request. Check that the bot polls for updates, answers
`/help`, and logs the failure.

**Acceptance Scenarios**:

1. **Given** Telegram returns an error for the command list, **When** the bot
   starts, **Then** it logs the error with the reason and keeps starting.
2. **Given** the bot started after such a failure, **When** a user types
   `/start` or any of the four menu commands by hand, **Then** the bot
   answers as usual.
3. **Given** the command list was published, **When** the bot finishes
   starting, **Then** the log records that the menu was set and how many
   commands it has.

---

### Edge Cases

- Input, conversion, and calculation: N/A. No amounts, currencies, dates, or
  tax rules change.
- The command list cannot be published (network error, Telegram error,
  timeout): the bot logs the failure and runs without a menu update. Users
  keep the menu from the last successful start, or no menu. The next start
  tries again.
- `/start` typed by hand or sent by Telegram's Start button in a new chat:
  unchanged. The bot shows the welcome message, although `/start` is not in
  the menu.
- An unknown command such as `/foo`, typed by hand: unchanged. It is handled
  as text input (spec 004).
- A command picked from the menu during a flow: same as typing it. `/cancel`
  and `/help` reset the session; `/calculate` and `/convert` start a new flow
  in place of the current one, as they do now.
- Concurrent users: the list is the same for all chats; publishing it once at
  start does not depend on users.
- Shutdown: the menu stays set while the bot is offline. Users who pick a
  command then get no answer until the bot is back, as with a typed command
  today.
- A Telegram app shows an older list after a release: apps may cache the
  menu. Reopening the chat shows the current list.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: At every start, the bot MUST publish its command list to
  Telegram as the default list for all chats and all user languages. No
  manual BotFather step is required.
- **FR-002**: The list MUST contain exactly `/calculate`, `/convert`,
  `/cancel`, and `/help`, in this order, with the descriptions in
  [Command Menu](#command-menu). `/start` MUST NOT be in the list.
- **FR-003**: The bot MUST register its command handlers only from the
  command list and for `/start`, so it has no handler of its own for any
  command outside the list except `/start`. An automated test MUST fail if a
  command in the list gets no answer of its own.
- **FR-004**: Each description MUST be one line of English text that states
  the same purpose as the command's entry in `/help` and in the README "Bot
  Commands" list. An automated test checks the format; review checks the
  purpose.
- **FR-005**: The published list MUST replace the default list set earlier,
  whether by BotFather or by an earlier release.
- **FR-006**: If publishing fails, the bot MUST log the failure with its
  reason at error level and continue to start. It MUST NOT retry before the
  next start.
- **FR-007**: After a successful publish, the bot MUST log at info level that
  the command menu was set, with the number of commands.
- **FR-008**: Picking a command from the menu MUST behave the same as typing
  it. Command handling, including `/start`, the `/start` and `/help` texts,
  the flows, and the handling of unknown commands MUST NOT change.
- **FR-009**: `README.md` MUST state in "Bot Commands" that the bot sets the
  Telegram command menu at start with every command except `/start`, that
  BotFather `/setcommands` is not needed, and that a list set there is
  replaced on the next start.
- **FR-010**: The `/start` welcome text and the `/help` text MUST each list
  exactly the commands of the menu. Their wording stays as it is (FR-008).
  An automated test MUST fail if either text lists a command that is not in
  the menu or leaves out a menu command.

### Key Entities

- **Bot command**: A command the bot answers: its name without the slash
  (lowercase, for example `calculate`), a one-line description, and its
  position in the menu. `/start` is a bot command that is not in the menu.
- **Command list**: The ordered bot commands that the bot publishes to
  Telegram at every start (FR-001–FR-003).
- **Command menu**: What Telegram shows from the command list to all users
  of the bot when they type `/` or open the Menu button. Telegram keeps the
  last published list until it is replaced. "The bot sets the command menu"
  (FR-007, FR-009, the log line `bot: command menu set`) means that the bot
  published the command list.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For a newly created bot, typing `/` after the bot's first start
  lists all 4 menu commands with their descriptions, with no setup in
  BotFather.
- **SC-002**: Each of the 4 commands picked from the menu gets the same reply
  as the typed command (4 of 4).
- **SC-003**: When Telegram rejects the command list at start, the bot still
  answers all 5 commands typed by hand (`/start` and the 4 menu commands),
  and the log names the failure.
- **SC-004**: After any number of restarts, the menu lists exactly the 4
  commands, with no missing, extra, or repeated entries.

## Assumptions

- The menu is in English, like the rest of the bot. No per-language lists
  are published.
- One default list applies to all chats. No per-chat, group-only, or
  admin-only lists.
- Lists that someone set in BotFather for a specific language or scope are
  not cleared. Telegram prefers them over the default list for the users they
  cover. Clearing them is out of scope; the README note (FR-009) covers the
  default list only.
- The bot's Menu button keeps Telegram's default behavior of showing the
  command list. This feature does not change the Menu button.
- New users reach `/start` through Telegram's Start button, which Telegram
  shows in a chat that has no messages yet.
- The bot publishes the list once per start, before or while it begins
  answering users. The menu may appear up to 10 s after start, the publish
  timeout (plan research R3).
- No setting to turn the feature off. An operator who wants a different
  menu changes the command list in the code.
- Dependency: Telegram Bot API, the same service the bot already uses for
  messages.
